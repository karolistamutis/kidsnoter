package downloading

import (
	"bytes"
	"context"
	"fmt"
	"github.com/karolistamutis/kidsnoter/listing"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/karolistamutis/kidsnoter/models"
	"github.com/karolistamutis/kidsnoter/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const concurrentDownloads = 5

var (
	downloadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "downloads_total",
		Help: "The total number of downloads",
	}, []string{"type"})
	downloadErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "download_errors_total",
		Help: "The total number of download errors",
	}, []string{"type"})
	downloadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "download_duration_seconds",
		Help:    "The duration of downloads in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	}, []string{"type"})
	downloadSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "download_size_bytes",
		Help:    "The size of downloads in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
	}, []string{"type"})
)

type Downloader struct {
	lister    listing.Lister
	client    *http.Client
	overwrite bool
	tmpl      *template.Template
}

// NewDownloader creates a new Downloader instance
func NewDownloader(lister listing.Lister, client *http.Client, overwrite bool) (*Downloader, error) {
	tmpl, err := template.ParseFiles("templates/album.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Downloader{
		lister:    lister,
		client:    client,
		overwrite: overwrite,
		tmpl:      tmpl,
	}, nil
}

// DownloadAlbums downloads all albums for a given child
func (d *Downloader) DownloadAlbums(ctx context.Context, childID int, outputDir string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Hour)
	defer cancel()

	albumChan := make(chan *models.Album)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	go func() {
		defer close(albumChan)
		if err := d.lister.ListAlbums(ctx, childID, albumChan); err != nil {
			errChan <- err
		}
	}()

	semaphore := make(chan struct{}, concurrentDownloads)
	var downloadErrs []error
	mu := &sync.Mutex{}

	totalAlbums := 0
	successfulAlbums := 0

	for album := range albumChan {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			totalAlbums++
			wg.Add(1)
			go func(a *models.Album) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if err := d.downloadAlbum(ctx, a, outputDir); err != nil {
					mu.Lock()
					downloadErrs = append(downloadErrs, err)
					downloadErrors.WithLabelValues("album").Inc()
					mu.Unlock()
				} else {
					successfulAlbums++
					downloadsTotal.WithLabelValues("album").Inc()
				}
			}(album)
		}
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		downloadErrs = append(downloadErrs, err)
	}

	if len(downloadErrs) > 0 {
		return fmt.Errorf("encountered errors during download: %v", downloadErrs)
	}
	return nil
}

func (d *Downloader) downloadAlbum(ctx context.Context, album *models.Album, outputDir string) error {
	timer := prometheus.NewTimer(downloadDuration.WithLabelValues("album"))
	defer timer.ObserveDuration()

	outputDir, err := util.ExpandTilde(outputDir)
	if err != nil {
		return fmt.Errorf("failed to expand output directory path: %w", err)
	}

	albumDir := filepath.Join(outputDir, album.GeneratedFolderName)

	logger.Log.Infof("Downloading album \"%s\" to directory %s", album.Title, albumDir)

	if err := os.MkdirAll(albumDir, 0755); err != nil {
		if os.IsExist(err) && !d.overwrite {
			logger.Log.Infof("Skipping album \"%s\", directory already exists", album.Title)
			return nil
		}
		return fmt.Errorf("failed to create album directory: %w", err)
	} else {
		// TODO new albums
	}

	if err := d.writeAlbumMetadata(album, albumDir); err != nil {
		return fmt.Errorf("failed to write album metadata: %w", err)
	}

	for _, image := range album.Images {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := d.downloadImage(ctx, image, albumDir); err != nil {
				logger.Log.Errorf("Error downloading image %s, %v", image.DownloadLink, err)
				downloadErrors.WithLabelValues("image").Inc()
			} else {
				downloadsTotal.WithLabelValues("image").Inc()
			}
		}
	}

	if album.Video != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := d.downloadVideo(ctx, album.Video, albumDir); err != nil {
				logger.Log.Errorf("Error downloading video for album %d: %v", album.ID, err)
				downloadErrors.WithLabelValues("video").Inc()
			} else {
				downloadsTotal.WithLabelValues("video").Inc()
			}
		}
	}

	return nil
}

func (d *Downloader) downloadImage(ctx context.Context, image *models.Image, albumDir string) error {
	imageExtension, err := util.ExtractExtensionFromURL(image.DownloadLink)
	if err != nil {
		return fmt.Errorf("failed to extract image file extension: %v", err)
	}
	fileName := filepath.Join(albumDir, fmt.Sprintf("%d.%s", image.ID, imageExtension))
	if !d.overwrite && util.FileExistsAndMatches(fileName, image.FileSize) {
		logger.Log.Infof("Skipping image %s: File already exists and matches size", fileName)
		return nil
	}
	return util.RetryWithBackoff(ctx, func() error {
		return d.downloadFile(ctx, image.DownloadLink, fileName, image.FileSize, "image")
	})
}

func (d *Downloader) downloadVideo(ctx context.Context, video *models.Video, albumDir string) error {
	videoExtension, err := util.ExtractExtensionFromURL(video.DownloadLink)
	if err != nil {
		return fmt.Errorf("failed to extract video file extension: %v", err)
	}
	videoFileName := filepath.Join(albumDir, fmt.Sprintf("video.%s", videoExtension))
	if !d.overwrite && util.FileExistsAndMatches(videoFileName, video.FileSize) {
		logger.Log.Infof("Skipping video %s: File already exists and matches size", videoFileName)
		return nil
	}
	return util.RetryWithBackoff(ctx, func() error {
		return d.downloadFile(ctx, video.DownloadLink, videoFileName, video.FileSize, "video")
	})
}

func (d *Downloader) downloadFile(ctx context.Context, url, filepath string, expectedSize int, fileType string) error {
	timer := prometheus.NewTimer(downloadDuration.WithLabelValues(fileType))
	defer timer.ObserveDuration()

	if !d.overwrite {
		if existing, err := os.Stat(filepath); err == nil {
			if existing.Size() == int64(expectedSize) {
				logger.Log.Infof("Skipping file %s: Already exists and matches size.", filepath)
				return nil
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	downloadSize.WithLabelValues(fileType).Observe(float64(expectedSize))

	return nil
}

func (d *Downloader) writeAlbumMetadata(album *models.Album, albumDir string) error {
	descriptionFile := filepath.Join(albumDir, "description.md")

	// Check if the metadata file exists
	if _, err := os.Stat(descriptionFile); os.IsNotExist(err) {
		return d.createAlbumMetadata(descriptionFile, album)
	}

	// File exists, read and parse it
	content, err := os.ReadFile(descriptionFile)
	if err != nil {
		return fmt.Errorf("failed to read album metadata file: %v", err)
	}

	existingTitle, existingDate, err := util.ParseFrontmatter(string(content))
	if err != nil {
		// Something is wrong reading the frontmatter, just update the file
		return d.createAlbumMetadata(descriptionFile, album)
	}

	// Check if update needed or overwrite is set
	if existingTitle != album.Title || existingDate != album.Date || d.overwrite {
		return d.createAlbumMetadata(descriptionFile, album)
	}

	return nil
}

func (d *Downloader) createAlbumMetadata(descriptionFile string, album *models.Album) error {
	data := struct {
		Title   string
		Date    string
		Content string
	}{
		Title:   album.Title,
		Date:    album.Date,
		Content: album.Content,
	}

	content := new(bytes.Buffer)
	if err := d.tmpl.Execute(content, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	f, err := os.Create(descriptionFile)
	if err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}
	defer f.Close()

	if err := os.WriteFile(descriptionFile, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

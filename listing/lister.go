package listing

import (
	"context"
	"fmt"
	"github.com/karolistamutis/kidsnoter/config"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/karolistamutis/kidsnoter/models"
	"github.com/karolistamutis/kidsnoter/util"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Lister interface {
	ListChildren(ctx context.Context) ([]*models.Child, error)
	ListAlbums(ctx context.Context, childID int, albumChan chan<- *models.Album) error
	Children() ([]*models.Child, error)
}

type lister struct {
	client   *http.Client
	children []*models.Child
}

func NewLister(client *http.Client) Lister {
	return &lister{client: client}
}

func (l *lister) ListChildren(ctx context.Context) ([]*models.Child, error) {
	var children []*models.Child

	err := util.RetryWithBackoff(ctx, func() error {
		apiInfoURL := config.GetAPIInfoURL()
		req, err := http.NewRequestWithContext(ctx, "GET", apiInfoURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for URL %s: %v", apiInfoURL, err)
		}

		resp, err := l.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute request for URL %s: %v", apiInfoURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-OK HTTP status %s from URL %s", resp.Status, apiInfoURL)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body from URL %s: %v", apiInfoURL, err)
		}

		var p fastjson.Parser
		value, err := p.ParseBytes(body)
		if err != nil {
			return fmt.Errorf("failed to parse JSON data: %v", err)
		}

		// Extract the children data
		childrenArray := value.GetArray("children")
		if childrenArray == nil {
			return fmt.Errorf("no children data found in JSON from URL %s", apiInfoURL)
		}

		children = make([]*models.Child, 0, len(childrenArray))
		for _, child := range childrenArray {
			id := child.GetInt("id")
			name := string(child.GetStringBytes("name"))
			gender := string(child.GetStringBytes("gender"))
			dateOfBirth := string(child.GetStringBytes("date_birth"))

			enrollment := child.GetArray("enrollment")
			if len(enrollment) == 0 {
				return fmt.Errorf("enrollment details missing for child ID %d", id)
			}

			centerID := enrollment[0].GetInt("center_id")
			classID := enrollment[0].GetInt("belong_to_class")

			if centerID == 0 || classID == 0 {
				return fmt.Errorf("invalid center or class IDs for child %s [ID: %d]", name, id)
			}

			children = append(children, &models.Child{
				ID:          id,
				CenterID:    centerID,
				ClassID:     classID,
				Name:        name,
				Gender:      gender,
				DateOfBirth: dateOfBirth,
			})
		}
		l.children = children
		return nil
	})

	if err != nil {
		return nil, err
	}

	return l.children, nil
}

func (l *lister) Children() ([]*models.Child, error) {
	if l.children == nil {
		return nil, fmt.Errorf("children data has not been populated, call ListChildren first")
	}
	return l.children, nil
}

func (l *lister) ListAlbums(ctx context.Context, childID int, albumChan chan<- *models.Album) error {
	// If childID is 0, list albums for all children in l.children slice.
	var wg sync.WaitGroup
	var errs []error  // Slice to collect errors
	var mu sync.Mutex // Protects access to errs

	// Handle case where specific child ID is provided
	if childID != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.listAlbumsForChild(ctx, childID, albumChan); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	} else {
		// Ensure children data is populated
		if l.children == nil || len(l.children) == 0 {
			return fmt.Errorf("children data has not been populated, check child ID or name first")
		}

		// Iterate over all children
		for _, child := range l.children {
			child := child // Capture range variable

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := l.listAlbumsForChild(ctx, child.ID, albumChan); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}()
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("encountered errors during album listing: %v", errs)
	}

	return nil
}

func (l *lister) listAlbumsForChild(ctx context.Context, childID int, albumChan chan<- *models.Album) error {
	// Find the child name from l.children.
	var childName string
	for _, child := range l.children {
		if child.ID == childID {
			childName = child.Name
			break
		}
	}

	if childName == "" {
		return fmt.Errorf("child with ID %d not found", childID)
	}

	next := fmt.Sprintf(config.GetAPIAlbumURL(), childID)
	for next != "" {
		body, err := l.albumData(ctx, next)
		if err != nil {
			return fmt.Errorf("failed to make request for URL %s: %v", next, err)
		}

		p := fastjson.Parser{}
		value, err := p.ParseBytes(body)
		if err != nil {
			return fmt.Errorf("failed to parse JSON data from URL %s: %v", next, err)
		}

		next, err = l.nexAlbumsURL(next, value)
		if err != nil {
			return fmt.Errorf("failed to extract next URL for URL %s: %v", next, err)
		}

		err = l.streamPageAlbums(childName, value, albumChan)
		if err != nil {
			return fmt.Errorf("failed to get page albums from URL %s: %v", next, err)
		}
	}
	return nil
}

func (l *lister) albumData(ctx context.Context, albumsURL string) ([]byte, error) {
	var body []byte
	var respErr error

	err := util.RetryWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", albumsURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for URL %s: %v", albumsURL, err)
		}

		resp, err := l.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute request for URL %s: %v", albumsURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-OK HTTP status %s from URL %s", resp.Status, albumsURL)
		}

		body, respErr = io.ReadAll(resp.Body)
		return respErr
	})

	if err != nil {
		return nil, err
	}
	return body, nil
}

func (l *lister) nexAlbumsURL(baseURL string, value *fastjson.Value) (string, error) {
	next := value.GetStringBytes("next")
	if next == nil {
		return "", nil
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %s: %v", baseURL, err)
	}

	q := u.Query()
	q.Set("page", string(next))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (l *lister) streamPageAlbums(childName string, value *fastjson.Value, albumChan chan<- *models.Album) error {
	albumArray := value.GetArray("results")
	if albumArray == nil {
		return fmt.Errorf("no album data found")
	}

	for _, album := range albumArray {
		id := album.GetInt("id")
		date := string(album.GetStringBytes("created"))
		title := string(album.GetStringBytes("title"))
		content := string(album.GetStringBytes("content"))
		video := l.getAlbumVideo(album.GetObject("attached_video"))
		images := l.getAlbumImages(album.GetArray("attached_images"))

		logger.Log.Debugf("got %d images for album \"%s\"", len(images), title)

		generatedFolderName, err := util.GenerateFolderName(childName, date, id, title)
		if err != nil {
			logger.Log.Warnf("failed to generate folder name for album ID %d: %v, skipping", id, err)
			continue
		}
		logger.Log.Debugf("generated folder name: %s for album title %s", generatedFolderName, title)

		albumChan <- &models.Album{
			ID:                  id,
			GeneratedFolderName: generatedFolderName,
			Date:                date,
			Title:               title,
			Content:             content,
			Video:               video,
			Images:              images,
		}
	}
	return nil
}

func (l *lister) getAlbumVideo(videoObject *fastjson.Object) *models.Video {
	if videoObject == nil {
		return nil
	}
	id, err := videoObject.Get("id").Int()
	if err != nil {
		id = 0
	}
	fileName := strings.Trim(videoObject.Get("original_file_name").String(), "\"")
	fileSize, err := videoObject.Get("file_size").Int()
	if err != nil {
		fileSize = 0
	}
	downloadLink := strings.Trim(videoObject.Get("high").String(), "\"")

	return &models.Video{
		ID:           id,
		FileName:     fileName,
		FileSize:     fileSize,
		DownloadLink: downloadLink,
	}
}

func (l *lister) getAlbumImages(imageArray []*fastjson.Value) []*models.Image {
	if imageArray == nil {
		return nil
	}

	var images []*models.Image
	for _, image := range imageArray {
		images = append(images, &models.Image{
			ID:           image.GetInt("id"),
			FileSize:     image.GetInt("file_size"),
			DownloadLink: string(image.GetStringBytes("original")),
		})
	}
	return images
}

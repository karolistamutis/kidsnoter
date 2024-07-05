package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/karolistamutis/kidsnoter/config"
	"github.com/karolistamutis/kidsnoter/downloading"
	"github.com/karolistamutis/kidsnoter/listing"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run in continuous album synchronization mode",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().Bool("overwrite", false, "Overwrite existing files")
	RootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	overwrite, err := cmd.Flags().GetBool("overwrite")
	if err != nil {
		return fmt.Errorf("error getting overwrite flag: %w", err)
	}

	return serveAlbums(ctx, overwrite)
}

func serveAlbums(ctx context.Context, overwrite bool) error {
	lister := listing.NewLister(client)
	albumDir := config.GetAlbumDir()
	if albumDir == "" {
		return fmt.Errorf("missing or empty invalid album_dir setting")
	}

	syncInterval := config.GetSyncInterval()
	if syncInterval <= 0 {
		return fmt.Errorf("invalid sync interval: %v, must be greater than 0", syncInterval)
	}

	downloader, err := downloading.NewDownloader(lister, client, overwrite)
	if err != nil {
		return fmt.Errorf("error creating downloader: %w", err)
	}

	logger.Log.Infof("Starting continuous album synchronization with interval: %v", syncInterval)
	logger.Log.Debugf("overwrite flag is: %v", overwrite)

	for {
		if err := syncAllAlbums(ctx, lister, downloader, albumDir); err != nil {
			logger.Log.Errorf("Error during synchronization: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(syncInterval):
		}
	}
}

func syncAllAlbums(ctx context.Context, lister listing.Lister, downloader *downloading.Downloader, albumDir string) error {
	children, err := lister.ListChildren(ctx)
	if err != nil {
		return fmt.Errorf("error listing children: %w", err)
	}

	var syncErrors []error
	for _, child := range children {
		if err := downloadAlbumsForChild(ctx, downloader, child, albumDir); err != nil {
			logger.Log.Errorf("Error downloading albums for child %s: %v", child.Name, err)
			syncErrors = append(syncErrors, err)
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("encountered %d errors during synchronization", len(syncErrors))
	}

	return nil
}

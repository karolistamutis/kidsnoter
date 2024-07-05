package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/karolistamutis/kidsnoter/config"
	"github.com/karolistamutis/kidsnoter/downloading"
	"github.com/karolistamutis/kidsnoter/listing"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/karolistamutis/kidsnoter/models"
	"github.com/spf13/cobra"
)

var downloadAlbumsCmd = &cobra.Command{
	Use:   "download-albums",
	Short: "Download the photo albums of a child",
	Long: `This command allows you to download the albums for a given child ID or name.

Specify either --child-id ID or --child-name NAME to proceed. If neither is specified, it will download albums for all children in the account.`,
	RunE: runDownloadAlbums,
}

func init() {
	addChildFlags(downloadAlbumsCmd)
	downloadAlbumsCmd.Flags().Bool("overwrite", false, "Overwrite existing files in the output directory")
	RootCmd.AddCommand(downloadAlbumsCmd)
}

func runDownloadAlbums(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	childID, childName, err := getChildFlags(cmd)
	if err != nil {
		return err
	}

	albumDir := config.GetAlbumDir()
	if albumDir == "" {
		return fmt.Errorf("missing or empty invalid album_dir setting")
	}

	overwrite, err := cmd.Flags().GetBool("overwrite")
	if err != nil {
		return fmt.Errorf("error getting overwrite flag: %w", err)
	}
	logger.Log.Debugf("overwrite flag is: %v", overwrite)

	lister := listing.NewLister(client)
	downloader, err := downloading.NewDownloader(lister, client, overwrite)
	if err != nil {
		return fmt.Errorf("error creating downloader: %w", err)
	}

	children, err := lister.ListChildren(ctx)
	if err != nil {
		return fmt.Errorf("error listing children: %w", err)
	}

	childrenToProcess, err := filterChildren(children, childID, childName)
	if err != nil {
		return err
	}

	for _, child := range childrenToProcess {
		if err := downloadAlbumsForChild(ctx, downloader, child, albumDir); err != nil {
			return err
		}
	}

	fmt.Println("Albums downloaded successfully")
	return nil
}

func downloadAlbumsForChild(ctx context.Context, downloader *downloading.Downloader, child *models.Child, outputDir string) error {
	childDir := filepath.Join(outputDir, child.Name)
	err := downloader.DownloadAlbums(ctx, child.ID, childDir)
	if err != nil {
		return fmt.Errorf("error downloading albums for %s: %w", child.Name, err)
	}
	return nil
}

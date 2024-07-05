package cmd

import (
	"context"
	"fmt"

	"github.com/karolistamutis/kidsnoter/listing"
	"github.com/karolistamutis/kidsnoter/models"
	"github.com/spf13/cobra"
)

var listAlbumsCmd = &cobra.Command{
	Use:   "list-albums",
	Short: "List the photo albums of a child",
	Long: `This command allows you to list the albums for a given child ID or name.

Specify either --child-id ID or --child-name NAME to proceed. If neither is specified, it will list albums for all children in the account.`,
	RunE: runListAlbums,
}

func init() {
	addChildFlags(listAlbumsCmd)
	RootCmd.AddCommand(listAlbumsCmd)
}

func runListAlbums(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	childID, childName, err := getChildFlags(cmd)
	if err != nil {
		return err
	}

	lister := listing.NewLister(client)

	children, err := lister.ListChildren(ctx)
	if err != nil {
		return fmt.Errorf("error listing children: %w", err)
	}

	childrenToProcess, err := filterChildren(children, childID, childName)
	if err != nil {
		return err
	}

	for _, child := range childrenToProcess {
		if err := listAlbumsForChild(ctx, lister, child); err != nil {
			return err
		}
	}

	return nil
}

func listAlbumsForChild(ctx context.Context, lister listing.Lister, child *models.Child) error {
	fmt.Printf("Listing albums for %s (ID: %d)\n", child.Name, child.ID)

	albumChan := make(chan *models.Album)
	errChan := make(chan error, 1)

	go func() {
		defer close(albumChan)
		defer close(errChan)
		if err := lister.ListAlbums(ctx, child.ID, albumChan); err != nil {
			errChan <- fmt.Errorf("error listing albums: %w", err)
		}
	}()

	albumCount := 0
	for album := range albumChan {
		fmt.Printf("ID [%d], on %s, \"%s\"\n", album.ID, album.Date, album.Title)
		albumCount++
	}

	if err := <-errChan; err != nil {
		return err
	}

	fmt.Printf("\nTotal of %d albums for %s\n\n", albumCount, child.Name)

	return nil
}

package cmd

import (
	"context"
	"fmt"

	"github.com/karolistamutis/kidsnoter/listing"
	"github.com/spf13/cobra"
)

var listChildrenCmd = &cobra.Command{
	Use:   "list-children",
	Short: "List the children under your account",
	Long:  `This command allows you to list the children associated with your account (i.e. your kids).`,
	RunE:  runListChildren,
}

func init() {
	RootCmd.AddCommand(listChildrenCmd)
}

func runListChildren(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	return listChildren(ctx)
}

func listChildren(ctx context.Context) error {
	lister := listing.NewLister(client)
	children, err := lister.ListChildren(ctx)
	if err != nil {
		return fmt.Errorf("error listing children: %w", err)
	}

	for _, child := range children {
		fmt.Printf("%s [%d], date of birth: %s, center: %d, class: %d\n",
			child.Name, child.ID, child.DateOfBirth, child.CenterID, child.ClassID)
	}

	return nil
}

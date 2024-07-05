package cmd

import (
	"fmt"

	"github.com/karolistamutis/kidsnoter/models"
	"github.com/spf13/cobra"
)

func getChildFlags(cmd *cobra.Command) (int, string, error) {
	childID, err := cmd.Flags().GetInt("child-id")
	if err != nil {
		return 0, "", fmt.Errorf("error getting child-id flag: %w", err)
	}

	childName, err := cmd.Flags().GetString("child-name")
	if err != nil {
		return 0, "", fmt.Errorf("error getting child-name flag: %w", err)
	}

	if childID != 0 && childName != "" {
		return 0, "", fmt.Errorf("set either --child-id or --child-name, not both")
	}

	return childID, childName, nil
}

func filterChildren(children []*models.Child, childID int, childName string) ([]*models.Child, error) {
	if childID == 0 && childName == "" {
		return children, nil
	}

	for _, child := range children {
		if child.ID == childID || child.Name == childName {
			return []*models.Child{child}, nil
		}
	}

	return nil, fmt.Errorf("couldn't find given child information")
}

func addChildFlags(cmd *cobra.Command) {
	cmd.Flags().Int("child-id", 0, "Child's ID on kidsnote.com, check with ./kidsnoter list-children")
	cmd.Flags().String("child-name", "", "Child's name on kidsnote.com, must match output of ./kidsnoter list-children")
}

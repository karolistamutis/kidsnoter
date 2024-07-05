package cmd

import "github.com/spf13/cobra"

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

func init() {
	completion := completionCommand()

	// Mark completion as hidden.
	completion.Hidden = true
	RootCmd.AddCommand(completion)
}

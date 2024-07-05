package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/karolistamutis/kidsnoter/config"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/spf13/cobra"
)

var Version = "0.0.1"

var (
	verbosity int
	ErrSilent = errors.New("SilentErr")
)

var RootCmd = &cobra.Command{
	Use:           "kidsnoter",
	Short:         "Kidsnoter is a tool that lets you download all your kids' albums from kidsnote.com",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runRoot,
}

func init() {
	RootCmd.SetFlagErrorFunc(flagErrorFunc)
	RootCmd.PersistentPreRunE = loginPreRun
	RootCmd.PersistentFlags().CountVarP(&verbosity, "verbosity", "v", "increase verbosity level")

	cobra.OnInitialize(initConfig)
}

func runRoot(cmd *cobra.Command, args []string) error {
	return printVersion(cmd.Context())
}

func printVersion(ctx context.Context) error {
	fmt.Printf("kidsnoter version %s\n", Version)
	return nil
}

func flagErrorFunc(cmd *cobra.Command, err error) error {
	cmd.Println(err)
	cmd.Println(cmd.UsageString())
	return ErrSilent
}

func initConfig() {
	err := config.InitConfig()
	if err != nil {
		fmt.Printf("Error initializing config: %s\n", err)
		os.Exit(1)
	}
	logger.Configure(verbosity)
}

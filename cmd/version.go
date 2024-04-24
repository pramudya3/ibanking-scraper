package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is a current application version
	Version string

	// GitCommit is a git commit SHA of the current application
	GitCommit string
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Print the version number of periksain-adapter",
		Long:         `All software has versions. This is periksain adapter`,
		Example:      `  periksain-adapter version`,
		SilenceUsage: false,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		printVersion()
	}

	return cmd
}

func printVersion() {
	if len(Version) == 0 {
		fmt.Println("Version: dev")
	} else {
		fmt.Println("Version:", Version)
	}
	fmt.Println("Git Commit:", GitCommit)
}

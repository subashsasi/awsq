package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables are set at build time using -ldflags
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of awsq",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("awsq %s (commit: %s, built: %s)\n", Version, CommitSHA, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

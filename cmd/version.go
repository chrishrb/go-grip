package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-grip",
	Long:  `Display the version information for go-grip.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (commit: %s, built: %s)\n", Version, CommitHash, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

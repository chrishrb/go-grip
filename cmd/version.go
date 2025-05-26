package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version   = "0.1.0"
	BuildDate = "2023-07-15"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-grip",
	Long:  `Display the version information for go-grip.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s (built: %s)\n", "ver", "build")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

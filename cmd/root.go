package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	theme       string
	boundingBox bool

	browser bool
	host    string
	port    int

	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "go-grip [command] <args>",
	Short: "Render markdown document as html",
	Long: `go-grip is a tool for rendering markdown documents as HTML.

Available commands:
  go-grip render FILE   - Generate static HTML from markdown
  go-grip serve FILE    - Serve markdown via local HTTP server `,

	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// no flag is used as root,
	//TODO: potential for backwards compatibility
}

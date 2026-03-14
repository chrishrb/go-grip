package cmd

import (
	"os"

	"github.com/chrishrb/go-grip/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-grip [file]",
	Short: "Render markdown document as html",
	Args:  cobra.MatchAll(cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		browser, _ := cmd.Flags().GetBool("browser")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")
		noReload, _ := cmd.Flags().GetBool("no-reload")

		var file string
		if len(args) == 1 {
			file = args[0]
		}

		parser := internal.NewParser()
		server := internal.NewServer(host, port, boundingBox, browser, !noReload, parser)
		return server.Serve(file)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("browser", "b", true, "Open new browser tab")
	rootCmd.Flags().StringP("host", "H", "localhost", "Host to use")
	rootCmd.Flags().IntP("port", "p", 6419, "Port to use")
	rootCmd.Flags().Bool("bounding-box", true, "Add bounding box to HTML")
	rootCmd.Flags().Bool("no-reload", false, "Disable automatic browser reload on file changes")
}

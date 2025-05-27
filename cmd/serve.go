package cmd

import (
	"fmt"

	"github.com/chrishrb/go-grip/pkg"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve FILE",
	Short: "Run as a server and serve the markdown file",
	Long: `Start a local server to render and serve the markdown file.

The server will watch for changes to the file and automatically refresh the browser.
This is useful for live previewing markdown as you edit it.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		parser := pkg.NewParser(theme)
		srv := pkg.NewServer(host, port, theme, boundingBox, browser, parser)

		if err := srv.Serve(file); err != nil {
			return fmt.Errorf("server error: %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&theme, "theme", "auto", "Select CSS theme [light/dark/auto]")
	serveCmd.Flags().BoolVar(&boundingBox, "bounding-box", true, "Add bounding box to HTML output")
	serveCmd.Flags().BoolVarP(&browser, "browser", "b", true, "Open browser tab automatically")
	serveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to listen on")
	serveCmd.Flags().IntVarP(&port, "port", "p", 6419, "Port to listen on")
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chrishrb/go-grip/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-grip [file]",
	Short: "Render markdown document as html",
	Args:  cobra.MatchAll(cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		theme, _ := cmd.Flags().GetString("theme")
		browser, _ := cmd.Flags().GetBool("browser")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")
		outputDir, _ := cmd.Flags().GetString("output")
		runAsServer, _ := cmd.Flags().GetBool("server")

		var file string
		if len(args) == 1 {
			file = args[0]
		}

		parser := pkg.NewParser(theme)
		srv := pkg.NewServer(host, port, theme, boundingBox, browser, parser)

		if runAsServer {
			return srv.Serve(file)
		}

		if outputDir == "" {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				return fmt.Errorf("failed to get cache directory: %v", err)
			}
			outputDir = filepath.Join(cacheDir, "go-grip")
		}

		return srv.GenerateStaticSite(file, outputDir)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("theme", "auto", "Select css theme [light/dark/auto]")
	rootCmd.Flags().BoolP("browser", "b", true, "Open new browser tab")
	rootCmd.Flags().StringP("host", "H", "localhost", "Host to use")
	rootCmd.Flags().IntP("port", "p", 6419, "Port to use")
	rootCmd.Flags().Bool("bounding-box", true, "Add bounding box to HTML")
	rootCmd.Flags().StringP("output", "o", "", "Output directory for static files (default: cache directory)")
	rootCmd.Flags().Bool("server", false, "Run as server instead of generating static files")
}

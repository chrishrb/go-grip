package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chrishrb/go-grip/pkg"
	"github.com/spf13/cobra"
)

var directoryMode bool

var renderCmd = &cobra.Command{
	Use:   "render [file|directory]",
	Short: "render md document as static html",
	Long: `Render a markdown file(s) as static HTML.

Basic usage:
  go-grip render FILE				# generate static HTML for a single file
  go-grip render FILE --output DIR	# specify output directory
  go-grip render --directory DIR	# render all markdown files in directory`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		parser := pkg.NewParser(theme)
		srv := pkg.NewServer(host, port, theme, boundingBox, browser, parser)

		if outputDir == "" {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				return fmt.Errorf("failed to get cache directory: %v", err)
			}
			outputDir = filepath.Join(cacheDir, "go-grip")
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		if directoryMode {
			return renderDirectory(srv, input, outputDir)
		} else {
			return renderSingleFile(srv, input, outputDir)
		}
	},
}

func renderSingleFile(srv *pkg.Server, filePath string, outputDir string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %s - %v", filePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("expected a file but got a directory '%s'. Use --directory flag for directories", filePath)
	}

	if filepath.Ext(filePath) != ".md" {
		return fmt.Errorf("file '%s' must be a markdown file with .md extension", filePath)
	}

	if err := srv.GenerateSingleFile(filePath, outputDir); err != nil {
		return fmt.Errorf("failed to generate HTML: %v", err)
	}

	return nil
}

func renderDirectory(srv *pkg.Server, dirPath string, outputDir string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("directory not found: %s - %v", dirPath, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("expected a directory but got a file '%s'. Remove --directory flag for single files", dirPath)
	}

	if err := srv.GenerateDirectoryFiles(dirPath, outputDir); err != nil {
		return fmt.Errorf("failed to generate HTML files: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(renderCmd)

	renderCmd.Flags().StringVar(&theme, "theme", "auto", "Select CSS theme [light/dark/auto]")
	renderCmd.Flags().BoolVar(&boundingBox, "bounding-box", true, "Add bounding box to HTML output")
	renderCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for static files")
	renderCmd.Flags().BoolVarP(&directoryMode, "directory", "d", false, "Render all markdown files in directory")
}

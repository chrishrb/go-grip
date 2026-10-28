package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/chrishrb/go-grip/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "go-grip [file]",
	Short:   "Render markdown document as html",
	Version: buildVersion(),
	Args:    cobra.MatchAll(cobra.OnlyValidArgs),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := logLevel(cmd)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		browser, _ := cmd.Flags().GetBool("browser")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")
		noReload, _ := cmd.Flags().GetBool("no-reload")
		logger, err := newLogger(cmd, os.Stderr)
		if err != nil {
			return err
		}

		var file string
		if len(args) == 1 {
			file = args[0]
		}

		parser := internal.NewParser()
		server := internal.NewServer(host, port, boundingBox, browser, !noReload, parser, logger)
		return server.Serve(file)
	},
}

// logLevel returns the validated --log-level value.
func logLevel(cmd *cobra.Command) (slog.Level, error) {
	value, _ := cmd.Flags().GetString("log-level")
	level, err := internal.ParseLogLevel(value)
	if err != nil {
		return 0, fmt.Errorf("--log-level: %w", err)
	}
	return level, nil
}

// newLogger builds the one logger the server and its watcher share.
func newLogger(cmd *cobra.Command, w io.Writer) (*slog.Logger, error) {
	level, err := logLevel(cmd)
	if err != nil {
		return nil, err
	}
	return internal.NewLogger(w, level), nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.Flags().BoolP("browser", "b", true, "Open new browser tab")
	rootCmd.Flags().StringP("host", "H", "localhost", "Host to use")
	rootCmd.Flags().IntP("port", "p", 6419, "Port to use")
	rootCmd.Flags().Bool("bounding-box", true, "Add bounding box to HTML")
	rootCmd.Flags().Bool("no-reload", false, "Disable automatic browser reload on file changes")
	rootCmd.Flags().String("log-level", "info", "Log level on stderr: error, warn, info or debug")
}

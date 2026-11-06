package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/chrishrb/go-grip/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run as an MCP server (stdio) exposing a preview_markdown tool",
	Long: `Run go-grip as a Model Context Protocol server over stdin/stdout, so AI
tools such as Claude Code can preview Markdown files in your browser.

Register it with Claude Code:

  claude mcp add --scope user go-grip -- go-grip mcp`,
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := logLevel(cmd)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		browser, _ := cmd.Flags().GetBool("browser")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")
		noReload, _ := cmd.Flags().GetBool("no-reload")
		// stdout carries the MCP protocol, so logs must stay on stderr
		logger, err := newLogger(cmd, os.Stderr)
		if err != nil {
			return err
		}

		previewer := internal.NewPreviewer(boundingBox, !noReload, internal.NewParser(), logger)
		return newMCPServer(previewer, browser).Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

// previewInput is the argument of the preview_markdown tool.
type previewInput struct {
	Path        string `json:"path" jsonschema:"path of the Markdown file or folder to preview, absolute or relative to the server's working directory"`
	OpenBrowser *bool  `json:"open_browser,omitempty" jsonschema:"open the preview in the browser (defaults to the server's --browser flag)"`
}

// newMCPServer returns the MCP server with the preview_markdown tool.
// openDefault applies when a call does not set open_browser.
func newMCPServer(previewer *internal.Previewer, openDefault bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "go-grip", Version: buildVersion()}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name: "preview_markdown",
		Description: "Render a Markdown file (or folder) with GitHub styling in the user's browser via go-grip. " +
			"Use when the user asks to preview, view or open a Markdown file such as a plan, README or design doc. " +
			"The page live-reloads when the file changes, so call this once per file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewInput) (*mcp.CallToolResult, any, error) {
		open := openDefault
		if in.OpenBrowser != nil {
			open = *in.OpenBrowser
		}
		addr, err := previewer.Preview(in.Path, open)
		if err != nil {
			// a returned error becomes a tool result with IsError set
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Previewing %s at %s (live-reloads on save)", in.Path, addr),
			}},
		}, nil, nil
	})
	return server
}

func init() {
	mcpCmd.Flags().BoolP("browser", "b", true, "Open a browser tab for each preview")
	mcpCmd.Flags().Bool("bounding-box", true, "Add bounding box to HTML")
	mcpCmd.Flags().Bool("no-reload", false, "Disable automatic browser reload on file changes")
	mcpCmd.Flags().String("log-level", "info", "Log level on stderr: error, warn, info or debug")
	rootCmd.AddCommand(mcpCmd)
}

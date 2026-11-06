package cmd

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/chrishrb/go-grip/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectMCP(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := newMCPServer(internal.NewPreviewer(false, false, internal.NewParser(), nil), false)
	serverT, clientT := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverT, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestMCPListsPreviewTool(t *testing.T) {
	session := connectMCP(t)
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(res.Tools) != 1 || res.Tools[0].Name != "preview_markdown" {
		t.Fatalf("expected only preview_markdown, got %+v", res.Tools)
	}
}

func TestMCPPreviewMarkdown(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(file, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	session := connectMCP(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "preview_markdown",
		Arguments: map[string]any{"path": file, "open_browser": false},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %+v", res.Content)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	addr := regexp.MustCompile(`http://\S+`).FindString(text)
	if addr == "" {
		t.Fatalf("no URL in %q", text)
	}
	resp, err := http.Get(addr)
	if err != nil {
		t.Fatalf("GET %s: %v", addr, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", addr, resp.StatusCode)
	}
}

func TestMCPPreviewMissingPathIsToolError(t *testing.T) {
	session := connectMCP(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "preview_markdown",
		Arguments: map[string]any{"path": filepath.Join(t.TempDir(), "nope.md")},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected a tool error, got %+v", res.Content)
	}
}

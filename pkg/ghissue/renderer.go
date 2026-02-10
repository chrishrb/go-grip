package ghissue

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// HTMLRenderer renders GitHubIssue nodes as HTML links
type HTMLRenderer struct {
	config *Config
}

// NewHTMLRenderer creates a new HTML renderer for GitHub issue/PR references
func NewHTMLRenderer(config *Config) renderer.NodeRenderer {
	return &HTMLRenderer{
		config: config,
	}
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGitHubIssue, r.renderGitHubIssue)
}

func (r *HTMLRenderer) renderGitHubIssue(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {

	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*GitHubIssue)

	// If no repository is set, just render as plain text
	if len(n.Repository) == 0 {
		_, _ = w.WriteString("#")
		_, _ = w.Write(n.Number)
		return ast.WalkContinue, nil
	}

	repository := string(n.Repository)
	number := string(n.Number)

	// Parse owner and repo
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		_, _ = w.WriteString("#")
		_, _ = w.Write(n.Number)
		return ast.WalkContinue, nil
	}

	owner := parts[0]
	repo := parts[1]

	// Build the URL (same for issues and prs)
	url := fmt.Sprintf("https://github.com/%s/%s/issues/%s", owner, repo, number)

	// Render as a link
	_, _ = w.WriteString(`<a href="`)
	_, _ = w.WriteString(url)
	_, _ = w.WriteString(`" class="issue-link">`)

	// Display text - show full repo#number for external refs, just #number for internal
	if n.IsExternal {
		// External reference: show owner/repo#123
		_, _ = w.Write(n.Repository)
		_, _ = w.WriteString("#")
		_, _ = w.Write(n.Number)
	} else {
		// Internal reference: show #123
		_, _ = w.WriteString("#")
		_, _ = w.Write(n.Number)
	}

	_, _ = w.WriteString(`</a>`)

	return ast.WalkContinue, nil
}

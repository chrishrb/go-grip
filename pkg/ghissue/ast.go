package ghissue

import (
	"github.com/yuin/goldmark/ast"
)

// KindGitHubIssue is a NodeKind of the GitHubIssue node.
var KindGitHubIssue = ast.NewNodeKind("GitHubIssue")

// GitHubIssue represents a GitHub issue or PR reference in markdown
type GitHubIssue struct {
	ast.BaseInline

	// Repository in the format "owner/repo"
	// Empty string means use the default repository
	Repository []byte

	// Number is the issue or PR number
	Number []byte

	// IsExternal indicates if this was an external reference (owner/repo#123)
	// vs an internal reference (#123)
	IsExternal bool
}

// Dump implements Node.Dump
func (n *GitHubIssue) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements Node.Kind
func (n *GitHubIssue) Kind() ast.NodeKind {
	return KindGitHubIssue
}

// NewGitHubIssue creates a new GitHubIssue node
func NewGitHubIssue(repo, number []byte) *GitHubIssue {
	return &GitHubIssue{
		Repository: repo,
		Number:     number,
		IsExternal: false, // Default to internal
	}
}

// NewExternalGitHubIssue creates a new GitHubIssue node for external references
func NewExternalGitHubIssue(repo, number []byte) *GitHubIssue {
	return &GitHubIssue{
		Repository: repo,
		Number:     number,
		IsExternal: true,
	}
}

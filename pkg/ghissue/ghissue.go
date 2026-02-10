// Package ghissue provides a Goldmark extension for GitHub issue and PR references.
// It transforms references like #123 or owner/repo#123 into clickable links that point
// to the actual issue or pull request on GitHub.
package ghissue

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Config holds configuration for the GitHub issue extension
type Config struct {
	// Repository is the default repository in the format "owner/repo"
	// Used for resolving references like #123
	Repository string

	// GitHubToken is an optional GitHub personal access token for API calls
	// If not provided, unauthenticated requests will be made (rate limited)
	GitHubToken string
}

// Option is a functional option for configuring the extension
type Option func(*Config)

// WithRepository sets the default repository for issue/PR references
func WithRepository(repo string) Option {
	return func(c *Config) {
		c.Repository = repo
	}
}

// WithGitHubToken sets the GitHub API token for authenticated requests
func WithGitHubToken(token string) Option {
	return func(c *Config) {
		c.GitHubToken = token
	}
}

// Extender implements goldmark.Extender to add GitHub issue/PR reference support
type Extender struct {
	config Config
}

// Extend extends the Goldmark parser and renderer with GitHub issue/PR functionality
func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			// Use high priority to run before hashtag extension
			util.Prioritized(NewParser(&e.config), 500),
		),
		parser.WithASTTransformers(
			// Transformer handles external references that parser skips
			util.Prioritized(NewTransformer(&e.config), 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(&e.config), 100),
		),
	)
}

// New creates a new GitHub issue/PR Extender with the given options
func New(opts ...Option) *Extender {
	config := Config{}
	for _, opt := range opts {
		opt(&config)
	}

	// Auto-detect repository if not provided
	if config.Repository == "" {
		config.Repository = DetectRepository()
	}

	return &Extender{config: config}
}

package internal

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/chrishrb/go-grip/pkg/alert"
	"github.com/chrishrb/go-grip/pkg/details"
	"github.com/chrishrb/go-grip/pkg/footnote"
	"github.com/chrishrb/go-grip/pkg/frontmatter"
	"github.com/chrishrb/go-grip/pkg/ghissue"
	"github.com/chrishrb/go-grip/pkg/highlighting"
	"github.com/chrishrb/go-grip/pkg/mathjax"
	"github.com/chrishrb/go-grip/pkg/slug"
	"github.com/chrishrb/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
)

const (
	FeatureDetails  = "details"
	FeatureFootnote = "footnote"
	FeatureGHIssue  = "ghissue"
	FeatureMathJax  = "mathjax"
	FeatureMermaid  = "mermaid"
)

var supportedMarkdownFeatures = []string{
	FeatureDetails,
	FeatureFootnote,
	FeatureGHIssue,
	FeatureMathJax,
	FeatureMermaid,
}

type Parser struct {
	disabledFeatures map[string]struct{}
}

func NewParser(disabledFeatures []string) *Parser {
	// Normalize feature names once so the rest of the parser can use simple lookups.
	disabled := make(map[string]struct{}, len(disabledFeatures))
	for _, feature := range disabledFeatures {
		disabled[strings.ToLower(feature)] = struct{}{}
	}
	return &Parser{disabledFeatures: disabled}
}

func SupportedMarkdownFeatures() []string {
	return slices.Clone(supportedMarkdownFeatures)
}

func ValidateMarkdownFeatures(features []string) error {
	var invalid []string
	for _, feature := range features {
		feature = strings.ToLower(strings.TrimSpace(feature))
		if !slices.Contains(supportedMarkdownFeatures, feature) {
			invalid = append(invalid, feature)
		}
	}
	if len(invalid) == 0 {
		return nil
	}

	return fmt.Errorf(
		"unknown markdown feature(s): %s (supported: %s)",
		strings.Join(invalid, ", "),
		strings.Join(supportedMarkdownFeatures, ", "),
	)
}

func (m Parser) FeatureEnabled(name string) bool {
	_, disabled := m.disabledFeatures[strings.ToLower(name)]
	return !disabled
}

func (m Parser) TemplateFeatures() map[string]bool {
	// Keep template asset toggles derived from the same feature registry as parsing.
	features := make(map[string]bool, len(supportedMarkdownFeatures))
	for _, feature := range supportedMarkdownFeatures {
		features[feature] = m.FeatureEnabled(feature)
	}
	return features
}

func (m Parser) MdToHTML(input []byte) ([]byte, error) {
	var prefix []byte
	if fm, body, ok := frontmatter.Extract(input); ok {
		if table, err := frontmatter.RenderTable(fm); err == nil {
			prefix = table
			input = body
		}
	}

	// Always-on GitHub-style extensions stay in the base list; optional ones are gated below.
	extensions := []goldmark.Extender{
		extension.Linkify,
		extension.Table,
		extension.Strikethrough,
		tasklist.TaskList,
		emoji.Emoji,
		&hashtag.Extender{},
		alert.New(),
		highlighting.Highlighting,
	}
	if m.FeatureEnabled(FeatureFootnote) {
		extensions = append(extensions, footnote.Footnote)
	}
	if m.FeatureEnabled(FeatureMermaid) {
		extensions = append(extensions, &mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true})
	}
	if m.FeatureEnabled(FeatureMathJax) {
		extensions = append(extensions, mathjax.MathJax)
	}
	if m.FeatureEnabled(FeatureGHIssue) {
		extensions = append(extensions, ghissue.New())
	}
	if m.FeatureEnabled(FeatureDetails) {
		extensions = append(extensions, details.New())
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	// Generate heading ids like GitHub does (e.g. keep underscores), so that
	// anchor links written against GitHub's ids resolve here too.
	ctx := parser.NewContext(parser.WithIDs(slug.NewIDs()))

	var buf bytes.Buffer
	if err := md.Convert(input, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}
	return append(prefix, buf.Bytes()...), nil
}

// LocalRefs returns the local image destinations referenced by the Markdown
// input, unescaped, without query or fragment, deduplicated and in document
// order. Destinations with a scheme (https:, data:, ...) and protocol-relative
// ones (//host/...) are skipped. Images written as raw HTML are not reported.
func (m Parser) LocalRefs(input []byte) []string {
	if _, body, ok := frontmatter.Extract(input); ok {
		input = body
	}

	md := goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough))
	doc := md.Parser().Parse(text.NewReader(input))

	refs := []string{}
	seen := make(map[string]bool)
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		img, ok := n.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}
		if ref, ok := localRef(string(img.Destination)); ok && !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
		return ast.WalkContinue, nil
	})
	return refs
}

// localRef reports the unescaped path of an image destination that points at
// a local file.
func localRef(dest string) (string, bool) {
	if dest == "" || strings.HasPrefix(dest, "//") {
		return "", false
	}
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" {
		return "", false
	}
	return u.Path, true
}

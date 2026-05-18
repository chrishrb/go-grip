package internal

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/chrishrb/go-grip/pkg/alert"
	"github.com/chrishrb/go-grip/pkg/details"
	"github.com/chrishrb/go-grip/pkg/footnote"
	"github.com/chrishrb/go-grip/pkg/ghissue"
	"github.com/chrishrb/go-grip/pkg/highlighting"
	"github.com/chrishrb/go-grip/pkg/mathjax"
	"github.com/chrishrb/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
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
	var buf bytes.Buffer
	if err := md.Convert(input, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

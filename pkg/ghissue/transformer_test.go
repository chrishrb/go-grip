package ghissue

import (
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		config         *Config
		expectedIssues []expectedIssue
	}{
		{
			name:  "internal reference",
			input: "See #123 for details",
			config: &Config{
				Repository: "owner/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "owner/repo",
					number:     "123",
					isExternal: false,
				},
			},
		},
		{
			name:  "external reference",
			input: "Check kubernetes/kubernetes#456",
			config: &Config{
				Repository: "owner/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "kubernetes/kubernetes",
					number:     "456",
					isExternal: true,
				},
			},
		},
		{
			name:  "multiple internal references",
			input: "See #100 and #200",
			config: &Config{
				Repository: "test/project",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "test/project",
					number:     "100",
					isExternal: false,
				},
				{
					repository: "test/project",
					number:     "200",
					isExternal: false,
				},
			},
		},
		{
			name:  "mixed references",
			input: "See #100 and owner/repo#200",
			config: &Config{
				Repository: "default/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "default/repo",
					number:     "100",
					isExternal: false,
				},
				{
					repository: "owner/repo",
					number:     "200",
					isExternal: true,
				},
			},
		},
		{
			name:  "internal reference at start",
			input: "#999 is important",
			config: &Config{
				Repository: "my/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "my/repo",
					number:     "999",
					isExternal: false,
				},
			},
		},
		{
			name:  "internal reference at end",
			input: "Fixed in #777",
			config: &Config{
				Repository: "my/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "my/repo",
					number:     "777",
					isExternal: false,
				},
			},
		},
		{
			name:  "no references",
			input: "Just regular text",
			config: &Config{
				Repository: "owner/repo",
			},
			expectedIssues: []expectedIssue{},
		},
		{
			name:  "reference without config repository",
			input: "See #123",
			config: &Config{
				Repository: "",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "",
					number:     "123",
					isExternal: false,
				},
			},
		},
		{
			name:  "hyphenated owner and repo",
			input: "See my-org/my-repo#42",
			config: &Config{
				Repository: "default/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "my-org/my-repo",
					number:     "42",
					isExternal: true,
				},
			},
		},
		{
			name:  "reference in middle of sentence",
			input: "The issue #555 was fixed yesterday",
			config: &Config{
				Repository: "test/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "test/repo",
					number:     "555",
					isExternal: false,
				},
			},
		},
		{
			name:  "multiple consecutive references",
			input: "#1 #2 #3",
			config: &Config{
				Repository: "test/repo",
			},
			expectedIssues: []expectedIssue{
				{
					repository: "test/repo",
					number:     "1",
					isExternal: false,
				},
				{
					repository: "test/repo",
					number:     "2",
					isExternal: false,
				},
				{
					repository: "test/repo",
					number:     "3",
					isExternal: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple parser and parse the input
			p := parser.NewParser(
				parser.WithBlockParsers(parser.DefaultBlockParsers()...),
				parser.WithInlineParsers(parser.DefaultInlineParsers()...),
				parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
			)
			source := []byte(tt.input)
			reader := text.NewReader(source)
			doc := p.Parse(reader).(*ast.Document)

			// Create and apply the transformer
			transformer := NewTransformer(tt.config)
			transformer.Transform(doc, reader, parser.NewContext())

			// Walk the AST and collect all GitHubIssue nodes
			var foundIssues []expectedIssue
			_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering {
					return ast.WalkContinue, nil
				}

				if n.Kind() == KindGitHubIssue {
					issue := n.(*GitHubIssue)
					foundIssues = append(foundIssues, expectedIssue{
						repository: string(issue.Repository),
						number:     string(issue.Number),
						isExternal: issue.IsExternal,
					})
				}

				return ast.WalkContinue, nil
			})

			// Verify we found the expected number of issues
			if len(foundIssues) != len(tt.expectedIssues) {
				t.Errorf("Expected %d issues, found %d", len(tt.expectedIssues), len(foundIssues))
				for i, issue := range foundIssues {
					t.Logf("Found issue %d: repo=%q number=%q isExternal=%v",
						i, issue.repository, issue.number, issue.isExternal)
				}
				return
			}

			// Verify each issue matches
			for i, expected := range tt.expectedIssues {
				found := foundIssues[i]
				if found.repository != expected.repository {
					t.Errorf("Issue %d: expected repository %q, got %q",
						i, expected.repository, found.repository)
				}
				if found.number != expected.number {
					t.Errorf("Issue %d: expected number %q, got %q",
						i, expected.number, found.number)
				}
				if found.isExternal != expected.isExternal {
					t.Errorf("Issue %d: expected isExternal %v, got %v",
						i, expected.isExternal, found.isExternal)
				}
			}
		})
	}
}

func TestTransformerNoConfigRepository(t *testing.T) {
	input := "See #123"
	source := []byte(input)

	p := parser.NewParser(
		parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
	reader := text.NewReader(source)
	doc := p.Parse(reader).(*ast.Document)

	// Test with nil config
	transformer := NewTransformer(nil)
	transformer.Transform(doc, reader, parser.NewContext())

	// Walk and find issues
	var foundIssues []*GitHubIssue
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindGitHubIssue {
			foundIssues = append(foundIssues, n.(*GitHubIssue))
		}
		return ast.WalkContinue, nil
	})

	if len(foundIssues) != 1 {
		t.Errorf("Expected 1 issue with nil config, found %d", len(foundIssues))
		return
	}

	if string(foundIssues[0].Number) != "123" {
		t.Errorf("Expected number '123', got %q", string(foundIssues[0].Number))
	}
	if len(foundIssues[0].Repository) != 0 {
		t.Errorf("Expected empty repository, got %q", string(foundIssues[0].Repository))
	}
}

func TestTransformerSkipsCodeSpan(t *testing.T) {
	input := "`#123`"
	source := []byte(input)

	p := parser.NewParser(
		parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
	reader := text.NewReader(source)
	doc := p.Parse(reader).(*ast.Document)

	config := &Config{Repository: "owner/repo"}
	transformer := NewTransformer(config)
	transformer.Transform(doc, reader, parser.NewContext())

	// Walk and find issues - should find none inside code span
	var foundIssues []*GitHubIssue
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindGitHubIssue {
			foundIssues = append(foundIssues, n.(*GitHubIssue))
		}
		return ast.WalkContinue, nil
	})

	if len(foundIssues) != 0 {
		t.Errorf("Expected 0 issues inside code span, found %d", len(foundIssues))
	}
}

func TestTransformerSkipsLinks(t *testing.T) {
	input := "[#123](http://example.com)"
	source := []byte(input)

	p := parser.NewParser()
	reader := text.NewReader(source)
	doc := p.Parse(reader).(*ast.Document)

	config := &Config{Repository: "owner/repo"}
	transformer := NewTransformer(config)
	transformer.Transform(doc, reader, parser.NewContext())

	// Walk and find issues - should find none inside links
	var foundIssues []*GitHubIssue
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindGitHubIssue {
			foundIssues = append(foundIssues, n.(*GitHubIssue))
		}
		return ast.WalkContinue, nil
	})

	if len(foundIssues) != 0 {
		t.Errorf("Expected 0 issues inside link, found %d", len(foundIssues))
	}
}

type expectedIssue struct {
	repository string
	number     string
	isExternal bool
}

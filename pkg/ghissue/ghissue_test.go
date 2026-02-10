package ghissue

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestGitHubIssueExtension(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		repo              string
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:  "internal issue reference",
			input: "See #123 for details",
			repo:  "owner/repo",
			expectContains: []string{
				`<a href="https://github.com/owner/repo/issues/123" class="issue-link">#123</a>`,
			},
		},
		{
			name:  "external issue reference",
			input: "Check out grafana/grafana#10",
			repo:  "owner/repo",
			expectContains: []string{
				`<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`,
			},
			expectNotContains: []string{
				"grafana/grafana<a", // Repo name should be inside the link
			},
		},
		{
			name:  "multiple internal references",
			input: "See #100 and #200",
			repo:  "test/project",
			expectContains: []string{
				`#100</a>`,
				`#200</a>`,
				` and `, // Text between links preserved
			},
		},
		{
			name:  "mixed internal and external references",
			input: "See #100 and grafana/grafana#10",
			repo:  "chrishrb/go-grip",
			expectContains: []string{
				`<a href="https://github.com/chrishrb/go-grip/issues/100" class="issue-link">#100</a>`,
				`<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`,
			},
			expectNotContains: []string{
				"grafana/grafana<a",
				`href="https://github.com/chrishrb/go-grip/issues/10"`, // Should use correct repo
			},
		},
		{
			name:  "multiple external references",
			input: "kubernetes/kubernetes#200 and grafana/grafana#10",
			repo:  "owner/repo",
			expectContains: []string{
				`kubernetes/kubernetes#200</a>`,
				`grafana/grafana#10</a>`,
			},
		},
		{
			name:  "no repository configured",
			input: "See #123",
			repo:  "",
			expectContains: []string{
				"#123", // Should render as plain text or without repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(WithRepository(tt.repo)),
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()

			for _, expected := range tt.expectContains {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if bytes.Contains([]byte(output), []byte(notExpected)) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestComplexScenarios(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		repo              string
		expectContains    []string
		expectNotContains []string
	}{
		{
			name: "full user scenario",
			input: `See #100 and #120

Same goes for external repositories. e.g. grafana/grafana#10 should lead to the grafana/grafana repository.

kubernetes/kubernetes#200`,
			repo: "chrishrb/go-grip",
			expectContains: []string{
				`<a href="https://github.com/chrishrb/go-grip/issues/100" class="issue-link">#100</a>`,
				`<a href="https://github.com/chrishrb/go-grip/issues/120" class="issue-link">#120</a>`,
				`<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`,
				`<a href="https://github.com/kubernetes/kubernetes/issues/200" class="issue-link">kubernetes/kubernetes#200</a>`,
			},
			expectNotContains: []string{
				"grafana/grafana<a",
				"kubernetes/kubernetes<a",
				`href="https://github.com/chrishrb/go-grip/issues/10"`,
				`href="https://github.com/chrishrb/go-grip/issues/200"`,
			},
		},
		{
			name:  "hyphenated owner and repo",
			input: "See my-org/my-repo#42",
			repo:  "default/repo",
			expectContains: []string{
				`<a href="https://github.com/my-org/my-repo/issues/42" class="issue-link">my-org/my-repo#42</a>`,
			},
		},
		{
			name:  "consecutive references",
			input: "#1 #2 #3",
			repo:  "test/repo",
			expectContains: []string{
				`#1</a>`,
				`#2</a>`,
				`#3</a>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(WithRepository(tt.repo)),
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()

			for _, expected := range tt.expectContains {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if bytes.Contains([]byte(output), []byte(notExpected)) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		repo  string
	}{
		{
			name:  "code span should not parse",
			input: "`#123`",
			repo:  "owner/repo",
		},
		{
			name:  "link should not parse",
			input: "[#123](http://example.com)",
			repo:  "owner/repo",
		},
		{
			name:  "plain text without references",
			input: "Just regular text",
			repo:  "owner/repo",
		},
		{
			name:  "invalid format with space",
			input: "# 123",
			repo:  "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(WithRepository(tt.repo)),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			// Just verify it doesn't crash
			if buf.Len() == 0 {
				t.Error("Expected some output")
			}
		})
	}
}

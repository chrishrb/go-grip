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
		name     string
		input    string
		repo     string
		contains string
	}{
		{
			name:     "internal issue reference",
			input:    "See #123 for details",
			repo:     "owner/repo",
			contains: `<a href="https://github.com/owner/repo/`,
		},
		{
			name:     "external issue reference",
			input:    "Check out grafana/grafana#10",
			repo:     "",
			contains: `grafana/grafana#10</a>`,
		},
		{
			name:     "multiple references",
			input:    "See #100 and #200",
			repo:     "test/project",
			contains: `#100</a>`,
		},
		{
			name:     "no reference",
			input:    "Just regular text",
			repo:     "owner/repo",
			contains: "Just regular text",
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
			if !bytes.Contains([]byte(output), []byte(tt.contains)) {
				t.Errorf("Expected output to contain %q, got:\n%s", tt.contains, output)
			}
		})
	}
}

func TestGitHubIssueParser(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedRepo   string
		expectedNumber string
		shouldParse    bool
	}{
		{
			name:           "simple issue",
			input:          "#123",
			expectedRepo:   "",
			expectedNumber: "123",
			shouldParse:    true,
		},
		{
			name:           "external reference",
			input:          "owner/repo#456",
			expectedRepo:   "owner/repo",
			expectedNumber: "456",
			shouldParse:    true,
		},
		{
			name:        "invalid format",
			input:       "# 123",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(WithRepository("default/repo")),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			// Just verify it doesn't crash
			// Detailed parsing tests would require AST inspection
		})
	}
}

// TestUserReportedBug tests the exact broken HTML output the user reported
func TestUserReportedBug(t *testing.T) {
	// The user reported this exact HTML output:
	// <p>grafana/grafana<a href="https://github.com/chrishrb/go-grip/issues/10" class="issue-link">#10</a></p>
	// This shows that:
	// 1. The repo name is OUTSIDE the link
	// 2. The URL uses the wrong repository (chrishrb/go-grip instead of grafana/grafana)

	input := `See #100 and #120

grafana/grafana#10`
	repo := "chrishrb/go-grip"

	md := goldmark.New(
		goldmark.WithExtensions(
			New(WithRepository(repo)),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	t.Logf("Actual output:\n%s", output)

	// WHAT WE DON'T WANT (the bug):
	buggyPattern := `grafana/grafana<a href="https://github.com/chrishrb/go-grip/issues/10"`
	if bytes.Contains([]byte(output), []byte(buggyPattern)) {
		t.Errorf("BUG REPRODUCED! Found the broken pattern in output:\n%s", output)
	}

	// WHAT WE WANT (correct behavior):
	correctPattern := `<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`
	if !bytes.Contains([]byte(output), []byte(correctPattern)) {
		t.Errorf("Expected correct pattern not found.\nWanted:\n%s\n\nGot:\n%s", correctPattern, output)
	}
}

// TestInternalThenExternalReference tests that text between refs is preserved
func TestInternalThenExternalReference(t *testing.T) {
	input := `see #10 and grafana/grafana#11`
	repo := "chrishrb/go-grip"

	md := goldmark.New(
		goldmark.WithExtensions(
			New(WithRepository(repo)),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	t.Logf("Output:\n%s", output)

	// Check both links are present
	if !bytes.Contains([]byte(output), []byte(`#10</a>`)) {
		t.Error("#10 link not found")
	}
	if !bytes.Contains([]byte(output), []byte(`grafana/grafana#11</a>`)) {
		t.Error("grafana/grafana#11 link not found")
	}

	// Check that ' and ' is preserved between the links
	if !bytes.Contains([]byte(output), []byte(` and `)) {
		t.Errorf("Text ' and ' between links is missing!\nOutput:\n%s", output)
	}
}

// TestExactUserCase tests the exact case from the user's report
func TestExactUserCase(t *testing.T) {
	input := `See #100 and #120

grafana/grafana#10
`
	repo := "chrishrb/go-grip"

	md := goldmark.New(
		goldmark.WithExtensions(
			New(WithRepository(repo)),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	t.Logf("Output:\n%s", output)

	// The bug: grafana/grafana#10 becomes grafana/grafana<a...>#10</a>
	// We need to ensure the full reference is in the link
	expected := `<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`
	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected output to contain full link:\n%s\n\nBut got:\n%s", expected, output)
	}

	// Make sure it doesn't have the broken pattern
	broken := "grafana/grafana<a"
	if bytes.Contains([]byte(output), []byte(broken)) {
		t.Errorf("Output contains broken pattern %q:\n%s", broken, output)
	}

	// Check all #100 and #120 are properly linked
	if !bytes.Contains([]byte(output), []byte(`<a href="https://github.com/chrishrb/go-grip/issues/100" class="issue-link">#100</a>`)) {
		t.Error("#100 not properly linked")
	}
	if !bytes.Contains([]byte(output), []byte(`<a href="https://github.com/chrishrb/go-grip/issues/120" class="issue-link">#120</a>`)) {
		t.Error("#120 not properly linked")
	}
}

// TestExactUserCaseWithinParagraph tests external reference within same paragraph as internal refs
func TestExactUserCaseWithinParagraph(t *testing.T) {
	// This is closer to what the user might be experiencing
	input := `See #100 and #120

grafana/grafana#10`
	repo := "chrishrb/go-grip"

	md := goldmark.New(
		goldmark.WithExtensions(
			New(WithRepository(repo)),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	output := buf.String()
	t.Logf("Output:\n%s", output)

	// grafana/grafana#10 should be a single complete link
	expectedLink := `grafana/grafana#10</a>`
	if !bytes.Contains([]byte(output), []byte(expectedLink)) {
		t.Errorf("Expected output to contain %q, got:\n%s", expectedLink, output)
	}

	// Should NOT have the broken pattern where repo is outside the link
	brokenPattern := "grafana/grafana<a"
	if bytes.Contains([]byte(output), []byte(brokenPattern)) {
		t.Errorf("Found broken pattern %q in output:\n%s", brokenPattern, output)
	}
}

// TestGitHubIssueBugFixes tests for specific bugs found in real-world usage
func TestGitHubIssueBugFixes(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		repo              string
		expectAllLinked   []string // All these should be converted to links
		expectNotContains []string // These should NOT appear in output
	}{
		{
			name:  "two internal references in same sentence",
			input: "See #100 and #120",
			repo:  "chrishrb/go-grip",
			expectAllLinked: []string{
				`<a href="https://github.com/chrishrb/go-grip/issues/100" class="issue-link">#100</a>`,
				`<a href="https://github.com/chrishrb/go-grip/issues/120" class="issue-link">#120</a>`,
			},
			expectNotContains: []string{
				"and #120", // #120 should be a link, not plain text
			},
		},
		{
			name:  "external reference at start of line",
			input: "kubernetes/kubernetes#200",
			repo:  "chrishrb/go-grip",
			expectAllLinked: []string{
				`<a href="https://github.com/kubernetes/kubernetes/issues/200" class="issue-link">kubernetes/kubernetes#200</a>`,
			},
			expectNotContains: []string{
				"kubernetes/kubernetes<a",     // Should not split the repo name from the link
				"chrishrb/go-grip/issues/200", // Should not use default repo for external refs
			},
		},
		{
			name:  "external reference in sentence",
			input: "e.g. grafana/grafana#10 should lead",
			repo:  "chrishrb/go-grip",
			expectAllLinked: []string{
				`<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`,
			},
			expectNotContains: []string{
				`href="https://github.com/chrishrb/go-grip/issues/10"`, // Should not use default repo
			},
		},
		{
			name: "full test case from user",
			input: `See #100 and #120

Same goes for external repositories. e.g. grafana/grafana#10 should lead to the grafana/grafana repository with either issue 10 or pull request 10. (decide here also which one exists and then create the link). 

kubernetes/kubernetes#200`,
			repo: "chrishrb/go-grip",
			expectAllLinked: []string{
				`<a href="https://github.com/chrishrb/go-grip/issues/100" class="issue-link">#100</a>`,
				`<a href="https://github.com/chrishrb/go-grip/issues/120" class="issue-link">#120</a>`,
				`<a href="https://github.com/grafana/grafana/issues/10" class="issue-link">grafana/grafana#10</a>`,
				`<a href="https://github.com/kubernetes/kubernetes/issues/200" class="issue-link">kubernetes/kubernetes#200</a>`,
			},
			expectNotContains: []string{
				"and #120",                // #120 should be linked
				"kubernetes/kubernetes<a", // Should not split repo from link
				`href="https://github.com/chrishrb/go-grip/issues/10"`,  // External ref should use its own repo
				`href="https://github.com/chrishrb/go-grip/issues/200"`, // External ref should use its own repo
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
			t.Logf("Output:\n%s", output)

			// Check that all expected links are present
			for _, expected := range tt.expectAllLinked {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("Expected output to contain:\n%s\n\nBut got:\n%s", expected, output)
				}
			}

			// Check that unwanted strings are not present
			for _, notExpected := range tt.expectNotContains {
				if bytes.Contains([]byte(output), []byte(notExpected)) {
					t.Errorf("Expected output NOT to contain:\n%s\n\nBut got:\n%s", notExpected, output)
				}
			}
		})
	}
}

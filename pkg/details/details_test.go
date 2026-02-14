package details

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestDetailsExtension(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		contains []string
	}{
		{
			name: "basic details with summary",
			markdown: `<details>

<summary><h3>YAML</h3></summary>

**bold text**

</details>`,
			contains: []string{
				`<details id="details-`,
				`<summary><h3>YAML</h3></summary>`,
				`<strong>bold text</strong>`,
				`</details>`,
				`sessionStorage`,
			},
		},
		{
			name: "multiple details elements",
			markdown: `<details>
<summary>First</summary>
Content 1
</details>

<details>
<summary>Second</summary>
Content 2
</details>`,
			contains: []string{
				`<details id="details-1-`,
				`<details id="details-2-`,
				`<summary>First</summary>`,
				`<summary>Second</summary>`,
			},
		},
		{
			name: "details with markdown content",
			markdown: `<details>
<summary>Click to expand</summary>

This is **bold** and this is *italic*.

- List item 1
- List item 2

</details>`,
			contains: []string{
				`<details id="details-`,
				`<summary>Click to expand</summary>`,
				`<strong>bold</strong>`,
				`<em>italic</em>`,
				`<li>List item 1</li>`,
				`<li>List item 2</li>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					New(),
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithHardWraps(),
					html.WithXHTML(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.markdown), &buf); err != nil {
				t.Fatal(err)
			}

			output := buf.String()

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
				}
			}
		})
	}
}

func TestDetailsIDGeneration(t *testing.T) {
	markdown := `<details>
<summary>Test</summary>
Content
</details>`

	md := goldmark.New(
		goldmark.WithExtensions(
			New(),
		),
	)

	var buf1, buf2 bytes.Buffer

	// Render twice
	if err := md.Convert([]byte(markdown), &buf1); err != nil {
		t.Fatal(err)
	}

	// Note: We need a new markdown instance because the transformer maintains state
	md2 := goldmark.New(
		goldmark.WithExtensions(
			New(),
		),
	)

	if err := md2.Convert([]byte(markdown), &buf2); err != nil {
		t.Fatal(err)
	}

	output1 := buf1.String()
	output2 := buf2.String()

	// Both should contain an ID
	if !strings.Contains(output1, `id="details-`) {
		t.Error("First render should contain an ID")
	}

	if !strings.Contains(output2, `id="details-`) {
		t.Error("Second render should contain an ID")
	}

	// Script should only be rendered once per document
	scriptCount := strings.Count(output1, "<script>")
	if scriptCount != 1 {
		t.Errorf("Expected exactly 1 script tag, got %d", scriptCount)
	}
}

func TestCustomIDPrefix(t *testing.T) {
	markdown := `<details>
<summary>Test</summary>
Content
</details>`

	md := goldmark.New(
		goldmark.WithExtensions(
			NewWithPrefix("custom-"),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, `id="custom-`) {
		t.Errorf("Expected custom prefix, got:\n%s", output)
	}
}

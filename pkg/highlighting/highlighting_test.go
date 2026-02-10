package highlighting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestHighlighting(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======
`+"``` go\n"+`func main() {
    fmt.Println("ok")
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	// The output should contain the highlighted code with clipboard button
	output := buffer.String()
	if !strings.Contains(output, "<h1>Title</h1>") {
		t.Error("failed to render title")
	}
	if !strings.Contains(output, `<div class="highlight notranslate position-relative overflow-auto"">`) {
		t.Error("failed to render highlight div")
	}
	if !strings.Contains(output, "clipboard-copy") {
		t.Error("failed to render clipboard button")
	}
	if !strings.Contains(output, "func") {
		t.Error("failed to render code content")
	}
}

func TestHighlightingWithoutLanguage(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Highlighting,
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======
`+"```"+`
func main() {
    fmt.Println("ok")
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	// Code without language is treated as plaintext and should still be highlighted
	output := buffer.String()
	if !strings.Contains(output, "<h1>Title</h1>") {
		t.Error("failed to render title")
	}
	if !strings.Contains(output, `<div class="highlight notranslate position-relative overflow-auto"">`) {
		t.Error("failed to render highlight div")
	}
	if !strings.Contains(output, "func main()") {
		t.Error("failed to render code content")
	}
}

func TestHighlightingCpp(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Highlighting,
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======

`+"```"+`cpp
#include <iostream>
int main() {
    std::cout<< "hello" << std::endl;
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	output := buffer.String()
	if !strings.Contains(output, "<h1>Title</h1>") {
		t.Error("failed to render title")
	}
	if !strings.Contains(output, `<div class="highlight notranslate position-relative overflow-auto"">`) {
		t.Error("failed to render highlight div")
	}
	if !strings.Contains(output, "#include") {
		t.Error("failed to render code content")
	}
	if !strings.Contains(output, "clipboard-copy") {
		t.Error("failed to render clipboard button")
	}
}

func TestHighlightingHttp(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte("```http"+`
GET /foo HTTP/1.1
Content-Type: application/json
User-Agent: foo

{
  "hello": "world"
}
`+"```"), &buffer); err != nil {
		t.Fatal(err)
	}

	output := buffer.String()
	if !strings.Contains(output, `<div class="highlight notranslate position-relative overflow-auto"">`) {
		t.Error("failed to render highlight div")
	}
	if !strings.Contains(output, "GET") {
		t.Error("failed to render HTTP method")
	}
	if !strings.Contains(output, "Content-Type") {
		t.Error("failed to render HTTP header")
	}
	if !strings.Contains(output, "clipboard-copy") {
		t.Error("failed to render clipboard button")
	}
}

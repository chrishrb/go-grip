package frontmatter_test

import (
	"strings"
	"testing"

	"github.com/chrishrb/go-grip/pkg/frontmatter"
)

// ---------- Extract tests ----------

func TestExtract_HappyPath(t *testing.T) {
	src := []byte("---\ntitle: Hello\ndate: 2024-01-01\n---\n# Body\n")
	fm, body, ok := frontmatter.Extract(src)
	if !ok {
		t.Fatal("expected frontmatter to be found")
	}
	if !strings.Contains(string(fm), "title: Hello") {
		t.Errorf("frontmatter missing title, got: %q", fm)
	}
	if !strings.HasPrefix(string(body), "# Body") {
		t.Errorf("body should start with '# Body', got: %q", body)
	}
}

func TestExtract_NoOpeningDash_BlankFirstLine(t *testing.T) {
	src := []byte("\n---\ntitle: Hello\n---\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when first line is blank")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src on no-match")
	}
}

func TestExtract_NoOpeningDash_HeadingFirst(t *testing.T) {
	src := []byte("# Title\n---\nfoo: bar\n---\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when first line is a heading")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src on no-match")
	}
}

func TestExtract_NoClosingDelimiter(t *testing.T) {
	src := []byte("---\ntitle: Hello\nno closer here\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when no closing delimiter")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src on no-match")
	}
}

func TestExtract_InvalidYAML(t *testing.T) {
	src := []byte("---\n: invalid: yaml: :\n---\n# Body\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when YAML is invalid")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src on invalid YAML")
	}
}

func TestExtract_BOMPrefixed(t *testing.T) {
	// BOM before --- means not at byte offset 0
	src := []byte("\xEF\xBB\xBF---\ntitle: Hello\n---\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when --- is preceded by BOM")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src on no-match")
	}
}

func TestExtract_EmptyFrontmatter(t *testing.T) {
	src := []byte("---\n---\n# Body\n")
	_, _, ok := frontmatter.Extract(src)
	// Empty YAML is valid (unmarshals to nil map) -- should still extract
	if !ok {
		t.Fatal("expected empty frontmatter block to be found")
	}
}

func TestExtract_CRLFLineEndings(t *testing.T) {
	src := []byte("---\r\ntitle: Hello\r\n---\r\n# Body\r\n")
	fm, body, ok := frontmatter.Extract(src)
	if !ok {
		t.Fatal("expected frontmatter to be found with CRLF line endings")
	}
	if !strings.Contains(string(fm), "title: Hello") {
		t.Errorf("frontmatter missing title, got: %q", fm)
	}
	if !strings.Contains(string(body), "# Body") {
		t.Errorf("body missing content, got: %q", body)
	}
}

func TestExtract_MidDocumentDash(t *testing.T) {
	// --- appearing only in the middle of the document
	src := []byte("# Title\nsome text\n---\nmore text\n")
	_, body, ok := frontmatter.Extract(src)
	if ok {
		t.Fatal("should not extract frontmatter when --- is mid-document")
	}
	if string(body) != string(src) {
		t.Error("body should equal original src")
	}
}

func TestExtract_BodyPreservedAfterFrontmatter(t *testing.T) {
	body_content := "# My Doc\n\nSome paragraph.\n"
	src := []byte("---\ntitle: Test\n---\n" + body_content)
	_, body, ok := frontmatter.Extract(src)
	if !ok {
		t.Fatal("expected frontmatter found")
	}
	if string(body) != body_content {
		t.Errorf("body mismatch\nwant: %q\ngot:  %q", body_content, body)
	}
}

// ---------- RenderTable tests ----------

func TestRenderTable_HappyPath(t *testing.T) {
	fm := []byte("title: Hello\ndate: 2024-01-01\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `class="frontmatter-table"`) {
		t.Error("missing frontmatter-table class")
	}
	if !strings.Contains(s, "<th>title</th>") {
		t.Error("missing title key cell")
	}
	if !strings.Contains(s, "<td>Hello</td>") {
		t.Error("missing title value cell")
	}
	if !strings.Contains(s, "<th>date</th>") {
		t.Error("missing date key cell")
	}
}

func TestRenderTable_KeyOrder(t *testing.T) {
	// Keys must appear in document order, not map iteration order
	fm := []byte("z_last: 1\na_first: 2\nm_middle: 3\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	zPos := strings.Index(s, "z_last")
	aPos := strings.Index(s, "a_first")
	mPos := strings.Index(s, "m_middle")
	if zPos < 0 || aPos < 0 || mPos < 0 {
		t.Fatalf("not all keys present in output: %s", s)
	}
	if zPos >= aPos || aPos >= mPos {
		t.Errorf("keys not in document order: z@%d a@%d m@%d", zPos, aPos, mPos)
	}
}

func TestRenderTable_NestedMapValue(t *testing.T) {
	fm := []byte("author:\n  name: Alice\n  email: alice@example.com\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<th>author</th>") {
		t.Error("missing author key")
	}
	// Value must be a string representation, not a sub-table
	if strings.Contains(s, "<table") && strings.Count(s, "<table") > 1 {
		t.Error("nested value should not produce a sub-table")
	}
}

func TestRenderTable_SequenceValue(t *testing.T) {
	fm := []byte("tags:\n  - go\n  - markdown\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<th>tags</th>") {
		t.Error("missing tags key")
	}
}

func TestRenderTable_HTMLEscaping(t *testing.T) {
	fm := []byte("desc: <script>alert('xss')</script>\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "<script>") {
		t.Error("raw <script> tag must be HTML-escaped in output")
	}
	if !strings.Contains(s, "&lt;script&gt;") {
		t.Error("expected HTML-escaped &lt;script&gt; in output")
	}
}

func TestRenderTable_AmpersandEscaping(t *testing.T) {
	fm := []byte("title: cats & dogs\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(out), "cats & dogs") {
		t.Error("bare & must be HTML-escaped")
	}
	if !strings.Contains(string(out), "cats &amp; dogs") {
		t.Error("expected &amp; in output")
	}
}

func TestRenderTable_SingleKey(t *testing.T) {
	fm := []byte("key: value\n")
	out, err := frontmatter.RenderTable(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<th>key</th>") || !strings.Contains(s, "<td>value</td>") {
		t.Errorf("single key/value not rendered correctly: %s", s)
	}
}

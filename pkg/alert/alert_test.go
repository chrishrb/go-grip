package alert

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestAlertRendering(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name: "note alert",
			markdown: `> [!NOTE]
> This is a note alert.`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
<p>This is a note alert.</p>
</div>
`,
		},
		{
			name: "tip alert",
			markdown: `> [!TIP]
> This is a tip alert.`,
			want: `<div class="markdown-alert markdown-alert-tip">
<p class="markdown-alert-title"><svg class="octicon octicon-light-bulb mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path></svg>Tip</p>
<p>This is a tip alert.</p>
</div>
`,
		},
		{
			name: "important alert",
			markdown: `> [!IMPORTANT]
> This is an important alert.`,
			want: `<div class="markdown-alert markdown-alert-important">
<p class="markdown-alert-title"><svg class="octicon octicon-report mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>Important</p>
<p>This is an important alert.</p>
</div>
`,
		},
		{
			name: "warning alert",
			markdown: `> [!WARNING]
> This is a warning alert.`,
			want: `<div class="markdown-alert markdown-alert-warning">
<p class="markdown-alert-title"><svg class="octicon octicon-alert mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>Warning</p>
<p>This is a warning alert.</p>
</div>
`,
		},
		{
			name: "caution alert",
			markdown: `> [!CAUTION]
> This is a caution alert.`,
			want: `<div class="markdown-alert markdown-alert-caution">
<p class="markdown-alert-title"><svg class="octicon octicon-stop mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Caution</p>
<p>This is a caution alert.</p>
</div>
`,
		},
		{
			name: "alert with multiple paragraphs",
			markdown: `> [!NOTE]
> First paragraph.
>
> Second paragraph.`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
<p>First paragraph.</p>
<p>Second paragraph.</p>
</div>
`,
		},
		{
			name: "alert with inline formatting",
			markdown: `> [!TIP]
> This has **bold** and *italic* text.`,
			want: `<div class="markdown-alert markdown-alert-tip">
<p class="markdown-alert-title"><svg class="octicon octicon-light-bulb mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path></svg>Tip</p>
<p>This has <strong>bold</strong> and <em>italic</em> text.</p>
</div>
`,
		},
		{
			name: "alert with code",
			markdown: `> [!WARNING]
> Use ` + "`code`" + ` carefully.`,
			want: `<div class="markdown-alert markdown-alert-warning">
<p class="markdown-alert-title"><svg class="octicon octicon-alert mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>Warning</p>
<p>Use <code>code</code> carefully.</p>
</div>
`,
		},
		{
			name: "alert with link",
			markdown: `> [!IMPORTANT]
> See [documentation](https://example.com) for details.`,
			want: `<div class="markdown-alert markdown-alert-important">
<p class="markdown-alert-title"><svg class="octicon octicon-report mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>Important</p>
<p>See <a href="https://example.com">documentation</a> for details.</p>
</div>
`,
		},
		{
			name: "alert with list",
			markdown: `> [!NOTE]
> Items to note:
> - Item 1
> - Item 2`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
<p>Items to note:</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
</ul>
</div>
`,
		},
		{
			name:     "regular blockquote without alert syntax",
			markdown: `> This is a regular blockquote.`,
			want: `<blockquote>
<p>This is a regular blockquote.</p>
</blockquote>
`,
		},
		{
			name: "blockquote with similar but invalid syntax",
			markdown: `> [NOTE]
> This should not be an alert.`,
			want: `<blockquote>
<p>[NOTE]<br />
This should not be an alert.</p>
</blockquote>
`,
		},
		{
			name: "multiple alerts in sequence",
			markdown: `> [!NOTE]
> First note.

> [!WARNING]
> Second warning.`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
<p>First note.</p>
</div>
<div class="markdown-alert markdown-alert-warning">
<p class="markdown-alert-title"><svg class="octicon octicon-alert mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>Warning</p>
<p>Second warning.</p>
</div>
`,
		},
		{
			name: "alert with text after marker on same line",
			markdown: `> [!NOTE] Important information
> This is the content.`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
<p>Important information<br />
This is the content.</p>
</div>
`,
		},
		{
			name:     "empty alert",
			markdown: `> [!NOTE]`,
			want: `<div class="markdown-alert markdown-alert-note">
<p class="markdown-alert-title"><svg class="octicon octicon-info mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>Note</p>
</div>
`,
		},
		{
			name: "alert with code block",
			markdown: `> [!TIP]
> Here's some code:
> ` + "```" + `
> code block
> ` + "```",
			want: `<div class="markdown-alert markdown-alert-tip">
<p class="markdown-alert-title"><svg class="octicon octicon-light-bulb mr-2" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path></svg>Tip</p>
<p>Here's some code:</p>
<pre><code>code block
</code></pre>
</div>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(New()),
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

			got := buf.String()
			t.Logf("Got:\n%s", got)
			t.Logf("Want:\n%s", tt.want)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAlertTypes(t *testing.T) {
	tests := []struct {
		alertType AlertType
		expected  string
	}{
		{AlertTypeNote, "note"},
		{AlertTypeTip, "tip"},
		{AlertTypeImportant, "important"},
		{AlertTypeWarning, "warning"},
		{AlertTypeCaution, "caution"},
	}

	for _, tt := range tests {
		t.Run(string(tt.alertType), func(t *testing.T) {
			alert := NewAlert(tt.alertType)
			assert.Equal(t, tt.alertType, alert.AlertType)
			assert.Equal(t, KindAlert, alert.Kind())
		})
	}
}

func TestAlertWithoutExtension(t *testing.T) {
	markdown := `> [!NOTE]
> This should remain a blockquote.`

	want := `<blockquote>
<p>[!NOTE]<br />
This should remain a blockquote.</p>
</blockquote>
`

	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	assert.Equal(t, want, got)
}

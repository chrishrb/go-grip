package tasklist

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestTaskListRendering(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name: "simple task list",
			markdown: `- [ ] Task 1
- [x] Task 2
- [ ] Task 3`,
			want: `<ul class="contains-task-list">
<li class="task-list-item"><input disabled="" type="checkbox" class="task-list-item-checkbox"> Task 1</li>
<li class="task-list-item"><input checked="" disabled="" type="checkbox" class="task-list-item-checkbox"> Task 2</li>
<li class="task-list-item"><input disabled="" type="checkbox" class="task-list-item-checkbox"> Task 3</li>
</ul>
`,
		},
		{
			name: "task list with text",
			markdown: `- [ ] Fix TODO lists (no dots before)
- [x] Fix TODO lists (no dots before)
- [x] Fix TODO lists (no dots before)`,
			want: `<ul class="contains-task-list">
<li class="task-list-item"><input disabled="" type="checkbox" class="task-list-item-checkbox"> Fix TODO lists (no dots before)</li>
<li class="task-list-item"><input checked="" disabled="" type="checkbox" class="task-list-item-checkbox"> Fix TODO lists (no dots before)</li>
<li class="task-list-item"><input checked="" disabled="" type="checkbox" class="task-list-item-checkbox"> Fix TODO lists (no dots before)</li>
</ul>
`,
		},
		{
			name: "ordered task list",
			markdown: `1. [ ] Task 1
2. [x] Task 2`,
			want: `<ol class="contains-task-list">
<li class="task-list-item"><input disabled="" type="checkbox" class="task-list-item-checkbox"> Task 1</li>
<li class="task-list-item"><input checked="" disabled="" type="checkbox" class="task-list-item-checkbox"> Task 2</li>
</ol>
`,
		},
		{
			name: "regular list without tasks",
			markdown: `- Regular item 1
- Regular item 2`,
			want: `<ul>
<li>Regular item 1</li>
<li>Regular item 2</li>
</ul>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(TaskList),
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

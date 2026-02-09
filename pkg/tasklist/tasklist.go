package tasklist

import (
	"regexp"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var taskListRegexp = regexp.MustCompile(`^\[([\sxX])\]\s*`)

type taskCheckBoxParser struct {
}

var defaultTaskCheckBoxParser = &taskCheckBoxParser{}

// NewTaskCheckBoxParser returns a new  InlineParser that can parse
// checkboxes in list items.
// This parser must take precedence over the parser.LinkParser.
func NewTaskCheckBoxParser() parser.InlineParser {
	return defaultTaskCheckBoxParser
}

func (s *taskCheckBoxParser) Trigger() []byte {
	return []byte{'['}
}

func (s *taskCheckBoxParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	// Given AST structure must be like
	// - List
	//   - ListItem         : parent.Parent
	//     - TextBlock      : parent
	//       (current line)
	if parent.Parent() == nil || parent.Parent().FirstChild() != parent {
		return nil
	}

	if parent.HasChildren() {
		return nil
	}
	if _, ok := parent.Parent().(*gast.ListItem); !ok {
		return nil
	}
	line, _ := block.PeekLine()
	m := taskListRegexp.FindSubmatchIndex(line)
	if m == nil {
		return nil
	}
	value := line[m[2]:m[3]][0]
	block.Advance(m[1])
	checked := value == 'x' || value == 'X'
	return ast.NewTaskCheckBox(checked)
}

func (s *taskCheckBoxParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// TaskCheckBoxHTMLRenderer is a renderer.NodeRenderer implementation that
// renders checkboxes in list items.
type TaskCheckBoxHTMLRenderer struct {
	html.Config
}

// NewTaskCheckBoxHTMLRenderer returns a new TaskCheckBoxHTMLRenderer.
func NewTaskCheckBoxHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TaskCheckBoxHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *TaskCheckBoxHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindTaskCheckBox, r.renderTaskCheckBox)
}

func (r *TaskCheckBoxHTMLRenderer) renderTaskCheckBox(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	n := node.(*ast.TaskCheckBox)

	if n.IsChecked {
		_, _ = w.WriteString(`<input checked="" disabled="" type="checkbox" class="task-list-item-checkbox"`)
	} else {
		_, _ = w.WriteString(`<input disabled="" type="checkbox" class="task-list-item-checkbox"`)
	}
	_, _ = w.WriteString("> ")
	return gast.WalkContinue, nil
}

// TaskListHTMLRenderer adds CSS classes to ul and li elements for lists containing task checkboxes.
type TaskListHTMLRenderer struct {
	html.Config
}

func NewTaskListHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TaskListHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *TaskListHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(gast.KindList, r.renderList)
	reg.Register(gast.KindListItem, r.renderListItem)
}

func (r *TaskListHTMLRenderer) renderList(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*gast.List)
	
	// Check if this is a task list
	isTaskList := false
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*gast.ListItem); ok {
			// Check all descendants, not just direct children
			var checkForTaskCheckBox func(gast.Node) bool
			checkForTaskCheckBox = func(n gast.Node) bool {
				if n.Kind() == ast.KindTaskCheckBox {
					return true
				}
				for gc := n.FirstChild(); gc != nil; gc = gc.NextSibling() {
					if checkForTaskCheckBox(gc) {
						return true
					}
				}
				return false
			}
			if checkForTaskCheckBox(li) {
				isTaskList = true
				break
			}
		}
	}
	
	if entering {
		// Render the opening tag manually
		tag := "ul"
		if n.IsOrdered() {
			tag = "ol"
		}
		_, _ = w.WriteString("<")
		_, _ = w.WriteString(tag)
		
		if isTaskList {
			_, _ = w.WriteString(` class="contains-task-list"`)
		}
		
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.ListAttributeFilter)
		}
		_, _ = w.WriteString(">")
		_, _ = w.WriteString("\n")
	} else {
		tag := "ul"
		if n.IsOrdered() {
			tag = "ol"
		}
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString(">")
		_, _ = w.WriteString("\n")
	}
	return gast.WalkContinue, nil
}

func (r *TaskListHTMLRenderer) renderListItem(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	// Check if this is a task list item by checking all descendants
	isTaskItem := false
	var checkForTaskCheckBox func(gast.Node) bool
	checkForTaskCheckBox = func(n gast.Node) bool {
		if n.Kind() == ast.KindTaskCheckBox {
			return true
		}
		for gc := n.FirstChild(); gc != nil; gc = gc.NextSibling() {
			if checkForTaskCheckBox(gc) {
				return true
			}
		}
		return false
	}
	isTaskItem = checkForTaskCheckBox(node)
	
	if entering {
		_, _ = w.WriteString("<li")
		
		if isTaskItem {
			_, _ = w.WriteString(` class="task-list-item"`)
		}
		
		if node.Attributes() != nil {
			html.RenderAttributes(w, node, html.ListItemAttributeFilter)
		}
		_, _ = w.WriteString(">")
	} else {
		_, _ = w.WriteString(`</li>`)
		_, _ = w.WriteString("\n")
	}
	return gast.WalkContinue, nil
}

type taskList struct {
}

// TaskList is an extension that allow you to use GFM task lists.
var TaskList = &taskList{}

func (e *taskList) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewTaskCheckBoxParser(), 0),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTaskCheckBoxHTMLRenderer(), 500),
		util.Prioritized(NewTaskListHTMLRenderer(), 501),
	))
}


package mermaid

import "github.com/yuin/goldmark/ast"

// Kind is the node kind of a Mermaid [Block] node.
var Kind = ast.NewNodeKind("MermaidBlock")

// Block is a Mermaid block.
//
//	```mermaid
//	graph TD;
//	    A-->B;
//	    A-->C;
//	    B-->D;
//	    C-->D;
//	```
//
// Its raw contents are the plain text of the Mermaid diagram.
type Block struct {
	ast.BaseBlock
}

// IsRaw reports that this block should be rendered as-is.
func (*Block) IsRaw() bool { return true }

// Kind reports that this is a MermaidBlock.
func (*Block) Kind() ast.NodeKind { return Kind }

// Dump dumps the contents of this block to stdout.
func (b *Block) Dump(src []byte, level int) {
	ast.DumpHelper(b, src, level, nil, nil)
}

// ScriptKind is the node kind of a Mermaid [ScriptBlock] node.
var ScriptKind = ast.NewNodeKind("MermaidScriptBlock")

// ScriptBlock marks where the Mermaid Javascript will be included.
//
// This is a placeholder and does not contain anything.
type ScriptBlock struct {
	ast.BaseBlock
}

// IsRaw reports that this block should be rendered as-is.
func (*ScriptBlock) IsRaw() bool { return true }

// Kind reports that this is a MermaidScriptBlock.
func (*ScriptBlock) Kind() ast.NodeKind { return ScriptKind }

// Dump dumps the contents of this block to stdout.
func (b *ScriptBlock) Dump(src []byte, level int) {
	ast.DumpHelper(b, src, level, nil, nil)
}

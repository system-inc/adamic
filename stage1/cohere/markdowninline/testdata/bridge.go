package markdown

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// AdamicLeaf reaches the actual Go printer with real AstPath ancestor and sibling stacks.
func AdamicLeaf(mode byte, text string) string {
	node := &Node{NodeType: "word", Value: text}
	parent := &Node{NodeType: "sentence", IsParent: true, Children: []*Node{node}}
	root := &Node{NodeType: "paragraph", IsParent: true, Children: []*Node{parent}}
	prose := "preserve"
	index, parentIndex := 0, 0
	switch mode {
	case 'w', 'e', 'f', 'n':
		root.NodeType = "emphasis"
		if mode == 'w' || mode == 'e' {
			flank := "a"
			if mode == 'e' {
				flank = " "
			}
			parent.Children = []*Node{{NodeType: "word", Value: flank}, node, {NodeType: "word", Value: flank}}
			index = 1
		}
		if mode == 'n' {
			root.Children = []*Node{{NodeType: "sentence", IsParent: true}, parent}
			parentIndex = 1
		}
	case 's':
		parent.Children = []*Node{{NodeType: "whitespace", Value: "\n"}, node, {NodeType: "whitespace", Value: "\n"}}
		index = 1
	case 'p':
	case 'c', 't', 'r', 'u':
		node.NodeType = "inlineCode"
		if mode == 't' || mode == 'u' {
			root.NodeType = "tableCell"
		}
		if mode == 'r' || mode == 'u' {
			prose = "never"
		}
	case 'k', 'v':
		node.NodeType = "wikiLink"
		if mode == 'v' {
			prose = "never"
		}
	case 'h', 'i':
		node.NodeType = "image"
		node.Alt, node.Title, node.URL = &text, &text, text
	case 'o':
		node.NodeType = "image"
		fallback := "fallback"
		node.Alt, node.OriginalAltText, node.URL = &fallback, &text, text
	case 'j', 'l', 'q':
		node.NodeType = "imageReference"
		node.Alt, node.Label = &text, &text
		node.ReferenceType = "full"
		if mode == 'l' {
			node.ReferenceType = "collapsed"
		}
		if mode == 'q' {
			node.ReferenceType = "shortcut"
		}
	case 'b':
		node.NodeType = "footnoteReference"
		node.Label = &text
	default:
		panic("unknown leaf mode")
	}
	path := printing.NewAstPath(root)
	options := &options{Settings: &settings{proseWrap: prose, singleQuote: mode == 'i'}}
	document := printing.Call(path, func(path *astPath) doc.Doc {
		return printMdast(path, options, func(any, any) doc.Doc { panic("leaf must not print children") }, nil)
	}, "children", parentIndex, "children", index)
	return doc.Print(document, doc.Options{PrintWidth: 80, TabWidth: 2})
}

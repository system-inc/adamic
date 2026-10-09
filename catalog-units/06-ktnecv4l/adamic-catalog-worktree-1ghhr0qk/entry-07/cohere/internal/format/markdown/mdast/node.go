// Package mdast is the syntax tree Prettier's markdown printer reads, and mdast-util-from-markdown, the
// compiler that builds it from micromark's events.
//
// # One struct for every node type
//
// Upstream's nodes are open JavaScript objects, and the printer reads fields across types (`node.value`
// on whatever it holds, `node.children` on anything) and adds its own in preprocessing. So a node here is
// one struct carrying every field any mdast type, extension or printer pass uses. Fields upstream can
// hold null distinctly from a value are pointers.
//
// # Positions are bytes
//
// micromark counts UTF-16 code units. The tree's offsets are converted to byte offsets into the original
// UTF-8 text, the unit every printer in cohere slices by. Lines and columns stay as micromark counts
// them; the printer reads columns only to compare where blocks start, after ASCII container prefixes.
//
// A node's Position is a pointer because upstream shares the position object between a node and the
// clones the printer's preprocessing makes, and keys a set on that identity.
package mdast

import "github.com/system-inc/cohere/internal/format/printing"

// Point is a place in the source: 1-based line and column (UTF-16 units), 0-based byte offset.
type Point struct {
	Line   int
	Column int
	Offset int
}

// Position is a node's span.
type Position struct {
	Start Point
	End   Point
}

// Node is any mdast node.
type Node struct {
	// NodeType is upstream's node.type. A field named Type would collide with the Type method the
	// print core requires.
	NodeType string
	Position *Position

	// Children is a parent's children. IsParent is upstream's `'children' in node`, which an empty list
	// also satisfies.
	Children []*Node
	IsParent bool

	// Value is a literal's value. IsLiteral is upstream's `'value' in node`; ValueNull is a wiki link
	// whose target never closed, where upstream's value stays null.
	Value     string
	IsLiteral bool
	ValueNull bool

	// Heading.
	Depth int

	// List and list item.
	Ordered bool
	Start   *int
	Spread  bool
	Checked *bool

	// Code and math.
	Lang *string
	Meta *string

	// Link, image, definition and their references.
	URL           string
	Title         *string
	Alt           *string
	Label         *string
	Identifier    string
	ReferenceType string

	// Table: per column, "left", "right", "center", or "" for upstream's null.
	Align []string

	// Front matter, added by parseMarkdown.
	FrontMatter *FrontMatter

	// Fields the printer's preprocess adds (src/language-markdown/print/preprocess.js). Raw is nil where
	// upstream's is undefined: a text node made by merging two.
	Raw                  *string
	IsIndented           bool
	IsAligned            bool
	OriginalAltText      *string
	HasIndentedCodeblock bool

	// Sentence, word and whitespace nodes, made by splitText. UsesCJSpaces caches a sentence's answer
	// once whitespace.js computes it, as upstream injects the property.
	Kind                   string
	IsCJ                   bool
	HasLeadingPunctuation  bool
	HasTrailingPunctuation bool
	UsesCJSpaces           *bool

	comments printing.CommentFields[*Node]
}

// FrontMatter is the front matter block upstream's parseFrontMatter finds.
type FrontMatter struct {
	Language         string
	ExplicitLanguage *string
	Value            string
	StartDelimiter   string
	EndDelimiter     string
	Raw              string
}

// Type is upstream's node.type, the method the print core reads it through.
func (node *Node) Type() string { return node.NodeType }

// Field is upstream's node[name] for the keys the print core navigates: "children", the only key the
// markdown printer walks, and "comments", which markdown never has.
func (node *Node) Field(name string) any {
	switch name {
	case "children":
		if !node.IsParent {
			return nil
		}
		return node.Children
	case "comments":
		return node.comments.Comments
	default:
		return nil
	}
}

// CommentData is the print core's comment bookkeeping. Markdown attaches no comments, so it stays empty.
func (node *Node) CommentData() *printing.CommentFields[*Node] { return &node.comments }

// ToString is mdast-util-to-string with its defaults (image alt and html included).
func ToString(node *Node) string {
	if node == nil {
		return ""
	}
	if node.IsLiteral {
		return node.Value
	}
	if node.Alt != nil && *node.Alt != "" {
		return *node.Alt
	}
	if node.IsParent {
		return toStringAll(node.Children)
	}
	return ""
}

func toStringAll(nodes []*Node) string {
	result := ""
	for _, node := range nodes {
		result += ToString(node)
	}
	return result
}

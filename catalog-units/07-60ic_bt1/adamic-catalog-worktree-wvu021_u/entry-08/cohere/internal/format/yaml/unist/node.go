// Package unist is the syntax tree Prettier's YAML printer reads: yaml-unist-parser 3.2.0's tree, built
// over the documents eemeli/yaml 2.9.0 composes.
//
// # One struct for every node type
//
// Upstream's nodes are open JavaScript objects (types.d.mts), and the printer reads fields across types.
// So a node here is one struct carrying every field any node type has. Fields upstream can hold null
// distinctly from a value are pointers.
//
// # Positions are bytes
//
// eemeli/yaml counts UTF-16 code units, as a JavaScript string does. Offsets are converted once, where
// the tree is built, to byte offsets into the original UTF-8 text, the unit every printer in cohere
// slices by. Lines and columns stay as upstream counts them: 1-based, columns in UTF-16 units.
package unist

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

// Node is any yaml-unist-parser node.
type Node struct {
	// NodeType is upstream's node.type. A field named Type would collide with the Type method the
	// print core requires.
	NodeType string
	Position Position

	// Parent is upstream's non-enumerable _parent, set by defineParents.
	Parent *Node

	// Children is a parent's children, in upstream's order. A document's are its head and body, a
	// mapping item's its key and value, and a key's, value's or item's its content, or none.
	Children []*Node

	// Value is a literal's value: comment, anchor, tag, alias, plain, quoted and block scalars.
	Value string

	// Content: tag and anchor, nil when absent, and the comments between them and the node.
	Tag            *Node
	Anchor         *Node
	MiddleComments []*Node

	// Comment attachment, by attach.mjs. Each field exists only on the types types.d.mts gives it.
	LeadingComments []*Node
	TrailingComment *Node
	EndComments     []*Node

	// Block scalars: chomping is "clip", "keep" or "strip"; Indent is the explicit indentation
	// indicator, nil when absent; IndicatorComment is a comment on the header line.
	Chomping         string
	Indent           *int
	IndicatorComment *Node

	// Document markers.
	DirectivesEndMarker bool
	DocumentEndMarker   bool

	// Directive.
	Name       string
	Parameters []string

	// Root: every comment in the file. Prettier's parser deletes it before printing.
	Comments []*Node

	comments printing.CommentFields[*Node]
}

// Type is upstream's node.type, the method the print core reads it through.
func (node *Node) Type() string { return node.NodeType }

// child is node.children[index], nil where upstream's is undefined.
func (node *Node) child(index int) *Node {
	if index < len(node.Children) {
		return node.Children[index]
	}
	return nil
}

// Field is upstream's node[name] for the keys the YAML printer's visitor keys name, the shortcuts its
// preprocess defines included (print-preprocess.js), and "comments" for the print core.
func (node *Node) Field(name string) any {
	switch name {
	case "children":
		return node.Children
	case "head":
		if node.NodeType == "document" {
			return node.child(0)
		}
	case "body":
		if node.NodeType == "document" {
			return node.child(1)
		}
	case "key":
		if node.NodeType == "mappingItem" || node.NodeType == "flowMappingItem" {
			return node.child(0)
		}
	case "value":
		if node.NodeType == "mappingItem" || node.NodeType == "flowMappingItem" {
			return node.child(1)
		}
	case "content":
		switch node.NodeType {
		case "documentBody", "sequenceItem", "flowSequenceItem", "mappingKey", "mappingValue":
			return node.child(0)
		}
	case "tag":
		return node.Tag
	case "anchor":
		return node.Anchor
	case "middleComments":
		return node.MiddleComments
	case "leadingComments":
		return node.LeadingComments
	case "trailingComment":
		return node.TrailingComment
	case "endComments":
		return node.EndComments
	case "indicatorComment":
		return node.IndicatorComment
	case "comments":
		return node.comments.Comments
	}
	return nil
}

// CommentData is the print core's comment bookkeeping. The YAML printer prints comments itself, so it
// stays empty.
func (node *Node) CommentData() *printing.CommentFields[*Node] { return &node.comments }

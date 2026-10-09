package css

// Go's spellings of the JavaScript printer-postcss.js and its print/ helpers lean on: the path and options
// types, path.node and path.map, the `cond ? doc : ""` idiom, and reading the plain objects (raws,
// source) and arrays the nodes carry.

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// astPath is upstream's AstPath over postcss nodes.
type astPath = printing.AstPath[*estree.Node]

// printerOptions is upstream's options object for one format.
type printerOptions = printing.Options[*estree.Node]

// currentNode is upstream's path.node.
func currentNode(path *astPath) *estree.Node {
	node, _ := path.Node()
	return node
}

// asNode is a path.match predicate's node argument as a node, or nil.
func asNode(value any) *estree.Node {
	node, _ := value.(*estree.Node)
	return node
}

// printAll is upstream's path.map(print, ...names).
func printAll(path *astPath, print printing.PrintFunc, names ...any) []doc.Doc {
	return printing.Map(path, func(*astPath, int, any) doc.Doc { return print(nil, nil) }, names...)
}

// choose is upstream's `condition ? doc : ""`, printing the doc only when it is chosen.
func choose(condition bool, document func() doc.Doc) doc.Doc {
	if condition {
		return document()
	}
	return doc.Text("")
}

// rawsOf is node.raws, postcss's and the sub-parsers' plain object of source details.
func rawsOf(node *estree.Node) map[string]any {
	raws, _ := node.Get("raws").(map[string]any)
	return raws
}

// rawString is node.raws[key] when it is a string, and "" otherwise.
func rawString(node *estree.Node, key string) string {
	value, _ := rawsOf(node)[key].(string)
	return value
}

// isPresent is JavaScript's truthiness for a property that holds an array, where an empty array is
// truthy: a nil slice stands for an absent one.
func isPresent(value any) bool {
	switch typed := value.(type) {
	case []*estree.Node:
		return typed != nil
	case []any:
		return typed != nil
	}
	return estree.IsTruthy(value)
}

// firstNode is list[0], or nil for an empty list.
func firstNode(list []*estree.Node) *estree.Node {
	if len(list) == 0 {
		return nil
	}
	return list[0]
}

// someNode and everyNode are Array.prototype.some and every.
func someNode(list []*estree.Node, predicate func(*estree.Node) bool) bool {
	for _, element := range list {
		if predicate(element) {
			return true
		}
	}
	return false
}

func everyNode(list []*estree.Node, predicate func(*estree.Node) bool) bool {
	for _, element := range list {
		if !predicate(element) {
			return false
		}
	}
	return true
}

// mustNode is a property access upstream makes without optional chaining: on a missing node JavaScript
// throws a TypeError, and so does the port, so a source the fork fails on fails here too.
func mustNode(node *estree.Node) *estree.Node {
	if node == nil {
		panic("TypeError: Cannot read properties of undefined (reading 'type')")
	}
	return node
}

// sliceText is String.prototype.slice for the non-negative offsets the printer passes, clamped as
// JavaScript clamps them.
func sliceText(text string, start int, end int) string {
	start = min(max(start, 0), len(text))
	end = min(max(end, 0), len(text))
	if start >= end {
		return ""
	}
	return text[start:end]
}

// characterBefore is originalText[offset - 1] for the ASCII comparisons the printer makes.
func characterBefore(text string, offset int) string {
	if offset-1 < 0 || offset-1 >= len(text) {
		return ""
	}
	return text[offset-1 : offset]
}

// sourceStartLine is node.source.start.line, and whether the node has a source.
func sourceStartLine(node *estree.Node) (int, bool) {
	source, _ := node.Get("source").(map[string]any)
	if source == nil {
		return 0, false
	}
	start, _ := source["start"].(map[string]any)
	switch line := start["line"].(type) {
	case int:
		return line, true
	case float64:
		return int(line), true
	}
	return 0, true
}

package selector

// postcss-selector-parser 2.2.3, dist/selectors/*.js: the node classes, reduced to what their
// constructors leave on the object. The methods (toString, walk, clone, remove...) are not ported:
// Prettier reads the tree the parser builds and never calls them on it.
//
// Every class's constructor runs node.js's first: it copies the options object's own keys onto the
// node in order (an undefined value included, as an own property that is undefined, which is a nil
// property here), then sets spaces. container.js adds nodes, the subclass sets type, and nesting.js and
// universal.js overwrite value, attribute.js adds raws. The JS object's type is the estree node's type.
//
// parent is left off the nodes, as the brief has it; the parser keeps it on the side (see parser.parents).

import "github.com/system-inc/cohere/internal/format/estree"

// types.js.
const (
	typeTag        = "tag"
	typeString     = "string"
	typeSelector   = "selector"
	typeRoot       = "root"
	typePseudo     = "pseudo"
	typeNesting    = "nesting"
	typeID         = "id"
	typeComment    = "comment"
	typeCombinator = "combinator"
	typeClass      = "class"
	typeAttribute  = "attribute"
	typeUniversal  = "universal"
)

// newSelectorNode is node.js's constructor followed by the subclass's: opts are the options object's
// keys and values in order.
func newSelectorNode(nodeType string, opts ...any) *estree.Node {
	node := estree.New(nodeType, 0, 0, opts...)
	// Prettier's parser never passes opts.spaces, so before and after are always the '' defaults.
	node.Set("spaces", map[string]any{"before": "", "after": ""})
	switch nodeType {
	case typeRoot, typeSelector, typePseudo:
		// container.js: if (!this.nodes) this.nodes = [].
		if node.Get("nodes") == nil {
			node.Set("nodes", []*estree.Node{})
		}
	case typeNesting:
		node.Set("value", "&")
	case typeUniversal:
		node.Set("value", "*")
	case typeAttribute:
		node.Set("raws", map[string]any{})
	}
	return node
}

// source is the {start: {line, column}, end: {line, column}} object the parser passes as opts.source.
// Each is an int, except where splitWord reads past a space token: nil for undefined, or NaN.
func source(startLine any, startColumn any, endLine any, endColumn any) map[string]any {
	return map[string]any{
		"start": map[string]any{"line": startLine, "column": startColumn},
		"end":   map[string]any{"line": endLine, "column": endColumn},
	}
}

// sourceOf is node.source, mutated in place where the parser moves a start or an end.
func sourceOf(node *estree.Node) map[string]any {
	value, _ := node.Get("source").(map[string]any)
	return value
}

// spacesOf is node.spaces.
func spacesOf(node *estree.Node) map[string]any {
	value, _ := node.Get("spaces").(map[string]any)
	return value
}

// rawsOf is attribute.raws.
func rawsOf(node *estree.Node) map[string]any {
	value, _ := node.Get("raws").(map[string]any)
	return value
}

// lastOf is container.last, nil where JavaScript has undefined.
func lastOf(container *estree.Node) *estree.Node {
	nodes := container.List("nodes")
	if len(nodes) == 0 {
		return nil
	}
	return nodes[len(nodes)-1]
}

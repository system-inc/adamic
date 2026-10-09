// Port of postcss-values-parser 2.0.1 lib/node.js, lib/container.js and the node classes: root.js,
// value.js, atword.js, colon.js, comma.js, comment.js, function.js, number.js, operator.js, paren.js,
// string.js, unicode-range.js, word.js.
//
// Each constructor builds the estree node with the own enumerable fields its class leaves on the
// JavaScript object, in the order JavaScript assigns them: Node's constructor sets raws, then copies the
// options it was given (value, source, sourceIndex and the class's extras), then the class sets type
// and its own fields. type is the node's type rather than a property. parent is left out, and so is
// everything functional (toString, clone, walk and the rest), which the parser never calls.

package values

import "github.com/system-inc/cohere/internal/format/estree"

// newRaws is Node's `this.raws = { before: "", after: "" }`.
func newRaws() map[string]any {
	return map[string]any{"before": "", "after": ""}
}

// position is a source start or end: { line, column }. column is in UTF-16 units, the library's own,
// and is NaN (a float64) where upstream computes it from an undefined variable.
func position(line int, column any) map[string]any {
	return map[string]any{"line": line, "column": column}
}

func source(start map[string]any, end map[string]any) map[string]any {
	return map[string]any{"start": start, "end": end}
}

// Root: Container with type 'root'.
func newRoot() *estree.Node {
	return estree.New("root", 0, 0, "raws", newRaws(), "nodes", []*estree.Node{})
}

// Value: Container with type 'value' and unbalanced 0.
func newValue() *estree.Node {
	return estree.New("value", 0, 0, "raws", newRaws(), "nodes", []*estree.Node{}, "unbalanced", float64(0))
}

// leaf builds the Node subclasses that take { value, source, sourceIndex }: colon, comma, operator,
// unicode-range and word.
func leaf(nodeType string, value string, nodeSource map[string]any, sourceIndex int) *estree.Node {
	return estree.New(nodeType, 0, 0, "raws", newRaws(), "value", value, "source", nodeSource, "sourceIndex", sourceIndex)
}

// Paren: Node with type 'paren' and parenType "".
func newParen(value string, nodeSource map[string]any, sourceIndex int) *estree.Node {
	node := leaf("paren", value, nodeSource, sourceIndex)
	node.Set("parenType", "")
	return node
}

// Comment: Node with type 'comment' and inline from its options.
func newComment(value string, inline bool, nodeSource map[string]any, sourceIndex int) *estree.Node {
	return estree.New("comment", 0, 0, "raws", newRaws(), "value", value, "inline", inline, "source", nodeSource, "sourceIndex", sourceIndex)
}

// AtWord: Container with type 'atword'.
func newAtWord(value string, nodeSource map[string]any, sourceIndex int) *estree.Node {
	return estree.New("atword", 0, 0, "raws", newRaws(), "value", value, "source", nodeSource, "sourceIndex", sourceIndex, "nodes", []*estree.Node{})
}

// NumberNode: Node with type 'number' and unit from its options, or "".
func newNumber(value string, nodeSource map[string]any, sourceIndex int, unit string) *estree.Node {
	return estree.New("number", 0, 0, "raws", newRaws(), "value", value, "source", nodeSource, "sourceIndex", sourceIndex, "unit", unit)
}

// FunctionNode: Container with type 'func', unbalanced starting at -1 so the parser knows no parens
// have been added yet.
func newFunc(value string, nodeSource map[string]any, sourceIndex int) *estree.Node {
	return estree.New("func", 0, 0, "raws", newRaws(), "value", value, "source", nodeSource, "sourceIndex", sourceIndex, "nodes", []*estree.Node{}, "unbalanced", float64(-1))
}

// StringNode: Node with type 'string' and quoted from its options.
func newString(value string, nodeSource map[string]any, sourceIndex int, quoted bool) *estree.Node {
	return estree.New("string", 0, 0, "raws", newRaws(), "value", value, "source", nodeSource, "sourceIndex", sourceIndex, "quoted", quoted)
}

// raws is node.raws.
func raws(node *estree.Node) map[string]any {
	return node.Get("raws").(map[string]any)
}

// appendRaw is node.raws[key] += text.
func appendRaw(node *estree.Node, key string, text string) {
	nodeRaws := raws(node)
	nodeRaws[key] = nodeRaws[key].(string) + text
}

// appendNode is Container's append: node.parent = this; this.nodes.push(node).
func appendNode(container *estree.Node, node *estree.Node) {
	container.Set("nodes", append(container.List("nodes"), node))
}

// last is Container's `get last`: the last child, or nil.
func last(container *estree.Node) *estree.Node {
	nodes := container.List("nodes")
	if len(nodes) == 0 {
		return nil
	}
	return nodes[len(nodes)-1]
}

func unbalanced(node *estree.Node) float64 {
	value, _ := node.Get("unbalanced").(float64)
	return value
}

func setUnbalanced(node *estree.Node, value float64) {
	node.Set("unbalanced", value)
}

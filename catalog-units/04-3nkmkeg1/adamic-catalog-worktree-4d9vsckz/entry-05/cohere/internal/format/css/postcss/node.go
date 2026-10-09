package postcss

// The node classes the parser builds: lib/root.js, lib/rule.js, lib/at-rule.js, lib/declaration.js and
// lib/comment.js, over lib/node.js and lib/container.js, reduced to their fields.
//
// A node is plain data while the parser runs, with offsets in UTF-16 units, and becomes an estree.Node
// in parse.go. keys records the order upstream creates its own enumerable fields in, after raws (which
// Node's constructor creates first) and type (the estree.Node's type), so the estree.Node's properties
// come out in the same order a for-in over the JavaScript object visits them. parent is kept for the
// parser and not handed out.
//
// Node's methods (clone, remove, toString, the proxy, the visitor) and Container's (each, walk,
// insertBefore, normalize) are not here: the parser calls only Container's push.

type node struct {
	nodeType string
	raws     map[string]any
	keys     []string

	parent *node
	nodes  []*node
	source *source

	// Rule.
	selector string
	// AtRule.
	name   string
	params string
	// Declaration.
	prop      string
	value     string
	important bool
	// Comment.
	text string
	// postcss-scss's NestedDeclaration (nested_declaration.go), a decl with nodes.
	isNested bool
}

// source is node.source without input: { start, end }, end unset until the parser finds it.
type source struct {
	start *position
	end   *position
}

// position is what Parser.getPosition returns: { column, line, offset }, offset in UTF-16 units.
type position struct {
	column int
	line   int
	offset int
}

// rawValue is what Parser.raw stores in raws[prop] when the clean value differs from the source text.
//
// scss is the field postcss-scss's raw adds, the text with `//` comments as written, where its raw
// (the same comments rewritten as /* */) differs; hasScss says it is set.
type rawValue struct {
	raw   string
	value string

	scss    string
	hasScss bool
}

// mark records that the field exists, the first time it is assigned.
func (each *node) mark(key string) {
	for _, existing := range each.keys {
		if existing == key {
			return
		}
	}
	each.keys = append(each.keys, key)
}

// Root's constructor: `if (!this.nodes) this.nodes = []`.
func newRoot() *node {
	return &node{nodeType: "root", raws: map[string]any{}, keys: []string{"nodes"}, nodes: []*node{}}
}

// Rule's constructor: `if (!this.nodes) this.nodes = []`.
func newRule() *node {
	return &node{nodeType: "rule", raws: map[string]any{}, keys: []string{"nodes"}, nodes: []*node{}}
}

// AtRule's constructor leaves nodes unset; the parser creates them when a block opens.
func newAtRule() *node {
	return &node{nodeType: "atrule", raws: map[string]any{}}
}

func newDeclaration() *node {
	return &node{nodeType: "decl", raws: map[string]any{}}
}

func newComment() *node {
	return &node{nodeType: "comment", raws: map[string]any{}}
}

// push is Container's push: the child joins this node's nodes with this node as its parent.
func (each *node) push(child *node) {
	child.parent = each
	each.nodes = append(each.nodes, child)
}

// hasNodes is `this.nodes` being truthy: a container whose nodes array exists, empty or not.
func (each *node) hasNodes() bool {
	return each.nodes != nil
}

// rawString is `this.raws[key] || ""` read as a string: the empty string where it is unset.
func (each *node) rawString(key string) string {
	text, _ := each.raws[key].(string)
	return text
}

// setField is `node[prop] = value` for the three fields Parser.raw writes.
func (each *node) setField(prop string, value string) {
	switch prop {
	case "selector":
		each.selector = value
	case "params":
		each.params = value
	case "value":
		each.value = value
	default:
		panic("postcss: raw writes selector, params or value, not " + prop)
	}
	each.mark(prop)
}

// emptyNodes is `[]`, for a parser method whose local variable is named node, as upstream's is.
func emptyNodes() []*node {
	return []*node{}
}

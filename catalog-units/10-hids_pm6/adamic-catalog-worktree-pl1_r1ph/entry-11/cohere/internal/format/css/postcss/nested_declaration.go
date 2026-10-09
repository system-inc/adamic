package postcss

// postcss-scss 4.0.9, lib/nested-declaration.js: a Container whose type is "decl", for a nested
// property with a value (`margin: 0 { left: 1px }`), holding the declarations in its block.
//
// The constructor runs Node's (raws), then sets type, isNested and nodes, so those are its first own
// fields after raws and type, in that order.
func newNestedDeclaration() *node {
	return &node{nodeType: "decl", raws: map[string]any{}, keys: []string{"isNested", "nodes"}, isNested: true, nodes: []*node{}}
}

package css

// src/language-css/utilities/get-value-root.js.
//
// Upstream climbs node.parent to the postcss-values-parser root. The Go trees carry no parent links,
// so parseNestedValue passes the root it is walking down to every node instead, and this is that root.

import "github.com/system-inc/cohere/internal/format/estree"

// valueRootWalk is the root of the postcss-values-parser tree being walked: what upstream's
// getValueRoot(node) climbs to from any node in it.
type valueRootWalk struct {
	root *estree.Node
}

func getValueRoot(walk valueRootWalk) *estree.Node {
	return walk.root
}

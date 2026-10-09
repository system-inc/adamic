package mediaquery

import "github.com/system-inc/cohere/internal/format/estree"

// postcss-media-query-parser 0.2.3, dist/nodes/Node.js and dist/nodes/Container.js.
//
// Both are constructors whose only state is own fields, so here they build an *estree.Node with those
// fields as properties in the constructor's assignment order: after, before, value, sourceIndex, then
// nodes for a container. type is the node's type, "" where the library leaves it undefined (a node
// whose type the post-pass in parseMediaQuery could not decide; Prettier's addMissingType fills it).
//
// A field the caller passes as undefined is a nil here (only the root, built by index.js without
// after, before or sourceIndex, does this), and Container fills it in place, keeping its position the
// way assigning an existing JavaScript property does.
//
// Not ported: parent (the brief leaves it out), and Container's walk and each, which Prettier never
// calls on this tree.

// newNode is `new Node(opts)`: a very generic node. Pretty much any element of a media query.
func newNode(after any, before any, nodeType string, value string, sourceIndex any) *estree.Node {
	return estree.New(nodeType, 0, 0,
		"after", after,
		"before", before,
		"value", value,
		"sourceIndex", sourceIndex,
	)
}

// newContainer is `new Container(opts)`: a node that contains other nodes.
func newContainer(after any, before any, nodeType string, value string, sourceIndex any, nodes []*estree.Node) *estree.Node {
	container := newNode(after, before, nodeType, value, sourceIndex)

	container.Set("nodes", nodes)

	if container.Get("after") == nil {
		if len(nodes) > 0 {
			container.Set("after", nodes[len(nodes)-1].Get("after"))
		} else {
			container.Set("after", "")
		}
	}

	if container.Get("before") == nil {
		if len(nodes) > 0 {
			container.Set("before", nodes[0].Get("before"))
		} else {
			container.Set("before", "")
		}
	}

	if container.Get("sourceIndex") == nil {
		// this.before.length: before is whitespace from the input, so its byte length is the byte offset.
		container.Set("sourceIndex", len(container.String("before")))
	}

	return container
}

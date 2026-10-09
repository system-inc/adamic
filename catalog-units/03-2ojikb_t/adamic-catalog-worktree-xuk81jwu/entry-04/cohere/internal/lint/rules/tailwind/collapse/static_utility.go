// The framework's static utility registrations, and the reading a static body produces.
//
// A static utility takes no value: `flex` is `display: flex` and `sr-only` is nine declarations, and
// neither consults the theme. So its reading is a constant, and the constant is PropertySort over
// the declarations it compiles to. This file holds the declaration type the generated table emits
// and the lookup that turns a body into a reading.
//
// # Why the table carries declarations rather than readings
//
// The base descriptor table carries `{order, count}` per static, which is the answer rather than the
// input. Carrying declarations instead means the reading is computed by the same PropertySort walk
// that answers a repository's own `@utility` blocks in descriptor_live.go, so the framework path and
// the repository path cannot drift apart about what a declaration list means. A table of readings
// can disagree with the walk; a table of declarations feeding the walk cannot.
//
// It also makes a wrong entry loud rather than plausible. A transposed reading is a number that
// still sorts; a transposed property name is visible against the engine's own compiled tree, which
// is what `internal/lint/rules/tailwind/tools/generate_statics/enumerate.mjs` captures.
package tailwind

// StaticDeclaration is one declaration of a static utility's compiled body.
//
// The fields mirror the declaration kind in ast.go rather than a reduced form, because PropertySort
// reads all of them: ValuePresent decides whether a declaration is counted at all, since upstream
// skips one whose value is `undefined` and counts one whose value is the empty string.
type StaticDeclaration struct {
	Property     string
	Value        string
	ValuePresent bool
	Important    bool
}

// FrameworkStaticReading returns the reading of a framework static utility, and whether it is one.
//
// The reading is computed rather than looked up. That is the point of the file: the same walk
// answers this and a repository's `@utility` block, so a future change to PropertySort moves both
// together or fails both together, and neither can quietly keep an old answer.
func FrameworkStaticReading(name string) (Reading, bool) {
	declarations, found := FrameworkStaticDeclarations[name]
	if !found {
		return Reading{}, false
	}
	sorted := PropertySort(nodesFromStaticDeclarations(declarations))
	return Reading{Order: sorted.Order, Count: sorted.Count}, true
}

// nodesFromStaticDeclarations rebuilds the declaration nodes PropertySort walks.
//
// A flat list rather than a rule wrapping one. The generator already flattened the engine's tree in
// the order PropertySort visits it, and re-wrapping would add a container the walk would descend
// into to reach the same sequence. Both produce the same reading; the flat form is what the
// generator's output actually describes, and shaping the input to match the data avoids implying a
// nesting that was measured away.
func nodesFromStaticDeclarations(declarations []StaticDeclaration) []*Node {
	nodes := make([]*Node, 0, len(declarations))
	for _, declaration := range declarations {
		nodes = append(nodes, &Node{
			Kind:         KindDeclaration,
			Property:     declaration.Property,
			Value:        declaration.Value,
			ValuePresent: declaration.ValuePresent,
			Important:    declaration.Important,
		})
	}
	return nodes
}

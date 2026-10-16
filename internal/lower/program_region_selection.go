package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"slices"
)

func ProgramRegionSCC[T comparable](edges, owning map[T][]T) map[T]bool {
	index, low := map[T]int{}, map[T]int{}
	active := map[T]bool{}
	stack := []T{}
	next := 0
	result := map[T]bool{}
	var visit func(T)
	visit = func(node T) {
		next++
		index[node] = next
		low[node] = next
		stack = append(stack, node)
		active[node] = true
		for _, link := range edges[node] {
			if index[link] == 0 {
				visit(link)
				if low[link] < low[node] {
					low[node] = low[link]
				}
			} else if active[link] && index[link] < low[node] {
				low[node] = index[link]
			}
		}
		if low[node] != index[node] {
			return
		}
		component := map[T]bool{}
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			active[last] = false
			component[last] = true
			if last == node {
				break
			}
		}
		selected := false
		for each := range component {
			for _, target := range owning[each] {
				if component[target] {
					selected = true
				}
			}
		}
		if selected {
			for each := range component {
				result[each] = true
			}
		}
	}
	for node := range edges {
		if index[node] == 0 {
			visit(node)
		}
	}
	return result
}

func (f *cycleFinder) programRegionSelection() map[cycleNode]bool {
	return f.programRegionSelectionMode(true)
}
func (f *cycleFinder) programRegionSelectionMode(indexed bool) map[cycleNode]bool {
	candidates := []cycleNode{}
	for proven := range f.where {
		candidates = append(candidates, cycleNode{proven: proven})
	}
	for _, shape := range f.shapes {
		candidates = append(candidates, cycleNode{proven: shape})
	}
	for local, declared := range f.l.result.Locals {
		if declared.Captured && !declared.Global {
			candidates = append(candidates, cycleNode{cell: local + 1})
		}
	}
	slices.SortFunc(candidates, func(a, b cycleNode) int {
		if a.proven == nil {
			if b.proven == nil {
				return a.cell - b.cell
			}
			return -1
		}
		if b.proven == nil {
			return 1
		}
		return int(a.proven.Id()) - int(b.proven.Id())
	})
	var shapes *programShapeIndex
	if indexed {
		shapes = f.programShapes(candidates)
	}
	edges, owning := map[cycleNode][]cycleNode{}, map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		if node.proven != nil && f.programScalarStorage(node.proven) {
			edges[node] = nil
			owning[node] = nil
			continue
		}
		owned := f.programOwningLinks(node)
		links := append([]cycleNode{}, owned...)
		if node.proven != nil && node.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(node.proven) && !f.template(node.proven) && !f.isFunction(node.proven) {
			views := candidates
			if shapes != nil {
				views = shapes.relatedCandidates(node.proven)
			}
			for _, view := range views {
				if view == node {
					continue
				}
				// A view can identify the same storage only when it has the same
				// owning field targets. Mere assignability cannot close a cycle.
				if view.proven != nil && !f.programScalarStorage(view.proven) && f.related(node.proven, view.proven) {
					other := f.programOwningLinks(view)
					if sameProgramLinks(owned, other) {
						links = append(links, view)
					}
				}
			}
		}
		owning[node] = owned
		edges[node] = links
		queue = append(queue, links...)
	}
	selected := ProgramRegionSCC(edges, owning)

	return selected
}

// sameProgramLinks prevents an acyclic structural view from bridging unrelated
// recursive layouts. Names do not enter the membership decision.
func sameProgramLinks(a, b []cycleNode) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}
	have := map[cycleNode]int{}
	for _, n := range a {
		have[n]++
	}
	for _, n := range b {
		have[n]--
		if have[n] < 0 {
			return false
		}
	}
	return true
}

// Empty-literal never[] views must not connect scalar scratch arrays/maps to a
// recursive object component merely because they are assignable to both.
func (f *cycleFinder) programScalarStorage(proven *checker.Type) bool {
	if !(f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
		return false
	}
	var scalar func(*checker.Type) bool
	scalar = func(value *checker.Type) bool {
		if value.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range value.Types() {
				if !scalar(member) {
					return false
				}
			}
			return true
		}
		return value.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0
	}
	for _, argument := range f.graphArguments(proven) {
		if !scalar(argument) {
			return false
		}
	}
	return true
}
func (f *cycleFinder) programOwningLinks(node cycleNode) []cycleNode {
	if node.cell != 0 {
		links := []cycleNode{}
		if proven := f.l.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		for _, proven := range f.l.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}
		// A capture into one interior slot owns the entire environment allocation.
		for _, function := range f.l.result.Functions {
			found := false
			for _, local := range function.FrameEnvironment {
				found = found || local == node.cell-1
			}
			if found {
				for _, local := range function.FrameEnvironment {
					if local != node.cell-1 {
						links = append(links, cycleNode{cell: local + 1})
					}
				}
			}
		}
		return links
	}
	proven := node.proven
	if proven == nil || f.weak(proven) || f.template(proven) {
		return nil
	}
	links := []cycleNode{}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			links = append(links, cycleNode{proven: member})
		}
		return links
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	switch {
	case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
		links = append(links, f.libraryIteratorCaptures(proven)...)
	case f.isFunction(proven):
		if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range f.l.statics {
				actual := f.l.checker.GetTypeOfSymbol(symbol)
				if f.l.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
				}
			}
		}
		for _, closure := range f.l.closureRecords {
			if f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
				for _, local := range f.l.result.Functions[closure.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.graphArguments(proven) {
			links = append(links, cycleNode{proven: argument})
		}
	default:
		if f.l.isStaticType(proven) {
			if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
				links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
			}
		}
		for _, field := range f.fields(proven) {
			links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
		}
		for _, accessor := range f.l.accessorCaptures {
			if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range f.l.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	}
	return links
}

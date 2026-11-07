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

func (finder *cycleFinder) programRegionSelection() map[cycleNode]bool {
	return finder.programRegionSelectionMode(true)
}
func (finder *cycleFinder) programRegionSelectionMode(indexed bool) map[cycleNode]bool {
	candidates := []cycleNode{}
	for proven := range finder.where {
		candidates = append(candidates, cycleNode{proven: proven})
	}
	for _, shape := range finder.shapes {
		candidates = append(candidates, cycleNode{proven: shape})
	}
	for local, declared := range finder.lowering.result.Locals {
		if declared.Captured && !declared.Global {
			candidates = append(candidates, cycleNode{cell: local + 1})
		}
	}
	slices.SortFunc(candidates, func(left, right cycleNode) int {
		if left.proven == nil {
			if right.proven == nil {
				return left.cell - right.cell
			}
			return -1
		}
		if right.proven == nil {
			return 1
		}
		return int(left.proven.Id()) - int(right.proven.Id())
	})
	var shapes *programShapeIndex
	if indexed {
		shapes = finder.programShapes(candidates)
	}
	edges, owning := map[cycleNode][]cycleNode{}, map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		if node.proven != nil && finder.programScalarStorage(node.proven) {
			edges[node] = nil
			owning[node] = nil
			continue
		}
		owned := finder.programOwningLinks(node)
		links := append([]cycleNode{}, owned...)
		if node.proven != nil &&
			node.proven.Flags()&checker.TypeFlagsObject != 0 &&
			!finder.weak(node.proven) &&
			!finder.template(node.proven) &&
			!finder.isFunction(node.proven) {
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
				if view.proven != nil &&
					!finder.programScalarStorage(view.proven) &&
					finder.related(node.proven, view.proven) {
					other := finder.programOwningLinks(view)
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
	// Compiler's container rule is applied after SCC selection, to concrete
	// allocation-site types after monomorphization. Views cannot admit a container
	// whose elements/keys/values are outside the selected component.
	for node := range selected {
		if node.proven != nil && finder.programContainer(node.proven) {
			delete(selected, node)
		}
	}
	for node := range edges {
		if node.proven != nil && finder.programContainerMember(node.proven, selected) {
			selected[node] = true
		}
	}
	return selected
}

func (finder *cycleFinder) programContainer(proven *checker.Type) bool {
	return finder.lowering.isArrayType(proven) ||
		checker.IsTupleType(proven) ||
		finder.lowering.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")
}

// Container chains and unions can connect to a selected graph without being an
// SCC themselves. An unknown generic argument supplies no positive evidence, but
// a concrete selected Map key does (Node keys are held strongly).
func (finder *cycleFinder) programContainerMember(proven *checker.Type, selected map[cycleNode]bool) bool {
	if !finder.programContainer(proven) {
		return false
	}
	active := map[*checker.Type]bool{}
	var member func(*checker.Type) bool
	member = func(value *checker.Type) bool {
		if active[value] {
			return false
		}
		active[value] = true
		defer delete(active, value)
		if value.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, part := range value.Types() {
				if member(part) {
					return true
				}
			}
			return false
		}
		if finder.programContainer(value) {
			for _, argument := range finder.graphArguments(value) {
				if member(argument) {
					return true
				}
			}
			return false
		}
		return !finder.template(value) && !finder.weak(value) && selected[cycleNode{proven: value}]
	}
	return member(proven)
}

// sameProgramLinks prevents an acyclic structural view from bridging unrelated
// recursive layouts. Names do not enter the membership decision.
func sameProgramLinks(left, right []cycleNode) bool {
	if len(left) == 0 || len(left) != len(right) {
		return false
	}
	have := map[cycleNode]int{}
	for _, node := range left {
		have[node]++
	}
	for _, node := range right {
		have[node]--
		if have[node] < 0 {
			return false
		}
	}
	return true
}

// Empty-literal never[] views must not connect scalar scratch arrays/maps to a
// recursive object component merely because they are assignable to both.
func (finder *cycleFinder) programScalarStorage(proven *checker.Type) bool {
	if !(finder.lowering.isArrayType(proven) ||
		checker.IsTupleType(proven) ||
		finder.lowering.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
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
	for _, argument := range finder.graphArguments(proven) {
		if !scalar(argument) {
			return false
		}
	}
	return true
}
func (finder *cycleFinder) programOwningLinks(node cycleNode) []cycleNode {
	if node.cell != 0 {
		links := []cycleNode{}
		if proven := finder.lowering.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		for _, proven := range finder.lowering.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}
		// A capture into one interior slot owns the entire environment allocation.
		for _, function := range finder.lowering.result.Functions {
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
	if proven == nil || finder.weak(proven) || finder.template(proven) {
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
	case finder.lowering.isLibraryType(proven, "MapIterator", "SetIterator"):
		links = append(links, finder.libraryIteratorCaptures(proven)...)
	case finder.isFunction(proven):
		if len(finder.lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range finder.lowering.statics {
				actual := finder.lowering.checker.GetTypeOfSymbol(symbol)
				if finder.lowering.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
				}
			}
		}
		for _, closure := range finder.lowering.closureRecords {
			if finder.lowering.checker.IsTypeAssignableTo(closure.proven, proven) {
				for _, local := range finder.lowering.result.Functions[closure.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	case finder.lowering.isArrayType(proven) ||
		checker.IsTupleType(proven) ||
		finder.lowering.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range finder.graphArguments(proven) {
			links = append(links, cycleNode{proven: argument})
		}
	default:
		if finder.lowering.isStaticType(proven) {
			if parent := finder.lowering.staticBase(finder.lowering.staticClass(proven, nil)); parent != nil {
				links = append(
					links,
					cycleNode{proven: finder.lowering.checker.GetTypeOfSymbol(finder.lowering.symbol(parent.Name()))},
				)
			}
		}
		for _, field := range finder.fields(proven) {
			links = append(links, cycleNode{proven: finder.lowering.checker.GetTypeOfSymbol(field)})
		}
		for _, accessor := range finder.lowering.accessorCaptures {
			if finder.lowering.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range finder.lowering.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	}
	return links
}

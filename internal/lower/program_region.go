package lower

// Program membership belongs to lowering. Runtime only manages its lifetime.

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// ProgramRegionSCC selects the strongly connected part containing an owning
// field/container/capture edge. Structural view edges connect possible identities,
// but a cycle of views alone is not a cycle of owned fields. The census adapter
// calls this same selection algorithm; it is not a second membership heuristic.
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

func (f *cycleFinder) programRegionTypes(modules []*ast.SourceFile) error {
	// Observe allocation and contextual identities even where all writes passed the
	// fresh proof had no refusal seeds (e.g. Weak trees).
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if ast.IsExpressionNode(node) {
				f.use(f.l.checker.GetTypeAtLocation(node), node)
				if contextual := f.l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
					f.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range f.l.closureRecords {
		f.use(closure.proven, closure.node)
	}
	for proven, node := range f.l.programAllocationTypes {
		f.use(proven, node)
	}
	selected := f.programRegionSelection()
	program := f.l.result
	program.ProgramTypes = map[int]bool{}
	for node := range selected {
		if node.proven != nil && !f.weak(node.proven) && !f.l.isLibraryType(node.proven, "Promise") {
			program.ProgramTypes[int(node.proven.Id())] = true
		}
		if node.cell != 0 {
			program.Locals[node.cell-1].ProgramRegion = true
		}
	}
	for _, closure := range f.l.closureRecords {
		program.Functions[closure.function].ProgramRegion = program.ProgramTypes[int(closure.proven.Id())]
	}
	// Capture environments are indivisible allocations; retaining an extra sibling
	// costs retention, while treating an interior cell as independently freeable is unsafe.
	for changed := true; changed; {
		changed = false
		for i := range program.Functions {
			function := &program.Functions[i]
			for _, local := range function.Environment {
				if program.Locals[local].ProgramRegion && !function.ProgramRegion {
					function.ProgramRegion = true
					changed = true
				}
			}
			if function.ProgramRegion {
				for _, local := range function.Environment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || program.Locals[local].ProgramRegion
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
		}
	}
	for _, instance := range f.l.instances {
		for _, local := range instance.thisLocals {
			for _, proven := range append([]*checker.Type{f.l.localTypes[local]}, f.l.localAlso[local]...) {
				if proven != nil && program.ProgramTypes[int(proven.Id())] {
					program.Classes[instance.class-1].ProgramRegion = true
				}
			}
		}
	}
	for symbol, instance := range f.l.statics {
		if program.ProgramTypes[int(f.l.checker.GetTypeOfSymbol(symbol).Id())] {
			program.Classes[instance.class-1].ProgramRegion = true
		}
	}

	if len(program.ProgramTypes) != 0 {
		parallel := false
		check := func(node any) bool {
			if _, ok := node.(ir.ParallelMap); ok {
				parallel = true
			}
			return true
		}
		walk(program.Main, check)
		for _, function := range program.Functions {
			walk(function.Body, check)
		}
		if parallel {
			return &Refused{Where: f.l.program.Where(modules[0].AsNode()), What: "parallelMap in a program with Program-region members", Fix: "keep Program members in the one-shot CLI thread; the Program region does not allocate or publish members across tasks"}
		}
	}
	return nil
}

// programRegionSelection is shared by executable lowering and the pinned census audit.
func (f *cycleFinder) programRegionSelection() map[cycleNode]bool {
	return f.selectProgramRegion(true)
}

func (f *cycleFinder) selectProgramRegion(indexed bool) map[cycleNode]bool {
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
	views := newProgramShapeIndex(f, candidates)
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
		// Record actual owned slots separately from possible structural views.
		owned := f.programOwnedLinks(node)
		links := append([]cycleNode{}, owned...)
		if node.proven != nil && node.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(node.proven) && !f.template(node.proven) && !f.isFunction(node.proven) {
			for _, view := range views.candidates(node.proven, indexed) {
				if view != node && view.proven != nil && view.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(view.proven) && !f.template(view.proven) && !f.isFunction(view.proven) && !f.programScalarStorage(view.proven) && f.related(node.proven, view.proven) {
					links = append(links, view)
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
		if node.proven != nil && f.programContainer(node.proven) {
			delete(selected, node)
		}
	}
	for node := range edges {
		if node.proven != nil && f.programContainerMember(node.proven, selected) {
			selected[node] = true
		}
	}
	return selected
}

const programContainerRule = "compiler container rule (allocation-site type after monomorphization): element, Map key/value, Set element or tuple member in selected component"

func (f *cycleFinder) programContainer(proven *checker.Type) bool {
	return f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")
}

// Container chains and unions can connect to a selected graph without being an
// SCC themselves. An unknown generic argument supplies no positive evidence, but
// a concrete selected Map key does (Node keys are held strongly).
func (f *cycleFinder) programContainerMember(proven *checker.Type, selected map[cycleNode]bool) bool {
	if !f.programContainer(proven) {
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
		if f.programContainer(value) {
			for _, argument := range f.l.checker.GetTypeArguments(value) {
				if member(argument) {
					return true
				}
			}
			return false
		}
		return !f.template(value) && !f.weak(value) && selected[cycleNode{proven: value}]
	}
	return member(proven)
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
	for _, argument := range f.l.checker.GetTypeArguments(proven) {
		if !scalar(argument) {
			return false
		}
	}
	return true
}

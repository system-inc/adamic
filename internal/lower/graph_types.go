package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// graphAllocation preserves both the allocation's concrete type and its contextual
// view. Ownership is selected only after the whole program's cycle proof runs.
func (l *lowering) graphAllocation(value ir.Expression, node *ast.Node) ir.Expression {
	ids := []int{int(l.concrete(l.checker.GetTypeAtLocation(node)).Id())}
	// Fresh literal identities may change across checker queries. The variable's
	// widened allocation view is stable and is the type the cycle finder visits.
	if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration && node.Parent.Name() != nil {
		ids = append(ids, int(l.concrete(l.checker.GetTypeAtLocation(node.Parent.Name())).Id()))
	}
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
		ids = append(ids, int(l.concrete(contextual).Id()))
	}
	switch value := value.(type) {
	case ir.ObjectLiteral:
		value.GraphTypes = ids
		return value
	case ir.ArrayLiteral:
		value.GraphTypes = ids
		return value
	case ir.MapNew:
		value.GraphTypes = ids
		return value
	case ir.SetNew:
		value.GraphTypes = ids
		return value
	case ir.ArrayMap:
		value.GraphTypes = ids
		return value
	case ir.ArrayFrom:
		value.GraphTypes = ids
		return value
	case ir.ArrayFill:
		value.GraphTypes = ids
		return value
	case ir.ArraySplice:
		value.GraphTypes = ids
		return value
	case ir.ArrayConcat:
		value.GraphTypes = ids
		return value
	case ir.MapEntries:
		value.GraphTypes = ids
		return value
	case ir.ArraySlice:
		value.GraphTypes = ids
		return value
	case ir.MapKeys:
		value.GraphTypes = ids
		return value
	case ir.MapValues:
		value.GraphTypes = ids
		return value
	case ir.SetValues:
		value.GraphTypes = ids
		return value
	case ir.ArrayVisit:
		value.GraphTypes = ids
		return value
	}
	return value
}

// The old refusal sites still determine the seeds. The ownership SCC includes
// all strong connecting paths, including readonly and proven-fresh slots.
func (f *cycleFinder) graphTypes(modules []*ast.SourceFile) error {
	seeds := map[cycleNode]bool{}
	for _, proven := range f.seen {
		if err := f.slotsOf(proven); err != nil {
			seeds[cycleNode{proven: proven}] = true
		}
	}
	for local, declared := range f.l.result.Locals {
		proven := f.l.localTypes[local]
		if declared.Captured && !declared.Global && proven != nil && !f.weak(proven) && f.reaches(proven, cycleNode{cell: local + 1}) {
			// An async frame owns its captured cells and is never a graph member, so a cycle through
			// one stays refused, as it was before graph regions.
			if declared.Function >= 0 && f.l.result.Functions[declared.Function].Async {
				return &Refused{
					Where: f.l.program.Where(f.l.localNodes[local]),
					What:  "async frame capture cycle: '" + declared.Name + "', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free",
					Fix:   "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)",
				}
			}
			seeds[cycleNode{cell: local + 1}] = true
		}
	}
	if len(seeds) == 0 {
		return nil
	}
	// Literals may have narrower types than the mutable view through which they
	// are later written. Include those allocation identities without creating seeds.
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
	edges := map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		links := f.graphLinks(node)
		// The finder's related relation treats compatible structural views as the
		// same possible allocation. Weak views never become strong edges.
		if node.proven != nil && node.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(node.proven) && !f.template(node.proven) && !f.isFunction(node.proven) {
			for _, view := range candidates {
				if view.proven != nil && view.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(view.proven) && !f.template(view.proven) && f.related(node.proven, view.proven) {
					links = append(links, view)
				}
			}
		}
		edges[node] = links
		queue = append(queue, links...)
	}
	// Tarjan's components, with a component selected iff it contains a seed.
	next := 0
	index, low := map[cycleNode]int{}, map[cycleNode]int{}
	active := map[cycleNode]bool{}
	stack := []cycleNode{}
	graph := map[cycleNode]bool{}
	var visit func(cycleNode)
	visit = func(node cycleNode) {
		next++
		index[node], low[node] = next, next
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
		component := []cycleNode{}
		selected := false
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			active[last] = false
			component = append(component, last)
			selected = selected || seeds[last]
			if last == node {
				break
			}
		}
		if selected {
			for _, each := range component {
				graph[each] = true
			}
		}
	}
	for node := range edges {
		if index[node] == 0 {
			visit(node)
		}
	}
	// An otherwise acyclic cache of graph elements belongs inside their graph too.
	for changed := true; changed; {
		changed = false
		for node := range edges {
			if graph[node] || node.proven == nil || f.weak(node.proven) {
				continue
			}
			proven := node.proven
			if f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
				for _, argument := range f.l.checker.GetTypeArguments(proven) {
					if graph[cycleNode{proven: argument}] {
						graph[node] = true
						changed = true
					}
					if argument.Flags()&checker.TypeFlagsUnion != 0 {
						for _, member := range argument.Types() {
							if graph[cycleNode{proven: member}] {
								graph[node] = true
								changed = true
							}
						}
					}
				}
			}
		}
	}
	f.l.result.GraphTypes = map[int]bool{}
	for node := range graph {
		if node.proven != nil && !f.weak(node.proven) {
			f.l.result.GraphTypes[int(node.proven.Id())] = true
		}
		if node.cell != 0 {
			f.l.result.Locals[node.cell-1].GraphCell = true
		}
	}
	for _, closure := range f.l.closureRecords {
		if f.l.result.GraphTypes[int(closure.proven.Id())] {
			f.l.result.Functions[closure.function].GraphClosure = true
		}
	}
	// One frame allocation is indivisible, even if siblings capture disjoint slots.
	for changed := true; changed; {
		changed = false
		for index := range f.l.result.Functions {
			function := &f.l.result.Functions[index]
			for _, local := range function.Environment {
				if f.l.result.Locals[local].GraphCell && !function.GraphClosure {
					function.GraphClosure = true
					changed = true
				}
			}
			if function.GraphClosure {
				for _, local := range function.Environment {
					if !f.l.result.Locals[local].GraphCell {
						f.l.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || f.l.result.Locals[local].GraphCell
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !f.l.result.Locals[local].GraphCell {
						f.l.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
		}
	}
	for _, instance := range f.l.instances {
		for _, local := range instance.thisLocals {
			types := append([]*checker.Type{f.l.localTypes[local]}, f.l.localAlso[local]...)
			for _, proven := range types {
				if proven != nil && f.l.result.GraphTypes[int(proven.Id())] {
					f.l.result.Classes[instance.class-1].Graph = true
				}
			}
		}
	}
	for symbol, instance := range f.l.statics {
		if f.l.result.GraphTypes[int(f.l.checker.GetTypeOfSymbol(symbol).Id())] {
			f.l.result.Classes[instance.class-1].Graph = true
		}
	}
	f.graphFlows()
	return nil
}

func (f *cycleFinder) graphLinks(node cycleNode) []cycleNode {
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
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
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
		for _, shape := range f.shapes {
			if shape != proven && f.l.checker.IsTypeAssignableTo(shape, proven) {
				links = append(links, cycleNode{proven: shape})
			}
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

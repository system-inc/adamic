package lower

import (
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// graphAllocation preserves both the allocation's concrete type and its contextual
// view. Ownership is selected only after the whole program's cycle proof runs.
func (lowering *lowering) graphAllocation(value ir.Expression, node *ast.Node) ir.Expression {
	ids := []int{lowering.cycleTypeID(lowering.concrete(lowering.checker.GetTypeAtLocation(node)))}
	// Fresh literal identities may change across checker queries. The variable's
	// widened allocation view is stable and is the type the cycle finder visits.
	if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration && node.Parent.Name() != nil {
		ids = append(
			ids,
			lowering.cycleTypeID(lowering.concrete(lowering.checker.GetTypeAtLocation(node.Parent.Name()))),
		)
	}
	if contextual := lowering.slotType(node); contextual != nil {
		ids = append(ids, lowering.cycleTypeID(lowering.concrete(contextual)))
	}
	switch value := value.(type) {
	case ir.ObjectLiteral:
		if value.Record {
			lowering.graphDynamicSite = node
		}
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

// graphSeed accepts only the slot-cycle proof; unrelated errors keep their identity.
func graphSeed(sourceError error) (bool, error) {
	var refused *Refused
	if errors.As(sourceError, &refused) && refused != nil && refused.cycle {
		return true, nil
	}
	return false, sourceError
}

// The old refusal sites still determine the seeds. The ownership SCC includes
// all strong connecting paths, including readonly and proven-fresh slots.
func (finder *cycleFinder) graphTypes(modules []*ast.SourceFile) error {
	// Graph regions don't follow record values, so a cycle through a record stays refused rather than seeded.
	for _, proven := range finder.seen {
		if finder.lowering.recordElement(proven) != nil {
			if slotsOfError := finder.slotsOf(proven); slotsOfError != nil {
				return slotsOfError
			}
		}
	}
	seeds := map[cycleNode]bool{}
	for _, proven := range finder.seen {
		seed, graphSeedError := graphSeed(finder.slotsOf(proven))
		if graphSeedError != nil {
			return graphSeedError
		}
		if seed {
			seeds[cycleNode{proven: proven}] = true
		}
	}
	for local, declared := range finder.lowering.result.Locals {
		proven := finder.lowering.localTypes[local]
		// A closed frame input (closed_frame_inputs.go) is proven not to close a cycle, so its cell
		// stays counted rather than seeding a graph region.
		if declared.Captured &&
			!declared.Global &&
			proven != nil &&
			!finder.weak(proven) &&
			finder.reaches(proven, cycleNode{cell: local + 1}) &&
			!finder.lowering.closedFrameInput(local) {
			// An async frame owns its captured cells and is never a graph member, so a cycle through
			// one stays refused, as it was before graph regions.
			if declared.Function >= 0 && finder.lowering.result.Functions[declared.Function].Suspends() {
				return &Refused{
					Where: finder.lowering.program.Where(finder.lowering.localNodes[local]),
					What: "async frame capture cycle: '" +
						declared.Name +
						"', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free",
					Fix: "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)",
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
				finder.use(finder.lowering.checker.GetTypeAtLocation(node), node)
				if contextual := finder.lowering.slotType(node); contextual != nil {
					finder.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range finder.lowering.closureRecords {
		finder.use(closure.proven, closure.node)
	}
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
	edges := map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		links := finder.graphLinks(node)
		// The finder's related relation treats compatible structural views as the
		// same possible allocation. Weak views never become strong edges.
		if node.proven != nil &&
			node.proven.Flags()&checker.TypeFlagsObject != 0 &&
			!finder.weak(node.proven) &&
			!finder.template(node.proven) &&
			!finder.isFunction(node.proven) {
			for _, view := range candidates {
				if view.proven != nil &&
					view.proven.Flags()&checker.TypeFlagsObject != 0 &&
					!finder.weak(view.proven) &&
					!finder.template(view.proven) &&
					finder.related(node.proven, view.proven) {
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
	var throughPromise []cycleNode
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
				if each.proven != nil &&
					finder.lowering.isLibraryType(each.proven, "Promise") &&
					throughPromise == nil {
					throughPromise = component
				}
			}
		}
	}
	for node := range edges {
		if index[node] == 0 {
			visit(node)
		}
	}
	// A region frees a cycle only when every value on it is a member. A Promise never is (its
	// payload is a counted value outside any region), so a cycle through one would keep its region
	// alive forever: refuse it, as before graph regions.
	if throughPromise != nil {
		var where *ast.Node
		for _, each := range throughPromise {
			if each.proven != nil && finder.where[each.proven] != nil {
				where = finder.where[each.proven]
				break
			}
		}
		if where == nil {
			where = modules[0].AsNode()
		}
		return &Refused{
			Where: finder.lowering.program.Where(where),
			What:  "a Promise whose payload can reach back to what holds it: a cycle reference counting can't free, and a Promise can't join a graph region",
			Fix:   "don't keep a Promise in a value its result can reach, or declare the field Weak<...> (adamic/cycle-capable)",
		}
	}
	// An otherwise acyclic cache of graph elements belongs inside their graph too.
	for changed := true; changed; {
		changed = false
		for node := range edges {
			if graph[node] || node.proven == nil || finder.weak(node.proven) {
				continue
			}
			proven := node.proven
			if finder.lowering.isArrayType(proven) ||
				checker.IsTupleType(proven) ||
				finder.lowering.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
				for _, argument := range finder.graphArguments(proven) {
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
	finder.lowering.result.GraphTypes = map[int]bool{}
	for node := range graph {
		if node.proven != nil && !finder.weak(node.proven) {
			finder.lowering.result.GraphTypes[finder.lowering.cycleTypeID(node.proven)] = true
		}
		if node.cell != 0 {
			finder.lowering.result.Locals[node.cell-1].GraphCell = true
		}
	}
	for _, closure := range finder.lowering.closureRecords {
		if finder.lowering.result.GraphTypes[int(closure.proven.Id())] {
			finder.lowering.result.Functions[closure.function].GraphClosure = true
		}
	}
	// One frame allocation is indivisible, even if siblings capture disjoint slots.
	for changed := true; changed; {
		changed = false
		for index := range finder.lowering.result.Functions {
			function := &finder.lowering.result.Functions[index]
			for _, local := range function.Environment {
				if finder.lowering.result.Locals[local].GraphCell && !function.GraphClosure {
					function.GraphClosure = true
					changed = true
				}
			}
			if function.GraphClosure {
				for _, local := range function.Environment {
					if !finder.lowering.result.Locals[local].GraphCell {
						finder.lowering.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || finder.lowering.result.Locals[local].GraphCell
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !finder.lowering.result.Locals[local].GraphCell {
						finder.lowering.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
		}
	}
	// Forwarded nested callbacks retain their parent frame through a path outside
	// the direct environment ownership proof. Keep that capture cycle refused.
	for _, closure := range finder.lowering.closureRecords {
		function := finder.lowering.result.Functions[closure.function]
		if function.GraphClosure && function.ForwardedNestedParent != 0 {
			return &Refused{
				Where: finder.lowering.program.Where(closure.node),
				What:  "a nested callback retains a cyclic parent environment through a forwarded frame",
				Fix:   "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> (adamic/cycle-capable)",
			}
		}
	}
	for _, instance := range finder.lowering.instances {
		for _, local := range instance.thisLocals {
			types := append([]*checker.Type{finder.lowering.localTypes[local]}, finder.lowering.localAlso[local]...)
			for _, proven := range types {
				if proven != nil && finder.lowering.result.GraphTypes[int(proven.Id())] {
					finder.lowering.result.Classes[instance.class-1].Graph = true
				}
			}
		}
	}
	for symbol, instance := range finder.lowering.statics {
		if finder.lowering.result.GraphTypes[int(finder.lowering.checker.GetTypeOfSymbol(symbol).Id())] {
			finder.lowering.result.Classes[instance.class-1].Graph = true
		}
	}
	finder.graphFlows()
	return finder.graphAdoptionBoundary()
}

func (finder *cycleFinder) graphLinks(node cycleNode) []cycleNode {
	if node.cell != 0 {
		links := []cycleNode{}
		if proven := finder.lowering.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		for _, proven := range finder.lowering.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}
		if finder.lowering.result.Locals[node.cell-1].IterationCell {
			return links
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
		for _, shape := range finder.shapes {
			if shape != proven && finder.lowering.checker.IsTypeAssignableTo(shape, proven) {
				links = append(links, cycleNode{proven: shape})
			}
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

// graphArguments only queries the checker on reference objects. A contextual
// union may describe an array initializer without itself being a TypeReference.
func (finder *cycleFinder) graphArguments(proven *checker.Type) []*checker.Type {
	if proven == nil ||
		proven.Flags()&checker.TypeFlagsObject == 0 ||
		proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return nil
	}
	return finder.lowering.checker.GetTypeArguments(proven)
}

// Graph adoption moves a fixed allocation. Checked-view adapters and dynamic
// shapes have additional owners outside that proof. Until their provenance is
// tracked separately, keep mixed programs outside graph ownership.
func (finder *cycleFinder) graphAdoptionBoundary() error {
	reason := ""
	var site *ast.Node
	// Legacy tagged-union views record checked fields without a ViewOrigin.
	if len(finder.lowering.result.ViewOrigins) != 0 || len(finder.lowering.result.CheckedFields) != 0 {
		reason = "a checked-view adapter in a graph program"
		site = finder.lowering.graphViewSite
	}
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if value.Record {
				reason = "a dynamic-shape object in a graph program"
				site = finder.lowering.graphDynamicSite
			}
		case ir.DynamicProperty, ir.DynamicKeyedProperty:
			reason = "a dynamic-shape object in a graph program"
			site = finder.lowering.graphDynamicSite
		case ir.ObjectCall:
			if value.Checked {
				reason = "a dynamic-shape object in a graph program"
				site = finder.lowering.graphDynamicSite
			}
		}
		return true
	}
	walk(finder.lowering.result.Main, inspect)
	for _, function := range finder.lowering.result.Functions {
		walk(function.Body, inspect)
	}
	if reason != "" {
		return &Refused{
			Where: finder.lowering.program.Where(site),
			What:  reason,
			Fix:   "keep checked views and dynamic shapes outside graph adoption (adamic/cycle-capable)",
		}
	}
	return nil
}

// cycleTypeID keeps the checker's allocation identity unchanged. The optional
// capture lets differential tests resolve even transient allocation types.
func (lowering *lowering) cycleTypeID(proven *checker.Type) int {
	id := int(proven.Id())
	if lowering.cycleTypes != nil {
		lowering.cycleTypes[id] = proven
	}
	return id
}

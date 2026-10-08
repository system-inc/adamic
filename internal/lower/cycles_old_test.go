package lower

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// oldCycleFinder freezes 6c891cca313e614ae46de794814bc7c0db4cae01's finder,
// ownership graph and fresh-write helpers as the differential oracle. The query
// wrappers compare all answers, not just those that happen to select a seed.
type oldCycleFinder struct {
	compare     *cycleFinder
	differences []string
	l           *lowering

	// seen is every type the program uses, in the order first met, with where it was met.
	seen  []*checker.Type
	where map[*checker.Type]*ast.Node

	// shapes is every seen object type that isn't an array, a map or a function: what a value seen as
	// some object type may really be. shaped are the literals' own types among them.
	shapes []*checker.Type
	shaped map[*checker.Type]bool

	// writes is every write in the program, judged by the relaxation for fresh writes (fresh.go),
	// worked out the first time a slot needs it.
	writes []fresh.Write

	libraryIterators []libraryIteratorCapture
}

func (l *lowering) oldCycles(modules []*ast.SourceFile) (*oldCycleFinder, error) {
	// Generated frame/Promise/reaction layouts are IR identities, never checker declarations.
	// Only their runtime-owned protocol edges are exempt. Source types, including a class
	// called adamic_async_frame, continue through slotsOf with no name-based escape hatch.
	if l.result.HasAsync() {
		for _, generated := range l.result.Generated {
			if !fresh.RuntimeBreaksCycles(generated) {
				return nil, fmt.Errorf("lower: unaudited generated cycle identity")
			}
		}
	}
	finder := &oldCycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			finder.libraryIteratorMade(node)
			switch {
			case node.Kind == ast.KindObjectLiteralExpression:
				// A literal is what a value may really be, and nothing is written through its own
				// type: it's seen as the type it's written into, or kept in a variable whose type
				// is used where it's declared.
				finder.shape(l.checker.GetTypeAtLocation(node), node)
			case node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression:
				// A value just made, by a literal, a call or new, is written through only once it's
				// kept somewhere, and where it's kept has a type of its own, used there. Until then
				// it's what a value may really be.
				finder.made(l.checker.GetTypeAtLocation(node), node)
			case ast.IsIdentifier(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression:
				// Where a value is named or reached, and so where a write can go through its type.
				if ast.IsIdentifier(node) && !ast.IsDeclarationName(node) && !ast.IsExpressionNode(node) {
					break
				}
				finder.use(l.checker.GetTypeAtLocation(node), node)
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	// Every instantiation lowering made, Box<Node> inside Maker<Node> included, which the source may
	// never write.
	for _, proven := range l.instantiated {
		finder.use(proven, l.classNodeFor(proven))
	}
	if os.Getenv("CYCLES_OLD") == "" {
		finder.compare = &cycleFinder{l: l, seen: finder.seen, where: finder.where, shapes: finder.shapes, shaped: finder.shaped, libraryIterators: finder.libraryIterators}
		finder.compareAllSlots()
	}
	err := finder.graphTypes(modules)
	if len(finder.differences) != 0 {
		return finder, fmt.Errorf("finder queries differ: %v", finder.differences)
	}
	if finder.compare != nil && !reflect.DeepEqual(finder.writes, finder.compare.writes) {
		return finder, fmt.Errorf("fresh write results differ: old=%v new=%v", finder.writes, finder.compare.writes)
	}
	return finder, err
}

func (l *lowering) oldFindCycles(modules []*ast.SourceFile) error {
	_, err := l.oldCycles(modules)
	return err
}

func (f *oldCycleFinder) made(proven *checker.Type, where *ast.Node) {
	switch {
	case proven == nil:
	case proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
		for _, member := range proven.Types() {
			f.made(member, where)
		}
	case proven.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(proven):
		f.use(proven, where)
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.use(argument, where)
		}
	default:
		f.shape(proven, where)
	}
}

func (f *oldCycleFinder) shape(proven *checker.Type, where *ast.Node) {
	if f.shaped[proven] {
		return
	}
	if f.shaped == nil {
		f.shaped = map[*checker.Type]bool{}
	}
	f.shaped[proven] = true
	f.shapes = append(f.shapes, proven)
	for _, field := range f.fields(proven) {
		f.use(f.l.checker.GetTypeOfSymbol(field), declaredAt(field, where))
	}
}

func (f *oldCycleFinder) use(proven *checker.Type, where *ast.Node) {
	if proven == nil {
		return
	}
	if _, isSeen := f.where[proven]; isSeen || f.template(proven) {
		return
	}
	f.where[proven] = where
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
		// An object literal's own type, as the literal expression has it. Nothing is written through
		// it: written into a declared type, the object is seen only as that, and kept in a variable
		// of its own, the variable's type is the literal's regular type, which is used where it's
		// declared. It's still what a value may really be.
		f.shape(proven, where)
		return
	}
	f.seen = append(f.seen, proven)
	flags := proven.Flags()
	if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			f.use(member, where)
		}
		return
	}
	if flags&checker.TypeFlagsObject == 0 {
		return
	}
	if f.isFunction(proven) {
		return
	}
	if f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise") {
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.use(argument, where)
		}
		return
	}
	if !f.shaped[proven] {
		f.shapes = append(f.shapes, proven)
	}
	for _, field := range f.fields(proven) {
		f.use(f.l.checker.GetTypeOfSymbol(field), declaredAt(field, where))
	}
}

func (f *oldCycleFinder) template(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return false
	}
	for _, argument := range f.l.checker.GetTypeArguments(proven) {
		if argument.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return true
		}
	}
	return false
}

func (f *oldCycleFinder) fields(proven *checker.Type) []*ast.Symbol {
	fields := []*ast.Symbol{}
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		if (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property) && !(f.l.isStaticType(proven) && property.Name == "prototype") {
			fields = append(fields, property)
		}
	}
	return fields
}

func (f *oldCycleFinder) isFunction(proven *checker.Type) bool {
	return len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || (len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && (!f.l.isStaticType(proven) || len(f.l.staticGlobals) == 0))
}

func (f *oldCycleFinder) oldWeak(proven *checker.Type) bool {
	representation, _ := f.l.representation(proven)
	return representation == ir.Weak
}

func (f *oldCycleFinder) oldSlotsOf(holder *checker.Type) error {
	l := f.l
	if holder.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(holder) {
		return nil
	}
	name := l.checker.TypeToString(holder)
	switch {
	case l.checker.IsArrayType(holder) || checker.IsTupleType(holder):
		if l.isLibraryType(holder, "ReadonlyArray") || (checker.IsTupleType(holder) && holder.TargetTupleType().IsReadonly()) {
			return nil
		}
		for _, element := range l.checker.GetTypeArguments(holder) {
			if f.weak(element) || !f.reaches(element, cycleNode{proven: holder}) {
				continue
			}
			write := f.unproven(fresh.WriteElement, holder, "")
			if write == nil {
				continue
			}
			target := f.present(element)
			return &Refused{
				Where: l.program.Where(f.where[holder]),
				What:  name + ", an array whose elements can reach back to an array like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
				Fix:   "declare the elements weak, Weak<" + target + ">[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly " + target + "[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)",
			}
		}
	case l.isLibraryType(holder, "ReadonlyMap"), l.isLibraryType(holder, "ReadonlySet"):
	case l.isLibraryType(holder, "Set"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 1 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteSetElement, holder, ""); write != nil {
				return &Refused{
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a set whose elements can reach back to a set like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlySet, built whole when it's made; or add to such a set only values this function made, or only to one it made (adamic/cycle-capable)",
				}
			}
		}
	case l.isLibraryType(holder, "Map"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 2 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a map whose keys can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlyMap, built whole when it's made; or set in such a map only keys this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
		if len(arguments) == 2 && !f.weak(arguments[1]) && f.reaches(arguments[1], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a map whose values can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "declare the values weak, Map<" + l.checker.TypeToString(arguments[0]) + ", Weak<" + f.present(arguments[1]) + ">> (import type { Weak } from 'adamic'), or make it a ReadonlyMap; or set in such a map only values this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
	default:
		for _, field := range f.fields(holder) {
			// A readonly field is a slot too: a constructor writes it, after this may have been put
			// somewhere (parent.kids.push(this)), and then that write can close a cycle. Every other
			// way a readonly field gets its value (a literal, an initializer) makes the holder with it,
			// and is no write at all.
			readonly := l.checker.IsReadonlySymbol(field)
			proven := l.checker.GetTypeOfSymbol(field)
			if f.weak(proven) || !f.reaches(proven, cycleNode{proven: holder}) {
				continue
			}
			write := f.unproven(fresh.WriteField, holder, field.Name)
			if write == nil {
				continue
			}
			where := f.where[holder]
			if len(field.Declarations) > 0 {
				where = field.Declarations[0]
			}
			target := l.checker.TypeToString(proven)
			kind, orReadonly := "a mutable field", "; or make it readonly"
			if readonly {
				kind, orReadonly = "a readonly field, which its constructor writes,", ""
			}
			return &Refused{
				Where: l.program.Where(where),
				What:  name + "." + l.cycleFieldName(field) + ", " + kind + " of type " + target + ", which can reach back to the " + name + " holding it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
				Fix:   "declare it " + l.cycleFieldName(field) + ": Weak<" + f.present(proven) + "> (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed" + orReadonly + "; or write into it only values this function made, or only into what it made (adamic/cycle-capable)",
			}
		}
	}
	return nil
}

func (f *oldCycleFinder) oldReaches(from *checker.Type, target cycleNode) bool {
	visited := map[cycleNode]bool{}
	queue := []cycleNode{{proven: from}}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if visited[node] {
			continue
		}
		visited[node] = true
		if node.cell != 0 {
			if node == target {
				return true
			}
			// An interior async cell retains its owner, not just its declared value.
			// Completion drops private slots; captured slots remain reachable through
			// any escaped cell. Follow every such slot irrespective of signature.
			local := f.l.result.Locals[node.cell-1]
			if local.EnvironmentCell && local.Function >= 0 {
				owner := f.l.result.Functions[local.Function]
				if owner.Async {
					for _, held := range owner.FrameEnvironment {
						if f.l.result.Locals[held].Captured {
							queue = append(queue, cycleNode{cell: held + 1})
						}
					}
				}
			}
			if proven := f.l.localTypes[node.cell-1]; proven != nil {
				queue = append(queue, cycleNode{proven: proven})
			}
			// A this shared by every instantiation of a class held the same way is each of them.
			for _, proven := range f.l.localAlso[node.cell-1] {
				queue = append(queue, cycleNode{proven: proven})
			}
			continue
		}
		proven := node.proven
		flags := proven.Flags()
		if f.weak(proven) || f.template(proven) {
			// A Weak holds nothing.
			continue
		}
		if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				queue = append(queue, cycleNode{proven: member})
			}
			continue
		}
		if flags&checker.TypeFlagsObject == 0 {
			continue
		}
		if target.proven != nil && f.related(proven, target.proven) {
			return true
		}
		switch {
		case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
			queue = append(queue, f.libraryIteratorCaptures(proven)...)
		case f.isFunction(proven):
			// Construct signatures can hide constructor objects behind an interface.
			if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
				for symbol := range f.l.statics {
					actual := f.l.checker.GetTypeOfSymbol(symbol)
					if f.l.checker.IsTypeAssignableTo(actual, proven) {
						queue = append(queue, cycleNode{proven: actual})
					}
				}
			}
			// What a function value holds is what it captured: the cells of every function value the
			// program makes that can be seen as this type.
			for _, closure := range f.l.closureRecords {
				if !f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
					continue
				}
				for _, local := range f.l.result.Functions[closure.function].Environment {
					queue = append(queue, cycleNode{cell: local + 1})
				}
			}
		case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
			for _, argument := range f.l.checker.GetTypeArguments(proven) {
				queue = append(queue, cycleNode{proven: argument})
			}
		default:
			if f.l.isStaticType(proven) {
				if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
					queue = append(queue, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
				}
			}
			for _, accessor := range f.l.accessorCaptures {
				if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
					for _, local := range f.l.result.Functions[accessor.function].Environment {
						queue = append(queue, cycleNode{cell: local + 1})
					}
				}
			}
			for _, field := range f.fields(proven) {
				queue = append(queue, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
			}
			// A value seen as this type may be any object type the program has that can be seen as
			// it, with fields this type doesn't show.
			for _, shape := range f.shapes {
				if shape != proven && f.l.checker.IsTypeAssignableTo(shape, proven) {
					queue = append(queue, cycleNode{proven: shape})
				}
			}
		}
	}
	return false
}

func (f *oldCycleFinder) present(proven *checker.Type) string {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return f.l.checker.TypeToString(proven)
	}
	members := []string{}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 {
			members = append(members, f.l.checker.TypeToString(member))
		}
	}
	return strings.Join(members, " | ")
}

func (f *oldCycleFinder) related(one *checker.Type, other *checker.Type) bool {
	if one == other {
		return true
	}
	if f.isFunction(one) || f.isFunction(other) {
		return false
	}
	return f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)
}

func (f *oldCycleFinder) graphTypes(modules []*ast.SourceFile) error {
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
				if each.proven != nil && f.l.isLibraryType(each.proven, "Promise") && throughPromise == nil {
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
			if each.proven != nil && f.where[each.proven] != nil {
				where = f.where[each.proven]
				break
			}
		}
		if where == nil {
			where = modules[0].AsNode()
		}
		return &Refused{
			Where: f.l.program.Where(where),
			What:  "a Promise whose payload can reach back to what holds it: a cycle reference counting can't free, and a Promise can't join a graph region",
			Fix:   "don't keep a Promise in a value its result can reach, or declare the field Weak<...> (adamic/cycle-capable)",
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

func (f *oldCycleFinder) graphLinks(node cycleNode) []cycleNode {
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

func (f *oldCycleFinder) graphFlows() {
	program := f.l.result
	next := 0
	program.Main = graphAllocationSites(reflect.ValueOf(program.Main), &next).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = graphAllocationSites(reflect.ValueOf(program.Functions[i].Body), &next).Interface().([]ir.Statement)
	}

	var graphType func(*checker.Type) bool
	graphType = func(proven *checker.Type) bool {
		if proven == nil || f.weak(proven) {
			return false
		}
		if program.GraphTypes[int(proven.Id())] {
			return true
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if graphType(member) {
					return true
				}
			}
		}
		return false
	}
	// Positive nodes are locals followed by function results.
	resultNode := func(function int) int { return len(program.Locals) + function + 1 }
	sources := map[int][]ir.Expression{}
	wanted := map[int]bool{}
	queue := []int{}
	demand := func(node int) {
		if !wanted[node] {
			wanted[node] = true
			queue = append(queue, node)
		}
	}
	for local, proven := range f.l.localTypes {
		if graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if graphType(proven) {
				demand(local + 1)
			}
		}
	}
	pending := []ir.Expression{}
	// Site retains the checker view of the actual slot being written.
	slot := func(site int, name string, argument int) bool {
		if site <= 0 || site > len(f.l.writeSites) {
			return false
		}
		holder := f.l.writeSites[site-1].holder
		if holder == nil {
			return false
		}
		if name != "" {
			for _, field := range f.fields(holder) {
				if f.l.cycleFieldMatches(holder, field.Name, name) {
					return graphType(f.l.checker.GetTypeOfSymbol(field))
				}
			}
			return false
		}
		args := f.l.checker.GetTypeArguments(holder)
		return argument < len(args) && graphType(args[argument])
	}
	// A counted aggregate can have graph-typed fields/elements. Seed those
	// initializer slots individually, rather than promoting its whole payload.
	returns := map[int]*checker.Type{}
	for symbol, index := range f.l.functions {
		signatures := f.l.checker.GetSignaturesOfType(f.l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[index] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	for _, closure := range f.l.closureRecords {
		signatures := f.l.checker.GetSignaturesOfType(closure.proven, checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[closure.function] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	var members func(ir.Expression, *checker.Type)
	members = func(expression ir.Expression, proven *checker.Type) {
		if expression == nil || proven == nil || f.weak(proven) {
			return
		}
		if graphType(proven) {
			pending = append(pending, expression)
		}
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			for _, field := range f.fields(proven) {
				for _, initializer := range value.Fields {
					if f.l.cycleFieldMatches(proven, field.Name, initializer.Name) {
						members(initializer.Value, f.l.checker.GetTypeOfSymbol(field))
					}
				}
			}
		case ir.ArrayLiteral:
			arguments := f.l.checker.GetTypeArguments(proven)
			if len(arguments) > 0 {
				for _, element := range value.Elements {
					members(element, arguments[0])
				}
			}
		case ir.Conditional:
			members(value.WhenTrue, proven)
			members(value.WhenNot, proven)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Assign:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Return:
				members(value.Value, returns[function])
				if function >= 0 {
					sources[resultNode(function)] = append(sources[resultNode(function)], value.Value)
				}
			case ir.Call:
				for _, target := range program.CallTargets(value) {
					if target < 0 || target >= len(program.Functions) {
						continue
					}
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// Opaque function values are a documented leak-only frontier.
					break
				}
				for _, target := range targets.Functions {
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.SetProperty:
				if slot(value.Site, value.Name, 0) {
					pending = append(pending, value.Value)
				}
			case ir.SetIndex:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.ArrayPush:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.MapSet:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Key)
				}
				if slot(value.Site, "", 1) {
					pending = append(pending, value.Value)
				}
			case ir.SetAdd:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	var follow func(ir.Expression)
	follow = func(expression ir.Expression) {
		if expression == nil {
			return
		}
		value := reflect.ValueOf(expression)
		if field := value.FieldByName("GraphTypes"); field.IsValid() {
			ids := field.Interface().([]int)
			program.GraphTypes[ids[len(ids)-1]] = true
			return
		}
		switch value := expression.(type) {
		case ir.Read:
			demand(value.Local + 1)
		case ir.Call:
			for _, target := range program.CallTargets(value) {
				demand(resultNode(target))
			}
		case ir.CallClosure:
			targets := program.ClosureTargets(value)
			if targets.Unknown {
				return
			}
			for _, target := range targets.Functions {
				demand(resultNode(target))
			}
		case ir.Conditional:
			follow(value.WhenTrue)
			follow(value.WhenNot)
		case ir.Box:
			follow(value.Value)
		case ir.Narrow:
			follow(value.Value)
		case ir.Unwrap:
			follow(value.Value)
		case ir.CheckedCast:
			follow(value.Value)
		}
	}
	for _, expression := range pending {
		follow(expression)
	}
	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, expression := range sources[node] {
			follow(expression)
		}
	}
}

func (f *oldCycleFinder) unproven(kind fresh.WriteKind, holder *checker.Type, name string) *fresh.Write {
	if f.writes == nil {
		f.writes = fresh.ProveWrites(f.l.result)
	}
	for index := range f.writes {
		write := &f.writes[index]
		if write.Proven {
			continue
		}
		if write.Kind == fresh.WriteUnknown {
			return write
		}
		if write.Kind != kind || (kind == fresh.WriteField && !f.l.cycleFieldMatches(holder, name, write.Name)) {
			continue
		}
		if write.Site == 0 {
			// A write lowering didn't record: it may be into anything of its kind.
			return write
		}
		written := f.l.writeSites[write.Site-1].holder
		if written == nil || f.template(written) || f.related(holder, written) {
			return write
		}
	}
	return nil
}

func (f *oldCycleFinder) writtenAt(write *fresh.Write) string {
	if write.Site == 0 {
		return "a write in " + f.l.functionName(write.Function)
	}
	return "the write at " + f.l.program.Where(f.l.writeSites[write.Site-1].node)
}

func (f *oldCycleFinder) libraryIteratorMade(node *ast.Node) {
	if node.Kind != ast.KindCallExpression {
		return
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	iterator := f.l.checker.GetTypeAtLocation(node)
	collection := f.l.checker.GetTypeAtLocation(access.Expression)
	if f.l.isLibraryType(iterator, "MapIterator", "SetIterator") &&
		f.l.isLibraryType(collection, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		f.libraryIterators = append(f.libraryIterators, libraryIteratorCapture{iterator, collection})
	}
}

func (f *oldCycleFinder) libraryIteratorCaptures(proven *checker.Type) []cycleNode {
	var held []cycleNode
	for _, capture := range f.libraryIterators {
		if f.l.checker.IsTypeAssignableTo(capture.iterator, proven) {
			held = append(held, cycleNode{proven: capture.collection})
		}
	}
	return held
}

// Compare every slot refusal and every reachability query, even when another
// slot would select the same ownership seed and hide that difference in the IR.
func (f *oldCycleFinder) slotsOf(holder *checker.Type) error {
	old := f.oldSlotsOf(holder)
	if f.compare != nil {
		next := f.compare.slotsOf(holder)
		if !reflect.DeepEqual(old, next) {
			f.differences = append(f.differences, fmt.Sprintf("slots of %s: old=%v new=%v", f.l.checker.TypeToString(holder), old, next))
		}
	}
	return old
}
func (f *oldCycleFinder) reaches(from *checker.Type, target cycleNode) bool {
	old := f.oldReaches(from, target)
	if f.compare != nil {
		next := f.compare.reaches(from, target)
		if old != next {
			f.differences = append(f.differences, fmt.Sprintf("reach from %d to %v: old=%t new=%t", from.Id(), target, old, next))
		}
	}
	return old
}

func (f *oldCycleFinder) weak(proven *checker.Type) bool {
	old := f.oldWeak(proven)
	if f.compare != nil && old != f.compare.weak(proven) {
		f.differences = append(f.differences, fmt.Sprintf("weak %d differs", proven.Id()))
	}
	return old
}

// A refusal ends slotsOf early. Probe the remaining slots too, so a change to a
// later slot cannot hide behind the same seed and the same first refusal.
func (f *oldCycleFinder) compareAllSlots() {
	check := func(holder, proven *checker.Type, kind fresh.WriteKind, name string) {
		weak := f.weak(proven)
		reachable := f.reaches(proven, cycleNode{proven: holder})
		if !weak && reachable {
			old := f.unproven(kind, holder, name)
			next := f.compare.cycleUnproven(kind, holder, name)
			if !reflect.DeepEqual(old, next) {
				f.differences = append(f.differences, fmt.Sprintf("write to %d.%s differs: old=%v new=%v", holder.Id(), name, old, next))
			}
		}
	}
	for _, holder := range f.seen {
		if holder.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(holder) {
			continue
		}
		switch {
		case f.l.checker.IsArrayType(holder) || checker.IsTupleType(holder):
			for _, element := range f.l.checker.GetTypeArguments(holder) {
				check(holder, element, fresh.WriteElement, "")
			}
		case f.l.isLibraryType(holder, "Map", "ReadonlyMap"):
			for _, element := range f.l.checker.GetTypeArguments(holder) {
				check(holder, element, fresh.WriteMapEntry, "")
			}
		case f.l.isLibraryType(holder, "Set", "ReadonlySet"):
			for _, element := range f.l.checker.GetTypeArguments(holder) {
				check(holder, element, fresh.WriteSetElement, "")
			}
		default:
			for _, field := range f.fields(holder) {
				check(holder, f.l.checker.GetTypeOfSymbol(field), fresh.WriteField, field.Name)
			}
		}
	}
	for local, declared := range f.l.result.Locals {
		if !declared.Captured || declared.Global {
			continue
		}
		if proven := f.l.localTypes[local]; proven != nil {
			f.weak(proven)
			f.reaches(proven, cycleNode{cell: local + 1})
		}
	}

}

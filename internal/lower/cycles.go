package lower

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// The cycle finder (docs/memory.md). Reference counting can't free a cycle, and Adamic has no
// collector, so a cycle must not be able to form. One forms only when a slot (a field, readonly or
// not, since a constructor writes a readonly one, an element of an array that isn't readonly, a map's
// value, a variable a function value captures) is set to something that can reach back to what
// holds the slot. Whether it can is visible in the types, so every such slot is found here, from the
// whole program, and each one is refused unless it's declared Weak<Target>, which doesn't count. A
// map's key and a set's element are slots too, now that they may be objects; neither can be declared
// weak, so one that can reach back is refused with the fix of a ReadonlyMap or ReadonlySet.
//
// Reaching is followed through everything the checker knows a value can hold: an object's fields,
// including the fields of every object type in the program that can be seen as it (a Dog seen as an
// Animal brings its owner along), an array's elements, a map's keys and values, and through a
// function type, the variables captured by every function value in the program that can be seen as
// it, since a type doesn't say what a function captured. A Weak isn't followed: it holds nothing.
//
// It runs after lowering, when the variables each function value captures are known.

// closureRecord is a function value the program makes: its type, and the function lowered for it.
type closureRecord struct {
	proven   *checker.Type
	function int
	node     *ast.Node
}

// cycleNode is one thing reaching is followed through: a type, or the cell of a captured local.
type cycleNode struct {
	proven *checker.Type
	cell   int
}

type cycleFinder struct {
	l *lowering

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

	// Outgoing edges are fixed after lowering and type collection. A node reached from many
	// slots must not rescan all program shapes and closures for every holder.
	edges                map[cycleNode][]cycleEdge
	shapeRepresentatives map[*checker.Type]*checker.Type
	shapeSignatures      map[string]*checker.Type
	shapeRelations       map[[2]*checker.Type]bool
}

// findCycles refuses the first cycle-capable slot that isn't declared Weak and has a write that isn't
// proven not to close a cycle (fresh.go), or returns nil.
func (l *lowering) findCycles(modules []*ast.SourceFile) error {
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
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
	for _, proven := range finder.seen {
		if err := finder.slotsOf(proven); err != nil {
			return err
		}
	}
	for local, declared := range l.result.Locals {
		if !declared.Captured || declared.Global {
			continue
		}
		proven, node := l.localTypes[local], l.localNodes[local]
		if proven == nil || node == nil || finder.weak(proven) {
			continue
		}
		if finder.reaches(proven, cycleNode{cell: local + 1}) {
			return &Refused{
				Where: l.program.Where(node),
				What:  "'" + declared.Name + "', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free",
				Fix:   "write the function as a function declaration (function " + declared.Name + "() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)",
			}
		}
	}
	return nil
}

// made notes the type of a value just made: an object type as a shape, and what anything else is
// made of as used.
func (f *cycleFinder) made(proven *checker.Type, where *ast.Node) {
	switch {
	case proven == nil:
	case proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
		for _, member := range proven.Types() {
			f.made(member, where)
		}
	case proven.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(proven):
		f.use(proven, where)
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.use(argument, where)
		}
	default:
		f.shape(proven, where)
	}
}

// shape notes an object type as one a value may really be, without examining its slots, and uses
// the types of its fields.
func (f *cycleFinder) shape(proven *checker.Type, where *ast.Node) {
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

// use notes a type the program uses, and every type it's made of.
func (f *cycleFinder) use(proven *checker.Type, where *ast.Node) {
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
	if f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
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

// declaredAt is where a field is declared, to point at what its type is made of, or otherwise.
func declaredAt(field *ast.Symbol, otherwise *ast.Node) *ast.Node {
	if len(field.Declarations) > 0 {
		return field.Declarations[0]
	}
	return otherwise
}

// template reports whether a type is a generic one not yet instantiated, or instantiated with any (a
// class seen from its own declaration): no value has it, and any relates to everything, so it would
// make every slot look cycle-capable.
func (f *cycleFinder) template(proven *checker.Type) bool {
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

// fields includes literal methods, which own closures. Class prototype methods hold no instance data.
func (f *cycleFinder) fields(proven *checker.Type) []*ast.Symbol {
	fields := []*ast.Symbol{}
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		if (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property) && !(f.l.isStaticType(proven) && property.Name == "prototype") {
			fields = append(fields, property)
		}
	}
	return fields
}

// isFunction reports whether a type is a function's, or a class's own (typeof Tag, which new calls).
// A class value holds its static fields, readonly in 0.1, and its prototype, which the checker shows
// as a field of the instance type and isn't one a program can set; seen as an object, that prototype
// read as a mutable field reaching back, and every class made with new was refused.
func (f *cycleFinder) isFunction(proven *checker.Type) bool {
	return len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || (len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && (!f.l.isStaticType(proven) || len(f.l.staticGlobals) == 0))
}

// weak reports whether a slot's type is a Weak<Target>, which holds nothing.
func (f *cycleFinder) weak(proven *checker.Type) bool {
	representation, _ := f.l.representation(proven)
	return representation == ir.Weak
}

// slotsOf refuses the first of a type's mutable slots that can close a cycle, unless every write
// into it is proven not to (fresh.go): then it stands, and a refusal names the write that isn't.
func (f *cycleFinder) slotsOf(holder *checker.Type) error {
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

// reaches reports whether a value of type from can reach target: a value seen as target's type
// (either way round, since either may be what the value really is), or target's cell.
func (f *cycleFinder) reaches(from *checker.Type, target cycleNode) bool {
	visited := map[cycleNode]bool{}
	queue := []cycleNode{{proven: from}}
	// Tracing records the same breadth-first search, including structural matches.
	// It does not change which edges are followed or which slots are refused.
	tracing := os.Getenv("ADAMIC_TRACE_CYCLES") == "1"
	type step struct {
		parent cycleNode
		edge   string
	}
	var parents map[cycleNode]step
	root := cycleNode{proven: from}
	if tracing {
		parents = map[cycleNode]step{root: {}}
	}
	var current cycleNode
	enqueue := func(next cycleNode, edge string) {
		if tracing {
			if _, found := parents[next]; !found {
				parents[next] = step{current, edge}
			}
		}
		queue = append(queue, next)
	}
	trace := func(last cycleNode, edge string) {
		if !tracing {
			return
		}
		path := []cycleNode{last}
		for path[len(path)-1] != root {
			path = append(path, parents[path[len(path)-1]].parent)
		}
		fmt.Fprintf(os.Stderr, "cycle reach: %s -- slot contents --> %s\n", f.cycleNodeName(target), f.cycleNodeName(root))
		for index := len(path) - 1; index > 0; index-- {
			next := path[index-1]
			fmt.Fprintf(os.Stderr, "  %s -- %s --> %s\n", f.cycleNodeName(path[index]), parents[next].edge, f.cycleNodeName(next))
		}
		fmt.Fprintf(os.Stderr, "  %s -- %s --> %s\n", f.cycleNodeName(last), edge, f.cycleNodeName(target))
	}
	for len(queue) > 0 {
		node := queue[0]
		current = node
		queue = queue[1:]
		identity := node
		if !tracing && node.proven != nil {
			// Identical plain literals have identical member types and shape relations.
			// Their only different outgoing edge is the omitted self-shape edge,
			// which cannot add reachability. Visit that equivalence class once.
			// Tracing keeps original nodes to preserve the exact breadth-first path.
			identity.proven = f.shapeRepresentative(node.proven)
		}
		if visited[identity] {
			continue
		}
		visited[identity] = true
		if node.cell != 0 {
			if node == target {
				trace(node, "same captured cell")
				return true
			}
			for _, edge := range f.outgoing(node, tracing) {
				enqueue(edge.next, edge.name)
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
				enqueue(cycleNode{proven: member}, "union or intersection member")
			}
			continue
		}
		if flags&checker.TypeFlagsObject == 0 {
			continue
		}
		if target.proven != nil && f.related(proven, target.proven) {
			trace(node, "related holder types")
			return true
		}
		for _, edge := range f.outgoing(node, tracing) {
			enqueue(edge.next, edge.name)
		}
	}
	return false
}

// cycleEdge retains the traversal order and the trace label of an outgoing edge.
type cycleEdge struct {
	next cycleNode
	name string
}

// outgoing memoizes only target-independent edges, never a reach decision. In particular,
// related holder types are still checked at the same point in each breadth-first search.
func (f *cycleFinder) outgoing(node cycleNode, tracing bool) []cycleEdge {
	if edges, found := f.edges[node]; found {
		return edges
	}
	var edges []cycleEdge
	appendEdge := func(next cycleNode, name string) {
		edges = append(edges, cycleEdge{next, name})
	}
	if node.cell != 0 {
		if proven := f.l.localTypes[node.cell-1]; proven != nil {
			appendEdge(cycleNode{proven: proven}, "captured cell contents")
		}
		// A this shared by every instantiation has all the original cell-content edges.
		for _, proven := range f.l.localAlso[node.cell-1] {
			appendEdge(cycleNode{proven: proven}, "captured cell contents")
		}
	} else {
		proven := node.proven
		switch {
		case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
			for _, capture := range f.libraryIteratorCaptures(proven) {
				appendEdge(capture, "library iterator capture")
			}
		case f.isFunction(proven):
			// Construct signatures can hide constructor objects behind an interface.
			if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
				for symbol := range f.l.statics {
					actual := f.l.checker.GetTypeOfSymbol(symbol)
					if f.l.checker.IsTypeAssignableTo(actual, proven) {
						appendEdge(cycleNode{proven: actual}, "assignable constructor")
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
					edge := "closure capture"
					if tracing {
						edge += " at " + fmt.Sprint(f.l.program.Where(closure.node))
					}
					appendEdge(cycleNode{cell: local + 1}, edge)
				}
			}
		case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
			for _, argument := range f.l.checker.GetTypeArguments(proven) {
				appendEdge(cycleNode{proven: argument}, "collection type argument")
			}
		default:
			if f.l.isStaticType(proven) {
				if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
					appendEdge(cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))}, "static base")
				}
			}
			for _, accessor := range f.l.accessorCaptures {
				if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
					for _, local := range f.l.result.Functions[accessor.function].Environment {
						appendEdge(cycleNode{cell: local + 1}, "accessor capture")
					}
				}
			}
			for _, field := range f.fields(proven) {
				appendEdge(cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)}, "field "+field.Name)
			}
			// A value seen as this type may be any object type the program has that can be seen as
			// it, with fields this type doesn't show.
			for _, shape := range f.shapes {
				if shape != proven && f.shapeAssignable(shape, proven) {
					appendEdge(cycleNode{proven: shape}, "assignable program shape")
				}
			}
		}
	}
	if f.edges == nil {
		f.edges = map[cycleNode][]cycleEdge{}
	}
	f.edges[node] = edges
	return edges
}

// shapeAssignable reuses checker decisions for identical plain object literals. It
// still emits each original shape edge: representatives never replace graph nodes.
func (f *cycleFinder) shapeAssignable(from, to *checker.Type) bool {
	key := [2]*checker.Type{f.shapeRepresentative(from), f.shapeRepresentative(to)}
	if result, found := f.shapeRelations[key]; found {
		return result
	}
	result := f.l.checker.IsTypeAssignableTo(from, to)
	if f.shapeRelations == nil {
		f.shapeRelations = map[[2]*checker.Type]bool{}
	}
	f.shapeRelations[key] = result
	return result
}

// shapeRepresentative interns only plain literal records whose members are declared
// directly in that literal. Equal keys retain freshness, member order, flags, readonly
// status and exact member types. The checker additionally confirms type identity.
// Classes, references, signatures, indexes, methods, accessors and spread members
// keep their original identities, including any nominal or declaration-dependent rules.
func (f *cycleFinder) shapeRepresentative(proven *checker.Type) *checker.Type {
	if representative, found := f.shapeRepresentatives[proven]; found {
		return representative
	}
	representative := proven
	literal := proven.Symbol()
	if literal != nil && len(literal.Declarations) == 1 && literal.Declarations[0].Kind == ast.KindObjectLiteralExpression &&
		proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsObjectLiteral != 0 &&
		proven.ObjectFlags()&checker.ObjectFlagsReference == 0 &&
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0 &&
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) == 0 &&
		len(f.l.checker.GetIndexInfosOfType(proven)) == 0 {
		var signature strings.Builder
		fmt.Fprintf(&signature, "%d/%d;", proven.Flags(), proven.ObjectFlags())
		plain := true
		for _, property := range f.l.checker.GetPropertiesOfType(proven) {
			if len(property.Declarations) != 1 {
				plain = false
				break
			}
			declaration := property.Declarations[0]
			if (declaration.Kind != ast.KindPropertyAssignment && declaration.Kind != ast.KindShorthandPropertyAssignment) ||
				declaration.Parent != literal.Declarations[0] {
				plain = false
				break
			}
			fmt.Fprintf(&signature, "%q/%d/%d/%t/%p;", property.Name, property.Flags, property.CheckFlags,
				f.l.checker.IsReadonlySymbol(property), f.l.checker.GetTypeOfSymbol(property))
		}
		if plain {
			key := signature.String()
			if earlier := f.shapeSignatures[key]; earlier != nil {
				if checker.Checker_isTypeIdenticalTo(f.l.checker, proven, earlier) {
					representative = earlier
				}
			} else {
				if f.shapeSignatures == nil {
					f.shapeSignatures = map[string]*checker.Type{}
				}
				f.shapeSignatures[key] = proven
			}
		}
	}
	if f.shapeRepresentatives == nil {
		f.shapeRepresentatives = map[*checker.Type]*checker.Type{}
	}
	f.shapeRepresentatives[proven] = representative
	return representative
}

// cycleNodeName identifies both type holders and captured variable holders in a debug path.
func (f *cycleFinder) cycleNodeName(node cycleNode) string {
	if node.cell != 0 {
		return "cell " + f.l.result.Locals[node.cell-1].Name
	}
	return f.l.checker.TypeToString(node.proven)
}

// present is a type as written without undefined, as Weak<Target> wants its Target.
func (f *cycleFinder) present(proven *checker.Type) string {
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

// related reports whether one value could be seen as both types: one is assignable to the other.
func (f *cycleFinder) related(one *checker.Type, other *checker.Type) bool {
	if one == other {
		return true
	}
	if f.isFunction(one) || f.isFunction(other) {
		return false
	}
	return f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)
}

func literalMethod(property *ast.Symbol) bool {
	for _, declaration := range property.Declarations {
		if declaration.Kind == ast.KindMethodDeclaration && declaration.Parent.Kind == ast.KindObjectLiteralExpression {
			return true
		}
	}
	return false
}

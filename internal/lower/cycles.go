package lower

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
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

// cycleNode is one thing reaching is followed through: a type, the cell of a captured local, or an
// Apple class by its Objective-C name, which the program may never name (what a subclass's
// stickTo: keeps, an NSRecord, whose binding no file of the program loads).
type cycleNode struct {
	proven *checker.Type
	cell   int
	objc   string
}

type cycleFinder struct {
	l *lowering

	// seen is every type the program uses, in the order first met, with where it was met. A type can be
	// met with no node (an instantiation lowering made whose type has no declaration, and what use derives
	// from it), so where is read only through whereOf, which never hands a refusal a nil node (#ew207ss).
	// whereFallback is the first module, where whereOf points when nothing nearer is known.
	seen          []*checker.Type
	where         map[*checker.Type]*ast.Node
	whereFallback *ast.Node

	// shapes is every seen object type that isn't an array, a map or a function: what a value seen as
	// some object type may really be. shaped are the literals' own types among them.
	shapes []*checker.Type
	shaped map[*checker.Type]bool

	// scalarShapes shares identical plain literal shapes. These have no hidden
	// members, signatures, or reference fields; their complete structural types
	// are the same, so repeating them adds no reachability edge.
	scalarShapes map[string]bool

	// writes is every write in the program, judged by the relaxation for fresh writes (fresh.go),
	// worked out the first time a slot needs it.
	writes []fresh.Write

	libraryIterators []libraryIteratorCapture
	// viewAlso is, for each type a checked view lets the program see a value through, the source types the
	// view can hand out there that the finder doesn't relate to it (viewAlsoTable), built once from every view.
	viewAlso map[*checker.Type][]*checker.Type
	// A type's slots and strong links do not change during a walk. Keep their checker
	// queries and compatible views once, rather than deriving them for every slot.
	// These caches are sound only while what reachLinks reads stays fixed once the first query
	// runs: shapes (whose length resets reachCache, reachEdges and shapeIndex), and also statics,
	// closureRecords (whose length resets closures), each function value's Environment, accessorCaptures,
	// localTypes, localAlso and, for functionCache, staticGlobals. All are complete before findCycles
	// today, and each lowering makes a new finder. A pass that adds to any of them mid-walk must reset
	// these caches too.
	fieldCache     map[*checker.Type][]*ast.Symbol
	functionCache  map[*checker.Type]bool
	templateCache  map[*checker.Type]bool
	weakCache      map[*checker.Type]bool
	relationCache  map[[2]*checker.Type]bool
	reachCache     map[cycleReach]bool
	reachEdges     map[cycleNode][]cycleNode
	reachShapes    int
	shapeIndex     map[string][]*checker.Type
	shapeIndexSize int
	literalTypes   map[*checker.Type]*checker.Type
	literalSlots   cycleSlotTrie
	shapeChoices   []*checker.Type
	writeIndex     map[cycleWriteKey][]int
	writeCache     map[cycleWriteSlot]*fresh.Write
	unknownWrite   int
	// closures is closureRecords grouped by type for the closure arm (closureLinksTo), and closureQueries
	// how many relation queries that arm has asked, which TestClosureArmRelatesEachCapturingTypeOnce reads.
	closures       *closureGroups
	closureQueries int

	// recordSlots are the seeds slotsOf found whose written value may be a record (noteRecordSlot), and
	// recordCells the captured cells graphTypes seeds, each judged once the graph types are selected
	// (judgeRecordSlots).
	recordSlots []recordSlotNote
	recordCells []recordSlotNote
	// recordStop is the first of them that can hold a record of a type graphTypes didn't select, which
	// no graph region holds, returned by graphTypes after the boundary's refusals.
	recordStop error
	// recordWrites is whether any write site's holder is a record (recordWritten): 0 unread, 1 yes, -1 no.
	recordWrites int
	// assigned is every local an ir.Assign writes after its declaration (reassigned), found once.
	assigned map[int]bool
	// selected is every type graphTypes made a graph type, which unplannedPair reads (#dps0kp6).
	selected []*checker.Type
	// loose is the loose set H and what the loose-slot stop keeps (loose_slots.go).
	loose looseSlots

	// nesting bounds use's walk through a generic that instantiates itself deeper at each level
	// (typeNesting, tsc's isDeeplyNestedType): Own<T> { next: Own<Own<T>> } makes a fresh type at every
	// field, so where never repeats and the walk ran without end (#ykd89v9). deepest is the first type
	// past the bound and deepestAt where it was met; findCycles stops there, since a program whose types
	// the finder can't see to the end of isn't proven free of cycles.
	nesting   typeNesting
	deepest   *checker.Type
	deepestAt *ast.Node
}

// findCycles refuses the first cycle-capable slot that isn't declared Weak and has a write that isn't
// proven not to close a cycle (fresh.go), or returns nil.
func (l *lowering) findCycles(modules []*ast.SourceFile) error {
	relationSizeTrace(l.checker, "findCycles start")
	defer relationSizeTrace(l.checker, "findCycles end")
	defer armTrace("findCycles end")
	// Generated frame/Promise/reaction layouts are IR identities, never checker declarations.
	// Only their runtime-owned protocol edges are exempt. Source types, including a class
	// called adamic_async_frame, continue through slotsOf with no name-based escape hatch.
	if l.result.HasAsync() {
		for _, generated := range l.result.Generated {
			if !fresh.RuntimeBreaksCycles(generated) {
				return fmt.Errorf("lower: unaudited generated cycle identity")
			}
		}
	}
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	if len(modules) > 0 {
		finder.whereFallback = modules[0].AsNode()
	}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if l.firstLinkSkips(node) {
				return false
			}
			finder.libraryIteratorMade(node)
			finder.looseEntry(node)
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
	relationSizeTrace(l.checker, "findCycles after the AST walk")
	// Every instantiation lowering made, Box<Node> inside Maker<Node> included, which the source may
	// never write.
	for _, proven := range l.instantiated {
		finder.use(proven, l.classNodeFor(proven))
	}
	relationSizeTrace(l.checker, "findCycles after instantiated")
	if finder.deepest != nil {
		where := finder.deepestAt
		if where == nil && len(modules) > 0 {
			where = modules[0].AsNode()
		}
		return l.notYet(where, "finding cycles through a type that instantiates itself deeper at each level ("+l.checker.TypeToString(finder.deepest)+")")
	}
	// Program membership changes ownership only after every ordinary graph check passes.
	relationSizeTrace(l.checker, "graphTypes start")
	if err := finder.graphTypes(modules); err != nil {
		relationSizeTrace(l.checker, "graphTypes stopped")
		return err
	}
	if err := finder.jsonReferenceReach(); err != nil {
		return err
	}
	// A delegate is held by the Apple object it's handed to, a slot no declaration shows and no write
	// can be proven fresh into, since Apple keeps it past the call: one that can reach back to its
	// holder is a cycle.
	for _, hold := range l.delegateHolds {
		if l.firstLinkSkips(hold.node) {
			continue
		}
		if finder.reaches(hold.delegate, cycleNode{proven: hold.holder}) {
			holder := l.checker.TypeToString(hold.holder)
			return &Refused{
				Where: l.program.Where(hold.node),
				What:  l.checker.TypeToString(hold.delegate) + ", handed to Apple as a delegate the " + holder + " holds from then on, which can reach back to that " + holder + ": a cycle reference counting can't free",
				Fix:   "hold the " + holder + " weakly where the delegate reaches it (Weak<" + holder + ">, import type { Weak } from 'adamic'), or keep the delegate apart from what it's the delegate of (adamic/cycle-capable)",
			}
		}
	}
	for local, declared := range l.result.Locals {
		if !declared.Captured || declared.Global {
			continue
		}
		proven, node := l.localTypes[local], l.localNodes[local]
		if l.firstLinkSkips(node) {
			continue
		}
		if proven == nil || node == nil || finder.weak(proven) || !finder.appleObject(proven) {
			continue
		}
		if finder.reaches(proven, cycleNode{cell: local + 1}) && !l.closedFrameInput(local) {
			kind := ""
			if declared.Function >= 0 && l.result.Functions[declared.Function].Async {
				kind = "async frame capture cycle: "
			}
			fix := "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)"
			if apple := finder.present(proven); finder.appleObject(proven) {
				// An Apple object a closure handed to Apple captures (a label an action sets): Apple may
				// hold the closure through what the object holds.
				fix = "capture it weakly, const " + declared.Name + "Held: Weak<" + apple + "> = " + declared.Name + " (import type { Weak } from 'adamic'), and keep it somewhere strong, as its superview does a view; or keep it in a module-level constant, which nothing frees (adamic/cycle-capable)"
			}
			return &Refused{
				Where: l.program.Where(node),
				What:  kind + "'" + declared.Name + "', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free",
				Fix:   fix,
			}
		}
	}
	if l.programDiscovery {
		l.resolveCountTypes(modules)
		l.programPlan = l.selectProgramMembership(finder, modules)
		return l.checkLazyViewReads()
	}
	l.resolveCountTypes(modules)
	if err := l.nodeProcessStorageABI(modules); err != nil {
		return err
	}
	if err := l.nodeFSDirectoryClosureABI(); err != nil {
		return err
	}
	return l.checkLazyViewReads()
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
	case f.l.recordElement(proven) != nil:
		f.use(f.l.recordElement(proven), where)
	case f.l.isArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
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
	key, scalar := f.scalarShapeKey(proven, where)
	if scalar {
		if f.scalarShapes[key] {
			return
		}
		if f.scalarShapes == nil {
			f.scalarShapes = map[string]bool{}
		}
		f.scalarShapes[key] = true
	}
	f.shapes = append(f.shapes, proven)
	for _, field := range f.fields(proven) {
		f.use(f.l.checker.GetTypeOfSymbol(field), declaredAt(field, where))
	}
	// A literal cast to a narrower type ({ name, back } as Named) has no variable to carry its regular type,
	// and its fresh type never relates to the narrower one, since the relation judges a fresh type for excess
	// properties. So its regular widened type is a shape too, as a variable's would be, and a value seen as
	// Named may be it, back and all (#sbvnh96). made and use both reach a literal through here. A scalar
	// shape holds no reference, so it can close no cycle and keeps its one shape (TestCycleScalarLiteralShapesOnce).
	if !scalar && proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
		f.use(f.l.relatedShape(proven), where)
	}
}

// scalarShapeKey is deliberately limited to plain, directly written scalar
// properties. Reference-bearing, spread, method, accessor, and computed shapes
// retain their individual identities and all their cycle edges.
func (f *cycleFinder) scalarShapeKey(proven *checker.Type, where *ast.Node) (string, bool) {
	if where == nil || where.Kind != ast.KindObjectLiteralExpression ||
		proven.Flags() != checker.TypeFlagsObject ||
		proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral == 0 ||
		len(f.l.checker.GetIndexInfosOfType(proven)) != 0 ||
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 ||
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return "", false
	}
	properties := where.AsObjectLiteralExpression().Properties.Nodes
	fields := f.l.checker.GetPropertiesOfType(proven)
	if len(properties) != len(fields) {
		return "", false
	}
	var key strings.Builder
	for i, property := range properties {
		if property.Kind != ast.KindPropertyAssignment || property.Name() == nil ||
			property.Name().Kind != ast.KindIdentifier {
			return "", false
		}
		field := fields[i]
		if field.Name != property.Name().Text() {
			return "", false
		}
		value := f.l.checker.GetTypeOfSymbol(field)
		// Exact checker identities retain distinct scalar literal values.
		flags := value.Flags()
		boolean := flags&checker.TypeFlagsBoolean != 0 && flags & ^(checker.TypeFlagsBoolean|checker.TypeFlagsUnion) == 0
		if !boolean && flags & ^(checker.TypeFlagsNumber|checker.TypeFlagsString|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsStringLiteral|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			return "", false
		}
		key.WriteString(strconv.Itoa(len(field.Name)))
		key.WriteByte(':')
		key.WriteString(field.Name)
		fmt.Fprintf(&key, ":%p:%d:%t;", value, field.Flags, f.l.checker.IsReadonlySymbol(field))
	}
	return key.String(), true
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
	ok, leave := f.nesting.enter(proven)
	defer leave()
	if !ok {
		if f.deepest == nil {
			f.deepest, f.deepestAt = proven, where
		}
		return
	}
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
	if element := f.l.recordElement(proven); element != nil {
		f.use(element, where)
		return
	}
	if f.l.isArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise") {
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

// whereOf is where a refusal about these types is reported, and never nil: where the first of them that
// was met with a node was met; else where the first that has a declaration in the program is declared, a
// reference by its target's (never the library's, the prelude's or Node's); else the first module. findCycles' instantiations are used with classNodeFor's node, nil for a
// type with no declaration, and what use derives from such a type inherits the nil, so on tsc a holder met
// only that way reached Program.Where with no node and the compiler panicked (#ew207ss). The refusal's What
// names the type, so this is only where it's reported. A finder built outside findCycles has no first module
// stored, and falls back to the program's first file.
func (f *cycleFinder) whereOf(types ...*checker.Type) *ast.Node {
	for _, proven := range types {
		if where := f.where[proven]; where != nil {
			return where
		}
	}
	for _, proven := range types {
		if proven == nil {
			continue
		}
		symbol := proven.Symbol()
		if symbol == nil && proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
			symbol = proven.Target().Symbol()
		}
		if symbol == nil {
			continue
		}
		// A declaration in the library, the prelude or Node's types is no place the program can change: an
		// array's is lib's Array, so the refusal points at the program's first file instead.
		for _, declaration := range symbol.Declarations {
			if file := ast.GetSourceFileOfNode(declaration); file != nil && !load.IsLibrary(file) && !load.IsPrelude(file) && !load.IsNodeLibrary(file) {
				return declaration
			}
		}
	}
	if f.whereFallback != nil {
		return f.whereFallback
	}
	return f.l.program.Files()[0].AsNode()
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
	if value, ok := f.templateCache[proven]; ok {
		return value
	}
	value := f.deriveTemplate(proven)
	if f.templateCache == nil {
		f.templateCache = map[*checker.Type]bool{}
	}
	f.templateCache[proven] = value
	return value
}

// fields includes literal methods, which own closures. Class prototype methods hold no instance data.
func (f *cycleFinder) fields(proven *checker.Type) []*ast.Symbol {
	if value, ok := f.fieldCache[proven]; ok {
		return value
	}
	value := f.deriveFields(proven)
	if f.fieldCache == nil {
		f.fieldCache = map[*checker.Type][]*ast.Symbol{}
	}
	f.fieldCache[proven] = value
	return value
}

// isFunction reports whether a type is a function's, or a class's own (typeof Tag, which new calls).
// A class value holds its static fields, readonly in 0.1, and its prototype, which the checker shows
// as a field of the instance type and isn't one a program can set; seen as an object, that prototype
// read as a mutable field reaching back, and every class made with new was refused.
func (f *cycleFinder) isFunction(proven *checker.Type) bool {
	if value, ok := f.functionCache[proven]; ok {
		return value
	}
	value := f.deriveIsFunction(proven)
	if f.functionCache == nil {
		f.functionCache = map[*checker.Type]bool{}
	}
	f.functionCache[proven] = value
	return value
}

// weak reports whether a slot's type is a Weak<Target>, which holds nothing.
func (f *cycleFinder) weak(proven *checker.Type) bool {
	if value, ok := f.weakCache[proven]; ok {
		return value
	}
	value := f.deriveWeak(proven)
	if f.weakCache == nil {
		f.weakCache = map[*checker.Type]bool{}
	}
	f.weakCache[proven] = value
	return value
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
	case l.isArrayType(holder) || checker.IsTupleType(holder):
		if l.isLibraryType(holder, "ReadonlyArray") || (checker.IsTupleType(holder) && holder.TargetTupleType().IsReadonly()) {
			return nil
		}
		for _, element := range l.checker.GetTypeArguments(holder) {
			if f.weak(element) || !f.reaches(element, cycleNode{proven: holder}) {
				continue
			}
			write := f.cycleUnproven(fresh.WriteElement, holder, "")
			if write == nil {
				continue
			}
			f.noteRecordSlot(element, holder, write)
			target := f.present(element)
			return &Refused{
				cycle: true,
				Where: l.program.Where(f.whereOf(holder)),
				What:  name + ", an array whose elements can reach back to an array like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
				Fix:   "declare the elements weak, Weak<" + target + ">[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly " + target + "[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)",
			}
		}
	case l.recordElement(holder) != nil:
		element := l.recordElement(holder)
		if !f.weak(element) && f.reaches(element, cycleNode{proven: holder}) {
			if write := f.cycleUnproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{cycle: true, Where: l.program.Where(f.whereOf(holder)), What: name + ", a record whose values can reach back to its holder: a cycle reference counting cannot free, and " + f.writtenAt(write) + " may close one", Fix: "use weak links or write only values proven unable to reach the record (adamic/cycle-capable)"}
			}
		}
	case l.isLibraryType(holder, "ReadonlyMap"), l.isLibraryType(holder, "ReadonlySet"):
	case l.isLibraryType(holder, "Set"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 1 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.cycleUnproven(fresh.WriteSetElement, holder, ""); write != nil {
				f.noteRecordSlot(arguments[0], holder, write)
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.whereOf(holder)),
					What:  name + ", a set whose elements can reach back to a set like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlySet, built whole when it's made; or add to such a set only values this function made, or only to one it made (adamic/cycle-capable)",
				}
			}
		}
	case l.isLibraryType(holder, "Map"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 2 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.cycleUnproven(fresh.WriteMapEntry, holder, ""); write != nil {
				f.noteRecordSlot(arguments[0], holder, write)
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.whereOf(holder)),
					What:  name + ", a map whose keys can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlyMap, built whole when it's made; or set in such a map only keys this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
		if len(arguments) == 2 && !f.weak(arguments[1]) && f.reaches(arguments[1], cycleNode{proven: holder}) {
			if write := f.cycleUnproven(fresh.WriteMapEntry, holder, ""); write != nil {
				f.noteRecordSlot(arguments[1], holder, write)
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.whereOf(holder)),
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
			write := f.cycleUnproven(fresh.WriteField, holder, field.Name)
			if write == nil {
				continue
			}
			f.noteRecordSlot(proven, holder, write)
			where := f.whereOf(holder)
			if len(field.Declarations) > 0 {
				where = field.Declarations[0]
			}
			target := l.checker.TypeToString(proven)
			kind, orReadonly := "a mutable field", "; or make it readonly"
			if readonly {
				kind, orReadonly = "a readonly field, which its constructor writes,", ""
			}
			return &Refused{
				cycle: true,
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
	// graphTypes adds allocation views after selecting seeds. Rebuild links for that
	// second graph, since a newly discovered shape may add a strong path.
	if f.reachEdges == nil || f.reachShapes != len(f.shapes) {
		f.reachEdges = map[cycleNode][]cycleNode{}
		f.reachCache = map[cycleReach]bool{}
		f.reachShapes = len(f.shapes)
	}
	key := cycleReach{from, target}
	if value, ok := f.reachCache[key]; ok {
		return value
	}
	visited := map[cycleNode]bool{}
	queue := []cycleNode{{proven: from}}
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		if node.proven != nil {
			node.proven = f.literalType(node.proven)
		}
		queue = queue[:len(queue)-1]
		if visited[node] {
			continue
		}
		visited[node] = true
		if (node.cell != 0 && node == target) || (node.proven != nil && !f.weak(node.proven) && !f.template(node.proven) && node.proven.Flags()&checker.TypeFlagsObject != 0 && target.proven != nil && f.related(node.proven, target.proven)) {
			f.reachCache[key] = true
			return true
		}
		links, ok := f.reachEdges[node]
		if !ok {
			links = f.reachLinks(node)
			f.reachEdges[node] = links
		}
		queue = append(queue, links...)
	}
	f.reachCache[key] = false
	return false
}

// appleHolds is what an Apple class's instances may hold, by the leaf table's derivation
// (load.AppleHolds), as nodes to follow: a class the program loads as its type, one it doesn't by its
// Objective-C name; any object as NSObject, whose holds are every class's; a protocol as its
// interface, which the program's delegates can be seen as; a block as every closure the program
// hands Apple. A class the table doesn't know holds anything.
func (f *cycleFinder) appleHolds(class string) []cycleNode {
	holds, err := load.AppleHolds()
	entries, known := holds[class]
	if err != nil || !known {
		entries = []string{"object", "block"}
	}
	nodes := []cycleNode{}
	for _, entry := range entries {
		kind, name, _ := strings.Cut(entry, " ")
		switch kind {
		case "class":
			if f.l.appleLeaf(name) {
				continue
			}
			if proven := f.l.appleType("class " + name); proven != nil {
				nodes = append(nodes, cycleNode{proven: proven})
			} else {
				nodes = append(nodes, cycleNode{objc: name})
			}
		case "protocol":
			if proven := f.l.appleType("protocol " + name); proven != nil {
				nodes = append(nodes, cycleNode{proven: proven})
			}
		case "object":
			if proven := f.l.appleType("class NSObject"); proven != nil {
				nodes = append(nodes, cycleNode{proven: proven})
			}
			if class != "NSObject" {
				nodes = append(nodes, cycleNode{objc: "NSObject"})
			}
		case "block":
			for _, handed := range f.l.appleHanded {
				nodes = append(nodes, cycleNode{proven: handed})
			}
		}
	}
	return nodes
}

// appleObject reports whether a type, undefined aside, is one of Apple's classes.
func (f *cycleFinder) appleObject(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if member.Flags()&checker.TypeFlagsUndefined == 0 && f.l.foreignClassName(member) == "" {
				return false
			}
		}
		return true
	}
	return f.l.foreignClassName(proven) != ""
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
	if one.Id() > other.Id() {
		one, other = other, one
	}
	key := [2]*checker.Type{one, other}
	if value, ok := f.relationCache[key]; ok {
		return value
	}
	value := f.deriveRelated(one, other)
	if f.relationCache == nil {
		f.relationCache = map[[2]*checker.Type]bool{}
	}
	f.relationCache[key] = value
	return value
}

func literalMethod(property *ast.Symbol) bool {
	for _, declaration := range property.Declarations {
		if declaration.Kind == ast.KindMethodDeclaration && declaration.Parent.Kind == ast.KindObjectLiteralExpression {
			return true
		}
	}
	return false
}

func (f *cycleFinder) deriveTemplate(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return false
	}
	for index, argument := range f.l.checker.GetTypeArguments(proven) {
		if argument.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			// An any where the type parameter's own default is any is the lib's, never written: Iterator<string>
			// is Iterator<string, any, any> by its TReturn and TNext. Treating it as a template made every slot
			// typed Iterator, Iterable or IterableIterator hold nothing for cycle finding, selection and the
			// region, so a Map's iterator stored in one leaked its Map (#66acn2a, Lower's find).
			if argument.Flags()&checker.TypeFlagsAny != 0 && anyByDefault(proven, index) {
				continue
			}
			return true
		}
	}
	return false
}

// anyByDefault is whether a reference's type parameter at index declares any as its default, as lib's TReturn
// and TNext do. It reads the declaration, since the checker gives the default no name the shim exports. A
// reference can't tell an any written in that place from the default's: Iterator<string, any> is the type
// Iterator<string> is, so both hold what their other arguments hold. That only finds more, never fewer, since an
// any argument is still a template where a walk meets it. The frozen scanning reference asks this too.
func anyByDefault(reference *checker.Type, index int) bool {
	target := reference.Target()
	if target == nil || target.AsInterfaceType() == nil {
		return false
	}
	parameters := target.AsInterfaceType().TypeParameters()
	if index >= len(parameters) || parameters[index].Symbol() == nil {
		return false
	}
	for _, declaration := range parameters[index].Symbol().Declarations {
		if declaration.Kind == ast.KindTypeParameter {
			if fallback := declaration.AsTypeParameterDeclaration().DefaultType; fallback != nil && fallback.Kind == ast.KindAnyKeyword {
				return true
			}
		}
	}
	return false
}

func (f *cycleFinder) deriveFields(proven *checker.Type) []*ast.Symbol {
	fields := []*ast.Symbol{}
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		if (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property) && !(f.l.isStaticType(proven) && property.Name == "prototype") {
			fields = append(fields, property)
		}
	}
	return fields
}

func (f *cycleFinder) deriveIsFunction(proven *checker.Type) bool {
	// The library Function interface has no call or construct signatures. Keep
	// the scanner's explicit classification so its caller field is not an owned edge.
	return f.l.isLibraryType(proven, "Function") || len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || (len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && (!f.l.isStaticType(proven) || len(f.l.staticGlobals) == 0))
}

func (f *cycleFinder) deriveWeak(proven *checker.Type) bool {
	representation, _ := f.l.representation(proven)
	return representation == ir.Weak
}

func (f *cycleFinder) reachLinks(node cycleNode) []cycleNode {
	if node.objc != "" {
		return f.appleHolds(node.objc)
	}
	links := []cycleNode{}
	if node.cell != 0 {
		// An interior async cell retains its owner, not just its declared value.
		// Completion drops private slots; captured slots remain reachable through
		// any escaped cell. Follow every such slot irrespective of signature.
		local := f.l.result.Locals[node.cell-1]
		if local.EnvironmentCell && local.Function >= 0 {
			owner := f.l.result.Functions[local.Function]
			if owner.Suspends() {
				for _, held := range owner.FrameEnvironment {
					if f.l.result.Locals[held].Captured {
						links = append(links, cycleNode{cell: held + 1})
					}
				}
			}
		}
		if proven := f.l.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		// A this shared by every instantiation of a class held the same way is each of them.
		for _, proven := range f.l.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}

		return links
	}
	proven := node.proven
	flags := proven.Flags()
	if f.weak(proven) || f.template(proven) {
		// A Weak holds nothing.
		return nil
	}
	if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			links = append(links, cycleNode{proven: member})
		}
		return links
	}
	if flags&checker.TypeFlagsObject == 0 {
		return nil
	}
	// A value seen as this type through a checked view may be a source type the view's walk met beneath it
	// that the finder doesn't relate to it (viewAlsoTable, #9489gt0).
	for _, held := range f.viewAlsoOf(proven) {
		links = append(links, cycleNode{proven: held})
	}
	switch {
	case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
		// A type can't say whether it's a collection's iterator or a generator declared as one. No
		// slot this walks holds a generator's call while generators aren't values (#qk8rztp S1,
		// TestNoSlotTheIteratorArmWalksHoldsAGenerator); once one can, this arm must add the frame's
		// links too (Lower, q6).
		links = append(links, f.libraryIteratorCaptures(proven)...)
	case f.isFunction(proven):
		// Construct signatures can hide constructor objects behind an interface.
		if f.l.isLibraryType(proven, "Function") || len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range f.l.statics {
				actual := f.l.checker.GetTypeOfSymbol(symbol)
				if f.l.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
				}
			}
		}
		// What a function value holds is what it captured: the cells of every function value the
		// program makes that can be seen as this type.
		links = append(links, f.closureLinksTo(proven)...)
	case f.l.recordElement(proven) != nil:
		links = append(links, cycleNode{proven: f.l.recordElement(proven)})
	case f.l.isArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			links = append(links, cycleNode{proven: argument})
		}
	default:
		if f.l.isStaticType(proven) {
			if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
				links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
			}
		}
		for _, accessor := range f.l.accessorCaptures {
			if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range f.l.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
		if class := f.l.foreignClassName(proven); !f.l.appleLeaf(class) {
			for _, field := range f.fields(proven) {
				links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
			}
			if class != "" {
				links = append(links, cycleNode{objc: class})
			}
		}
		// A value seen as this type may be any object type the program has that can be seen as
		// it, with fields this type doesn't show.
		for _, shape := range f.shapeCandidates(proven) {
			if shape != proven && f.l.checker.IsTypeAssignableTo(shape, proven) {
				links = append(links, cycleNode{proven: shape})
			}
		}
	}
	return links
}

// closureGroups is closureRecords grouped by type, for the closure arm of reachLinks and graphLinks
// (closureLinksTo). That arm related every function value the program makes to every function type a walk
// met: on tsc, 6,378 values times 8,318 types, 18 million queries, which spent the checker's relation
// budget inside 30 seconds (#c0eypws). A value that captured nothing adds no link, so it's never related,
// and values of one type are related once, by their type.
type closureGroups struct {
	// types are the distinct types of the function values that captured something, in the order first met.
	types []*checker.Type
	// captures are those values in closureRecords' order, so the links come out in the order relating
	// each value on its own added them.
	captures []closureCapture
	// records is how many closureRecords these were grouped from.
	records int
}

// closureCapture is one function value that captured something: its type's index in closureGroups.types,
// and the locals it captured (closureLinks).
type closureCapture struct {
	group  int
	locals []int
}

// groupClosures groups the lowering's closureRecords by type (closureGroups).
func groupClosures(l *lowering) *closureGroups {
	groups := &closureGroups{records: len(l.closureRecords)}
	groupOf := map[*checker.Type]int{}
	for _, closure := range l.closureRecords {
		locals := l.closureLinks(closure)
		if len(locals) == 0 {
			continue
		}
		group, grouped := groupOf[closure.proven]
		if !grouped {
			group = len(groups.types)
			groupOf[closure.proven] = group
			groups.types = append(groups.types, closure.proven)
		}
		groups.captures = append(groups.captures, closureCapture{group: group, locals: locals})
	}
	return groups
}

// closureLinksTo is the cells captured by every function value the program makes that can be seen as
// proven, which is what a value of that function type holds, in closureRecords' order. Each distinct type
// of a value that captured something is related to proven once (closureGroups).
func (f *cycleFinder) closureLinksTo(proven *checker.Type) []cycleNode {
	if f.closures == nil || f.closures.records != len(f.l.closureRecords) {
		f.closures = groupClosures(f.l)
	}
	assignable := make([]bool, len(f.closures.types))
	for group, closureType := range f.closures.types {
		f.closureQueries++
		assignable[group] = f.l.checker.IsTypeAssignableTo(closureType, proven)
	}
	links := []cycleNode{}
	for _, capture := range f.closures.captures {
		if !assignable[capture.group] {
			continue
		}
		for _, local := range capture.locals {
			links = append(links, cycleNode{cell: local + 1})
		}
	}
	return links
}

func (f *cycleFinder) deriveRelated(one *checker.Type, other *checker.Type) bool {
	if one == other {
		return true
	}
	if f.isFunction(one) || f.isFunction(other) {
		return false
	}
	// A fresh literal is related by its regular widened type, as shape holds it beside the fresh one
	// (the census): the relation judges a fresh type for excess properties, so { name, back } as Named would
	// otherwise never be what a value seen as Named reaches (#sbvnh96).
	one, other = f.l.relatedShape(one), f.l.relatedShape(other)
	return f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)
}

// viewAlsoOf is the source types a checked view can hand out as proven that the finder doesn't relate to it
// (viewAlsoTable), for this finder's views.
func (f *cycleFinder) viewAlsoOf(proven *checker.Type) []*checker.Type {
	if f.viewAlso == nil {
		f.viewAlso = viewAlsoTable(f.l, f.related, f.isFunction, f.shapes)
	}
	return f.viewAlso[proven]
}

// viewAlsoTable is, for each type a checked view lets the program see a value through, the source types the
// view can hand out there that related doesn't relate to it (#9489gt0, Lower's ruling). A view tests at most its
// top: interfaceCast's tag, a member cast's or a kind cast's (view.cast). Beneath that nothing is tested, so a
// field the target narrows (Base.f: Y1 | Y2 seen as Narrow.f: Y1) hands out a Y2 as a Y1, and a Y1 slot it's
// stored in holds what a Y2 holds. The finder took two types for one allocation only when they relate, so that
// cycle went unseen and leaked, in a graph program or a plain one. Each view's source and target are walked in
// step, through fields, index values, and array, tuple, Map and Set elements, and every source arm beneath an
// untested position that doesn't relate to a target arm there is an edge from that arm to it: reachLinks and
// graphLinks follow it, so the cycle is found and held as a graph region rather than stopped.
//
// It's keyed by type, so it's program-wide: every Y1 links to Y2, not only the one under Narrow, which finds
// more, never fewer. A literal at depth rules no arm out, since no read beneath a view checks one; were it to
// prune an edge, the next read road that skips its check would reopen this. A related arm pairs on, its fields
// with the target's. An unrelated one pairs on too, since a value read out of it through the target's field is
// its field's value seen as the target's. The walk is bounded by the pairs it has met, and by typeNesting: where
// a generic grows deeper on its path, the whole source is linked to the target and the walk stops there, which
// reachLinks then walks on its own bound.
//
// A view with no source names no type its value was. A function constructor's new is one of its candidates, each
// laid by the writes through its declared this (functionConstructor.thisType), so each this type is the source.
// The other road that records none today, a speculative result, admits only scalar fields, so nothing beneath it
// can be handed out. Any view with no source and no candidates whose target holds a field that isn't a plain
// scalar or a function links the target to every shape the program has, failing closed (Lower's should 2).
func viewAlsoTable(l *lowering, related func(one, other *checker.Type) bool, isFunction func(*checker.Type) bool, shapes []*checker.Type) map[*checker.Type][]*checker.Type {
	table := map[*checker.Type][]*checker.Type{}
	linked := map[[2]*checker.Type]bool{}
	link := func(viewed, held *checker.Type) {
		if viewed == held || linked[[2]*checker.Type{viewed, held}] {
			return
		}
		linked[[2]*checker.Type{viewed, held}] = true
		table[viewed] = append(table[viewed], held)
	}
	skip := func(proven *checker.Type) bool {
		return proven == nil || viewPairScalar(proven) || isFunction(proven)
	}
	arms := func(proven *checker.Type, flags checker.TypeFlags) []*checker.Type {
		if proven.Flags()&flags != 0 {
			return proven.Types()
		}
		return []*checker.Type{proven}
	}
	container := func(proven *checker.Type) bool {
		return l.isArrayType(proven) || checker.IsTupleType(proven) || l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")
	}
	type step struct {
		source, target *checker.Type
		tested         bool
	}
	walked := map[step]bool{}
	var nesting typeNesting
	var pair func(source, target *checker.Type, tested bool)
	members := func(source, target *checker.Type) {
		if container(source) && container(target) {
			sourceArguments, targetArguments := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			for index, targetArgument := range targetArguments {
				switch {
				case index < len(sourceArguments) && (checker.IsTupleType(source) == checker.IsTupleType(target) || checker.IsTupleType(target)):
					pair(sourceArguments[index], targetArgument, false)
				case len(sourceArguments) == 1 && l.isArrayType(source):
					pair(sourceArguments[0], targetArgument, false)
				}
			}
			if checker.IsTupleType(source) && l.isArrayType(target) && len(targetArguments) == 1 {
				for _, sourceArgument := range sourceArguments {
					pair(sourceArgument, targetArguments[0], false)
				}
			}
		}
		for _, info := range l.checker.GetIndexInfosOfType(target) {
			for _, own := range l.checker.GetIndexInfosOfType(source) {
				if own.KeyType() == info.KeyType() {
					pair(own.ValueType(), info.ValueType(), false)
				}
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(target) {
			held := l.checker.GetPropertyOfType(source, field.Name)
			if held == nil {
				// A source with an index signature holds the field among its values.
				for _, own := range l.checker.GetIndexInfosOfType(source) {
					pair(own.ValueType(), l.checker.GetTypeOfSymbol(field), false)
				}
				continue
			}
			pair(l.checker.GetTypeOfSymbol(held), l.checker.GetTypeOfSymbol(field), false)
		}
	}
	pair = func(source, target *checker.Type, tested bool) {
		source, target = l.concrete(source), l.concrete(target)
		if skip(source) || skip(target) {
			return
		}
		present, viewed := l.checker.GetNonNullableType(source), l.checker.GetNonNullableType(target)
		if present == viewed || walked[step{present, viewed, tested}] {
			return
		}
		walked[step{present, viewed, tested}] = true
		enteredSource, leaveSource := nesting.enter(present)
		defer leaveSource()
		enteredTarget, leaveTarget := nesting.enter(viewed)
		defer leaveTarget()
		if !enteredSource || !enteredTarget {
			for _, seen := range arms(viewed, checker.TypeFlagsUnion|checker.TypeFlagsIntersection) {
				link(seen, present)
			}
			return
		}
		for _, held := range arms(present, checker.TypeFlagsUnion) {
			if skip(held) {
				continue
			}
			for _, seen := range arms(viewed, checker.TypeFlagsUnion|checker.TypeFlagsIntersection) {
				if skip(seen) {
					continue
				}
				if related(held, seen) {
					members(held, seen)
					continue
				}
				// The cast's tag test told these arms apart at the top.
				if tested {
					continue
				}
				link(seen, held)
				members(held, seen)
			}
		}
	}
	for _, view := range l.viewSources {
		if view.source != nil {
			pair(view.source, view.target, view.cast)
			continue
		}
		candidates, sourced := 0, 0
		if roads := l.functionConstructorScan; roads != nil {
			for _, made := range l.functionConstructorNews {
				if made.node != view.node {
					continue
				}
				for _, index := range made.candidates {
					candidates++
					if index >= 0 && index < len(roads.ordered) && roads.ordered[index].thisType != nil {
						pair(roads.ordered[index].thisType, view.target, false)
						sourced++
					}
				}
			}
		}
		if candidates != 0 && sourced == candidates {
			continue
		}
		target := l.checker.GetNonNullableType(l.concrete(view.target))
		for _, field := range l.checker.GetPropertiesOfType(target) {
			if held := l.checker.GetTypeOfSymbol(field); skip(held) {
				continue
			}
			for _, seen := range arms(target, checker.TypeFlagsUnion|checker.TypeFlagsIntersection) {
				for _, shape := range shapes {
					link(seen, shape)
				}
			}
			break
		}
	}
	return table
}

func (f *cycleFinder) shapeCandidates(proven *checker.Type) []*checker.Type {
	if f.shapeIndex == nil || f.shapeIndexSize != len(f.shapes) {
		f.shapeIndex = map[string][]*checker.Type{}
		f.shapeChoices = nil
		included := map[*checker.Type]bool{}
		for _, shape := range f.shapes {
			shape = f.literalType(shape)
			if included[shape] {
				continue
			}
			included[shape] = true
			f.shapeChoices = append(f.shapeChoices, shape)
			for _, property := range f.l.checker.GetPropertiesOfType(shape) {
				f.shapeIndex[property.Name] = append(f.shapeIndex[property.Name], shape)
			}
		}
		f.shapeIndexSize = len(f.shapes)
	}
	candidates := f.shapeChoices
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsOptional == 0 && len(f.shapeIndex[property.Name]) < len(candidates) {
			candidates = f.shapeIndex[property.Name]
		}
	}
	return candidates
}

func (f *cycleFinder) literalType(proven *checker.Type) *checker.Type {
	if prior := f.literalTypes[proven]; prior != nil {
		return prior
	}
	if f.literalTypes == nil {
		f.literalTypes = map[*checker.Type]*checker.Type{}
	}
	actual := proven
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsObjectLiteral != 0 && !f.isFunction(proven) && len(f.l.checker.GetIndexInfosOfType(proven)) == 0 {
		properties := f.l.checker.GetPropertiesOfType(proven)
		// A paired setter can have a write type different from its getter's read
		// type. Keep accessor identities rather than reducing them to read slots.
		for _, property := range properties {
			if accessorSymbol(property) {
				f.literalTypes[proven] = proven
				return proven
			}
		}
		trie := &f.literalSlots
		for _, property := range properties {
			key := cycleSlotKey{
				flags:    property.Flags,
				held:     (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property),
				name:     property.Name,
				proven:   f.l.checker.GetTypeOfSymbol(property),
				optional: property.Flags&ast.SymbolFlagsOptional != 0,
				readonly: f.l.checker.IsReadonlySymbol(property),
			}
			if trie.next == nil {
				trie.next = map[cycleSlotKey]*cycleSlotTrie{}
			}
			next := trie.next[key]
			if next == nil {
				next = &cycleSlotTrie{}
				trie.next[key] = next
			}
			trie = next
		}
		flags := proven.ObjectFlags()
		if trie.proven == nil {
			trie.proven = map[checker.ObjectFlags]*checker.Type{}
		}
		if prior := trie.proven[flags]; prior != nil {
			actual = prior
		} else {
			trie.proven[flags] = proven
		}
	}
	f.literalTypes[proven] = actual
	return actual
}

func (f *cycleFinder) cycleUnproven(kind fresh.WriteKind, holder *checker.Type, name string) *fresh.Write {
	if f.writes == nil {
		f.writes = fresh.ProveWrites(f.l.result)
	}
	if f.writeIndex == nil {
		f.writeIndex = map[cycleWriteKey][]int{}
		f.writeCache = map[cycleWriteSlot]*fresh.Write{}
		f.unknownWrite = -1
		for index, write := range f.writes {
			if write.Proven {
				continue
			}
			if write.Kind == fresh.WriteUnknown {
				if f.unknownWrite < 0 {
					f.unknownWrite = index
				}
				continue
			}
			key := cycleWriteKey{write.Kind, write.Name}
			if write.Kind != fresh.WriteField {
				key.name = ""
			}
			f.writeIndex[key] = append(f.writeIndex[key], index)
		}
	}
	slot := cycleWriteSlot{holder, kind, name}
	if write, ok := f.writeCache[slot]; ok {
		return write
	}
	candidates := f.writeIndex[cycleWriteKey{kind, name}]
	// A private field has one source symbol and a written name for each native
	// instantiation. Match those names once, preserving the old write order.
	if kind == fresh.WriteField {
		property := f.l.checker.GetPropertyOfType(holder, name)
		if property != nil && len(property.Declarations) == 1 {
			declared := property.Declarations[0].Name()
			if declared != nil && declared.Kind == ast.KindPrivateIdentifier {
				candidates = append([]int(nil), candidates...)
				for key, indices := range f.writeIndex {
					if key.kind == kind && key.name != name && f.l.cycleFieldMatches(holder, name, key.name) {
						candidates = append(candidates, indices...)
					}
				}
			}
		}
	}
	first := f.unknownWrite
	for _, index := range candidates {
		if first >= 0 && index >= first {
			continue
		}
		write := &f.writes[index]
		if write.Site == 0 {
			first = index
			continue
		}
		written := f.l.writeSites[write.Site-1].holder
		if written == nil || f.template(written) || f.related(holder, written) {
			first = index
		}
	}
	var write *fresh.Write
	if first >= 0 {
		write = &f.writes[first]
	}
	f.writeCache[slot] = write
	return write
}

type cycleReach struct {
	from   *checker.Type
	target cycleNode
}

type cycleSlotKey struct {
	flags    ast.SymbolFlags
	held     bool
	name     string
	proven   *checker.Type
	optional bool
	readonly bool
}

type cycleSlotTrie struct {
	next   map[cycleSlotKey]*cycleSlotTrie
	proven map[checker.ObjectFlags]*checker.Type
}

type cycleWriteKey struct {
	kind fresh.WriteKind
	name string
}

type cycleWriteSlot struct {
	holder *checker.Type
	kind   fresh.WriteKind
	name   string
}

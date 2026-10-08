package lower

import (
	"fmt"
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

	// A type's slots and strong links do not change during a walk. Keep their checker
	// queries and compatible views once, rather than deriving them for every slot.
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
}

// findCycles refuses the first cycle-capable slot that isn't declared Weak and has a write that isn't
// proven not to close a cycle (fresh.go), or returns nil.
func (l *lowering) findCycles(modules []*ast.SourceFile) error {
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
	return l.cycleTypes(modules).graphTypes(modules)
}

// cycleTypes notes the source views and allocation shapes before following slots.
func (l *lowering) cycleTypes(modules []*ast.SourceFile) *cycleFinder {
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
	return finder
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
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
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
func (f *cycleFinder) deriveTemplate(proven *checker.Type) bool {
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
func (f *cycleFinder) deriveFields(proven *checker.Type) []*ast.Symbol {
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
func (f *cycleFinder) deriveIsFunction(proven *checker.Type) bool {
	return len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || (len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && (!f.l.isStaticType(proven) || len(f.l.staticGlobals) == 0))
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
func (f *cycleFinder) deriveWeak(proven *checker.Type) bool {
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
			write := f.cycleUnproven(fresh.WriteElement, holder, "")
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
			if write := f.cycleUnproven(fresh.WriteSetElement, holder, ""); write != nil {
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
			if write := f.cycleUnproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a map whose keys can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlyMap, built whole when it's made; or set in such a map only keys this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
		if len(arguments) == 2 && !f.weak(arguments[1]) && f.reaches(arguments[1], cycleNode{proven: holder}) {
			if write := f.cycleUnproven(fresh.WriteMapEntry, holder, ""); write != nil {
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
			write := f.cycleUnproven(fresh.WriteField, holder, field.Name)
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

type cycleReach struct {
	from   *checker.Type
	target cycleNode
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

// reachLinks holds exactly the strong links followed by the slot finder. Its cell
// edges differ from the ownership graph's: only an async cell retains its owner.
func (f *cycleFinder) reachLinks(node cycleNode) []cycleNode {
	links := []cycleNode{}
	if node.cell != 0 {
		// An interior async cell retains its owner, not just its declared value.
		// Completion drops private slots; captured slots remain reachable through
		// any escaped cell. Follow every such slot irrespective of signature.
		local := f.l.result.Locals[node.cell-1]
		if local.EnvironmentCell && local.Function >= 0 {
			owner := f.l.result.Functions[local.Function]
			if owner.Async {
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
	switch {
	case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
		links = append(links, f.libraryIteratorCaptures(proven)...)
	case f.isFunction(proven):
		// Construct signatures can hide constructor objects behind an interface.
		if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range f.l.statics {
				actual := f.l.checker.GetTypeOfSymbol(symbol)
				if f.l.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
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
				links = append(links, cycleNode{cell: local + 1})
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
		for _, accessor := range f.l.accessorCaptures {
			if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range f.l.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
		for _, field := range f.fields(proven) {
			links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
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

func (f *cycleFinder) deriveRelated(one *checker.Type, other *checker.Type) bool {
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

// A structural view with a required property can only hold shapes with that
// property. Use the shortest posting list; the checker still decides compatibility.
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

// Literal types with identical properties and strong slots have the same
// structural views. Accessors stay distinct from stored fields. Keep one walk for them; their allocation identities remain
// separate in where and in the ownership graph. Names are indexed once per type,
// rather than formatting a structural key at every visit.
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

// A closed literal tree assigned straight to a global has no identity the fresh
// proof can name later: global reads are outside values. Its allocation and
// escape cannot affect any other value. Leave that dead heap out of the proof,
// keeping the instruction itself and every initializer with local reads or effects.
// A global read is already outside in this proof and escaping outside changes no state.
func (f *cycleFinder) cycleUnproven(kind fresh.WriteKind, holder *checker.Type, name string) *fresh.Write {
	if f.writes == nil {
		program := *f.l.result
		program.Main = append([]ir.Statement(nil), program.Main...)
		for i, statement := range program.Main {
			switch statement := statement.(type) {
			case ir.Declare:
				if f.l.result.Locals[statement.Local].Global && f.cycleLiteralTree(statement.Value) {
					statement.Value = ir.Undefined{Of: statement.Value.Type()}
					program.Main[i] = statement
				}
			case ir.Assign:
				if f.l.result.Locals[statement.Local].Global && f.cycleLiteralTree(statement.Value) {
					statement.Value = ir.Undefined{Of: statement.Value.Type()}
					program.Main[i] = statement
				}
			}
		}
		f.writes = fresh.ProveWrites(&program)
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
func (f *cycleFinder) cycleLiteralTree(value ir.Expression) bool {
	switch value := value.(type) {
	case ir.Read:
		return f.l.result.Locals[value.Local].Global && !value.Checked
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.JSONNull, ir.Null:
		return true
	case ir.Unary:
		return f.cycleLiteralTree(value.Operand)
	case ir.Binary:
		return f.cycleLiteralTree(value.Left) && f.cycleLiteralTree(value.Right)
	case ir.ArrayLiteral:
		for _, spread := range value.Spread {
			if spread {
				return false
			}
		}
		for _, element := range value.Elements {
			if !f.cycleLiteralTree(element) {
				return false
			}
		}
		return true
	case ir.ObjectLiteral:
		if value.Spread != nil || len(value.Methods) != 0 {
			return false
		}
		for _, field := range value.Fields {
			if !f.cycleLiteralTree(field.Value) {
				return false
			}
		}
		for _, field := range value.Empty {
			if !f.cycleLiteralTree(field.Value) {
				return false
			}
		}
		return true
	}
	return false
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

package lower

import (
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
}

// findCycles refuses the first cycle-capable slot that isn't declared Weak and has a write that isn't
// proven not to close a cycle (fresh.go), or returns nil.
func (l *lowering) findCycles(modules []*ast.SourceFile) error {
	if err := l.checkDiscriminantConstruction(modules); err != nil {
		return err
	}
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if !finder.iteratorCollectionStaysEmpty(node, modules) {
				finder.libraryIteratorMade(node)
			}
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

// iteratorCollectionStaysEmpty proves a narrow acyclic case: a const bound to an empty library
// collection, used only to make iterators. With no alias, insertion or reassignment anywhere in
// the program, that collection can hold nothing, even when its declared element type can reach back.
func (f *cycleFinder) iteratorCollectionStaysEmpty(node *ast.Node, modules []*ast.SourceFile) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !ast.IsIdentifier(receiver) || !f.l.isLibraryType(f.l.checker.GetTypeAtLocation(receiver), "Map", "Set") {
		return false
	}
	symbol := f.l.checker.GetSymbolAtLocation(receiver)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || initializer.Kind != ast.KindNewExpression {
		return false
	}
	made := initializer.AsNewExpression()
	if (made.Arguments != nil && len(made.Arguments.Nodes) != 0) || (!f.l.isLibraryGlobal(made.Expression, "Map") && !f.l.isLibraryGlobal(made.Expression, "Set")) {
		return false
	}
	empty := true
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(use *ast.Node) bool {
			if ast.IsIdentifier(use) && use != declaration.Name() && f.l.checker.GetSymbolAtLocation(use) == symbol {
				access := use.Parent
				if access == nil || access.Kind != ast.KindPropertyAccessExpression || access.AsPropertyAccessExpression().Expression != use {
					empty = false
					return false
				}
				name := access.Name().Text()
				call := access.Parent
				if (name != "keys" && name != "values" && name != "entries") || call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != access || len(call.AsCallExpression().Arguments.Nodes) != 0 {
					empty = false
				}
			}
			return use.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	return empty
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
	if element := f.l.recordElement(proven); element != nil {
		f.use(element, where)
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
	case l.recordElement(holder) != nil:
		element := l.recordElement(holder)
		if !f.weak(element) && f.reaches(element, cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{Where: l.program.Where(f.where[holder]), What: name + ", a record whose values can reach back to its holder: a cycle reference counting cannot free, and " + f.writtenAt(write) + " may close one", Fix: "use weak links or write only values proven unable to reach the record (adamic/cycle-capable)"}
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
		if f.weak(proven) {
			// A Weak holds nothing.
			continue
		}
		// An iterator can flow into any interface it satisfies. Follow its hidden collection
		// before skipping templates: Iterator's default return and next arguments are any,
		// but they don't hide the collection held by a concrete iterator made here.
		if flags&checker.TypeFlagsObject != 0 {
			queue = append(queue, f.libraryIteratorCaptures(proven)...)
		}
		if f.template(proven) {
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
		case f.l.recordElement(proven) != nil:
			queue = append(queue, cycleNode{proven: f.l.recordElement(proven)})
		case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
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

package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsc hands a NodeArray, held as an array, to slots typed TextRange ({ pos: number; end: number }), which an
// object type would hold as ir.Object (#hae423v). Adamic's (b1) ruling on #jjxtwzs says ir.Object promises an
// adamic object, so a type whose value can be an array at run time is held as ir.Union, and every consumer asks
// the kind (range_unions.go classes each one). A closed range type is such a type: an object type whose
// properties are only pos and end, each a required number, with no call, construct or index signature, and no
// class instance. It's held as a Union in a program that has the NodeArray layout at all, whatever reaches
// what, and as an object everywhere else, so a program with no NodeArray lowers as it did.
//
// The program bit is a superset, never a walk that can miss (Lower's Must 1 on #hae423v): it's on when any
// interface the program declares, in any file, the bundled library and the prelude included, has an array in
// its ancestry (nodeArrayBase) and declares pos and end, which NodeArrayLayout requires of every NodeArray.
// That holds for every type nodeArrayLayoutOf recognizes, generic or not, and asks nothing of which value
// goes where. The bit's other half, that the program has a closed range type, is the per-type predicate
// itself: a program with none holds no type differently, whatever the bit says.

// rangeHeldAsUnion is representationReason's question: whether a type, already concrete, is a closed range
// type held as a Union. The bit is asked first, so a program without the layout asks the checker nothing more.
func (l *lowering) rangeHeldAsUnion(proven *checker.Type) bool {
	return l.rangeArraysPresent() && l.rangeShape(proven)
}

// rangeArraysPresent is the program bit, computed once, before lowering asks any representation
// (lowerWithElementKey), or at the first question of a lowering built another way. A lowering with no program
// to read fails closed: every closed range type is held as a Union there.
func (l *lowering) rangeArraysPresent() bool {
	if l.rangeArraysKnown {
		return l.rangeArrays
	}
	if l.program == nil {
		return true
	}
	l.rangeArrays, l.rangeArraysKnown = l.nodeArrayDeclared(), true
	return l.rangeArrays
}

// nodeArrayDeclared is whether any file of the program declares an interface nodeArrayInterface accepts. A
// declaration file has no bodies, so its interfaces are its statements' and its namespaces' alone, and only
// those are walked; every other file is walked whole, since an interface may be declared in a function's body.
func (l *lowering) nodeArrayDeclared() bool {
	seen := map[*ast.Symbol]bool{}
	var statements, whole ast.Visitor
	statements = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration:
			return l.nodeArrayInterface(node, seen)
		case ast.KindModuleDeclaration, ast.KindModuleBlock:
			return node.ForEachChild(statements)
		}
		return false
	}
	whole = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration && l.nodeArrayInterface(node, seen) {
			return true
		}
		return node.ForEachChild(whole)
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		visit := whole
		if file.IsDeclarationFile || load.IsLibrary(file) {
			visit = statements
		}
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// nodeArrayInterface is whether an interface declaration's type may have the NodeArray layout: an array in
// its ancestry, as nodeArrayBase finds one, and pos and end among its properties. Array and ReadonlyArray
// themselves are no NodeArray (nodeArrayLayoutOf). A declaration with no extends clause is passed over, since
// the symbol's declaration that has one is visited too.
func (l *lowering) nodeArrayInterface(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	if len(nodesOf(node.AsInterfaceDeclaration().HeritageClauses)) == 0 {
		return false
	}
	symbol := l.checker.GetSymbolOfDeclaration(node)
	if symbol == nil || seen[symbol] {
		return false
	}
	seen[symbol] = true
	declared := l.checker.GetDeclaredTypeOfSymbol(symbol)
	if declared == nil || l.isLibraryType(declared, "Array", "ReadonlyArray") || l.nodeArrayBase(declared, map[*checker.Type]bool{}) == nil {
		return false
	}
	return l.checker.GetPropertyOfType(declared, "pos") != nil && l.checker.GetPropertyOfType(declared, "end") != nil
}

// rangeShape is whether a type is a closed range type, memoized a lowering: the answer is the checker's
// alone, whatever the program bit says.
func (l *lowering) rangeShape(proven *checker.Type) bool {
	if shaped, known := l.rangeShapes[proven]; known {
		return shaped
	}
	shaped := l.rangeShapeAsked(proven)
	if l.rangeShapes == nil {
		l.rangeShapes = map[*checker.Type]bool{}
	}
	l.rangeShapes[proven] = shaped
	return shaped
}

// rangeShapeAsked is rangeShape's answer, asked of the checker: an object type, no array, tuple or class
// instance, with no call, construct or index signature, and at least one property, each of them pos or end, a
// required data property of type number. A getter named pos is no data property, and a class instance's
// fields have the class's layout, which an object read of the range would take as any object's.
func (l *lowering) rangeShapeAsked(proven *checker.Type) bool {
	if proven == nil || proven.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(proven) || checker.IsTupleType(proven) || isClassInstance(proven) {
		return false
	}
	if len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return false
	}
	properties := l.checker.GetPropertiesOfType(proven)
	if len(properties) == 0 {
		return false
	}
	for _, property := range properties {
		if property.Name != "pos" && property.Name != "end" {
			return false
		}
		if property.Flags&ast.SymbolFlagsProperty == 0 || property.Flags&ast.SymbolFlagsOptional != 0 {
			return false
		}
		if l.checker.GetTypeOfSymbol(property).Flags() != checker.TypeFlagsNumber {
			return false
		}
	}
	return true
}

// holdsRangeUnion is whether a value of a type may be held as a range type's Union: the type is a closed
// range type held as one, or a union with such a member. A type parameter no instantiation fixed may be one,
// so it fails closed.
func (l *lowering) holdsRangeUnion(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.holdsRangeUnion(member) {
				return true
			}
		}
		return false
	}
	return l.rangeHeldAsUnion(proven)
}

// rangeWhereObjectGoes is the stop for a value held as a range type's Union going where its slot holds an
// object or an array, or "" when it isn't one: tsc lets a TextRange go where a wider object type with only
// optional members beyond pos and end is expected, and fit hands a Union to such a slot unchanged, which the
// slot would read as its own kind without asking which the Union holds. The value site is where the checker
// still knows both types; range_unions.go's table classes this road as stopping. A type parameter no
// instantiation fixed is judged where it's concrete, as rangeReceiver judges one.
func (l *lowering) rangeWhereObjectGoes(node *ast.Node, value ir.Expression, slot *checker.Type) string {
	if value == nil || value.Type() != ir.Union || !l.rangeArraysPresent() {
		return ""
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if proven == nil || proven.Flags()&checker.TypeFlagsTypeParameter != 0 || !l.holdsRangeUnion(proven) {
		return ""
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.Name().Kind == ast.KindObjectBindingPattern && parent.AsVariableDeclaration().Initializer != nil && ast.SkipParentheses(parent.AsVariableDeclaration().Initializer) == node && l.rangeOnly(proven) {
		// const { pos, end } = range: the slot is the pattern's own type, at which nothing is held, and
		// destructureFrom reads each field of a range type's Union by kind (rangeDestructure), or stops.
		return ""
	}
	held, known := l.representation(slot)
	if !known || held != ir.Object && held != ir.Array {
		return ""
	}
	kind := "an object"
	if held == ir.Array {
		kind = "an array"
	}
	return "a value of range type " + l.checker.TypeToString(proven) + ", held as a union where a NodeArray may stand for it, going where " + l.checker.TypeToString(slot) + " is held as " + kind
}

// rangeReceiver is whether a field access's receiver is held as a Union because its type, without null and
// undefined, is a closed range type or a union with one.
func (l *lowering) rangeReceiver(receiver *ast.Node) bool {
	if !l.rangeArraysPresent() {
		return false
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(receiver))
	if proven == nil {
		return false
	}
	proven = l.checker.GetNonNullableType(proven)
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		// A type parameter no instantiation fixed is no range type's receiver here: its reads lower as the
		// instantiation's, where the parameter is concrete.
		return false
	}
	return l.holdsRangeUnion(proven)
}

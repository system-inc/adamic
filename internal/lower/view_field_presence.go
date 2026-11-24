package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A presence-only read of a field through a checked view answers what Node answers (Adamic's ruling on
// #y5n0x0j, Nov 24 19:10Z): `!!(value as T).f`, `(value as T).f` as a whole condition or under !, and
// `(value as T).f === undefined` or `!== undefined`. A view stops at its first read of a field its value
// lacks (untaggedUnionCast), which is right for a read whose value is used, and wrong for valid code asking
// whether the cast was true: Node answers false there. tsc's isFileWatcherWithModifiedTime asks it of a Date
// (`!!(value as FileWatcherWithModifiedTime).watcher`). Only a boolean leaves the test, so no value ever
// carries the field's type (Language's read of the premise, #y5n0x0j 19:15Z). The test reads the value's
// own slot, absence included, as a structural method's presence test does (MethodTruthiness), and makes
// no view.
//
// Adamic's three conditions are kept, with Language's two:
//   - Presence-only: truthinessOnly's positions (the operand of !, a condition, the right of an && tested
//     for truthiness alone), or one side of === or !== whose other side is undefined. typeof, ||, ??, a
//     value-reading && and == undefined, which null answers too, keep the view's read.
//   - Node's answer on every value: Node looks a name up through the prototype chain, and the test reads
//     the value's own slot, its class's methods and a record's own keys (adamic_object_method_truthy). The
//     two agree only where no prototype the value can have holds the name, so the road is taken only when
//     every member of the cast's source is a plain object the program declares (no class instance, array,
//     callable or dictionary) or the library's Date, no source member's type has the name as a method or
//     an accessor, Date's own members (getTime and its kin) are never the name, and the name is none of
//     Object.prototype's (objectPrototypeName). A program class instance seen as a plain member answers
//     through its class's methods natively and its prototype on Node, alike.
//   - No accessor answers the name: a name any accessor in the program declares (accessorNames) keeps the
//     view's read, and a library accessor written in lib.d.ts as a property (Map's size, RegExp's
//     lastIndex) is out with its member, since only Date is admitted among the library's.
//
// Native's Date keeps its time in a private slot named with #, as its other private host slots are
// (native/fields.go), so no field name a program can write reaches it. Anything this test doesn't take
// keeps the view's read and its stop, which fails closed.
func (l *lowering) viewFieldPresence(node *ast.Node, defined bool) (ir.Expression, bool, error) {
	operand := l.viewFieldPresenceOperand(node, defined)
	if operand == nil {
		return nil, false, nil
	}
	value, err := l.expression(operand)
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Object {
		return nil, true, l.notYet(node, "a presence test of a field through a view whose value isn't held as an object")
	}
	return ir.Property{Object: value, Name: l.fieldName(node.AsPropertyAccessExpression().Name()), Of: ir.Boolean, MethodTruthiness: !defined, SlotDefined: defined}, true, nil
}

// viewFieldPresenceOperand is the cast's operand when viewFieldPresence takes node, a truthiness test unless
// defined, and nil when the read keeps its view. It lowers nothing, so a consumer can ask before it lowers.
func (l *lowering) viewFieldPresenceOperand(node *ast.Node, defined bool) *ast.Node {
	if node.Kind != ast.KindPropertyAccessExpression || l.questionDot(node) || l.inOptionalChain(node) {
		return nil
	}
	if !defined && !truthinessOnly(node) {
		return nil
	}
	access := node.AsPropertyAccessExpression()
	cast := ast.SkipParentheses(access.Expression)
	if cast.Kind != ast.KindAsExpression || !l.presenceFieldDeclared(node) {
		return nil
	}
	proof, err := l.castProof(cast)
	if err != nil || !proof.untagged && !proof.overlap && !proof.structuralView {
		// A cast that stops keeps its own words, which lowering the cast as a value gives.
		return nil
	}
	source := l.concrete(l.checker.GetTypeAtLocation(cast.AsAsExpression().Expression))
	if l.includesUndefined(source) || l.includesNull(source) || !l.presenceAnsweredOwn(source, access.Name().Text()) {
		return nil
	}
	return cast.AsAsExpression().Expression
}

// presenceFieldDeclared is whether node reads a required data field each declaration of which is the
// program's own property: an optional one already reads as absent through the view, and a method, an
// accessor or a library's field isn't the test's to answer.
func (l *lowering) presenceFieldDeclared(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node.AsPropertyAccessExpression().Name())
	if symbol == nil || len(symbol.Declarations) == 0 || symbol.Flags&ast.SymbolFlagsOptional != 0 || accessorSymbol(symbol) || l.accessorNames[symbol.Name] {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if declaration.Kind != ast.KindPropertySignature && declaration.Kind != ast.KindPropertyDeclaration || load.IsLibrary(file) || load.IsNodeLibrary(file) || load.IsPrelude(file) {
			return false
		}
	}
	return true
}

// presenceAnsweredOwn is whether every value of source answers a lookup of name from its own slots as
// Node answers it through its prototype chain (viewFieldPresence's second condition).
func (l *lowering) presenceAnsweredOwn(source *checker.Type, name string) bool {
	if l.objectPrototypeName(name) {
		return false
	}
	for _, member := range castMembers(source) {
		if l.isLibraryType(member, "Date") {
			if l.checker.GetPropertyOfType(l.checker.GetApparentType(member), name) != nil {
				return false
			}
			continue
		}
		if !l.plainViewObject(member) || !l.programDeclared(member) {
			return false
		}
		if property := l.checker.GetPropertyOfType(member, name); property != nil && (property.Flags&ast.SymbolFlagsMethod != 0 || accessorSymbol(property)) {
			return false
		}
	}
	return true
}

// programDeclared is whether every declaration of a type's symbol is in the program's own files.
func (l *lowering) programDeclared(proven *checker.Type) bool {
	symbol := proven.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if load.IsLibrary(file) || load.IsNodeLibrary(file) || load.IsPrelude(file) {
			return false
		}
	}
	return true
}

// viewFieldDefined lowers `(value as T).f === undefined` and `!== undefined` when viewFieldPresence takes the
// read: whether the value holds the field as something other than undefined, negated for ===.
func (l *lowering) viewFieldDefined(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind
	if operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
		return nil, false, nil
	}
	read, other := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	if !undefinedIdentifier(other) {
		read, other = other, read
	}
	if !undefinedIdentifier(other) {
		return nil, false, nil
	}
	present, handled, err := l.viewFieldPresence(read, true)
	if !handled || err != nil {
		return nil, handled, err
	}
	if operator == ast.KindEqualsEqualsEqualsToken {
		return ir.Unary{Operator: ir.Not, Operand: present}, true, nil
	}
	return present, true, nil
}

// undefinedIdentifier is whether node is the identifier undefined, read as comparedWithUndefined reads it.
func undefinedIdentifier(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier && node.Text() == "undefined"
}

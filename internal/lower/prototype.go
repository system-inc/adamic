package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// libraryMember follows the declaration, not the spelling: an own callback named toString is
// still an own field. Every declaration must be from the library to specialize a call.
func (lowering *lowering) memberSymbol(node *ast.Node) *ast.Symbol {
	symbol := lowering.checker.GetSymbolAtLocation(node)
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			symbol = lowering.checker.GetPropertyOfType(
				lowering.checker.GetTypeAtLocation(access.Expression),
				key.Text(),
			)
		}
	}
	return symbol
}

func (lowering *lowering) libraryMember(node *ast.Node) bool {
	return lowering.librarySymbol(lowering.memberSymbol(node))
}

// A union can combine an own callback with an inherited method. That still cannot be an
// unconditional own-field load: any library root is enough to refuse the read.
func (lowering *lowering) inheritedLibraryMember(node *ast.Node) bool {
	return lowering.inheritedLibrarySymbol(lowering.memberSymbol(node))
}

func (lowering *lowering) inheritedLibrarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, root := range lowering.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return true
			}
		}
	}
	return false
}

// iteratorResultField reports whether done or value is read off a library iterator result as an own
// field, never a prototype member: a result object next returns holds both as its own on Node, whatever
// made it. It's the one home of that rule (#5t28afa), asked by every property read, element read and
// destructuring, in two arms:
//   - every library root of symbol is IteratorReturnResult's, both fields required and own. The prelude's
//     Map and Set iterators return one as their done arm, so inference reads that arm as the iterator's
//     return rather than its value;
//   - for a read off receiver (nil for a destructuring, which keeps the first arm alone, as before this
//     move), every member of the receiver's type is the library's IteratorYieldResult or
//     IteratorReturnResult, the result a hand-driven next() gives (#qk8rztp S2c-2a).
func (lowering *lowering) iteratorResultField(receiver *ast.Node, symbol *ast.Symbol) bool {
	if symbol == nil || (symbol.Name != "done" && symbol.Name != "value") {
		return false
	}
	if lowering.iteratorReturnResultRoots(symbol) {
		return true
	}
	if receiver == nil {
		return false
	}
	proven := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(receiver))
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if !lowering.isLibraryType(member, "IteratorYieldResult", "IteratorReturnResult") {
			return false
		}
	}
	return true
}

// iteratorReturnResultRoots reports whether every library root of symbol is declared in
// IteratorReturnResult.
func (lowering *lowering) iteratorReturnResultRoots(symbol *ast.Symbol) bool {
	for _, root := range lowering.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				continue
			}
			owner := declaration.Parent
			if owner == nil ||
				owner.Kind != ast.KindInterfaceDeclaration ||
				owner.Name().Text() != "IteratorReturnResult" {
				return false
			}
		}
	}
	return true
}

func (lowering *lowering) librarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	roots := lowering.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
		}
	}
	return true
}

func (lowering *lowering) prototypeRead(node *ast.Node, name string) error {
	if name == "isPrototypeOf" {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "isPrototypeOf",
			Fix:   "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant",
		}
	}
	return &Refused{
		Where: lowering.program.Where(node),
		What:  "inherited library member " + name + " read as an own field",
		Fix:   "prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)",
	}
}

func (lowering *lowering) objectPrototypeCall(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	return lowering.objectPrototypeCallArguments(node, receiver, name, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) objectPrototypeCallArguments(
	node, receiver *ast.Node,
	name string,
	written []*ast.Node,
) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !lowering.libraryMember(callee) {
		return nil, false, nil
	}
	switch name {
	case "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf":
	default:
		return nil, false, nil
	}
	if name == "isPrototypeOf" {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  "isPrototypeOf",
			Fix:   "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant",
		}
	}
	if name != "valueOf" && lowering.isLibraryType(lowering.checker.GetTypeAtLocation(receiver), errorClasses...) {
		return nil, true, lowering.notYet(
			node,
			name+" on Error (its prototype and non-enumerable own descriptors differ from plain objects)",
		)
	}
	of, _ := lowering.representation(lowering.checker.GetTypeAtLocation(receiver))
	if lowering.exactPlainObject(receiver) {
		of = ir.Object
	}
	if of == ir.Object || of == ir.Union {
		if reason := lowering.prototypeHazard(receiver, name); reason != "" {
			return nil, true, lowering.notYet(node, name+" through an object view ("+reason+")")
		}
	}
	if name == "toString" && of == ir.Number {
		return nil, false, nil
	} // Number's radix overload.
	if name == "toLocaleString" && of == ir.Number {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  name + " on a number",
			Fix:   "JavaScript uses locale-sensitive number formatting; use toString for deterministic formatting",
		}
	}
	if name == "toLocaleString" && of == ir.Array {
		element, elementTypeError := lowering.elementType(receiver)
		if elementTypeError != nil {
			return nil, true, elementTypeError
		}
		if element != ir.String && element != ir.Boolean {
			return nil, true, &Refused{
				Where: lowering.program.Where(node),
				What:  name + " on this array",
				Fix:   "JavaScript formats each element with its own toLocaleString, which may be locale-sensitive or user-defined; use toString for deterministic formatting",
			}
		}
	}
	if name == "hasOwnProperty" || name == "propertyIsEnumerable" {
		if of == ir.Record && name == "hasOwnProperty" {
			return lowering.recordObjectCallArguments(node, "hasOwn", append([]*ast.Node{receiver}, written...))
		}
		if of != ir.Object {
			return lowering.nonObjectOwnPropertyArguments(node, receiver, name, of, written)
		}
		if checker.IsTupleType(lowering.checker.GetTypeAtLocation(receiver)) {
			return nil, true, lowering.notYet(
				node,
				name+" on a tuple (its representation includes absent optional slots and no length descriptor)",
			)
		}
		// Every own field of a plain object or a class instance is enumerable. Methods are on its
		// prototype, absent from its shape, and defineProperty is refused.
		return lowering.hasOwnPropertyArguments(node, receiver, written)
	}
	if len(written) != 0 {
		return nil, true, lowering.notYet(node, name+" with arguments")
	}
	if checker.IsTupleType(lowering.checker.GetTypeAtLocation(receiver)) {
		return nil, true, lowering.notYet(node, name+" on a tuple (its native representation is not an array)")
	}
	if name == "toLocaleString" {
		toString := lowering.checker.GetPropertyOfType(lowering.checker.GetTypeAtLocation(receiver), "toString")
		if toString == nil ||
			len(toString.Declarations) == 0 ||
			!load.IsLibrary(ast.GetSourceFileOfNode(toString.Declarations[0])) {
			return nil, true, lowering.notYet(node, "toLocaleString delegating to a user-defined toString")
		}
	}
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if name == "valueOf" {
		if result, _ := lowering.representation(
			lowering.checker.GetTypeAtLocation(node),
		); result == ir.Union && of == ir.Object {
			return fit(value, result), true, nil
		} else if result != of {
			return nil, true, lowering.notYet(node, "valueOf whose library result type erases the "+
				typeName(of)+
				" representation to Object (keeping or returning that result needs a tagged object view)")
		}
		switch of {
		case ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map, ir.Closure:
			return value, true, nil
		}
	}
	switch of {
	case ir.String:
		return value, true, nil
	case ir.Boolean:
		return ir.BooleanToString{Value: value}, true, nil
	case ir.Array:
		return lowering.arrayMethodArguments(node, receiver, "join", written)
	case ir.Object, ir.Map:
		tag := "[object Object]"
		if of == ir.Map {
			tag = "[object Map]"
			if lowering.isSet(receiver) {
				tag = "[object Set]"
			}
		}
		text := ir.StringConstant{Index: lowering.constant(tag)}
		// Both arms are the same, but the test evaluates the receiver exactly once, preserving its
		// side effects and any narrowed-read check. No field is read to produce the constant tag.
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: text, WhenNot: text}, true, nil
	}
	return nil, true, lowering.notYet(
		node,
		name+" on a "+typeName(of)+" (function source text and mixed value dispatch are not represented)",
	)
}

// Structural object views can hide an override, an Error, or an artificial absent-spread slot.
// Until prototype dispatch and presence descriptors exist, refuse any compatible runtime shape
// in the whole program with one of those hazards. This deliberately errs toward refusal.
// An empty name tests own descriptors without dispatching an overridable receiver method.
func (lowering *lowering) prototypeHazard(receiver *ast.Node, name string) string {
	view := lowering.concrete(lowering.checker.GetTypeAtLocation(receiver))
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return "the program's runtime shapes are not known"
	}
	if name != "valueOf" && lowering.errorMayHide(receiver) {
		return "an Error may be hidden by the view"
	}
	reason := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if reason != "" {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression ||
			node.Kind == ast.KindRegularExpressionLiteral ||
			(node.Kind == ast.KindCallExpression &&
				lowering.isLibraryType(lowering.checker.GetTypeAtLocation(node), "RegExp", "RegExpStringIterator")) ||
			node.Kind == ast.KindNewExpression ||
			node.Kind == ast.KindArrayLiteralExpression ||
			node.Kind == ast.KindArrowFunction ||
			node.Kind == ast.KindNumericLiteral ||
			node.Kind == ast.KindStringLiteral ||
			node.Kind == ast.KindTrueKeyword ||
			node.Kind == ast.KindFalseKeyword {
			shape := lowering.checker.GetTypeAtLocation(node)
			if lowering.checker.IsTypeAssignableTo(shape, view) {
				if lowering.isLibraryType(shape, "RegExp", "RegExpStringIterator") &&
					!lowering.exactPlainObject(receiver) {
					// Regex objects have intrinsic prototype behavior and metadata slots. A structural
					// view cannot make those plain-object methods or own-property descriptors.
					reason = "a RegExp or its iterator may be hidden by the view"
					return true
				}
				if representation, known := lowering.representation(
					shape,
				); known && representation != ir.Object && !lowering.exactPlainObject(receiver) {
					reason = "a value with a different native representation may be hidden by the view"
					return true
				}
				member := lowering.checker.GetPropertyOfType(shape, name)
				if name != "" && member != nil && !lowering.librarySymbol(member) {
					reason = "a user-defined " + name + " may be hidden by the view"
					return true
				}
				if name == "toLocaleString" {
					member = lowering.checker.GetPropertyOfType(shape, "toString")
					if member != nil && !lowering.librarySymbol(member) {
						reason = "a user-defined toString may be hidden by the view"
						return true
					}
				}
				if name == "" || name == "hasOwnProperty" || name == "propertyIsEnumerable" {
					for _, field := range lowering.checker.GetPropertiesOfType(shape) {
						if strings.ContainsRune(field.Name, 0) {
							reason = "a field name contains NUL, which native shape names cannot represent"
							return true
						}
					}
					if node.Kind == ast.KindObjectLiteralExpression {
						for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
							if property.Kind == ast.KindSpreadAssignment && lowering.includesUndefined(
								lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
							) {
								reason = "a possibly absent spread has slots without own-property presence descriptors"
								return true
							}
						}
					}
					if node.Kind == ast.KindNewExpression {
						if declaration := lowering.classes[lowering.symbol(
							ast.SkipParentheses(node.AsNewExpression().Expression),
						)]; declaration != nil {
							for _, member := range declaration.Members() {
								if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient) {
									reason = "declared class fields have no JavaScript own property"
									return true
								}
								if member.Name() != nil && member.Name().Kind == ast.KindPrivateIdentifier {
									reason = "private class slots are not JavaScript own properties"
									return true
								}
							}
						}
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return reason
}

// errorMayHide reports whether receiver may hold an Error. Node keeps an Error's message as an own property
// it doesn't enumerate, and its name on the prototype, where natively both are fields like any other's, so
// no road that walks or tests own properties tells the truth about one yet. An Error made in the source,
// by new Error or any other of errorClasses, may hide when its type fits receiver's. One the runtime made, from a
// failing host call or a brand check, has no such node, so it may hide whenever the library's Error fits receiver's
// type and the program can take one: a catch that binds what it caught, or a Promise's rejection handler.
// prototypeHazard decides its own Error case here, so the two can't drift. This deliberately errs toward refusal.
func (lowering *lowering) errorMayHide(receiver *ast.Node) bool {
	view := lowering.concrete(lowering.checker.GetTypeAtLocation(receiver))
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return true
	}
	made, takes := false, false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if made {
			return true
		}
		if node.Kind == ast.KindNewExpression || node.Kind == ast.KindCallExpression {
			shape := lowering.checker.GetTypeAtLocation(node)
			if lowering.isLibraryType(shape, errorClasses...) && lowering.checker.IsTypeAssignableTo(shape, view) {
				made = true
				return true
			}
		}
		takes = takes || takesRejection(node)
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if made {
		return true
	}
	if !takes {
		return false
	}
	errorType := lowering.libraryErrorType()
	return errorType == nil || lowering.checker.IsTypeAssignableTo(errorType, view)
}

// takesRejection reports whether node hands the program a thrown value it didn't make: a catch that binds
// it, or a call that can pass a Promise's rejection reason to a handler, read by name alone so a receiver
// typed loosely still counts.
func takesRejection(node *ast.Node) bool {
	if node.Kind == ast.KindCatchClause {
		return node.AsCatchClause().VariableDeclaration != nil
	}
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	switch callee.Name().Text() {
	case "catch", "allSettled":
		return true
	case "then":
		return call.Arguments != nil && len(call.Arguments.Nodes) > 1
	}
	return false
}

// libraryErrorType is the library's Error interface, the type a caught Error the runtime made has, or nil
// if the library doesn't declare one.
func (lowering *lowering) libraryErrorType() *checker.Type {
	for _, file := range lowering.program.CompilerProgram().GetSourceFiles() {
		if !load.IsLibrary(file) {
			continue
		}
		for _, statement := range file.Statements.Nodes {
			if statement.Kind == ast.KindInterfaceDeclaration && statement.Name().Text() == "Error" {
				if symbol := lowering.checker.GetSymbolOfDeclaration(statement); symbol != nil {
					return lowering.checker.GetDeclaredTypeOfSymbol(symbol)
				}
			}
		}
	}
	return nil
}

// An unannotated const made directly from a plain literal has no boxed primitive hidden by a
// structural view. Its shape is fixed; spreads and annotated/widened names need the whole scan.
func (lowering *lowering) exactPlainObject(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := lowering.checker.GetSymbolAtLocation(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration ||
			declaration.AsVariableDeclaration().Type != nil ||
			declaration.Parent == nil ||
			declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
		node = ast.SkipParentheses(node)
	}
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind == ast.KindSpreadAssignment {
			return false
		}
	}
	return true
}

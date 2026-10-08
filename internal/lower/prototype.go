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
func (l *lowering) memberSymbol(node *ast.Node) *ast.Symbol {
	symbol := l.checker.GetSymbolAtLocation(node)
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			symbol = l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(access.Expression), key.Text())
		}
	}
	return symbol
}

func (l *lowering) libraryMember(node *ast.Node) bool {
	return l.librarySymbol(l.memberSymbol(node))
}

// A union can combine an own callback with an inherited method. That still cannot be an
// unconditional own-field load: any library root is enough to refuse the read.
func (l *lowering) inheritedLibraryMember(node *ast.Node) bool {
	return l.inheritedLibrarySymbol(l.memberSymbol(node))
}

func (l *lowering) inheritedLibrarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) librarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	roots := l.checker.GetRootSymbols(symbol)
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

func (l *lowering) prototypeRead(node *ast.Node, name string) error {
	if name == "isPrototypeOf" {
		return &Refused{Where: l.program.Where(node), What: "isPrototypeOf", Fix: "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant"}
	}
	return &Refused{Where: l.program.Where(node), What: "inherited library member " + name + " read as an own field", Fix: "prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)"}
}

func (l *lowering) objectPrototypeCall(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !l.libraryMember(callee) {
		return nil, false, nil
	}
	switch name {
	case "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf":
	default:
		return nil, false, nil
	}
	if l.regexGroups(receiver) {
		return nil, true, l.notYet(node, name+" on RegExp named groups (the dictionary has a null prototype, so inherited Object methods throw TypeError on Node)")
	}
	if name == "isPrototypeOf" {
		return nil, true, &Refused{Where: l.program.Where(node), What: "isPrototypeOf", Fix: "Adamic has no observable prototype chain; use instanceof for class identity or an explicit discriminant"}
	}
	if name != "valueOf" && l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Error") {
		return nil, true, l.notYet(node, name+" on Error (its prototype and non-enumerable own descriptors differ from plain objects)")
	}
	of, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	if of == ir.Object {
		if reason := l.prototypeHazard(receiver, name); reason != "" {
			return nil, true, l.notYet(node, name+" through an object view ("+reason+")")
		}
	}
	if name == "toString" && of == ir.Number {
		return nil, false, nil
	} // Number's radix overload.
	if name == "toLocaleString" && of == ir.Number {
		return nil, true, &Refused{Where: l.program.Where(node), What: name + " on a number", Fix: "JavaScript uses locale-sensitive number formatting; use toString for deterministic formatting"}
	}
	if name == "toLocaleString" && of == ir.Array {
		element, err := l.elementType(receiver)
		if err != nil {
			return nil, true, err
		}
		if element != ir.String && element != ir.Boolean {
			return nil, true, &Refused{Where: l.program.Where(node), What: name + " on this array", Fix: "JavaScript formats each element with its own toLocaleString, which may be locale-sensitive or user-defined; use toString for deterministic formatting"}
		}
	}
	if name == "hasOwnProperty" || name == "propertyIsEnumerable" {
		if of != ir.Object {
			return l.nonObjectOwnProperty(node, receiver, name, of)
		}
		if checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {
			return nil, true, l.notYet(node, name+" on a tuple (its representation includes absent optional slots and no length descriptor)")
		}
		// Every own field of a plain object or a class instance is enumerable. Methods are on its
		// prototype, absent from its shape, and defineProperty is refused.
		own, handled, err := l.hasOwnProperty(node, receiver)
		if err == nil && l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "MapIterator", "SetIterator") {
			// Iterator slots are implementation details. Even next is inherited in JavaScript.
			// Evaluate receiver and key once, in order, without exposing any native slots.
			return ir.Conditional{Condition: own, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: ir.BooleanConstant{Value: false}}, true, nil
		}
		return own, handled, err
	}
	if len(node.AsCallExpression().Arguments.Nodes) != 0 {
		return nil, true, l.notYet(node, name+" with arguments")
	}
	if checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, true, l.notYet(node, name+" on a tuple (its native representation is not an array)")
	}
	if name == "toLocaleString" {
		toString := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(receiver), "toString")
		if toString == nil || len(toString.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(toString.Declarations[0])) {
			return nil, true, l.notYet(node, "toLocaleString delegating to a user-defined toString")
		}
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if name == "valueOf" {
		if result, _ := l.representation(l.checker.GetTypeAtLocation(node)); result != of {
			return nil, true, l.notYet(node, "valueOf whose library result type erases the "+typeName(of)+" representation to Object (keeping or returning that result needs a tagged object view)")
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
		return l.arrayMethod(node, receiver, "join")
	case ir.Object, ir.Map:
		tag := "[object Object]"
		if of == ir.Map {
			tag = "[object Map]"
			if l.isSet(receiver) {
				tag = "[object Set]"
			}
		}
		if l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "MapIterator") {
			tag = "[object Map Iterator]"
		} else if l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "SetIterator") {
			tag = "[object Set Iterator]"
		}
		text := ir.StringConstant{Index: l.constant(tag)}
		// Both arms are the same, but the test evaluates the receiver exactly once, preserving its
		// side effects and any narrowed-read check. No field is read to produce the constant tag.
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: text, WhenNot: text}, true, nil
	}
	return nil, true, l.notYet(node, name+" on a "+typeName(of)+" (function source text and mixed value dispatch are not represented)")
}

// Structural object views can hide an override, an Error, or an artificial absent-spread slot.
// Until prototype dispatch and presence descriptors exist, refuse any compatible runtime shape
// in the whole program with one of those hazards. This deliberately errs toward refusal.
func (l *lowering) prototypeHazard(receiver *ast.Node, name string) string {
	view := l.concrete(l.checker.GetTypeAtLocation(receiver))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return "the program's runtime shapes are not known"
	}
	reason := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if reason != "" {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindRegularExpressionLiteral || (node.Kind == ast.KindCallExpression && l.isLibraryType(l.checker.GetTypeAtLocation(node), "RegExp", "RegExpStringIterator", "MapIterator", "SetIterator")) || node.Kind == ast.KindNewExpression || node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindArrowFunction || node.Kind == ast.KindNumericLiteral || node.Kind == ast.KindStringLiteral || node.Kind == ast.KindTrueKeyword || node.Kind == ast.KindFalseKeyword {
			shape := l.checker.GetTypeAtLocation(node)
			if l.checker.IsTypeAssignableTo(shape, view) {
				if l.isLibraryType(shape, "MapIterator", "SetIterator") && !l.isLibraryType(view, "MapIterator", "SetIterator") && !l.exactPlainObject(receiver) {
					reason = "a collection iterator may be hidden by the view"
					return true
				}
				if l.isLibraryType(shape, "RegExp", "RegExpStringIterator") && !l.exactPlainObject(receiver) {
					// Regex objects have intrinsic prototype behavior and metadata slots. A structural
					// view cannot make those plain-object methods or own-property descriptors.
					reason = "a RegExp or its iterator may be hidden by the view"
					return true
				}
				if representation, known := l.representation(shape); known && representation != ir.Object && !l.exactPlainObject(receiver) {
					reason = "a value with a different native representation may be hidden by the view"
					return true
				}
				if name != "valueOf" && l.isLibraryType(shape, "Error") {
					reason = "an Error may be hidden by the view"
					return true
				}
				member := l.checker.GetPropertyOfType(shape, name)
				if member != nil && !l.librarySymbol(member) {
					reason = "a user-defined " + name + " may be hidden by the view"
					return true
				}
				if name == "toLocaleString" {
					member = l.checker.GetPropertyOfType(shape, "toString")
					if member != nil && !l.librarySymbol(member) {
						reason = "a user-defined toString may be hidden by the view"
						return true
					}
				}
				if name == "hasOwnProperty" || name == "propertyIsEnumerable" {
					for _, field := range l.checker.GetPropertiesOfType(shape) {
						if strings.HasPrefix(field.Name, "#") {
							reason = "a public # name collides with private native slot names"
							return true
						}
						if strings.ContainsRune(field.Name, 0) {
							reason = "a field name contains NUL, which native shape names cannot represent"
							return true
						}
					}
					if node.Kind == ast.KindObjectLiteralExpression {
						for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
							if property.Kind == ast.KindSpreadAssignment && l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
								reason = "a possibly absent spread has slots without own-property presence descriptors"
								return true
							}
						}
					}
					if node.Kind == ast.KindNewExpression {
						if declaration := l.classes[l.symbol(ast.SkipParentheses(node.AsNewExpression().Expression))]; declaration != nil {
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

// An unannotated const made directly from a plain literal has no boxed primitive hidden by a
// structural view. Its shape is fixed; spreads and annotated/widened names need the whole scan.
func (l *lowering) exactPlainObject(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.checker.GetSymbolAtLocation(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Type != nil || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
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

package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognize the library declaration, never an unrelated variable named Object.
// This does not materialize Object.prototype or permit detached method values.
func (l *lowering) objectIntrinsic(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	access := node.AsPropertyAccessExpression()
	if l.isLibraryGlobal(access.Expression, "Object") {
		if _, known := objectStaticMethodLengths[access.Name().Text()]; known {
			return access.Name().Text()
		}
		return ""
	}
	prototype := ast.SkipParentheses(access.Expression)
	if prototype.Kind != ast.KindPropertyAccessExpression || prototype.Name().Text() != "prototype" || !l.isLibraryGlobal(prototype.AsPropertyAccessExpression().Expression, "Object") {
		return ""
	}
	switch access.Name().Text() {
	case "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf":
		return access.Name().Text()
	}
	return ""
}

func (l *lowering) libraryObjectBoundMethod(node *ast.Node) bool {
	if l.objectIntrinsic(node) == "" {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindCallExpression {
		call := parent.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "hasOwn" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") && len(call.Arguments.Nodes) == 2 && ast.SkipParentheses(call.Arguments.Nodes[0]) == ast.SkipParentheses(node)
	}
	if parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	switch parent.Name().Text() {
	case "name", "length", "prototype":
		return true
	case "call", "hasOwnProperty", "propertyIsEnumerable":
		return called(parent)
	}
	return false
}

func (l *lowering) libraryObjectIntrinsicCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" {
		return nil, false, nil
	}
	name := l.objectIntrinsic(callee.AsPropertyAccessExpression().Expression)
	if name == "" {
		return nil, false, nil
	}
	if name != "toString" && name != "hasOwnProperty" && name != "propertyIsEnumerable" {
		return nil, true, l.notYet(node, "Object intrinsic "+name+".call")
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	count := 1
	if name != "toString" {
		count = 2
	}
	if len(arguments) != count || hasSpread(node) {
		return nil, true, l.notYet(node, "Object.prototype."+name+".call with these arguments")
	}
	receiver := arguments[0]
	proven := l.checker.GetTypeAtLocation(receiver)
	of, known := l.representation(proven)
	tag := ""
	if name == "toString" {
		if l.libraryIteratorTagType(proven) {
			value, err := l.expression(receiver)
			if err != nil {
				return nil, true, err
			}
			return l.libraryIteratorString(value), true, nil
		}
		switch {
		case proven.Flags()&checker.TypeFlagsNull != 0:
			tag = "Null"
		case proven.Flags()&checker.TypeFlagsUndefined != 0:
			tag = "Undefined"
		case proven.Flags()&checker.TypeFlagsNumberLike != 0:
			tag = "Number"
		case proven.Flags()&checker.TypeFlagsStringLike != 0:
			tag = "String"
		case proven.Flags()&checker.TypeFlagsBooleanLike != 0:
			tag = "Boolean"
		case known && of == ir.Array:
			tag = "Array"
		case known && of == ir.Closure:
			tag = "Function"
		case l.exactObject(receiver, 0):
			tag = "Object"
		}
		if tag == "" {
			return nil, true, l.notYet(receiver, "Object.prototype.toString.call without a proven intrinsic tag (boxed primitives, host objects and Symbol.toStringTag are not represented)")
		}
		value, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		text := ir.StringConstant{Index: l.constant("[object " + tag + "]")}
		// A generated parameter preserves evaluation once, without testing a number as a
		// pointer or comparing an immortal string's address with NULL.
		function, _ := l.stringHelper("object_intrinsic_tag", []ir.Expression{value})
		l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: text}}
		return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, true, nil
	}
	if !known || l.includesNull(proven) || l.includesUndefined(proven) {
		return nil, true, l.notYet(receiver, "Object prototype own-property call with a nullish or mixed receiver")
	}
	if of == ir.Array && l.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray") {
		return nil, true, l.notYet(receiver, "Object prototype own-property call on a regex result array (additional own descriptors are not represented)")
	}
	if of != ir.Object {
		return l.nonObjectOwnPropertyArguments(node, receiver, name, of, arguments[1:])
	}
	if !l.exactObject(receiver, 0) {
		return nil, true, l.notYet(receiver, "Object prototype own-property call without a proven complete plain shape")
	}
	if key, _ := l.representation(l.checker.GetTypeAtLocation(arguments[1])); key != ir.String {
		return nil, true, l.notYet(arguments[1], "Object prototype property key requiring ToPropertyKey coercion")
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	return ir.HasOwn{Object: object, Key: key}, true, nil
}

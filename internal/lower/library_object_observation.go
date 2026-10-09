package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Node 24.19.0's intrinsic data descriptors. Every listed property is non-enumerable.
// A static observation does not make a prototype object or callable method value escape.
var objectConstructorNames = []string{"length", "name", "prototype", "assign", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "getOwnPropertyNames", "getOwnPropertySymbols", "hasOwn", "is", "preventExtensions", "seal", "create", "defineProperties", "defineProperty", "freeze", "getPrototypeOf", "setPrototypeOf", "isExtensible", "isFrozen", "isSealed", "keys", "entries", "fromEntries", "values", "groupBy"}
var objectPrototypeNames = []string{"constructor", "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__", "isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString"}

var objectStaticMethodLengths = map[string]float64{
	"assign": 2, "getOwnPropertyDescriptor": 2, "getOwnPropertyDescriptors": 1,
	"getOwnPropertyNames": 1, "getOwnPropertySymbols": 1, "hasOwn": 2, "is": 2,
	"preventExtensions": 1, "seal": 1, "create": 2, "defineProperties": 2,
	"defineProperty": 3, "freeze": 1, "getPrototypeOf": 1, "setPrototypeOf": 2,
	"isExtensible": 1, "isFrozen": 1, "isSealed": 1, "keys": 1, "entries": 1,
	"fromEntries": 1, "values": 1, "groupBy": 2,
}

func (l *lowering) objectPrototype(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Object")
}

func (l *lowering) libraryObjectObservation(node *ast.Node) (ir.Expression, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindTypeOfExpression && l.objectIntrinsic(node.AsTypeOfExpression().Expression) != "" {
		return ir.StringConstant{Index: l.constant("function")}, true
	}
	if name := l.objectIntrinsic(node); name != "" && truthinessUse(node) {
		return ir.BooleanConstant{Value: true}, true
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return nil, false
	}
	name := l.objectIntrinsic(node.AsPropertyAccessExpression().Expression)
	if name == "" {
		return nil, false
	}
	switch node.Name().Text() {
	case "name":
		return ir.StringConstant{Index: l.constant(name)}, true
	case "length":
		length := 0.0
		if name == "hasOwnProperty" || name == "propertyIsEnumerable" || name == "isPrototypeOf" {
			length = 1
		}
		method := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
		if l.isLibraryGlobal(method.AsPropertyAccessExpression().Expression, "Object") {
			length = objectStaticMethodLengths[name]
		}
		return ir.NumberConstant{Value: length}, true
	case "prototype":
		return ir.Undefined{}, true
	}
	return nil, false
}

func (l *lowering) libraryObjectStaticCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	receiver, method := access.Expression, access.Name().Text()
	names := []string(nil)
	if l.objectPrototype(receiver) {
		names = objectPrototypeNames
	}
	if l.isLibraryGlobal(receiver, "Object") {
		names = objectConstructorNames
	}
	if l.objectIntrinsic(receiver) != "" {
		names = []string{"length", "name"}
	}
	if names == nil {
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if method == "hasOwn" && l.isLibraryGlobal(receiver, "Object") && len(arguments) == 2 && !hasSpread(node) {
		target := arguments[0]
		switch {
		case l.isLibraryGlobal(target, "Object"):
			names = objectConstructorNames
		case l.objectPrototype(target):
			names = objectPrototypeNames
		case l.objectIntrinsic(target) != "":
			names = []string{"length", "name"}
		default:
			return nil, false, nil
		}
		arguments = arguments[1:]
		method = "hasOwnProperty"
	}
	if method == "toString" && l.objectPrototype(receiver) && len(arguments) == 0 {
		return ir.StringConstant{Index: l.constant("[object Object]")}, true, nil
	}
	if method != "hasOwnProperty" && method != "propertyIsEnumerable" {
		return nil, false, nil
	}
	if len(arguments) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "static Object own-property observation with these arguments")
	}
	key, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(arguments[0], "static Object property key requiring ToPropertyKey")
	}
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "key", Type: ir.String, Function: function})
	var result ir.Expression = ir.BooleanConstant{Value: false}
	if method == "hasOwnProperty" {
		for _, name := range names {
			result = ir.Binary{Operator: ir.Or, Left: result, Right: ir.Binary{Operator: ir.Equal, Left: ir.Read{Local: local, Of: ir.String}, Right: ir.StringConstant{Index: l.constant(name)}}}
		}
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "object_static_own", Parameters: []int{local}, Returns: ir.Boolean, Body: []ir.Statement{ir.Return{Value: result}}})
	return ir.Call{Function: function, Arguments: []ir.Expression{key}, Returns: ir.Boolean}, true, nil
}

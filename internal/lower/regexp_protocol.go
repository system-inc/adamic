package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognize well-known keys by library Symbol identity, never by spelling alone.
func (l *lowering) regexProtocolKey(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || !l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Symbol") {
		return "", false
	}
	name := node.Name().Text()
	switch name {
	case "match", "matchAll", "replace", "search", "split":
		return name, true
	}
	return "", false
}

func (l *lowering) regexProtocol(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	args := node.AsCallExpression().Arguments.Nodes
	prototypeCall := false
	if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "call" {
		callee = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		prototypeCall = true
	}
	if callee.Kind != ast.KindElementAccessExpression {
		return nil, false, nil
	}
	access := callee.AsElementAccessExpression()
	name, known := l.regexProtocolKey(access.ArgumentExpression)
	if !known {
		return nil, false, nil
	}
	receiver := access.Expression
	if prototypeCall {
		base := ast.SkipParentheses(receiver)
		if base.Kind != ast.KindPropertyAccessExpression || base.Name().Text() != "prototype" || !l.isLibraryGlobal(base.AsPropertyAccessExpression().Expression, "RegExp") {
			return nil, false, nil
		}
		if len(args) == 0 {
			return nil, true, l.notYet(node, "RegExp Symbol protocol without a proven receiver")
		}
		receiver, args = args[0], args[1:]
	}
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	if !l.isLibraryType(proven, "RegExp") {
		return nil, true, l.notYet(node, "RegExp Symbol protocol on an unproven receiver or a custom hook")
	}
	if access.QuestionDotToken != nil {
		return nil, true, l.notYet(node, "an optional RegExp Symbol protocol call")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{}
	for _, arg := range args {
		lowered, err := l.expression(arg)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	if len(arguments) == 0 || arguments[0].Type() != ir.String {
		return nil, true, l.notYet(node, "RegExp Symbol protocol input other than a string")
	}
	result := ir.Array
	switch name {
	case "search":
		result = ir.Number
	case "matchAll":
		result = ir.Object
	case "replace":
		result = ir.String
		if len(arguments) != 2 {
			return nil, true, l.notYet(node, "RegExp Symbol replacement arity")
		}
		if arguments[1].Type() == ir.Closure && l.regexReplacementCallback(args[1]) {
			name = "replaceCallback"
		} else if arguments[1].Type() != ir.String {
			return nil, true, l.notYet(node, "RegExp Symbol replacement callback with unproved capture/index/input/group parameters or result type")
		}
	case "split":
		if len(arguments) == 1 {
			arguments = append(arguments, ir.NumberConstant{Value: 4294967295})
		}
		if len(arguments) == 2 {
			if _, undefined := arguments[1].(ir.Undefined); undefined {
				arguments[1] = ir.NumberConstant{Value: 4294967295}
			}
			if arguments[1].Type() == ir.MaybeNumber {
				arguments[1] = ir.Coalesce{Value: arguments[1], Fallback: ir.NumberConstant{Value: 4294967295}, Of: ir.Number}
			}
		}
		if len(arguments) != 2 || arguments[1].Type() != ir.Number {
			return nil, true, l.notYet(node, "RegExp Symbol split limit other than a number")
		}
	}
	if name != "replace" && name != "replaceCallback" && name != "split" && len(arguments) != 1 {
		return nil, true, l.notYet(node, "RegExp Symbol protocol with extra arguments")
	}
	return ir.RegExpCall{Value: value, Arguments: arguments, Method: "symbol:" + name, Returns: result}, true, nil
}

// A string concatenation requests the default primitive conversion. Only the
// library RegExp representation is admitted; structural custom hooks are refused.
func (l *lowering) regexStringValue(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() == ir.Object && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)), "RegExp") {
		return ir.RegExpCall{Value: value, Method: "toString", Returns: ir.String}
	}
	return value
}

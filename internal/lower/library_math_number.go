package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// These are the exact binary64 constants V8 exposes, not computations with libm.
var libraryMathConstants = map[string]float64{
	"LN10": 0x1.26bb1bbb55516p+1, "LN2": 0x1.62e42fefa39efp-1,
	"LOG10E": 0x1.bcb7b1526e50ep-2, "LOG2E": 0x1.71547652b82fep+0,
	"SQRT1_2": 0x1.6a09e667f3bcdp-1, "SQRT2": 0x1.6a09e667f3bcdp+0,
}

func (l *lowering) libraryMathNumberProperty(node *ast.Node) (ir.Expression, bool) {
	access := node.AsPropertyAccessExpression()
	if l.isLibraryGlobal(access.Expression, "Math") {
		if value, known := libraryMathConstants[node.Name().Text()]; known {
			return ir.NumberConstant{Value: value}, true
		}
	}
	if l.isLibraryGlobal(access.Expression, "Number") {
		switch node.Name().Text() {
		case "length":
			return ir.NumberConstant{Value: 1}, true
		case "name":
			return ir.StringConstant{Index: l.constant("Number")}, true
		}
	}
	return nil, false
}

func (l *lowering) numberPrototype(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Number")
}

// libraryNumber converts only proven primitives. Objects can run arbitrary valueOf/toString code,
// so they stay NotYet rather than being guessed at. The runtime handles missing primitive values.
func (l *lowering) libraryNumber(node *ast.Node) (ir.Expression, error) {
	if ast.SkipParentheses(node).Kind == ast.KindNullKeyword {
		return ir.NumberConstant{}, nil
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if _, missing := value.(ir.Undefined); missing {
		return ir.NumberConstant{Value: math.NaN()}, nil
	}
	switch value.Type() {
	case ir.Number:
		return value, nil
	case ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
	case ir.Union:
		if !l.writable(l.checker.GetTypeAtLocation(node)) {
			return nil, l.notYet(node, "Number conversion of a union containing objects")
		}
	default:
		return nil, l.notYet(node, "Number conversion of an object")
	}
	return ir.NumberCall{Function: "convert", Arguments: []ir.Expression{value}}, nil
}

func (l *lowering) libraryMathNumberCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	written := node.AsCallExpression().Arguments.Nodes
	if l.isLibraryGlobal(callee, "isNaN") || l.isLibraryGlobal(callee, "isFinite") {
		name := callee.Text()
		if len(written) != 1 || hasSpread(node) {
			return nil, true, l.notYet(node, name+" with spread or other than one argument")
		}
		value, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.Number {
			return nil, true, l.notYet(node, name+" without a proven number argument")
		}
		return ir.NumberCall{Function: name, Arguments: []ir.Expression{value}}, true, nil
	}
	if l.isLibraryGlobal(callee, "Number") {
		if len(written) == 0 {
			return ir.NumberConstant{}, true, nil
		}
		if len(written) != 1 || hasSpread(node) {
			return nil, true, l.notYet(node, "Number with spread or extra arguments")
		}
		value, err := l.libraryNumber(written[0])
		return value, true, err
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	if l.isLibraryGlobal(receiver, "Math") && (name == "clz32" || name == "fround" || name == "imul") {
		count := 1
		if name == "imul" {
			count = 2
		}
		if len(written) > count || hasSpread(node) {
			return nil, true, l.notYet(node, "Math."+name+" with spread or extra arguments")
		}
		arguments := []ir.Expression{}
		for _, argument := range written {
			value, err := l.libraryNumber(argument)
			if err != nil {
				return nil, true, err
			}
			arguments = append(arguments, value)
		}
		for len(arguments) < count {
			arguments = append(arguments, ir.NumberConstant{Value: math.NaN()})
		}
		return ir.MathCall{Function: name, Arguments: arguments}, true, nil
	}
	if (l.isLibraryGlobal(receiver, "Number") || l.numberPrototype(receiver)) && name == "hasOwnProperty" {
		if len(written) != 1 {
			return nil, true, l.notYet(node, "Number.hasOwnProperty with other than one key")
		}
		key, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if key.Type() != ir.String {
			return nil, true, l.notYet(node, "Number.hasOwnProperty with a non-string key")
		}
		function := "hasOwnProperty"
		if l.numberPrototype(receiver) {
			function = "prototypeHasOwnProperty"
		}
		return ir.NumberCall{Function: function, Arguments: []ir.Expression{key}}, true, nil
	}
	// Number.prototype carries [[NumberData]] +0. Its methods may also be called with a primitive
	// number as this; no prototype object is represented as a plain object with invented fields.
	var value ir.Expression
	if l.numberPrototype(receiver) {
		value = ir.NumberConstant{}
	} else if name == "call" && ast.SkipParentheses(receiver).Kind == ast.KindPropertyAccessExpression {
		method := ast.SkipParentheses(receiver)
		if !l.numberPrototype(method.AsPropertyAccessExpression().Expression) {
			return nil, false, nil
		}
		name = method.Name().Text()
		if len(written) == 0 {
			return nil, true, l.notYet(node, "Number prototype method without a numeric receiver")
		}
		var err error
		if l.numberPrototype(written[0]) {
			value = ir.NumberConstant{}
		} else {
			value, err = l.expression(written[0])
			if err != nil {
				return nil, true, err
			}
			if value.Type() != ir.Number {
				return nil, true, l.notYet(node, "Number prototype method on a non-number receiver")
			}
		}
		written = written[1:]
	} else {
		return nil, false, nil
	}
	if name == "valueOf" && len(written) == 0 {
		return value, true, nil
	}
	if name != "toString" && name != "toFixed" && name != "toExponential" && name != "toPrecision" {
		return nil, true, l.notYet(node, "Number.prototype."+name)
	}
	if len(written) > 1 {
		return nil, true, l.notYet(node, "Number prototype format with extra arguments")
	}
	var argument ir.Expression
	if len(written) == 1 {
		var err error
		argument, err = l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if _, missing := argument.(ir.Undefined); missing {
			argument = nil
		} else if argument.Type() != ir.Number {
			return nil, true, l.notYet(node, "Number prototype format with a non-number argument")
		}
	}
	if name == "toFixed" {
		if argument == nil {
			argument = ir.NumberConstant{}
		}
		return ir.ToFixed{Value: value, Digits: argument}, true, nil
	}
	return ir.NumberFormat{Method: name, Value: value, Argument: argument}, true, nil
}

// A .call with an explicit this is bound, even though its method expression isn't the callee.
func (l *lowering) libraryNumberBoundMethod(node *ast.Node) bool {
	if !l.numberPrototype(node.AsPropertyAccessExpression().Expression) {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && called(parent)
}

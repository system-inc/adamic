package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
	if name, arity, intrinsic := l.libraryMathNumberFunction(access.Expression); intrinsic {
		switch node.Name().Text() {
		case "length":
			return ir.NumberConstant{Value: float64(arity)}, true
		case "name":
			return ir.StringConstant{Index: l.constant(name)}, true
		}
	}
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

// An immediate Number construction is used only for its intrinsic internal slot. No object
// identity escapes, no prototype can be changed, and its argument is evaluated exactly once.
func (l *lowering) libraryNumberConstruction(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindNewExpression && l.isLibraryGlobal(node.AsNewExpression().Expression, "Number")
}

func (l *lowering) libraryNumberSlot(node *ast.Node) (ir.Expression, error) {
	if l.numberPrototype(node) {
		return ir.NumberConstant{}, nil
	}
	node = ast.SkipParentheses(node)
	arguments := node.AsNewExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return ir.NumberConstant{}, nil
	}
	if len(arguments.Nodes) != 1 || arguments.Nodes[0].Kind == ast.KindSpreadElement {
		return nil, l.notYet(node, "immediate Number construction with spread or extra arguments")
	}
	return l.libraryNumber(arguments.Nodes[0])
}

// libraryNumber converts only proven primitives. Objects can run arbitrary valueOf/toString code,
// so they stay NotYet rather than being guessed at. The runtime handles missing primitive values.
func (l *lowering) libraryNumber(node *ast.Node) (ir.Expression, error) {
	if l.numberPrototype(node) || l.libraryNumberConstruction(node) {
		return l.libraryNumberSlot(node)
	}
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
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if proven.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		constant := ir.NumberConstant{Value: math.NaN()}
		if proven.Flags()&checker.TypeFlagsNull != 0 {
			constant.Value = 0
		}
		return l.nullableObservation("number_empty", value, constant, func(ir.Expression) ir.Expression { return constant }), nil
	}
	if value.Type() == ir.String && l.includesNull(l.checker.GetTypeAtLocation(node)) {
		return l.nullableObservation("number", value, ir.NumberConstant{}, func(read ir.Expression) ir.Expression {
			return ir.NumberCall{Function: "convert", Arguments: []ir.Expression{read}}
		}), nil
	}
	// Date is a proven intrinsic: its numeric valueOf reads the internal slot.
	// Generic objects still require user-defined coercion and remain NotYet.
	if l.isLibraryType(l.checker.GetTypeAtLocation(node), "Date") {
		return ir.DateCall{Method: "valueOf", Receiver: value, Returns: ir.Number}, nil
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
	if l.isLibraryGlobal(callee, "isFinite") || l.isLibraryGlobal(callee, "isNaN") {
		if len(written) > 1 || hasSpread(node) {
			return nil, true, l.notYet(node, "global numeric predicate with spread or extra arguments")
		}
		value := ir.Expression(ir.NumberConstant{Value: math.NaN()})
		if len(written) == 1 {
			var err error
			value, err = l.libraryNumber(written[0])
			if err != nil {
				return nil, true, err
			}
		}
		return ir.NumberCall{Function: callee.Text(), Arguments: []ir.Expression{value}}, true, nil
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
	if name == "hasOwnProperty" {
		if _, _, intrinsic := l.libraryMathNumberFunction(receiver); intrinsic {
			if len(written) != 1 || hasSpread(node) {
				return nil, true, l.notYet(node, "numeric intrinsic hasOwnProperty with spread or other than one key")
			}
			key, err := l.stringConversion(written[0])
			if err != nil {
				return nil, true, err
			}
			return l.libraryNumberFunctionOwnProperty(key), true, nil
		}
	}
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
	if l.numberPrototype(receiver) || l.libraryNumberConstruction(receiver) {
		var err error
		value, err = l.libraryNumberSlot(receiver)
		if err != nil {
			return nil, true, err
		}
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
		if l.numberPrototype(written[0]) || l.libraryNumberConstruction(written[0]) {
			value, err = l.libraryNumberSlot(written[0])
			if err != nil {
				return nil, true, err
			}
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
	} else if of, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); of == ir.Number && l.libraryMember(callee) {
		switch name {
		case "toString", "toFixed", "toExponential", "toPrecision", "valueOf":
		default:
			return nil, false, nil
		}
		var err error
		value, err = l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
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

// Intrinsic method observation reads metadata without detaching a receiver. The same named
// member of a user object is never specialized, and taking an intrinsic as a value stays refused.
func (l *lowering) libraryMathNumberFunction(node *ast.Node) (string, int, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return "", 0, false
	}
	receiver, name := node.AsPropertyAccessExpression().Expression, node.Name().Text()
	if !l.libraryMember(node) || len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) == 0 {
		return "", 0, false
	}
	if l.isLibraryGlobal(receiver, "Math") {
		arity, known := mathFunctions[name]
		if name == "clz32" || name == "fround" {
			arity, known = 1, true
		}
		if name == "imul" {
			arity, known = 2, true
		}
		if name == "random" {
			arity, known = 0, true
		}
		if arity < 0 {
			arity = 2
		}
		return name, arity, known
	}
	if l.numberPrototype(receiver) {
		switch name {
		case "valueOf", "toLocaleString":
			return name, 0, true
		case "toString", "toFixed", "toExponential", "toPrecision":
			return name, 1, true
		}
	}
	if l.isLibraryGlobal(receiver, "Number") {
		switch name {
		case "isFinite", "isNaN", "isInteger", "isSafeInteger", "parseFloat":
			return name, 1, true
		case "parseInt":
			return name, 2, true
		}
	}
	return "", 0, false
}

func (l *lowering) libraryMathNumberTypeOf(node *ast.Node) (ir.Expression, bool) {
	operand := ast.SkipParentheses(node.AsTypeOfExpression().Expression)
	text := ""
	if l.isLibraryGlobal(operand, "Math") || l.numberPrototype(operand) {
		text = "object"
	}
	if _, _, intrinsic := l.libraryMathNumberFunction(operand); intrinsic {
		text = "function"
	}
	if text == "" {
		return nil, false
	}
	return ir.StringConstant{Index: l.constant(text)}, true
}

// A .call with an explicit this is bound. typeof and name/length observations do not call it.
func (l *lowering) libraryNumberBoundMethod(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if _, _, intrinsic := l.libraryMathNumberFunction(node); intrinsic {
		if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "hasOwnProperty" && called(parent) {
			return true
		}
		if parent != nil && parent.Kind == ast.KindTypeOfExpression {
			return true
		}
		if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && (parent.Name().Text() == "length" || parent.Name().Text() == "name") {
			return true
		}
	}
	if !l.numberPrototype(node.AsPropertyAccessExpression().Expression) {
		return false
	}
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && called(parent)
}

// Numeric builtin functions own just their name and length. The helper evaluates and converts
// the key once; ordinary IR keeps both backends and the ownership analyses on the same path.
func (l *lowering) libraryNumberFunctionOwnProperty(key ir.Expression) ir.Expression {
	function := len(l.result.Functions)
	parameter := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "key", Type: ir.String, Function: function})
	read := ir.Read{Local: parameter, Of: ir.String}
	result := ir.Binary{Operator: ir.Or,
		Left:  ir.Binary{Operator: ir.Equal, Left: read, Right: ir.StringConstant{Index: l.constant("length")}},
		Right: ir.Binary{Operator: ir.Equal, Left: read, Right: ir.StringConstant{Index: l.constant("name")}},
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_number_function_own_property", Parameters: []int{parameter}, Returns: ir.Boolean, Body: []ir.Statement{ir.Return{Value: result}}})
	return ir.Call{Function: function, Arguments: []ir.Expression{key}, Returns: ir.Boolean}
}

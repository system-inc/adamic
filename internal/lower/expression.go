package lower

import (
	"errors"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// typeOf is what's left at runtime of the type the checker proved for a node: a number, a boolean or
// a string. A union counts when every member is the same one ('Fizz' | 'Buzz' is a string).
func (l *lowering) typeOf(node *ast.Node) (ir.Type, error) {
	if valueType, isKnown := representation(l.checker.GetTypeAtLocation(node)); isKnown {
		return valueType, nil
	}
	return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
}

func representation(proven *checker.Type) (ir.Type, bool) {
	flags := proven.Flags()
	switch {
	case flags&checker.TypeFlagsNumberLike != 0:
		return ir.Number, true
	case flags&checker.TypeFlagsStringLike != 0:
		return ir.String, true
	case flags&checker.TypeFlagsBooleanLike != 0:
		return ir.Boolean, true
	case flags&checker.TypeFlagsUnion != 0:
		var shared ir.Type
		for _, member := range proven.Types() {
			memberType, isKnown := representation(member)
			if !isKnown || (shared != 0 && memberType != shared) {
				return 0, false
			}
			shared = memberType
		}
		return shared, shared != 0
	}
	return 0, false
}

// expression lowers a value.
func (l *lowering) expression(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindNumericLiteral:
		return l.numericLiteral(node)
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return ir.StringConstant{Index: l.constant(node.Text())}, nil
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return ir.BooleanConstant{Value: node.Kind == ast.KindTrueKeyword}, nil
	case ast.KindIdentifier:
		local, isLocal := l.local(node)
		if !isLocal {
			return nil, l.notYet(node, "reading "+node.Text())
		}
		return ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)}, nil
	case ast.KindPrefixUnaryExpression:
		return l.prefix(node)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		left, err := l.expression(binary.Left)
		if err != nil {
			return nil, err
		}
		right, err := l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		return l.combine(node, binary.OperatorToken.Kind, left, right)
	case ast.KindTemplateExpression:
		return l.template(node)
	case ast.KindConditionalExpression:
		return l.conditional(node)
	case ast.KindCallExpression:
		call, err := l.call(node)
		if err != nil {
			return nil, err
		}
		if call.Type() == 0 {
			// The checker allows a void call where a value goes only in places 0.1 refuses anyway
			// (a template of void prints "undefined"); stage 0 says so rather than guess.
			return nil, l.notYet(node, "a void call used as a value")
		}
		return call, nil
	}
	return nil, l.notYet(node, describe(node))
}

// numericLiteral is the value the checker read from the literal, so 0x1F, 1_000 and 1e3 all mean
// what JavaScript says they mean without a second parser here.
func (l *lowering) numericLiteral(node *ast.Node) (ir.Expression, error) {
	literal := l.checker.GetTypeAtLocation(node)
	if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, errors.New("lower: " + l.program.Where(node) + ": the checker gave a numeric literal a type that isn't a number literal")
	}
	value := reflect.ValueOf(literal.AsLiteralType().Value())
	if value.Kind() != reflect.Float64 {
		return nil, errors.New("lower: " + l.program.Where(node) + ": a numeric literal's value isn't a float64")
	}
	return ir.NumberConstant{Value: value.Float()}, nil
}

func (l *lowering) prefix(node *ast.Node) (ir.Expression, error) {
	prefix := node.AsPrefixUnaryExpression()
	operand, err := l.expression(prefix.Operand)
	if err != nil {
		return nil, err
	}
	switch {
	case prefix.Operator == ast.KindMinusToken && operand.Type() == ir.Number:
		return ir.Unary{Operator: ir.Negate, Operand: operand}, nil
	case prefix.Operator == ast.KindPlusToken && operand.Type() == ir.Number:
		return ir.Unary{Operator: ir.Plus, Operand: operand}, nil
	case prefix.Operator == ast.KindExclamationToken && operand.Type() == ir.Boolean:
		return ir.Unary{Operator: ir.Not, Operand: operand}, nil
	}
	return nil, l.notYet(node, describe(node)+" on a "+typeName(operand.Type()))
}

var arithmetic = map[ast.Kind]ir.Operator{
	ast.KindPlusToken:             ir.Add,
	ast.KindMinusToken:            ir.Subtract,
	ast.KindAsteriskToken:         ir.Multiply,
	ast.KindSlashToken:            ir.Divide,
	ast.KindPercentToken:          ir.Remainder,
	ast.KindAsteriskAsteriskToken: ir.Power,
}

var comparisons = map[ast.Kind]ir.Operator{
	ast.KindLessThanToken:          ir.Less,
	ast.KindLessThanEqualsToken:    ir.LessOrEqual,
	ast.KindGreaterThanToken:       ir.Greater,
	ast.KindGreaterThanEqualsToken: ir.GreaterOrEqual,
}

// combine lowers a binary operator on two lowered operands.
func (l *lowering) combine(node *ast.Node, operator ast.Kind, left ir.Expression, right ir.Expression) (ir.Expression, error) {
	both := func(want ir.Type) bool { return left.Type() == want && right.Type() == want }
	if operator == ast.KindPlusToken && both(ir.String) {
		return ir.Concat{Parts: []ir.Expression{left, right}}, nil
	}
	if lowered, isArithmetic := arithmetic[operator]; isArithmetic && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if lowered, isComparison := comparisons[operator]; isComparison && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) && left.Type() == right.Type() {
		lowered := ir.Equal
		if operator == ast.KindExclamationEqualsEqualsToken {
			lowered = ir.NotEqual
		}
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if (operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken) && both(ir.Boolean) {
		lowered := ir.And
		if operator == ast.KindBarBarToken {
			lowered = ir.Or
		}
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	return nil, l.notYet(node, describe(node)+" with a "+typeName(left.Type())+" and a "+typeName(right.Type()))
}

// template lowers a template literal to a Concat, writing each value as String() would.
func (l *lowering) template(node *ast.Node) (ir.Expression, error) {
	template := node.AsTemplateExpression()
	parts := []ir.Expression{}
	if head := template.Head.Text(); head != "" {
		parts = append(parts, ir.StringConstant{Index: l.constant(head)})
	}
	for _, span := range template.TemplateSpans.Nodes {
		value, err := l.expression(span.AsTemplateSpan().Expression)
		if err != nil {
			return nil, err
		}
		switch value.Type() {
		case ir.Number:
			value = ir.NumberToString{Value: value}
		case ir.Boolean:
			value = ir.BooleanToString{Value: value}
		}
		parts = append(parts, value)
		if literal := span.AsTemplateSpan().Literal.Text(); literal != "" {
			parts = append(parts, ir.StringConstant{Index: l.constant(literal)})
		}
	}
	return ir.Concat{Parts: parts}, nil
}

func (l *lowering) conditional(node *ast.Node) (ir.Expression, error) {
	conditional := node.AsConditionalExpression()
	condition, err := l.condition(conditional.Condition)
	if err != nil {
		return nil, err
	}
	whenTrue, err := l.expression(conditional.WhenTrue)
	if err != nil {
		return nil, err
	}
	whenNot, err := l.expression(conditional.WhenFalse)
	if err != nil {
		return nil, err
	}
	if whenTrue.Type() != whenNot.Type() {
		return nil, l.notYet(node, "a conditional whose branches have different types")
	}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot}, nil
}

func typeName(valueType ir.Type) string {
	switch valueType {
	case ir.Number:
		return "number"
	case ir.Boolean:
		return "boolean"
	case ir.String:
		return "string"
	}
	return "value"
}

// call lowers a call to one of the module's functions.
func (l *lowering) call(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	function, isFunction := l.functions[l.checker.GetSymbolAtLocation(callee)]
	if !ast.IsIdentifier(callee) || !isFunction {
		return nil, l.notYet(node, "a call to "+describe(callee))
	}
	arguments := []ir.Expression{}
	for _, argument := range call.Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, lowered)
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns}, nil
}

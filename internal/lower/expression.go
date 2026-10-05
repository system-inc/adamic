package lower

import (
	"errors"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// typeOf is what's left at runtime of the type the checker proved for a node: a number, a boolean or
// a string. A union counts when every member is the same one ('Fizz' | 'Buzz' is a string).
func (l *lowering) typeOf(node *ast.Node) (ir.Type, error) {
	if valueType, isKnown := l.representation(l.checker.GetTypeAtLocation(node)); isKnown {
		return valueType, nil
	}
	return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
}

func (l *lowering) representation(proven *checker.Type) (ir.Type, bool) {
	flags := proven.Flags()
	if flags&checker.TypeFlagsTypeParameter != 0 {
		// Inside a generic class, a type parameter is what this instantiation made it.
		substituted, isKnown := l.substitution[proven]
		return substituted, isKnown
	}
	switch {
	case flags&checker.TypeFlagsNumberLike != 0:
		return ir.Number, true
	case flags&checker.TypeFlagsStringLike != 0:
		return ir.String, true
	case flags&checker.TypeFlagsBooleanLike != 0:
		return ir.Boolean, true
	case flags&checker.TypeFlagsObject != 0 && l.checker.IsArrayType(proven):
		return ir.Array, true
	case flags&checker.TypeFlagsObject != 0 && l.isLibraryType(proven, "Map", "ReadonlyMap"):
		return ir.Map, true
	case flags&checker.TypeFlagsObject != 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0:
		return ir.Object, true
	case flags&checker.TypeFlagsObject != 0:
		// An object with call signatures is a function, held as a closure.
		return ir.Closure, true
	case flags&checker.TypeFlagsUnion != 0:
		var shared ir.Type
		for _, member := range proven.Types() {
			if member.Flags()&checker.TypeFlagsUndefined != 0 {
				// undefined joins a union of references as a null pointer; it's checked below that
				// the rest are references.
				continue
			}
			memberType, isKnown := l.representation(member)
			if !isKnown || (shared != 0 && memberType != shared) {
				return 0, false
			}
			shared = memberType
		}
		if shared != 0 && !shared.IsReference() && l.includesUndefined(proven) {
			// number | undefined is a present-and-value pair; boolean | undefined is not yet.
			if shared == ir.Number {
				return ir.MaybeNumber, true
			}
			return 0, false
		}
		return shared, shared != 0
	}
	return 0, false
}

// isLibraryType reports whether a type is one of the library's, by name: Map, not a program's own
// interface that happens to be called Map.
func (l *lowering) isLibraryType(proven *checker.Type, names ...string) bool {
	symbol := proven.Symbol()
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return false
	}
	for _, name := range names {
		if symbol.Name == name {
			return true
		}
	}
	return false
}

func (l *lowering) includesUndefined(proven *checker.Type) bool {
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			return true
		}
	}
	return false
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
		if !isLocal && node.Text() == "undefined" {
			return ir.Undefined{}, nil
		}
		if !isLocal {
			return nil, l.notYet(node, "reading "+node.Text())
		}
		read := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)})
		if l.result.Locals[local].Type == ir.MaybeNumber {
			// Where the checker has narrowed it to number, it's read as one.
			if narrowed, _ := l.representation(l.checker.GetTypeAtLocation(node)); narrowed == ir.Number {
				read = ir.Unwrap{Value: read}
			}
		}
		return read, nil
	case ast.KindPrefixUnaryExpression:
		return l.prefix(node)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			return l.coalesce(node)
		}
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
	case ast.KindObjectLiteralExpression:
		return l.objectLiteral(node)
	case ast.KindArrayLiteralExpression:
		return l.arrayLiteral(node)
	case ast.KindPropertyAccessExpression:
		return l.property(node)
	case ast.KindElementAccessExpression:
		return l.elementAccess(node)
	case ast.KindNewExpression:
		return l.newExpression(node)
	case ast.KindThisKeyword:
		if l.this < 0 {
			return nil, l.notYet(node, "this outside a method")
		}
		l.touch(l.this)
		return ir.Read{Local: l.this, Of: ir.Object}, nil
	case ast.KindArrowFunction:
		return l.closure(node)
	case ast.KindAsExpression:
		return l.cast(node)
	case ast.KindFunctionExpression:
		return nil, l.notYet(node, "a function expression (an arrow function captures this as written)")
	case ast.KindCallExpression:
		if lowered, isBuiltin, err := l.builtin(node); isBuiltin {
			if err == nil && lowered.Type() == 0 {
				// forEach is void; as a value it's undefined, which only places 0.1 refuses would use.
				return nil, l.notYet(node, "a void call used as a value")
			}
			return lowered, err
		}
		call, err := l.callOrMethod(node)
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
	if lowered, isComparison := comparisons[operator]; isComparison && (both(ir.Number) || both(ir.String)) {
		// Strings compare in UTF-16 code unit order, as JavaScript's do.
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		// x === undefined tests for a missing reference, whatever x's type.
		_, leftUndefined := left.(ir.Undefined)
		_, rightUndefined := right.(ir.Undefined)
		if leftUndefined != rightUndefined {
			value := left
			if leftUndefined {
				value = right
			}
			if !value.Type().IsReference() && value.Type() != ir.MaybeNumber {
				return nil, l.notYet(node, "comparing a "+typeName(value.Type())+" with undefined")
			}
			test := ir.Expression(ir.IsUndefined{Value: value})
			if operator == ast.KindExclamationEqualsEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return test, nil
		}
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

// call lowers a call to one of the module's functions, or to a function value.
func (l *lowering) call(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	function, isFunction := l.functions[l.symbol(callee)]
	if !ast.IsIdentifier(callee) || !isFunction {
		if calleeType, _ := l.representation(l.checker.GetTypeAtLocation(callee)); calleeType == ir.Closure {
			return l.callClosure(node)
		}
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

// coalesce lowers value ?? fallback, and value ?? panic('why'), evaluating the right side only when
// the left is missing.
func (l *lowering) coalesce(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	value, err := l.expression(binary.Left)
	if err != nil {
		return nil, err
	}
	present := value.Type()
	if present == ir.MaybeNumber {
		present = ir.Number
	} else if !present.IsReference() {
		// A value that can't be missing: ?? never runs its right side.
		return value, nil
	}
	right := ast.SkipParentheses(binary.Right)
	if right.Kind == ast.KindCallExpression && l.isPreludeFunction(right.AsCallExpression().Expression, "panic") && len(right.AsCallExpression().Arguments.Nodes) == 1 {
		message, err := l.expression(right.AsCallExpression().Arguments.Nodes[0])
		if err != nil {
			return nil, err
		}
		return ir.Coalesce{Value: value, Panic: message, Of: present}, nil
	}
	fallback, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if fallback.Type() != present {
		return nil, l.notYet(node, "?? whose sides have different types")
	}
	return ir.Coalesce{Value: value, Fallback: fallback, Of: present}, nil
}

// closure lowers an arrow function to a function of its own and the closure that captures it.
func (l *lowering) closure(node *ast.Node) (ir.Expression, error) {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "closure", Closure: true})
	l.closures = append(l.closures, index)
	err := l.lowerFunction(index, node, -1)
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}

// callClosure lowers a call through a function value.
func (l *lowering) callClosure(node *ast.Node) (ir.Expression, error) {
	closure, err := l.expression(node.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, lowered)
	}
	var returns ir.Type
	if result := l.checker.GetTypeAtLocation(node); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = l.representation(result); !isKnown {
			return nil, l.notYet(node, "a call returning "+l.checker.TypeToString(result))
		}
	}
	for _, argument := range arguments {
		if argument.Type() == ir.MaybeNumber {
			return nil, l.notYet(node, "passing number | undefined to a function value")
		}
	}
	if returns == ir.MaybeNumber {
		return nil, l.notYet(node, "a function value returning number | undefined")
	}
	return ir.CallClosure{Closure: closure, Arguments: arguments, Returns: returns}, nil
}

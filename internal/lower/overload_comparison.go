package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Each public overload must admit two values of the same primitive kind.
// Immutable parameters preserve that correlation inside the union-typed body.
// This authorizes the existing same-kind comparisons, never mixed coercion.
func (l *lowering) overloadComparisonPair(node *ast.Node) bool {
	binary := node.AsBinaryExpression()
	left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	if !ast.IsIdentifier(left) || !ast.IsIdentifier(right) {
		return false
	}
	for _, operand := range []*ast.Node{left, right} {
		t := l.checker.GetTypeAtLocation(operand)
		if l.includesUndefined(t) || l.includesNull(t) {
			return false
		}
	}
	owner := node.Parent
	for owner != nil && !ast.IsFunctionLike(owner) {
		owner = owner.Parent
	}
	if owner == nil || owner.Kind != ast.KindFunctionDeclaration || len(owner.TypeParameters()) != 0 {
		return false
	}
	symbols := []*ast.Symbol{l.symbol(left), l.symbol(right)}
	positions := []int{-1, -1}
	for index, parameter := range owner.Parameters() {
		for side, symbol := range symbols {
			if ast.IsIdentifier(parameter.Name()) && l.symbol(parameter.Name()) == symbol {
				if parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
					return false
				}
				positions[side] = index
			}
		}
	}
	if positions[0] < 0 || positions[1] < 0 {
		return false
	}
	symbol := l.symbol(owner.Name())
	if symbol == nil {
		return false
	}
	overloads := 0
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() != nil {
			continue
		}
		signature := l.checker.GetSignatureFromDeclaration(declaration)
		parameters := signature.Parameters()
		var kind ir.Type
		for _, position := range positions {
			if position >= len(parameters) {
				return false
			}
			t := l.checker.GetTypeOfSymbol(parameters[position])
			if l.includesNull(t) {
				return false
			}
			of, known := l.representation(l.checker.GetNonNullableType(t))
			if !known || of != ir.Number && of != ir.String {
				return false
			}
			if kind == 0 {
				kind = of
			} else if kind != of {
				return false
			}
		}
		overloads++
	}
	if overloads == 0 {
		return false
	}
	unchanged := true
	var target func(*ast.Node) bool
	target = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && (l.symbol(node) == symbols[0] || l.symbol(node) == symbols[1]) {
			unchanged = false
		}
		return node.ForEachChild(target)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			_, compound := compoundAssignments[binary.OperatorToken.Kind]
			if binary.OperatorToken.Kind == ast.KindEqualsToken || compound || logicalAssignment(binary.OperatorToken.Kind) {
				target(binary.Left)
			}
		}
		if node.Kind == ast.KindPrefixUnaryExpression {
			unary := node.AsPrefixUnaryExpression()
			if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
				target(unary.Operand)
			}
		}
		if node.Kind == ast.KindPostfixUnaryExpression {
			target(node.AsPostfixUnaryExpression().Operand)
		}
		if node.Kind == ast.KindForOfStatement || node.Kind == ast.KindForInStatement {
			target(node.AsForInOrOfStatement().Initializer)
		}
		return node.ForEachChild(visit)
	}
	owner.Body().ForEachChild(visit)
	return unchanged
}

func (l *lowering) overloadUnionComparison(node *ast.Node, operator ir.Operator, left, right ir.Expression) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{left, right})
	a, c := b.read(b.parameters[0]), b.read(b.parameters[1])
	for _, kind := range []ir.Type{ir.String, ir.Number} {
		name := "string"
		if kind == ir.Number {
			name = "number"
		}
		matches := func(value ir.Expression) ir.Expression {
			return ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(name)}}
		}
		condition := ir.Binary{Operator: ir.And, Left: matches(a), Right: matches(c)}
		b.body = append(b.body, ir.If{Condition: condition, Then: []ir.Statement{ir.Return{Value: ir.Binary{Operator: operator, Left: ir.Narrow{Value: a, To: kind}, Right: ir.Narrow{Value: c, To: kind}}}}})
	}
	b.body = append(b.body, ir.Panic{Message: ir.StringConstant{Index: l.constant("overloaded comparison operands violate their same-kind primitive contract")}})
	return b.finish("overload_primitive_comparison", ir.BooleanConstant{Value: false})
}

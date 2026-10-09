package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Flow narrowing describes values, not the layout of a binding's existing array.
func (l *lowering) arrayStorageType(node *ast.Node) *checker.Type {
	actual := l.arrayPredicateType(node)
	skipped := ast.SkipParentheses(node)
	var symbol *ast.Symbol
	if ast.IsIdentifier(skipped) {
		symbol = l.symbol(skipped)
	}
	if skipped.Kind == ast.KindPropertyAccessExpression {
		symbol = l.checker.GetSymbolAtLocation(skipped.Name())
	}
	if symbol != nil {
		declared := l.concrete(l.checker.GetTypeOfSymbol(symbol))
		if l.checker.IsArrayType(l.checker.GetNonNullableType(declared)) {
			declared = l.checker.GetNonNullableType(declared)
			if of, known := l.kept(l.checker.GetElementTypeOfArrayType(declared)); known && of == ir.Union {
				return declared
			}
		}
	}
	return actual
}

func (l *lowering) checkedArrayNarrow(value ir.Expression, to ir.Type) ir.Expression {
	if to == ir.Union {
		return value
	}
	name := "object"
	switch to.Present() {
	case ir.Number:
		name = "number"
	case ir.Boolean:
		name = "boolean"
	case ir.String:
		name = "string"
	case ir.Closure:
		name = "function"
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant(name)}})
	if to.IsMaybe() {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "array element does not match its narrowed type"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	result := b.finish("narrowed_array_element", ir.Narrow{Value: held, To: to})
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return result
}

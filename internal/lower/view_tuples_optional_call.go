package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// The original tuple consumer uses receiver?.forEach(callback). Keep the
// receiver once and evaluate the callback only in the present branch. This
// statement hook reuses ArrayVisit and its checked element extraction.
func (l *lowering) tupleOptionalForEach(node *ast.Node) ([]ir.Statement, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().QuestionDotToken == nil || callee.Name().Text() != "forEach" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	present := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	element := l.viewArrayElementType(present)
	if element == nil || !fixedViewTuple(element) || l.includesNull(l.checker.GetTypeAtLocation(receiver)) {
		return nil, false, nil
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if array.Type() != ir.Array {
		return nil, false, nil
	}
	of, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	held := l.libraryLocal(l.functionIndex, "tuple_optional_receiver", ir.Array)
	visit, _, err := l.arrayVisit(node, held, of, "forEach")
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{
		ir.Declare{Local: held.Local, Value: array},
		ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: held}}, Then: []ir.Statement{ir.Evaluate{Value: visit}}},
	}, true, nil
}

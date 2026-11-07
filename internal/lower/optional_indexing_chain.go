package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// This continuation owns only a?.field[index]. The receiver guard must skip
// the index, while a missing element still reaches a following optional read.
func (lowering *lowering) optionalIndexContinuation(node *ast.Node) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	receiver := access.Expression
	// Parentheses terminate a continuous chain; do not skip them here.
	if receiver.Kind != ast.KindPropertyAccessExpression {
		return nil, lowering.notYet(node, "an optional chain longer than one step")
	}
	field := receiver.AsPropertyAccessExpression()
	if field.QuestionDotToken == nil {
		return nil, lowering.notYet(node, "an optional chain longer than one step")
	}
	value, propertyError := lowering.property(receiver)
	if propertyError != nil {
		return nil, propertyError
	}
	property, known := value.(ir.Property)
	if !known ||
		property.Absent ||
		property.Method ||
		property.Of == ir.Weak ||
		!(property.Of == ir.Array || property.Of == ir.String || property.Of.IsTypedArray()) {
		return nil, lowering.notYet(node, "an indexing continuation without a required array or string field")
	}
	function := -1
	if lowering.function != nil {
		function = lowering.functionIndex
	}
	local := len(lowering.result.Locals)
	base := property.Object
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "optional_index_chain_base", Type: base.Type(), Function: function, ExpressionAssigned: true},
	)
	read := ir.Read{Local: local, Of: base.Type()}
	property.Object, property.Optional = read, false
	indexed, optionalIndexValueError := lowering.optionalIndexValue(node, property)
	if optionalIndexValueError != nil {
		return nil, optionalIndexValueError
	}
	of := indexed.Type()
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: base}}, Result: ir.Conditional{
		Condition: ir.Unary{
			Operator: ir.Not,
			Operand:  ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}},
		},
		WhenTrue: indexed, WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, nil
}

package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Clearing or decrementing length only removes elements, so it needs no hole
// representation. Existing ArrayPop preserves ownership, and Throw is catchable.
func (l *lowering) arrayLengthRemoval(target *ast.Node, array ir.Expression, clear bool) ([]ir.Statement, error) {
	element, err := l.elementType(target.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{array})
	held := b.read(b.parameters[0])
	pop := ir.Evaluate{Value: ir.ArrayPop{Array: held, Element: element}}
	if clear {
		b.body = append(b.body, ir.Loop{Condition: ir.Binary{Operator: ir.Greater, Left: ir.Length{Array: held}, Right: ir.NumberConstant{Value: 0}}, Body: []ir.Statement{pop}})
	} else {
		b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: ir.Length{Array: held}, Right: ir.NumberConstant{Value: 0}}, Then: []ir.Statement{ir.Throw{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant("Invalid array length")}, Name: ir.StringConstant{Index: l.constant("RangeError")}}}}, Else: []ir.Statement{pop}})
	}
	return []ir.Statement{ir.Evaluate{Value: b.finish("array_length_removal", ir.BooleanConstant{Value: false})}}, nil
}

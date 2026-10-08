package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Fixed tuples keep numeric fields, rather than array storage. Read the selected
// field anew each iteration so writes to later elements remain visible.
func (l *lowering) forOfTuple(node, name *ast.Node, source ir.Expression, proven *checker.Type) ([]ir.Statement, error) {
	elements := l.checker.GetTypeArguments(proven)
	if checker.TupleType_combinedFlags(proven.TargetTupleType())&checker.ElementFlagsNonRequired != 0 || len(elements) == 0 {
		return nil, l.notYet(node, "for...of over a tuple with optional, rest or no elements")
	}
	if !ast.IsIdentifier(name) {
		return nil, l.notYet(name, "destructuring in for...of over a tuple")
	}
	local, err := l.declareLocal(name)
	if err != nil {
		return nil, err
	}
	of := l.result.Locals[local].Type
	for _, element := range elements {
		stored, known := l.representation(element)
		if !known || stored != of || (of != ir.String && of != ir.Number && of != ir.Boolean) {
			return nil, l.notYet(node, "for...of over a tuple without homogeneous primitive storage")
		}
	}
	held := l.iterationLocal("iteration_tuple", ir.Object, l.functionIndex)
	index := l.iterationLocal("iteration_index", ir.Number, l.functionIndex)
	read := ir.Read{Local: held, Of: ir.Object}
	position := ir.Read{Local: index, Of: ir.Number}
	var value ir.Expression = ir.Property{Object: read, Name: strconv.Itoa(len(elements) - 1), Of: of}
	for slot := len(elements) - 2; slot >= 0; slot-- {
		value = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: position, Right: ir.NumberConstant{Value: float64(slot)}}, WhenTrue: ir.Property{Object: read, Name: strconv.Itoa(slot), Of: of}, WhenNot: value}
	}
	body, err := l.statement(node.AsForInOrOfStatement().Statement)
	if err != nil {
		return nil, err
	}
	loop := ir.Loop{
		Condition: ir.Binary{Operator: ir.Less, Left: position, Right: ir.NumberConstant{Value: float64(len(elements))}},
		Body:      append([]ir.Statement{ir.Declare{Local: local, Value: value}}, body...),
		Update:    []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: position, Right: ir.NumberConstant{Value: 1}}}},
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{ir.Declare{Local: held, Value: source}, ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}}, loop}}}, nil
}

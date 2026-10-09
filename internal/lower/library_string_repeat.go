package lower

// Copyright 2019 the V8 project authors. BSD-3-Clause; see THIRD_PARTY_NOTICES.md.
// Port of V8 13.6.233.17 src/builtins/string-repeat.tq into ordinary typed IR.
import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) stringRepeatCall(node *ast.Node, value ir.Expression, written []*ast.Node, provided []ir.Expression) (ir.Expression, bool, error) {
	if len(written) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "repeat without one proven number argument")
	}
	var count ir.Expression
	var err error
	if provided != nil {
		count = provided[0]
	} else {
		count, err = l.expression(written[0])
	}
	if err != nil {
		return nil, true, err
	}
	if count.Type() != ir.Number {
		return nil, true, l.notYet(node, "repeat without a proven number count")
	}
	function, reads := l.stringHelper("repeat", []ir.Expression{value, count})
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: function})
		return index
	}
	n, power, text := local("count", ir.Number), local("power", ir.String), local("text", ir.String)
	readN := ir.Read{Local: n, Of: ir.Number}
	readPower := ir.Read{Local: power, Of: ir.String}
	readText := ir.Read{Local: text, Of: ir.String}
	zero := ir.NumberConstant{Value: 0}
	empty := ir.StringConstant{Index: l.constant("")}
	binary := func(op ir.Operator, left, right ir.Expression) ir.Expression {
		return ir.Binary{Operator: op, Left: left, Right: right}
	}
	failure := func(message ir.Expression) []ir.Statement {
		return []ir.Statement{ir.Throw{Value: ir.MakeError{
			Name: ir.StringConstant{Index: l.constant("RangeError")}, Message: message,
		}}}
	}
	invalidCount := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Invalid count value: ")}, ir.NumberToString{Value: reads[1]}}}
	l.result.Functions[function].Body = []ir.Statement{
		ir.Declare{Local: n, Value: stringInteger(reads[1])},
		ir.If{Condition: binary(ir.Or, binary(ir.Less, readN, zero), ir.Unary{Operator: ir.Not, Operand: ir.NumberCall{Function: "isFinite", Arguments: []ir.Expression{readN}}}), Then: failure(invalidCount)},
		ir.If{Condition: binary(ir.Or, binary(ir.Equal, readN, zero), binary(ir.Equal, ir.StringLength{Value: reads[0]}, zero)), Then: []ir.Statement{ir.Return{Value: empty}}},
		// No observable allocation precedes this equivalent length check. Every later
		// doubling is bounded by the final length, so the concat panic is unreachable.
		ir.If{Condition: binary(ir.Greater, binary(ir.Multiply, ir.StringLength{Value: reads[0]}, readN), ir.NumberConstant{Value: 536870888}), Then: failure(ir.StringConstant{Index: l.constant("Invalid string length")})},
		ir.Declare{Local: power, Value: reads[0]}, ir.Declare{Local: text, Value: empty},
		ir.Loop{Condition: binary(ir.Greater, readN, zero), Body: []ir.Statement{
			ir.If{Condition: binary(ir.Equal, binary(ir.Remainder, readN, ir.NumberConstant{Value: 2}), ir.NumberConstant{Value: 1}), Then: []ir.Statement{ir.Assign{Local: text, Value: ir.Concat{Parts: []ir.Expression{readText, readPower}}}}},
			ir.Assign{Local: n, Value: ir.MathCall{Function: "floor", Arguments: []ir.Expression{binary(ir.Divide, readN, ir.NumberConstant{Value: 2})}}},
			ir.If{Condition: binary(ir.Greater, readN, zero), Then: []ir.Statement{ir.Assign{Local: power, Value: ir.Concat{Parts: []ir.Expression{readPower, readPower}}}}},
		}}, ir.Return{Value: readText},
	}
	return ir.Call{Function: function, Arguments: []ir.Expression{value, count}, Returns: ir.String}, true, nil
}

package lower

import "github.com/system-inc/adamic/internal/ir"

// Port of V8's src/builtins/string-substr.tq. The receiver is already a string;
// numeric operands have no observable conversion. Bind all operands once before
// clamping start relative to the UTF-16 length and limiting the requested count.
func (l *lowering) stringSubstrValues(value ir.Expression, arguments []ir.Expression) ir.Expression {
	values := append([]ir.Expression{value}, arguments...)
	function, reads := l.stringHelper("substr", values)
	zero := ir.NumberConstant{Value: 0}
	size := ir.StringLength{Value: reads[0]}
	start := ir.Expression(zero)
	if len(reads) > 1 {
		start = stringInteger(reads[1])
	}
	start = ir.Conditional{Condition: ir.Binary{Operator: ir.Less, Left: start, Right: zero}, WhenTrue: ir.MathCall{Function: "max", Arguments: []ir.Expression{ir.Binary{Operator: ir.Add, Left: size, Right: start}, zero}}, WhenNot: ir.MathCall{Function: "min", Arguments: []ir.Expression{start, size}}}
	count := ir.Expression(ir.Binary{Operator: ir.Subtract, Left: size, Right: start})
	if len(reads) > 2 {
		count = ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{stringInteger(reads[2]), zero}}, count}}
	}
	result := ir.StringCall{Method: "slice", Value: reads[0], Arguments: []ir.Expression{start, ir.Binary{Operator: ir.Add, Left: start, Right: count}}}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}
}

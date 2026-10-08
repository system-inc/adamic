package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"math"
)

func (l *lowering) stringRangeThrow(message ir.Expression) ir.Statement {
	return ir.Throw{Value: ir.BuiltinError{Kind: 3, Message: message}}
}

// Port of V8 src/builtins/builtins-string.cc StringFromCodePoint validation.
// All arguments are evaluated before the first validation. Spread remains the
// language owner's argument-count/stack-limit dependency around a try.
func (l *lowering) stringCodePointsChecked(call ir.StringFromCodes) ir.Expression {
	if !call.CodePoints || call.Spread != nil {
		return call
	}
	f, reads := l.stringHelper("codepoints_checked", call.Codes)
	body := []ir.Statement{}
	for _, code := range reads {
		noninteger := ir.Binary{Operator: ir.NotEqual, Left: code, Right: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{code}}}
		outside := ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Less, Left: code, Right: ir.NumberConstant{Value: 0}}, Right: ir.Binary{Operator: ir.Greater, Left: code, Right: ir.NumberConstant{Value: 0x10ffff}}}
		message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Invalid code point ")}, ir.NumberToString{Value: code}}}
		body = append(body, ir.If{Condition: ir.Binary{Operator: ir.Or, Left: noninteger, Right: outside}, Then: []ir.Statement{l.stringRangeThrow(message)}})
	}
	body = append(body, ir.Return{Value: ir.StringFromCodes{Codes: reads, CodePoints: true, Validated: true}})
	l.result.Functions[f].Body = body
	return ir.Call{Function: f, Arguments: call.Codes, Returns: ir.String}
}

// Port of V8 src/builtins/string-repeat.tq, with the same 64-bit V8 maximum
// already enforced by adamic_string_check_length. Throws use shared error identity.
func (l *lowering) stringRepeatChecked(text, count ir.Expression) ir.Expression {
	f, reads := l.stringHelper("repeat_checked", []ir.Expression{text, count})
	n := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "count", Type: ir.Number, Function: f})
	read := ir.Read{Local: n, Of: ir.Number}
	negative := ir.Binary{Operator: ir.Less, Left: read, Right: ir.NumberConstant{Value: 0}}
	infinite := ir.Binary{Operator: ir.Equal, Left: ir.MathCall{Function: "abs", Arguments: []ir.Expression{read}}, Right: ir.NumberConstant{Value: math.Inf(1)}}
	message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Invalid count value: ")}, ir.NumberToString{Value: reads[1]}}}
	length := ir.Binary{Operator: ir.Multiply, Left: ir.StringLength{Value: reads[0]}, Right: read}
	tooLong := ir.Binary{Operator: ir.Greater, Left: length, Right: ir.NumberConstant{Value: 536870888}}
	l.result.Functions[f].Body = []ir.Statement{
		ir.Declare{Local: n, Value: stringInteger(reads[1])},
		ir.If{Condition: ir.Binary{Operator: ir.Or, Left: negative, Right: infinite}, Then: []ir.Statement{l.stringRangeThrow(message)}},
		ir.If{Condition: tooLong, Then: []ir.Statement{l.stringRangeThrow(ir.StringConstant{Index: l.constant("Invalid string length")})}},
		ir.Return{Value: ir.StringCall{Method: "repeat", Value: reads[0], Arguments: []ir.Expression{read}, Validated: true}},
	}
	return ir.Call{Function: f, Arguments: []ir.Expression{text, count}, Returns: ir.String}
}

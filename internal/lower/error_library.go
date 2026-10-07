package lower

import (
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A checked library call uses an ordinary IR throw, so the existing exception cleanup
// handles the error before any runtime routine can enter its non-unwinding failure path.
func (l *lowering) checkedLibrary(where *ast.Node, arguments []ir.Expression, build func([]ir.Expression) (ir.Expression, ir.Expression, ir.Expression)) ir.Expression {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "error_checked_library", LibraryGuarded: true})
	reads := []ir.Expression{}
	for _, value := range arguments {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: value.Type(), Function: index})
		l.result.Functions[index].Parameters = append(l.result.Functions[index].Parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	operation, invalid, message := build(reads)
	errorClass := l.errorInstance("RangeError")
	failure := ir.Call{Function: errorClass.constructor, Returns: ir.Object, Arguments: []ir.Expression{message, ir.Undefined{Of: ir.Union}}}
	l.result.Functions[index].Returns = operation.Type()
	l.result.Functions[index].Body = []ir.Statement{ir.If{Condition: invalid, Then: []ir.Statement{ir.Throw{Value: failure}}}, ir.Return{Value: operation}}
	return ir.Call{Function: index, Returns: operation.Type(), Arguments: arguments, Pure: true}
}

func errorOutside(value ir.Expression, low, high float64) ir.Expression {
	return ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Less, Left: value, Right: ir.NumberConstant{Value: low}}, Right: ir.Binary{Operator: ir.Greater, Left: value, Right: ir.NumberConstant{Value: high}}}
}

func (l *lowering) checkedToFixed(where *ast.Node, value, digits ir.Expression) ir.Expression {
	if constantWithin(digits, 0, 100) {
		return ir.ToFixed{Value: value, Digits: digits}
	}
	return l.checkedLibrary(where, []ir.Expression{value, digits}, func(args []ir.Expression) (ir.Expression, ir.Expression, ir.Expression) {
		integer := ir.MathCall{Function: "trunc", Arguments: []ir.Expression{args[1]}}
		return ir.ToFixed{Value: args[0], Digits: args[1]}, errorOutside(integer, 0, 100), ir.StringConstant{Index: l.constant("toFixed() digits argument must be between 0 and 100")}
	})
}

func (l *lowering) checkedNumberFormat(where *ast.Node, format ir.NumberFormat) ir.Expression {
	if format.Argument == nil || constantWithin(format.Argument, formatArguments[format.Method][0], formatArguments[format.Method][1]) {
		return format
	}
	return l.checkedLibrary(where, []ir.Expression{format.Value, format.Argument}, func(args []ir.Expression) (ir.Expression, ir.Expression, ir.Expression) {
		bounds := formatArguments[format.Method]
		invalid := errorOutside(ir.MathCall{Function: "trunc", Arguments: []ir.Expression{args[1]}}, bounds[0], bounds[1])
		// NaN becomes zero in ToIntegerOrInfinity. Precision and radix reject that zero.
		if bounds[0] > 0 {
			invalid = ir.Binary{Operator: ir.Or, Left: invalid, Right: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{args[1]}}}
		}
		if format.Method != "toString" {
			invalid = ir.Binary{Operator: ir.And, Left: ir.NumberCall{Function: "isFinite", Arguments: []ir.Expression{args[0]}}, Right: invalid}
		}
		message := format.Method + "() argument must be between "
		if format.Method == "toString" {
			message = "toString() radix argument must be between 2 and 36"
		} else if format.Method == "toPrecision" {
			message += "1 and 100"
		} else {
			message += "0 and 100"
		}
		return ir.NumberFormat{Method: format.Method, Value: args[0], Argument: args[1]}, invalid, ir.StringConstant{Index: l.constant(message)}
	})
}

func (l *lowering) checkedStringCall(where *ast.Node, call ir.StringCall) ir.Expression {
	if call.Method != "repeat" && call.Method != "normalize" {
		return call
	}
	if call.Method == "normalize" && isNormalizationForm(call.Arguments[0], l.result.Strings) {
		if text, ok := call.Value.(ir.StringConstant); ok && len(l.result.Strings[text.Index]) <= 536870888/36 {
			return call
		}
	}
	if call.Method == "repeat" && constantWithin(call.Arguments[0], 0, 1) {
		return call
	}
	if call.Method == "repeat" {
		if text, ok := call.Value.(ir.StringConstant); ok && constantWithin(call.Arguments[0], 0, math.MaxFloat64) {
			count := call.Arguments[0].(ir.NumberConstant).Value
			// UTF-8 byte length is an upper bound on UTF-16 units.
			if float64(len(l.result.Strings[text.Index]))*math.Trunc(count) <= 536870888 {
				return call
			}
		}
	}
	guarded := l.checkedLibrary(where, append([]ir.Expression{call.Value}, call.Arguments...), func(args []ir.Expression) (ir.Expression, ir.Expression, ir.Expression) {
		var invalid, message ir.Expression
		if call.Method == "repeat" {
			integer := ir.MathCall{Function: "trunc", Arguments: []ir.Expression{args[1]}}
			invalid = ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Less, Left: integer, Right: ir.NumberConstant{}}, Right: ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Equal, Left: integer, Right: ir.NumberConstant{Value: math.Inf(1)}}, Right: ir.Binary{Operator: ir.Equal, Left: integer, Right: ir.NumberConstant{Value: math.Inf(-1)}}}}
			message = ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Invalid count value: ")}, ir.NumberToString{Value: args[1]}}}
			tooLong := ir.Binary{Operator: ir.Greater, Left: ir.Binary{Operator: ir.Multiply, Left: integer, Right: ir.StringLength{Value: args[0]}}, Right: ir.NumberConstant{Value: 536870888}}
			message = ir.Conditional{Condition: invalid, WhenTrue: message, WhenNot: ir.StringConstant{Index: l.constant("Invalid string length")}}
			invalid = ir.Binary{Operator: ir.Or, Left: invalid, Right: tooLong}
		} else {
			valid := ir.Expression(ir.BooleanConstant{})
			for _, form := range []string{"NFC", "NFD", "NFKC", "NFKD"} {
				valid = ir.Binary{Operator: ir.Or, Left: valid, Right: ir.Binary{Operator: ir.Equal, Left: args[1], Right: ir.StringConstant{Index: l.constant(form)}}}
			}
			invalid = ir.Unary{Operator: ir.Not, Operand: valid}
			message = ir.StringConstant{Index: l.constant("The normalization form should be one of NFC, NFD, NFKC, NFKD.")}
		}
		return ir.StringCall{Method: call.Method, Value: args[0], Arguments: args[1:]}, invalid, message
	})
	if call.Method == "normalize" {
		// Unicode's longest compatibility decomposition has 18 code points. Counting
		// two UTF-16 units per point also bounds supplementary decompositions.
		text, constant := call.Value.(ir.StringConstant)
		if !constant || len(l.result.Strings[text.Index]) > 536870888/36 {
			l.result.Functions[guarded.(ir.Call).Function].LibraryGuarded = false
		}
	}
	return guarded
}

// A narrowing can be invalidated by a call. Property reads throw JavaScript's
// nominal TypeError; the ordinary IR throw also lets an enclosing catch narrow it.
func (l *lowering) errorDefined(value ir.Expression, message string) ir.Expression {
	index := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: value.Type(), Function: index})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "error_defined", Parameters: []int{local}, Returns: value.Type(), LibraryGuarded: true})
	read := ir.Read{Local: local, Of: value.Type()}
	failure := l.errorInstance("TypeError")
	missing := ir.Expression(ir.IsUndefined{Value: read})
	if strings.Contains(message, "properties of null") {
		missing = ir.IsNull{Value: read}
	}
	l.result.Functions[index].Body = []ir.Statement{
		ir.If{Condition: missing, Then: []ir.Statement{ir.Throw{Value: ir.Call{Function: failure.constructor, Returns: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: l.constant(message)}, ir.Undefined{Of: ir.Union}}}}}},
		ir.Return{Value: read},
	}
	return ir.Call{Function: index, Returns: value.Type(), Arguments: []ir.Expression{value}}
}

// Evaluate the receiver and right side before checking the write, as JavaScript does.
func (l *lowering) errorSetProperty(write ir.SetProperty) ir.Statement {
	index := len(l.result.Functions)
	object := len(l.result.Locals)
	value := object + 1
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: ir.Object, Function: index}, ir.Local{Name: "value", Type: write.Value.Type(), Function: index})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "error_set_property", Parameters: []int{object, value}, LibraryGuarded: true})
	read := ir.Read{Local: object, Of: ir.Object}
	failure := l.errorInstance("TypeError")
	stored := write
	stored.Object = read
	stored.Value = ir.Read{Local: value, Of: write.Value.Type()}
	l.result.Functions[index].Body = []ir.Statement{
		ir.If{Condition: ir.IsUndefined{Value: read}, Then: []ir.Statement{ir.Throw{Value: ir.Call{Function: failure.constructor, Returns: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: l.constant("Cannot set properties of undefined (setting '" + write.Name + "')")}, ir.Undefined{Of: ir.Union}}}}}},
		stored,
	}
	return ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{write.Object, write.Value}}}
}

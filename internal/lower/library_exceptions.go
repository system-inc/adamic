package lower

import (
	"math"
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// Library validation is ordinary IR: both backends throw the same nominal Error,
// and existing call effects carry the exceptional edge through every owning frame.
// The primitive is reached only after its arguments have been checked.
func (l *lowering) libraryExceptions() {
	count := len(l.result.Functions)
	l.libraryExceptionTree(reflect.ValueOf(&l.result.Main).Elem())
	for index := 0; index < count; index++ {
		l.libraryExceptionTree(reflect.ValueOf(&l.result.Functions[index].Body).Elem())
	}
}

func (l *lowering) libraryExceptionTree(value reflect.Value) {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		l.libraryExceptionTree(copy)
		if expression, ok := copy.Interface().(ir.Expression); ok {
			value.Set(reflect.ValueOf(l.libraryException(expression)))
		} else if statement, ok := copy.Interface().(ir.Statement); ok {
			value.Set(reflect.ValueOf(l.libraryStatement(statement)))
		} else {
			value.Set(copy)
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			l.libraryExceptionTree(value.Field(index))
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			l.libraryExceptionTree(value.Index(index))
		}
	}
}

func binary(operator ir.Operator, left, right ir.Expression) ir.Expression {
	return ir.Binary{Operator: operator, Left: left, Right: right}
}
func number(value float64) ir.Expression { return ir.NumberConstant{Value: value} }
func finite(value ir.Expression) ir.Expression {
	return ir.NumberCall{Function: "isFinite", Arguments: []ir.Expression{value}}
}
func integer(value ir.Expression) ir.Expression {
	if value.Type() == ir.MaybeNumber {
		value = ir.Coalesce{Value: value, Fallback: number(0), Of: ir.Number}
	}
	return ir.Conditional{Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}}, WhenTrue: number(0), WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{value}}, Of: ir.Number}
}
func outside(value ir.Expression, low, high float64) ir.Expression {
	return binary(ir.Or, binary(ir.Less, value, number(low)), binary(ir.Greater, value, number(high)))
}
func (l *lowering) libraryThrow(b *libraryArrayBuilder, condition ir.Expression, message ir.Expression) {
	b.body = append(b.body, ir.If{Condition: condition, Then: []ir.Statement{ir.Throw{Value: ir.Box{Value: ir.MakeError{Message: message, Constructor: "RangeError"}}}}})
}
func (l *lowering) checkedLibrary(b *libraryArrayBuilder, name string, result ir.Expression) ir.Expression {
	call := b.finish("exception_"+name, result)
	l.result.Functions[b.function].CheckedLibrary = true
	return call
}
func (l *lowering) libraryException(expression ir.Expression) ir.Expression {
	text := func(message string) ir.Expression { return ir.StringConstant{Index: l.constant(message)} }
	switch value := expression.(type) {
	case ir.RegExpCall:
		if value.Method == "matchAll" || value.Method == "replaceAll" {
			b := l.libraryArrayBuilder(append([]ir.Expression{value.Value}, value.Arguments...))
			regex := b.read(b.parameters[1])
			condition := ir.Unary{Operator: ir.Not, Operand: ir.Property{Object: regex, Name: "global", Of: ir.Boolean}}
			message := text("String.prototype." + value.Method + " called with a non-global RegExp argument")
			b.body = append(b.body, ir.If{Condition: condition, Then: []ir.Statement{ir.Throw{Value: ir.Box{Value: ir.MakeError{Message: message, Constructor: "TypeError"}}}}})
			value.Value = b.read(b.parameters[0])
			value.Arguments = append([]ir.Expression{}, value.Arguments...)
			for index := range value.Arguments {
				value.Arguments[index] = b.read(b.parameters[index+1])
			}
			return l.checkedLibrary(b, "regexp_global", value)
		}
	case ir.ObjectCall:
		if value.Method == "assign" && l.objectCanFreeze() {
			b := l.libraryArrayBuilder(value.Arguments)
			target := b.read(b.parameters[0])
			for _, parameter := range b.parameters[1:] {
				source := b.read(parameter)
				keys := b.read(b.declare("keys", ir.ObjectCall{Method: "keys", Arguments: []ir.Expression{source}, Returns: ir.Array, Element: ir.String}))
				condition := binary(ir.And, ir.ObjectCall{Method: "isFrozen", Arguments: []ir.Expression{fit(target, ir.Union)}, Returns: ir.Boolean}, binary(ir.Greater, ir.Length{Array: keys}, number(0)))
				key := ir.ArrayIndex{Array: keys, Index: number(0), Element: ir.String}
				message := ir.Concat{Parts: []ir.Expression{text("Cannot assign to read only property '"), key, text("' of object '#<Object>'")}}
				b.body = append(b.body, ir.If{Condition: condition, Then: []ir.Statement{ir.Throw{Value: ir.Box{Value: ir.MakeError{Message: message, Constructor: "TypeError"}}}}})
				b.body = append(b.body, ir.Evaluate{Value: ir.ObjectCall{Method: "assign", Arguments: []ir.Expression{target, source}, Returns: ir.Object}})
			}
			return l.checkedLibrary(b, "object_assign", target)
		}
	case ir.StringCall:
		if value.Method == "repeat" {
			b := l.libraryArrayBuilder(append([]ir.Expression{value.Value}, value.Arguments...))
			receiver, argument := b.read(b.parameters[0]), b.read(b.parameters[1])
			count := b.read(b.declare("count", integer(argument)))
			invalid := binary(ir.Or, binary(ir.Less, count, number(0)), ir.Unary{Operator: ir.Not, Operand: finite(count)})
			message := ir.Concat{Parts: []ir.Expression{text("Invalid count value: "), ir.NumberToString{Value: argument}}}
			l.libraryThrow(b, invalid, message)
			l.libraryThrow(b, binary(ir.Greater, binary(ir.Multiply, ir.StringLength{Value: receiver}, count), number(536870888)), text("Invalid string length"))
			value.Value = receiver
			value.Arguments = []ir.Expression{count}
			return l.checkedLibrary(b, "repeat", value)
		}
		if value.Method == "normalize" && len(value.Arguments) > 0 && !isNormalizationForm(value.Arguments[0], l.result.Strings) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Value, value.Arguments[0]})
			receiver, form := b.read(b.parameters[0]), b.read(b.parameters[1])
			valid := binary(ir.Equal, form, text("NFC"))
			for _, name := range []string{"NFD", "NFKC", "NFKD"} {
				valid = binary(ir.Or, valid, binary(ir.Equal, form, text(name)))
			}
			l.libraryThrow(b, ir.Unary{Operator: ir.Not, Operand: valid}, text("The normalization form should be one of NFC, NFD, NFKC, NFKD."))
			value.Value = receiver
			value.Arguments = []ir.Expression{form}
			return l.checkedLibrary(b, "normalize", value)
		}
	case ir.ToFixed:
		if !constantWithin(value.Digits, 0, 100) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Value, value.Digits})
			receiver, digits := b.read(b.parameters[0]), b.read(b.declare("digits", integer(b.read(b.parameters[1]))))
			l.libraryThrow(b, outside(digits, 0, 100), text("toFixed() digits argument must be between 0 and 100"))
			value.Value = receiver
			value.Digits = digits
			return l.checkedLibrary(b, "toFixed", value)
		}
	case ir.NumberFormat:
		bounds := formatArguments[value.Method]
		if value.Argument != nil && !constantWithin(value.Argument, bounds[0], bounds[1]) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Value, value.Argument})
			receiver, digits := b.read(b.parameters[0]), b.read(b.declare("digits", integer(b.read(b.parameters[1]))))
			invalid := outside(digits, bounds[0], bounds[1])
			message := "toString() radix argument must be between 2 and 36"
			if value.Method != "toString" {
				invalid = binary(ir.And, finite(receiver), invalid)
				if value.Method == "toPrecision" {
					message = "toPrecision() argument must be between 1 and 100"
				} else {
					message = "toExponential() argument must be between 0 and 100"
				}
			}
			l.libraryThrow(b, invalid, text(message))
			value.Value = receiver
			value.Argument = digits
			return l.checkedLibrary(b, value.Method, value)
		}
	case ir.TypedArrayNew:
		if !value.FromArray && !constantWithin(value.Source, 0, 9007199254740991) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Source})
			argument := b.read(b.parameters[0])
			size := b.read(b.declare("size", integer(argument)))
			l.libraryThrow(b, outside(size, 0, 9007199254740991), ir.Concat{Parts: []ir.Expression{text("Invalid typed array length: "), ir.NumberToString{Value: argument}}})
			value.Source = size
			return l.checkedLibrary(b, "typed_array", value)
		}
	case ir.TypedArraySet:
		arguments := append([]ir.Expression{value.Array}, value.Arguments...)
		b := l.libraryArrayBuilder(arguments)
		receiver, source := b.read(b.parameters[0]), b.read(b.parameters[1])
		offset := number(0)
		if len(b.parameters) > 2 {
			offset = b.read(b.declare("offset", integer(b.read(b.parameters[2]))))
		}
		invalid := binary(ir.Or, binary(ir.Less, offset, number(0)), binary(ir.Greater, binary(ir.Add, offset, ir.Length{Array: source}), ir.Length{Array: receiver}))
		l.libraryThrow(b, invalid, text("offset is out of bounds"))
		value.Array = receiver
		value.Arguments = []ir.Expression{source, fit(offset, ir.MaybeNumber)}
		b.body = append(b.body, ir.Evaluate{Value: value})
		return l.checkedLibrary(b, "typed_array_set", ir.Undefined{Of: ir.Object})
	case ir.ArrayFill:
		if value.Array == nil && !constantWithin(value.Length, 0, 4294967295) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Length, value.Value})
			length := b.read(b.parameters[0])
			valid := binary(ir.And, ir.Unary{Operator: ir.Not, Operand: outside(length, 0, 4294967295)}, binary(ir.Equal, length, integer(length)))
			l.libraryThrow(b, ir.Unary{Operator: ir.Not, Operand: valid}, text("Invalid array length"))
			value.Length = length
			value.Value = b.read(b.parameters[1])
			return l.checkedLibrary(b, "array_length", value)
		}
	case ir.ArrayFrom:
		if !constantWithin(value.Length, math.Inf(-1), 4294967295) {
			b := l.libraryArrayBuilder([]ir.Expression{value.Length, value.Callback})
			length := b.read(b.declare("length", integer(b.read(b.parameters[0]))))
			l.libraryThrow(b, binary(ir.Greater, length, number(4294967295)), text("Invalid array length"))
			value.Length = length
			value.Callback = b.read(b.parameters[1])
			return l.checkedLibrary(b, "array_from", value)
		}
	}
	return expression
}

func (l *lowering) libraryStatement(statement ir.Statement) ir.Statement {
	value, ok := statement.(ir.SetProperty)
	if !ok || !l.objectCanFreeze() || value.Name == "" || value.Name[0] == '#' {
		return statement
	}
	b := l.libraryArrayBuilder([]ir.Expression{value.Object, value.Value})
	receiver := b.read(b.parameters[0])
	condition := ir.ObjectCall{Method: "isFrozen", Arguments: []ir.Expression{fit(receiver, ir.Union)}, Returns: ir.Boolean}
	message := ir.StringConstant{Index: l.constant("Cannot assign to read only property '" + value.Name + "' of object '#<Object>'")}
	b.body = append(b.body, ir.If{Condition: condition, Then: []ir.Statement{ir.Throw{Value: ir.Box{Value: ir.MakeError{Message: message, Constructor: "TypeError"}}}}})
	value.Object = receiver
	value.Value = b.read(b.parameters[1])
	b.body = append(b.body, value)
	return ir.Evaluate{Value: l.checkedLibrary(b, "frozen_write", ir.Undefined{Of: ir.Object})}
}

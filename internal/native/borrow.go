package native

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// Borrowed reads (docs/memory.md, "Borrowed parameters", extended to reads).
//
// A reference read from somewhere a call could write (a global, a captured variable, a field) is
// retained the moment JavaScript reads it, because something later in the statement might write
// that place and free the value while it's still to be used. That's a retain and a release around
// every text.charCodeAt(position) of a global text: ten million pairs in the tokenizer benchmark.
//
// The count buys nothing when nothing can run between the read and its use. So a read is lent,
// taken without a count, when it's a direct operand of a consumer: an operation that uses the
// value and is done with it (its result is a number, a boolean, a new value, or one with its own
// count), whose every operand is pure, meaning it writes no variable, field, element or map, calls
// no code, and frees nothing. Then from the read to the consumer's operation, nothing but pure
// evaluation happens, and the place read still holds the value.
//
// One more thing makes that true in C: a consumer's result is often a C expression the emitter
// hands its parent, evaluated wherever the parent puts it, possibly after a sibling's call. A
// consumer that lent a read has its result evaluated where it stands (snapshotted), right after its
// operands, as JavaScript evaluates it.
//
// value is the wrapper the emitter calls for every expression; evaluate does the work.
func (e *emitter) value(expression ir.Expression) string {
	e.depth++
	defer func() { e.depth-- }()
	outerLendAt, outerLent := e.lendAt, e.lent
	// A read evaluated here may be lent when its parent is a consumer that said so.
	e.lendable = e.depth == outerLendAt
	if consumes(expression) {
		e.lendAt = e.depth + 1
	} else {
		e.lendAt = -1
	}
	e.lent, e.self = false, false
	result := e.evaluate(expression)
	// self is whether this expression's own value is lent (a read that took no count), which its
	// parent must hear; lent is whether one of its operands was.
	self, lent := e.self, e.lent
	e.lendAt, e.lent, e.self = outerLendAt, outerLent || self, false
	if lent && consumes(expression) && !expression.Type().IsReference() && expression.Type() != 0 {
		result = e.snapshot(expression.Type(), result)
	}
	return result
}

// lendable reports whether a read of a value of this type can be lent: a reference, held as a
// pointer. A union holds its reference inside a box, and stays counted.
func lendable(valueType ir.Type) bool {
	return valueType.IsReference() && valueType != ir.Union
}

// consumes reports whether an expression is a consumer whose operands it can lend: pure, done with
// its operands when it's evaluated (not handing one on as its own value, as a conditional or ??
// does), and with every operand pure.
//
// Every pure operation whose value is one of its operands, as it is (a narrowing, a cast, a box),
// must be on the list below: on pure's list without being here, it lends the read it hands on, and
// its parent keeps that uncounted pointer while a later operand frees it (ir.Defined, once pure,
// did: borrow_defined_lent.a). TestPassThroughsAreNotConsumers holds every pure kind to a decision.
func consumes(expression ir.Expression) bool {
	switch expression.(type) {
	case ir.Read, ir.Conditional, ir.Coalesce, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.MaybeOf, ir.Defined, ir.Undefined,
		ir.NumberConstant, ir.BooleanConstant, ir.StringConstant:
		return false
	}
	return pure(expression)
}

// pure reports whether evaluating an expression can't write a variable, a field, an element or a
// map, call code, or free anything a statement holds: a fixed list of the IR's operations, whatever
// they're made of being pure too. Anything not on the list, an operation added later included, is
// not pure.
func pure(expression ir.Expression) bool {
	if !pureKind(expression) {
		return false
	}
	operands := true
	eachOperand(expression, func(operand ir.Expression) {
		if !pure(operand) {
			operands = false
		}
	})
	return operands
}

// pureKind reports whether an expression's own operation is on pure's list, whatever its operands.
func pureKind(expression ir.Expression) bool {
	switch expression.(type) {
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Read, ir.Undefined,
		ir.Unary, ir.Binary, ir.NumberToString, ir.BooleanToString, ir.Concat, ir.Length, ir.StringLength,
		ir.CharCodeAt, ir.StringIndex, ir.ArrayIndex, ir.Property, ir.MapGet, ir.MapHas, ir.MapSize, ir.HasOwn,
		ir.IsUndefined, ir.Unwrap, ir.MaybeOf, ir.Box, ir.Narrow, ir.TypeOf, ir.Conditional, ir.Coalesce,
		ir.MathCall, ir.NumberCall, ir.ToFixed, ir.NumberFormat, ir.Trim, ir.StringCall, ir.CodePoints,
		ir.ArraySearch, ir.CheckedCast, ir.UnionToString, ir.MaybeToString, ir.Defined, ir.InstanceOf:
	default:
		return false
	}
	return true
}

// eachOperand visits the expressions an IR node holds directly.
func eachOperand(expression ir.Expression, visit func(ir.Expression)) {
	var each func(value reflect.Value)
	each = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if operand, ok := value.Interface().(ir.Expression); ok {
				visit(operand)
				return
			}
			each(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				each(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				each(value.Index(index))
			}
		}
	}
	each(reflect.ValueOf(expression))
}

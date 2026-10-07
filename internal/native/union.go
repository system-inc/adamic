package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// converted is value, of type from, as a value of type to: a number or a boolean made a Maybe pair
// or put in a union's box, a reference made a union. fresh says it's a new reference the caller now
// owns (a boxed number); otherwise it's borrowed as value was.
func converted(from ir.Type, to ir.Type, value string) (string, bool) {
	switch {
	case from == to:
		return value, false
	case to.IsMaybe() && from == to.Present():
		return maybe(to, value), false
	case to == ir.Weak:
		return fmt.Sprintf("adamic_weak_of(%s)", value), true
	case to != ir.Union:
		panic(fmt.Sprintf("native: no conversion from %d to %d", from, to))
	case from == ir.Number:
		return fmt.Sprintf("adamic_box_number(%s)", value), true
	case from == ir.Boolean:
		return fmt.Sprintf("((%s) ? &adamic_box_true.heap : &adamic_box_false.heap)", value), false
	case from == ir.MaybeNumber:
		return fmt.Sprintf("((%s).present ? adamic_box_number((%s).number) : NULL)", value, value), true
	case from == ir.MaybeBoolean:
		return fmt.Sprintf("(!(%s).present ? NULL : (%s).boolean ? &adamic_box_true.heap : &adamic_box_false.heap)", value, value), false
	}
	return fmt.Sprintf("((adamic_heap *)(%s))", value), false
}

// box emits a value where a union goes.
func (e *emitter) box(value ir.Expression) string {
	boxed, fresh := converted(value.Type(), ir.Union, e.value(value))
	if fresh {
		return e.own(ir.Union, boxed)
	}
	return boxed
}

// narrow emits a union the checker proved to be one member, as that member. A number or a boolean is
// read out of its box now, as JavaScript reads the variable; a reference is the union's, borrowed.
func (e *emitter) narrow(narrow ir.Narrow) string {
	value := e.value(narrow.Value)
	switch narrow.To {
	case ir.Number:
		return e.snapshot(ir.Number, fmt.Sprintf("((const adamic_number_box *)%s)->number", value))
	case ir.Boolean:
		return e.snapshot(ir.Boolean, fmt.Sprintf("((const adamic_boolean_box *)%s)->boolean", value))
	case ir.MaybeNumber:
		return e.snapshot(ir.MaybeNumber, fmt.Sprintf("%s == NULL ? %s : %s", value, zero(ir.MaybeNumber), maybe(ir.MaybeNumber, fmt.Sprintf("((const adamic_number_box *)%s)->number", value))))
	case ir.MaybeBoolean:
		return e.snapshot(ir.MaybeBoolean, fmt.Sprintf("%s == NULL ? %s : %s", value, zero(ir.MaybeBoolean), maybe(ir.MaybeBoolean, fmt.Sprintf("((const adamic_boolean_box *)%s)->boolean", value))))
	}
	return fmt.Sprintf("((%s)%s)", cType(narrow.To), value)
}

// typeOf is the single native emission path for typeof, including comparisons and switches.
// Scalars have no heap header; dynamic references use the runtime's one classifier.
func (e *emitter) typeOf(observation ir.TypeOf) string {
	value := observation.Value
	operand, null := e.typeOfReference(observation)
	named := func(name string) string { return "&adamic_typeof_" + name }
	if _, intrinsic := value.(ir.LibraryGlobal); intrinsic {
		// Built-in identities are opaque headers, so JSON has no class metadata to inspect.
		if value.Type() == ir.Object {
			return named("object")
		}
		return named("function")
	}
	switch value.Type() {
	case ir.Number:
		return named("number")
	case ir.Boolean:
		return named("boolean")
	case ir.MaybeNumber, ir.MaybeBoolean:
		return fmt.Sprintf("((%s).present ? %s : %s)", operand, named(typeName(value.Type().Present())), named("undefined"))
	}
	return fmt.Sprintf("adamic_union_typeof((const adamic_heap *)(%s), %s)", operand, null)
}

// typeName is the name typeof gives a number or a boolean.
func typeName(valueType ir.Type) string {
	if valueType == ir.Boolean {
		return "boolean"
	}
	return "number"
}

// Functions and declarations moved unchanged from emit.go.
package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

func cType(valueType ir.Type) string {
	switch valueType {
	case ir.Number:
		return "double"
	case ir.Boolean:
		return "bool"
	case ir.Object:
		return "adamic_object *"
	case ir.Uint8Array, ir.Int32Array, ir.Float64Array:
		return "adamic_typed_array *"
	case ir.Array:
		return "adamic_array *"
	case ir.Map:
		return "adamic_map *"
	case ir.Closure:
		return "adamic_closure *"
	case ir.MaybeNumber:
		return "adamic_maybe_number"
	case ir.MaybeBoolean:
		return "adamic_maybe_boolean"
	case ir.Union:
		return "adamic_heap *"
	case ir.Weak:
		return "adamic_weak *"
	}
	return "adamic_string *"
}

// member is the adamic_value member that holds a value of the type.
func member(valueType ir.Type) string {
	switch valueType {
	case ir.Number, ir.MaybeNumber:
		return "number"
	case ir.Boolean:
		return "boolean"
	case ir.MaybeBoolean:
		return "maybe_boolean"
	}
	return "reference"
}

// zero is a declared local's value before its first assignment. The checker proves no read comes
// first, so it's never seen; a string still needs a real one, since assigning releases the old.
func zero(valueType ir.Type) string {
	switch valueType {
	case ir.Number:
		return "0.0"
	case ir.Boolean:
		return "false"
	case ir.String:
		return "&adamic_string_empty"
	case ir.MaybeNumber:
		return "(adamic_maybe_number){false, 0.0}"
	case ir.MaybeBoolean:
		return "(adamic_maybe_boolean){false, false}"
	}
	return "NULL"
}

// cNumber is a double as a C literal that reads back as exactly the same bits: hexadecimal floating
// point, which C99 reads exactly, unlike decimal.
func cNumber(value float64) string {
	switch {
	case value != value:
		return "NAN"
	case value > 1.7976931348623157e308:
		return "HUGE_VAL"
	case value < -1.7976931348623157e308:
		return "(-HUGE_VAL)"
	}
	return "(" + strconv.FormatFloat(value, 'x', -1, 64) + ")"
}

// absent is an omitted argument, distinct from an uninitialized local's placeholder.
// References use NULL; scalar optionals carry their explicit absence bit.
func absent(valueType ir.Type) string {
	if valueType.IsMaybe() {
		return zero(valueType)
	}
	if valueType.IsReference() {
		return "NULL"
	}
	panic("native: omitted argument has no undefined representation")
}

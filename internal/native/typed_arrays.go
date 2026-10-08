package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func typedArrayKind(kind ir.Type) string {
	switch kind {
	case ir.Uint16Array:
		return "adamic_typed_array_uint16"
	case ir.Uint8Array:
		return "adamic_typed_array_uint8"
	case ir.Int32Array:
		return "adamic_typed_array_int32"
	case ir.Float64Array:
		return "adamic_typed_array_float64"
	}
	panic(fmt.Sprintf("native: no typed array kind for %d", kind))
}

// Evaluate optional arguments once, preserving undefined as absent for runtime's flags.
func (e *emitter) typedArrayOffset(arguments []ir.Expression, index int) (string, string) {
	if index >= len(arguments) {
		return "0.0", "false"
	}
	argument := e.snapshot(ir.MaybeNumber, e.value(arguments[index]))
	return argument + ".number", argument + ".present"
}

func (e *emitter) typedArrayValue(expression ir.Expression) string {
	switch v := expression.(type) {
	case ir.TypedArrayNew:
		source := e.value(v.Source)
		function := "adamic_typed_array_new"
		if v.FromArray {
			function = "adamic_typed_array_from_numbers"
		}
		return e.own(v.Of, fmt.Sprintf("%s(%s, %s)", function, typedArrayKind(v.Of), source))
	case ir.TypedArraySort:
		array := e.value(v.Array)
		e.line("adamic_typed_array_sort(%s);", array)
		return array
	case ir.TypedArrayFill:
		array := e.value(v.Array)
		value := e.snapshot(ir.Number, e.value(v.Arguments[0]))
		start, hasStart := e.typedArrayOffset(v.Arguments, 1)
		end, hasEnd := e.typedArrayOffset(v.Arguments, 2)
		e.line("adamic_typed_array_fill(%s, %s, %s, %s, %s, %s);", array, value, start, end, hasStart, hasEnd)
		return array
	case ir.TypedArraySet:
		array := e.value(v.Array)
		source := e.value(v.Arguments[0])
		offset, hasOffset := e.typedArrayOffset(v.Arguments, 1)
		e.line("adamic_typed_array_set_from(%s, %s, %s, %s);", array, source, offset, hasOffset)
		return "0"
	case ir.TypedArraySubarray:
		array := e.value(v.Array)
		start, present := e.typedArrayOffset(v.Arguments, 0)
		end, hasEnd := e.typedArrayOffset(v.Arguments, 1)
		start = fmt.Sprintf("(%s ? %s : 0.0)", present, start)
		return e.own(v.Type(), fmt.Sprintf("adamic_typed_array_subarray(%s, %s, %s, %s)", array, start, end, hasEnd))
	}
	panic(fmt.Sprintf("native: no typed array emission for %T", expression))
}

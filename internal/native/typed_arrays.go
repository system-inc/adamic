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
func (nativeEmitter *emitter) typedArrayOffset(arguments []ir.Expression, index int) (string, string) {
	if index >= len(arguments) {
		return "0.0", "false"
	}
	argument := nativeEmitter.snapshot(ir.MaybeNumber, nativeEmitter.value(arguments[index]))
	return argument + ".number", argument + ".present"
}

func (nativeEmitter *emitter) typedArrayValue(expression ir.Expression) string {
	switch typed := expression.(type) {
	case ir.TypedArrayNew:
		source := nativeEmitter.value(typed.Source)
		function := "adamic_typed_array_new"
		if typed.FromArray {
			function = "adamic_typed_array_from_numbers"
		}
		return nativeEmitter.own(typed.Of, fmt.Sprintf("%s(%s, %s)", function, typedArrayKind(typed.Of), source))
	case ir.TypedArraySort:
		array := nativeEmitter.value(typed.Array)
		nativeEmitter.line("adamic_typed_array_sort(%s);", array)
		return array
	case ir.TypedArrayFill:
		array := nativeEmitter.value(typed.Array)
		value := nativeEmitter.snapshot(ir.Number, nativeEmitter.value(typed.Arguments[0]))
		start, hasStart := nativeEmitter.typedArrayOffset(typed.Arguments, 1)
		end, hasEnd := nativeEmitter.typedArrayOffset(typed.Arguments, 2)
		nativeEmitter.line(
			"adamic_typed_array_fill(%s, %s, %s, %s, %s, %s);",
			array,
			value,
			start,
			end,
			hasStart,
			hasEnd,
		)
		return array
	case ir.TypedArraySet:
		array := nativeEmitter.value(typed.Array)
		source := nativeEmitter.value(typed.Arguments[0])
		offset, hasOffset := nativeEmitter.typedArrayOffset(typed.Arguments, 1)
		nativeEmitter.line("adamic_typed_array_set_from(%s, %s, %s, %s);", array, source, offset, hasOffset)
		return "0"
	case ir.TypedArraySubarray:
		array := nativeEmitter.value(typed.Array)
		start, present := nativeEmitter.typedArrayOffset(typed.Arguments, 0)
		end, hasEnd := nativeEmitter.typedArrayOffset(typed.Arguments, 1)
		start = fmt.Sprintf("(%s ? %s : 0.0)", present, start)
		return nativeEmitter.own(
			typed.Type(),
			fmt.Sprintf("adamic_typed_array_subarray(%s, %s, %s, %s)", array, start, end, hasEnd),
		)
	}
	panic(fmt.Sprintf("native: no typed array emission for %T", expression))
}

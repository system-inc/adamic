package native

import (
	"github.com/system-inc/adamic/internal/ir"
)

// Scalar views can cross a function boundary without changing the array's slots.
// The array's reference flag identifies boxed storage; unbox only after checking its brand.
func (e *emitter) arrayReadSlot(array, slot string, of ir.Type) string {
	if of != ir.Number && of != ir.Boolean {
		return slot
	}
	result := e.temporary()
	e.line("adamic_value %s = %s;", result, slot)
	e.line("if (%s->references) {", array)
	kind, box, field := "number", "adamic_number_box", "number"
	if of == ir.Boolean {
		kind, box, field = "boolean", "adamic_boolean_box", "boolean"
	}
	e.line("\tif (%s.reference == NULL || ((const adamic_heap *)%s.reference)->kind != adamic_kind_%s) {", slot, slot, kind)
	e.line("\t\tstatic const char message[] = \"array element does not match its narrowed type\";")
	e.line("\t\tadamic_panic(message, sizeof message - 1);")
	e.line("\t}")
	e.line("\t%s.%s = ((const %s *)%s.reference)->%s;", result, field, box, slot, field)
	e.line("}")
	return result
}

func (e *emitter) arrayCallbackSlot(array, slot string, callback ir.Expression, kind, position int, fallback ir.Type) string {
	layout := e.program.ClosureArgumentLayout(ir.CallClosure{Closure: callback, FunctionType: kind})
	of := fallback
	if position < len(layout.Fixed) {
		of = layout.Fixed[position]
	}
	return e.arrayReadSlot(array, slot, of)
}

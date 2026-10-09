package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Keep metadata at allocation, not at assertion. Wrapping dispatch here also
// covers arrays produced inside callees, irrespective of source lowering order.
func (e *emitter) evaluate(expression ir.Expression) string {
	value := e.evaluateWithoutViewArrays(expression)
	if !ir.HasArrayViews(e.program) {
		return value
	}
	storage := ir.Type(0)
	switch expression := expression.(type) {
	case ir.ArrayLiteral:
		storage = expression.Element
	case ir.ArrayMap:
		storage = expression.Result
	case ir.ArrayFrom:
		storage = expression.Element
	case ir.ArrayFill:
		if expression.Array == nil {
			storage = expression.Element
		}
	case ir.ArrayVisit:
		if expression.Method == "filter" {
			storage = expression.Element
		}
	case ir.CodePoints:
		storage = ir.String
	case ir.MapEntries:
		storage = ir.Object
	case ir.StringCall:
		if expression.Method == "split" {
			storage = ir.String
		}
	}
	if storage != 0 {
		e.line("adamic_array_view_storage(%s, %d);", value, storage)
	}
	return value
}

func (e *emitter) emitViewArrayRead(read ir.ArrayIndex, array, index string) string {
	snapshot, slot := e.temporary(), e.temporary()
	e.line("adamic_value %s;", snapshot)
	e.line("adamic_value *%s = adamic_view_array_at(%s, %s, %t, %t, %d, %s, %s, &%s);", slot, array, index, read.Relative, read.UndefinedAllowed, read.Element, cString(read.ViewType), cString(read.View), snapshot)
	if read.Required {
		e.line("if (%s == NULL) adamic_view_array_missing(%s, %s);", slot, cString(read.View), cString(read.ViewType))
	}
	if len(read.ViewAllowed) != 0 {
		tests := []string{}
		for _, literal := range read.ViewAllowed {
			switch literal.Of {
			case ir.Number:
				tests = append(tests, fmt.Sprintf("%s.number == %s", snapshot, cNumber(literal.Number)))
			case ir.Boolean:
				tests = append(tests, fmt.Sprintf("%s.boolean == %t", snapshot, literal.Boolean))
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				tests = append(tests, fmt.Sprintf("adamic_string_equal(%s.reference, &%s)", snapshot, name))
			}
		}
		e.line("if (%s != NULL && !(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", slot, strings.Join(tests, " || "), cString(read.View), cString(read.ViewType), read.Element, snapshot)
	}
	return slot
}

func (e *emitter) viewArrayElementSlot(read ir.ArrayViewRead, array, index string) string {
	slot := e.emitViewArrayRead(read.Index(), array, index)
	value := e.temporary()
	fallback := "(adamic_value){.reference = NULL}"
	if read.Element == ir.MaybeNumber {
		fallback = "(adamic_value){.number = " + slotted(ir.MaybeNumber, zero(ir.MaybeNumber)) + "}"
	}
	e.line("adamic_value %s = %s == NULL ? %s : *%s;", value, slot, fallback, slot)
	return value
}

func (e *emitter) viewArrayMutation(array string, element ir.Type) {
	if ir.HasArrayViews(e.program) {
		e.line("adamic_view_array_storage_check(%s, %d, %s);", array, element, cString("<array write>"))
	}
}

func (e *emitter) emitViewArrayPop(pop ir.ArrayPop) string {
	array := e.snapshot(ir.Array, e.value(pop.Array))
	slot := e.emitViewArrayRead(pop.ViewRead.Index(), array, fmt.Sprintf("(double)%s->length - 1", array))
	var result string
	if pop.Type().IsMaybe() {
		result = e.snapshot(pop.Type(), maybeSlot(pop.Element, slot))
	} else {
		result = e.own(pop.Element, fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s->reference)", slot, cType(pop.Element), slot))
	}
	e.line("if (%s->length != 0) {", array)
	e.line("\t%s->length--;", array)
	e.line("\tif (%s->references) adamic_release(%s->elements[%s->length].reference);", array, array, array)
	e.line("}")
	return result
}

func (e *emitter) viewArrayJoin(join ir.ArrayJoin) string {
	array := e.own(ir.Array, "adamic_retain("+e.value(join.Array)+")")
	separator := e.snapshot(ir.String, e.value(join.Separator))
	copy := e.own(ir.Array, fmt.Sprintf("adamic_array_new(0,%t)", join.Element.IsReference()))
	index := e.temporary()
	e.line("for(size_t %s=0;%s<%s->length;%s++){", index, index, array, index)
	e.indent++
	value := e.viewArrayElementSlot(join.ViewRead, array, "(double)"+index)
	if join.Element.IsReference() {
		e.line("adamic_retain(%s.reference);", value)
	}
	e.line("adamic_array_push(%s,%s);", copy, value)
	e.indent--
	e.line("}")
	return e.own(ir.String, fmt.Sprintf("adamic_array_join(%s,%s,%s)", copy, separator, joinKind(join.Element)))
}

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Keep metadata at allocation, not at assertion. Wrapping dispatch here also
// covers arrays produced inside callees, irrespective of source lowering order.
func (e *emitter) evaluate(expression ir.Expression) string {
	if record, ok := expression.(ir.ArrayRecord); ok {
		return e.emitViewArrayRecord(record)
	}
	if properties, ok := expression.(ir.ArrayProperties); ok {
		return e.emitViewArrayProperties(properties)
	}
	if sort, ok := expression.(ir.ArraySort); ok && sort.OptionalComparator {
		return e.emitViewOptionalArraySort(sort)
	}
	if join, ok := expression.(ir.ArrayJoin); ok && join.Stringify {
		return e.emitViewArrayString(join)
	}
	value := e.evaluateWithoutViewArrays(expression)
	if !ir.HasArrayViews(e.program) {
		return value
	}
	e.viewArraySourceCertificate(expression, value)
	storage := ir.Type(0)
	switch expression := expression.(type) {
	case ir.ArrayHoles:
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
	if read.Element == ir.Object && read.ViewContract != 0 {
		e.line("if (%s != NULL) {", slot)
		e.indent++
		e.viewObjectUnion(ir.Property{View: read.View, ViewContract: read.ViewContract}, snapshot+".reference")
		e.indent--
		e.line("}")
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

func (e *emitter) viewArrayMutation(array string, element ir.Type, value string) {
	if ir.HasArrayViews(e.program) && element == ir.Object {
		e.viewArrayReferenceWrite(array, value)
		return
	}
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
	e.line("adamic_view_array_pop_commit(%s);", array)
	return result
}

func (e *emitter) emitViewArrayString(join ir.ArrayJoin) string {
	array := e.value(join.Array)
	separator := e.value(join.Separator)
	allowed := e.viewArrayStringLiterals(join.ViewRead.ViewAllowed)
	return e.own(ir.String, fmt.Sprintf("adamic_view_array_string(%s, %s, %t, %d, %t, %s, %s, %d, %s)", array, separator, join.ViewRead.View != "", join.Element, join.ViewRead.UndefinedAllowed, cString(join.ViewRead.ViewType), cString(join.ViewRead.View), len(join.ViewRead.ViewAllowed), allowed))
}

func (e *emitter) emitViewOptionalArraySort(sort ir.ArraySort) string {
	array := e.value(sort.Array)
	callback := e.value(sort.Callback)
	sortFunction := "adamic_array_sort"
	if sort.Element == ir.MaybeNumber {
		sortFunction = "adamic_array_sort_undefined_last"
	}
	e.line("if (%s != NULL) {", callback)
	e.indent++
	e.line("%s(%s, adamic_compare_closure, %s);", sortFunction, array, callback)
	e.closureThrown()
	e.indent--
	e.line("} else {")
	e.indent++
	e.line("%s(%s, adamic_view_array_default_compare, (void *)(uintptr_t)%d);", sortFunction, array, sort.Element)
	e.indent--
	e.line("}")
	return array
}

func (e *emitter) viewArrayStringLiterals(literals []ir.ViewLiteral) string {
	if len(literals) == 0 {
		return "NULL"
	}
	values := []string{}
	for _, literal := range literals {
		switch literal.Of {
		case ir.Number:
			values = append(values, "{.number = "+cNumber(literal.Number)+"}")
		case ir.Boolean:
			values = append(values, fmt.Sprintf("{.boolean = %t}", literal.Boolean))
		case ir.String:
			name := e.temporary()
			e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
			values = append(values, "{.reference = &"+name+"}")
		}
	}
	return "(adamic_value[]){" + strings.Join(values, ", ") + "}"
}

// join consumes all positions now; the original array is never scanned by a
// field read. A temporary string array keeps physical source storage unchanged.
func (e *emitter) emitViewArrayJoin(join ir.ArrayJoin) string {
	array := e.snapshot(ir.Array, e.value(join.Array))
	separator := e.snapshot(ir.String, e.value(join.Separator))
	strings := e.own(ir.Array, "adamic_array_new(0, true)")
	index := e.temporary()
	e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, array, index)
	e.indent++
	read := join.ViewRead.Index()
	read.Required = false
	slot := e.emitViewArrayRead(read, array, "(double)"+index)
	empty, yes, no := e.temporary(), e.temporary(), e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"\");", empty), fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"true\");", yes), fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"false\");", no))
	text := e.temporary()
	e.line("adamic_string *%s = &%s;", text, empty)
	e.line("if (%s != NULL) {", slot)
	e.indent++
	switch join.Element {
	case ir.Number:
		e.line("%s = adamic_string_from_number(%s->number);", text, slot)
	case ir.MaybeNumber:
		value := e.temporary()
		e.line("adamic_maybe_number %s = adamic_maybe_number_unpack(%s->number);", value, slot)
		e.line("if (%s.present) %s = adamic_string_from_number(%s.number);", value, text, value)
	case ir.Boolean:
		e.line("%s = %s->boolean ? &%s : &%s;", text, slot, yes, no)
	case ir.String:
		e.line("%s = adamic_retain(%s->reference);", text, slot)
	}
	e.indent--
	e.line("}")
	e.line("adamic_array_push(%s, (adamic_value){.reference = %s});", strings, text)
	e.indent--
	e.line("}")
	return e.own(ir.String, fmt.Sprintf("adamic_array_join(%s, %s, adamic_join_strings)", strings, separator))
}

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Keep metadata at allocation, not at assertion. Wrapping dispatch here also
// covers arrays produced inside callees, irrespective of source lowering order.
func (e *emitter) evaluate(expression ir.Expression) string {
	if read, ok := expression.(ir.ArrayIndex); ok && !read.TupleUnion && read.Element == ir.Union && read.View != "" {
		if _, primitive := ir.PrimitiveViewMembers(e.program, read.ViewContract); primitive {
			return e.emitPrimitiveArrayIndex(read)
		}
	}
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
	if literal, ok := expression.(ir.ObjectLiteral); ok {
		e.line("%s->tuple = %t;", value, literal.Tuple)
	}
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

func (e *emitter) viewArrayReadOwner(element ir.Type) string {
	if element != ir.Union {
		return "NULL"
	}
	return e.own(ir.Array, "adamic_array_new(0, true)")
}

func (e *emitter) emitViewArrayRead(read ir.ArrayIndex, array, index string) string {
	if read.TupleUnion {
		return e.emitTupleArrayUnionIndex(read, array, index)
	}
	return e.emitViewArrayReadWithOwner(read, array, index, e.viewArrayReadOwner(read.Element))
}

func (e *emitter) emitViewArrayReadWithOwner(read ir.ArrayIndex, array, index, owner string) string {
	return e.emitViewArrayReadChecked(read, array, index, owner, true)
}

// Producer scans use the physical adapter, then their own recursive certificate.
func (e *emitter) emitViewArrayReadChecked(read ir.ArrayIndex, array, index, owner string, logical bool) string {
	snapshot, slot := e.temporary(), e.temporary()
	wanted := fmt.Sprint(read.Element)
	nominal, _, _ := mapNominalContract(e.program, read.ViewContract)
	if read.Element == ir.Union && nominal.NominalClass != 0 {
		wanted = "adamic_view_array_nominal_union"
	}
	e.line("adamic_value %s;", snapshot)
	e.line("adamic_value *%s = adamic_view_array_at(%s, %s, %t, %t, %s, %s, %s, &%s, %s);", slot, array, index, read.Relative, read.UndefinedAllowed, wanted, cString(read.ViewType), cString(read.View), snapshot, owner)
	if read.Required {
		e.line("if (%s == NULL) adamic_view_array_missing(%s, %s);", slot, cString(read.View), cString(read.ViewType))
	}
	if len(read.ViewAllowed) != 0 && read.Element != ir.Union {
		tests := []string{}
		for _, literal := range read.ViewAllowed {
			switch literal.Of {
			case ir.Number:
				tests = append(tests, fmt.Sprintf("%s.number == %s", snapshot, cNumber(literal.Number)))
			case ir.Boolean:
				actual := snapshot + ".boolean"
				if read.Element == ir.MaybeBoolean {
					actual = "adamic_maybe_boolean_unpack(" + snapshot + ".maybe_boolean).boolean"
				}
				tests = append(tests, fmt.Sprintf("%s == %t", actual, literal.Boolean))
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				tests = append(tests, fmt.Sprintf("adamic_string_equal(%s.reference, &%s)", snapshot, name))
			}
		}
		e.line("if (%s != NULL && !(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", slot, strings.Join(tests, " || "), cString(read.View), cString(read.ViewType), read.Element, snapshot)
	}
	if logical && read.Element == ir.Object && read.ViewContract != 0 {
		e.line("if (%s != NULL) {", slot)
		e.indent++
		e.viewObjectUnion(ir.Property{View: read.View, ViewContract: read.ViewContract}, snapshot+".reference")
		e.nominalViewRead(read.ViewContract, snapshot+".reference", read.View, false)
		e.indent--
		e.line("}")
	}
	if logical && read.Element == ir.Union {
		e.line("if (%s != NULL) {", slot)
		if nominal.NominalClass != 0 {
			e.nominalViewRead(read.ViewContract, snapshot+".reference", read.View, false)
		} else {
			e.arrayPrimitiveUnionRead(read.ViewContract, snapshot+".reference", read.View, read.ViewType)
		}
		e.line("}")
	}
	return slot
}

func (e *emitter) viewArrayElementSlot(read ir.ArrayViewRead, array, index, owner string) string {
	slot := ""
	if read.TupleUnion {
		slot = e.emitTupleArrayUnionSlot(read.Index(), array, index, false)
	} else {
		slot = e.emitViewArrayReadWithOwner(read.Index(), array, index, owner)
	}
	value := e.temporary()
	fallback := "(adamic_value){.reference = NULL}"
	if read.Element == ir.MaybeNumber {
		fallback = "(adamic_value){.number = " + slotted(ir.MaybeNumber, zero(ir.MaybeNumber)) + "}"
	}
	if read.Element == ir.MaybeBoolean {
		fallback = "(adamic_value){.maybe_boolean = 2}"
	}
	e.line("adamic_value %s = %s == NULL ? %s : *%s;", value, slot, fallback, slot)
	return value
}

func (e *emitter) viewArrayMutation(array string, element ir.Type, value string) {
	if ir.HasArrayViews(e.program) && (element == ir.Object || element == ir.Union) {
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
	array := e.snapshot(ir.Array, e.value(join.Array))
	separator := e.value(join.Separator)
	if join.Element == ir.Union {
		result := e.own(ir.String, "&adamic_string_undefined")
		e.line("if (%s != NULL) {", array)
		outer := e.owned
		e.owned = nil
		text := e.emitViewArrayJoinSource(join, array, separator)
		e.line("%s = adamic_retain(%s);", result, text)
		e.end()
		e.owned = outer
		e.line("}")
		return result
	}
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
	e.line("%s(%s, %s, %s);", sortFunction, array, e.viewCallableBoxedComparator(sort.Element), callback)
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
	return e.emitViewArrayJoinSource(join, array, separator)
}

func (e *emitter) emitViewArrayJoinSource(join ir.ArrayJoin, array, separator string) string {
	strings := e.own(ir.Array, "adamic_array_new(0, true)")
	owner := e.viewArrayReadOwner(join.Element)
	index := e.temporary()
	e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, array, index)
	e.indent++
	read := join.ViewRead.Index()
	read.Required = false
	slot := e.emitViewArrayReadWithOwner(read, array, "(double)"+index, owner)
	empty, yes, no := e.temporary(), e.temporary(), e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"\");", empty), fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"true\");", yes), fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(\"false\");", no))
	text := e.temporary()
	e.line("adamic_string *%s = &%s;", text, empty)
	e.line("if (%s != NULL) {", slot)
	e.indent++
	switch join.Element {
	case ir.Union:
		e.line("if (%s->reference != NULL && %s->reference != &adamic_null) %s = adamic_union_to_string(%s->reference);", slot, slot, text, slot)
	case ir.Number:
		e.line("%s = adamic_string_from_number(%s->number);", text, slot)
	case ir.MaybeNumber:
		value := e.temporary()
		e.line("adamic_maybe_number %s = adamic_maybe_number_unpack(%s->number);", value, slot)
		e.line("if (%s.present) %s = adamic_string_from_number(%s.number);", value, text, value)
	case ir.MaybeBoolean:
		e.line("%s = adamic_maybe_boolean_unpack(%s->maybe_boolean).boolean ? &%s : &%s;", text, slot, yes, no)
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

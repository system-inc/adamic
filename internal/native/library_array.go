package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) libraryArraySearch(search ir.ArraySearch) string {
	array := e.value(search.Array)
	value := e.value(search.Value)
	from := "0.0"
	if search.From != nil {
		from = e.value(search.From)
	}
	element := search.Element
	if element == ir.Union {
		value = e.snapshot(ir.Union, value)
		// A temporary match vector preserves strict equality, SameValueZero and the runtime
		// search's starting-index rules without comparing boxes by their addresses.
		copy := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%s->length, false)", array))
		index, item := e.temporary(), e.temporary()
		e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, array, index)
		e.line("\tadamic_heap *%s = %s->elements[%s].reference;", item, array, index)
		equal := fmt.Sprintf("adamic_union_equal(%s, %s)", item, value)
		if search.Includes {
			equal += fmt.Sprintf(" || (%s != NULL && %s != NULL && %s->kind == adamic_kind_number && %s->kind == adamic_kind_number && isnan(((adamic_number_box *)%s)->number) && isnan(((adamic_number_box *)%s)->number))", item, value, item, value, item, value)
		}
		e.line("\tadamic_array_push(%s, (adamic_value){.number = (%s) ? 1.0 : 0.0});", copy, equal)
		e.line("}")
		array, value, element = copy, "1.0", ir.Number
	}
	if element == ir.MaybeBoolean {
		// Search compares the three tagged values, including present undefined. Use numerical
		// tags in a temporary dense array so no uninitialized union padding enters equality.
		copy := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%s->length, false)", array))
		index := e.temporary()
		e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, array, index)
		e.line("\tadamic_array_push(%s, (adamic_value){.number = %s->elements[%s].maybe_boolean});", copy, array, index)
		e.line("}")
		array, value, element = copy, fmt.Sprintf("(double)adamic_maybe_boolean_pack(%s)", value), ir.Number
	}
	found := fmt.Sprintf("adamic_array_search_from(%s, %s, %s, %t, %s, %t, %t)", array, borrowed(element, value), equality(element), search.Includes, from, search.From != nil, search.Last)
	if search.Includes {
		return e.snapshot(ir.Boolean, found+" != -1")
	}
	return e.snapshot(ir.Number, found)
}

func (e *emitter) libraryArrayJoin(join ir.ArrayJoin) string {
	array := e.value(join.Array)
	separator := e.value(join.Separator)
	return e.own(ir.String, fmt.Sprintf("adamic_array_join_nested(%s, %s, %s, %d)", array, separator, joinKind(join.Element), join.Depth))
}

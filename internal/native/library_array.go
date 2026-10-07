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
	function := "adamic_array_search_from"
	if e.hasArrayHoles() {
		function = "adamic_array_holes_search_from"
	}
	found := fmt.Sprintf("%s(%s, %s, %s, %t, %s, %t, %t)", function, array, borrowed(search.Element, value), equality(search.Element), search.Includes, from, search.From != nil, search.Last)
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

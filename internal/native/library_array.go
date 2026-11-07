package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) libraryArraySearch(search ir.ArraySearch) string {
	array := nativeEmitter.value(search.Array)
	value := nativeEmitter.value(search.Value)
	from := "0.0"
	if search.From != nil {
		from = nativeEmitter.value(search.From)
	}
	found := fmt.Sprintf(
		"adamic_array_search_from(%s, %s, %s, %t, %s, %t, %t)",
		array,
		borrowed(search.Element, value),
		equality(search.Element),
		search.Includes,
		from,
		search.From != nil,
		search.Last,
	)
	if search.Includes {
		return nativeEmitter.snapshot(ir.Boolean, found+" != -1")
	}
	return nativeEmitter.snapshot(ir.Number, found)
}

func (nativeEmitter *emitter) libraryArrayJoin(join ir.ArrayJoin) string {
	array := nativeEmitter.value(join.Array)
	separator := nativeEmitter.value(join.Separator)
	return nativeEmitter.own(
		ir.String,
		fmt.Sprintf("adamic_array_join_nested(%s, %s, %s, %d)", array, separator, joinKind(join.Element), join.Depth),
	)
}

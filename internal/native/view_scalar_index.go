package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) scalarCastIndexSnapshot(read ir.ArrayIndex) string {
	e.declarations = append(e.declarations, "#include \"view_array_primitives.h\"")
	array := e.own(ir.Array, "adamic_retain("+e.value(read.Array)+")")
	index := e.value(read.Index)
	snapshot := e.temporary()
	e.line("adamic_view_union_value %s = adamic_view_primitive_array_snapshot(%s, %s, false, %s, %s);", snapshot, array, index, cString(read.View), cString(read.ViewType))
	return e.own(ir.Union, fmt.Sprintf("adamic_view_primitive_array_box(%s)", snapshot))
}

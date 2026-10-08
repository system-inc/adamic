package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitPrimitiveArrayIndex(read ir.ArrayIndex) string {
	contracts, ok := ir.PrimitiveViewMembers(e.program, read.ViewContract)
	if !ok {
		panic("compiler bug: incomplete primitive array read contract")
	}
	members := []string{}
	for _, member := range contracts {
		kind := int(member.Of)
		if member.Kind == ir.ViewUndefined {
			kind = 9
		}
		if len(member.Allowed) == 0 {
			members = append(members, fmt.Sprintf("{%d, false, {.number = 0}, 0}", kind))
			continue
		}
		for _, literal := range member.Allowed {
			value := ""
			switch literal.Of {
			case ir.Number:
				value = "{.number = " + cNumber(literal.Number) + "}"
			case ir.Boolean:
				value = fmt.Sprintf("{.boolean = %t}", literal.Boolean)
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				value = "{.reference = &" + name + "}"
			}
			members = append(members, fmt.Sprintf("{%d, true, %s, 0}", kind, value))
		}
	}
	e.declarations = append(e.declarations, "#include \"view_array_primitives.h\"")
	// Index callbacks can replace the receiver. Keep its original identity alive.
	array := e.own(ir.Array, "adamic_retain("+e.value(read.Array)+")")
	index := e.value(read.Index)
	snapshot := e.temporary()
	e.line("adamic_view_union_value %s = adamic_view_primitive_array_snapshot(%s, %s, %t, %s, %s);", snapshot, array, index, read.Relative, cString(read.View), cString(read.ViewType))
	e.line("(void)adamic_view_mixed_union_select(&%s, (const adamic_view_union_member[]){%s}, %d, NULL, NULL, %s, %s);", snapshot, strings.Join(members, ", "), len(members), cString(read.View), cString(read.ViewType))
	return e.own(ir.Union, "adamic_view_primitive_array_box("+snapshot+")")
}

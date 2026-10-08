package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) viewObjectPrimitive(property ir.Property) string {
	contract := e.program.ViewContracts[property.ViewContract-1]
	members := []string{}
	undefined := false
	for _, id := range contract.Members {
		child := e.program.ViewContracts[id-1]
		if child.Kind == ir.ViewUndefined {
			undefined = true
			continue
		}
		if child.Kind != ir.ViewScalar && child.Kind != ir.ViewObject && child.Kind != ir.ViewArray {
			panic("compiler bug: unavailable object primitive member adapter")
		}
		if child.Unsupported != "" {
			panic("compiler bug: unsupported union member")
		}
		if len(child.Allowed) == 0 {
			members = append(members, fmt.Sprintf("{%d, false, {.number = 0}}", child.Of))
		} else {
			for _, literal := range child.Allowed {
				value := ""
				switch literal.Of {
				case ir.Boolean:
					value = fmt.Sprintf("{.boolean = %t}", literal.Boolean)
				case ir.Number:
					value = "{.number = " + cNumber(literal.Number) + "}"
				case ir.String:
					name := e.temporary()
					e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
					value = "{.reference = &" + name + "}"
				}
				members = append(members, fmt.Sprintf("{%d, true, %s}", child.Of, value))
			}
		}
	}
	e.declarations = append(e.declarations, "#include \"view_unions_object_primitive.h\"")
	object := e.value(property.Object)
	value := fmt.Sprintf("adamic_object_primitive_view(%s, %s, &%s, (const adamic_object_primitive_member[]){%s}, %d, %t, %t, %t, %s, %s)", object, cString(property.Name), e.cache(), strings.Join(members, ", "), len(members), undefined, property.Absent, property.Optional, cString(contract.Name), cString(property.View))
	result := e.own(ir.Union, value)
	for _, id := range contract.Members {
		if e.program.ViewContracts[id-1].FixedTuple {
			e.line("if (%s != NULL && ((const adamic_heap *)%s)->kind == adamic_kind_object) {", result, result)
			e.viewTuple(ir.Property{View: property.View, ViewContract: id}, result)
			e.line("}")
		}
	}
	return result
}

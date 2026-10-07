package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// Test union membership at the returned object, then leave member payload reads
// checked. The caller has already evaluated the receiver exactly once.
func (e *emitter) viewObjectUnion(property ir.Property, object string) {
	if property.ViewContract == 0 {
		return
	}
	contract := e.program.ViewContracts[int(property.ViewContract)-1]
	if contract.Kind != ir.ViewUnion {
		return
	}
	for _, field := range contract.Fields {
		tag := e.program.ViewContracts[int(field.Contract)-1]
		if tag.Kind != ir.ViewScalar || len(tag.Allowed) == 0 {
			continue
		}
		slot := e.temporary()
		expression := property.View + "." + field.Name
		e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(field.Name), e.cache(), tag.Of, cString(tag.Name), cString(expression))
		tests := []string{}
		for _, allowed := range tag.Allowed {
			switch allowed.Of {
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(allowed.String)))
				tests = append(tests, fmt.Sprintf("adamic_string_equal((const adamic_string *)%s.reference, &%s)", slot, name))
			case ir.Number:
				tests = append(tests, fmt.Sprintf("%s.number == %s", slot, strconv.FormatFloat(allowed.Number, 'g', -1, 64)))
			case ir.Boolean:
				tests = append(tests, fmt.Sprintf("%s.boolean == %t", slot, allowed.Boolean))
			}
		}
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString(expression), cString(tag.Name), tag.Of, slot)
		return
	}
	panic("compiler bug: object union has no checked discriminant")
}

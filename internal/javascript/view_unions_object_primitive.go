package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) viewObjectPrimitive(property ir.Property) string {
	contract := e.program.ViewContracts[property.ViewContract-1]
	tests := []string{}
	for _, id := range contract.Members {
		child := e.program.ViewContracts[id-1]
		if child.Kind == ir.ViewUndefined {
			tests = append(tests, "v === undefined")
			continue
		}
		if child.Unsupported != "" {
			panic("compiler bug: unsupported union member")
		}
		if len(child.Allowed) != 0 {
			for _, literal := range child.Allowed {
				value := ""
				switch literal.Of {
				case ir.String:
					value = quote(literal.String)
				case ir.Number:
					value = strconv.FormatFloat(literal.Number, 'g', -1, 64)
				case ir.Boolean:
					value = strconv.FormatBool(literal.Boolean)
				}
				tests = append(tests, "v === "+value)
			}
		} else if child.Kind == ir.ViewArray && child.Of == ir.Array {
			tests = append(tests, "Array.isArray(v)")
		} else if child.Kind == ir.ViewObject && child.Of == ir.Object {
			tests = append(tests, "v !== null && typeof v === 'object' && !Array.isArray(v) && !(v instanceof Map)")
		} else if child.Kind == ir.ViewScalar {
			kind := map[ir.Type]string{ir.String: "string", ir.Number: "number", ir.Boolean: "boolean"}[child.Of]
			tests = append(tests, "typeof v === "+quote(kind))
		} else {
			panic("compiler bug: unavailable object primitive member adapter")
		}
	}
	return fmt.Sprintf("((o) => { const v = adamicReadField(o, %s, %s, false, false, %s); if (!(%s)) panic('field read failed: ' + %s + ' matches no member of ' + %s + '; expected ' + %s + ', found ' + (v === undefined ? 'undefined' : v === null ? 'null' : Array.isArray(v) ? 'array' : v instanceof Map ? 'Map' : typeof v)); return v; })(%s)", quote(property.Name), quote(property.View), quote(contract.Name), strings.Join(tests, " || "), quote(property.View), quote(contract.Name), quote(contract.Name), e.value(property.Object))
}

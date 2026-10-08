package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) emitPrimitiveArrayIndex(read ir.ArrayIndex) string {
	members, ok := ir.PrimitiveViewMembers(e.program, read.ViewContract)
	if !ok {
		panic("compiler bug: incomplete primitive array read contract")
	}
	tests := []string{}
	for _, member := range members {
		if member.Kind == ir.ViewUndefined {
			tests = append(tests, "v === undefined")
			continue
		}
		if len(member.Allowed) == 0 {
			tests = append(tests, "typeof v === "+quote(map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[member.Of]))
			continue
		}
		for _, literal := range member.Allowed {
			value := ""
			switch literal.Of {
			case ir.Number:
				value = strconv.FormatFloat(literal.Number, 'g', -1, 64)
			case ir.Boolean:
				value = strconv.FormatBool(literal.Boolean)
			case ir.String:
				value = quote(literal.String)
			}
			tests = append(tests, "v === "+value)
		}
	}
	return fmt.Sprintf("((a, i) => { if (!Array.isArray(a)) panic('cast failed: field read failed: ' + %s + ' matches no member of ' + %s + '; expected ' + %s + ', found unsupported representation'); if (%t) { i = Math.trunc(Number(i)) || 0; if (i < 0) i += a.length; } const v = a[i]; if (!(%s)) panic('cast failed: field read failed: ' + %s + ' matches no member of ' + %s + '; expected ' + %s + ', found ' + (v === null ? 'null' : Array.isArray(v) ? 'array' : typeof v)); return v; })(%s, %s)", quote(read.View), quote(read.ViewType), quote(read.ViewType), read.Relative, strings.Join(tests, " || "), quote(read.View), quote(read.ViewType), quote(read.ViewType), e.value(read.Array), e.value(read.Index))
}

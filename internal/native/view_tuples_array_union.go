package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// This uses the existing mixed-union selector and the existing tuple shape
// witness. Normalization and ownership belong to the caller, not selection.
func (e *emitter) viewTupleHeapUnion(id ir.ViewContractID, value, expression string) bool {
	ids, ok := ir.TupleViewMembers(e.program, id)
	if !ok {
		return false
	}
	e.declarations = append(e.declarations, "#include \"view_tuples.h\"", "#include \"view_unions_mixed.h\"")
	callback := e.temporary() + "_tuple_match"
	cases, members := []string{}, []string{}
	for _, childID := range ids {
		c := e.program.ViewContracts[childID-1]
		kind := "adamic_view_union_undefined"
		if c.FixedTuple {
			kind = "adamic_view_union_object"
			cases = append(cases, fmt.Sprintf("case %d: return adamic_tuple_matches(value->payload.reference, %d);", childID, len(c.Tuple)))
		} else if c.Kind == ir.ViewScalar {
			kind = map[ir.Type]string{ir.Number: "adamic_view_union_number", ir.String: "adamic_view_union_string", ir.Boolean: "adamic_view_union_boolean"}[c.Of]
		}
		if len(c.Allowed) == 0 {
			members = append(members, fmt.Sprintf("{%s, false, {.number=0}, %d}", kind, childID))
			continue
		}
		for _, literal := range c.Allowed {
			payload := ""
			switch literal.Of {
			case ir.Number:
				payload = "{.number=" + cNumber(literal.Number) + "}"
			case ir.Boolean:
				payload = fmt.Sprintf("{.boolean=%t}", literal.Boolean)
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				payload = "{.reference=&" + name + "}"
			}
			members = append(members, fmt.Sprintf("{%s, true, %s, %d}", kind, payload, childID))
		}
	}
	e.declarations = append(e.declarations, "static bool "+callback+"(void *context, const adamic_view_union_member *member, const adamic_view_union_value *value) { (void)context; switch(member->contract) { "+strings.Join(cases, " ")+" default: return false; } }")
	snapshot := e.temporary()
	e.line("adamic_view_union_value %s = adamic_view_union_heap((const adamic_heap *)%s);", snapshot, value)
	e.line("(void)adamic_view_mixed_union_select(&%s, (const adamic_view_union_member[]){%s}, %d, %s, NULL, %s, %s);", snapshot, strings.Join(members, ","), len(members), callback, cString(expression), cString(e.program.ViewContracts[id-1].Name))
	return true
}

// Index reads own normalization until the existing statement cleanup. Callback
// and loop consumers remain refused until their distinct transfer is certified.
func (e *emitter) emitTupleArrayUnionIndex(read ir.ArrayIndex, array, index string) string {
	e.declarations = append(e.declarations, "#include \"view_tuples.h\"")
	snapshot, present, slot := e.temporary(), e.temporary(), e.temporary()
	e.line("adamic_value %s;", snapshot)
	e.line("bool %s = adamic_tuple_array_union_at(%s, %s, %t, %t, %s, %s, &%s);", present, array, index, read.Relative, read.UndefinedAllowed, cString(read.ViewType), cString(read.View), snapshot)
	value := e.own(ir.Union, snapshot+".reference")
	if read.Required {
		e.line("if (!%s) adamic_view_array_missing(%s, %s);", present, cString(read.View), cString(read.ViewType))
	}
	e.line("if (%s) {", present)
	e.indent++
	if !e.viewTupleHeapUnion(read.ViewContract, value, read.View) {
		panic("compiler bug: missing tuple union plan")
	}
	e.indent--
	e.line("}")
	e.line("adamic_value *%s = %s ? &%s : NULL;", slot, present, snapshot)
	return slot
}

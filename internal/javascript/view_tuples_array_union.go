package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) tupleArrayUnionCheck(id ir.ViewContractID, expression, value string) string {
	ids, ok := ir.TupleViewMembers(e.program, id)
	if !ok {
		panic("compiler bug: missing tuple union member plan")
	}
	members, cases := []string{}, []string{}
	for _, childID := range ids {
		c := e.program.ViewContracts[childID-1]
		kind := "undefined"
		if c.FixedTuple {
			kind = "object"
			if c.TupleVariable {
				upper := fmt.Sprintf("snapshot.value.length <= %d", len(c.Tuple))
				if c.TupleRest != 0 {
					upper = "true"
				}
				cases = append(cases, fmt.Sprintf("case %d: return Array.isArray(snapshot.value) && snapshot.value.length >= %d && %s;", childID, c.TupleMinimum, upper))
			} else {
				cases = append(cases, fmt.Sprintf("case %d: return Array.isArray(snapshot.value) && snapshot.value.length === %d;", childID, len(c.Tuple)))
			}
		} else if c.Kind == ir.ViewScalar {
			kind = map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean"}[c.Of]
		}
		if len(c.Allowed) == 0 {
			members = append(members, fmt.Sprintf("{kind:%s,contract:%d}", quote(kind), childID))
			continue
		}
		for _, literal := range c.Allowed {
			value := quote(literal.String)
			if literal.Of == ir.Number {
				value = strconv.FormatFloat(literal.Number, 'g', -1, 64)
			}
			if literal.Of == ir.Boolean {
				value = strconv.FormatBool(literal.Boolean)
			}
			members = append(members, fmt.Sprintf("{kind:%s,contract:%d,literal:true,value:%s}", quote(kind), childID, value))
		}
	}
	return "((v) => {" + viewMixedUnionsRuntime + " const snapshot = {kind:v === undefined ? 'undefined' : v === null ? 'null' : typeof v,value:v}; adamicViewMixedUnionSelect(snapshot,[" + strings.Join(members, ",") + "],(member,snapshot) => {switch(member.contract){" + strings.Join(cases, " ") + "default:return false;}}," + quote(expression) + "," + quote(e.program.ViewContracts[id-1].Name) + "); return v;})(" + value + ")"
}

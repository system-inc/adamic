package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// IntersectionRuntime consumes shared normalized snapshots. It grants neither
// readiness nor ownership and cannot replace the receiver's per-read guard.
func IntersectionRuntime() string { return viewIntersectionsRuntime }

const viewIntersectionsRuntime = `
const adamicViewIntersectionMatches = (snapshot, members, match) => {
 if (snapshot === undefined || snapshot.kind === 'unknown' || members.length === 0 || match === undefined) return false;
 for (const member of members) {
  if (!member || !match(member, snapshot)) return false;
 }
 return true;
};
const adamicViewIntersectionRequire = (snapshot, members, match, expression, declared) => {
 if (adamicViewIntersectionMatches(snapshot, members, match)) return;
 const found = snapshot === undefined || snapshot.kind === 'unknown' ? 'unsupported representation' : snapshot.kind;
 panic('field read failed: ' + expression + ' does not satisfy every member; expected ' + declared + ', found ' + found);
};
`

// Retain one evaluated receiver and one snapshot per combined field.
func (e *emitter) viewObjectIntersection(property ir.Property, value string) string {
	body := e.viewIntersectionFields(property.ViewContract, "adamicIntersectionValue", property.View, map[ir.ViewContractID]bool{})
	return "((adamicIntersectionValue) => { if (adamicIntersectionValue !== undefined) {" + body + "} return adamicIntersectionValue; })(" + value + ")"
}
func (e *emitter) viewIntersectionFields(id ir.ViewContractID, object, expression string, active map[ir.ViewContractID]bool) string {
	if active[id] {
		return ""
	}
	active[id] = true
	defer delete(active, id)
	body := []string{}
	contract := e.program.ViewContracts[id-1]
	for index, field := range contract.Fields {
		child := e.program.ViewContracts[field.Contract-1]
		if child.Unsupported != "" || child.Kind == ir.ViewUnknown || child.Kind == ir.ViewCallable || child.Kind == ir.ViewUndefined || child.Of == ir.Union {
			continue
		}
		path := expression + "." + field.Name
		slot := fmt.Sprintf("adamicIntersectionField%d_%d", id, index)
		allowed := []string{}
		for _, literal := range child.Allowed {
			switch literal.Of {
			case ir.String:
				allowed = append(allowed, quote(literal.String))
			case ir.Number:
				allowed = append(allowed, strconv.FormatFloat(literal.Number, 'g', -1, 64))
			case ir.Boolean:
				allowed = append(allowed, strconv.FormatBool(literal.Boolean))
			}
		}
		optional := field.Optional || child.Undefined
		body = append(body, fmt.Sprintf("{ const %s = adamicViewField(%s,%s,%s,%d,%s,[%s],%t,%t);", slot, object, quote(field.Name), quote(path), child.Of, quote(child.Name), strings.Join(allowed, ","), optional, optional))
		if child.Of == ir.Object {
			if child.Kind == ir.ViewUnion {
				body = append(body, "if ("+slot+" !== undefined) {"+e.viewObjectUnion(ir.Property{View: path, ViewContract: field.Contract}, slot)+";}")
			} else {
				body = append(body, "if ("+slot+" !== undefined) {"+e.viewIntersectionFields(field.Contract, slot, path, active)+"}")
			}
		}
		body = append(body, "}")
	}
	return strings.Join(body, "\n")
}

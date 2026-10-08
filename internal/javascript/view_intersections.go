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
	if e.program.ViewContracts[property.ViewContract-1].IntersectionRecursive {
		return e.viewRecursiveIntersection(property, value)
	}
	body := e.viewIntersectionFields(property.ViewContract, "adamicIntersectionValue", property.View, map[ir.ViewContractID]bool{})
	return "((adamicIntersectionValue) => { if (adamicIntersectionValue !== undefined) {" + body + "} return adamicIntersectionValue; })(" + value + ")"
}
func (e *emitter) viewIntersectionFields(id ir.ViewContractID, object, expression string, active map[ir.ViewContractID]bool, skip ...string) string {
	if active[id] {
		return ""
	}
	active[id] = true
	defer delete(active, id)
	body := []string{}
	contract := e.program.ViewContracts[id-1]
	for index, field := range contract.Fields {
		if len(skip) > 0 && field.Name == skip[0] {
			continue
		}
		child := e.program.ViewContracts[field.Contract-1]
		if child.Unsupported != "" || child.Kind == ir.ViewUnknown || child.Kind == ir.ViewUndefined || child.Of == ir.Union {
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

func (e *emitter) viewIntersectionUnion(property ir.Property, value string) string {
	root := e.program.ViewContracts[property.ViewContract-1]
	var tag ir.ViewContract
	for _, field := range root.Fields {
		if field.Name == root.IntersectionTag {
			tag = e.program.ViewContracts[field.Contract-1]
		}
	}
	body := fmt.Sprintf("const adamicIntersectionTag = adamicViewField(adamicIntersectionValue,%s,%s,%d,%s);", quote(root.IntersectionTag), quote(property.View+"."+root.IntersectionTag), tag.Of, quote(tag.Name))
	for index, member := range root.Members {
		arm := e.program.ViewContracts[member-1]
		allowed := []string{}
		for _, field := range arm.Fields {
			if field.Name == root.IntersectionTag {
				for _, literal := range e.program.ViewContracts[field.Contract-1].Allowed {
					switch literal.Of {
					case ir.Number:
						allowed = append(allowed, strconv.FormatFloat(literal.Number, 'g', -1, 64))
					case ir.Boolean:
						allowed = append(allowed, strconv.FormatBool(literal.Boolean))
					case ir.String:
						allowed = append(allowed, quote(literal.String))
					}
				}
			}
		}
		prefix := "if"
		if index != 0 {
			prefix = "else if"
		}
		body += prefix + " ([" + strings.Join(allowed, ",") + "].includes(adamicIntersectionTag)) {" + e.viewIntersectionFields(member, "adamicIntersectionValue", property.View, map[ir.ViewContractID]bool{}, root.IntersectionTag) + "}"
	}
	body += "else {panic(" + quote("field read failed: "+property.View+"."+root.IntersectionTag+" expected "+tag.Name+", found ") + "+typeof adamicIntersectionTag+' '+adamicIntersectionTag); }"
	return "((adamicIntersectionValue) => { if (adamicIntersectionValue !== undefined) {" + body + "} return adamicIntersectionValue; })(" + value + ")"
}

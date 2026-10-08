package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

type recursiveIntersectionField struct {
	Name      string            `json:"name"`
	Expected  string            `json:"expected"`
	Type      ir.Type           `json:"type"`
	Optional  bool              `json:"optional"`
	Undefined bool              `json:"undefined"`
	Child     ir.ViewContractID `json:"child"`
	Allowed   []any             `json:"allowed"`
}

func (e *emitter) viewRecursiveIntersection(property ir.Property, value string) string {
	descriptors := [][]recursiveIntersectionField{nil}
	ids := ir.RecursiveIntersectionObjects(e.program, property.ViewContract)
	indices := map[ir.ViewContractID]ir.ViewContractID{}
	for i, id := range ids {
		indices[id] = ir.ViewContractID(i + 1)
	}
	for _, id := range ids {
		contract := e.program.ViewContracts[id-1]
		fields := []recursiveIntersectionField{}
		for _, field := range contract.Fields {
			child := e.program.ViewContracts[field.Contract-1]
			if child.Unsupported != "" && child.Unsupported != "recursive intersection payload" || child.Kind == ir.ViewUnknown {
				continue
			}
			target := field.Contract
			if child.ObjectPresent != 0 {
				target = child.ObjectPresent
			}
			nested := ir.ViewContractID(0)
			if e.program.ViewContracts[target-1].Kind == ir.ViewObject {
				nested = indices[target]
			}
			allowed := []any{}
			for _, literal := range child.Allowed {
				switch literal.Of {
				case ir.Number:
					allowed = append(allowed, literal.Number)
				case ir.String:
					allowed = append(allowed, literal.String)
				case ir.Boolean:
					allowed = append(allowed, literal.Boolean)
				}
			}
			fields = append(fields, recursiveIntersectionField{field.Name, child.Name, child.Of, field.Optional, child.Undefined, nested, allowed})
		}
		descriptors = append(descriptors, fields)
	}
	data, _ := json.Marshal(descriptors)
	return fmt.Sprintf(`((value) => {
 const contracts = %s;
 const active = new Map();
 const check = (object,id,path) => {
  if (object === undefined) return;
  let ids = active.get(object);
  if (ids !== undefined && ids.has(id)) return;
  if (ids === undefined) { ids = new Set(); active.set(object,ids); }
  ids.add(id);
  for (const field of contracts[id] || []) {
   const next = path + '.' + field.name;
   const slot = adamicViewField(object,field.name,next,field.type,field.expected,field.allowed,field.optional,false,field.undefined);
   if (slot !== undefined && field.child !== 0) check(slot,field.child,next);
  }
  ids.delete(id);
 };
 check(value,%d,%s); return value;
})(%s)`, data, indices[ir.RecursiveIntersectionPresent(e.program, property.ViewContract)], quote(property.View), value)
}

func (e *emitter) viewIntersectionNullishRead(property ir.Property, value string) string {
	if property.ViewContract == 0 {
		return value
	}
	id := property.ViewContract
	c := e.program.ViewContracts[id-1]
	if c.Kind == ir.ViewNullable {
		id = c.Element
		if id == 0 {
			return value
		}
		c = e.program.ViewContracts[id-1]
	}
	if !c.Intersection {
		return value
	}
	property.ViewContract = id
	return "((present) => present === undefined || present === null ? present : " + e.viewObjectUnion(property, "present") + ")(" + value + ")"
}

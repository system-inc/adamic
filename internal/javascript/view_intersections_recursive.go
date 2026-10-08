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

type boundedIntersectionArm struct {
	Allowed  []any             `json:"allowed"`
	Contract ir.ViewContractID `json:"contract"`
}

type boundedIntersectionContract struct {
	Fields      []recursiveIntersectionField `json:"fields"`
	Tag         string                       `json:"tag,omitempty"`
	TagExpected string                       `json:"tagExpected,omitempty"`
	TagType     ir.Type                      `json:"tagType,omitempty"`
	Arms        []boundedIntersectionArm     `json:"arms,omitempty"`
}

func boundedLiterals(allowed []ir.ViewLiteral) []any {
	values := []any{}
	for _, literal := range allowed {
		switch literal.Of {
		case ir.Number:
			values = append(values, literal.Number)
		case ir.String:
			values = append(values, literal.String)
		case ir.Boolean:
			values = append(values, literal.Boolean)
		}
	}
	return values
}

// The bounded table holds every object and selected union the read can reach.
// A descriptor already active on the path leaves its fields to their own reads.
func (e *emitter) viewBoundedIntersection(property ir.Property, value string) string {
	contracts := e.program.ViewContracts
	ids, _ := ir.BoundedIntersectionContracts(e.program, property.ViewContract, nil)
	indices := map[ir.ViewContractID]ir.ViewContractID{}
	for i, id := range ids {
		indices[id] = ir.ViewContractID(i + 1)
	}
	descriptors := []boundedIntersectionContract{{}}
	for _, id := range ids {
		contract := contracts[id-1]
		if contract.Kind == ir.ViewUnion {
			tag, arms := ir.IntersectionUnionArms(e.program, id)
			descriptor := boundedIntersectionContract{Fields: []recursiveIntersectionField{}}
			if tag == "" {
				var discriminant ir.ViewContractID
				tag, discriminant = ir.UnionDiscriminant(e.program, id)
				descriptor.TagExpected, descriptor.TagType = contracts[discriminant-1].Name, contracts[discriminant-1].Of
				descriptor.Arms = []boundedIntersectionArm{{boundedLiterals(contracts[discriminant-1].Allowed), 0}}
			} else {
				for _, field := range contract.Fields {
					if field.Name == tag {
						descriptor.TagExpected, descriptor.TagType = contracts[field.Contract-1].Name, contracts[field.Contract-1].Of
					}
				}
				for _, arm := range arms {
					descriptor.Arms = append(descriptor.Arms, boundedIntersectionArm{boundedLiterals(ir.IntersectionArmLiterals(e.program, arm, tag)), indices[arm]})
				}
			}
			descriptor.Tag = tag
			descriptors = append(descriptors, descriptor)
			continue
		}
		fields := []recursiveIntersectionField{}
		for _, field := range contract.Fields {
			kind, next := ir.BoundedIntersectionField(e.program, field.Contract, nil)
			if kind == ir.BoundedDeferred {
				continue
			}
			child := contracts[field.Contract-1]
			nested := ir.ViewContractID(0)
			if kind == ir.BoundedObject || kind == ir.BoundedArms {
				nested = indices[next]
			}
			fields = append(fields, recursiveIntersectionField{field.Name, child.Name, child.Of, field.Optional, child.Undefined, nested, boundedLiterals(child.Allowed)})
		}
		descriptors = append(descriptors, boundedIntersectionContract{Fields: fields})
	}
	data, _ := json.Marshal(descriptors)
	return fmt.Sprintf(`((value) => {
 const contracts = %s;
 const active = new Set();
 const check = (object,id,path) => {
  if (object === undefined || object === null || active.has(id)) return;
  active.add(id);
  const contract = contracts[id];
  if (contract.tag !== undefined) {
   const next = path + '.' + contract.tag;
   const tag = adamicViewField(object,contract.tag,next,contract.tagType,contract.tagExpected);
   const arm = contract.arms.find(arm => arm.allowed.includes(tag));
   if (arm === undefined) panic('field read failed: ' + next + ' expected ' + contract.tagExpected + ', found ' + typeof tag + ' ' + tag);
   if (arm.contract !== 0) check(object,arm.contract,path);
  } else {
   for (const field of contract.fields) {
    const next = path + '.' + field.name;
    const slot = adamicViewField(object,field.name,next,field.type,field.expected,field.allowed,field.optional,false,field.undefined);
    if (slot !== undefined && field.child !== 0) check(slot,field.child,next);
   }
  }
  active.delete(id);
 };
 check(value,1,%s); return value;
})(%s)`, data, quote(property.View), value)
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

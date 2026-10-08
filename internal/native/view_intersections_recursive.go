package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) viewRecursiveIntersection(property ir.Property, object string) {
	name := e.temporary() + "_recursive"
	contracts := e.program.ViewContracts
	fields := []string{"{NULL,0,NULL,NULL,0,NULL,0}"}
	ids := ir.RecursiveIntersectionObjects(e.program, property.ViewContract)
	indices := map[ir.ViewContractID]int{}
	for i, id := range ids {
		indices[id] = i + 1
	}
	for index, id := range ids {
		contract := contracts[id-1]
		parts := []string{}
		for _, field := range contract.Fields {
			child := contracts[field.Contract-1]
			if child.Unsupported != "" && child.Unsupported != "recursive intersection payload" || child.Kind == ir.ViewUnknown {
				continue
			}
			target := field.Contract
			if child.ObjectPresent != 0 {
				target = child.ObjectPresent
			}
			nested := 0
			if contracts[target-1].Kind == ir.ViewObject {
				nested = indices[target]
			}
			allowed := []string{}
			for _, literal := range child.Allowed {
				allowed = append(allowed, fmt.Sprintf("{%d,%s,%t,%s,%d}", literal.Of, strconv.FormatFloat(literal.Number, 'g', -1, 64), literal.Boolean, cString(literal.String), len(literal.String)))
			}
			literalName := "NULL"
			if len(allowed) > 0 {
				literalName = fmt.Sprintf("%s_%d_%d_literals", name, index, len(parts))
				e.declarations = append(e.declarations, "static const adamic_intersection_literal "+literalName+"[] = {"+strings.Join(allowed, ",")+"};")
			}
			parts = append(parts, fmt.Sprintf("{%s,%s,%d,%t,%t,%d,%s,%d}", cString(field.Name), cString(child.Name), child.Of, field.Optional, child.Undefined, nested, literalName, len(allowed)))
		}
		if len(parts) == 0 {
			fields = append(fields, "{NULL,0,NULL,NULL,0,NULL,0}")
			continue
		}
		fieldName := fmt.Sprintf("%s_%d_fields", name, index)
		e.declarations = append(e.declarations, "static const adamic_intersection_field "+fieldName+"[] = {"+strings.Join(parts, ",")+"};")
		fields = append(fields, fmt.Sprintf("{%s,%d,NULL,NULL,0,NULL,0}", fieldName, len(parts)))
	}
	e.declarations = append(e.declarations, "static const adamic_intersection_contract "+name+"[] = {"+strings.Join(fields, ",")+"};")
	e.line("adamic_intersection_recursive_require(%s,%s,%d,%s);", object, name, indices[ir.RecursiveIntersectionPresent(e.program, property.ViewContract)], cString(property.View))
}

// The bounded table holds every object and selected union the read can reach.
// Field kinds and literals are checked at each level; a descriptor already
// active on the path leaves its fields to their own checked reads.
func (e *emitter) viewBoundedIntersection(property ir.Property, object string) {
	name := e.temporary() + "_bounded"
	contracts := e.program.ViewContracts
	ids, _ := ir.BoundedIntersectionContracts(e.program, property.ViewContract, nil)
	indices := map[ir.ViewContractID]int{}
	for i, id := range ids {
		indices[id] = i + 1
	}
	literals := func(prefix string, allowed []ir.ViewLiteral) string {
		if len(allowed) == 0 {
			return "NULL"
		}
		parts := []string{}
		for _, literal := range allowed {
			parts = append(parts, fmt.Sprintf("{%d,%s,%t,%s,%d}", literal.Of, strconv.FormatFloat(literal.Number, 'g', -1, 64), literal.Boolean, cString(literal.String), len(literal.String)))
		}
		e.declarations = append(e.declarations, "static const adamic_intersection_literal "+prefix+"[] = {"+strings.Join(parts, ",")+"};")
		return prefix
	}
	entries := []string{"{NULL,0,NULL,NULL,0,NULL,0}"}
	for index, id := range ids {
		contract := contracts[id-1]
		if contract.Kind == ir.ViewUnion {
			tag, arms := ir.IntersectionUnionArms(e.program, id)
			armParts := []string{}
			var tagContract ir.ViewContract
			if tag == "" {
				var discriminant ir.ViewContractID
				tag, discriminant = ir.UnionDiscriminant(e.program, id)
				tagContract = contracts[discriminant-1]
				armParts = append(armParts, fmt.Sprintf("{%s,%d,0}", literals(fmt.Sprintf("%s_%d_tag", name, index), tagContract.Allowed), len(tagContract.Allowed)))
			} else {
				for _, field := range contract.Fields {
					if field.Name == tag {
						tagContract = contracts[field.Contract-1]
					}
				}
				for armIndex, arm := range arms {
					allowed := ir.IntersectionArmLiterals(e.program, arm, tag)
					armParts = append(armParts, fmt.Sprintf("{%s,%d,%d}", literals(fmt.Sprintf("%s_%d_%d_arm", name, index, armIndex), allowed), len(allowed), indices[arm]))
				}
			}
			armName := fmt.Sprintf("%s_%d_arms", name, index)
			e.declarations = append(e.declarations, "static const adamic_intersection_arm "+armName+"[] = {"+strings.Join(armParts, ",")+"};")
			entries = append(entries, fmt.Sprintf("{NULL,0,%s,%s,%d,%s,%d}", cString(tag), cString(tagContract.Name), tagContract.Of, armName, len(armParts)))
			continue
		}
		parts := []string{}
		for _, field := range contract.Fields {
			kind, next := ir.BoundedIntersectionField(e.program, field.Contract, nil)
			if kind == ir.BoundedDeferred {
				continue
			}
			child := contracts[field.Contract-1]
			nested := 0
			if kind == ir.BoundedObject || kind == ir.BoundedArms {
				nested = indices[next]
			}
			allowed := literals(fmt.Sprintf("%s_%d_%d_literals", name, index, len(parts)), child.Allowed)
			parts = append(parts, fmt.Sprintf("{%s,%s,%d,%t,%t,%d,%s,%d}", cString(field.Name), cString(child.Name), child.Of, field.Optional, child.Undefined, nested, allowed, len(child.Allowed)))
		}
		if len(parts) == 0 {
			entries = append(entries, "{NULL,0,NULL,NULL,0,NULL,0}")
			continue
		}
		fieldName := fmt.Sprintf("%s_%d_fields", name, index)
		e.declarations = append(e.declarations, "static const adamic_intersection_field "+fieldName+"[] = {"+strings.Join(parts, ",")+"};")
		entries = append(entries, fmt.Sprintf("{%s,%d,NULL,NULL,0,NULL,0}", fieldName, len(parts)))
	}
	e.declarations = append(e.declarations, "static const adamic_intersection_contract "+name+"[] = {"+strings.Join(entries, ",")+"};")
	e.line("adamic_intersection_bounded_require(%s,%s,1,%s);", object, name, cString(property.View))
}

// Nullable physical conversion must not erase a present intersection obligation.
func (e *emitter) viewIntersectionNullishRead(property ir.Property, value string) {
	if property.ViewContract == 0 {
		return
	}
	id := property.ViewContract
	c := e.program.ViewContracts[id-1]
	if c.Kind == ir.ViewNullable {
		id = c.Element
		if id == 0 {
			return
		}
		c = e.program.ViewContracts[id-1]
	}
	if !c.Intersection {
		return
	}
	property.ViewContract = id
	e.line("if (%s != NULL && %s != &adamic_null) {", value, value)
	e.viewObjectUnion(property, fmt.Sprintf("(const adamic_object *)%s", value))
	e.line("}")
}

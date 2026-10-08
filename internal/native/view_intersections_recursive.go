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
	fields := []string{"{NULL,0}"}
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
			fields = append(fields, "{NULL,0}")
			continue
		}
		fieldName := fmt.Sprintf("%s_%d_fields", name, index)
		e.declarations = append(e.declarations, "static const adamic_intersection_field "+fieldName+"[] = {"+strings.Join(parts, ",")+"};")
		fields = append(fields, fmt.Sprintf("{%s,%d}", fieldName, len(parts)))
	}
	e.declarations = append(e.declarations, "static const adamic_intersection_contract "+name+"[] = {"+strings.Join(fields, ",")+"};")
	e.line("adamic_intersection_recursive_require(%s,%s,%d,%s);", object, name, indices[ir.RecursiveIntersectionPresent(e.program, property.ViewContract)], cString(property.View))
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

package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// The arrow holds the returned object, preserving a side-effectful receiver.
func (e *emitter) viewObjectUnion(property ir.Property, value string) string {
	if property.ViewContract == 0 {
		return value
	}
	contract := e.program.ViewContracts[int(property.ViewContract)-1]
	if contract.Intersection {
		return e.viewObjectIntersection(property, value)
	}
	if contract.IntersectionTag != "" {
		return e.viewIntersectionUnion(property, value)
	}
	if contract.Kind != ir.ViewUnion || (contract.Of != ir.Object && contract.Of != ir.Array) {
		return value
	}
	if untaggedObjectUnion(e.program.ViewContracts, contract) {
		return e.viewUntaggedObjectUnion(property, value)
	}
	for _, field := range contract.Fields {
		tag := e.program.ViewContracts[int(field.Contract)-1]
		if tag.Kind != ir.ViewScalar || len(tag.Allowed) == 0 {
			continue
		}
		allowed := []string{}
		for _, literal := range tag.Allowed {
			switch literal.Of {
			case ir.String:
				allowed = append(allowed, quote(literal.String))
			case ir.Number:
				allowed = append(allowed, strconv.FormatFloat(literal.Number, 'g', -1, 64))
			case ir.Boolean:
				allowed = append(allowed, strconv.FormatBool(literal.Boolean))
			}
		}
		return fmt.Sprintf("((adamicViewValue) => { adamicViewField(adamicViewValue, %s, %s, %d, %s, [%s]); return adamicViewValue; })(%s)", quote(field.Name), quote(property.View+"."+field.Name), tag.Of, quote(tag.Name), strings.Join(allowed, ", "), value)
	}
	panic("compiler bug: object union has no checked discriminant")
}

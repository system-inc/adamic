package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

type checkedViewReadHook func(*emitter, ir.Property) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) checkedViewField(property ir.Property) string {
	if property.Namespace {
		return fmt.Sprintf("adamicNamespaceRead(%s, %s, %d)", e.value(property.Object), quote(property.Name), property.Of)
	}
	if property.Absent && (property.Of == ir.MaybeNumber || property.Of == ir.MaybeBoolean || property.Of == ir.String) {
		return fmt.Sprintf("adamicOptionalViewField(%s, %s, %s, %d, %s, [%s])", e.value(property.Object), quote(property.Name), quote(property.View), property.Of, quote(property.ViewType), e.values(property.ViewAllowed))
	}
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property); handled {
			return value
		}
	}
	expected := property.ViewType
	if expected == "" {
		expected = map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}[property.Of]
	}
	return fmt.Sprintf("adamicViewField(%s, %s, %s, %d, %s, [%s])", e.value(property.Object), quote(property.Name), quote(property.View), property.Of, quote(expected), e.values(property.ViewAllowed))
}

// V1 records which checked field contracts permit absence. Required-field
// construction keeps its existing write path and readiness transitions.
func (e *emitter) optionalViewFieldName(name string) bool {
	for _, contract := range e.program.ViewContracts {
		for _, field := range contract.Fields {
			if field.Name == name && field.Optional {
				return true
			}
		}
	}
	return false
}

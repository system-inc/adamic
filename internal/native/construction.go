package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func constructionFields(fields []ir.Field) bool {
	for _, field := range fields {
		if field.Absent {
			return true
		}
	}
	return false
}

func (e *emitter) constructionNeeded() bool {
	found := false
	walkExpressions(e.program, func(expression ir.Expression) {
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			found = found || constructionFields(value.Fields)
		case ir.ArrayLiteral:
			found = found || len(value.Metadata) != 0
		}
	})
	return found
}

func (e *emitter) nodeArrayLiteral(literal ir.ArrayLiteral) string {
	// Metadata is reserved, never evaluated as an invented initial value.
	for _, field := range literal.Metadata {
		if !field.Absent {
			panic("native: NodeArray metadata must be absent at allocation")
		}
	}
	array := e.own(ir.Array, fmt.Sprintf("adamic_node_array_new(%d, %t, &%s)", len(literal.Elements), literal.Element.IsReference(), e.shape(literal.Metadata)))
	for index, field := range literal.Metadata {
		e.line("adamic_object_field_types(%s->metadata)[%d] = %d;", array, index, field.Value.Type())
	}
	for index, element := range literal.Elements {
		value := e.value(element)
		if literal.Spread != nil && literal.Spread[index] {
			e.line("adamic_array_append(%s, %s);", array, value)
			continue
		}
		if literal.Element.IsReference() {
			value = retained(value)
		}
		e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(literal.Element), slotted(literal.Element, value))
	}
	return array
}

func (e *emitter) nodeArrayProperty(property ir.Property) string {
	array := e.value(property.Object)
	object := e.snapshot(ir.Object, "("+array+")->metadata")
	slot := fmt.Sprintf("adamic_object_read(%s, %s, &%s, %s)", object, cString(property.Name), e.cache(), cString(property.Readiness))
	if property.View != "" {
		value := e.temporary()
		e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", value, object, cString(property.Name), e.cache(), property.Of, cString(property.ViewType), cString(property.View))
		return e.snapshot(property.Of, unslotted(property.Of, value+"."+member(property.Of)))
	}
	value := unslotted(property.Of, slot+"->"+member(property.Of))
	if property.Of.IsReference() {
		return e.own(property.Of, "adamic_retain("+value+")")
	}
	return e.snapshot(property.Of, value)
}

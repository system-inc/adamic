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

func (nativeEmitter *emitter) constructionNeeded() bool {
	found := false
	walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			found = found || constructionFields(value.Fields)
		case ir.ArrayLiteral:
			found = found || len(value.Metadata) != 0
		}
	})
	return found
}

func (nativeEmitter *emitter) nodeArrayLiteral(literal ir.ArrayLiteral) string {
	// Metadata is reserved, never evaluated as an invented initial value.
	for _, field := range literal.Metadata {
		if !field.Absent {
			panic("native: NodeArray metadata must be absent at allocation")
		}
	}
	array := nativeEmitter.own(ir.Array, fmt.Sprintf(
		"adamic_node_array_make(%d, %s, &%s)",
		len(literal.Elements),
		arrayElement(literal.Element),
		nativeEmitter.shape(literal.Metadata),
	))
	for index, field := range literal.Metadata {
		nativeEmitter.line("adamic_object_field_types(%s->metadata)[%d] = %d;", array, index, field.Value.Type())
	}
	for index, element := range literal.Elements {
		value := nativeEmitter.value(element)
		if literal.Spread != nil && literal.Spread[index] {
			nativeEmitter.line("adamic_array_append(%s, %s);", array, value)
			continue
		}
		if literal.Element.IsReference() {
			value = retained(value)
		}
		nativeEmitter.line(
			"adamic_array_push(%s, (adamic_value){.%s = %s});",
			array,
			member(literal.Element),
			slotted(literal.Element, value),
		)
	}
	return array
}

func (nativeEmitter *emitter) nodeArrayProperty(property ir.Property) string {
	array := nativeEmitter.value(property.Object)
	object := nativeEmitter.snapshot(ir.Object, "("+array+")->metadata")
	slot := fmt.Sprintf(
		"adamic_object_read(%s, %s, &%s, %s)",
		object,
		cString(property.Name),
		nativeEmitter.cache(),
		cString(property.Readiness),
	)
	if property.View != "" {
		value := nativeEmitter.temporary()
		nativeEmitter.line(
			"adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);",
			value,
			object,
			cString(property.Name),
			nativeEmitter.cache(),
			property.Of,
			cString(property.ViewType),
			cString(property.View),
		)
		return nativeEmitter.snapshot(property.Of, unslotted(property.Of, value+"."+member(property.Of)))
	}
	value := unslotted(property.Of, slot+"->"+member(property.Of))
	if property.Of.IsReference() {
		return nativeEmitter.own(property.Of, "adamic_retain("+value+")")
	}
	return nativeEmitter.snapshot(property.Of, value)
}

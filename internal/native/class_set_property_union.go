package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A readonly widened view can observe a scalar slot as a union. Its actual shape
// decides how to box that slot; a field declared as a union already holds a box.
func (e *emitter) unionField(property ir.Property) string {
	object := e.snapshot(ir.Object, e.value(property.Object))
	slot := e.temporary()
	lookup := e.fieldSlot(object, property.Name, property.Class)
	if property.Absent {
		lookup = fmt.Sprintf("adamic_object_optional_field(%s, %s, &%s)", object, cString(property.Name), e.cache())
	}
	if property.Optional {
		lookup = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, lookup)
	}
	e.line("adamic_value *%s = %s;", slot, lookup)
	result := e.own(ir.Union, "NULL")
	e.line("if (%s != NULL) {", slot)
	e.indent++
	seen := map[string]bool{}
	walkExpressions(e.program, func(expression ir.Expression) {
		literal, ok := expression.(ir.ObjectLiteral)
		if !ok || literal.Spread != nil {
			return
		}
		for _, field := range literal.Fields {
			if field.Name != property.Name || field.Value.Type().IsReference() {
				continue
			}
			shape := e.literalShape(literal)
			if seen[shape] {
				continue
			}
			seen[shape] = true
			raw := unslotted(field.Value.Type(), fmt.Sprintf("%s->%s", slot, member(field.Value.Type())))
			boxed, fresh := converted(field.Value.Type(), ir.Union, raw)
			if !fresh {
				boxed = retained(boxed)
			}
			e.line("if (%s->shape == &%s) { %s = %s; } else", object, shape, result, boxed)
		}
	})
	e.line("{ %s = adamic_retain(%s->reference); }", result, slot)
	e.indent--
	e.line("}")
	return result
}

func (e *emitter) unionFieldKind(call ir.NumberCall) string {
	value := e.snapshot(ir.Union, e.value(call.Arguments[0]))
	to := ir.Type(call.Arguments[1].(ir.NumberConstant).Value)
	kind := "adamic_kind_object"
	switch to {
	case ir.Array:
		kind = "adamic_kind_array"
	case ir.Map:
		kind = "adamic_kind_map"
	default:
		if to.IsTypedArray() {
			kind = "adamic_kind_typed_array"
		}
	}
	return e.snapshot(ir.Boolean, fmt.Sprintf("(%s != NULL && %s->kind == %s)", value, value, kind))
}

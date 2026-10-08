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
	e.declarations = append(e.declarations, "const adamic_object *adamic_union_slot_owner(const adamic_object *, const adamic_value *);", "adamic_heap *adamic_union_runtime_field(const adamic_object *, const adamic_value *);")
	owner := e.temporary()
	e.line("const adamic_object *%s = adamic_union_slot_owner(%s, %s);", owner, object, slot)
	seen := map[string]bool{}
	emit := func(fields []ir.Field, shape string) {
		for _, field := range fields {
			if field.Name != property.Name || field.Value.Type().IsReference() || seen[shape] {
				continue
			}
			seen[shape] = true
			raw := unslotted(field.Value.Type(), fmt.Sprintf("%s->%s", slot, member(field.Value.Type())))
			boxed, fresh := converted(field.Value.Type(), ir.Union, raw)
			if !fresh {
				boxed = retained(boxed)
			}
			e.line("if (%s->shape == &%s) { %s = %s; } else", owner, shape, result, boxed)
		}
	}
	for _, class := range e.program.Classes {
		fields := class.PublicFields
		if !class.Literal {
			fields = nil
			for _, field := range class.Fields {
				if !field.Private {
					fields = append(fields, field)
				}
			}
		}
		emit(fields, e.publicClassShape(class))
	}
	walkExpressions(e.program, func(expression ir.Expression) {
		literal, ok := expression.(ir.ObjectLiteral)
		if !ok {
			return
		}
		if literal.Spread == nil {
			emit(literal.Fields, e.literalShape(literal))
		}
		if literal.SpreadMaybeUndefined {
			fields := emptyFields(literal)
			emit(fields, e.shape(fields))
		}
	})
	e.line("{ %s = adamic_union_runtime_field(%s, %s); }", result, owner, slot)
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
	}
	return e.snapshot(ir.Boolean, fmt.Sprintf("(%s != NULL && %s->kind == %s)", value, value, kind))
}

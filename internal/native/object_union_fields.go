package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A readonly scalar field may be viewed through a union field without copying its
// object. Box that view from its actual layout, while preserving reference identity.
func (e *emitter) objectUnionProperty(property ir.Property) string {
	object := e.snapshot(ir.Object, e.value(property.Object))
	e.declarations = append(e.declarations, `#include "object_spread_extend.h"`)
	fieldShape := e.temporary()
	e.line("const adamic_shape *%s = %s == NULL ? NULL : adamic_object_spread_field_shape(%s->shape, %s);", fieldShape, object, object, cString(property.Name))
	result := e.temporary()
	e.line("adamic_heap *%s = NULL;", result)
	if property.Optional {
		e.line("if (%s != NULL) {", object)
	}
	slot := e.temporary()
	lookup := e.fieldSlot(object, property.Name, property.Class)
	if property.Absent {
		lookup = fmt.Sprintf("adamic_object_optional_field(%s, %s, &%s)", object, cString(property.Name), e.cache())
	}
	e.line("adamic_value *%s = %s;", slot, lookup)
	if property.Absent {
		e.line("if (%s != NULL) {", slot)
	}
	seen := map[string]bool{}
	first := true
	register := func(fields []ir.Field, shape string) {
		for _, field := range fields {
			if field.Name != property.Name {
				continue
			}
			if seen[shape] {
				return
			}
			seen[shape] = true
			prefix := "if"
			if !first {
				prefix = "else if"
			}
			first = false
			e.line("%s (%s == &%s) {", prefix, fieldShape, shape)
			raw := unslotted(field.Value.Type(), slot+"->"+member(field.Value.Type()))
			boxed, fresh := converted(field.Value.Type(), ir.Union, raw)
			if !fresh {
				boxed = retained(boxed)
			}
			e.line("%s = %s;", result, boxed)
			e.line("}")
			return
		}
	}
	walkExpressions(e.program, func(expression ir.Expression) {
		literal, ok := expression.(ir.ObjectLiteral)
		if !ok {
			return
		}
		shape := e.literalShape(literal)
		if literal.Spread != nil {
			shape = e.shape(literal.Fields)
		}
		register(literal.Fields, shape)
	})
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
		register(fields, e.publicClassShape(class))
	}
	if first {
		e.line("%s = adamic_retain(%s->reference);", result, slot)
	} else {
		e.line("else { %s = adamic_retain(%s->reference); }", result, slot)
	}
	if property.Absent {
		e.line("}")
	}
	if property.Optional {
		e.line("}")
	}
	e.owned = append(e.owned, result)
	return result
}

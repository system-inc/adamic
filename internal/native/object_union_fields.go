package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A readonly scalar field may be viewed through a union field without copying its
// object. Box that view from its actual layout, while preserving reference identity.
func (e *emitter) objectUnionProperty(property ir.Property) string {
	object := e.snapshot(ir.Object, e.value(property.Object))
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
	walkExpressions(e.program, func(expression ir.Expression) {
		literal, ok := expression.(ir.ObjectLiteral)
		if !ok || literal.Spread != nil {
			return
		}
		for _, field := range literal.Fields {
			if field.Name != property.Name {
				continue
			}
			shape := e.literalShape(literal)
			if seen[shape] {
				return
			}
			seen[shape] = true
			prefix := "if"
			if !first {
				prefix = "else if"
			}
			first = false
			e.line("%s (%s->shape == &%s) {", prefix, object, shape)
			raw := unslotted(field.Value.Type(), slot+"->"+member(field.Value.Type()))
			boxed, fresh := converted(field.Value.Type(), ir.Union, raw)
			if !fresh {
				boxed = retained(boxed)
			}
			e.line("%s = %s;", result, boxed)
			e.line("}")
			return
		}
	})
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

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nullishViewField(property ir.Property) string {
	e.declarations = append(e.declarations, "#include \"view_nullish.h\"")
	value := e.own(ir.Union, fmt.Sprintf("adamic_object_nullish_view(%s, %s, &%s, %d, %t, %t, %t, %t, %s, %s)", e.value(property.Object), cString(property.Name), e.cache(), property.NullishKinds, property.NullAllowed, property.UndefinedAllowed, property.Absent, property.Optional, cString(property.View), cString(property.ViewType)))
	if len(property.ViewAllowed) != 0 {
		var tests []string
		for _, literal := range property.ViewAllowed {
			kind := map[ir.Type]string{ir.Number: "adamic_kind_number", ir.Boolean: "adamic_kind_boolean", ir.String: "adamic_kind_string"}[literal.Type()]
			actual := func() string { converted, _ := converted(ir.Union, literal.Type(), value); return converted }()
			tests = append(tests, fmt.Sprintf("(%s->kind == %s && %s)", value, kind, e.binary(ir.Equal, literal.Type(), actual, e.value(literal))))
		}
		e.line("if (%s != NULL && %s != &adamic_null && !(%s)) adamic_nullish_failure(%s, %s, %s);", value, value, strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), value)
	}
	if property.Of == ir.Union {
		return value
	}
	narrowed := fmt.Sprintf("((%s)%s)", cType(property.Of), value)
	if property.Of.IsMaybe() {
		narrowed = e.narrowValue(value, property.Of)
	}
	if property.NullAllowed && property.Of.IsReference() {
		narrowed = fmt.Sprintf("(%s == &adamic_null ? NULL : %s)", value, narrowed)
	}
	if property.Of == ir.MaybeNumber || property.Of == ir.MaybeBoolean {
		return narrowed
	}
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("adamic_retain(%s)", narrowed))
	}
	return narrowed
}

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
	e.nullishMemberSelection(property, value)
	e.mapViewCertificate(property, value)
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

func (e *emitter) nullishMemberSelection(property ir.Property, value string) {
	if property.ViewContract == 0 {
		return
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind == ir.ViewNullable {
		if contract.Element == 0 {
			return
		}
		property.ViewContract = contract.Element
		contract = e.program.ViewContracts[contract.Element-1]
	}
	if contract.Kind != ir.ViewUnion {
		return
	}
	e.line("if (%s != NULL && %s != &adamic_null) {", value, value)
	if contract.Of == ir.Object {
		e.viewObjectUnion(property, "(adamic_object *)"+value)
	} else {
		tests := []string{}
		for _, id := range contract.Members {
			member := e.program.ViewContracts[id-1]
			if member.Kind == ir.ViewUndefined || member.Kind == ir.ViewNull {
				continue
			}
			kind := map[ir.Type]string{ir.Number: "adamic_kind_number", ir.Boolean: "adamic_kind_boolean", ir.String: "adamic_kind_string", ir.Object: "adamic_kind_object", ir.Array: "adamic_kind_array", ir.Map: "adamic_kind_map"}[member.Of]
			if kind == "" {
				panic("compiler bug: unavailable nullable union member")
			}
			allowed := []string{}
			for _, literal := range member.Allowed {
				var constant ir.Expression
				switch literal.Of {
				case ir.Number:
					constant = ir.NumberConstant{Value: literal.Number}
				case ir.Boolean:
					constant = ir.BooleanConstant{Value: literal.Boolean}
				case ir.String:
					name := e.temporary()
					e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
					allowed = append(allowed, fmt.Sprintf("adamic_string_equal((const adamic_string *)%s, &%s)", value, name))
					continue
				}
				actual, _ := converted(ir.Union, member.Of, value)
				allowed = append(allowed, e.binary(ir.Equal, member.Of, actual, e.value(constant)))
			}
			test := value + "->kind == " + kind
			if len(allowed) != 0 {
				test += " && (" + strings.Join(allowed, " || ") + ")"
			}
			tests = append(tests, "("+test+")")
		}
		e.line("if (!(%s)) adamic_nullish_failure(%s, %s, %s);", strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), value)
		for _, id := range contract.Members {
			member := e.program.ViewContracts[id-1]
			if member.Kind == ir.ViewMap {
				mapped := property
				mapped.ViewContract = id
				e.line("if(%s->kind==adamic_kind_map){", value)
				e.mapViewCertificate(mapped, value)
				e.line("}")
			}
		}

	}
	e.line("}")
}

package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// caughtProperty reads a dynamic own field. Generated layouts prove scalar
// representations; the runtime handles reference fields and absent fields.
func (e *emitter) caughtProperty(call ir.ObjectCall, arguments []string) string {
	name := e.program.Strings[call.Arguments[1].(ir.StringConstant).Index]
	object := arguments[0]
	result := e.own(ir.Union, "NULL")
	handled := e.snapshot(ir.Boolean, "false")
	seen := map[string]bool{}
	walkExpressions(e.program, func(expression ir.Expression) {
		literal, ok := expression.(ir.ObjectLiteral)
		if !ok {
			return
		}
		for index, field := range literal.Fields {
			_, literalNull := field.Value.(ir.Null)
			if field.Name != name || (field.Value.Type().IsReference() && !field.Null && !literalNull) {
				continue
			}
			shape := e.literalShape(literal)
			if seen[shape] {
				continue
			}
			seen[shape] = true
			of := field.Value.Type()
			value := unslotted(of, fmt.Sprintf("((adamic_object *)%s)->slots[%d].%s", object, index, member(of)))
			boxed, _ := converted(of, ir.Union, value)
			if field.Null || literalNull {
				boxed = fmt.Sprintf("(%s == NULL ? &adamic_null : (adamic_heap *)adamic_retain(%s))", value, value)
			}
			e.line("if (%s != NULL && %s->kind == adamic_kind_object && ((adamic_object *)%s)->shape == &%s) {", object, object, object, shape)
			e.indent++
			e.line("%s = %s;", result, boxed)
			e.line("%s = true;", handled)
			e.indent--
			e.line("}")
		}
	})
	e.line("if (!%s) %s = adamic_caught_property(%s, %s);", handled, result, object, arguments[1])
	e.checkThrown()
	return result
}

func (e *emitter) checkCaughtType(narrow ir.Narrow, value string) {
	kind := map[ir.Type]string{ir.String: "string", ir.Number: "number", ir.Boolean: "boolean", ir.Object: "object", ir.Array: "array", ir.Map: "map", ir.Closure: "closure"}[narrow.To.Present()]
	if kind == "" {
		panic("native: unknown checked catch representation")
	}
	condition := fmt.Sprintf("(%s == NULL || %s->kind != adamic_kind_%s)", value, value, kind)
	if narrow.Optional {
		condition = fmt.Sprintf("(%s != NULL && %s->kind != adamic_kind_%s)", value, value, kind)
	}
	message := "adamic/catch-type: caught value does not match its typed use"
	e.line("if %s adamic_panic(%s, %d);", condition, cString(message), len(message))
}

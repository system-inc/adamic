package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// ordinaryConversionBody bridges verified, non-accessor method signatures to the
// runtime protocol. The callback's regular IR helper keeps dispatch, this, packed
// return values and exception cleanup identical to every other method call.
// Accessors stay refused, so lookup in the typed callback is observationally
// identical to Get followed immediately by Call; nothing can run between them.
func (e *emitter) ordinaryConversionBody(function ir.Function) {
	protocol := function.OrdinaryConversion
	e.declarations = append(e.declarations, "#include <string.h>")
	prefix := fmt.Sprintf("adamic_primitive_%d", e.functionIndex)
	receiverType := e.program.Locals[function.Parameters[0]].Type
	getter := []string{fmt.Sprintf("static adamic_primitive_method %s_get(adamic_heap *receiver, const char *name) {", prefix), "\t(void)receiver;"}
	for _, name := range []string{"valueOf", "toString"} {
		method, exists := protocol.Methods[name]
		if !exists || method.Function == 0 && method.Builtin == "" {
			continue
		}
		callback := prefix + "_" + name
		lines := []string{fmt.Sprintf("static adamic_primitive %s(adamic_heap *receiver, void *owner) {", callback), "\t(void)owner;"}
		if method.Function == 0 {
			if method.Builtin == "valueOf" {
				lines = append(lines, "\treturn (adamic_primitive){adamic_primitive_object, {.reference = adamic_retain(receiver)}};")
			} else {
				lines = append(lines, "\t(void)receiver;", "\tstatic adamic_string text = ADAMIC_STRING(\"[object Object]\");", "\treturn (adamic_primitive){adamic_primitive_string, {.reference = &text}};")
			}
		} else {
			target := method.Function - 1
			returns := e.program.Functions[target].Returns
			receiver := "receiver"
			if e.reuse.consumed[e.program.Functions[target].Parameters[0]] {
				receiver = "adamic_retain(receiver)"
			}
			call := fmt.Sprintf("%s((%s)%s)", e.functionName(target), cType(receiverType), receiver)
			if returns == 0 {
				lines = append(lines, "\t"+call+";", "\treturn (adamic_primitive){adamic_primitive_undefined, {.reference = NULL}};")
			} else {
				lines = append(lines, fmt.Sprintf("\t%s value = %s;", cType(returns), call), "\tif (adamic_thrown != NULL) return (adamic_primitive){adamic_primitive_undefined, {.reference = NULL}};")
				lines = append(lines, primitiveResult(returns, method)...)
			}
		}
		lines = append(lines, "}")
		e.declarations = append(e.declarations, strings.Join(lines, "\n"))
		getter = append(getter, fmt.Sprintf("\tif (strcmp(name, %s) == 0) return (adamic_primitive_method){%s, NULL};", cString(name), callback))
	}
	getter = append(getter, "\t(void)name;", "\treturn (adamic_primitive_method){NULL, NULL};", "}")
	e.declarations = append(e.declarations, strings.Join(getter, "\n"))
	hint := "adamic_hint_default"
	if protocol.StringHint {
		hint = "adamic_hint_string"
	}
	e.line("adamic_primitive primitive = adamic_ordinary_to_primitive((adamic_heap *)%s, %s, %s_get);", e.localName(function.Parameters[0]), hint, prefix)
	e.checkThrown()
	e.line("adamic_string *text;")
	e.line("switch (primitive.kind) {")
	e.line("case adamic_primitive_undefined: text = &adamic_string_undefined; break;")
	e.line("case adamic_primitive_null: { static adamic_string null_text = ADAMIC_STRING(\"null\"); text = &null_text; break; }")
	e.line("case adamic_primitive_boolean: text = primitive.value.boolean ? &adamic_string_true : &adamic_string_false; break;")
	e.line("case adamic_primitive_number: text = adamic_string_from_number(primitive.value.number); break;")
	e.line("case adamic_primitive_string: text = primitive.value.reference; break;")
	e.line("default: adamic_unreachable();")
	e.line("}")
	e.releaseScopes(e.functionDepth)
	e.line("return text;")
}

// primitiveResult transfers string/object ownership, and unwraps scalar boxes
// before releasing them. NULL and the immortal null sentinel stay distinct.
func primitiveResult(of ir.Type, method ir.PrimitiveMethod) []string {
	result := func(kind, member, value string) string {
		return fmt.Sprintf("\treturn (adamic_primitive){adamic_primitive_%s, {.%s = %s}};", kind, member, value)
	}
	if method.Null || method.Undefined {
		return []string{"\t(void)value;", result(map[bool]string{true: "null", false: "undefined"}[method.Null], "reference", "NULL")}
	}
	if of.IsReference() && of != ir.Union && (method.NullAbsent || method.UndefinedAbsent) {
		kind := "undefined"
		if method.NullAbsent {
			kind = "null"
		}
		prefix := "\tif (value == NULL) return (adamic_primitive){adamic_primitive_" + kind + ", {.reference = NULL}};"
		if of == ir.String {
			return []string{prefix, result("string", "reference", "value")}
		}
		return []string{prefix, result("object", "reference", "value")}
	}
	switch of {
	case ir.Number:
		return []string{result("number", "number", "value")}
	case ir.Boolean:
		return []string{result("boolean", "boolean", "value")}
	case ir.MaybeNumber, ir.MaybeBoolean:
		kind, member := "number", "number"
		if of == ir.MaybeBoolean {
			kind, member = "boolean", "boolean"
		}
		return []string{"\tif (!value.present) return (adamic_primitive){adamic_primitive_undefined, {.reference = NULL}};", result(kind, member, "value."+member)}
	case ir.String:
		return []string{"\tif (value == NULL) return (adamic_primitive){adamic_primitive_undefined, {.reference = NULL}};", result("string", "reference", "value")}
	case ir.Union:
		return []string{
			"\tif (value == NULL) return (adamic_primitive){adamic_primitive_undefined, {.reference = NULL}};",
			"\tif (value == &adamic_null) return (adamic_primitive){adamic_primitive_null, {.reference = NULL}};",
			"\tif (value->kind == adamic_kind_number) { double number = ((adamic_number_box *)value)->number; adamic_release(value); return (adamic_primitive){adamic_primitive_number, {.number = number}}; }",
			"\tif (value->kind == adamic_kind_boolean) { bool boolean = ((adamic_boolean_box *)value)->boolean; adamic_release(value); return (adamic_primitive){adamic_primitive_boolean, {.boolean = boolean}}; }",
			"\tif (value->kind == adamic_kind_string) return (adamic_primitive){adamic_primitive_string, {.reference = value}};",
			result("object", "reference", "value"),
		}
	default:
		return []string{result("object", "reference", "value")}
	}
}

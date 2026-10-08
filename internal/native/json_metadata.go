package native

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// Slot descriptions follow the IR representation, never the ownership bitmap.
func jsonStorage(t ir.Type) string {
	switch t {
	case 0:
		return "&adamic_json_undefined_schema"
	case ir.Number:
		return "&adamic_json_number_schema"
	case ir.Boolean:
		return "&adamic_json_boolean_schema"
	case ir.String:
		return "&adamic_json_string_schema"
	case ir.MaybeNumber:
		return "&adamic_json_maybe_number_schema"
	case ir.Object, ir.Array, ir.Map, ir.Closure, ir.Union:
		return "&adamic_json_union_schema"
	default:
		return "NULL"
	}
}

func (e *emitter) shapeJSON(name string, names []string, types []ir.Type, private []bool, tuple bool, hook bool) string {
	e.declarations = append(e.declarations, `#include "json_metadata.h"`)
	indices := []int{}
	for i, field := range names {
		if field == "toJSON" {
			return "NULL"
		}
		if i < len(private) && private[i] {
			continue
		}
		if jsonStorage(types[i]) == "NULL" {
			return "NULL"
		}
		indices = append(indices, i)
	}
	index := func(name string) (uint64, bool) {
		n, err := strconv.ParseUint(name, 10, 32)
		return n, err == nil && n < 4294967295 && strconv.FormatUint(n, 10) == name
	}
	sort.SliceStable(indices, func(a, b int) bool {
		x, xi := index(names[indices[a]])
		y, yi := index(names[indices[b]])
		if xi != yi {
			return xi
		}
		return xi && x < y
	})
	fields := []string{}
	for _, i := range indices {
		key := fmt.Sprintf("%s_json_key_%d", name, i)
		e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", key, cString(names[i])))
		fields = append(fields, fmt.Sprintf("{&%s, %d, %s}", key, i, jsonStorage(types[i])))
	}
	list := "NULL"
	if len(fields) > 0 {
		list = name + "_json_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_field %s[] = {%s};", list, strings.Join(fields, ", ")))
	}
	kind := "object"
	if hook {
		kind = "toJSON"
	}
	if tuple {
		kind = "tuple"
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_json_schema %s_json = {.kind = adamic_json_%s, .count = %d, .fields = %s};", name, kind, len(fields), list))
	return "&" + name + "_json"
}

func (e *emitter) jsonHook(function int) string {
	if name, ok := e.jsonHooks[function]; ok {
		return name
	}
	method := e.program.Functions[function]
	if method.Closure || method.Async || len(method.Parameters) == 0 || e.program.Locals[method.Parameters[0]].Type != ir.Object {
		return "NULL"
	}
	// A receiver plus optionally the key. Other calling conventions remain unproven.
	if len(method.Parameters) < 1 || len(method.Parameters) > 2 {
		return "NULL"
	}
	if len(method.Parameters) == 2 && e.program.Locals[method.Parameters[1]].Type != ir.String {
		return "NULL"
	}
	if method.Returns != ir.MaybeBoolean && jsonStorage(method.Returns) == "NULL" {
		return "NULL"
	}
	name := fmt.Sprintf("adamic_json_hook_%d", function)
	if e.jsonHooks == nil {
		e.jsonHooks = map[int]string{}
	}
	e.jsonHooks[function] = name
	args := []string{"self"}
	if e.reuse.consumed[method.Parameters[0]] {
		args[0] = "adamic_retain(self)"
	}
	if len(method.Parameters) == 2 {
		key := "(adamic_string *)key"
		if e.reuse.consumed[method.Parameters[1]] {
			key = "adamic_retain(" + key + ")"
		}
		args = append(args, key)
	}
	call := e.functionName(function) + "(" + strings.Join(args, ", ") + ")"
	lines := []string{fmt.Sprintf("static adamic_json_result %s(adamic_object *self, const adamic_string *key) {", name), " (void)key;"}
	if method.Returns == 0 {
		lines = append(lines, " "+call+";", " return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};")
	} else if method.Returns == ir.MaybeBoolean {
		lines = append(lines, " adamic_maybe_boolean result = "+call+";", " if (adamic_thrown != NULL || !result.present) return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};", " return (adamic_json_result){{.boolean = result.boolean}, &adamic_json_boolean_schema};")
	} else {
		lines = append(lines, fmt.Sprintf(" adamic_value result = {.%s = %s};", member(method.Returns), slotted(method.Returns, call)))
		if method.Returns.IsReference() {
			lines = append(lines, " if (adamic_thrown != NULL) { adamic_release(result.reference); return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema}; }")
		} else {
			lines = append(lines, " if (adamic_thrown != NULL) return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};")
		}
		if method.Returns.IsReference() {
			lines = append(lines, " if (result.reference == NULL) return (adamic_json_result){{.reference = NULL}, &adamic_json_undefined_schema};")
		}
		lines = append(lines, fmt.Sprintf(" return (adamic_json_result){result, %s};", jsonStorage(method.Returns)))
	}
	lines = append(lines, "}")
	e.declarations = append(e.declarations, `#include "json_metadata.h"`, strings.Join(lines, "\n"))
	return name
}

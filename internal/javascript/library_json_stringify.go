package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) jsonReadSchema(schema *ir.JSONSchema) string {
	if schema == nil {
		return "undefined"
	}
	check := "undefined"
	if schema.ArrayRead.Element != 0 && ir.HasArrayViews(e.program) {
		check = e.viewArrayChecker(schema.ArrayRead)
	}
	fields := []string{}
	for _, field := range schema.Fields {
		fields = append(fields, quote(e.program.Strings[field.Name])+":"+e.jsonReadSchema(field.Schema))
	}
	return "{check:" + check + ",element:" + e.jsonReadSchema(schema.Element) + ",fields:{" + strings.Join(fields, ",") + "}}"
}

// Normalize only after all three arguments have been evaluated. Copying an array earlier would
// hide mutations made by the space argument. A closure's C-like backend wrapper is represented as
// an actual function for JSON, which omits it from objects and writes null in arrays.
const jsonStringifyRuntime = `const adamicJSONValue = (value, schema) => {
 if (value instanceof AdamicClosure) return () => undefined;
 if (Array.isArray(value)) return value.map((item,index) => adamicJSONValue(schema?.check ? schema.check(item) : item,schema?.element || schema?.fields?.[String(index)]));
 if (value !== null && typeof value === 'object' && !(value instanceof Map) && !(value instanceof Set)) {
  const copy = {};
  for (const key of Object.keys(value)) Object.defineProperty(copy, key, {value: adamicJSONValue(value[key],schema?.fields?.[key]), enumerable: true, configurable: true, writable: true});
  return copy;
 }
 return value;
};
const adamicJSONStringify = (value, replacer, space, schema, replacerSchema) => { const keys=adamicJSONValue(replacer,replacerSchema); return JSON.stringify(adamicJSONValue(value,schema), keys, space); };
`

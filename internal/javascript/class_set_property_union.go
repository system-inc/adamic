package javascript

import "github.com/system-inc/adamic/internal/ir"

// Only object-like union members need a tag beyond JavaScript's typeof.
func (e *emitter) unionFieldKind(call ir.NumberCall) string {
	value := e.value(call.Arguments[0])
	to := ir.Type(call.Arguments[1].(ir.NumberConstant).Value)
	test := "(v !== null && typeof v === 'object' && !Array.isArray(v) && !(v instanceof Map) && !(v instanceof Set) && !ArrayBuffer.isView(v))"
	switch to {
	case ir.Array:
		test = "Array.isArray(v)"
	case ir.Map:
		test = "(v instanceof Map || v instanceof Set)"
	}
	return "((v) => " + test + ")(" + value + ")"
}

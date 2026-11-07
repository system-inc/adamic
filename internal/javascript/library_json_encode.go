package javascript

import "github.com/system-inc/adamic/internal/ir"

func (javascriptEmitter *emitter) jsonEncode(expression ir.JSONEncode) string {
	return "encodeJson(" +
		javascriptEmitter.value(expression.Value) +
		", " +
		javascriptEmitter.jsonDescriptor(expression.Schema) +
		")"
}

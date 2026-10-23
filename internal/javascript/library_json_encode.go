package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) jsonEncode(expression ir.JSONEncode) string {
	return "encodeJson(" + e.value(expression.Value) + ", " + e.jsonDescriptor(expression.Schema) + ")"
}

package javascript

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonEncode(expression ir.JSONEncode) string {
	descriptor, err := json.Marshal(expression.Schema)
	if err != nil {
		panic(err)
	}
	return "encodeJson(" + e.value(expression.Value) + ", " + string(descriptor) + ")"
}

package javascript

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonDecode(expression ir.JSONDecode) string {
	descriptor, err := json.Marshal(expression.Schema)
	if err != nil {
		panic(err)
	}
	return "decodeJson(" + e.value(expression.Text) + ", " + string(descriptor) + ")"
}

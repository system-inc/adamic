package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonEncode(expression ir.JSONEncode) string {
	e.declarations = append(e.declarations, `#include "json_encode.h"`)
	schema := e.jsonDecodeSchema(expression.Schema)
	value := e.jsonSlot(expression.Value)
	return e.own(ir.String, fmt.Sprintf("adamic_json_encode(%s, &%s)", value, schema))
}

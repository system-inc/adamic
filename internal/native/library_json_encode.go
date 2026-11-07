package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) jsonEncode(expression ir.JSONEncode) string {
	nativeEmitter.declarations = append(nativeEmitter.declarations, `#include "json_encode.h"`)
	schema := nativeEmitter.jsonDataSchema(expression.Schema, true)
	value := nativeEmitter.jsonSlot(expression.Value)
	return nativeEmitter.own(ir.String, fmt.Sprintf("adamic_json_encode(%s, &%s)", value, schema))
}

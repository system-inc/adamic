package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) processCall(call ir.ProcessCall) string {
	if value, known := e.nodeProcessCall(call); known {
		return value
	}
	panic(fmt.Sprintf("native: unknown process operation %s", call.Operation))
}

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) processCall(call ir.ProcessCall) string {
	if value, known := e.nodeProcessCall(call); known {
		return value
	}
	switch call.Operation {
	case "stdoutTTY":
		return e.snapshot(ir.MaybeBoolean, "adamic_process_is_tty(adamic_stdout)")
	case "stderrTTY":
		return e.snapshot(ir.MaybeBoolean, "adamic_process_is_tty(adamic_stderr)")
	case "env":
		return e.own(ir.String, "adamic_process_environment("+e.value(call.Arguments[0])+")")

	}
	panic(fmt.Sprintf("native: unknown process operation %s", call.Operation))
}

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) processCall(call ir.ProcessCall) string {
	if value, known := nativeEmitter.nodeProcessCall(call); known {
		return value
	}
	switch call.Operation {
	case "exitCode":
		return nativeEmitter.snapshot(ir.MaybeNumber, "adamic_process_exit_code()")

	case "setExitCode":
		code := nativeEmitter.snapshot(ir.MaybeNumber, nativeEmitter.value(call.Arguments[0]))
		nativeEmitter.line("adamic_process_set_exit_code(%s);", code)
		nativeEmitter.checkThrown()
		return code
	case "exit":
		nativeEmitter.line("adamic_process_exit(%s);", nativeEmitter.value(call.Arguments[0]))
		nativeEmitter.checkThrown()
		return "0.0"
	}
	panic(fmt.Sprintf("native: unknown process operation %s", call.Operation))
}

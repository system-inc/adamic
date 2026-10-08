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
	case "exitCode":
		e.hostHelper("exitCode")
		return e.snapshot(ir.MaybeNumber, "adamic_library_exit_code()")
	case "stdoutTTY":
		return e.snapshot(ir.MaybeBoolean, "adamic_process_is_tty(adamic_stdout)")
	case "stderrTTY":
		return e.snapshot(ir.MaybeBoolean, "adamic_process_is_tty(adamic_stderr)")
	case "env":
		return e.own(ir.String, "adamic_host_environment("+e.value(call.Arguments[0])+")")
	case "setExitCode":
		e.hostHelper("setExitCode")
		code := e.snapshot(ir.MaybeNumber, e.value(call.Arguments[0]))
		e.line("adamic_library_set_exit_code(%s);", code)
		e.checkThrown()
		return code
	case "exit":
		e.hostHelper("setExitCode")
		e.line("adamic_library_set_exit_code(%s);", e.value(call.Arguments[0]))
		e.checkThrown()
		e.line("adamic_process_exit_now(adamic_host_exit_status());")
		return "0.0"
	}
	panic(fmt.Sprintf("native: unknown process operation %s", call.Operation))
}

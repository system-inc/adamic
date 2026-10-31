package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	if call.Operation == "tmpdir" {
		return "adamicNodeTmpdir()", true
	}
	values := map[string]string{"performance": "performance", "nextTickFeature": "!!process.nextTick", "eol": "\"\\n\"", "handle": "process.stdout._handle", "cwd": "process.cwd()", "platform": "process.platform", "pid": "process.pid", "argv": "process.argv", "execArgv": "process.execArgv", "memoryUsage": "process.memoryUsage()", "now": "performance.now()", "timeOrigin": "performance.timeOrigin", "stackTraceLimit": "Error.stackTraceLimit"}
	if value, known := values[call.Operation]; known {
		return value, true
	}
	if call.Operation == "setStackTraceLimit" {
		return "(Error.stackTraceLimit = " + e.value(call.Arguments[0]) + ")", true
	}
	methods := map[string]string{"chdir": "process.chdir", "stdoutWrite": "process.stdout.write", "mark": "performance.mark", "measure": "performance.measure", "clearMarks": "performance.clearMarks", "clearMeasures": "performance.clearMeasures"}
	if method, known := methods[call.Operation]; known {
		return method + "(" + e.values(call.Arguments) + ")", true
	}
	return "", false
}

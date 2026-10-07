package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	values := map[string]string{"nextTickFeature": "!!process.nextTick", "eol": "\"\\n\"", "handle": "process.stdout._handle", "cwd": "process.cwd()", "platform": "process.platform", "pid": "process.pid", "argv": "process.argv", "execArgv": "process.execArgv", "columns": "process.stdout.columns", "memoryUsage": "process.memoryUsage()", "now": "performance.now()", "timeOrigin": "performance.timeOrigin"}
	if value, known := values[call.Operation]; known {
		return value, true
	}
	methods := map[string]string{"stdoutWrite": "process.stdout.write", "mark": "performance.mark", "measure": "performance.measure", "clearMarks": "performance.clearMarks", "clearMeasures": "performance.clearMeasures"}
	if method, known := methods[call.Operation]; known {
		return method + "(" + e.values(call.Arguments) + ")", true
	}
	return "", false
}

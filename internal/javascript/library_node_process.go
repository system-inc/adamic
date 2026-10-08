package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	if call.Operation == "osPlatform" {
		return "adamicNodeOSPlatform()", true
	}
	if call.Operation == "homedir" {
		return "adamicNodeHomedir()", true
	}
	if call.Operation == "tmpdir" {
		return "adamicNodeTmpdir()", true
	}
	if call.Operation == "envSet" {
		return "(process.env[" + e.value(call.Arguments[0]) + "] = " + e.value(call.Arguments[1]) + ")", true
	}
	if call.Operation == "envDelete" {
		return "delete process.env[" + e.value(call.Arguments[0]) + "]", true
	}
	values := map[string]string{"performance": "performance", "nextTickFeature": "!!process.nextTick", "eol": "\"\\n\"", "handle": "process.stdout._handle", "cwd": "process.cwd()", "platform": "process.platform", "pid": "process.pid", "argv": "process.argv", "execArgv": "process.execArgv", "columns": "process.stdout.columns", "memoryUsage": "process.memoryUsage()", "now": "performance.now()", "timeOrigin": "performance.timeOrigin"}
	if value, known := values[call.Operation]; known {
		return value, true
	}
	methods := map[string]string{"chdir": "process.chdir", "stdoutWrite": "process.stdout.write", "mark": "performance.mark", "measure": "performance.measure", "clearMarks": "performance.clearMarks", "clearMeasures": "performance.clearMeasures"}
	if method, known := methods[call.Operation]; known {
		return method + "(" + e.values(call.Arguments) + ")", true
	}
	return "", false
}

package javascript

import "github.com/system-inc/adamic/internal/ir"

// process.memoryUsage()'s result has Object's prototype on Node, but native has no honest answer yet
// for four of its five fields, so a run-time key read stops on it as on a host object (Adamic's
// ruling on #ey7aysy). It keeps Node's own fields for lowering's static read of heapUsed, under a
// prototype that isn't Object's, so adamicDynamicGet's host test (dynamic_keyed.go) stops on it as
// native's host_object does. The helper takes the result rather than calling process itself, since
// every module carries it and one for Workers may name no Node global (workers_test.go).
const hostMemoryUsageRuntime = "class AdamicHostObject {}\nconst adamicHostMemoryUsage = usage => Object.setPrototypeOf(usage, AdamicHostObject.prototype);\n"

// process.stdout._handle is native's adamic_node_stdout_handle: absent where Node has no handle, and
// otherwise a host object whose one member, setBlocking, is a closure. Node's Pipe or TTY handle holds
// setBlocking as a bare host function, and every call road here calls a closure's code: the method
// call's adamicCallee, and tsc's sys.ts guard, which reads the function-typed field and hands it to
// adamicCall. Each reads the member from this value, so it's made a closure here, once (#qa1g5f3).
// It stays a host value for a key read, as Node's Pipe is.
const hostStdoutHandleRuntime = "const adamicHostStdoutHandle = handle => handle === undefined ? undefined : Object.setPrototypeOf({ setBlocking: new AdamicClosure((closure, values) => handle.setBlocking(...values), []) }, AdamicHostObject.prototype);\n"

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	if call.Operation == "tmpdir" {
		return "adamicNodeTmpdir()", true
	}
	values := map[string]string{"performance": "performance", "nextTickFeature": "!!process.nextTick", "eol": "\"\\n\"", "handle": "adamicHostStdoutHandle(process.stdout._handle)", "cwd": "process.cwd()", "platform": "process.platform", "pid": "process.pid", "argv": "process.argv", "execArgv": "process.execArgv", "memoryUsage": "adamicHostMemoryUsage(process.memoryUsage())", "now": "performance.now()", "timeOrigin": "performance.timeOrigin", "stackTraceLimit": "Error.stackTraceLimit"}
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

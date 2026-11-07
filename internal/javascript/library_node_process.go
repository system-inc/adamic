package javascript

import "github.com/system-inc/adamic/internal/ir"

// process.memoryUsage()'s result has Object's prototype on Node, but native has no honest answer yet
// for four of its five fields, so a run-time key read stops on it as on a host object (Adamic's
// ruling on #ey7aysy). It keeps Node's own fields for lowering's static read of heapUsed, under a
// prototype that isn't Object's, so adamicDynamicGet's host test (dynamic_keyed.go) stops on it as
// native's host_object does. The helper takes the result rather than calling process itself, since
// every module carries it and one for Workers may name no Node global (workers_test.go).
const hostMemoryUsageRuntime = "class AdamicHostObject {}\nconst adamicHostMemoryUsage = usage => Object.setPrototypeOf(usage, AdamicHostObject.prototype);\n"

func (javascriptEmitter *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	if call.Operation == "tmpdir" {
		return "adamicNodeTmpdir()", true
	}
	values := map[string]string{
		"performance":     "performance",
		"nextTickFeature": "!!process.nextTick",
		"eol":             "\"\\n\"",
		"handle":          "process.stdout._handle",
		"cwd":             "process.cwd()",
		"platform":        "process.platform",
		"pid":             "process.pid",
		"argv":            "process.argv",
		"execArgv":        "process.execArgv",
		"memoryUsage":     "adamicHostMemoryUsage(process.memoryUsage())",
		"now":             "performance.now()",
		"timeOrigin":      "performance.timeOrigin",
		"stackTraceLimit": "Error.stackTraceLimit",
	}
	if value, known := values[call.Operation]; known {
		return value, true
	}
	if call.Operation == "setStackTraceLimit" {
		return "(Error.stackTraceLimit = " + javascriptEmitter.value(call.Arguments[0]) + ")", true
	}
	methods := map[string]string{
		"chdir":         "process.chdir",
		"stdoutWrite":   "process.stdout.write",
		"mark":          "performance.mark",
		"measure":       "performance.measure",
		"clearMarks":    "performance.clearMarks",
		"clearMeasures": "performance.clearMeasures",
	}
	if method, known := methods[call.Operation]; known {
		return method + "(" + javascriptEmitter.values(call.Arguments) + ")", true
	}
	return "", false
}

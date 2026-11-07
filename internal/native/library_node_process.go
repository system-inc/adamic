package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	functions := map[string]string{
		"tmpdir":             "adamic_node_tmpdir",
		"performance":        "adamic_node_performance",
		"chdir":              "adamic_node_chdir",
		"nextTickFeature":    "adamic_node_next_tick_feature",
		"eol":                "adamic_node_eol",
		"handle":             "adamic_node_stdout_handle",
		"cwd":                "adamic_node_cwd",
		"platform":           "adamic_node_platform",
		"pid":                "adamic_node_pid",
		"argv":               "adamic_node_argv",
		"execArgv":           "adamic_node_exec_argv",
		"stdoutWrite":        "adamic_node_stdout_write",
		"memoryUsage":        "adamic_node_memory_usage",
		"now":                "adamic_node_performance_now",
		"timeOrigin":         "adamic_node_time_origin",
		"mark":               "adamic_node_mark",
		"measure":            "adamic_node_measure",
		"clearMarks":         "adamic_node_clear_marks",
		"clearMeasures":      "adamic_node_clear_measures",
		"stackTraceLimit":    "adamic_node_stack_trace_limit",
		"setStackTraceLimit": "adamic_node_set_stack_trace_limit",
	}
	function, known := functions[call.Operation]
	if !known {
		return "", false
	}
	args := []string{}
	for _, argument := range call.Arguments {
		args = append(args, nativeEmitter.value(argument))
	}
	code := function + "(" + strings.Join(args, ", ") + ")"
	if call.Operation == "clearMarks" ||
		call.Operation == "clearMeasures" ||
		call.Operation == "stdoutWrite" ||
		call.Operation == "chdir" {
		nativeEmitter.line("%s;", code)
		if call.Operation == "chdir" {
			nativeEmitter.checkThrown()
		}
		return "NULL", true
	}
	if call.Of.IsReference() {
		value := nativeEmitter.own(call.Of, code)
		if call.Operation == "cwd" || call.Operation == "measure" {
			nativeEmitter.checkThrown()
		}
		return value, true
	}
	return nativeEmitter.snapshot(call.Of, code), true
}

package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	functions := map[string]string{
		"nextTickFeature": "adamic_node_next_tick_feature", "eol": "adamic_node_eol", "handle": "adamic_node_stdout_handle", "cwd": "adamic_node_cwd", "platform": "adamic_node_platform", "pid": "adamic_node_pid",
		"argv": "adamic_node_argv", "execArgv": "adamic_node_exec_argv", "columns": "adamic_node_columns",
		"stdoutWrite": "adamic_node_stdout_write", "memoryUsage": "adamic_node_memory_usage",
		"now": "adamic_node_performance_now", "timeOrigin": "adamic_node_time_origin",
		"mark": "adamic_node_mark", "measure": "adamic_node_measure",
		"clearMarks": "adamic_node_clear_marks", "clearMeasures": "adamic_node_clear_measures",
	}
	function, known := functions[call.Operation]
	if !known {
		return "", false
	}
	args := []string{}
	for _, argument := range call.Arguments {
		args = append(args, e.value(argument))
	}
	code := function + "(" + strings.Join(args, ", ") + ")"
	if call.Operation == "clearMarks" || call.Operation == "clearMeasures" || call.Operation == "stdoutWrite" {
		e.line("%s;", code)
		return "NULL", true
	}
	if call.Of.IsReference() {
		return e.own(call.Of, code), true
	}
	return e.snapshot(call.Of, code), true
}

package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeProcessCall(call ir.ProcessCall) (string, bool) {
	functions := map[string]string{
		"tmpdir": "adamic_node_tmpdir", "performance": "adamic_library_performance", "chdir": "adamic_library_chdir", "envSet": "adamic_node_environment_set", "envDelete": "adamic_node_environment_delete", "nextTickFeature": "adamic_node_next_tick_feature", "eol": "adamic_host_eol", "cwd": "adamic_library_cwd", "platform": "adamic_node_platform", "pid": "adamic_node_pid",
		"argv": "adamic_library_argv", "execArgv": "adamic_library_execArgv", "columns": "adamic_node_columns",
		"stdoutWrite": "adamic_library_stdout_write", "stdout": "adamic_library_stdout", "memoryUsage": "adamic_library_memory_usage",
		"execPath": "adamic_library_exec_path", "dirname": "adamic_library_dirname",
		"now": "adamic_host_performance_now", "timeOrigin": "adamic_host_time_origin", "dateNow": "adamic_host_date_now",
		"mark": "adamic_node_mark", "measure": "adamic_node_measure",
		"clearMarks": "adamic_node_clear_marks", "clearMeasures": "adamic_node_clear_measures",
	}
	function, known := functions[call.Operation]
	if call.Operation == "filename" {
		function, known = functions["execPath"], true
		e.hostHelper("execPath")
	}
	if call.Operation == "handle" {
		return "NULL", true
	}
	if !known {
		return "", false
	}
	switch call.Operation {
	case "cwd", "chdir", "argv", "execArgv", "stdout", "stdoutWrite", "execPath", "dirname", "memoryUsage", "performance":
		e.hostHelper(call.Operation)
	}
	args := []string{}
	for _, argument := range call.Arguments {
		args = append(args, e.value(argument))
	}
	if call.Operation == "envSet" {
		e.line("adamic_node_environment_set(%s, %s);", args[0], args[1])
		return args[1], true
	}
	code := function + "(" + strings.Join(args, ", ") + ")"
	if call.Operation == "clearMarks" || call.Operation == "clearMeasures" || call.Operation == "stdoutWrite" || call.Operation == "chdir" {
		e.line("%s;", code)
		if call.Operation == "chdir" || call.Operation == "stdoutWrite" {
			e.checkThrown()
		}
		return "NULL", true
	}
	if call.Of.IsReference() {
		value := e.own(call.Of, code)
		if call.Operation == "cwd" || call.Operation == "measure" || call.Operation == "memoryUsage" || call.Operation == "execPath" || call.Operation == "filename" || call.Operation == "dirname" || call.Operation == "argv" || call.Operation == "execArgv" {
			e.checkThrown()
		}
		return value, true
	}
	return e.snapshot(call.Of, code), true
}

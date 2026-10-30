package native

import (
	"fmt"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) nodeFSFile(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for i, argument := range call.Arguments {
		args[i] = e.value(argument)
	}
	code := fmt.Sprintf("adamic_fs_file_%s(%s)", call.Operation, strings.Join(args, ", "))
	if call.Operation == "host_join" {
		array := "NULL"
		if len(args) != 0 {
			array = "(adamic_string *const[]){" + strings.Join(args, ", ") + "}"
		}
		code = fmt.Sprintf("adamic_fs_file_host_join(%d, %s)", len(args), array)
	}
	if call.Operation == "host_stdout_columns" {
		const unavailable = `#ifdef ADAMIC_TARGET_WASI
#error "wasi preview1 reports no terminal size"
#endif`
		if !slices.Contains(e.declarations, unavailable) {
			e.declarations = append(e.declarations, unavailable)
		}
		code = "adamic_host_stdout_columns()"
	}
	if call.Operation == "host_stdout_is_tty" {
		code = "adamic_host_stdout_is_tty()"
	}
	if call.Operation == "host_stdout_write" {
		code = "(adamic_write_raw(adamic_stdout, " + args[0] + "), 0.0)"
	}
	if call.Operation == "host_env" {
		code = e.nodeEnvironmentRead(args[0])
	}
	if call.Operation == "host_realpath" {
		code = e.snapshot(ir.String, fmt.Sprintf("adamic_host_realpath(%s)", args[0]))
		e.line("if (%s == NULL) { adamic_fs_file_realpath_error(%s); }", code, args[0])
	}
	result := ""
	if call.Of.IsReference() {
		result = e.own(call.Of, code)
	} else {
		result = e.snapshot(call.Of, code)
	}
	if call.MayThrow() {
		e.checkThrown()
	}
	return result
}

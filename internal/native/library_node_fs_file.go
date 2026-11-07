package native

import (
	"fmt"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) nodeFSFile(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		args[index] = nativeEmitter.value(argument)
	}
	code := fmt.Sprintf("adamic_fs_file_%s(%s)", call.Operation, strings.Join(args, ", "))
	if call.Operation == "host_watch_file" {
		code = "adamic_fs_file_watch(" + strings.Join(args, ", ") + ")"
	}
	if call.Operation == "host_unwatch_file" {
		callback := "NULL"
		if len(args) == 2 {
			callback = args[1]
		}
		code = "adamic_fs_file_unwatch(" + args[0] + ", " + callback + ")"
	}
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
		if !slices.Contains(nativeEmitter.declarations, unavailable) {
			nativeEmitter.declarations = append(nativeEmitter.declarations, unavailable)
		}
		code = "adamic_host_stdout_columns()"
	}
	if call.Operation == "host_stdout_is_tty" {
		code = "adamic_host_stdout_is_tty()"
	}
	if call.Operation == "host_stdout_write" {
		code = "(adamic_write_raw(adamic_stdout, " + args[0] + "), 0.0)"
	}
	if call.Operation == "host_env_set" {
		nativeEmitter.line("adamic_host_environment_set(%s, %s);", args[0], args[1])
		code = "adamic_retain(" + args[1] + ")"
	}
	if call.Operation == "host_env_delete" {
		code = "adamic_host_environment_delete(" + args[0] + ")"
	}
	if call.Operation == "host_env" {
		code = nativeEmitter.nodeEnvironmentRead(args[0])
	}
	if call.Operation == "host_realpath" {
		code = nativeEmitter.snapshot(ir.String, fmt.Sprintf("adamic_host_realpath(%s)", args[0]))
		nativeEmitter.line("if (%s == NULL) { adamic_fs_file_realpath_error(%s); }", code, args[0])
	}
	result := ""
	if call.Of.IsReference() {
		result = nativeEmitter.own(call.Of, code)
	} else {
		result = nativeEmitter.snapshot(call.Of, code)
	}
	if call.MayThrow() {
		nativeEmitter.checkThrown()
	}
	return result
}

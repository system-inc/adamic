package native

import (
	"fmt"
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

package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) nodeFSFile(call ir.NodeFSFile) string {
	e.nodeFSFileWASIMarker(call)
	args := make([]string, len(call.Arguments))
	for i, argument := range call.Arguments {
		args[i] = e.value(argument)
	}
	code := fmt.Sprintf("adamic_fs_file_%s(%s)", call.Operation, strings.Join(args, ", "))
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

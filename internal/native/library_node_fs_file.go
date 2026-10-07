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
	result := ""
	if call.Of.IsReference() {
		result = e.own(call.Of, code)
	} else {
		result = e.snapshot(call.Of, code)
	}
	if call.MayThrow() {
		classes := map[string]int{}
		for index, class := range e.program.Classes {
			if class.Definition >= 1<<30 && class.Definition < (1<<30)+7 {
				classes[class.Name] = index + 1
			}
		}
		if classes["Error"] == 0 || classes["TypeError"] == 0 || classes["RangeError"] == 0 {
			panic("filesystem failures need nominal error classes")
		}
		e.line("adamic_fs_file_error_classes(&adamic_class_%d, &adamic_class_%d, &adamic_class_%d);", classes["Error"], classes["TypeError"], classes["RangeError"])
		e.checkThrown()
	}
	return result
}

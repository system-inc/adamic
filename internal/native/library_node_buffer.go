package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	args := []string{}
	for _, arg := range call.Arguments {
		args = append(args, e.value(arg))
	}
	switch call.Function {
	case "buffer_from", "buffer_string", "hash_update":
		args = append(args, fmt.Sprint(call.Encoding))
	}
	code := "adamic_node_" + call.Function + "(" + strings.Join(args, ", ") + ")"
	if !call.Returns.IsReference() {
		return e.snapshot(call.Returns, code)
	}
	return e.own(call.Returns, code)
}

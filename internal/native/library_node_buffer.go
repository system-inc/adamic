package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	args := []string{}
	for _, arg := range call.Arguments {
		args = append(args, nativeEmitter.value(arg))
	}
	switch call.Function {
	case "buffer_from", "buffer_string", "hash_update":
		args = append(args, fmt.Sprint(call.Encoding))
	}
	code := "adamic_node_" + call.Function + "(" + strings.Join(args, ", ") + ")"
	if !call.Returns.IsReference() {
		return nativeEmitter.snapshot(call.Returns, code)
	}
	return nativeEmitter.own(call.Returns, code)
}

package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	commandArguments := []string{}
	for _, argument := range call.Arguments {
		commandArguments = append(commandArguments, nativeEmitter.value(argument))
	}
	switch call.Function {
	case "buffer_from", "buffer_string", "hash_update":
		commandArguments = append(commandArguments, fmt.Sprint(call.Encoding))
	}
	code := "adamic_node_" + call.Function + "(" + strings.Join(commandArguments, ", ") + ")"
	if !call.Returns.IsReference() {
		return nativeEmitter.snapshot(call.Returns, code)
	}
	return nativeEmitter.own(call.Returns, code)
}

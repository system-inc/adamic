package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (javascriptEmitter *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	encoding := []string{"utf8", "utf16le", "base64", "hex"}[call.Encoding]
	commandArguments := make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		commandArguments[index] = javascriptEmitter.value(argument)
	}
	switch call.Function {
	case "buffer_from":
		return "Buffer.from(" + commandArguments[0] + ", " + quote(encoding) + ")"
	case "buffer_set":
		return "(" + commandArguments[0] + "[" + commandArguments[1] + "] = " + commandArguments[2] + ")"
	case "buffer_copy":
		return "Buffer.from(" + commandArguments[0] + ")"
	case "buffer_string":
		return fmt.Sprintf(
			"(%s).toString(%s, %s, %s)",
			commandArguments[0],
			quote(encoding),
			commandArguments[1],
			commandArguments[2],
		)
	case "hash_new":
		return "adamicNodeCreateHash('sha256')"
	case "hash_update":
		return "(" + commandArguments[0] + ").update(" + commandArguments[1] + ", " + quote(encoding) + ")"
	case "hash_digest":
		return "(" + commandArguments[0] + ").digest('hex')"
	}
	panic("javascript: unknown Buffer/crypto host operation")
}

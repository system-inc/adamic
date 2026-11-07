package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (javascriptEmitter *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	encoding := []string{"utf8", "utf16le", "base64", "hex"}[call.Encoding]
	args := make([]string, len(call.Arguments))
	for index, arg := range call.Arguments {
		args[index] = javascriptEmitter.value(arg)
	}
	switch call.Function {
	case "buffer_from":
		return "Buffer.from(" + args[0] + ", " + quote(encoding) + ")"
	case "buffer_set":
		return "(" + args[0] + "[" + args[1] + "] = " + args[2] + ")"
	case "buffer_copy":
		return "Buffer.from(" + args[0] + ")"
	case "buffer_string":
		return fmt.Sprintf("(%s).toString(%s, %s, %s)", args[0], quote(encoding), args[1], args[2])
	case "hash_new":
		return "adamicNodeCreateHash('sha256')"
	case "hash_update":
		return "(" + args[0] + ").update(" + args[1] + ", " + quote(encoding) + ")"
	case "hash_digest":
		return "(" + args[0] + ").digest('hex')"
	}
	panic("javascript: unknown Buffer/crypto host operation")
}

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
	// Buffer.from(number[]) converts each number with ToUint8, as a Uint8Array made from them does;
	// Buffer.from(aBuffer) copies bytes. One name serves both, so the JavaScript backend's
	// Buffer.from stays one call.
	if call.Function == "buffer_copy" && call.Arguments[0].Type() == ir.Array {
		code = "adamic_typed_array_from_numbers(adamic_typed_array_uint8, " + args[0] + ")"
	}
	if !call.Returns.IsReference() {
		return e.snapshot(call.Returns, code)
	}
	result := e.own(call.Returns, code)
	// Buffer.allocUnsafe's size out of range, or one it can't allocate, leaves Node's RangeError
	// pending and makes no Buffer: the NULL it owns is let go of on the way to the catch.
	if call.MayThrow() {
		e.checkThrown()
	}
	return result
}

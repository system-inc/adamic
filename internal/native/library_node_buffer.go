package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeBufferCall(call ir.NodeBufferCall) string {
	if call.Function == "buffer_index_of" {
		return e.nodeBufferIndexOf(call)
	}
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

// nodeBufferIndexOf is Buffer.indexOf: a string value searched as its UTF-8 bytes, a number as one
// byte, each its own runtime function (node_buffer.c). A number value is read before the offset, as
// TypedArrayFill reads its own, and the offset is held with its present flag, so an absent or
// undefined one is Node's default, as typedArrayOffset passes a typed array's.
func (e *emitter) nodeBufferIndexOf(call ir.NodeBufferCall) string {
	buffer := e.value(call.Arguments[0])
	function, value := "adamic_node_buffer_index_of_string", e.value(call.Arguments[1])
	if call.Arguments[1].Type() == ir.Number {
		function, value = "adamic_node_buffer_index_of_number", e.snapshot(ir.Number, value)
	}
	offset, present := e.typedArrayOffset(call.Arguments, 2)
	return e.snapshot(ir.Number, fmt.Sprintf("%s(%s, %s, %s, %s)", function, buffer, value, offset, present))
}

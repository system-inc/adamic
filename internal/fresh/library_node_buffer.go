package fresh

import "github.com/system-inc/adamic/internal/ir"

// Lowering admits numeric bytes and immutable strings only. These operations
// cannot capture a user reference or call user code. Hash.update returns its
// receiver, so preserve that alias instead of treating it as a fresh hash.
func (a *analysis) nodeBufferCall(call ir.NodeBufferCall) value {
	switch call.Function {
	case "buffer_from", "buffer_copy", "buffer_string", "buffer_set", "hash_new", "hash_update", "hash_digest":
		var receiver value
		for index, argument := range call.Arguments {
			held := a.value(argument)
			if index == 0 {
				receiver = held
			}
		}
		switch call.Function {
		case "buffer_from", "buffer_copy", "hash_new":
			return a.fresh(anyField, value{})
		case "hash_update":
			return receiver
		}
		return value{}
	default:
		a.unknown(call)
		return a.call(a.operands(call), call.Type())
	}
}

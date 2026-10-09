package flow

import (
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// A Date setter changes scalar internal data, without capturing any reference argument.
func (n *inference) dateCall(call ir.DateCall) shape {
	receiver := shape{}
	if call.Receiver != nil {
		receiver = n.value(call.Receiver)
	}
	for _, argument := range call.Arguments {
		n.value(argument)
	}
	if strings.HasPrefix(call.Method, "set") {
		n.store(receiver, shape{})
	}
	if call.Returns == ir.Object {
		return shape{fresh: true}
	}
	return shape{}
}

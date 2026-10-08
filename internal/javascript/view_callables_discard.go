package javascript

import "github.com/system-inc/adamic/internal/ir"

// The ordinary calling convention discards the result after independently
// checking the stored producer against the erased marker's zero-argument call.
func (e *emitter) emitViewCallableDiscard(call ir.CallClosure) string {
	property := ir.Property{View: call.DiscardView, ViewContract: call.DiscardContract}
	return "adamicCall(" + e.emitViewCallableCertificate(property, e.value(call.Closure)) + ", [])"
}

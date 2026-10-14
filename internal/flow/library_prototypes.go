package flow

import "github.com/system-inc/adamic/internal/ir"

// Prototype mutation can reject a cycle or an integrity restriction.
func objectPrototypeMayThrow(node any) bool {
	call, ok := node.(ir.ObjectCall)
	return ok && call.Method == "setPrototypeOf"
}

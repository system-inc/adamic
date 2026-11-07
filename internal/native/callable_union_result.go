package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// callableUnionResult decodes the producer's ABI before boxing the invocation's
// result. A synthesized union signature does not change its producers' slots.
// References already come back owned; scalar boxes acquire their own ownership.
func (nativeEmitter *emitter) callableUnionResult(result, closure, method string) string {
	boxed := nativeEmitter.temporary()
	nativeEmitter.line("adamic_heap *%s;", boxed)
	branches := 0
	for index, function := range nativeEmitter.program.Functions {
		condition := ""
		if function.Closure && closure != "" {
			condition = fmt.Sprintf(
				"(%s != NULL && %s)",
				closure,
				nativeEmitter.unionClosureCodeIdentity(closure, index),
			)
		} else if !function.Closure && function.Receiver && method != "" && nativeEmitter.dispatchable(index) {
			thunk := nativeEmitter.methodThunk(index)
			if closure != "" {
				if nativeEmitter.program.ClosureConventionNeeded() {
					field := "code"
					if nativeEmitter.program.PackedCountNeeded(index) {
						field = "counted_code"
					}
					condition = fmt.Sprintf(
						"(%s == NULL && %s.counted == %t && %s.%s == %s)",
						closure,
						method,
						nativeEmitter.program.PackedCountNeeded(index),
						method,
						field,
						thunk,
					)
				} else {
					condition = fmt.Sprintf("(%s == NULL && %s == %s)", closure, method, thunk)
				}
			} else {
				condition = fmt.Sprintf("%s == %s", method, thunk)
			}
		}
		if condition == "" {
			continue
		}
		value := "NULL"
		if function.Returns != 0 {
			value = unslotted(function.Returns, result+"."+member(function.Returns))
			if function.Returns.IsReference() {
				value = fmt.Sprintf("(%s)%s", cType(function.Returns), value)
			}
			value, _ = converted(function.Returns, ir.Union, value)
		}
		prefix := "if"
		if branches != 0 {
			prefix = "else if"
		}
		nativeEmitter.line("%s (%s) { %s = %s; }", prefix, condition, boxed, value)
		branches++
	}
	// Every runtime code pointer was emitted from a function in this program.
	// Missing producer evidence must never reinterpret a scalar as a reference.
	if branches == 0 {
		nativeEmitter.line("adamic_unreachable();")
	} else {
		nativeEmitter.line("else { adamic_unreachable(); }")
	}
	return nativeEmitter.own(ir.Union, boxed)
}

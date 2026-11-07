package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A callable union's synthesized signature describes legal arguments, not the
// selected implementation's slots. Adapt each argument after checking the callee.
func (nativeEmitter *emitter) checkedCallableArguments(call ir.CallClosure, closure string) (string, string) {
	nativeEmitter.declarations = append(nativeEmitter.declarations, "#include \"view_callables_contract.h\"")
	if len(call.Spread) != 0 {
		panic("compiler bug: unavailable checked callable spread adapter")
	}
	slots := []string{}
	for index, argument := range call.Arguments {
		value := nativeEmitter.value(argument)
		ordinary := fmt.Sprintf("(adamic_value){.%s = %s}", member(argument.Type()), slotted(argument.Type(), value))
		tests := []string{}
		for _, producer := range nativeEmitter.unionCallableProducers() {
			function := nativeEmitter.program.Functions[producer.Function]
			if index < len(function.Parameters) &&
				nativeEmitter.program.Locals[function.Parameters[index]].Type == ir.Union {
				tests = append(tests, nativeEmitter.unionClosureCodeIdentity(closure, producer.Function))
			}
		}
		if len(tests) != 0 && argument.Type() != ir.Union {
			boxed, fresh := converted(argument.Type(), ir.Union, value)
			if fresh {
				boxed = nativeEmitter.own(ir.Union, boxed)
			}
			ordinary = "(" +
				strings.Join(tests, " || ") +
				" ? (adamic_value){.reference = " +
				boxed +
				"} : " +
				ordinary +
				")"
		}
		slots = append(slots, ordinary)
	}
	return nativeEmitter.closureSlots(call, slots, "", fmt.Sprint(len(slots))), fmt.Sprint(len(slots))
}

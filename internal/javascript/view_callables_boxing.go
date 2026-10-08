package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// JavaScript needs no boxing, but the shared unknown-producer boundary must stop
// in both backends. Evaluate callee and arguments before entering this adapter.
func (e *emitter) viewCallableBoxedDispatch(call ir.CallClosure) string {
	if len(call.GenericInstances) != 0 {
		cases := []string{}
		for _, pair := range call.GenericInstances {
			if pair[0] >= 0 {
				cases = append(cases, "if (callee instanceof AdamicClosure && callee.code === "+functionName(e.program, pair[0])+") return "+functionName(e.program, pair[1])+"(callee, arguments_);")
			}
		}
		return "((callee, arguments_) => {" + strings.Join(cases, " ") + "panic(\"generic callable: no compatible producer instantiation\");})"
	}
	needed := call.Returns == ir.Union
	for _, argument := range call.Arguments {
		needed = needed || argument.Type() == ir.Union
	}
	if !needed {
		return "adamicCall"
	}
	choices := []string{}
	for index, function := range e.program.Functions {
		if function.Closure {
			choices = append(choices, "code === "+functionName(e.program, index))
		}
	}
	known := "false"
	if len(choices) != 0 {
		known = strings.Join(choices, " || ")
	}
	return "((callee, arguments_) => { const code = callee instanceof AdamicClosure ? callee.code : callee; if (" + known + ") return adamicCall(callee, arguments_); panic(\"callable ABI adapter: unknown producer signature\"); })"
}

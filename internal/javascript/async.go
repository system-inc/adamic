package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func asyncJavaScript(program *ir.Program) string {
	var out strings.Builder
	out.WriteString("const adamicTypeOf = (value) => typeof value;\n")
	out.WriteString("const adamicUnready = (name) => { throw new ReferenceError(`Cannot access '${name}' before initialization`); };\n")
	e := &emitter{program: program, indent: 1}
	for index, fn := range program.Async.Functions {
		parameters := []string{}
		for _, local := range fn.Parameters {
			parameters = append(parameters, e.name(local))
		}
		fmt.Fprintf(&out, "async function adamic_async_%d(%s) {\n", index, strings.Join(parameters, ", "))
		for _, local := range fn.Locals {
			ready := false
			for _, parameter := range fn.Parameters {
				ready = ready || parameter == local
			}
			fmt.Fprintf(&out, "let %s = %t;\n", readyName(local), ready)
		}
		for _, state := range fn.States {
			for _, statement := range state.Body {
				switch statement := statement.(type) {
				case ir.Declare:
					// This closed primitive graph has no synchronous declaration prologue.
					// Preserve checked reads across suspension with function-local readiness.
					e.line("let %s = %s;", e.name(statement.Local), e.value(statement.Value))
					e.line("%s = %t;", readyName(statement.Local), !statement.Uninitialized)
				case ir.WriteLine:
					e.statements([]ir.Statement{statement})
				default:
					panic("javascript async: unknown lowered statement")
				}
			}
			out.WriteString(e.out.String())
			e.out.Reset()
			if state.Await != nil {
				wait := state.Await
				value := "Promise.resolve()"
				if wait.Function >= 0 {
					args := []string{}
					for _, arg := range wait.Arguments {
						args = append(args, e.value(arg))
					}
					value = fmt.Sprintf("adamic_async_%d(%s)", wait.Function, strings.Join(args, ", "))
				} else if wait.Value != nil {
					value = "Promise.resolve(" + e.value(wait.Value) + ")"
				}
				if state.Target >= 0 {
					fmt.Fprintf(&out, "let %s = await %s;\n%s = true;\n", e.name(state.Target), value, readyName(state.Target))
				} else {
					fmt.Fprintf(&out, "await %s;\n", value)
				}
			} else if state.Throw != nil {
				fmt.Fprintf(&out, "throw new Error(%s);\n", e.value(state.Throw))
			} else if state.Result != nil {
				fmt.Fprintf(&out, "return %s;\n", e.value(state.Result))
			} else {
				out.WriteString("return;\n")
			}
		}
		out.WriteString("}\n")
	}
	fmt.Fprintf(&out, "await adamic_async_%d();\n", program.Async.Entry)
	return out.String()
}

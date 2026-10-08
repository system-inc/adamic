package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// A void view permits a value-returning target. Its packed reference result is
// still owned. Identify the selected code, rather than interpreting scalar bits
// as a pointer or forgetting a reference returned through the discarded view.
func (e *emitter) discardClosureResult(call ir.CallClosure, closure, receiver, method, invocation string) {
	if call.Direct > 0 && e.program.Functions[call.Direct-1].Returns.IsReference() {
		e.line("adamic_release((%s).reference);", invocation)
		return
	}
	conditions := []string{}
	for index, function := range e.program.Functions {
		if !function.Closure || !function.Returns.IsReference() || closure == "" {
			continue
		}
		member := "code"
		condition := fmt.Sprintf("%s->code == %s", closure, e.functionName(index))
		if e.program.ClosureConventionNeeded() {
			counted := e.program.PackedCountNeeded(index)
			if counted {
				member = "counted_code"
			}
			condition = fmt.Sprintf("%s->counted == %t && %s->%s == %s", closure, counted, closure, member, e.functionName(index))
		}
		conditions = append(conditions, "("+condition+")")
	}
	if receiver != "" {
		property := call.Closure.(ir.Property)
		seen := map[int]bool{}
		walkExpressions(e.program, func(value ir.Expression) {
			literal, known := value.(ir.ObjectLiteral)
			if !known {
				return
			}
			for _, candidate := range literal.Methods {
				index := candidate.Function
				if candidate.Name != property.Name || seen[index] || !e.program.Functions[index].Returns.IsReference() {
					continue
				}
				seen[index] = true
				code := e.methodThunk(index)
				if method == code {
					conditions = append(conditions, "true")
					continue
				}
				condition := method + " == " + code
				if e.program.ClosureConventionNeeded() {
					member := "code"
					counted := e.program.PackedCountNeeded(index)
					if counted {
						member = "counted_code"
					}
					condition = fmt.Sprintf("%s.counted == %t && %s.%s == %s", method, counted, method, member, code)
				}
				conditions = append(conditions, "("+condition+")")
			}
		})
	}
	if len(conditions) == 0 {
		e.line("%s;", invocation)
		return
	}
	result := e.temporary()
	e.line("adamic_value %s = %s;", result, invocation)
	e.closureThrown()
	condition := strings.Join(conditions, " || ")
	if closure != "" {
		// Prototype dispatch has no closure. Test that pointer before its code.
		parts := []string{}
		for _, part := range conditions {
			if strings.Contains(part, "->") {
				part = "(" + closure + " != NULL && " + part + ")"
			}
			parts = append(parts, part)
		}
		condition = strings.Join(parts, " || ")
	}
	e.line("if (%s) adamic_release(%s.reference);", condition, result)
}

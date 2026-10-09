package javascript

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) optionalMethodCall(call ir.CallClosure) string {
	property := call.Closure.(ir.Property)
	counted := false
	for _, function := range e.program.Functions {
		counted = counted || function.ArgumentsCount != 0
	}
	method := "fn(object, ...values)"
	if counted {
		method = "adamicDirect(fn, [object, ...values])"
	}
	presence := ""
	if call.OptionalPresent != 0 {
		presence = e.name(call.OptionalPresent-1) + " = true; "
	}
	return fmt.Sprintf("((object) => { const fn = object[%s]; if (fn === undefined || fn === null) return undefined; %sconst values = [%s]; return Object.hasOwn(object, %s) ? (fn.receiver ? adamicCall(fn, [object, ...values]) : adamicCall(fn, values)) : %s; })(%s)", quote(property.Name), presence, e.callValues(call.Arguments, call.Spread), quote(property.Name), method, e.value(property.Object))
}

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
	selection := fmt.Sprintf("(Object.hasOwn(object, %s) ? object[%s] : (adamicStaticMethods.get(object) && Object.hasOwn(adamicStaticMethods.get(object), %s) ? adamicStaticMethods.get(object)[%s] : object[%s]))", quote(property.Name), quote(property.Name), quote(property.Name), quote(property.Name), quote(property.Name))
	if property.CallableAccessor {
		selection = fmt.Sprintf("(adamicFindAccessor(object, %s) ? adamicGetAccessor(object, %s) : object[%s])", quote(property.Name), quote(property.Name), quote(property.Name))
	}
	guard := "if (fn === undefined || fn === null) return undefined; "
	if call.RequiredCallable {
		guard = ""
	}
	presence := ""
	if call.OptionalPresent != 0 {
		presence = e.name(call.OptionalPresent-1) + " = true; "
	}
	return fmt.Sprintf("((object) => { const fn = %s; const own = Object.hasOwn(object, %s); %s%sconst values = [%s]; if (!(fn instanceof AdamicClosure) && typeof fn !== 'function') panic('TypeError: optional call value is not callable'); return (own || fn instanceof AdamicClosure) ? (fn.receiver ? adamicCall(fn, [object, ...values]) : adamicCall(fn, values)) : %s; })(%s)", selection, quote(property.Name), guard, presence, e.callValues(call.Arguments, call.Spread), method, e.value(property.Object))
}

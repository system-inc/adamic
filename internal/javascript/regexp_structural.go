package javascript

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// Optional reads hand out one shared prototype function in the backend's closure shape.
// Detachment throws through Node itself, and identity agrees across different RegExp objects.
// Calls off an object keep their own receiver in adamicCallee and adamicOptionalMethodCall.
const regexpFunctionRuntime = "const adamicRegExpTest = new AdamicClosure((closure, values) => Reflect.apply(RegExp.prototype.test, undefined, values), [], false, 1);\nconst adamicRegExpFunction = value => value === RegExp.prototype.test ? adamicRegExpTest : value;\n"

func (e *emitter) objectFieldRead(expression ir.Property) string {
	if expression.Unset && expression.View != "" {
		return fmt.Sprintf("adamicPlaceholderViewField(%s, %s, %s, %d, %s, [%s], %t, %t)", e.value(expression.Object), quote(expression.Name), quote(expression.View), expression.UnsetType, quote(expression.ViewType), e.values(expression.ViewAllowed), expression.Optional, expression.Absent)
	}
	if expression.View != "" {
		return e.checkedViewField(expression)
	}
	if expression.Readiness != "" {
		return fmt.Sprintf("adamicReadField(%s, %s, %s, %t, %t)", e.value(expression.Object), quote(expression.Name), quote(expression.Readiness), expression.Optional, expression.Absent)
	}
	if expression.Optional {
		return e.value(expression.Object) + "?.[" + quote(expression.Name) + "]"
	}
	return e.value(expression.Object) + "[" + quote(expression.Name) + "]"
}

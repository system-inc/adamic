package javascript

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) objectFieldRead(expression ir.Property) string {
	if expression.MethodTruthiness {
		read := "adamicMethodSlot(receiver, " + quote(expression.Name) + ")"
		if expression.Optional {
			read = "receiver != null && " + read
		}
		return "((receiver) => Boolean(" + read + "))(" + e.value(expression.Object) + ")"
	}
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

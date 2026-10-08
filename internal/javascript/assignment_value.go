package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) assignmentValue(expression ir.AssignmentValue) string {
	switch store := expression.Store.(type) {
	case ir.Assign:
		value := e.value(assignmentRaw(expression, store.Value))
		if store.Checked {
			return fmt.Sprintf("((value) => { if (!%s) adamicUnready(%s); return %s = value; })(%s)", readyName(store.Local), quote(e.program.Locals[store.Local].Name), e.variable(store.Local), value)
		}
		return fmt.Sprintf("(%s = %s)", e.variable(store.Local), value)
	case ir.SetProperty:
		return fmt.Sprintf("(%s[%s] = %s)", e.value(store.Object), quote(store.Name), e.value(assignmentRaw(expression, store.Value)))
	case ir.SetIndex:
		return fmt.Sprintf("((array, index, value) => { adamicSetIndex(array, index, value); return value; })(%s, %s, %s)", e.value(store.Array), e.value(store.Index), e.value(assignmentRaw(expression, store.Value)))
	}
	panic("javascript: assignment value without a store")
}

func assignmentRaw(expression ir.AssignmentValue, fitted ir.Expression) ir.Expression {
	if expression.Value != nil {
		return expression.Value
	}
	return fitted
}

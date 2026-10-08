package fresh

import "github.com/system-inc/adamic/internal/ir"

// Typed arrays hold only numbers, so their stores cannot introduce reference cycles.
// Views conservatively share the parent's abstract identity and every operand is evaluated.
func (a *analysis) typedArray(expression ir.Expression) (value, bool) {
	switch v := expression.(type) {
	case ir.TypedArrayNew:
		a.value(v.Source)
		return a.fresh(anyField, value{}), true
	case ir.TypedArraySort:
		return a.value(v.Array), true
	case ir.TypedArrayFill:
		array := a.value(v.Array)
		for _, argument := range v.Arguments {
			a.value(argument)
		}
		return array, true
	case ir.TypedArraySubarray:
		array := a.value(v.Array)
		for _, argument := range v.Arguments {
			a.value(argument)
		}
		return array, true
	case ir.TypedArraySet:
		a.value(v.Array)
		for _, argument := range v.Arguments {
			a.value(argument)
		}
		return value{}, true
	}
	return value{}, false
}

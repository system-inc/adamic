package fresh

import "github.com/system-inc/adamic/internal/ir"

// Typed arrays hold only numbers, so their stores cannot introduce reference cycles.
// Views conservatively share the parent's abstract identity and every operand is evaluated.
func (analysis *analysis) typedArray(expression ir.Expression) (value, bool) {
	switch typed := expression.(type) {
	case ir.TypedArrayNew:
		analysis.value(typed.Source)
		return analysis.fresh(anyField, value{}), true
	case ir.TypedArraySort:
		return analysis.value(typed.Array), true
	case ir.TypedArrayFill:
		array := analysis.value(typed.Array)
		for _, argument := range typed.Arguments {
			analysis.value(argument)
		}
		return array, true
	case ir.TypedArraySubarray:
		array := analysis.value(typed.Array)
		for _, argument := range typed.Arguments {
			analysis.value(argument)
		}
		return array, true
	case ir.TypedArraySet:
		analysis.value(typed.Array)
		for _, argument := range typed.Arguments {
			analysis.value(argument)
		}
		return value{}, true
	}
	return value{}, false
}

package fresh

import "github.com/system-inc/adamic/internal/ir"

// Typed arrays hold only numbers, so their stores cannot introduce reference cycles.
// Views conservatively share the parent's abstract identity and every operand is evaluated.
func (analysis *analysis) typedArray(expression ir.Expression) (value, bool) {
	switch v := expression.(type) {
	case ir.TypedArrayNew:
		analysis.value(v.Source)
		return analysis.fresh(anyField, value{}), true
	case ir.TypedArraySort:
		return analysis.value(v.Array), true
	case ir.TypedArrayFill:
		array := analysis.value(v.Array)
		for _, argument := range v.Arguments {
			analysis.value(argument)
		}
		return array, true
	case ir.TypedArraySubarray:
		array := analysis.value(v.Array)
		for _, argument := range v.Arguments {
			analysis.value(argument)
		}
		return array, true
	case ir.TypedArraySet:
		analysis.value(v.Array)
		for _, argument := range v.Arguments {
			analysis.value(argument)
		}
		return value{}, true
	}
	return value{}, false
}

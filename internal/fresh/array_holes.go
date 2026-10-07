package fresh

import "github.com/system-inc/adamic/internal/ir"

// A length constructor allocates an empty reachability graph: holes own no
// references. Length writes can remove edges but never introduce an edge, so
// retaining existing edges is a conservative approximation after a shrink.
func (a *analysis) arrayHoles(expression ir.Expression) value {
	switch expression := expression.(type) {
	case ir.ArrayHoles:
		a.value(expression.Length)
		return a.fresh(elementKey, value{})
	case ir.ArraySetLength:
		a.value(expression.Array)
		a.value(expression.Length)
	case ir.ArrayRangeErrorIs:
		a.value(expression.Value)
	}
	return value{}
}

package fresh

import "github.com/system-inc/adamic/internal/ir"

// These nodes do not themselves write an existing slot. Constructor and identity
// operations have precise models. Reflected keys and the current closure keep
// the opaque-call model: operands escape, results are outside and parameters clobber.
func (a *analysis) libraryLanguage(expression ir.Expression) (value, bool) {
	if value, known := a.typedArray(expression); known {
		return value, true
	}
	switch expression := expression.(type) {
	case ir.ArrayHoles:
		// A length-form constructor evaluates its scalar length, including any
		// nested writes, and returns fresh storage with no present references.
		a.value(expression.Length)
		return a.fresh(elementKey, value{}), true
	case ir.ArrayRangeErrorIs:
		// Identity inspection borrows the error; it neither stores nor escapes it.
		a.value(expression.Value)
		return value{}, true
	case ir.NodeBufferCall:
		return a.nodeBufferCall(expression), true
	case ir.ObjectKeys, ir.ClosureSelf, ir.LibraryGlobal:
		return a.call(a.operands(expression), expression.Type()), true
	}
	return value{}, false
}

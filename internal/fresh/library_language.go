package fresh

import "github.com/system-inc/adamic/internal/ir"

// These language nodes do not write into an existing slot. Keep the previous opaque-call model
// for their reachability until the cycle finder has a more precise model of reflected keys and
// the current closure: operands escape, results are outside, and parameters are clobbered.
func (a *analysis) libraryLanguage(expression ir.Expression) (value, bool) {
	switch expression := expression.(type) {
	case ir.ArrayHoles, ir.ArraySetLength, ir.ArrayRangeErrorIs:
		return a.arrayHoles(expression), true
	case ir.NodeBufferCall:
		return a.nodeBufferCall(expression), true
	case ir.ProcessCall:
		// Process state holds only a number. Environment results are copied immutable strings;
		// evaluate operand effects, but no operation captures or writes a mutable heap slot.
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		return value{}, true
	case ir.ObjectKeys, ir.ClosureSelf, ir.LibraryGlobal:
		return a.call(a.operands(expression), expression.Type()), true
	}
	return value{}, false
}

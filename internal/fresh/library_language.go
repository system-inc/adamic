package fresh

import "github.com/system-inc/adamic/internal/ir"

// These language nodes do not write into an existing slot. Keep the previous opaque-call model
// for their reachability until the cycle finder has a more precise model of reflected keys and
// the current closure: operands escape, results are outside, and parameters are clobbered.
func (a *analysis) libraryLanguage(expression ir.Expression) (value, bool) {
	switch expression.(type) {
	case ir.ObjectKeys, ir.ClosureSelf, ir.LibraryGlobal:
		return a.call(a.operands(expression), expression.Type()), true
	}
	return value{}, false
}

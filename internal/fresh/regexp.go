package fresh

import "github.com/system-inc/adamic/internal/ir"

// regexCall knows only the native regex operations, not arbitrary calls on an
// object named RegExp. Lowering excludes replacement callbacks and overriding
// regex properties. These methods can update lastIndex or an iterator's done
// flag, but cannot store a reference to a user object. Their operands still run.
func (a *analysis) regexCall(expression ir.RegExpCall) value {
	switch expression.Method {
	case "test", "exec", "match", "matchAll", "next", "iteratorDone",
		"replace", "replaceAll", "split", "search":
		a.value(expression.Value)
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		if mutable(expression.Type()) {
			// A fresh result container can acquire user references later. Its
			// nested metadata is conservatively outside, so reads cannot lose
			// aliases to groups, indices, iterator results or their children.
			return a.fresh(anyField, outsideValue())
		}
		return value{}
	default:
		// A future method might call user code or store an argument. It must
		// remain unknown until its effects have been proved too.
		a.unknown(expression)
		return a.call(a.operands(expression), expression.Type())
	}
}

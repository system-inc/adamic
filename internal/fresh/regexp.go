package fresh

import "github.com/system-inc/adamic/internal/ir"

// regexCall knows only the native regex operations, not arbitrary calls on an
// object named RegExp. Lowering excludes overriding regex properties. Intrinsic
// methods cannot store a reference to a user object; their operands still run.
// The admitted replacement callbacks use the ordinary conservative call effects.
func (a *analysis) regexCall(expression ir.RegExpCall) value {
	switch expression.Method {
	case "replaceCallback", "replaceAllCallback", "symbol:replaceCallback":
		return a.call(a.operands(expression), expression.Type())
	case "test", "exec", "match", "matchAll", "next", "iteratorDone",
		"replace", "replaceAll", "split", "search", "toString",
		"symbol:match", "symbol:matchAll", "symbol:replace", "symbol:split", "symbol:search":
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

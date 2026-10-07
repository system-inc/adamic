package fresh

import "github.com/system-inc/adamic/internal/ir"

// regexCall knows only the native regex operations, not arbitrary calls on an
// object named RegExp. Lowering excludes overriding regex properties. Replacement callbacks
// retain ordinary call escape and clobber effects. These methods can update lastIndex or an iterator's done
// flag, but cannot store a reference to a user object. Their operands still run.
func (a *analysis) regexCall(expression ir.RegExpCall) value {
	switch expression.Method {
	case "test", "exec", "match", "matchAll", "next", "iteratorDone",
		"replace", "replaceAll", "split", "search", "toString",
		"symbol:match", "symbol:matchAll", "symbol:replace", "symbol:search", "symbol:split":
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
	case "replaceCallback", "replaceAllCallback", "symbol:replaceCallback":
		return a.call(a.operands(expression), expression.Type())
	default:
		// A future method might call user code or store an argument. It must
		// remain unknown until its effects have been proved too.
		a.unknown(expression)
		return a.call(a.operands(expression), expression.Type())
	}
}

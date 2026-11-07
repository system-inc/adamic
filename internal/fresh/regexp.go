package fresh

import "github.com/system-inc/adamic/internal/ir"

// regexCall knows only the native regex operations, not arbitrary calls on an
// object named RegExp. Typed replacement callbacks are conservative calls;
// lowering excludes overriding regex properties. These methods can update lastIndex or an iterator's done
// flag, but cannot store a reference to a user object. Their operands still run.
func (analysis *analysis) regexCall(expression ir.RegExpCall) value {
	if expression.Replacement != nil {
		return analysis.call(analysis.operands(expression), expression.Type())
	}
	switch expression.Method {
	case "test", "exec", "match", "matchAll", "next", "iteratorDone",
		"replace", "replaceAll", "split", "search":
		analysis.value(expression.Value)
		for _, argument := range expression.Arguments {
			analysis.value(argument)
		}
		if mutable(expression.Type()) {
			// A fresh result container can acquire user references later. Its
			// nested metadata is conservatively outside, so reads cannot lose
			// aliases to groups, indices, iterator results or their children.
			return analysis.fresh(anyField, outsideValue())
		}
		return value{}
	default:
		// A future method might call user code or store an argument. It must
		// remain unknown until its effects have been proved too.
		analysis.unknown(expression)
		return analysis.call(analysis.operands(expression), expression.Type())
	}
}

package fresh

import "github.com/system-inc/adamic/internal/ir"

// regexCall knows only the native regex operations, not arbitrary calls on an
// object named RegExp. Lowering excludes overriding regex properties. Builtin
// operations may update scalar metadata but cannot store user references.
// Functional replacement has ordinary call escape and clobber effects. Every
// receiver and argument still runs in source order.
func (a *analysis) regexCall(expression ir.RegExpCall) value {
	switch expression.Method {
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
	case "compile":
		// Only immutable bytecode and scalar metadata are installed. Return
		// the receiver alias, never a falsely fresh object.
		receiver := a.value(expression.Value)
		for _, argument := range expression.Arguments {
			a.value(argument)
		}
		return receiver
	case "replaceCallback", "replaceAllCallback", "symbol:replaceCallback":
		// User callbacks may retain operands and mutate reachable objects. This
		// is the same conservative effect as a known ordinary closure call.
		return a.call(a.operands(expression), expression.Type())
	default:
		// A future method might call user code or store an argument. It must
		// remain unknown until its effects have been proved too.
		a.unknown(expression)
		return a.call(a.operands(expression), expression.Type())
	}
}

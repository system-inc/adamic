package lower

import "github.com/system-inc/adamic/internal/ir"

// IteratorYieldResult permits an omitted done flag. Its boolean test is false on
// both omission and explicit false; preserve the optional value everywhere else.
func libraryIteratorDoneTruth(value ir.Expression) ir.Expression {
	if call, ok := value.(ir.RegExpCall); ok && call.Method == "iteratorDone" && call.Returns == ir.MaybeBoolean {
		return ir.Coalesce{Value: value, Fallback: ir.BooleanConstant{Value: false}, Of: ir.Boolean}
	}
	return value
}

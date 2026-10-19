package lower

import (
	"github.com/system-inc/adamic/internal/ir"
)

// voidDiscards says not yet to a void call that has no one way to let go of its result. A void
// call's result is released, never read (ir.Discard): a function value's by its targets' Returns,
// or by the closure's own flag where they're Unknown or disagree. A call through an interface whose
// object has no field of the name runs a class's method instead, which carries no flag, bounded by
// every method table holding the name, so those must agree. The IR keeps no source position, so the
// site is named by the function it's in, as the other whole-program passes name theirs.
func (l *lowering) voidDiscards() error {
	var failure error
	in := "at the module's top level"
	check := func(node any) bool {
		if failure != nil {
			return false
		}
		if call, isCall := node.(ir.CallClosure); isCall && call.Returns == 0 && l.result.MethodDiscard(call) == ir.DiscardMixed {
			failure = &NotYet{Where: l.result.Source, What: "a void call through an interface " + in + " where the class methods of its name disagree on whether their result is a reference to release; give the methods one result type, or call through a field"}
			return false
		}
		return true
	}
	walk(l.result.Main, check)
	for _, function := range l.result.Functions {
		in = "in " + function.Name
		walk(function.Body, check)
	}
	return failure
}

package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// widenedReadsOf counts a lowered program's source reads of a field by name, split by whether each is a widened
// read checked where it's made (#zpzm9ta): marked in WidenedReads under its own key, keeping the view readiness
// keeps for it. An honest value passes a checked read and a plain one alike, so this is what shows the check
// behavior can't. A read marked but stripped of its view would read the field raw, the mutant where readiness
// strips the view anyway, so it fails t.
func widenedReadsOf(t *testing.T, program *ir.Program, name string) (checked, other int) {
	t.Helper()
	inspect := func(node any) bool {
		read, isProperty := node.(ir.Property)
		if !isProperty || read.Name != name || read.ViewWhere == "" {
			return true
		}
		switch {
		case !program.WidenedReads[ir.WidenedReadKey(read)]:
			other++
		case read.View == "":
			t.Errorf("%s: the read of %s is marked widened but has no view, so it reads the field raw", read.ViewWhere, name)
		default:
			checked++
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return checked, other
}

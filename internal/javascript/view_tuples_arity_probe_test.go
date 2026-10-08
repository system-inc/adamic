package javascript

import (
	"fmt"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestTupleArityIRPlanProbe(t *testing.T) {
	for _, rest := range []bool{false, true} {
		c := ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object, FixedTuple: true, TupleVariable: true, TupleMinimum: 1, Name: "Tuple", Tuple: []ir.ViewContractID{1, 1}}
		if rest {
			c.TupleRest = 1
		}
		e := emitter{program: &ir.Program{ViewContracts: []ir.ViewContract{c}}}
		for _, sample := range []struct {
			value string
			valid bool
			found string
		}{
			{"[7]", true, ""}, {"[7,undefined]", true, ""}, {"[]", false, "array"}, {"[7,8,9]", rest, "array"}, {"({'0':7})", false, "object"},
		} {
			checked, ok := e.viewTuple(ir.Property{ViewContract: 1, View: "selected"}, sample.value)
			if !ok {
				t.Fatal("tuple plan absent")
			}
			if sample.valid {
				runViewNode(t, viewTestRuntime+checked+";console.log('ok');", "ok\n", "", 0)
			} else {
				runViewNode(t, viewTestRuntime+checked+";", "", fmt.Sprintf("adamic: panic: cast failed: field read failed: selected is not a Tuple; expected Tuple, found %s\n", sample.found), 70)
			}
		}
	}
}

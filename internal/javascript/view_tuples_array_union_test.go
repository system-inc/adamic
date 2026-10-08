package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"strings"
	"testing"
)

func TestTupleHeapUnionSelectionProbe(t *testing.T) {
	p := &ir.Program{ViewContracts: []ir.ViewContract{
		{Kind: ir.ViewScalar, Of: ir.Number, Name: "ID"},
		{Kind: ir.ViewObject, Of: ir.Object, Name: "Tuple", FixedTuple: true, Tuple: []ir.ViewContractID{1, 1}},
		{Kind: ir.ViewUnion, Of: ir.Union, Name: "ID | Tuple", Members: []ir.ViewContractID{1, 2}},
	}}
	e := emitter{program: p}
	for _, sample := range []struct{ value, found string }{{"7", ""}, {"[7,9]", ""}, {"({'0':7,'1':9})", "object"}, {"false", "boolean"}, {"[7]", "object"}} {
		checked := e.tupleArrayUnionCheck(3, "item", sample.value)
		if os.Getenv("ADAMIC_TUPLE_SELECTOR_MUTANT") == "shape" {
			checked = strings.ReplaceAll(checked, "return Array.isArray(snapshot.value) && snapshot.value.length === 2;", "return true;")
		}
		if sample.found == "" {
			runViewNode(t, viewTestRuntime+"const v="+checked+";console.log(Array.isArray(v)?v[0]:v);", "7\n", "", 0)
			continue
		}
		runViewNode(t, viewTestRuntime+checked+";", "", "adamic: panic: cast failed: field read failed: item matches no member of ID | Tuple; expected ID | Tuple, found "+sample.found+"\n", 70)
	}
}

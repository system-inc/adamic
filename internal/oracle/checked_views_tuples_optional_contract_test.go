package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"testing"
)

func TestCheckedViewTupleOptionalContract(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"optional-view-absent", "7\nundefined\n", "", ""},
		{"optional-view-undefined", "7\nundefined\n", "", ""},
		{"optional-view-present", "7\nstring\n", "", ""},
		{"optional-map-absent", "7\nundefined\n", "", ""},
		{"optional-map-undefined", "7\nundefined\n", "", ""},
		{"optional-map-present", "7\nstring\n", "", ""},
	})
}
func TestCheckedViewTupleOptionalAbsentMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_OPTIONAL_CONTRACT_MUTANT") != "present" {
		t.Skip("opt-in absent-as-present mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "optional-view-absent")
	index := len(p.Strings)
	p.Strings = append(p.Strings, "invented")
	changed := changeOptionalProducerLiteral(p, func(v ir.ObjectLiteral) bool { return v.Tuple && len(v.Fields) == 1 }, func(v ir.ObjectLiteral) ir.ObjectLiteral {
		v.Fields = append(v.Fields, ir.Field{Name: "1", Value: ir.StringConstant{Index: index}})
		return v
	})
	if changed != 1 {
		t.Fatalf("changed %d producers", changed)
	}
	actual, _ := nativelyUncached(t, p)
	if diff := disagreement(run{stdout: []byte("7\nundefined\n")}, actual); diff != "" {
		t.Fatalf("absent-as-present mutant caught: %s; %#v", diff, actual)
	}
	t.Fatal("mutant escaped")
}

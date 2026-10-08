package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"testing"
)

func TestCheckedViewTupleRestContract(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"rest-view-zero", "7\nundefined\nundefined\n", "", ""},
		{"rest-view-one", "7\nstring\nundefined\n", "", ""},
		{"rest-view-many", "7\nstring\nstring\n", "", ""},
		{"rest-map-zero", "7\nundefined\nundefined\n", "", ""},
		{"rest-map-one", "7\nstring\nundefined\n", "", ""},
		{"rest-map-many", "7\nstring\nstring\n", "", ""},
		{"rest-view-wrong-tail", "7\nboolean\n", "7\n", "field read failed: selected[3] is not a string; expected string, found boolean"},
		{"rest-map-wrong-contract", "read\n", "", "Map contract failed: target.values; expected ReadonlyMap<number, readonly [number, ...number[]]>, found Map<number, readonly [number, ...string[]]>"},
	})
}
func TestCheckedViewTupleRestAbsentMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_REST_CONTRACT_MUTANT") != "present" {
		t.Skip("opt-in zero-rest-as-present mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "rest-view-zero")
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
	if diff := disagreement(run{stdout: []byte("7\nundefined\nundefined\n")}, actual); diff != "" {
		t.Fatalf("zero-rest-as-present mutant caught: %s; %#v", diff, actual)
	}
	t.Fatal("mutant escaped")
}

func TestCheckedViewTupleRestSchemaMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_REST_CONTRACT_MUTANT") != "schema" {
		t.Skip("opt-in erased rest schema mutant")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "rest-map-wrong-contract")
	var number ir.ViewContractID
	for i, c := range p.ViewContracts {
		if c.Kind == ir.ViewScalar && c.Of == ir.Number {
			number = ir.ViewContractID(i + 1)
			break
		}
	}
	changed := 0
	for i, c := range p.ViewContracts {
		if c.TupleRest != 0 && p.ViewContracts[c.TupleRest-1].Of == ir.String {
			p.ViewContracts[i].TupleRest = number
			changed++
		}
	}
	if number == 0 || changed == 0 {
		t.Fatal("mutant changed no rest schemas")
	}
	actual := onJavaScriptBackend(t, p)
	if actual.exitCode != 70 {
		t.Fatalf("erased rest schema mutant caught: expected named refusal, got %#v", actual)
	}
	t.Fatal("mutant escaped")
}

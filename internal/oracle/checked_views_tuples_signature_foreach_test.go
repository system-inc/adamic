package oracle

import (
	"os"
	"testing"
)

func TestCheckedViewTupleOriginalSignatureForEach(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"signature-foreach-tuple", "object\ndone\n", "", ""},
		{"signature-foreach-scalar", "number\ndone\n", "", ""},
		{"signature-foreach-undefined", "done\n", "", ""},
		{"signature-foreach-wrong-kind", "boolean\ndone\n", "", "cast failed: field read failed: selected[element] matches no member of IncrementalBuildInfoEmitSignature; expected IncrementalBuildInfoEmitSignature, found boolean"},
		{"signature-foreach-wrong-arity", "object\ndone\n", "", "cast failed: field read failed: selected[element] matches no member of IncrementalBuildInfoEmitSignature; expected IncrementalBuildInfoEmitSignature, found object"},
	})
}
func TestCheckedViewTupleOriginalSignatureForEachMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_SIGNATURE_FOREACH_MUTANT") != "skip" {
		t.Skip("opt-in skip tuple alternative guard")
	}
	root, _ := tupleOriginalInputs(t)
	_, p := tupleOriginalProgram(t, root, "signature-foreach-wrong-arity")
	changed := 0
	for i, c := range p.ViewContracts {
		if c.FixedTuple && len(c.Tuple) == 2 {
			p.ViewContracts[i].Tuple = p.ViewContracts[i].Tuple[:1]
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("changed %d tuple alternatives", changed)
	}
	actual := onJavaScriptBackend(t, p)
	if actual.exitCode != 70 {
		t.Fatalf("tuple alternative arity mutant caught: expected named refusal; %#v", actual)
	}
	// An escaped mutant must leave this runner green, exposing the missing kill.
	return
}

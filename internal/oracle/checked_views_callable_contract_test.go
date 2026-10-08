package oracle

import "testing"

// Positive source control for the highest-read fixed representation shape.
// This is not certification of the new shape helper's shared dispatch wiring.
func TestCheckedViewCallableContractControl(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "lane5/probes/number-good")
	want := run{stdout: []byte("8\n")}
	if difference := disagreement(want, onNode(t, path)); difference != "" {
		t.Fatal("Node: " + difference)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
}

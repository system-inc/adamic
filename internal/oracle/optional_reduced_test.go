package oracle

import (
	"path/filepath"
	"testing"
)

func TestOptionalReducedSource(t *testing.T) {
	path, err := filepath.Abs(repository + "/internal/oracle/testdata/optional_widening_reduced.a")
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || string(expected.stdout) != "7\n" {
		t.Fatalf("Node: %#v", expected)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Error(diff)
		}
	}
}

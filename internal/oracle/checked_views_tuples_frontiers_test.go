package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Preserve the exact handed-off input and certify its live array-backed tuple read.
func TestCheckedViewTupleLane4bCastFrontier(t *testing.T) {
	root, _ := tupleOriginalInputs(t)
	source, err := os.ReadFile("../../stage3/interface-downcasts/tuples/frontiers/incremental-tuple-element-frontier.a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.ReplaceAll(string(source), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(root, "compiler/builder.d.ts"))))
	file := filepath.Join(t.TempDir(), "frontier.a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("string\n")}, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("string\n")}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s; got %#v", diff, got)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestCheckedViewTupleLane4bOutSignature(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"lane4b-out-signature-string", "string\n", "", ""},
		{"lane4b-out-signature-tuple", "object\n", "", ""},
		{"lane4b-out-signature-wrong", "boolean\n", "", "field read failed: node.outSignature matches no member of EmitSignature | undefined; expected EmitSignature | undefined, found boolean"},
	})
}

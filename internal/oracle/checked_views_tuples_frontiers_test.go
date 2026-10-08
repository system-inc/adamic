package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Preserve the exact handed-off input. This is a blocker witness, not runtime
// certification: homogeneous array storage cannot be asserted into tuple slots.
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
	_, err = lowered(t, file)
	if err == nil || !strings.Contains(err.Error(), "a cast the runtime can't check") || !strings.Contains(err.Error(), "adamic/no-unchecked-cast") {
		t.Fatal("frontier unexpectedly admitted without an array-to-tuple storage witness")
	}
	t.Logf("original source Node=string; retained compiler boundary: %v", err)
}

func TestCheckedViewTupleLane4bOutSignature(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"lane4b-out-signature-string", "string\n", "", ""},
		{"lane4b-out-signature-tuple", "object\n", "", ""},
		{"lane4b-out-signature-wrong", "boolean\n", "", "field read failed: node.outSignature matches no member of EmitSignature | undefined; expected EmitSignature | undefined, found boolean"},
	})
}

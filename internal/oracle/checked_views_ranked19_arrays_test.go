package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewRanked19OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "19", 2, 22, 21)
}

// Not parallel: the explicit refresh writes this group's rows in the shared counts file.
func TestCheckedViewRanked19ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "19")
}

// The next four-read candidate stops at union admission before its array field is selected.
func TestCheckedViewRanked19NextFrontier(t *testing.T) {
	declarations, directory, _ := originalArrayInputs(t, "19")
	if declarations == "" {
		t.Skip("original19 declaration inputs required")
	}
	input, err := os.ReadFile(filepath.Join(directory, "incremental-file-names-frontier.a"))
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.Replace(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))), 1)
	file := filepath.Join(t.TempDir(), "frontier.a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("a;b\n")}, onNode(t, file)); diff != "" {
		t.Fatal("Node: " + diff)
	}
	_, err = lowered(t, file)
	suffix := "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
	if err == nil || !strings.HasSuffix(err.Error(), suffix) {
		t.Fatalf("union admission frontier changed: %v", err)
	}
	t.Log(err)
}

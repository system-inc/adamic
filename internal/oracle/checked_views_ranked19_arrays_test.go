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

// Preserve the handed-off four-read candidate after lazy common-array admission.
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
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("a;b\n")}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatal(diff)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

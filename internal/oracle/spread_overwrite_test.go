package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A field written after a leading spread, { ...source, k: undefined }, lands in the slot the spread brings
// for it when the literal takes or copies the source object. Family 65 found writes native couldn't keep there
// where Node and the JavaScript backend agree (#xy74m5x, main ac44317c): a number slot written as a reference
// released the number's bits as a pointer (exit 139), and a field optional in the source's type had no slot in
// an actual {} (exit 70 for a string, 139 for a number). Those literals now build their own shape at run time
// (adamic_object_spread_with, #k7g6f9h): spread_with_test.go holds them to Node, and the rows that keep the stop.

// The neighbors still lower and agree with Node: a spread with no overwrite, an overwrite whose value the
// source's slot holds alike over a source that may be undefined (its fallback object is laid out by the
// literal), one over a present source beside a field the spread carries through, and Seats' string row,
// undefined over a string slot, where both are references, so native releases the string and stores undefined.
func TestSpreadOverwriteNeighborsAgree(t *testing.T) {
	t.Parallel()
	for file, node := range map[string]string{"neighbors.ts": "1\nundefined\n2\n2\none 2\n", "string_source.ts": "undefined\nundefined\n"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_overwrite", file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// spreadOverwriteLocation is the path:line:column where text first occurs in the file, as Program.Where prints it.
func spreadOverwriteLocation(path, text string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for number, line := range strings.Split(string(source), "\n") {
		if column := strings.Index(line, text); column >= 0 {
			return fmt.Sprintf("%s:%d:%d", filepath.ToSlash(path), number+1, column+1), nil
		}
	}
	return "", fmt.Errorf("%q isn't in %s", text, path)
}

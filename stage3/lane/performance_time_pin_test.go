package lane

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage3/lane/adaptedtree"
)

// TestPerformanceTimeYewmv46AdaptedProductMatchesLowerPin compares the keyed,
// already built adapted-tree product with Lower's pin. Only the adapted side's
// CRLF is replaced by LF. Nothing is trimmed or otherwise normalized, including
// spaces and the final newline of the exact 16 lines (13 through 28).
func TestPerformanceTimeYewmv46AdaptedProductMatchesLowerPin(t *testing.T) {
	t.Parallel()
	tree := filepath.Join(adaptedtree.Product(t), "tree")
	t.Logf("adapted tree: %s", tree)
	source, err := os.ReadFile(filepath.Join(tree, "src/compiler/performanceCore.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	lines := bytes.SplitAfter(source, []byte("\n"))
	if len(lines) < 28 {
		t.Fatalf("performanceCore.ts has fewer than 28 lines")
	}
	actual := bytes.Join(lines[12:28], nil)
	pin, err := os.ReadFile(filepath.Join(repositoryRoot(t), "internal/lower/testdata/host_binding/performance_core_adapted.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, pin) {
		t.Fatalf("performanceCore.ts lines 13 through 28 disagree with Lower's pin\nadapted: %q\npin: %q", actual, pin)
	}
}

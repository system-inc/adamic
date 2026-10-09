package lint

import (
 "path/filepath"
 "strings"
 "testing"
)

// TestNodeTableIsLinkOnly requires the same output with a copy of every row appended to the node table,
// attached to nothing. Stage 1 reads the table only by following links from the root, and the flat copy
// of typescript-go's tree (#k4fm1vf) depends on it: its tables hold rows no link reaches. A rule or
// harness pass that walks the table by row reports on the copies and fails here.
// Not parallel: uses the shared upstream capture and unmodified port products before timing samples.
// Not parallel: compiler work shares the CPU used by interleaved throughput timing samples.
// Not parallel: uses shared upstream capture and oracle products.
func TestNodeTableIsLinkOnly(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := generated(t)
	for _, row := range upstream(t) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	path := manifest(t, recoveryRows(t, oracle, rows))
	binary := buildPort(t, directory, false)
	plain := execute(t, "", binary, "--manifest", path)
	junk := execute(t, "", binary, "--manifest", path, "--junk-rows")
	if diff := difference(junk.output, plain.output); diff != "" {
		t.Fatalf("output changed with unattached rows in the node table: %s", diff)
	}
	t.Logf("%d rows: identical with and without unattached node rows, %d bytes", len(rows), len(plain.output))
}


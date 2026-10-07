package validation

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: the pinned Go fixture capture sets a process-wide destination.
// This source-only check does not certify emitted JavaScript or native: the
// required runtime RegExp constructor is still refused by the shared lowerer.
func TestSourceNodeParity(t *testing.T) {
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	rows := append(upstream(t), witnesses(t)...)
	custom := filepath.Join(t.TempDir(), "custom.ts")
	if err := os.WriteFile(custom, []byte("const classes='!flex';\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rows = append(rows, custom+"\t"+ruleName(slugs[0])+"\t\t\tfalse\t{\"variables\":[\"^classes$\"]}")
	for i, row := range rows {
		if blocked(row) {
			rows[i] = projected(t, oracle, []string{row})[0]
		}
	}
	path := withVariantFacts(t, oracle, manifest(t, rows))
	want := clean(t, run(t, "", oracle, "--manifest", path))
	entry := filepath.Join(port, "rules", slugs[0], "driver.a")
	got := clean(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repo(t), "oracle/node.mjs"), entry, "--manifest", path))
	if !bytes.Equal(got, want) {
		t.Fatal(difference(got, want))
	}
	t.Logf("source Node and actual Go: %d bytes across %d fixture/witness rows, explicit Go JSX projection and design-system facts; no native/emitted certification", len(want), len(rows))
}

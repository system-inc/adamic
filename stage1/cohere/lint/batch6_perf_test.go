package lint

import (
	"bytes"
	"path/filepath"
	"testing"
)

// Not parallel: executable dispatch mutants are checked against Go and both backends.
func TestBatch6DispatchMutants(t *testing.T) {
	oracle := goOracle(t)
	path := manifest(t, batch6Generated(t))
	want := execute(t, "", oracle, "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"constructor dispatch lost", "case 'Constructor':", "case 'CaseBlock':"},
		{"module statement dispatch lost", "case 'VariableStatement':", "case 'EmptyStatement':"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutantFile(t, "registry.ts", change.from, change.to)
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", node(t, directory, path, false)},
				{"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)},
			} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s mutant survived", side.name)
				}
				t.Logf("caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}

// Not parallel: the minimal compiler/runtime reproducer runs under sanitizers.
func TestBatch6GetterProbe(t *testing.T) {
	directory, err := filepath.Abs("perf_probes/getter")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	source := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "Identifier", "3000000")
	native := execute(t, "", buildPort(t, directory, true), "Identifier", "3000000")
	if !bytes.Equal(source.output, native.output) || string(source.output) != "3000000\n" {
		t.Fatalf("getter probe disagrees: Node=%q native=%q", source.output, native.output)
	}
	t.Logf("Node and sanitized native both print %q; no sanitizer or leak report", source.output)
}

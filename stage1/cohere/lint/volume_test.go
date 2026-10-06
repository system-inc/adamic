package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func volumeGenerated(t *testing.T) []string {
	t.Helper()
	source := "interface Bare { n: number }; interface WrongType { value: number }; type Alias = string; const Choice={Yes:'Yes'} as const; function guard(x: unknown): x is string { return true; } console.log('one'); console['warn']('two'); let count=0; count++; for(let i=0;i<3;i++){count++;}\n"
	path := filepath.Join(t.TempDir(), "volume.ts")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return []string{path, path + "\tno-plusplus\t\t\t\t{\"AllowForLoopAfterthoughts\":true}"}
}

// Not parallel: sanitized mutant builds are bounded and completed before timing.
func TestVolumeMutants(t *testing.T) {
	path := manifest(t, volumeGenerated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"predicate kind omitted", "node.kind === 'TypePredicate' &&", "node.kind === 'NeverKeyword' &&"},
		{"console receiver widened", "this.node(receiver).text === 'console'", "this.node(receiver).text !== 'console'"},
		{"for update option ignored", "this.settings.read('allowforloopafterthoughts', 'false') === 'true'", "this.settings.read('allowforloopafterthoughts', 'false') === 'false'"},
		{"interface accepts type suffix", "? ['Interface', 'Properties', 'Options']", "? ['Type', 'Interface', 'Properties', 'Options']"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to, "volume.ts")
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s survived on %s", change.name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.name, side.name, difference(side.run.output, want))
			}
		})
	}
}

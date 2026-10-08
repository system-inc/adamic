package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These source mutants must compile and run cleanly. Only disagreement with
// the original Node program counts as detection, on every enabled backend.
func TestNodeBufferHost29Mutants(t *testing.T) {
	for _, row := range []struct{ name, before, after string }{
		{"views", "parent.subarray(1, 5)", "Buffer.from(parent.subarray(1, 5))"},
		{"latin1", "Buffer.from(text, 'latin1')", "Buffer.from(text, 'utf8')"},
	} {
		t.Run(row.name, func(t *testing.T) {
			original, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_buffer_"+row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(data), row.before, row.after, 1)
			if source == string(data) {
				t.Fatal("mutant changed nothing")
			}
			path := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, original)
			got, binary := natively(t, program)
			observations := map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program)}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				observations["WASI"] = onWASI(t, native.C(program))
			}
			for backend, observed := range observations {
				if observed.exitCode != 0 || len(observed.stderr) != 0 {
					t.Fatalf("%s mutant failed uncleanly: %+v", backend, observed)
				}
				if difference := disagreement(truth, observed); difference != "stdout differs" {
					t.Fatalf("%s: expected only Node stdout disagreement, got %q", backend, difference)
				}
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Fatal(leaked)
			}
			t.Log("clean mutant caught only by Node stdout comparison")
		})
	}
}

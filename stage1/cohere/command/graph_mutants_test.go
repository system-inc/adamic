package command

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphAndProcessExecutableMutants(t *testing.T) {
	TestOriginalGraphAndProcessBoundaries(t)
	for _, mutant := range []struct{ name, file, from, to string }{
		{"reverse direction", "graph.ts", "row.push(edge.from)", "row.push(toward)"},
		{"closure limit", "graph.ts", "seen.size > 500", "seen.size > 50000"},
		{"child verdict", "process.ts", "error !== '' ? 1 : child.exitCode", "error !== '' ? 1 : 0"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			original := mutantSource(t, mutant.file, mutant.from, mutant.to)
			source := filepath.Join(filepath.Dir(original), "graph_probe.ts")
			program := lowered(t, source)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := graphTests(t, source, binary)
			text := string(result.stdout) + string(result.stderr)
			if result.exitCode == 0 {
				t.Fatal("mutant survived")
			}
			for _, side := range []string{"ADAMIC_NATIVE_PROBE", "ADAMIC_NODE_PROBE", "ADAMIC_BACKEND_PROBE"} {
				if !strings.Contains(text, side) {
					t.Fatalf("missing %s comparison: %s", side, text)
				}
			}
			if strings.Count(text, "error <nil> stderr ") < 3 || strings.Contains(text, "AddressSanitizer") || strings.Contains(text, "runtime error:") {
				t.Fatalf("mutant did not execute cleanly: %s", text)
			}
			if err := os.WriteFile(filepath.Join(t.TempDir(), "failure.log"), []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
			t.Logf("Go command boundary caught %s after clean native, Node and backend execution", mutant.name)
		})
	}
}

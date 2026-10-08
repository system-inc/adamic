package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This is tsc's absent-crypto fallback, not a replacement for crypto.createHash.
// Keep its function bytes tied to upstream so a host scout cannot change the algorithm to fit native.
func TestHostScoutDjb2ShapeAndMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/host_scout_djb2.a"))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := os.ReadFile(filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/sys.ts"))
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(upstream), "\r\n", "\n")
	start := strings.Index(source, "export function generateDjb2Hash(")
	if start < 0 {
		t.Fatal("upstream function not found")
	}
	end := strings.Index(source[start:], "\n}")
	if end < 0 {
		t.Fatal("upstream function end not found")
	}
	if !strings.Contains(string(fixture), source[start:start+end+2]) {
		t.Fatal("fixture changed upstream function bytes")
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node oracle failed: %+v", truth)
	}
	for _, mutant := range []bool{false, true} {
		name := "original"
		selected := path
		if mutant {
			name = "wrong hash seed"
			if strings.Count(string(fixture), "let acc = 5381;") != 1 {
				t.Fatal("hash seed mutation must change exactly one site")
			}
			selected = filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(selected, []byte(strings.Replace(string(fixture), "let acc = 5381;", "let acc = 5382;", 1)), 0600); err != nil {
				t.Fatal(err)
			}
		}
		t.Run(name, func(t *testing.T) {
			program, err := lowered(t, selected)
			if err != nil {
				t.Fatal(err)
			}
			got, binary := nativelyUncached(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("native failed outside output comparison: %+v", got)
			}
			if leaked := leaksUncached(t, program, binary); leaked != "" {
				t.Fatal(leaked)
			}
			wantDifference := ""
			if mutant {
				wantDifference = "stdout differs"
			}
			if difference := disagreement(truth, got); difference != wantDifference {
				t.Fatalf("comparison %q, want %q", difference, wantDifference)
			}
			if mutant {
				t.Log("wrong seed mutant caught only by Node stdout; sanitizers and leak check clean")
			}
		})
	}
}

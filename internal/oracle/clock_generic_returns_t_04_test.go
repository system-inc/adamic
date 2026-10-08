package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/clock_generic_returns_t_04.a", true, false})
}

func TestClockGenericReturnsT04Mutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/clock_generic_returns_t_04.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(source), `typeArguments: ["string", "number"]`, `typeArguments: [] as string[]`, 1)
	if mutant == string(source) {
		t.Fatal("mutant target absent")
	}
	mutated := filepath.Join(t.TempDir(), "clock_generic_returns_t_04_drop_type_arguments.a")
	if err := os.WriteFile(mutated, []byte(mutant), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, mutated)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant must execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("%s: want stdout disagreement, got %q", name, difference)
		}
		t.Logf("clock_generic_returns_t_04_drop_type_arguments %s: Node %q, mutant %q", name, want.stdout, got.stdout)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

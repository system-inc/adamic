package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/representation_clock_children_optional.a", true, false})
}

func TestClockChildrenUndefinedIsEmpty(t *testing.T) {
	t.Run("clock-children-undefined-is-empty", func(t *testing.T) {
		original, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/representation_clock_children_optional.a"))
		if err != nil {
			t.Fatal(err)
		}
		source, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		mutant := strings.Replace(string(source), "? -1 :", "? 0 :", 1)
		if mutant == string(source) {
			t.Fatal("mutant changed nothing")
		}
		path := filepath.Join(t.TempDir(), "mutant.a")
		if err := os.WriteFile(path, []byte(mutant), 0644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		want := onNode(t, original)
		if want.exitCode != 0 || string(want.stdout) != "2\n-1\n" || len(want.stderr) != 0 {
			t.Fatalf("unexpected Node oracle: %+v", want)
		}
		native, binary := natively(t, program)
		if native.exitCode != 0 || len(native.stderr) != 0 {
			t.Fatalf("mutant did not finish cleanly: %+v", native)
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
		for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, got); difference != "stdout differs" {
				t.Fatalf("%s mutant: %q", name, difference)
			}
		}
		t.Logf("stdout kills mutant: Node %q mutant %q", want.stdout, native.stdout)
	})
}

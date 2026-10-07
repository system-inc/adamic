package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

// Register these fixtures without changing the shared oracle harness.
func init() {
	for _, name := range []string{"93122fa_s1.a", "93122fa_s2.a", "switch_empty_neighbors.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Dropping a trailing empty group must fail solely on the source oracle's stdout.
func TestSwitchTrailingEmptyMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/93122fa_s1.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		for at, statement := range function.Body {
			switched, ok := statement.(ir.Switch)
			if !ok || len(switched.Cases) == 0 {
				continue
			}
			last := switched.Cases[len(switched.Cases)-1]
			if len(last.Body) != 0 {
				continue
			}
			switched.Cases = switched.Cases[:len(switched.Cases)-1]
			function.Body[at] = switched
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside stdout comparison: %+v", got)
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("backend mutant caught by %q", difference)
	}
	t.Logf("Node %q; mutant %q; caught only by stdout", truth.stdout, got.stdout)
}

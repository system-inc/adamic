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
func TestSwitchTrailingEmptyMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/93122fa_s1.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
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
		test.Fatal("mutant changed nothing")
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside stdout comparison: %+v", got)
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "stdout differs" {
		test.Fatalf("backend mutant caught by %q", difference)
	}
	test.Logf("Node %q; mutant %q; caught only by stdout", truth.stdout, got.stdout)
}

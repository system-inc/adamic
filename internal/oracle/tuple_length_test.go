package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"93122fa_t_v5.a", "93122fa_t_v1.a", "93122fa_t_v2.a", "93122fa_t_t5.a", "tuple_length_neighbors.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Both mutants finish leak-clean and are caught by Node's stdout alone.
func TestTupleLengthViewMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"93122fa_t_v1.a", "93122fa_t_v2.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
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
				if function.Name != "tuple_view_length" {
					continue
				}
				returned, ok := function.Body[len(function.Body)-1].(ir.Return)
				if !ok {
					t.Fatal("length helper has no return")
				}
				choice, ok := returned.Value.(ir.Conditional)
				if !ok {
					t.Fatal("length helper has no runtime choice")
				}
				choice.Condition = ir.BooleanConstant{Value: false}
				returned.Value = choice
				function.Body[len(function.Body)-1] = returned
				changed = true
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: %+v", got)
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
		})
	}
}

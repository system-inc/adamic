package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_string_split_limit.a", true, false})
}

func TestLibraryStringSplitLimitMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"limit not ToUint32", "limit ignored"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_split_limit.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				slice, ok := value.(ir.ArraySlice)
				if !ok || len(slice.Arguments) != 2 {
					return value
				}
				split, ok := slice.Array.(ir.StringCall)
				if !ok || split.Method != "split" {
					return value
				}
				limit, ok := slice.Arguments[1].(ir.Binary)
				if !ok || limit.Operator != ir.ShiftRightUnsigned {
					return value
				}
				changed = true
				if name == "limit ignored" {
					return split
				}
				slice.Arguments = []ir.Expression{slice.Arguments[0], limit.Left}
				return slice
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no expression")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside stdout comparison: %+v", got)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
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

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
	}{"internal/oracle/testdata/library_string_last_index_of_position.a", true, false})
}

func TestLastIndexOfPositionMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"NaN as zero", "end not widened"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_last_index_of_position.a"))
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
				if function.Name != "library_string_lastIndexOf" {
					continue
				}
				mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
					if name == "NaN as zero" {
						if node, ok := value.(ir.Conditional); ok {
							node.WhenTrue = ir.NumberConstant{Value: 0}
							changed = true
							return node
						}
					} else if node, ok := value.(ir.Binary); ok && node.Operator == ir.Add {
						changed = true
						return node.Left
					}
					return value
				})
			}
			if !changed {
				t.Fatal("mutant changed nothing")
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
				t.Fatalf("native mutant caught by %q", difference)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
				t.Fatalf("JavaScript mutant caught by %q", difference)
			}
			t.Logf("Node %q; mutant %q; caught only by stdout", truth.stdout, got.stdout)
		})
	}
}

package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_string_starts_ends_position.a", true, false})
}

func TestStringStartsEndsPositionMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unclamped start", "position dropped"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_starts_ends_position.a"))
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
				if function.Name != "library_string_startsWith" && function.Name != "library_string_endsWith" {
					continue
				}
				statement := function.Body[0].(ir.Return)
				call := statement.Value.(ir.StringCall)
				slice := call.Value.(ir.StringCall)
				if name == "unclamped start" {
					if call.Method != "startsWith" {
						continue
					}
					slice.Arguments[0] = ir.Read{Local: function.Parameters[2], Of: ir.Number}
					call.Value = slice
				} else {
					call.Value = slice.Value
				}
				statement.Value = call
				function.Body[0] = statement
				changed = true
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
				t.Fatalf("mutant caught by %q", difference)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
				t.Fatalf("backend mutant caught by %q", difference)
			}
			t.Logf("Node %q; mutant %q; caught only by stdout", truth.stdout, got.stdout)
		})
	}
}

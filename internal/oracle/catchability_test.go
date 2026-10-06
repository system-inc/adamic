package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/9984394_lib_dispatch.a", "internal/oracle/testdata/9984394_defined_in_try.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	for _, path := range []string{"internal/oracle/testdata/catchability-refused/9984394_lib_codepoint.a", "internal/oracle/testdata/catchability-refused/9984394_lib_dispatch_codepoint.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, false, false})
	}
}

// Refusal is itself an oracle boundary: Node must catch the failure, and Lower
// must reject a program whose runtime primitive cannot reach that catch yet.
func TestCodePointCatchabilityBoundary(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"9984394_lib_codepoint.a", "9984394_lib_dispatch_codepoint.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/catchability-refused", name))
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte("made 1\nfromCodePoint caught\n")}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatalf("Node probe: %s", difference)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "String.fromCodePoint") {
				t.Fatalf("want String.fromCodePoint catchability refusal, got %v", err)
			}
		})
	}
}

func TestIntegrationCatchabilityMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, fixture string
		mutate        func(*ir.Program) bool
	}{
		{"interface toFixed guard removed", "9984394_lib_dispatch.a", errorGuardMutant("toFixed() digits argument")},
		{"narrowed TypeError becomes panic", "9984394_defined_in_try.a", func(program *ir.Program) bool {
			return changeErrorFunction(program, "error_defined", func(value any) any {
				if thrown, ok := value.(ir.Throw); ok {
					call := thrown.Value.(ir.Call)
					return ir.Panic{Message: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: errorString(program, "TypeError: ")}, call.Arguments[0]}}}
				}
				return value
			})
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutant.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if !mutant.mutate(program) {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, program)
			if difference := disagreement(want, got); difference == "" {
				t.Fatal("mutant survived Node comparison")
			} else {
				t.Logf("caught: %s (Node exit %d, mutant exit %d)", difference, want.exitCode, got.exitCode)
			}
		})
	}
}

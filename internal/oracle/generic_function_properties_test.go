package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/generic-function-property.a", "internal/oracle/testdata/generic_function_properties.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Mutating the concrete number body preserves dispatch and ownership, so only
// the external Node observation can catch the wrong returned result.
func TestGenericFunctionPropertyMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_function_properties.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if !strings.HasPrefix(function.Name, "generic_closure_") || function.Returns != ir.Number {
			continue
		}
		for index, statement := range function.Body {
			if returned, ok := statement.(ir.Return); ok && returned.Value != nil {
				call, ok := returned.Value.(ir.CallClosure)
				if !ok {
					continue
				}
				if _, forwarding := call.Closure.(ir.MakeClosure); !forwarding {
					continue
				}
				returned.Value = ir.Binary{Operator: ir.Add, Left: returned.Value, Right: ir.NumberConstant{Value: 1}}
				function.Body[index] = returned
				changed++
			}
		}
	}
	if changed == 0 {
		t.Fatal("number specialization absent")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	original := onNode(t, path)
	for name, result := range map[string]run{"native": execute(t, binary), "JavaScript": onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s mutant failed before comparison: %d %s", name, result.exitCode, result.stderr)
		}
		if difference := disagreement(original, result); difference != "stdout differs" {
			t.Fatalf("%s: want stdout differs, got %q", name, difference)
		}
		t.Logf("%s mutant caught only by Node stdout", name)
	}
}

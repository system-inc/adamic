package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var conditionRepresentations = []string{"number", "string", "object", "array", "function", "optional_reference", "optional_number", "optional_boolean", "optional_string", "null", "union", "sites", "evaluation"}

func init() {
	for _, name := range conditionRepresentations {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/conditions_" + name + ".a", true, false})
	}
}

func TestTypeScriptConditionsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range conditionRepresentations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/conditions_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			// Store source as .a in the repository; exercise the TypeScript loader explicitly.
			path := filepath.Join(t.TempDir(), "conditions.ts")
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			native, binary := nativelyUncached(t, program)
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			t.Logf("Node stdout %q; both backends agree", truth.stdout)
		})
	}
}

func TestConditionNaNMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/conditions_number.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	changed := 0
	mutate := func(node any) any {
		if call, ok := node.(ir.NumberCall); ok && call.Function == "toBoolean" && call.Arguments[0].Type() == ir.Number {
			changed++
			// A naive nonzero comparison incorrectly makes NaN truthy.
			return ir.Binary{Operator: ir.NotEqual, Left: call.Arguments[0], Right: ir.NumberConstant{Value: 0}}
		}
		return node
	}
	for i := range program.Functions {
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, mutate)
	}
	if changed == 0 {
		t.Fatal("NaN mutant changed no conditions")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("mutant was not caught only by stdout: %#v", got)
		}
	}
	t.Logf("NaN-as-truthy mutant caught in both backends by stdout; changed %d conditions", changed)
}

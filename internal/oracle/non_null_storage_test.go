package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"local", "field", "static", "identifier", "assigned"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_storage_" + name + ".a", true, name != "assigned"})
	}
}

func TestNonNullStorageReadinessMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, kind, expression, node string }{
		{"local", "variable", "value", "value"},
		{"field", "field", "state.value", "value"},
		{"static", "field", "State.value", "value"},
		{"identifier", "field", "node.escapedText", "escapedText"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_storage_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			sourceValue := "undefined"
			if probe.name == "identifier" {
				sourceValue = "true"
			}
			if probe.name == "field" {
				sourceValue = "null"
			}
			if difference := disagreement(run{stdout: []byte("before\n" + sourceValue + "\n")}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: read before assignment: " + probe.kind + " '" + probe.node + "' in " + probe.expression + "\n"), exitCode: 70}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			changes := 0
			mutate := func(node any) any {
				switch value := node.(type) {
				case ir.Read:
					if probe.kind == "variable" && value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				case ir.Property:
					if probe.kind == "field" && value.Name == probe.node && value.Readiness != "" {
						value.Readiness = ""
						changes++
						return value
					}
				}
				return node
			}
			program.Main = mutateReadiness(program.Main, mutate)
			for i := range program.Functions {
				program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, mutate)
			}
			if changes != 1 {
				t.Fatalf("removed %d reads, want one", changes)
			}
			mutant, _ := nativelyUncached(t, program)
			if mutant.exitCode != 0 || disagreement(want, mutant) == "" {
				t.Fatalf("mutant escaped runtime output pin: %#v", mutant)
			}
			t.Logf("drop %s readiness caught: stdout %q, exit %d", probe.name, mutant.stdout, mutant.exitCode)
		})
	}
}

func TestNonNullStorageUnionRemainsRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_refuse_storage_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("before\nundefined\n")}, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "an uninitialized object property without a supported stored type" || refused.Fix != "use a supported scalar or reference field type" {
		t.Fatalf("want unsupported structural union slot refusal, got %v", err)
	}
}

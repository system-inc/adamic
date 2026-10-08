package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCheckedViewsNextMutants(t *testing.T) {
	for _, fixture := range []string{"optional-boolean-misfit", "optional-false-misfit", "required-boolean-misfit", "branch-file-misfit", "branch-start-misfit", "branch-length-misfit", "branch-bottom-number-misfit", "branch-bottom-string-misfit", "generic-site-misfit", "overload-result-misfit", "overload-callback-misfit"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := checkedWriteFixture(t, fixture)
			changes := 0
			change := func(value ir.Expression) ir.Expression {
				if result, ok := value.(ir.ContractResult); ok {
					changes++
					return result.Value
				}
				if array, ok := value.(ir.ArrayLiteral); ok && array.Never {
					array.Never = false
					if fixture == "branch-bottom-string-misfit" {
						array.Element = ir.String
					}
					changes++
					return array
				}
				if object, ok := value.(ir.ObjectLiteral); ok {
					for i, field := range object.Fields {
						if field.Contract == nil {
							continue
						}
						contract := *field.Contract
						switch {
						case strings.HasPrefix(fixture, "optional-") && field.Name == "multiLine":
							contract.Allowed = nil
						case fixture == "required-boolean-misfit" && field.Name == "multiLine":
							contract.Nullable = true
						case strings.HasPrefix(fixture, "branch-") && field.Name == strings.TrimSuffix(strings.TrimPrefix(fixture, "branch-"), "-misfit"):
							contract.Nullable = true
						case fixture == "generic-site-misfit" && field.Name == "flags":
							contract.Allowed = nil
						default:
							continue
						}
						object.Fields[i].Contract = &contract
						changes++
					}
					return object
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
			if changes == 0 {
				t.Fatal("mutant changed no contract")
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			got, binary := nativelyUncached(t, program)
			for name, result := range map[string]run{"native sanitized": got, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, result); difference != "" {
					t.Fatalf("%s mutant must run Node behavior: %s", name, difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("contract mutant changed %d contracts; original exit 70 is caught, mutant matches Node", changes)
		})
	}
}

func TestCheckedViewsNextAdamicRefusals(t *testing.T) {
	for _, name := range []string{"refused-optional-boolean", "refused-branch-diagnostic", "refused-branch-bottom", "refused-overload-result"} {
		path, pathError := filepath.Abs(filepath.Join(repository, "stage3/checked-writes", name+".a"))
		if pathError != nil {
			t.Fatal(pathError)
		}
		_, err := lowered(t, path)
		expected, readError := os.ReadFile(path + ".refused")
		if readError != nil {
			t.Fatal(readError)
		}
		relative := "stage3/checked-writes/" + name + ".a"
		if err == nil || strings.ReplaceAll(err.Error(), path, relative) != strings.TrimSpace(string(expected)) {
			t.Fatalf("%s: expected pinned path and fix %s, got %v", name, expected, err)
		}
		t.Log(err)
	}
}

func TestCheckedViewsNextProvenAdamic(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/checked-writes/proven-generic.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.WriteChecks) != 0 {
		t.Fatal("proven generic inserted a check")
	}
	truth := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	for name, result := range map[string]run{"native sanitized": got, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, result); d != "" {
			t.Fatalf("%s: %s", name, d)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var conditionAssertionFixtures = []struct {
	name, stdout, prefix string
	proven, stops        bool
}{
	{"proven", "1\ncontinued\n", "", true, false},
	{"proven_narrow", "5\n", "", true, false},
	{"lying", "continued\n", "", false, true},
	{"gated", "continued\n", "", false, true},
	{"saved", "1\ncontinued\n", "", false, false},
	{"false_saved", "1\ncontinued\n", "1\n", false, true},
}

func conditionAssertionInput(t *testing.T, name string, checked bool) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/condition_assertions", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	if !checked {
		return path
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	copy := filepath.Join(t.TempDir(), name+".ts")
	if err = os.WriteFile(copy, source, 0644); err != nil {
		t.Fatal(err)
	}
	return copy
}
func TestConditionAssertionsOracle(t *testing.T) {
	for _, fixture := range conditionAssertionFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			node := onNode(t, conditionAssertionInput(t, fixture.name, false))
			if difference := disagreement(run{stdout: []byte(fixture.stdout)}, node); difference != "" {
				t.Fatal(difference)
			}
			original, err := lowered(t, conditionAssertionInput(t, fixture.name, false))
			if fixture.proven {
				if err != nil {
					t.Fatal(err)
				}
				got, binary := nativelyUncached(t, original)
				for _, result := range []run{got, onJavaScriptBackend(t, original)} {
					if difference := disagreement(node, result); difference != "" {
						t.Fatal(difference)
					}
				}
				if report := leaksUncached(t, original, binary); report != "" {
					t.Fatal(report)
				}
			} else if err == nil || !strings.Contains(err.Error(), "type predicate") {
				t.Fatalf("unproved .a assertion admitted: %v", err)
			}
			program, err := lowered(t, conditionAssertionInput(t, fixture.name, true))
			if err != nil {
				t.Fatal(err)
			}
			if len(program.PredicateChecks.Sites) != 1 {
				t.Fatalf("missing assertion site: %+v", program.PredicateChecks)
			}
			if fixture.proven && (program.PredicateChecks.Proven != 1 || program.PredicateChecks.Checked != 0) {
				t.Fatalf("proof did not retire check: %+v", program.PredicateChecks)
			}
			if !fixture.proven && program.PredicateChecks.Checked != 1 {
				t.Fatalf("unproved assertion lacks check: %+v", program.PredicateChecks)
			}
			want := node
			if fixture.stops {
				site := program.PredicateChecks.Sites[0]
				source := "false"
				if fixture.name == "false_saved" {
					source = "boolean"
				}
				want = run{stdout: []byte(fixture.prefix), stderr: []byte(fmt.Sprintf("adamic: panic: predicate assertCondition at %s asserts branch: source %s, target truthy condition\n", site.Where, source)), exitCode: 70}
			}
			got, binary := nativelyUncached(t, program)
			for _, result := range []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Fatalf("%s: got %q %q exit %d", difference, result.stdout, result.stderr, result.exitCode)
				}
			}
			if !fixture.stops {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if fixture.stops {
				changes := 0
				for i := range program.Functions {
					if strings.Contains(program.Functions[i].Name, "_predicate_check_") {
						program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(n any) any {
							if _, ok := n.(ir.Panic); ok {
								changes++
								return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
							}
							return n
						})
					}
				}
				if changes != 1 {
					t.Fatalf("dropped %d checks", changes)
				}
				mutant, _ := nativelyUncached(t, program)
				for _, result := range []run{mutant, onJavaScriptBackend(t, program)} {
					if disagreement(want, result) == "" {
						t.Fatal("dropped assertion check escaped")
					}
					if difference := disagreement(node, result); difference != "" {
						t.Fatal("drop-check mutant did not reproduce Node: " + difference)
					}
				}
				t.Log("dropped check lets the falsy value through and is caught in both backends")
			}
		})
	}
}
func TestConditionAssertionReevaluationMutant(t *testing.T) {
	path := conditionAssertionInput(t, "saved", true)
	node := onNode(t, conditionAssertionInput(t, "saved", false))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var argument ir.Expression
	mutateReadiness(program.Main, func(n any) any {
		if call, ok := n.(ir.Call); ok && strings.Contains(program.Functions[call.Function].Name, "_predicate_check_") {
			argument = call.Arguments[0]
		}
		return n
	})
	if argument == nil {
		t.Fatal("missing saved original argument")
	}
	changes := 0
	for i := range program.Functions {
		if strings.Contains(program.Functions[i].Name, "_predicate_check_") {
			program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(n any) any {
				if truth, ok := n.(ir.Truthy); ok {
					changes++
					truth.Value = argument
					return truth
				}
				return n
			})
		}
	}
	if changes != 1 {
		t.Fatalf("changed %d truthiness tests", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, result := range []run{mutant, onJavaScriptBackend(t, program)} {
		if disagreement(node, result) == "" || result.exitCode != 70 || string(result.stdout) != "1\n2\n" {
			t.Fatalf("reevaluation mutant not caught: %q %q exit %d", result.stdout, result.stderr, result.exitCode)
		}
	}
	t.Log("reevaluation mutant runs the condition twice and exits 70; both backends disagree with Node's single read")
}
func init() { additionalFixtureCounts = append(additionalFixtureCounts, conditionAssertionCountRows) }
func conditionAssertionCountRows(t *testing.T) []string {
	var rows []string
	for _, fixture := range conditionAssertionFixtures {
		modes := []bool{true}
		if fixture.proven {
			modes = append(modes, false)
		}
		for _, checked := range modes {
			t.Run(fmt.Sprintf("condition-counts/%s/%t", fixture.name, checked), func(t *testing.T) {
				program, err := lowered(t, conditionAssertionInput(t, fixture.name, checked))
				if err != nil {
					t.Fatal(err)
				}
				binary := filepath.Join(t.TempDir(), "counted")
				if err = native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				result := execute(t, binary)
				match := countsLine.FindSubmatch(result.stderr)
				if match == nil {
					t.Fatal("missing allocation counts")
				}
				path := "internal/lower/testdata/condition_assertions/" + fixture.name + ".a"
				if checked {
					path += " (.ts checked mode)"
				}
				rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6]))
			})
		}
	}
	return rows
}

func TestConditionAssertionCountsAreRecorded(t *testing.T) {
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range conditionAssertionCountRows(t) {
		if !strings.Contains(string(recorded), row+"\n") {
			t.Fatalf("missing condition allocation row: %s", row)
		}
	}
	index := strings.Index(string(recorded), predicateCountsHeader)
	if index < 0 || string(recorded[index:]) != predicateCountsTable(t) {
		t.Fatal("condition direction counts changed")
	}
}

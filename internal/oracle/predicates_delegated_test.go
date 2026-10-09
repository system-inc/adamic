package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var delegatedPredicateFixtures = []struct {
	name, stdout, predicate string
	proven                  bool
}{
	{"primitive", "true:true\ntrue:false\nfalse:false\n", "", true},
	{"literals", "true:true:false\ntrue:false\n", "", true},
	{"switch", "true:true\ntrue:false\nfalse:false\nfalse:false\n", "", true},
	{"paths", "true:true\nfalse:false\n", "", true},
	{"lying", "true\n", "helper", false},
	{"negation", "true\n", "delegated", false},
	{"cycle", "true\n", "first", false},
	{"effect", "true\n", "delegated", false},
	{"identity", "true\n", "delegated", false},
	{"mixed_false", "text\n", "delegated", false},
}

func delegatedPredicateInput(t *testing.T, name string, checked bool) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/predicates_delegated", name+".a"))
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
func delegatedPredicateOracle(t *testing.T, name string) {
	t.Helper()
	var fixtureIndex int
	for i, fixture := range delegatedPredicateFixtures {
		if fixture.name == name {
			fixtureIndex = i
		}
	}
	fixture := delegatedPredicateFixtures[fixtureIndex]
	node := onNode(t, delegatedPredicateInput(t, name, false))
	if difference := disagreement(run{stdout: []byte(fixture.stdout)}, node); difference != "" {
		t.Fatal("source Node: " + difference)
	}
	for _, checked := range []bool{false, true} {
		program, err := lowered(t, delegatedPredicateInput(t, name, checked))
		if !fixture.proven && !checked {
			if err == nil || !strings.Contains(err.Error(), "type predicate") {
				t.Fatalf("unproved .a body admitted: %v", err)
			}
			t.Logf(".a refuses: %v", err)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.proven {
			if program.PredicateChecks.Checked != 0 {
				t.Fatalf("proved helpers retain checks: %+v", program.PredicateChecks)
			}
		} else if program.PredicateChecks.Checked == 0 {
			t.Fatal("unproved helper lost its check")
		}
		got, binary := nativelyUncached(t, program)
		results := []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)}
		if fixture.proven {
			for _, result := range results {
				if difference := disagreement(node, result); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			// A semantically wrong helper return must disagree with source Node
			// after the direction checks have been retired by a body proof.
			if !checked {
				helper := map[string]string{"primitive": "isString", "literals": "isPlus", "switch": "selectedKind", "paths": "stringBody"}[name]
				changes := 0
				for index := range program.Functions {
					if program.Functions[index].Name == helper {
						program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, func(node any) any {
							if returned, ok := node.(ir.Return); ok {
								changes++
								returned.Value = ir.BooleanConstant{Value: false}
								return returned
							}
							return node
						})
					}
				}
				if changes == 0 {
					t.Fatal("default-return mutant changed nothing")
				}
				mutant, _ := nativelyUncached(t, program)
				for _, result := range []run{mutant, onJavaScriptBackend(t, program)} {
					if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(node, result) == "" {
						t.Fatalf("default-return mutant was not caught solely by output: %+v", result)
					}
				}
				t.Log("default-false helper mutant caught by Node output in both backends")
			}
		} else {
			branch := "true"
			if name == "mixed_false" {
				branch = "false"
			}
			for _, result := range results {
				if result.exitCode != 70 || len(result.stdout) != 0 || !strings.Contains(string(result.stderr), "predicate "+fixture.predicate+" at ") || !strings.Contains(string(result.stderr), branch+" branch: source ") || !strings.Contains(string(result.stderr), ", target ") {
					t.Fatalf("missing checked failure: stdout=%q stderr=%q exit=%d", result.stdout, result.stderr, result.exitCode)
				}
				if difference := disagreement(results[0], result); difference != "" {
					t.Fatal(difference)
				}
			}
			t.Logf("Node %q exit 0; compiled check %q exit 70", node.stdout, got.stderr)
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
			if changes == 0 {
				t.Fatal("removed-check mutant changed nothing")
			}
			mutant, _ := nativelyUncached(t, program)
			for _, result := range []run{mutant, onJavaScriptBackend(t, program)} {
				if name != "mixed_false" {
					if difference := disagreement(node, result); difference != "" {
						t.Fatal("removed guards did not reproduce Node: " + difference)
					}
				} else if strings.Contains(string(result.stderr), "predicate delegated") {
					t.Fatal("removed false predicate check still ran")
				}
				if disagreement(got, result) == "" {
					t.Fatal("removed predicate checks escaped")
				}
			}
			t.Log("removed-check mutant caught in native and JavaScript")
		}
	}
}
func TestDelegatedPredicatePrimitiveOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "primitive")
}
func TestDelegatedPredicateLiteralOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "literals")
}
func TestDelegatedPredicateSwitchOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "switch")
}
func TestDelegatedPredicatePathsOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "paths")
}
func TestDelegatedPredicateLyingHelperOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "lying")
}
func TestDelegatedPredicateNegationOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "negation")
}
func TestDelegatedPredicateCycleOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "cycle")
}
func TestDelegatedPredicateEffectOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "effect")
}
func TestDelegatedPredicateIdentityOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "identity")
}
func init() { additionalFixtureCounts = append(additionalFixtureCounts, delegatedPredicateCountRows) }
func delegatedPredicateCountRows(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, fixture := range delegatedPredicateFixtures {
		checked := !fixture.proven
		program, err := lowered(t, delegatedPredicateInput(t, fixture.name, checked))
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
		name := "internal/lower/testdata/predicates_delegated/" + fixture.name + ".a"
		if checked {
			name += " (.ts checked mode)"
		}
		rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", name, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	return rows
}
func TestDelegatedPredicateCountsAreRecorded(t *testing.T) {
	t.Parallel()
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range delegatedPredicateCountRows(t) {
		if !strings.Contains(string(recorded), row+"\n") {
			t.Fatalf("unrecorded delegation allocation row: %s", row)
		}
	}
}

func TestDelegatedPredicateMixedFalseOracle(t *testing.T) {
	t.Parallel()
	delegatedPredicateOracle(t, "mixed_false")
}

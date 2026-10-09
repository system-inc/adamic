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

var checkedPredicateFixtures = []struct {
	name, nodeOut, checkedOut, branch, source string
	checked, unobservable                     int
}{
	{"true", "claimed\n", "", "true", "string | number", 2, 0},
	{"true_unused", "continued\n", "", "true", "\"text\"", 1, 1},
	{"false", "continued\n", "", "false", "string | number", 2, 0},
	{"false_alias", "continued\n", "", "false", "string | number", 2, 0},
	{"false_compound_unchanged", "continued\n", "", "", "", 1, 1},
	{"asserts", "called\nasserted\n", "called\n", "asserts", "string | number", 1, 0},
	{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", "", "", 1, 1},
	{"asserts_valid", "called\n7\n", "", "", "", 1, 0},
	{"false_unchanged", "continued\n", "", "", "", 1, 1},
}

func predicateCheckedInput(t *testing.T, name string) (string, string) {
	t.Helper()
	witness, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/predicates_checked", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(witness)
	if err != nil {
		t.Fatal(err)
	}
	checked := filepath.Join(t.TempDir(), name+".ts")
	if err = os.WriteFile(checked, source, 0644); err != nil {
		t.Fatal(err)
	}
	return witness, checked
}

func TestCheckedPredicateOracle(t *testing.T) {
	for _, probe := range checkedPredicateFixtures {
		t.Run(probe.name, func(t *testing.T) {
			witness, path := predicateCheckedInput(t, probe.name)
			node := onNode(t, witness)
			expectedNode := run{stdout: []byte(probe.nodeOut)}
			if difference := disagreement(expectedNode, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			counts := program.PredicateChecks
			if counts.Checked != probe.checked || counts.Unobservable != probe.unobservable || counts.Proven != 0 || len(counts.Sites) != 1 {
				t.Fatalf("direction counts %+v", counts)
			}
			want := node
			if probe.branch != "" {
				site := counts.Sites[0]
				// The expected source and target strings are pins, not emitted diagnostics.
				message := fmt.Sprintf("adamic: panic: predicate %s at %s %s branch: source %s, target number\n", site.Function, site.Where, probe.branch, probe.source)
				want = run{stdout: []byte(probe.checkedOut), stderr: []byte(message), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				countedBinary := filepath.Join(t.TempDir(), "counted")
				if err = native.Build(native.C(program), countedBinary, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				if report := unbalanced(t, execute(t, countedBinary)); report != "" {
					t.Fatal(report)
				}
			}
			if probe.branch != "" {
				changes := 0
				for i := range program.Functions {
					if strings.Contains(program.Functions[i].Name, "_predicate_check_") {
						program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
							if panic, ok := node.(ir.Panic); ok {
								constant, ok := panic.Message.(ir.StringConstant)
								if ok && strings.Contains(program.Strings[constant.Index], " "+probe.branch+" branch:") {
									changes++
									return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
								}
							}
							return node
						})
					}
				}
				if changes != 1 {
					t.Fatalf("check mutant removed %d guards", changes)
				}
				mutant, _ := nativelyUncached(t, program)
				for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
					if disagreement(want, got) == "" {
						t.Fatal("removed-check mutant escaped exit-70 contract")
					}
					if difference := disagreement(node, got); difference != "" {
						t.Fatal("mutant failed to execute source semantics: " + difference)
					}
				}
				t.Logf("%s lying body exits 70; removed-%s-check mutant executes Node result %q and is caught", probe.name, probe.branch, node.stdout)
			}
		})
	}
}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, checkedPredicateCountRows) }
func checkedPredicateCountRows(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, probe := range checkedPredicateFixtures {
		t.Run("checked-predicate-counts/"+probe.name, func(t *testing.T) {
			witness, path := predicateCheckedInput(t, probe.name)
			program, err := lowered(t, path)
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
				t.Fatalf("no counts: %#v", result)
			}
			root, err := filepath.Abs(repository)
			if err != nil {
				t.Fatal(err)
			}
			relative, err := filepath.Rel(root, witness)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, fmt.Sprintf("| %s (.ts checked mode) | %s | %s | %s | %s | %s | %s |", filepath.ToSlash(relative), match[1], match[2], match[3], match[4], match[5], match[6]))
		})
	}
	return rows
}

// The global writer registers these adapted inputs. This focused verifier holds
// only their rows and the predicate table to the recorded counts.
func TestCheckedPredicateCountsAreRecorded(t *testing.T) {
	rows := checkedPredicateCountRows(t)
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if !strings.Contains(string(recorded), row+"\n") {
			t.Fatalf("unrecorded checked predicate row: %s", row)
		}
	}
	index := strings.Index(string(recorded), predicateCountsHeader)
	if index < 0 {
		t.Fatal("missing predicate table")
	}
	if table := predicateCountsTable(t); string(recorded[index:]) != table {
		t.Fatal("predicate direction table changed")
	}
}

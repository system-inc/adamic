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

var predicateKindFixtures = []struct{ name, stdout string }{
	{"enum", "identifier:name\nliteral:literal\n"},
	{"alias", "name\nother\n"},
	{"optional", "absent\npresent\nother\n"},
	{"fields", "undefined\n"},
	{"optional_read", "absent:missing\npresent:value\nother\n"},
	{"optional_alias", "missing:false:missing\nundefined:true:missing\nvalue:true:payload\n"},
	{"optional_wrong", "undefined:missing\n"},
	{"optional_receiver", "0\n80\nvalue\n"},
}

func predicateKindInput(t *testing.T, name string, checked bool) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/predicates_kind", name+".a"))
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
	target := filepath.Join(t.TempDir(), name+".ts")
	if err = os.WriteFile(target, source, 0644); err != nil {
		t.Fatal(err)
	}
	return target
}

func TestPredicateKindProofOracle(t *testing.T) {
	for _, probe := range predicateKindFixtures {
		t.Run(probe.name, func(t *testing.T) {
			node := onNode(t, predicateKindInput(t, probe.name, false))
			if difference := disagreement(run{stdout: []byte(probe.stdout)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, checked := range []bool{false, true} {
				t.Run(fmt.Sprint("checked-source-", checked), func(t *testing.T) {
					program, err := lowered(t, predicateKindInput(t, probe.name, checked))
					if err != nil {
						t.Fatal(err)
					}
					if program.PredicateChecks.Checked != 0 {
						t.Fatalf("proven kind retains checks: %+v", program.PredicateChecks)
					}
					if checked && (program.PredicateChecks.Proven != 2 || len(program.PredicateChecks.Sites) != 1) {
						t.Fatalf("missing directional proof: %+v", program.PredicateChecks)
					}
					for _, function := range program.Functions {
						if strings.Contains(function.Name, "_predicate_check_") {
							t.Fatal("body proof retained a predicate wrapper")
						}
					}
					if probe.name != "enum" && !program.CheckedFields["escapedText"] {
						t.Fatal("open tag proof lost its remaining field view")
					}
					want := node
					if probe.name == "fields" {
						want = run{stderr: []byte("adamic: panic: field read failed: node.escapedText is not initialized; expected string, found missing\n"), exitCode: 70}
					}
					if probe.name == "optional_wrong" {
						want = run{stderr: []byte("adamic: panic: field read failed: node.text is not a string; expected string, found nullish\n"), exitCode: 70}
					}
					sanitized, binary := nativelyUncached(t, program)
					for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if difference := disagreement(want, got); difference != "" {
							t.Fatalf("%s: got stdout %q stderr %q exit %d", difference, got.stdout, got.stderr, got.exitCode)
						}
					}
					if want.exitCode == 0 {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
						counted := filepath.Join(t.TempDir(), "counted")
						if err = native.Build(native.C(program), counted, native.Options{Count: true}); err != nil {
							t.Fatal(err)
						}
						if report := unbalanced(t, execute(t, counted)); report != "" {
							t.Fatal(report)
						}
					}
					if !checked && probe.name != "fields" && probe.name != "optional_wrong" {
						changes := 0
						for i := range program.Functions {
							if program.Functions[i].Name == "isIdentifier" {
								program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
									if returned, ok := node.(ir.Return); ok {
										changes++
										returned.Value = ir.BooleanConstant{Value: false}
										return returned
									}
									return node
								})
							}
						}
						if changes != 1 {
							t.Fatalf("default-false body mutant changed %d returns", changes)
						}
						mutant, _ := nativelyUncached(t, program)
						for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
							if disagreement(node, got) == "" {
								for index, function := range program.Functions {
									t.Logf("function %d %s body %#v", index, function.Name, function.Body)
								}
								t.Fatalf("default-false kind body mutant escaped source Node: stdout %q", got.stdout)
							}
						}
						t.Logf("%s default-false proof mutant caught by native and JavaScript Node disagreement", probe.name)
					}
				})
			}
		})
	}
}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, predicateKindCountRows) }
func predicateKindCountRows(t *testing.T) []string {
	var rows []string
	for _, probe := range predicateKindFixtures {
		t.Run("kind-proof-counts/"+probe.name, func(t *testing.T) {
			path := predicateKindInput(t, probe.name, false)
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
			relative, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", filepath.ToSlash(relative), match[1], match[2], match[3], match[4], match[5], match[6]))
		})
	}
	return rows
}

func TestPredicateKindCountsAreRecorded(t *testing.T) {
	rows := predicateKindCountRows(t)
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if !strings.Contains(string(recorded), row+"\n") {
			t.Fatalf("unrecorded kind proof row: %s", row)
		}
	}
}

// A payload-based presence mutant collapses present-undefined into missing.
func TestPredicateOptionalAliasPresenceMutant(t *testing.T) {
	path := predicateKindInput(t, "optional_alias", false)
	node := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := range program.Functions {
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(n any) any {
			if call, ok := n.(ir.HasProperty); ok && call.Name == "text" {
				changed++
				return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: ir.Property{Object: ir.Narrow{Value: call.Object, To: ir.Object}, Name: "text", Of: ir.String, Absent: true}}}
			}
			return n
		})
	}
	if changed != 1 {
		t.Fatalf("changed %d presence queries", changed)
	}
	got, _ := nativelyUncached(t, program)
	for _, result := range []run{got, onJavaScriptBackend(t, program)} {
		if disagreement(node, result) == "" {
			t.Fatal("payload-based presence mutant escaped Node")
		}
		t.Logf("presence mutant caught: stdout=%q stderr=%q exit=%d", result.stdout, result.stderr, result.exitCode)
	}
}

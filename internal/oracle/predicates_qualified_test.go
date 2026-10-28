package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

var qualifiedPredicateShapes = []string{"namespace", "type", "each", "defined", "reexport", "module", "nested", "false_read"}

func init() {
	for _, shape := range qualifiedPredicateShapes {
		for _, suffix := range []string{"_truth", "_lie"} {
			fixtures = append(fixtures, struct {
				path            string
				lowers, checked bool
			}{"internal/oracle/testdata/predicate_qualified/" + shape + suffix + "/main.ts", true, suffix == "_lie"})
		}
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/predicate_qualified/void/main.ts", true, false})
}

// Each road's truthful call agrees with source Node; each lie must panic at the
// qualified call on JS and native under ASan/UBSan, before the following print.
// In particular these catch callFunction bypassing censusOverloadResult for a
// namespace, a re-export chain, or an ESM namespace import.
func qualifiedPredicateAgreement(t *testing.T, shape string) {
	t.Helper()
	for _, suffix := range []string{"_truth", "_lie"} {
		t.Run(shape+suffix, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_qualified", shape+suffix, "main.ts"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			sourceOut := "called\nafter\n"
			if shape == "false_read" && suffix == "_lie" {
				sourceOut = "called\nundefined\n"
			}
			if d := disagreement(run{stdout: []byte(sourceOut)}, node); d != "" {
				t.Fatal("source Node: " + d)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			js := onJavaScriptBackend(t, program)
			native, _ := nativelyUncached(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"JavaScript", js}, {"ASan/UBSan native", native}} {
				got := backend.got
				if suffix == "_lie" {
					branch := "asserts"
					if shape == "nested" {
						branch = "true"
					}
					if shape == "false_read" {
						branch = "false"
					}
					if got.exitCode != 70 || string(got.stdout) != "called\n" || !strings.HasPrefix(string(got.stderr), "adamic: panic: predicate ") || !strings.Contains(string(got.stderr), path+":") || !strings.Contains(string(got.stderr), " "+branch+" branch:") {
						t.Errorf("%s qualified lie did not stop at the %s call check: exit %d stdout %q stderr %q", backend.name, branch, got.exitCode, got.stdout, got.stderr)
					}
				} else if d := disagreement(node, got); d != "" {
					t.Error(backend.name + ": " + d)
				}
			}
			if d := disagreement(js, native); d != "" {
				t.Fatal("ASan/UBSan native: " + d)
			}
		})
	}
}

func TestQualifiedPredicateNamespaceAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "namespace")
}
func TestQualifiedPredicateGenericTypeAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "type")
}
func TestQualifiedPredicateArrayAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "each")
}
func TestQualifiedPredicateDefinedAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "defined")
}
func TestQualifiedPredicateReexportAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "reexport")
}
func TestQualifiedPredicateModuleAssertions(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "module")
}
func TestQualifiedPredicateNestedNamespace(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "nested")
}
func TestQualifiedPredicateFalseDirection(t *testing.T) {
	t.Parallel()
	qualifiedPredicateAgreement(t, "false_read")
}

func TestQualifiedPredicateVoidOverloadDoesNotCheck(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_qualified/void/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if d := disagreement(run{stdout: []byte("called\nafter\n")}, node); d != "" {
		t.Fatal(d)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	js := onJavaScriptBackend(t, program)
	if d := disagreement(node, js); d != "" {
		t.Fatal(d)
	}
	native, _ := nativelyUncached(t, program)
	if d := disagreement(node, native); d != "" {
		t.Fatal(d)
	}
}

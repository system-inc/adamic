package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

var qualifiedPredicateShapes = []string{
	"namespace",
	"type",
	"each",
	"defined",
	"reexport",
	"module",
	"nested",
	"false_read",
}

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
func qualifiedPredicateAgreement(test *testing.T, shape string) {
	test.Helper()
	for _, suffix := range []string{"_truth", "_lie"} {
		test.Run(shape+suffix, func(test *testing.T) {
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/predicate_qualified", shape+suffix, "main.ts"),
			)
			if err != nil {
				test.Fatal(err)
			}
			node := onNode(test, path)
			sourceOut := "called\nafter\n"
			if shape == "false_read" && suffix == "_lie" {
				sourceOut = "called\nundefined\n"
			}
			if d := disagreement(run{stdout: []byte(sourceOut)}, node); d != "" {
				test.Fatal("source Node: " + d)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			js := onJavaScriptBackend(test, program)
			native, _ := nativelyUncached(test, program)
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
					if got.exitCode != 70 ||
						string(got.stdout) != "called\n" ||
						!strings.HasPrefix(string(got.stderr), "adamic: panic: predicate ") ||
						!strings.Contains(string(got.stderr), path+":") ||
						!strings.Contains(string(got.stderr), " "+branch+" branch:") {
						test.Errorf(
							"%s qualified lie did not stop at the %s call check: exit %d stdout %q stderr %q",
							backend.name,
							branch,
							got.exitCode,
							got.stdout,
							got.stderr,
						)
					}
				} else if d := disagreement(node, got); d != "" {
					test.Error(backend.name + ": " + d)
				}
			}
			if d := disagreement(js, native); d != "" {
				test.Fatal("ASan/UBSan native: " + d)
			}
		})
	}
}

func TestQualifiedPredicateNamespaceAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "namespace")
}
func TestQualifiedPredicateGenericTypeAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "type")
}
func TestQualifiedPredicateArrayAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "each")
}
func TestQualifiedPredicateDefinedAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "defined")
}
func TestQualifiedPredicateReexportAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "reexport")
}
func TestQualifiedPredicateModuleAssertions(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "module")
}
func TestQualifiedPredicateNestedNamespace(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "nested")
}
func TestQualifiedPredicateFalseDirection(test *testing.T) {
	test.Parallel()
	qualifiedPredicateAgreement(test, "false_read")
}

func TestQualifiedPredicateVoidOverloadDoesNotCheck(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_qualified/void/main.ts"))
	if err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	if d := disagreement(run{stdout: []byte("called\nafter\n")}, node); d != "" {
		test.Fatal(d)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	js := onJavaScriptBackend(test, program)
	if d := disagreement(node, js); d != "" {
		test.Fatal(d)
	}
	native, _ := nativelyUncached(test, program)
	if d := disagreement(node, native); d != "" {
		test.Fatal(d)
	}
}

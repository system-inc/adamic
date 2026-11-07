package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func checkMiscompile2AStop(test *testing.T, name, output, diagnostic string, refused bool) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/miscompile_2a_refused", name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != output || len(node.stderr) != 0 {
		test.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	test.Logf("Node: exit=0 stdout=%q", node.stdout)
	// Both backends share lowering. A stop here prevents C, JavaScript and a
	// sanitizer build from receiving the unsafe program at all.
	_, loweredError := lowered(test, path)
	var pending *lower.NotYet
	var rejection *lower.Refused
	if loweredError == nil ||
		(!refused && !errors.As(loweredError, &pending)) ||
		(refused && !errors.As(loweredError, &rejection)) ||
		!strings.Contains(loweredError.Error(), diagnostic) ||
		!strings.Contains(loweredError.Error(), name+".a:") {
		test.Fatalf("want path-bearing stop %q, got %v", diagnostic, loweredError)
	}
	test.Log(loweredError)
}

func TestMiscompile2AErrorSpread(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(test, "9984394_error_spread", "no name no message\n", "object spread of Error storage", false)
}
func TestMiscompile2AOptionalError(test *testing.T) {
	test.Parallel()
	checkMiscompile2AAdmitted(test, "internal/oracle/testdata/miscompile_2a/4ddd17f_opt_3.a", "[] [m]\n")
}
func TestMiscompile2AMaybeSetter(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(test, "classfeat_maybe_setter", "3\n", "setter with an optional numeric input", false)
}
func TestMiscompile2ALongName(test *testing.T) {
	test.Parallel()
	checkMiscompile2AAdmitted(test, "internal/oracle/testdata/miscompile_2a/4ddd17f_long_name.a", "2\n")
}
func TestMiscompile2ADerivedSymbol(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(test, "iterators_derived_symbol", "1,2\n", "computed member name in a derived class", false)
}
func TestMiscompile2AOverrideSource(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(
		test,
		"iterators_override_source",
		"300\n3\n",
		"computed member name in a derived class",
		false,
	)
}
func TestMiscompile2AConstructorArrow(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(
		test,
		"d7054e9_this_arrow",
		"undefined:NaN|set\n",
		"this escaping a constructor through a closure",
		true,
	)
}
func TestMiscompile2AConstructorNumberArrow(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(
		test,
		"d7054e9_this_arrow_number",
		"NaN\nNaN,NaN\n",
		"this escaping a constructor through a closure",
		true,
	)
}

func TestMiscompile2AErrorSubclassSpread(test *testing.T) {
	test.Parallel()
	checkMiscompile2AStop(
		test,
		"error_subclass_spread",
		"no name no message\n",
		"object spread of Error storage",
		false,
	)
}

func checkMiscompile2AAdmitted(test *testing.T, relative, output string) {
	test.Helper()
	path := relative
	if !filepath.IsAbs(path) {
		path = filepath.Join(repository, relative)
	}
	path, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != output || len(node.stderr) != 0 {
		test.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	if got := onJavaScriptBackend(test, program); disagreement(node, got) != "" {
		test.Fatalf("JavaScript: %s", disagreement(node, got))
	}
	got, binary := natively(test, program)
	if difference := disagreement(node, got); difference != "" {
		test.Fatalf("%s; native stderr: %s", difference, got.stderr)
	}
	if got := released(test, program); disagreement(node, got) != "" {
		test.Fatalf("release native: %s", disagreement(node, got))
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestMiscompile2AControls(test *testing.T) {
	test.Parallel()
	checkMiscompile2AAdmitted(
		test,
		"internal/oracle/testdata/miscompile_2a/controls.a",
		"2\n2\n3\nplain shown\nm\n1,2\n[] [m]\nError bad\nundefined\n3\n",
	)
}
func TestMiscompile2AMarker(test *testing.T) {
	test.Parallel()
	checkMiscompile2AAdmitted(
		test,
		"internal/oracle/testdata/census_never_rest_marker.a",
		"function\nfunction\nok\nfunction:function\nfunction\nfunction\nfunction\nfunction\nfunction:1\nfunction\n",
	)
}

// Keep the stored census program unchanged, and exercise its existing optional
// entry point with an omitted argument using a separate test driver.
func TestMiscompile2AMarkerUndefined(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/census_never_rest_marker.a"))
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "marker-undefined.a")
	source = append(source, []byte(`
try { fail(); } catch (error) {
 if (error instanceof Error) { console.log('[' + error.message + ']'); }
}
`)...)
	if writeError := os.WriteFile(path, source, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	checkMiscompile2AAdmitted(
		test,
		path,
		"function\nfunction\nok\nfunction:function\nfunction\nfunction\nfunction\nfunction\nfunction:1\nfunction\n[]\n",
	)
}

func TestMiscompile2AUndefinedError(test *testing.T) {
	test.Parallel()
	checkMiscompile2AAdmitted(
		test,
		"internal/oracle/testdata/miscompile_2a/error_undefined_message.a",
		"[] []\n[] [m]\n[] 1\n",
	)
}

func init() {
	for _, name := range []string{
		"9984394_error_spread",
		"classfeat_maybe_setter",
		"iterators_derived_symbol",
		"iterators_override_source",
		"error_subclass_spread",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/miscompile_2a_refused/" + name + ".a", false, false})
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/miscompile_2a/4ddd17f_long_name.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/miscompile_2a/controls.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/miscompile_2a/4ddd17f_opt_3.a", true, false})

	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/miscompile_2a/error_undefined_message.a", true, false})

}

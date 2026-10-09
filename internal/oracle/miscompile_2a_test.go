package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func checkMiscompile2AStop(t *testing.T, name, output, diagnostic string, refused bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/miscompile_2a", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != output || len(node.stderr) != 0 {
		t.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	t.Logf("Node: exit=0 stdout=%q", node.stdout)
	// Both backends share lowering. A stop here prevents C, JavaScript and a
	// sanitizer build from receiving the unsafe program at all.
	_, err = lowered(t, path)
	var pending *lower.NotYet
	var rejection *lower.Refused
	if err == nil || (!refused && !errors.As(err, &pending)) || (refused && !errors.As(err, &rejection)) || !strings.Contains(err.Error(), diagnostic) || !strings.Contains(err.Error(), name+".a:") {
		t.Fatalf("want path-bearing stop %q, got %v", diagnostic, err)
	}
	t.Log(err)
}

func TestMiscompile2AErrorSpread(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "9984394_error_spread", "no name no message\n", "object spread of Error storage", false)
}
func TestMiscompile2AOptionalError(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "4ddd17f_opt_3", "[] [m]\n", "possibly undefined message", false)
}
func TestMiscompile2AMaybeSetter(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "classfeat_maybe_setter", "3\n", "setter with an optional numeric input", false)
}
func TestMiscompile2ALongName(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "4ddd17f_long_name", "2\n", "property name longer than 4095 bytes", false)
}
func TestMiscompile2ADerivedSymbol(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "iterators_derived_symbol", "1,2\n", "computed member name in a derived class", false)
}
func TestMiscompile2AOverrideSource(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "iterators_override_source", "300\n3\n", "computed member name in a derived class", false)
}
func TestMiscompile2AConstructorArrow(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "d7054e9_this_arrow", "undefined:NaN|set\n", "this escaping a constructor through a closure", true)
}
func TestMiscompile2AConstructorNumberArrow(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "d7054e9_this_arrow_number", "NaN\nNaN,NaN\n", "this escaping a constructor through a closure", true)
}

func TestMiscompile2AErrorSubclassSpread(t *testing.T) {
	t.Parallel()
	checkMiscompile2AStop(t, "error_subclass_spread", "no name no message\n", "object spread of Error storage", false)
}

func checkMiscompile2AAdmitted(t *testing.T, relative, output string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != output || len(node.stderr) != 0 {
		t.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if got := onJavaScriptBackend(t, program); disagreement(node, got) != "" {
		t.Fatalf("JavaScript: %s", disagreement(node, got))
	}
	got, binary := natively(t, program)
	if difference := disagreement(node, got); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestMiscompile2AControls(t *testing.T) {
	t.Parallel()
	checkMiscompile2AAdmitted(t, "internal/oracle/testdata/miscompile_2a/controls.a", "2\n2\n3\nplain shown\nm\n1,2\n[] [m]\nError bad\nundefined\n3\n")
}
func TestMiscompile2AMarker(t *testing.T) {
	t.Parallel()
	checkMiscompile2AAdmitted(t, "internal/oracle/testdata/census_never_rest_marker.a", "function\nfunction\nok\nfunction:function\nfunction\nfunction\nfunction\nfunction\nfunction:1\nfunction\n")
}

func init() {
	for _, name := range []string{"9984394_error_spread", "4ddd17f_opt_3", "classfeat_maybe_setter", "4ddd17f_long_name", "iterators_derived_symbol", "iterators_override_source", "error_subclass_spread"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/miscompile_2a/" + name + ".a", false, false})
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/miscompile_2a/controls.a", true, false})
}

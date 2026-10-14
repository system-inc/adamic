package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

const gettersCensusCallbackRefusal = "Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)"

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/getters_census_class_map.a", true, false},
		{"internal/oracle/testdata/getters_census_throwing_object.a", true, false},
		{"internal/oracle/testdata/getters_census_binary_inline.a", true, false},
		{"internal/oracle/testdata/getters_census_closure_number_methods.a", false, false},
		{"internal/oracle/testdata/getters_census_lazy_object.a", false, false},
		{"internal/oracle/testdata/getters_census_closure_number_next.a", false, false},
		{"internal/oracle/testdata/getters_census_closure_union_next.a", false, false},
		{"internal/oracle/testdata/getters_census_snapshot_next.a", false, false},
		{"internal/oracle/testdata/getters_census_callback_pair_methods.a", false, false},
		{"internal/oracle/testdata/getters_census_higher_callback_pair_methods.a", false, false},
	} {
		fixtures = append(fixtures, fixture)
	}
}

// Each observation is a top-level leaf, so the test budget applies to one shape.
// Refused witnesses stay outside the passing fixture list; the shared oracle
// only recognizes NotYet in its non-lowering list.
func observeGettersCensus(t *testing.T, shape, refusal, want string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/getters_census_"+shape+".a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	t.Logf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != want {
		t.Fatalf("source failed: %+v", node)
	}
	program, err := lowered(t, path)
	if refusal != "" {
		if err == nil || !strings.Contains(err.Error(), refusal) {
			t.Fatalf("want refusal %q, got %v", refusal, err)
		}
		var notYet *lower.NotYet
		var refused *lower.Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Fatalf("not a compiler refusal: %v", err)
		}
		t.Logf("Lower: %v; JavaScript/release/sanitized not reached", err)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	backend := onJavaScriptBackend(t, program)
	sanitized, binary := natively(t, program)
	release := released(t, program)
	for name, result := range map[string]run{"JavaScript": backend, "release": release, "sanitized": sanitized} {
		t.Logf("%s: exit %d stdout %q stderr %q", name, result.exitCode, result.stdout, result.stderr)
		if difference := disagreement(node, result); difference != "" {
			t.Errorf("%s: %s", name, difference)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestGettersCensusClosureNumberMethods(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "closure_number_methods", "stage 0 can't lower a non-accessor method in an accessor literal yet", "0\n1\n")
}

func TestGettersCensusLazyObject(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "lazy_object", "stage 0 can't lower reading factory yet", "constructed\ninitialize\n9\n10\n")
}

func TestGettersCensusPrimaryFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "primary_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n7\n7\ninitialize 7\n7\n")
}

func TestGettersCensusOptionalBooleanFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "optional_boolean_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}

func TestGettersCensusUpdateNodeFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "update_node_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusUnaryNodeFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "unary_node_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusOptionalTypeFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "optional_type_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusUpdateTypeFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "update_type_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n15\n15\ninitialize 7\n15\n")
}

func TestGettersCensusOptionalCommentFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "optional_comment_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}

func TestGettersCensusUpdateCommentFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "update_comment_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusBinaryFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "binary_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusOverloadedFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "overloaded_function", "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)", "10\n")
}

func TestGettersCensusUnaryFunction(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "unary_function", gettersCensusCallbackRefusal, "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusThrowingObject(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "throwing_object", "", "constructed\nNot supported\n")
}

func TestGettersCensusClosureNumberNext(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "closure_number_next", "stage 0 can't lower a non-accessor method in an accessor literal yet", "0\n1\n")
}

func TestGettersCensusClosureUnionNext(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "closure_union_next", "stage 0 can't lower a non-accessor method in an accessor literal yet", "undefined\nbad mapping\n")
}

func TestGettersCensusSnapshotNext(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "snapshot_next", "stage 0 can't lower a non-accessor method in an accessor literal yet", "0/1\n")
}

func TestGettersCensusClassMap(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "class_map", "", "0/0/0\n1/1/1\n9\n")
}

func TestGettersCensusCallbackPairMethods(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "callback_pair_methods", "stage 0 can't lower a non-accessor method in an accessor literal yet", "3\n13\ndone\ninitialized\n")
}

func TestGettersCensusHigherCallbackPairMethods(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "higher_callback_pair_methods", "stage 0 can't lower a non-accessor method in an accessor literal yet", "3\n13\ndone\ninitialized\n")
}

func TestGettersCensusBinaryInline(t *testing.T) {
	t.Parallel()
	observeGettersCensus(t, "binary_inline", "", "constructed\ninitialize 7\n12\n16\ninitialize 7\n10\n")
}

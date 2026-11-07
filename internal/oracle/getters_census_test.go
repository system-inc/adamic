package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

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
func observeGettersCensus(test *testing.T, shape, refusal, want string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/getters_census_"+shape+".a"))
	if err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	test.Logf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != want {
		test.Fatalf("source failed: %+v", node)
	}
	program, err := lowered(test, path)
	if refusal != "" {
		if err == nil || !strings.Contains(err.Error(), refusal) {
			test.Fatalf("want refusal %q, got %v", refusal, err)
		}
		var notYet *lower.NotYet
		var refused *lower.Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			test.Fatalf("not a compiler refusal: %v", err)
		}
		test.Logf("Lower: %v; JavaScript/release/sanitized not reached", err)
		return
	}
	if err != nil {
		test.Fatal(err)
	}
	// These formerly cycle-refused callback shapes now require graph ownership.
	// The overloaded shape returns above with its independent cast refusal.
	if strings.HasSuffix(shape, "_function") && len(program.GraphTypes) == 0 {
		test.Fatal("callback admitted without graph ownership")
	}
	backend := onJavaScriptBackend(test, program)
	sanitized, binary := natively(test, program)
	release := released(test, program)
	for name, result := range map[string]run{"JavaScript": backend, "release": release, "sanitized": sanitized} {
		test.Logf("%s: exit %d stdout %q stderr %q", name, result.exitCode, result.stdout, result.stderr)
		if difference := disagreement(node, result); difference != "" {
			test.Errorf("%s: %s", name, difference)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestGettersCensusClosureNumberMethods(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"closure_number_methods",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"0\n1\n",
	)
}

func TestGettersCensusLazyObject(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"lazy_object",
		"stage 0 can't lower reading factory yet",
		"constructed\ninitialize\n9\n10\n",
	)
}

func TestGettersCensusPrimaryFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "primary_function", "", "constructed\ninitialize 7\n7\n7\ninitialize 7\n7\n")
}

func TestGettersCensusOptionalBooleanFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "optional_boolean_function", "", "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}

func TestGettersCensusUpdateNodeFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "update_node_function", "", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusUnaryNodeFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "unary_node_function", "", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusOptionalTypeFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "optional_type_function", "", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusUpdateTypeFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "update_type_function", "", "constructed\ninitialize 7\n15\n15\ninitialize 7\n15\n")
}

func TestGettersCensusOptionalCommentFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "optional_comment_function", "", "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}

func TestGettersCensusUpdateCommentFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "update_comment_function", "", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusBinaryFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "binary_function", "", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}

func TestGettersCensusOverloadedFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"overloaded_function",
		"Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)",
		"10\n",
	)
}

func TestGettersCensusUnaryFunction(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "unary_function", "", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func TestGettersCensusThrowingObject(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "throwing_object", "", "constructed\nNot supported\n")
}

func TestGettersCensusClosureNumberNext(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"closure_number_next",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"0\n1\n",
	)
}

func TestGettersCensusClosureUnionNext(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"closure_union_next",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"undefined\nbad mapping\n",
	)
}

func TestGettersCensusSnapshotNext(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"snapshot_next",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"0/1\n",
	)
}

func TestGettersCensusClassMap(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "class_map", "", "0/0/0\n1/1/1\n9\n")
}

func TestGettersCensusCallbackPairMethods(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"callback_pair_methods",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"3\n13\ndone\ninitialized\n",
	)
}

func TestGettersCensusHigherCallbackPairMethods(test *testing.T) {
	test.Parallel()
	observeGettersCensus(
		test,
		"higher_callback_pair_methods",
		"stage 0 can't lower a non-accessor method in an accessor literal yet",
		"3\n13\ndone\ninitialized\n",
	)
}

func TestGettersCensusBinaryInline(test *testing.T) {
	test.Parallel()
	observeGettersCensus(test, "binary_inline", "", "constructed\ninitialize 7\n12\n16\ninitialize 7\n10\n")
}

package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"binary", "unary", "primary", "unary_node", "update_node", "optional_boolean", "optional_type", "update_type", "optional_comment", "update_comment"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/getters_census_" + name + "_function.a", true, false})
	}
}
func observeClosureTargetsCensus(t *testing.T, shape, want string) {
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

func TestGettersCensusPrimaryFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "primary_function", "constructed\ninitialize 7\n7\n7\ninitialize 7\n7\n")
}
func TestGettersCensusOptionalBooleanFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "optional_boolean_function", "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}
func TestGettersCensusUpdateNodeFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "update_node_function", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}
func TestGettersCensusUnaryNodeFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "unary_node_function", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}
func TestGettersCensusOptionalTypeFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "optional_type_function", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}
func TestGettersCensusUpdateTypeFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "update_type_function", "constructed\ninitialize 7\n15\n15\ninitialize 7\n15\n")
}
func TestGettersCensusOptionalCommentFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "optional_comment_function", "constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n")
}
func TestGettersCensusUpdateCommentFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "update_comment_function", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}
func TestGettersCensusBinaryFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "binary_function", "constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n")
}
func TestGettersCensusUnaryFunction(t *testing.T) {
	t.Parallel()
	observeClosureTargetsCensus(t, "unary_function", "constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n")
}

func observeClosureCycleRefusal(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycle_closure_targets_refused", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "stored\n" {
		t.Fatalf("Node: %+v", node)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("want cycle refusal, got %v", err)
	}
	want := "'memo', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)"
	if name == "map" {
		want = "Map<number, () => number>, a map whose values can reach back to a map like it: a cycle reference counting can't free, and the write at <fixture>:14:1 may close one (the value written reaches something this function didn't make or let escape, and what it's written into wasn't made here, where make is called from the top level); declare the values weak, Map<number, Weak<() => number>> (import type { Weak } from 'adamic'), or make it a ReadonlyMap; or set in such a map only values this function made, or only in one it made (adamic/cycle-capable)"
	}
	got := strings.ReplaceAll(refused.What+"; "+refused.Fix, path, "<fixture>")
	if got != want {
		t.Fatalf("refusal changed:\nwant %s\ngot  %s", want, got)
	}
	t.Log(err)
}
func TestClosureCycleDirect(t *testing.T) { t.Parallel(); observeClosureCycleRefusal(t, "direct") }
func TestClosureCycleJoined(t *testing.T) { t.Parallel(); observeClosureCycleRefusal(t, "joined") }
func TestClosureCycleMap(t *testing.T)    { t.Parallel(); observeClosureCycleRefusal(t, "map") }

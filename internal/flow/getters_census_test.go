package flow

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
)

// These source-Node witnesses deliberately stop before IR exists. The oracle's
// TestGettersCensus leaves hold each stop; the flow corpus needs runnable IR.
var gettersCensusStops = map[string]string{
	"closure_number_methods":       "a non-accessor method in an accessor literal",
	"closure_number_next":          "a non-accessor method in an accessor literal",
	"closure_union_next":           "a non-accessor method in an accessor literal",
	"snapshot_next":                "a non-accessor method in an accessor literal",
	"callback_pair_methods":        "a non-accessor method in an accessor literal",
	"higher_callback_pair_methods": "a non-accessor method in an accessor literal",
	"lazy_object":                  "reading factory",
	"primary_function":             "adamic/cycle-capable",
	"optional_boolean_function":    "adamic/cycle-capable",
	"update_node_function":         "adamic/cycle-capable",
	"unary_node_function":          "adamic/cycle-capable",
	"optional_type_function":       "adamic/cycle-capable",
	"update_type_function":         "adamic/cycle-capable",
	"optional_comment_function":    "adamic/cycle-capable",
	"update_comment_function":      "adamic/cycle-capable",
	"binary_function":              "adamic/cycle-capable",
	"unary_function":               "adamic/cycle-capable",
	"overloaded_function":          "adamic/no-unchecked-cast",
}

func refusedGettersCensusFixture(path string) bool {
	name := filepath.Base(path)
	if !strings.HasPrefix(name, "getters_census_") || !strings.HasSuffix(name, ".a") {
		return false
	}
	shape := strings.TrimSuffix(strings.TrimPrefix(name, "getters_census_"), ".a")
	_, stopped := gettersCensusStops[shape]
	return stopped
}

func TestFlowGettersCensusRefusalClassification(t *testing.T) {
	t.Parallel()
	runnable := map[string]bool{}
	for _, path := range programs(t) {
		runnable[filepath.Base(path)] = true
	}
	for shape, reason := range gettersCensusStops {
		name := "getters_census_" + shape + ".a"
		if runnable[name] {
			t.Errorf("deliberate refusal %s entered the runnable flow corpus", name)
		}
		path, err := filepath.Abs(filepath.Join("../oracle/testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		checked, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		program, err := Lower(context.Background(), checked)
		if program != nil || err == nil || !strings.Contains(err.Error(), path+":") || !strings.Contains(err.Error(), reason) {
			t.Errorf("excluded %s must still stop at its named path for %q: %v", name, reason, err)
		}
	}
	for _, shape := range []string{"class_map", "binary_inline", "throwing_object"} {
		if name := "getters_census_" + shape + ".a"; !runnable[name] {
			t.Errorf("runnable getter %s left the flow corpus", name)
		}
	}
}

func TestFlowGettersCensusClassMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/getters_census_class_map.a")
}

func TestFlowGettersCensusBinaryInline(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/getters_census_binary_inline.a")
}

func TestFlowGettersCensusThrowingObject(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/getters_census_throwing_object.a")
}

func TestFlowTraceSnapshotsDoNotInvokeGetters(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	source := traceRuntime + `
let getterCalls = 0;
const holder = {
	get value() { getterCalls++; throw new Error('snapshot invoked value getter'); },
	get code() { getterCalls++; throw new Error('snapshot invoked code getter'); },
	child: { count: 1 },
};
adamicPoint(0, [[0, () => holder]]);
holder.child.count = 2;
adamicPoint(1, [[0, () => holder]]);
if (getterCalls !== 0) throw new Error('snapshot executed user code');
if (!adamicTrace.includes('mutated 0 0')) throw new Error('snapshot lost nested data mutation');
`
	if output, err := exec.CommandContext(ctx, "node", "--input-type=module", "-e", source).CombinedOutput(); err != nil {
		t.Fatalf("snapshot changed getter behavior or lost a data mutation: %v\n%s", err, output)
	}
}

package oracle

import (
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range graphOracleFixtures {
		if path == "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a" {
			continue
		}
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

var graphOracleFixtures = []string{
	"internal/oracle/testdata/canonical_graph.a",
	"internal/oracle/testdata/canonical_graph_counted.a",
	"internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a",
	"internal/oracle/testdata/graph_regions/classification_conditional.a",
	"internal/oracle/testdata/graph_regions/classification_mixed.a",
	"internal/oracle/testdata/graph_regions/classification_override.a",
	"internal/oracle/testdata/graph_regions/classification_return.a",
	"internal/oracle/testdata/graph_regions_accessor.a",
	"internal/oracle/testdata/graph_regions_anchor.a",
	"internal/oracle/testdata/graph_regions_cache.a",
	"internal/oracle/testdata/graph_regions_closure.a",
	"internal/oracle/testdata/graph_regions_counted_container.a",
	"internal/oracle/testdata/graph_regions_coverage_arrays.a",
	"internal/oracle/testdata/graph_regions_coverage_boundary.a",
	"internal/oracle/testdata/graph_regions_coverage_closure_loop.a",
	"internal/oracle/testdata/graph_regions_coverage_iterator.a",
	"internal/oracle/testdata/graph_regions_coverage_maps.a",
	"internal/oracle/testdata/graph_regions_coverage_merge.a",
	"internal/oracle/testdata/graph_regions_coverage_payload_loop.a",
	"internal/oracle/testdata/graph_regions_coverage_readonly.a",
	"internal/oracle/testdata/graph_regions_coverage_spread.a",
	"internal/oracle/testdata/graph_regions_coverage_throw.a",
	"internal/oracle/testdata/graph_regions_entries.a",
	"internal/oracle/testdata/graph_regions_escape.a",
	"internal/oracle/testdata/graph_regions_flow.a",
	"internal/oracle/testdata/graph_regions_generic_capture.a",
	"internal/oracle/testdata/graph_regions_list.a",
	"internal/oracle/testdata/graph_regions_literal_method.a",
	"internal/oracle/testdata/graph_regions_literals.a",
	"internal/oracle/testdata/graph_regions_nested_disjoint.a",
	"internal/oracle/testdata/graph_regions_nested_self.a",
	"internal/oracle/testdata/graph_regions_parse.a",
	"internal/oracle/testdata/graph_regions_regression_01.a",
	"internal/oracle/testdata/graph_regions_regression_02.a",
	"internal/oracle/testdata/graph_regions_regression_03.a",
	"internal/oracle/testdata/graph_regions_regression_04.a",
	"internal/oracle/testdata/graph_regions_regression_05.a",
	"internal/oracle/testdata/graph_regions_regression_06.a",
	"internal/oracle/testdata/graph_regions_regression_07.a",
	"internal/oracle/testdata/graph_regions_regression_08.a",
	"internal/oracle/testdata/graph_regions_regression_09.a",
	"internal/oracle/testdata/graph_regions_regression_10.a",
	"internal/oracle/testdata/graph_regions_regression_11.a",
	"internal/oracle/testdata/graph_regions_regression_12.a",
	"internal/oracle/testdata/graph_regions_regression_13.a",
	"internal/oracle/testdata/graph_regions_regression_14.a",
	"internal/oracle/testdata/graph_regions_regression_15.a",
	"internal/oracle/testdata/graph_regions_static.a",
	"internal/oracle/testdata/graph_regions_static_interface.a",
	"internal/oracle/testdata/graph_regions_static_parent.a",
	"internal/oracle/testdata/graph_regions_static_private.a",
	"internal/oracle/testdata/graph_regions_structural_literal.a",
	"internal/oracle/testdata/graph_regions_symbols.a",
	"internal/oracle/testdata/graph_regions_throw.a",
	"internal/oracle/testdata/graph_regions_weak_mixed.a",
	"internal/oracle/testdata/weak_region_review.a",
}

// One complete witness per test keeps both sanitizer and counted work bounded.
func graphOracleWitness(t *testing.T, path string) {
	t.Helper()
	graphOracleWitnessWithLayout(t, path, false)
}

func graphOracleWitnessWithLayout(t *testing.T, path string, million bool) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, absolute)
	if truth.exitCode != 0 {
		t.Fatalf("Node: %d %s", truth.exitCode, truth.stderr)
	}
	program, err := lowered(t, absolute)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := natively(t, program)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	code := native.C(program)
	countedBinary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(code, countedBinary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, arguments := pinnedStack(countedBinary)
	var countedResult run
	if emitsParallelMap(code) {
		countedResult = executeParallel(t, "1", false, name, arguments...)
	} else {
		countedResult = execute(t, name, arguments...)
	}
	if countedResult.exitCode != truth.exitCode || string(countedResult.stdout) != string(truth.stdout) {
		t.Fatalf("counted disagrees with Node: %d %s %s", countedResult.exitCode, countedResult.stdout, countedResult.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(countedResult)); report != "" {
		t.Fatal(report)
	}
	graphOracleCounts(t, path)
	graphMeasureFixture(t, path, million)
}

func graphOracleCounts(t *testing.T, path string) {
	t.Helper()
	_, measure := counted(t, path, false, nil, false, false)
	row := measure()
	if *updateCounts {
		measuredCounts.Store(path, row)
	} else {
		rows, err := recordedCounts()
		if err != nil {
			t.Fatal(err)
		}
		if row != rows[path] {
			t.Fatalf("counts changed: got %s; recorded %s", row, rows[path])
		}
	}
	t.Log(row)
}

func TestGraphOracleCanonicalGraph(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/canonical_graph.a")
}

func TestGraphOracleCanonicalGraphCounted(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/canonical_graph_counted.a")
}

func TestGraphOracleGraphRegionsPerTask(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a")
}

func TestGraphOracleGraphRegionsClassificationConditional(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions/classification_conditional.a")
}

func TestGraphOracleGraphRegionsClassificationMixed(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions/classification_mixed.a")
}

func TestGraphOracleGraphRegionsClassificationOverride(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions/classification_override.a")
}

func TestGraphOracleGraphRegionsClassificationReturn(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions/classification_return.a")
}

func TestGraphOracleGraphRegionsAccessor(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_accessor.a")
}

func TestGraphOracleGraphRegionsAnchor(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_anchor.a")
}

func TestGraphOracleGraphRegionsCache(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_cache.a")
}

func TestGraphOracleGraphRegionsClosure(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_closure.a")
}

func TestGraphOracleGraphRegionsCountedContainer(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_counted_container.a")
}

func TestGraphOracleGraphRegionsCoverageArrays(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_arrays.a")
}

func TestGraphOracleGraphRegionsCoverageBoundary(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_boundary.a")
}

func TestGraphOracleGraphRegionsCoverageClosureLoop(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_closure_loop.a")
}

func TestGraphOracleGraphRegionsCoverageIterator(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_iterator.a")
}

func TestGraphOracleGraphRegionsCoverageMaps(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_maps.a")
}

func TestGraphOracleGraphRegionsCoverageMerge(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_merge.a")
}

func TestGraphOracleGraphRegionsCoveragePayloadLoop(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_payload_loop.a")
}

func TestGraphOracleGraphRegionsCoverageReadonly(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_readonly.a")
}

func TestGraphOracleGraphRegionsCoverageSpread(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_spread.a")
}

func TestGraphOracleGraphRegionsCoverageThrow(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_coverage_throw.a")
}

func TestGraphOracleGraphRegionsEntries(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_entries.a")
}

func TestGraphOracleGraphRegionsEscape(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_escape.a")
}

func TestGraphOracleGraphRegionsFlow(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_flow.a")
}

func TestGraphOracleGraphRegionsGenericCapture(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_generic_capture.a")
}

func TestGraphOracleGraphRegionsList(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_list.a")
}

func TestGraphOracleGraphRegionsLiteralMethod(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_literal_method.a")
}

func TestGraphOracleGraphRegionsLiterals(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_literals.a")
}

func TestGraphOracleGraphRegionsNestedDisjoint(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_nested_disjoint.a")
}

func TestGraphOracleGraphRegionsNestedSelf(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_nested_self.a")
}

func TestGraphOracleGraphRegionsParse(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_parse.a")
}

func TestGraphOracleGraphRegionsRegression01(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_01.a")
}

func TestGraphOracleGraphRegionsRegression02(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_02.a")
}

func TestGraphOracleGraphRegionsRegression03(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_03.a")
}

func TestGraphOracleGraphRegionsRegression04(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_04.a")
}

func TestGraphOracleGraphRegionsRegression05(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_05.a")
}

func TestGraphOracleGraphRegionsRegression06(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_06.a")
}

func TestGraphOracleGraphRegionsRegression07(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_07.a")
}

func TestGraphOracleGraphRegionsRegression08(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_08.a")
}

func TestGraphOracleGraphRegionsRegression09(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_09.a")
}

func TestGraphOracleGraphRegionsRegression10(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_10.a")
}

func TestGraphOracleGraphRegionsRegression11(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_11.a")
}

func TestGraphOracleGraphRegionsRegression12(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_12.a")
}

func TestGraphOracleGraphRegionsRegression13(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_13.a")
}

func TestGraphOracleGraphRegionsRegression14(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_14.a")
}

func TestGraphOracleGraphRegionsRegression15(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_regression_15.a")
}

func TestGraphOracleGraphRegionsStatic(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_static.a")
}

func TestGraphOracleGraphRegionsStaticInterface(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_static_interface.a")
}

func TestGraphOracleGraphRegionsStaticParent(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_static_parent.a")
}

func TestGraphOracleGraphRegionsStaticPrivate(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_static_private.a")
}

func TestGraphOracleGraphRegionsStructuralLiteral(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_structural_literal.a")
}

func TestGraphOracleGraphRegionsSymbols(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_symbols.a")
}

func TestGraphOracleGraphRegionsThrow(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_throw.a")
}

func TestGraphOracleGraphRegionsWeakMixed(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/graph_regions_weak_mixed.a")
}

func TestGraphOracleWeakRegionReview(t *testing.T) {
	t.Parallel()
	graphOracleWitness(t, "internal/oracle/testdata/weak_region_review.a")
}

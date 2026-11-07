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
func graphOracleWitness(test *testing.T, path string) {
	test.Helper()
	graphOracleWitnessWithLayout(test, path, false)
}

func graphOracleWitnessWithLayout(test *testing.T, path string, million bool) {
	test.Helper()
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, absolute)
	if truth.exitCode != 0 {
		test.Fatalf("Node: %d %s", truth.exitCode, truth.stderr)
	}
	program, loweredError := lowered(test, absolute)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	actual, binary := natively(test, program)
	if difference := disagreement(truth, actual); difference != "" {
		test.Fatal(difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "" {
		test.Fatal(difference)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	code := native.C(program)
	countedBinary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(code, countedBinary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	name, arguments := pinnedStack(countedBinary)
	var countedResult run
	if emitsParallelMap(code) {
		countedResult = executeParallel(test, "1", false, name, arguments...)
	} else {
		countedResult = execute(test, name, arguments...)
	}
	if countedResult.exitCode != truth.exitCode || string(countedResult.stdout) != string(truth.stdout) {
		test.Fatalf(
			"counted disagrees with Node: %d %s %s",
			countedResult.exitCode,
			countedResult.stdout,
			countedResult.stderr,
		)
	}
	if report := leakcheck.Unbalanced(leakRun(countedResult)); report != "" {
		test.Fatal(report)
	}
	graphOracleCounts(test, path)
	graphMeasureFixture(test, path, million)
}

func graphOracleCounts(test *testing.T, path string) {
	test.Helper()
	_, measure := counted(test, path, false, nil, false, false)
	row := measure()
	if *updateCounts {
		measuredCounts.Store(path, row)
	} else {
		rows, recordedCountsError := recordedCounts()
		if recordedCountsError != nil {
			test.Fatal(recordedCountsError)
		}
		if row != rows[path] {
			test.Fatalf("counts changed: got %s; recorded %s", row, rows[path])
		}
	}
	test.Log(row)
}

func TestGraphOracleCanonicalGraph(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/canonical_graph.a")
}

func TestGraphOracleCanonicalGraphCounted(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/canonical_graph_counted.a")
}

func TestGraphOracleGraphRegionsPerTask(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a")
}

func TestGraphOracleGraphRegionsClassificationConditional(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions/classification_conditional.a")
}

func TestGraphOracleGraphRegionsClassificationMixed(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions/classification_mixed.a")
}

func TestGraphOracleGraphRegionsClassificationOverride(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions/classification_override.a")
}

func TestGraphOracleGraphRegionsClassificationReturn(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions/classification_return.a")
}

func TestGraphOracleGraphRegionsAccessor(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_accessor.a")
}

func TestGraphOracleGraphRegionsAnchor(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_anchor.a")
}

func TestGraphOracleGraphRegionsCache(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_cache.a")
}

func TestGraphOracleGraphRegionsClosure(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_closure.a")
}

func TestGraphOracleGraphRegionsCountedContainer(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_counted_container.a")
}

func TestGraphOracleGraphRegionsCoverageArrays(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_arrays.a")
}

func TestGraphOracleGraphRegionsCoverageBoundary(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_boundary.a")
}

func TestGraphOracleGraphRegionsCoverageClosureLoop(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_closure_loop.a")
}

func TestGraphOracleGraphRegionsCoverageIterator(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_iterator.a")
}

func TestGraphOracleGraphRegionsCoverageMaps(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_maps.a")
}

func TestGraphOracleGraphRegionsCoverageMerge(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_merge.a")
}

func TestGraphOracleGraphRegionsCoveragePayloadLoop(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_payload_loop.a")
}

func TestGraphOracleGraphRegionsCoverageReadonly(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_readonly.a")
}

func TestGraphOracleGraphRegionsCoverageSpread(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_spread.a")
}

func TestGraphOracleGraphRegionsCoverageThrow(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_coverage_throw.a")
}

func TestGraphOracleGraphRegionsEntries(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_entries.a")
}

func TestGraphOracleGraphRegionsEscape(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_escape.a")
}

func TestGraphOracleGraphRegionsFlow(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_flow.a")
}

func TestGraphOracleGraphRegionsGenericCapture(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_generic_capture.a")
}

func TestGraphOracleGraphRegionsList(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_list.a")
}

func TestGraphOracleGraphRegionsLiteralMethod(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_literal_method.a")
}

func TestGraphOracleGraphRegionsLiterals(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_literals.a")
}

func TestGraphOracleGraphRegionsNestedDisjoint(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_nested_disjoint.a")
}

func TestGraphOracleGraphRegionsNestedSelf(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_nested_self.a")
}

func TestGraphOracleGraphRegionsParse(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_parse.a")
}

func TestGraphOracleGraphRegionsRegression01(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_01.a")
}

func TestGraphOracleGraphRegionsRegression02(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_02.a")
}

func TestGraphOracleGraphRegionsRegression03(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_03.a")
}

func TestGraphOracleGraphRegionsRegression04(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_04.a")
}

func TestGraphOracleGraphRegionsRegression05(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_05.a")
}

func TestGraphOracleGraphRegionsRegression06(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_06.a")
}

func TestGraphOracleGraphRegionsRegression07(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_07.a")
}

func TestGraphOracleGraphRegionsRegression08(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_08.a")
}

func TestGraphOracleGraphRegionsRegression09(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_09.a")
}

func TestGraphOracleGraphRegionsRegression10(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_10.a")
}

func TestGraphOracleGraphRegionsRegression11(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_11.a")
}

func TestGraphOracleGraphRegionsRegression12(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_12.a")
}

func TestGraphOracleGraphRegionsRegression13(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_13.a")
}

func TestGraphOracleGraphRegionsRegression14(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_14.a")
}

func TestGraphOracleGraphRegionsRegression15(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_regression_15.a")
}

func TestGraphOracleGraphRegionsStatic(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_static.a")
}

func TestGraphOracleGraphRegionsStaticInterface(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_static_interface.a")
}

func TestGraphOracleGraphRegionsStaticParent(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_static_parent.a")
}

func TestGraphOracleGraphRegionsStaticPrivate(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_static_private.a")
}

func TestGraphOracleGraphRegionsStructuralLiteral(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_structural_literal.a")
}

func TestGraphOracleGraphRegionsSymbols(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_symbols.a")
}

func TestGraphOracleGraphRegionsThrow(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_throw.a")
}

func TestGraphOracleGraphRegionsWeakMixed(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/graph_regions_weak_mixed.a")
}

func TestGraphOracleWeakRegionReview(test *testing.T) {
	test.Parallel()
	graphOracleWitness(test, "internal/oracle/testdata/weak_region_review.a")
}

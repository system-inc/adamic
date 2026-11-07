package flow

import (
	"path/filepath"
	"strings"
	"testing"
)

// Async unit 2 exposes these existing fixtures to all four ordinary flow checks.
// Each program has its own bounded top-level unit instead of growing the remainder.
var asyncFlowUnits = []struct {
	path string
	test func(*testing.T)
}{
	{"../oracle/testdata/async_catch_return.a", TestFlowAsyncCatchReturn},
	{"../oracle/testdata/async_catches.a", TestFlowAsyncCatches},
	{"../oracle/testdata/async_closures.a", TestFlowAsyncClosures},
	{"../oracle/testdata/async_conditions.a", TestFlowAsyncConditions},
	{"../oracle/testdata/async_control.a", TestFlowAsyncControl},
	{"../oracle/testdata/async_escaped_capture.a", TestFlowAsyncEscapedCapture},
	{"../oracle/testdata/async_frame_capture_safe.a", TestFlowAsyncFrameCaptureSafe},
	{"../oracle/testdata/async_fresh_holder.a", TestFlowAsyncFreshHolder},
	{"../oracle/testdata/async_live_values.a", TestFlowAsyncLiveValues},
	{"../oracle/testdata/async_methods.a", TestFlowAsyncMethods},
	{"../oracle/testdata/async_mutation.a", TestFlowAsyncMutation},
	{"../oracle/testdata/async_operand_order.a", TestFlowAsyncOperandOrder},
	{"../oracle/testdata/async_promise_expression.a", TestFlowAsyncPromiseExpression},
	{"../oracle/testdata/async_receiver.a", TestFlowAsyncReceiver},
	{"../oracle/testdata/async_snapshots.a", TestFlowAsyncSnapshots},
	{"../oracle/testdata/async_void_value.a", TestFlowAsyncVoidValue},
	{"../oracle/testdata/async_coverage_discard.a", TestFlowAsyncCoverageDiscard},
	{"../oracle/testdata/async_coverage_parameters.a", TestFlowAsyncCoverageParameters},
	{"../oracle/testdata/async_coverage_reject_eager.a", TestFlowAsyncCoverageRejectEager},
	{"../oracle/testdata/async_coverage_reject_empty.a", TestFlowAsyncCoverageRejectEmpty},
	{"../oracle/testdata/async_coverage_reject_nested.a", TestFlowAsyncCoverageRejectNested},
	{"../oracle/testdata/async_coverage_typeof.a", TestFlowAsyncCoverageTypeof},
	{"../oracle/testdata/async_coverage_unions.a", TestFlowAsyncCoverageUnions},
	{"../oracle/testdata/async_coverage_values.a", TestFlowAsyncCoverageValues},
	{"../oracle/testdata/async_nested.a", TestFlowAsyncNested},
	{"../oracle/testdata/async_plain.a", TestFlowAsyncPlain},
	{"../oracle/testdata/async_three.a", TestFlowAsyncThree},
	{"../oracle/testdata/async_throw.a", TestFlowAsyncThrow},
	{"../oracle/testdata/async_typeof.a", TestFlowAsyncTypeof},
}

func TestFlowAsyncCoverageDiscard(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_discard.a")
}

func TestFlowAsyncCoverageParameters(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_parameters.a")
}

func TestFlowAsyncCoverageRejectEager(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_reject_eager.a")
}

func TestFlowAsyncCoverageRejectEmpty(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_reject_empty.a")
}

func TestFlowAsyncCoverageRejectNested(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_reject_nested.a")
}

func TestFlowAsyncCoverageTypeof(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_typeof.a")
}

func TestFlowAsyncCoverageUnions(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_unions.a")
}

func TestFlowAsyncCoverageValues(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_coverage_values.a")
}

func TestFlowAsyncNested(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_nested.a")
}

func TestFlowAsyncPlain(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_plain.a")
}

func TestFlowAsyncThree(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_three.a")
}

func TestFlowAsyncThrow(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_throw.a")
}

func TestFlowAsyncTypeof(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_typeof.a")
}

func TestAsyncFlowUnitsCoverExistingFixtures(test *testing.T) {
	test.Parallel()
	seen := map[string]bool{}
	for _, unit := range asyncFlowUnits {
		if seen[unit.path] {
			test.Fatalf("duplicate async flow unit: %s", unit.path)
		}
		seen[unit.path] = true
		if unit.test == nil {
			test.Fatalf("missing async test: %s", unit.path)
		}
	}
	for _, path := range programs(test) {
		if strings.HasPrefix(filepath.Base(path), "async_") && !seen[filepath.ToSlash(path)] {
			test.Fatalf("async fixture has no bounded flow unit: %s", path)
		}
	}
}

package flow

import "testing"

func init() {
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_completion.a", TestFlowAsyncLabelsAsyncFinallyCompletion})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_labels.a", TestFlowAsyncLabelsAsyncLabels})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_labels_finally.a", TestFlowAsyncLabelsAsyncLabelsFinally})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_labels_block_finally.a", TestFlowAsyncLabelsAsyncLabelsBlockFinally})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_edges.a", TestFlowAsyncLabelsLabelBlockEdges})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterables_cells_labels.a", TestFlowAsyncLabelsAsyncIterablesCellsLabels})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_break_sync.a", TestFlowAsyncLabelsLabelBlockBreakSync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_continue_sync.a", TestFlowAsyncLabelsLabelBlockContinueSync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_switch_break_sync.a", TestFlowAsyncLabelsLabelBlockSwitchBreakSync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_outer_continue_sync.a", TestFlowAsyncLabelsLabelBlockOuterContinueSync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_break_async.a", TestFlowAsyncLabelsLabelBlockBreakAsync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_continue_async.a", TestFlowAsyncLabelsLabelBlockContinueAsync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_switch_break_async.a", TestFlowAsyncLabelsLabelBlockSwitchBreakAsync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/label_block_outer_continue_async.a", TestFlowAsyncLabelsLabelBlockOuterContinueAsync})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_return.a", TestFlowAsyncLabelsAsyncFinallyReturn})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_throw.a", TestFlowAsyncLabelsAsyncFinallyThrow})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_break.a", TestFlowAsyncLabelsAsyncFinallyBreak})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_continue.a", TestFlowAsyncLabelsAsyncFinallyContinue})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_void_return.a", TestFlowAsyncLabelsAsyncFinallyVoidReturn})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_finally_snapshot.a", TestFlowAsyncLabelsAsyncFinallySnapshot})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_gap_label_admission.a", TestFlowAsyncLabelsAsyncGapLabelAdmission})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_gap_catch_admission.a", TestFlowAsyncLabelsAsyncGapCatchAdmission})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_gap_finally_admission.a", TestFlowAsyncLabelsAsyncGapFinallyAdmission})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_gap_finally_return_admission.a", TestFlowAsyncLabelsAsyncGapFinallyReturnAdmission})
}
func TestFlowAsyncLabelsAsyncFinallyCompletion(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_completion.a")
}
func TestFlowAsyncLabelsAsyncLabels(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_labels.a")
}
func TestFlowAsyncLabelsAsyncLabelsFinally(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_labels_finally.a")
}
func TestFlowAsyncLabelsAsyncLabelsBlockFinally(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_labels_block_finally.a")
}
func TestFlowAsyncLabelsLabelBlockEdges(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_edges.a")
}
func TestFlowAsyncLabelsAsyncIterablesCellsLabels(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_iterables_cells_labels.a")
}
func TestFlowAsyncLabelsLabelBlockBreakSync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_break_sync.a")
}
func TestFlowAsyncLabelsLabelBlockContinueSync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_continue_sync.a")
}
func TestFlowAsyncLabelsLabelBlockSwitchBreakSync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_switch_break_sync.a")
}
func TestFlowAsyncLabelsLabelBlockOuterContinueSync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_outer_continue_sync.a")
}
func TestFlowAsyncLabelsLabelBlockBreakAsync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_break_async.a")
}
func TestFlowAsyncLabelsLabelBlockContinueAsync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_continue_async.a")
}
func TestFlowAsyncLabelsLabelBlockSwitchBreakAsync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_switch_break_async.a")
}
func TestFlowAsyncLabelsLabelBlockOuterContinueAsync(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/label_block_outer_continue_async.a")
}
func TestFlowAsyncLabelsAsyncFinallyReturn(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_return.a")
}
func TestFlowAsyncLabelsAsyncFinallyThrow(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_throw.a")
}
func TestFlowAsyncLabelsAsyncFinallyBreak(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_break.a")
}
func TestFlowAsyncLabelsAsyncFinallyContinue(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_continue.a")
}
func TestFlowAsyncLabelsAsyncFinallyVoidReturn(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_void_return.a")
}
func TestFlowAsyncLabelsAsyncFinallySnapshot(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_finally_snapshot.a")
}
func TestFlowAsyncLabelsAsyncGapLabelAdmission(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_gap_label_admission.a")
}
func TestFlowAsyncLabelsAsyncGapCatchAdmission(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_gap_catch_admission.a")
}
func TestFlowAsyncLabelsAsyncGapFinallyAdmission(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_gap_finally_admission.a")
}
func TestFlowAsyncLabelsAsyncGapFinallyReturnAdmission(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_gap_finally_return_admission.a")
}

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
func TestFlowAsyncLabelsAsyncFinallyCompletion(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_completion.a")
}
func TestFlowAsyncLabelsAsyncLabels(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_labels.a")
}
func TestFlowAsyncLabelsAsyncLabelsFinally(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_labels_finally.a")
}
func TestFlowAsyncLabelsAsyncLabelsBlockFinally(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_labels_block_finally.a")
}
func TestFlowAsyncLabelsLabelBlockEdges(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_edges.a")
}
func TestFlowAsyncLabelsAsyncIterablesCellsLabels(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterables_cells_labels.a")
}
func TestFlowAsyncLabelsLabelBlockBreakSync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_break_sync.a")
}
func TestFlowAsyncLabelsLabelBlockContinueSync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_continue_sync.a")
}
func TestFlowAsyncLabelsLabelBlockSwitchBreakSync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_switch_break_sync.a")
}
func TestFlowAsyncLabelsLabelBlockOuterContinueSync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_outer_continue_sync.a")
}
func TestFlowAsyncLabelsLabelBlockBreakAsync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_break_async.a")
}
func TestFlowAsyncLabelsLabelBlockContinueAsync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_continue_async.a")
}
func TestFlowAsyncLabelsLabelBlockSwitchBreakAsync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_switch_break_async.a")
}
func TestFlowAsyncLabelsLabelBlockOuterContinueAsync(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/label_block_outer_continue_async.a")
}
func TestFlowAsyncLabelsAsyncFinallyReturn(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_return.a")
}
func TestFlowAsyncLabelsAsyncFinallyThrow(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_throw.a")
}
func TestFlowAsyncLabelsAsyncFinallyBreak(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_break.a")
}
func TestFlowAsyncLabelsAsyncFinallyContinue(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_continue.a")
}
func TestFlowAsyncLabelsAsyncFinallyVoidReturn(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_void_return.a")
}
func TestFlowAsyncLabelsAsyncFinallySnapshot(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_finally_snapshot.a")
}
func TestFlowAsyncLabelsAsyncGapLabelAdmission(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_gap_label_admission.a")
}
func TestFlowAsyncLabelsAsyncGapCatchAdmission(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_gap_catch_admission.a")
}
func TestFlowAsyncLabelsAsyncGapFinallyAdmission(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_gap_finally_admission.a")
}
func TestFlowAsyncLabelsAsyncGapFinallyReturnAdmission(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_gap_finally_return_admission.a")
}

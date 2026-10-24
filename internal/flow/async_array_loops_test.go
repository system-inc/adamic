package flow

import "testing"

func init() {
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_top_catch_binding.a", TestFlowAsyncLoopsTopCatchBinding})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_top_catch_after_await.a", TestFlowAsyncLoopsTopCatchAfterAwait})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_top_loop.a", TestFlowAsyncLoopsTopLoop})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_body.a", TestFlowAsyncLoopsForOfBody})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_mutation.a", TestFlowAsyncLoopsForOfMutation})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_return.a", TestFlowAsyncLoopsForOfReturn})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_switch_body.a", TestFlowAsyncLoopsSwitchBody})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_loop_cells.a", TestFlowAsyncLoopsLoopCells})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_loop_body_capture.a", TestFlowAsyncLoopsLoopBodyCapture})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_capture.a", TestFlowAsyncLoopsForOfCapture})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_cells.a", TestFlowAsyncLoopsForOfCells})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_body_cells.a", TestFlowAsyncLoopsForOfBodyCells})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_break_cells.a", TestFlowAsyncLoopsForOfBreakCells})
}
func TestFlowAsyncLoopsTopCatchBinding(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_top_catch_binding.a")
}
func TestFlowAsyncLoopsTopCatchAfterAwait(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_top_catch_after_await.a")
}
func TestFlowAsyncLoopsTopLoop(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_top_loop.a")
}
func TestFlowAsyncLoopsForOfBody(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_body.a")
}
func TestFlowAsyncLoopsForOfMutation(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_mutation.a")
}
func TestFlowAsyncLoopsForOfReturn(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_return.a")
}
func TestFlowAsyncLoopsSwitchBody(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_switch_body.a")
}
func TestFlowAsyncLoopsLoopCells(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_loop_cells.a")
}
func TestFlowAsyncLoopsLoopBodyCapture(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_loop_body_capture.a")
}
func TestFlowAsyncLoopsForOfCapture(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_capture.a")
}
func TestFlowAsyncLoopsForOfCells(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_cells.a")
}
func TestFlowAsyncLoopsForOfBodyCells(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_body_cells.a")
}
func TestFlowAsyncLoopsForOfBreakCells(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_break_cells.a")
}

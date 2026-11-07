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
func TestFlowAsyncLoopsTopCatchBinding(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_top_catch_binding.a")
}
func TestFlowAsyncLoopsTopCatchAfterAwait(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_top_catch_after_await.a")
}
func TestFlowAsyncLoopsTopLoop(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_top_loop.a")
}
func TestFlowAsyncLoopsForOfBody(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_body.a")
}
func TestFlowAsyncLoopsForOfMutation(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_mutation.a")
}
func TestFlowAsyncLoopsForOfReturn(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_return.a")
}
func TestFlowAsyncLoopsSwitchBody(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_switch_body.a")
}
func TestFlowAsyncLoopsLoopCells(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_loop_cells.a")
}
func TestFlowAsyncLoopsLoopBodyCapture(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_loop_body_capture.a")
}
func TestFlowAsyncLoopsForOfCapture(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_capture.a")
}
func TestFlowAsyncLoopsForOfCells(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_cells.a")
}
func TestFlowAsyncLoopsForOfBodyCells(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_body_cells.a")
}
func TestFlowAsyncLoopsForOfBreakCells(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_for_of_break_cells.a")
}

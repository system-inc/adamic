package flow

import "testing"

func TestFlowAsyncCatchReturn(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_catch_return.a")
}

func TestFlowAsyncCatches(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_catches.a")
}

func TestFlowAsyncClosures(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_closures.a")
}

func TestFlowAsyncConditions(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_conditions.a")
}

func TestFlowAsyncControl(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_control.a")
}

func TestFlowAsyncEscapedCapture(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_escaped_capture.a")
}

func TestFlowAsyncFrameCaptureSafe(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_frame_capture_safe.a")
}

func TestFlowAsyncFreshHolder(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_fresh_holder.a")
}

func TestFlowAsyncLiveValues(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_live_values.a")
}

func TestFlowAsyncMethods(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_methods.a")
}

func TestFlowAsyncMutation(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_mutation.a")
}

func TestFlowAsyncOperandOrder(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_operand_order.a")
}

func TestFlowAsyncPromiseExpression(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_promise_expression.a")
}

func TestFlowAsyncReceiver(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_receiver.a")
}

func TestFlowAsyncSnapshots(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_snapshots.a")
}

func TestFlowAsyncVoidValue(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkAllFlowProgram(test, "../oracle/testdata/async_void_value.a")
}

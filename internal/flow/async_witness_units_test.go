package flow

import "testing"

func TestFlowAsyncCatchReturn(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_catch_return.a")
}

func TestFlowAsyncCatches(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_catches.a")
}

func TestFlowAsyncClosures(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_closures.a")
}

func TestFlowAsyncConditions(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_conditions.a")
}

func TestFlowAsyncControl(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_control.a")
}

func TestFlowAsyncEscapedCapture(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_escaped_capture.a")
}

func TestFlowAsyncFrameCaptureSafe(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_frame_capture_safe.a")
}

func TestFlowAsyncFreshHolder(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_fresh_holder.a")
}

func TestFlowAsyncLiveValues(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_live_values.a")
}

func TestFlowAsyncMethods(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_methods.a")
}

func TestFlowAsyncMutation(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_mutation.a")
}

func TestFlowAsyncOperandOrder(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_operand_order.a")
}

func TestFlowAsyncPromiseExpression(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_promise_expression.a")
}

func TestFlowAsyncReceiver(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_receiver.a")
}

func TestFlowAsyncSnapshots(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_snapshots.a")
}

func TestFlowAsyncVoidValue(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_void_value.a")
}

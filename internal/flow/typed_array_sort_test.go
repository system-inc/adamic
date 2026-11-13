package flow

import "testing"

func TestFlowTypedArraySortUint8(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_uint8array.a")
}

func TestFlowTypedArraySortUint16(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_uint16array.a")
}

func TestFlowTypedArraySortInt32(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_int32array.a")
}

func TestFlowTypedArraySortFloat64(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_float64array.a")
}

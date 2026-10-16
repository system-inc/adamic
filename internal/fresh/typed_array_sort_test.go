package fresh_test

import "testing"

func TestFreshTypedArraySortUint8(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/typed_arrays_sort_uint8array.a")
}

func TestFreshTypedArraySortUint16(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/typed_arrays_sort_uint16array.a")
}

func TestFreshTypedArraySortInt32(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/typed_arrays_sort_int32array.a")
}

func TestFreshTypedArraySortFloat64(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/typed_arrays_sort_float64array.a")
}

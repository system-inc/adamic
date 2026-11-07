package fresh_test

import "testing"

func TestFreshTypedArraySortUint8(test *testing.T) {
	test.Parallel()
	beginFreshUnit(test)
	checkFreshProgram(test, "../oracle/testdata/typed_arrays_sort_uint8array.a")
}

func TestFreshTypedArraySortUint16(test *testing.T) {
	test.Parallel()
	beginFreshUnit(test)
	checkFreshProgram(test, "../oracle/testdata/typed_arrays_sort_uint16array.a")
}

func TestFreshTypedArraySortInt32(test *testing.T) {
	test.Parallel()
	beginFreshUnit(test)
	checkFreshProgram(test, "../oracle/testdata/typed_arrays_sort_int32array.a")
}

func TestFreshTypedArraySortFloat64(test *testing.T) {
	test.Parallel()
	beginFreshUnit(test)
	checkFreshProgram(test, "../oracle/testdata/typed_arrays_sort_float64array.a")
}

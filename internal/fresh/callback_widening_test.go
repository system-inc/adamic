package fresh_test

import "testing"

func TestFreshCallbackWideningLibraryMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/library-map.a")
}
func TestFreshCallbackWideningNumberCapture(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-capture.a")
}
func TestFreshCallbackWideningNumberEvery(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-every.a")
}
func TestFreshCallbackWideningNumberFilter(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-filter.a")
}
func TestFreshCallbackWideningNumberFind(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-find.a")
}
func TestFreshCallbackWideningNumberFindIndex(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-findIndex.a")
}
func TestFreshCallbackWideningNumberFindLast(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-findLast.a")
}
func TestFreshCallbackWideningNumberForEach(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-forEach.a")
}
func TestFreshCallbackWideningNumberGeneric(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-generic.a")
}
func TestFreshCallbackWideningNumberIndex(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-index.a")
}
func TestFreshCallbackWideningNumberMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-map.a")
}
func TestFreshCallbackWideningNumberNamedSort(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-named-sort.a")
}
func TestFreshCallbackWideningNumberReduceAccumulator(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-reduce-accumulator.a")
}
func TestFreshCallbackWideningNumberReduce(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-reduce.a")
}
func TestFreshCallbackWideningNumberSome(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-some.a")
}
func TestFreshCallbackWideningNumberSort(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-sort.a")
}
func TestFreshCallbackWideningNumberThrow(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-throw.a")
}
func TestFreshCallbackWideningNumberToSorted(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-toSorted.a")
}
func TestFreshCallbackWideningObjectEvery(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-every.a")
}
func TestFreshCallbackWideningObjectFilter(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-filter.a")
}
func TestFreshCallbackWideningObjectFind(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-find.a")
}
func TestFreshCallbackWideningObjectForEach(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-forEach.a")
}
func TestFreshCallbackWideningObjectGeneric(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-generic.a")
}
func TestFreshCallbackWideningObjectMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-map.a")
}
func TestFreshCallbackWideningObjectReduce(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-reduce.a")
}
func TestFreshCallbackWideningObjectSome(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-some.a")
}
func TestFreshCallbackWideningObjectSort(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-sort.a")
}

func TestFreshCallbackWideningNumberGenericMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-generic-map.a")
}

func TestFreshCallbackWideningObjectGenericMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/object-generic-map.a")
}

func TestFreshCallbackWideningNumberArguments(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/number-arguments.a")
}

func TestFreshCallbackWideningBooleanMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/boolean-map.a")
}

func TestFreshCallbackWideningStringMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/string-map.a")
}

func TestFreshCallbackWideningOptionalNumberMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/optional-number-map.a")
}

func TestFreshCallbackWideningNullableObjectMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/nullable-object-map.a")
}

func TestFreshCallbackWideningOptionalObjectMap(test *testing.T) {
	test.Parallel()
	checkFreshProgram(test, "../oracle/testdata/callback_widening/optional-object-map.a")
}

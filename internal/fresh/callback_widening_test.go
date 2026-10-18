package fresh_test

import "testing"

func TestFreshCallbackWideningLibraryMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/library-map.a")
}
func TestFreshCallbackWideningNumberCapture(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-capture.a")
}
func TestFreshCallbackWideningNumberEvery(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-every.a")
}
func TestFreshCallbackWideningNumberFilter(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-filter.a")
}
func TestFreshCallbackWideningNumberFind(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-find.a")
}
func TestFreshCallbackWideningNumberFindIndex(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-findIndex.a")
}
func TestFreshCallbackWideningNumberFindLast(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-findLast.a")
}
func TestFreshCallbackWideningNumberForEach(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-forEach.a")
}
func TestFreshCallbackWideningNumberGeneric(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-generic.a")
}
func TestFreshCallbackWideningNumberIndex(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-index.a")
}
func TestFreshCallbackWideningNumberMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-map.a")
}
func TestFreshCallbackWideningNumberNamedSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-named-sort.a")
}
func TestFreshCallbackWideningNumberReduceAccumulator(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-reduce-accumulator.a")
}
func TestFreshCallbackWideningNumberReduce(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-reduce.a")
}
func TestFreshCallbackWideningNumberSome(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-some.a")
}
func TestFreshCallbackWideningNumberSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-sort.a")
}
func TestFreshCallbackWideningNumberThrow(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-throw.a")
}
func TestFreshCallbackWideningNumberToSorted(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-toSorted.a")
}
func TestFreshCallbackWideningObjectEvery(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-every.a")
}
func TestFreshCallbackWideningObjectFilter(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-filter.a")
}
func TestFreshCallbackWideningObjectFind(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-find.a")
}
func TestFreshCallbackWideningObjectForEach(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-forEach.a")
}
func TestFreshCallbackWideningObjectGeneric(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-generic.a")
}
func TestFreshCallbackWideningObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-map.a")
}
func TestFreshCallbackWideningObjectReduce(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-reduce.a")
}
func TestFreshCallbackWideningObjectSome(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-some.a")
}
func TestFreshCallbackWideningObjectSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-sort.a")
}

func TestFreshCallbackWideningNumberGenericMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-generic-map.a")
}

func TestFreshCallbackWideningObjectGenericMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-generic-map.a")
}

func TestFreshCallbackWideningNumberArguments(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-arguments.a")
}

func TestFreshCallbackWideningBooleanMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/boolean-map.a")
}

func TestFreshCallbackWideningStringMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/string-map.a")
}

func TestFreshCallbackWideningOptionalNumberMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/optional-number-map.a")
}

func TestFreshCallbackWideningNullableObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/nullable-object-map.a")
}

func TestFreshCallbackWideningOptionalObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/optional-object-map.a")
}

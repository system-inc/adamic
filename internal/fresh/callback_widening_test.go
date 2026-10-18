package fresh_test

import "testing"

func TestCallbackWideningLibraryMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/library-map.a")
}
func TestCallbackWideningNumberCapture(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-capture.a")
}
func TestCallbackWideningNumberEvery(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-every.a")
}
func TestCallbackWideningNumberFilter(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-filter.a")
}
func TestCallbackWideningNumberFind(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-find.a")
}
func TestCallbackWideningNumberFindIndex(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-findIndex.a")
}
func TestCallbackWideningNumberFindLast(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-findLast.a")
}
func TestCallbackWideningNumberForEach(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-forEach.a")
}
func TestCallbackWideningNumberGeneric(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-generic.a")
}
func TestCallbackWideningNumberIndex(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-index.a")
}
func TestCallbackWideningNumberMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-map.a")
}
func TestCallbackWideningNumberNamedSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-named-sort.a")
}
func TestCallbackWideningNumberReduceAccumulator(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-reduce-accumulator.a")
}
func TestCallbackWideningNumberReduce(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-reduce.a")
}
func TestCallbackWideningNumberSome(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-some.a")
}
func TestCallbackWideningNumberSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-sort.a")
}
func TestCallbackWideningNumberThrow(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-throw.a")
}
func TestCallbackWideningNumberToSorted(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-toSorted.a")
}
func TestCallbackWideningObjectEvery(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-every.a")
}
func TestCallbackWideningObjectFilter(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-filter.a")
}
func TestCallbackWideningObjectFind(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-find.a")
}
func TestCallbackWideningObjectForEach(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-forEach.a")
}
func TestCallbackWideningObjectGeneric(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-generic.a")
}
func TestCallbackWideningObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-map.a")
}
func TestCallbackWideningObjectReduce(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-reduce.a")
}
func TestCallbackWideningObjectSome(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-some.a")
}
func TestCallbackWideningObjectSort(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-sort.a")
}

func TestCallbackWideningNumberGenericMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-generic-map.a")
}

func TestCallbackWideningObjectGenericMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/object-generic-map.a")
}

func TestCallbackWideningNumberArguments(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/number-arguments.a")
}

func TestCallbackWideningBooleanMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/boolean-map.a")
}

func TestCallbackWideningStringMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/string-map.a")
}

func TestCallbackWideningOptionalNumberMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/optional-number-map.a")
}

func TestCallbackWideningNullableObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/nullable-object-map.a")
}

func TestCallbackWideningOptionalObjectMap(t *testing.T) {
	t.Parallel()
	checkFreshProgram(t, "../oracle/testdata/callback_widening/optional-object-map.a")
}

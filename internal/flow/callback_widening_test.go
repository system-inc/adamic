package flow

import "testing"

func TestFlowCallbackWideningLibraryMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/library-map.a")
}
func TestFlowCallbackWideningNumberCapture(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-capture.a")
}
func TestFlowCallbackWideningNumberEvery(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-every.a")
}
func TestFlowCallbackWideningNumberFilter(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-filter.a")
}
func TestFlowCallbackWideningNumberFind(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-find.a")
}
func TestFlowCallbackWideningNumberFindIndex(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-findIndex.a")
}
func TestFlowCallbackWideningNumberFindLast(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-findLast.a")
}
func TestFlowCallbackWideningNumberForEach(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-forEach.a")
}
func TestFlowCallbackWideningNumberGeneric(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-generic.a")
}
func TestFlowCallbackWideningNumberIndex(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-index.a")
}
func TestFlowCallbackWideningNumberMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-map.a")
}
func TestFlowCallbackWideningNumberNamedSort(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-named-sort.a")
}
func TestFlowCallbackWideningNumberReduceAccumulator(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-reduce-accumulator.a")
}
func TestFlowCallbackWideningNumberReduce(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-reduce.a")
}
func TestFlowCallbackWideningNumberSome(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-some.a")
}
func TestFlowCallbackWideningNumberSort(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-sort.a")
}
func TestFlowCallbackWideningNumberThrow(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-throw.a")
}
func TestFlowCallbackWideningNumberToSorted(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-toSorted.a")
}
func TestFlowCallbackWideningObjectEvery(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-every.a")
}
func TestFlowCallbackWideningObjectFilter(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-filter.a")
}
func TestFlowCallbackWideningObjectFind(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-find.a")
}
func TestFlowCallbackWideningObjectForEach(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-forEach.a")
}
func TestFlowCallbackWideningObjectGeneric(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-generic.a")
}
func TestFlowCallbackWideningObjectMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-map.a")
}
func TestFlowCallbackWideningObjectReduce(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-reduce.a")
}
func TestFlowCallbackWideningObjectSome(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-some.a")
}
func TestFlowCallbackWideningObjectSort(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-sort.a")
}

func TestFlowCallbackWideningNumberGenericMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-generic-map.a")
}

func TestFlowCallbackWideningObjectGenericMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/object-generic-map.a")
}

func TestFlowCallbackWideningNumberArguments(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/number-arguments.a")
}

func TestFlowCallbackWideningBooleanMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/boolean-map.a")
}

func TestFlowCallbackWideningStringMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/string-map.a")
}

func TestFlowCallbackWideningOptionalNumberMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/optional-number-map.a")
}

func TestFlowCallbackWideningNullableObjectMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/nullable-object-map.a")
}

func TestFlowCallbackWideningOptionalObjectMap(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/callback_widening/optional-object-map.a")
}

func TestSlice2OptionalError(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/miscompile_2a/4ddd17f_opt_3.a")
}

func TestSlice2MiscompileControls(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/miscompile_2a/controls.a")
}

func TestSlice2UndefinedError(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/miscompile_2a/error_undefined_message.a")
}

func TestSlice2WeakSingle(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/weak_miscompiles/4ddd17f_weak_single_narrowed.a")
}

func TestSlice2WeakUnion(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/weak_miscompiles/4ddd17f_weak_union_narrowed.a")
}

func TestSlice2WeakBooleanFilter(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/weak_miscompiles/boolean_filter.a")
}

func TestSlice2WeakHeldUnion(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/weak_miscompiles/held_union.a")
}

func TestSlice2WeakMatchingMethods(test *testing.T) {
	test.Parallel()
	checkAllFlowProgram(test, "../oracle/testdata/weak_miscompiles/matching_methods.a")
}

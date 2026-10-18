package flow

import "testing"

func TestFlowCallbackWideningLibraryMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/library-map.a")
}
func TestFlowCallbackWideningNumberCapture(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-capture.a")
}
func TestFlowCallbackWideningNumberEvery(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-every.a")
}
func TestFlowCallbackWideningNumberFilter(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-filter.a")
}
func TestFlowCallbackWideningNumberFind(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-find.a")
}
func TestFlowCallbackWideningNumberFindIndex(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-findIndex.a")
}
func TestFlowCallbackWideningNumberFindLast(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-findLast.a")
}
func TestFlowCallbackWideningNumberForEach(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-forEach.a")
}
func TestFlowCallbackWideningNumberGeneric(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-generic.a")
}
func TestFlowCallbackWideningNumberIndex(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-index.a")
}
func TestFlowCallbackWideningNumberMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-map.a")
}
func TestFlowCallbackWideningNumberNamedSort(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-named-sort.a")
}
func TestFlowCallbackWideningNumberReduceAccumulator(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-reduce-accumulator.a")
}
func TestFlowCallbackWideningNumberReduce(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-reduce.a")
}
func TestFlowCallbackWideningNumberSome(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-some.a")
}
func TestFlowCallbackWideningNumberSort(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-sort.a")
}
func TestFlowCallbackWideningNumberThrow(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-throw.a")
}
func TestFlowCallbackWideningNumberToSorted(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-toSorted.a")
}
func TestFlowCallbackWideningObjectEvery(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-every.a")
}
func TestFlowCallbackWideningObjectFilter(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-filter.a")
}
func TestFlowCallbackWideningObjectFind(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-find.a")
}
func TestFlowCallbackWideningObjectForEach(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-forEach.a")
}
func TestFlowCallbackWideningObjectGeneric(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-generic.a")
}
func TestFlowCallbackWideningObjectMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-map.a")
}
func TestFlowCallbackWideningObjectReduce(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-reduce.a")
}
func TestFlowCallbackWideningObjectSome(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-some.a")
}
func TestFlowCallbackWideningObjectSort(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-sort.a")
}

func TestFlowCallbackWideningNumberGenericMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-generic-map.a")
}

func TestFlowCallbackWideningObjectGenericMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/object-generic-map.a")
}

func TestFlowCallbackWideningNumberArguments(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/number-arguments.a")
}

func TestFlowCallbackWideningBooleanMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/boolean-map.a")
}

func TestFlowCallbackWideningStringMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/string-map.a")
}

func TestFlowCallbackWideningOptionalNumberMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/optional-number-map.a")
}

func TestFlowCallbackWideningNullableObjectMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/nullable-object-map.a")
}

func TestFlowCallbackWideningOptionalObjectMap(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/callback_widening/optional-object-map.a")
}

func TestSlice2OptionalError(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/miscompile_2a/4ddd17f_opt_3.a")
}

func TestSlice2MiscompileControls(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/miscompile_2a/controls.a")
}

func TestSlice2UndefinedError(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/miscompile_2a/error_undefined_message.a")
}

func TestSlice2WeakSingle(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/weak_miscompiles/4ddd17f_weak_single_narrowed.a")
}

func TestSlice2WeakUnion(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/weak_miscompiles/4ddd17f_weak_union_narrowed.a")
}

func TestSlice2WeakBooleanFilter(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/weak_miscompiles/boolean_filter.a")
}

func TestSlice2WeakHeldUnion(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/weak_miscompiles/held_union.a")
}

func TestSlice2WeakMatchingMethods(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/weak_miscompiles/matching_methods.a")
}

package flow

import "testing"

func init() {
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_string.a", TestFlowAsyncLiveSourceForOfString})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_map.a", TestFlowAsyncLiveSourceForOfMap})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_collections_mutation.a", TestFlowAsyncLiveSourceForOfCollectionsMutation})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_ascii_function.a", TestFlowAsyncLiveIterableStringAsciiFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_ascii_module.a", TestFlowAsyncLiveIterableStringAsciiModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_emoji_function.a", TestFlowAsyncLiveIterableStringEmojiFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_emoji_module.a", TestFlowAsyncLiveIterableStringEmojiModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_surrogates_function.a", TestFlowAsyncLiveIterableStringSurrogatesFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_surrogates_module.a", TestFlowAsyncLiveIterableStringSurrogatesModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_entries_function.a", TestFlowAsyncLiveIterableMapEntriesFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_entries_module.a", TestFlowAsyncLiveIterableMapEntriesModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_keys_function.a", TestFlowAsyncLiveIterableMapKeysFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_keys_module.a", TestFlowAsyncLiveIterableMapKeysModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_values_function.a", TestFlowAsyncLiveIterableMapValuesFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_values_module.a", TestFlowAsyncLiveIterableMapValuesModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_values_function.a", TestFlowAsyncLiveIterableSetValuesFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_values_module.a", TestFlowAsyncLiveIterableSetValuesModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_entries_function.a", TestFlowAsyncLiveIterableSetEntriesFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_entries_module.a", TestFlowAsyncLiveIterableSetEntriesModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_clear_function.a", TestFlowAsyncLiveIterableMapClearFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_clear_module.a", TestFlowAsyncLiveIterableMapClearModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_readd_function.a", TestFlowAsyncLiveIterableMapReaddFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_readd_module.a", TestFlowAsyncLiveIterableMapReaddModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_rehash_function.a", TestFlowAsyncLiveIterableMapRehashFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_rehash_module.a", TestFlowAsyncLiveIterableMapRehashModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_nested_maps_function.a", TestFlowAsyncLiveIterableNestedMapsFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_nested_maps_module.a", TestFlowAsyncLiveIterableNestedMapsModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_break_function.a", TestFlowAsyncLiveIterableMapBreakFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_break_module.a", TestFlowAsyncLiveIterableMapBreakModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_return_function.a", TestFlowAsyncLiveIterableMapReturnFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_map_return_module.a", TestFlowAsyncLiveIterableMapReturnModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_break_function.a", TestFlowAsyncLiveIterableStringBreakFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_string_break_module.a", TestFlowAsyncLiveIterableStringBreakModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_return_function.a", TestFlowAsyncLiveIterableSetReturnFunction})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_iterable_set_return_module.a", TestFlowAsyncLiveIterableSetReturnModule})
	asyncFlowUnits = append(asyncFlowUnits, struct {
		path string
		test func(*testing.T)
	}{"../oracle/testdata/async_for_of_string_admission.a", TestFlowAsyncLiveForOfStringAdmission})
}
func TestFlowAsyncLiveIterableStringAsciiFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_ascii_function.a")
}
func TestFlowAsyncLiveIterableStringAsciiModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_ascii_module.a")
}
func TestFlowAsyncLiveIterableStringEmojiFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_emoji_function.a")
}
func TestFlowAsyncLiveIterableStringEmojiModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_emoji_module.a")
}
func TestFlowAsyncLiveIterableStringSurrogatesFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_surrogates_function.a")
}
func TestFlowAsyncLiveIterableStringSurrogatesModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_surrogates_module.a")
}
func TestFlowAsyncLiveIterableMapEntriesFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_entries_function.a")
}
func TestFlowAsyncLiveIterableMapEntriesModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_entries_module.a")
}
func TestFlowAsyncLiveIterableMapKeysFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_keys_function.a")
}
func TestFlowAsyncLiveIterableMapKeysModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_keys_module.a")
}
func TestFlowAsyncLiveIterableMapValuesFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_values_function.a")
}
func TestFlowAsyncLiveIterableMapValuesModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_values_module.a")
}
func TestFlowAsyncLiveIterableSetValuesFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_values_function.a")
}
func TestFlowAsyncLiveIterableSetValuesModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_values_module.a")
}
func TestFlowAsyncLiveIterableSetEntriesFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_entries_function.a")
}
func TestFlowAsyncLiveIterableSetEntriesModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_entries_module.a")
}
func TestFlowAsyncLiveIterableMapClearFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_clear_function.a")
}
func TestFlowAsyncLiveIterableMapClearModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_clear_module.a")
}
func TestFlowAsyncLiveIterableMapReaddFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_readd_function.a")
}
func TestFlowAsyncLiveIterableMapReaddModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_readd_module.a")
}
func TestFlowAsyncLiveIterableMapRehashFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_rehash_function.a")
}
func TestFlowAsyncLiveIterableMapRehashModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_rehash_module.a")
}
func TestFlowAsyncLiveIterableNestedMapsFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_nested_maps_function.a")
}
func TestFlowAsyncLiveIterableNestedMapsModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_nested_maps_module.a")
}
func TestFlowAsyncLiveIterableMapBreakFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_break_function.a")
}
func TestFlowAsyncLiveIterableMapBreakModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_break_module.a")
}
func TestFlowAsyncLiveIterableMapReturnFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_return_function.a")
}
func TestFlowAsyncLiveIterableMapReturnModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_map_return_module.a")
}
func TestFlowAsyncLiveIterableStringBreakFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_break_function.a")
}
func TestFlowAsyncLiveIterableStringBreakModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_string_break_module.a")
}
func TestFlowAsyncLiveIterableSetReturnFunction(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_return_function.a")
}
func TestFlowAsyncLiveIterableSetReturnModule(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_iterable_set_return_module.a")
}
func TestFlowAsyncLiveForOfStringAdmission(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_string_admission.a")
}

func TestFlowAsyncLiveSourceForOfString(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_string.a")
}

func TestFlowAsyncLiveSourceForOfMap(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_map.a")
}

func TestFlowAsyncLiveSourceForOfCollectionsMutation(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/async_for_of_collections_mutation.a")
}

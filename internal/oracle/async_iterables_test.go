package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_refused/async_refuse_user_iterator_function.a", false, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_refused/async_refuse_user_iterator_module.a", false, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_string.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_map.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_collections_mutation.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_ascii_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_ascii_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_emoji_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_emoji_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_surrogates_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_surrogates_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_entries_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_entries_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_keys_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_keys_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_values_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_boolean_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_values_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_values_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_boolean_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_values_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_entries_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_entries_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_clear_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_clear_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_readd_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_readd_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_rehash_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_rehash_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_nested_maps_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_nested_maps_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_break_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_break_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_return_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_map_return_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_break_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_string_break_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_return_function.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_return_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_for_of_string_admission.a", true, false})
}
func TestAsyncLiveIterableStringAsciiFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_ascii_function", "")
}
func TestAsyncLiveIterableStringAsciiModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_ascii_module", "")
}
func TestAsyncLiveIterableStringEmojiFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_emoji_function", "")
}
func TestAsyncLiveIterableStringEmojiModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_emoji_module", "")
}
func TestAsyncLiveIterableStringSurrogatesFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_surrogates_function", "")
}
func TestAsyncLiveIterableStringSurrogatesModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_surrogates_module", "")
}
func TestAsyncLiveIterableMapEntriesFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_entries_function", "")
}
func TestAsyncLiveIterableMapEntriesModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_entries_module", "")
}
func TestAsyncLiveIterableMapKeysFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_keys_function", "")
}
func TestAsyncLiveIterableMapKeysModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_keys_module", "")
}
func TestAsyncLiveIterableMapValuesFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_values_function", "")
}
func TestAsyncLiveIterableMapValuesModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_values_module", "")
}
func TestAsyncLiveIterableSetValuesFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_values_function", "")
}
func TestAsyncLiveIterableSetValuesModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_values_module", "")
}
func TestAsyncLiveIterableSetEntriesFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_entries_function", "")
}
func TestAsyncLiveIterableSetEntriesModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_entries_module", "")
}
func TestAsyncLiveIterableMapClearFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_clear_function", "")
}
func TestAsyncLiveIterableMapClearModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_clear_module", "")
}
func TestAsyncLiveIterableMapReaddFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_readd_function", "")
}
func TestAsyncLiveIterableMapReaddModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_readd_module", "")
}
func TestAsyncLiveIterableMapRehashFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_rehash_function", "")
}
func TestAsyncLiveIterableMapRehashModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_rehash_module", "")
}
func TestAsyncLiveIterableNestedMapsFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_nested_maps_function", "")
}
func TestAsyncLiveIterableNestedMapsModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_nested_maps_module", "")
}
func TestAsyncLiveIterableMapBreakFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_break_function", "")
}
func TestAsyncLiveIterableMapBreakModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_break_module", "")
}
func TestAsyncLiveIterableMapReturnFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_return_function", "")
}
func TestAsyncLiveIterableMapReturnModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_return_module", "")
}
func TestAsyncLiveIterableStringBreakFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_break_function", "")
}
func TestAsyncLiveIterableStringBreakModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_string_break_module", "")
}
func TestAsyncLiveIterableSetReturnFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_return_function", "")
}
func TestAsyncLiveIterableSetReturnModule(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_return_module", "")
}
func TestAsyncLiveForOfStringAdmission(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_string_admission", "")
}

// Advance twice instead of once. Both emitters must still compile and clean up;
// only the source oracle may reject this change in live iteration order.
func asyncLiveWrongEntry(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Interface {
			out := reflect.New(value.Type()).Elem()
			if !value.IsNil() {
				out.Set(rewrite(value.Elem()))
			}
			return out
		}
		if loop, ok := value.Interface().(ir.Loop); ok && !changed && len(loop.Body) > 0 {
			if decl, ok := loop.Body[0].(ir.Declare); ok {
				if call, ok := decl.Value.(ir.CallClosure); ok {
					if property, ok := call.Closure.(ir.Property); ok && property.Name == "next" {
						changed = true
						loop.Body = append([]ir.Statement{ir.Evaluate{Value: call}}, loop.Body...)
						return reflect.ValueOf(loop)
					}
				}
			}
		}
		switch value.Kind() {
		case reflect.Struct:
			out := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				out.Field(index).Set(rewrite(value.Field(index)))
			}
			return out
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for index := 0; index < value.Len(); index++ {
				out.Index(index).Set(rewrite(value.Index(index)))
			}
			return out
		}
		return value
	}
	for index := range program.Functions {
		program.Functions[index].Body = rewrite(
			reflect.ValueOf(program.Functions[index].Body),
		).Interface().([]ir.Statement)
	}
	if !changed {
		test.Fatal("no iterator step mutated")
	}
	code := native.C(program)
	binary := filepath.Join(test.TempDir(), "wrong-entry")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
	for backend, result := range map[string]run{
		"native":     execute(test, binary),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if disagreement(truth, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("%s must be caught only by Node stdout: %+v", backend, result)
		}
	}
	test.Log("Node stdout caught an extra iterator advance in native and JavaScript")
}

func TestAsyncLiveMutantIterableMapBreakFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_break_function")
}

func TestAsyncLiveMutantIterableMapBreakModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_break_module")
}

func TestAsyncLiveMutantIterableMapClearFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_clear_function")
}

func TestAsyncLiveMutantIterableMapClearModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_clear_module")
}

func TestAsyncLiveMutantIterableMapEntriesFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_entries_function")
}

func TestAsyncLiveMutantIterableMapEntriesModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_entries_module")
}

func TestAsyncLiveMutantIterableMapKeysFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_keys_function")
}

func TestAsyncLiveMutantIterableMapKeysModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_keys_module")
}

func TestAsyncLiveMutantIterableMapReaddFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_readd_function")
}

func TestAsyncLiveMutantIterableMapReaddModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_readd_module")
}

func TestAsyncLiveMutantIterableMapRehashFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_rehash_function")
}

func TestAsyncLiveMutantIterableMapRehashModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_rehash_module")
}

func TestAsyncLiveMutantIterableMapReturnFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_return_function")
}

func TestAsyncLiveMutantIterableMapReturnModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_return_module")
}

func TestAsyncLiveMutantIterableMapValuesFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_values_function")
}

func TestAsyncLiveMutantIterableMapValuesModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_map_values_module")
}

func TestAsyncLiveMutantIterableNestedMapsFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_nested_maps_function")
}

func TestAsyncLiveMutantIterableNestedMapsModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_nested_maps_module")
}

func TestAsyncLiveMutantIterableSetEntriesFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_entries_function")
}

func TestAsyncLiveMutantIterableSetEntriesModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_entries_module")
}

func TestAsyncLiveMutantIterableSetReturnFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_return_function")
}

func TestAsyncLiveMutantIterableSetReturnModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_return_module")
}

func TestAsyncLiveMutantIterableSetValuesFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_values_function")
}

func TestAsyncLiveMutantIterableSetValuesModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_set_values_module")
}

func TestAsyncLiveMutantIterableStringAsciiFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_ascii_function")
}

func TestAsyncLiveMutantIterableStringAsciiModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_ascii_module")
}

func TestAsyncLiveMutantIterableStringBreakFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_break_function")
}

func TestAsyncLiveMutantIterableStringBreakModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_break_module")
}

func TestAsyncLiveMutantIterableStringEmojiFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_emoji_function")
}

func TestAsyncLiveMutantIterableStringEmojiModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_emoji_module")
}

func TestAsyncLiveMutantIterableStringSurrogatesFunction(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_surrogates_function")
}

func TestAsyncLiveMutantIterableStringSurrogatesModule(test *testing.T) {
	test.Parallel()
	asyncLiveWrongEntry(test, "async_iterable_string_surrogates_module")
}

func TestAsyncLiveSourceForOfString(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_string", "")
}

func TestAsyncLiveSourceForOfMap(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_map", "")
}

func TestAsyncLiveSourceForOfCollectionsMutation(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_for_of_collections_mutation", "")
}

// Drop only the frame's owned iterator, leaving all values and control flow intact.
// The source oracle agrees; the counted and sanitizer cleanup checks must fail.
func asyncLiveReleaseMutant(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	function, local := -1, -1
	var find func(reflect.Value)
	find = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface {
			if !value.IsNil() {
				find(value.Elem())
			}
			return
		}
		if declare, ok := value.Interface().(ir.Declare); ok {
			switch declare.Value.(type) {
			case ir.CollectionIterator, ir.StringIterator:
				local = declare.Local
			}
		}
		switch value.Kind() {
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				find(value.Field(index))
			}
		case reflect.Slice:
			for index := 0; index < value.Len(); index++ {
				find(value.Index(index))
			}
		}
	}
	for index, f := range program.Functions {
		find(reflect.ValueOf(f.Body))
		if local >= 0 {
			function = index
			break
		}
	}
	if function < 0 {
		test.Fatal("iterator slot not found")
	}
	slot := -1
	for index, l := range program.Functions[function].FrameEnvironment {
		if l == local {
			slot = index
		}
	}
	if slot < 0 {
		test.Fatal("iterator not owned by frame")
	}
	code := native.C(program)
	readable := strings.Join(program.Functions[function].Source.Declaration, "_")
	begin, end := -1, -1
	offset := 0
	for {
		at := strings.Index(code[offset:], "static void adamic_async_finish_")
		if at < 0 {
			break
		}
		at += offset
		lineEnd := strings.IndexByte(code[at:], '\n')
		if lineEnd < 0 {
			break
		}
		if strings.Contains(code[at:at+lineEnd], "_"+readable+"_") {
			begin = at
			tail := strings.Index(code[begin:], "static void adamic_async_abandon_")
			if tail >= 0 {
				end = begin + tail
			}
			break
		}
		offset = at + lineEnd
	}
	if begin < 0 || end < begin {
		test.Fatal("missing stable iterator frame cleanup")
	}
	cleanup := fmt.Sprintf("adamic_release(adamic_slot_%d);", slot)
	part := code[begin:end]
	if strings.Count(part, cleanup) != 1 {
		test.Fatal("cleanup slot not unique")
	}
	code = code[:begin] + strings.Replace(part, cleanup, "(void)0;", 1) + code[end:]
	binary := filepath.Join(test.TempDir(), "leaked-iterator")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	report := leakcheck.Report(test, code, binary)
	if !strings.Contains(report, "detected memory leaks") && !strings.Contains(report, "heap values leaked:") {
		test.Fatalf("iterator cleanup mutant survived: %s", report)
	}
	counted := filepath.Join(test.TempDir(), "counted-iterator")
	if buildError := native.Build(code, counted, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if report := leakcheck.Unbalanced(
		leakRun(execute(test, counted)),
	); !strings.Contains(report, "heap values leaked:") {
		test.Fatalf("counted cleanup mutant survived: %s", report)
	}
	test.Log("dropping the iterator release is caught by sanitizer and counted cleanup")
}
func TestAsyncLiveReleaseMapBreak(test *testing.T) {
	test.Parallel()
	asyncLiveReleaseMutant(test, "async_iterable_map_break_function")
}
func TestAsyncLiveReleaseMapReturn(test *testing.T) {
	test.Parallel()
	asyncLiveReleaseMutant(test, "async_iterable_map_return_function")
}
func TestAsyncLiveReleaseStringBreak(test *testing.T) {
	test.Parallel()
	asyncLiveReleaseMutant(test, "async_iterable_string_break_function")
}
func TestAsyncLiveReleaseSetReturn(test *testing.T) {
	test.Parallel()
	asyncLiveReleaseMutant(test, "async_iterable_set_return_function")
}

func TestAsyncLiveUserIteratorFunctionRemainsUnproven(test *testing.T) {
	test.Parallel()
	asyncLoopsRefusal(test, "async_refuse_user_iterator_function", "async for-of over a user iterator", "")
}

func TestAsyncLiveUserIteratorModuleRemainsUnproven(test *testing.T) {
	test.Parallel()
	asyncLoopsRefusal(test, "async_refuse_user_iterator_module", "async for-of over a user iterator", "")
}

func TestAsyncLiveIterableMapBooleanFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_map_boolean_function", "false\n")
}

func TestAsyncLiveIterableSetBooleanFunction(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterable_set_boolean_function", "false\n")
}

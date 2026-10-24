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
	}{"internal/oracle/testdata/async_iterable_map_values_module.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterable_set_values_function.a", true, false})
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
func TestAsyncLiveIterableStringAsciiFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_ascii_function", "")
}
func TestAsyncLiveIterableStringAsciiModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_ascii_module", "")
}
func TestAsyncLiveIterableStringEmojiFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_emoji_function", "")
}
func TestAsyncLiveIterableStringEmojiModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_emoji_module", "")
}
func TestAsyncLiveIterableStringSurrogatesFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_surrogates_function", "")
}
func TestAsyncLiveIterableStringSurrogatesModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_surrogates_module", "")
}
func TestAsyncLiveIterableMapEntriesFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_entries_function", "")
}
func TestAsyncLiveIterableMapEntriesModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_entries_module", "")
}
func TestAsyncLiveIterableMapKeysFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_keys_function", "")
}
func TestAsyncLiveIterableMapKeysModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_keys_module", "")
}
func TestAsyncLiveIterableMapValuesFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_values_function", "")
}
func TestAsyncLiveIterableMapValuesModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_values_module", "")
}
func TestAsyncLiveIterableSetValuesFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_values_function", "")
}
func TestAsyncLiveIterableSetValuesModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_values_module", "")
}
func TestAsyncLiveIterableSetEntriesFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_entries_function", "")
}
func TestAsyncLiveIterableSetEntriesModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_entries_module", "")
}
func TestAsyncLiveIterableMapClearFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_clear_function", "")
}
func TestAsyncLiveIterableMapClearModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_clear_module", "")
}
func TestAsyncLiveIterableMapReaddFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_readd_function", "")
}
func TestAsyncLiveIterableMapReaddModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_readd_module", "")
}
func TestAsyncLiveIterableMapRehashFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_rehash_function", "")
}
func TestAsyncLiveIterableMapRehashModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_rehash_module", "")
}
func TestAsyncLiveIterableNestedMapsFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_nested_maps_function", "")
}
func TestAsyncLiveIterableNestedMapsModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_nested_maps_module", "")
}
func TestAsyncLiveIterableMapBreakFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_break_function", "")
}
func TestAsyncLiveIterableMapBreakModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_break_module", "")
}
func TestAsyncLiveIterableMapReturnFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_return_function", "")
}
func TestAsyncLiveIterableMapReturnModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_map_return_module", "")
}
func TestAsyncLiveIterableStringBreakFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_break_function", "")
}
func TestAsyncLiveIterableStringBreakModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_string_break_module", "")
}
func TestAsyncLiveIterableSetReturnFunction(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_return_function", "")
}
func TestAsyncLiveIterableSetReturnModule(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterable_set_return_module", "")
}
func TestAsyncLiveForOfStringAdmission(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_string_admission", "")
}

// Advance twice instead of once. Both emitters must still compile and clean up;
// only the source oracle may reject this change in live iteration order.
func asyncLiveWrongEntry(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		if v.Kind() == reflect.Interface {
			out := reflect.New(v.Type()).Elem()
			if !v.IsNil() {
				out.Set(rewrite(v.Elem()))
			}
			return out
		}
		if loop, ok := v.Interface().(ir.Loop); ok && !changed && len(loop.Body) > 0 {
			if decl, ok := loop.Body[0].(ir.Declare); ok {
				if call, ok := decl.Value.(ir.CallClosure); ok {
					if prop, ok := call.Closure.(ir.Property); ok && prop.Name == "next" {
						changed = true
						loop.Body = append([]ir.Statement{ir.Evaluate{Value: call}}, loop.Body...)
						return reflect.ValueOf(loop)
					}
				}
			}
		}
		switch v.Kind() {
		case reflect.Struct:
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(rewrite(v.Field(i)))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(rewrite(v.Index(i)))
			}
			return out
		}
		return v
	}
	for i := range p.Functions {
		p.Functions[i].Body = rewrite(reflect.ValueOf(p.Functions[i].Body)).Interface().([]ir.Statement)
	}
	if !changed {
		t.Fatal("no iterator step mutated")
	}
	code := native.C(p)
	binary := filepath.Join(t.TempDir(), "wrong-entry")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
	for backend, result := range map[string]run{"native": execute(t, binary), "JavaScript": onJavaScriptBackend(t, p)} {
		if disagreement(truth, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s must be caught only by Node stdout: %+v", backend, result)
		}
	}
	t.Log("Node stdout caught an extra iterator advance in native and JavaScript")
}

func TestAsyncLiveMutantIterableMapBreakFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_break_function")
}

func TestAsyncLiveMutantIterableMapBreakModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_break_module")
}

func TestAsyncLiveMutantIterableMapClearFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_clear_function")
}

func TestAsyncLiveMutantIterableMapClearModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_clear_module")
}

func TestAsyncLiveMutantIterableMapEntriesFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_entries_function")
}

func TestAsyncLiveMutantIterableMapEntriesModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_entries_module")
}

func TestAsyncLiveMutantIterableMapKeysFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_keys_function")
}

func TestAsyncLiveMutantIterableMapKeysModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_keys_module")
}

func TestAsyncLiveMutantIterableMapReaddFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_readd_function")
}

func TestAsyncLiveMutantIterableMapReaddModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_readd_module")
}

func TestAsyncLiveMutantIterableMapRehashFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_rehash_function")
}

func TestAsyncLiveMutantIterableMapRehashModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_rehash_module")
}

func TestAsyncLiveMutantIterableMapReturnFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_return_function")
}

func TestAsyncLiveMutantIterableMapReturnModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_return_module")
}

func TestAsyncLiveMutantIterableMapValuesFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_values_function")
}

func TestAsyncLiveMutantIterableMapValuesModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_map_values_module")
}

func TestAsyncLiveMutantIterableNestedMapsFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_nested_maps_function")
}

func TestAsyncLiveMutantIterableNestedMapsModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_nested_maps_module")
}

func TestAsyncLiveMutantIterableSetEntriesFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_entries_function")
}

func TestAsyncLiveMutantIterableSetEntriesModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_entries_module")
}

func TestAsyncLiveMutantIterableSetReturnFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_return_function")
}

func TestAsyncLiveMutantIterableSetReturnModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_return_module")
}

func TestAsyncLiveMutantIterableSetValuesFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_values_function")
}

func TestAsyncLiveMutantIterableSetValuesModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_set_values_module")
}

func TestAsyncLiveMutantIterableStringAsciiFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_ascii_function")
}

func TestAsyncLiveMutantIterableStringAsciiModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_ascii_module")
}

func TestAsyncLiveMutantIterableStringBreakFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_break_function")
}

func TestAsyncLiveMutantIterableStringBreakModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_break_module")
}

func TestAsyncLiveMutantIterableStringEmojiFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_emoji_function")
}

func TestAsyncLiveMutantIterableStringEmojiModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_emoji_module")
}

func TestAsyncLiveMutantIterableStringSurrogatesFunction(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_surrogates_function")
}

func TestAsyncLiveMutantIterableStringSurrogatesModule(t *testing.T) {
	t.Parallel()
	asyncLiveWrongEntry(t, "async_iterable_string_surrogates_module")
}

func TestAsyncLiveSourceForOfString(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_string", "")
}

func TestAsyncLiveSourceForOfMap(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_map", "")
}

func TestAsyncLiveSourceForOfCollectionsMutation(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_for_of_collections_mutation", "")
}

// Drop only the frame's owned iterator, leaving all values and control flow intact.
// The source oracle agrees; the counted and sanitizer cleanup checks must fail.
func asyncLiveReleaseMutant(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	function, local := -1, -1
	var find func(reflect.Value)
	find = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		if v.Kind() == reflect.Interface {
			if !v.IsNil() {
				find(v.Elem())
			}
			return
		}
		if d, ok := v.Interface().(ir.Declare); ok {
			switch d.Value.(type) {
			case ir.CollectionIterator, ir.StringIterator:
				local = d.Local
			}
		}
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				find(v.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				find(v.Index(i))
			}
		}
	}
	for i, f := range p.Functions {
		find(reflect.ValueOf(f.Body))
		if local >= 0 {
			function = i
			break
		}
	}
	if function < 0 {
		t.Fatal("iterator slot not found")
	}
	slot := -1
	for i, l := range p.Functions[function].FrameEnvironment {
		if l == local {
			slot = i
		}
	}
	if slot < 0 {
		t.Fatal("iterator not owned by frame")
	}
	code := native.C(p)
	readable := strings.Join(p.Functions[function].Source.Declaration, "_")
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
		t.Fatal("missing stable iterator frame cleanup")
	}
	cleanup := fmt.Sprintf("adamic_release(adamic_slot_%d);", slot)
	part := code[begin:end]
	if strings.Count(part, cleanup) != 1 {
		t.Fatal("cleanup slot not unique")
	}
	code = code[:begin] + strings.Replace(part, cleanup, "(void)0;", 1) + code[end:]
	binary := filepath.Join(t.TempDir(), "leaked-iterator")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	report := leakcheck.Report(t, code, binary)
	if !strings.Contains(report, "detected memory leaks") && !strings.Contains(report, "heap values leaked:") {
		t.Fatalf("iterator cleanup mutant survived: %s", report)
	}
	counted := filepath.Join(t.TempDir(), "counted-iterator")
	if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Unbalanced(leakRun(execute(t, counted))); !strings.Contains(report, "heap values leaked:") {
		t.Fatalf("counted cleanup mutant survived: %s", report)
	}
	t.Log("dropping the iterator release is caught by sanitizer and counted cleanup")
}
func TestAsyncLiveReleaseMapBreak(t *testing.T) {
	t.Parallel()
	asyncLiveReleaseMutant(t, "async_iterable_map_break_function")
}
func TestAsyncLiveReleaseMapReturn(t *testing.T) {
	t.Parallel()
	asyncLiveReleaseMutant(t, "async_iterable_map_return_function")
}
func TestAsyncLiveReleaseStringBreak(t *testing.T) {
	t.Parallel()
	asyncLiveReleaseMutant(t, "async_iterable_string_break_function")
}
func TestAsyncLiveReleaseSetReturn(t *testing.T) {
	t.Parallel()
	asyncLiveReleaseMutant(t, "async_iterable_set_return_function")
}

func TestAsyncLiveUserIteratorFunctionRemainsUnproven(t *testing.T) {
	t.Parallel()
	asyncLoopsRefusal(t, "async_refuse_user_iterator_function", "async for-of over a user iterator", "")
}

func TestAsyncLiveUserIteratorModuleRemainsUnproven(t *testing.T) {
	t.Parallel()
	asyncLoopsRefusal(t, "async_refuse_user_iterator_module", "async for-of over a user iterator", "")
}

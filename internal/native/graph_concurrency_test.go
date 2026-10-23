package native

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

// These witnesses exercise runtime allocation and ownership directly, before any
// compiler graph classification is enabled. Mutants must compile and fail at run time.
func graphRuntimeGuard(t *testing.T, source, file, old, changed, diagnostic string, modes ...bool) {
	t.Helper()
	options := Options{Sanitize: true, Count: true}
	files, flags, compiler, version, err := runtimeInputs("", options, []string{"-DADAMIC_CANONICAL_CLOSURES=1"})
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("graph-runtime")
	if err != nil {
		t.Fatal(err)
	}
	if len(modes) == 0 {
		modes = []bool{false, true}
	}
	for _, mutant := range modes {
		if mutant {
			found := false
			for i := range files {
				if files[i].name != file {
					continue
				}
				text := string(files[i].contents)
				if strings.Count(text, old) != 1 {
					t.Fatalf("missing or ambiguous mutant in %s: %q", file, old)
				}
				files[i].contents = []byte(strings.Replace(text, old, changed, 1))
				found = true
			}
			if !found {
				t.Fatal("missing mutant file", file)
			}
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			t.Fatal(err)
		}
		binary := buildHarnessWithLibrary(t, source, flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1")
		output, err := command.CombinedOutput()
		release()
		if mutant {
			if err == nil || !strings.Contains(string(output), diagnostic) {
				t.Fatalf("mutant survived or failed differently: %v\n%s", err, output)
			}
			t.Logf("caught %s mutant by %s", file, diagnostic)
		} else if err != nil {
			t.Fatalf("control: %v\n%s", err, output)
		}
	}
}

const graphObjectPrelude = `#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
static const char *const names[] = {"next", "number", "last"};
static const bool refs[] = {true, false, false};
static const adamic_field_kind shape_kinds[] = {adamic_field_reference, adamic_field_number, adamic_field_number};
static const adamic_shape shape __attribute__((unused)) = {3, names, refs, NULL, shape_kinds};
`

func TestGraphAdoptionObjectBytes(t *testing.T) {
	t.Parallel()
	graphAdoptionObjectBytes(t, false)
}

func TestGraphAdoptionObjectBytesMutant(t *testing.T) {
	t.Parallel()
	graphAdoptionObjectBytes(t, true)
}

func graphAdoptionObjectBytes(t *testing.T, mutant bool) {
	t.Helper()
	source := graphObjectPrelude + `int main(void) {
 adamic_object *object = adamic_object_new(&shape);
 adamic_object_field_types(object)[2] = 173;
 object = adamic_graph_adopt_owned(object, adamic_object_size(3));
 assert(adamic_object_field_types(object)[2] == 173);
 adamic_release(object);
 return 0;
}
`
	graphRuntimeGuard(t, source, "heap.c", "memcpy(heap, old, size);", "memcpy(heap, old, size - 3 * (sizeof(size_t) + 2));", "Assertion", mutant)
}

func TestGraphAdoptionSelfPointer(t *testing.T) {
	t.Parallel()
	source := graphObjectPrelude + `int main(void) {
 adamic_object *object = adamic_object_new(&shape);
 object->slots[0].reference = object;
 object = adamic_graph_adopt_owned(object, adamic_object_size(3));
 assert(object->slots[0].reference == object);
 adamic_release(object);
 return 0;
}
`
	graphRuntimeGuard(t, source, "graph_regions.c", "object->slots[i].reference = heap;", "object->slots[i].reference = value;", "AddressSanitizer: heap-use-after-free")
}

// Undefined reference slots are gaps on this base, which has no separate hole bitmap.
// An owned child beyond gaps must be joined during adoption, then found by the report walk.
func TestGraphAdoptionSparseArray(t *testing.T) {
	t.Parallel()
	source := graphObjectPrelude + `int main(void) {
 adamic_cell *child = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 1}, false), sizeof(adamic_cell));
 adamic_array *array = adamic_array_new(8, true);
 for (size_t i = 0; i < 7; i++) adamic_array_push(array, (adamic_value){.reference = NULL});
 adamic_array_push(array, (adamic_value){.reference = adamic_retain(child)});
 array = adamic_graph_adopt_owned(array, sizeof *array);
 adamic_release(child);
 assert(((adamic_cell *)array->elements[7].reference)->value.number == 1);
 adamic_release(array);
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`
	graphRuntimeGuard(t, source, "graph_regions.c", "for (size_t i = 0; i < array->length; i++) { visit(array->elements[i].reference); }", "for (size_t i = 0; i < array->length && array->elements[i].reference != NULL; i++) { visit(array->elements[i].reference); }", "graph edge crosses regions")
}

func TestGraphRegionSparseArrayRelease(t *testing.T) {
	t.Parallel()
	source := graphObjectPrelude + `int main(void) {
 adamic_array *array = adamic_array_new(8, true);
 for (size_t i = 0; i < 7; i++) adamic_array_push(array, (adamic_value){.reference = NULL});
 static adamic_string a = ADAMIC_STRING("built "), b = ADAMIC_STRING("now");
 adamic_array_push(array, (adamic_value){.reference = adamic_string_concat(2, (adamic_string *const[]){&a, &b})});
 array = adamic_graph_adopt_owned(array, sizeof *array);
 adamic_release(array);
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`
	graphRuntimeGuard(t, source, "heap.c", "index < array->length; index++", "index < array->length && array->elements[index].reference != NULL; index++", "Assertion")
}

func TestGraphSharedAllocationCannotBeAdopted(t *testing.T) {
	t.Parallel()
	graphSharedAllocationCannotBeAdopted(t, false)
}

func TestGraphSharedAllocationCannotBeAdoptedMutant(t *testing.T) {
	t.Parallel()
	graphSharedAllocationCannotBeAdopted(t, true)
}

func graphSharedAllocationCannotBeAdopted(t *testing.T, mutant bool) {
	t.Helper()
	source := `#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
int main(void) {
 adamic_cell *cell = adamic_cell_new((adamic_value){.number = 1}, false);
 adamic_share(cell);
 adamic_graph_adopt(cell, sizeof *cell);
 assert(0 && "shared allocation adopted");
}
`
	// Here the control is the intentional refusal, and the mutant must reach the assertion.
	graphRefusalGuard(t, source, mutant)
}

func graphRefusalGuard(t *testing.T, source string, mutant bool) {
	t.Helper()
	files, flags, compiler, version, err := runtimeInputs("", Options{Sanitize: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("graph-refusal")
	if err != nil {
		t.Fatal(err)
	}
	if mutant {
		for i := range files {
			if files[i].name == "graph_regions.c" {
				files[i].contents = []byte(strings.Replace(string(files[i].contents), "(heap->slab & (ADAMIC_SHARED_HEADER | ADAMIC_REGION_VALUE)) != 0 || heap->references != 1", "adamic_reference_count(heap) != 1", 1))
			}
		}
	}
	library, err := cachedRuntime(files, flags, compiler, version, cache)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildHarnessWithLibrary(t, source, flags, compiler, library)
	command, release := boundedrun.Command(5*time.Second, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	output, err := command.CombinedOutput()
	release()
	expected := "graph adoption needs a new allocation"
	if mutant {
		expected = "Assertion"
	}
	if err == nil || !strings.Contains(string(output), expected) {
		t.Fatalf("mutant=%t: %v %s", mutant, err, output)
	}
}

func TestGraphParallelOutsideCountIsRefused(t *testing.T) {
	t.Parallel()
	if goruntime.GOOS != "linux" {
		t.Skip("requires Linux ThreadSanitizer")
	}
	const source = `#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "graph_regions.h"
#include "parallel.h"
#include <pthread.h>
#include <assert.h>
static pthread_barrier_t barrier;
static adamic_value work(adamic_closure *self, adamic_value *arguments) {
 size_t index = (size_t)arguments[0].number;
 adamic_cell *member = self->cells[index]->value.reference;
 pthread_barrier_wait(&barrier);
 for (size_t i = 0; i < 10000; i++) { adamic_retain(member); adamic_release(member); }
 return (adamic_value){.number = 1};
}
int main(void) {
 pthread_barrier_init(&barrier, NULL, 2);
 adamic_cell *a = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 1}, false), sizeof *a);
 adamic_cell *b = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 2}, false), sizeof *b);
 adamic_graph_merge(a, b);
 adamic_closure *callback = adamic_closure_new(work, 2);
 callback->cells[0] = adamic_cell_new((adamic_value){.reference = a}, true);
 callback->cells[1] = adamic_cell_new((adamic_value){.reference = b}, true);
 adamic_array *items = adamic_array_new(2, false);
 adamic_array_push(items, (adamic_value){.number = 0});
 adamic_array_push(items, (adamic_value){.number = 1});
 adamic_array *results = adamic_parallel_map(items, callback, false);
 adamic_release(results); adamic_release(items); adamic_release(callback);
 adamic_parallel_shutdown();
 pthread_barrier_destroy(&barrier);
 return 0;
}
`
	files, flags, compiler, version, err := runtimeInputs("", Options{ThreadSanitize: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("graph-task-refusal")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []bool{false, true} {
		if mutant {
			for i := range files {
				if files[i].name == "share.c" {
					old := `adamic_panic("a graph region can't cross into parallel work yet", sizeof "a graph region can't cross into parallel work yet" - 1);`
					text := string(files[i].contents)
					if strings.Count(text, old) != 1 {
						t.Fatal("missing graph sharing guard")
					}
					files[i].contents = []byte(strings.Replace(text, old, "continue;", 1))
				}
			}
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			t.Fatal(err)
		}
		binary := buildHarnessWithLibrary(t, source, flags, compiler, library)
		command, release := boundedrun.Command(10*time.Second, binary)
		command.Env = append(os.Environ(), "ADAMIC_THREADS=2", "TSAN_OPTIONS=halt_on_error=1")
		output, err := command.CombinedOutput()
		release()
		expected := "a graph region can't cross into parallel work yet"
		if mutant {
			expected = "WARNING: ThreadSanitizer: data race"
		}
		if err == nil || !strings.Contains(string(output), expected) {
			t.Fatalf("mutant=%t: %v\n%s", mutant, err, output)
		}
		if mutant && !strings.Contains(string(output), "graph_regions.c") {
			t.Fatalf("race did not involve region counts: %s", output)
		}
		t.Logf("mutant=%t caught by %s", mutant, expected)
	}
}

func TestGraphCanonicalClosureAdoption(t *testing.T) {
	t.Parallel()
	source := graphObjectPrelude + `static adamic_value code(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments; return (adamic_value){.number = 1};
}
int main(void) {
 adamic_environment *environment = adamic_environment_new(1);
 environment = adamic_graph_adopt(environment, sizeof *environment + sizeof(adamic_cell));
 adamic_cell *cell = &environment->cells[0];
 adamic_closure *first = adamic_closure_canonical_graph(cell, code, 1, (adamic_cell *const[]){cell});
 assert(adamic_graph_is(first));
 adamic_closure *second = adamic_closure_canonical_graph(cell, code, 1, (adamic_cell *const[]){cell});
 assert(first == second);
 adamic_release(first); adamic_release(second); adamic_release(environment);
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`
	graphRuntimeGuard(t, source, "closure.c", "\tif (graph) { closure = adamic_graph_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]); }", "\tif (graph) { (void)closure; }", "Assertion")
}

// This base's Map buffers are out of line, so its relocation hook is a no-op.
// Instrument that hook in both runs to prove adoption invokes the extension seam;
// changing only the call site must fail, even though this layout needs no fixup.
func TestGraphAdoptionCallsMapRelocation(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
int main(void) {
 adamic_map *map = adamic_map_new(false, false);
 adamic_map_set(map, (adamic_value){.number = 7}, (adamic_value){.number = 9});
 map = adamic_graph_adopt_owned(map, sizeof *map);
 assert(map->iterating == 1);
 assert(adamic_map_get(map, (adamic_value){.number = 7})->number == 9);
 map->iterating = 0;
 adamic_release(map);
 return 0;
}
`
	files, flags, compiler, version, err := runtimeInputs("", Options{Sanitize: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("graph-relocation")
	if err != nil {
		t.Fatal(err)
	}
	for i := range files {
		if files[i].name == "map.c" {
			text := string(files[i].contents)
			if strings.Count(text, "\t(void)to;") != 1 {
				t.Fatal("missing relocation observation seam")
			}
			files[i].contents = []byte(strings.Replace(text, "\t(void)to;", "\tto->iterating++;", 1))
		}
	}
	for _, mutant := range []bool{false, true} {
		if mutant {
			for i := range files {
				if files[i].name == "heap.c" {
					old := "adamic_map_relocated((adamic_map *)old, (adamic_map *)heap);"
					text := string(files[i].contents)
					if strings.Count(text, old) != 1 {
						t.Fatal("missing relocation call")
					}
					files[i].contents = []byte(strings.Replace(text, old, "(void)heap;", 1))
				}
			}
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			t.Fatal(err)
		}
		binary := buildHarnessWithLibrary(t, source, flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary)
		output, err := command.CombinedOutput()
		release()
		if mutant {
			if err == nil || !strings.Contains(string(output), "Assertion") {
				t.Fatalf("relocation mutant: %v %s", err, output)
			}
		} else if err != nil {
			t.Fatalf("relocation control: %v %s", err, output)
		}
	}
}

const graphOutsideCounts = `#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
int main(void) {
 adamic_cell *a = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 1}, false), sizeof *a);
 adamic_retain(a);
 adamic_cell *b = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 2}, false), sizeof *b);
 adamic_graph_merge(a, b);
 adamic_release(a); adamic_release(a); adamic_release(b);
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`

func TestGraphMembersCountOnTheirRegion(t *testing.T) {
	t.Parallel()
	old := "ADAMIC_OWNERSHIP void adamic_release(void *value) {\n\tADAMIC_COUNT_RELEASE();\n\tadamic_heap *heap = value;\n\tif (adamic_program_is(heap)) { return; }\n\tif (heap != NULL && (heap->slab & ADAMIC_SLOW_COUNT) == 0)"
	graphRuntimeGuard(t, graphOutsideCounts, "adamic.h", old, strings.Replace(old, "ADAMIC_SLOW_COUNT", "ADAMIC_SHARED_HEADER", 1), "Assertion")
}

func TestGraphMergeSumsOutsideCounts(t *testing.T) {
	t.Parallel()
	graphRuntimeGuard(t, graphOutsideCounts, "graph_regions.c", "a->references += two->references;", "a->references += 0;", "AddressSanitizer: heap-use-after-free")
}

const graphMapAdoptions = `#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
static const char *const names[] = {"0", "1"};
static const bool references[] = {false, true};
static const adamic_field_kind pair_kinds[] = {adamic_field_number, adamic_field_reference};
static const adamic_shape pair = {2, names, references, NULL, pair_kinds};
int main(void) {
 adamic_cell *child = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 19}, false), sizeof *child);
 adamic_map *map = adamic_graph_adopt(adamic_map_new(false, true), sizeof(adamic_map));
 adamic_map_set(map, (adamic_value){.number = 7}, (adamic_value){.reference = adamic_graph_hold(map, child)});
 adamic_release(child);
 adamic_array *entries = adamic_map_entries(map, &pair);
 assert(entries->length == 1);
 adamic_object *tuple = entries->elements[0].reference;
 assert(((adamic_cell *)tuple->slots[1].reference)->value.number == 19);
 adamic_object *iterator = adamic_collection_iterator(map, 2, 1, 4, false);
 assert(adamic_graph_is(iterator));
 adamic_release(entries); adamic_release(iterator); adamic_release(map);
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`

func TestGraphMapEntriesAdoptAllObjectBytes(t *testing.T) {
	t.Parallel()
	graphRuntimeGuard(t, graphMapAdoptions, "map.c", "adamic_object_size(pair->count)", "sizeof *tuple + pair->count * sizeof(adamic_value)", "graph adoption truncates object storage")
}
func TestGraphIteratorStateAdoptsAllObjectBytes(t *testing.T) {
	t.Parallel()
	graphRuntimeGuard(t, graphMapAdoptions, "map_set.c", "adamic_object_size(state->shape->count)", "sizeof *state + state->shape->count * sizeof(adamic_value)", "graph adoption truncates object storage")
}
func TestGraphIteratorObjectAdoptsAllObjectBytes(t *testing.T) {
	t.Parallel()
	graphRuntimeGuard(t, graphMapAdoptions, "map_set.c", "adamic_object_size(object->shape->count)", "sizeof *object + object->shape->count * sizeof(adamic_value)", "graph adoption truncates object storage")
}

// Graph canonical entry points preserve the same result ownership as ordinary ones.
func TestGraphCanonicalClosureDiscards(t *testing.T) {
	t.Parallel()
	const source = `#define ADAMIC_CANONICAL_CLOSURES 1
#define ADAMIC_CLOSURE_CONVENTION 1
#define ADAMIC_CLOSURE_DISCARDS 1
#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
static adamic_string a = ADAMIC_STRING("made "), b = ADAMIC_STRING("now");
static adamic_value code(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments;
 return (adamic_value){.reference = adamic_string_concat(2, (adamic_string *const[]){&a, &b})};
}
static adamic_value counted_code(adamic_closure *self, adamic_value *arguments, size_t count) {
 assert(count == 0); return code(self, arguments);
}
int main(void) {
 for (int graph = 0; graph < 2; graph++) {
  for (int counted = 0; counted < 2; counted++) {
   adamic_environment *environment = adamic_environment_new(1);
   if (graph) environment = adamic_graph_adopt(environment, sizeof *environment + sizeof(adamic_cell));
   adamic_cell *cell = &environment->cells[0];
   adamic_closure *closure, *cached;
   if (counted) {
    closure = graph ? adamic_counted_closure_canonical_graph(cell, counted_code, 1, &cell, adamic_discard_reference)
                    : adamic_counted_closure_canonical(cell, counted_code, 1, &cell, adamic_discard_reference);
    cached = graph ? adamic_counted_closure_canonical_graph(cell, counted_code, 1, &cell, adamic_discard_reference)
                   : adamic_counted_closure_canonical(cell, counted_code, 1, &cell, adamic_discard_reference);
   } else {
    closure = graph ? adamic_closure_canonical_graph(cell, code, 1, &cell, adamic_discard_reference)
                    : adamic_closure_canonical(cell, code, 1, &cell, adamic_discard_reference);
    cached = graph ? adamic_closure_canonical_graph(cell, code, 1, &cell, adamic_discard_reference)
                   : adamic_closure_canonical(cell, code, 1, &cell, adamic_discard_reference);
   }
   assert(closure == cached && closure->discard == adamic_discard_reference);
   adamic_closure_discard(closure, adamic_closure_call(closure, NULL, 0));
   adamic_release(cached); adamic_release(closure); adamic_release(environment);
  }
 }
 assert(adamic_counted.allocations == adamic_counted.frees);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "discards")
	if err := Build(source, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	command, release := boundedrun.Command(5*time.Second, binary)
	defer release()
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("canonical discard ownership: %v\n%s", err, output)
	}
}

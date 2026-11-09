package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Adoption makes a region of one (#5nrs3q2): a copy that holds members of a long-lived graph keeps
// the counts its children hold, as counted edges out of its own region, and frees at its last release
// instead of joining the graph until the graph dies. These drive each seam of that rule in the runtime
// alone, judged by the counted build (allocations equal frees, and the region invariant, which panics
// "a counted edge lies inside a graph region" when a merge or free finds the counted edges and the
// members disagree) and by ASan.
//
// Mutants, each red here:
//   - adamic_graph_adopt_owned joins its member children again: "dropped" peaks with the loop count.
//   - adamic_graph_merge skips uncount_edges: "stored" and "closed" panic on the invariant.
//   - adamic_graph_drop returns on a counted edge out of the region: "overwrite" leaks a graph.
//   - adamic_graph_escape retains a counted edge out of the region: "escape" leaks a graph.
//   - release_outside skips a counted edge out of the region: "teardown" leaks a graph.
//   - adamic_array_sort stores before its take again: "sorted" frees a member early (ASan) and
//     leaks the one it wrote over.
const graphRegionOfOneHarness = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *const names[] = {"next", "other", "label"};
static const bool references[] = {true, true, true};
static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference};
static const adamic_shape shape = {3, names, references, NULL, kinds};
static adamic_string prefix = ADAMIC_STRING("node ");
static adamic_string name_a = ADAMIC_STRING("a"), name_b = ADAMIC_STRING("b"), name_c = ADAMIC_STRING("c");
// Each label is built at run time, so a freed node's label reads under ASan.
static adamic_object *node(adamic_string *name) {
 adamic_object *object = adamic_object_new(&shape);
 object = adamic_graph_adopt(object, adamic_object_size(3));
 object->slots[2].reference = adamic_string_concat(2, (adamic_string *const[]){&prefix, name});
 return object;
}
static void link(adamic_object *from, size_t index, void *to) {
 void *old = from->slots[index].reference;
 from->slots[index].reference = adamic_graph_hold(from, to);
 adamic_graph_drop(from, old);
}
// copy is what a planned slice of a graph's array is: made, filled with counted members, adopted.
static adamic_array *copy(adamic_object *a, adamic_object *b) {
 adamic_array *array = adamic_array_make(0, adamic_ir_Union);
 adamic_array_push(array, (adamic_value){.reference = adamic_retain(a)});
 if (b != NULL) { adamic_array_push(array, (adamic_value){.reference = adamic_retain(b)}); }
 return adamic_graph_adopt_owned(array, sizeof *array);
}
static int by_label(adamic_value left, adamic_value right, void *context) {
 (void)context;
 adamic_string *one = ((adamic_object *)left.reference)->slots[2].reference;
 adamic_string *two = ((adamic_object *)right.reference)->slots[2].reference;
 return memcmp(one->bytes, two->bytes, one->length < two->length ? one->length : two->length);
}
static void label(adamic_object *object) {
 adamic_string *text = object->slots[2].reference;
 printf("%.*s\n", (int)text->length, text->bytes);
}

int main(int count, char **arguments) {
 if (count != 3) { return 2; }
 const char *mode = arguments[1];
 long loops = strtol(arguments[2], NULL, 10);
 // a and b are one long-lived graph: each names the other.
 adamic_object *a = node(&name_a), *b = node(&name_b);
 link(a, 0, b); link(b, 0, a);
 if (strcmp(mode, "dropped") == 0) {
  // A copy read and dropped each time round: its region of one frees at its last release.
  size_t seen = 0;
  for (long i = 0; i < loops; i++) {
   adamic_array *held = copy(a, b);
   seen += held->length;
   adamic_release(held);
  }
  printf("seen %zu\n", seen);
 } else if (strcmp(mode, "stored") == 0) {
  // A copy stored into the graph joins it, and the merge takes the counts off its edges inside.
  adamic_array *held = copy(a, b);
  link(a, 1, held);
  adamic_release(held);
  label(((adamic_array *)a->slots[1].reference)->elements[1].reference);
 } else if (strcmp(mode, "closed") == 0) {
  // Language's pin: a dropped copy takes a member of another graph, then a member stores the copy.
  adamic_object *c = node(&name_c);
  adamic_array *held = copy(a, NULL);
  adamic_array_push(held, (adamic_value){.reference = adamic_graph_hold(held, c)});
  link(c, 1, held);
  adamic_release(held);
  adamic_release(c);
  label(b);
 } else if (strcmp(mode, "escape") == 0) {
  // Popping a counted edge out of the region hands its count to the statement.
  adamic_array *held = copy(a, b);
  adamic_object *popped = adamic_graph_escape(held, held->elements[--held->length].reference);
  label(popped);
  adamic_release(popped);
  adamic_release(held);
 } else if (strcmp(mode, "overwrite") == 0) {
  // Writing over a counted edge out of the region releases it.
  adamic_object *c = node(&name_c);
  adamic_array *held = copy(a, NULL);
  adamic_array_set(held, 0, (adamic_value){.reference = adamic_graph_hold(held, c)});
  adamic_release(c);
  label(((adamic_array *)held)->elements[0].reference);
  adamic_release(held);
 } else if (strcmp(mode, "teardown") == 0) {
  // The graph's last outside hold is the copy's counted edge: it lives until the copy goes.
  adamic_array *held = copy(a, NULL);
  adamic_release(a);
  adamic_release(b);
  a = b = NULL;
  label(held->elements[0].reference);
  adamic_release(held);
  return 0;
 } else if (strcmp(mode, "sorted") == 0) {
  // sort writes back through adamic_graph_take, which merges as hold does, and does it before the
  // store, so the merge un-counts the slot's old edge and the take releases the work copy's count.
  adamic_array *held = copy(b, a);
  adamic_array_sort(held, by_label, NULL);
  label(held->elements[0].reference);
  adamic_release(held);
 } else {
  return 3;
 }
 adamic_release(a);
 adamic_release(b);
 return 0;
}
`

var graphRegionOfOnePeak = regexp.MustCompile(`peak (\d+)`)

func TestGraphRegionOfOne(t *testing.T) {
	t.Parallel()
	sanitized := filepath.Join(t.TempDir(), "sanitized")
	if err := native.Build(graphRegionOfOneHarness, sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	counted := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(graphRegionOfOneHarness, counted, native.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	run := func(binary, mode, loops string) leakcheck.Run {
		command := exec.Command(binary, mode, loops)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		exit := 0
		if err != nil {
			exit = -1
			if command.ProcessState != nil {
				exit = command.ProcessState.ExitCode()
			}
		}
		return leakcheck.Run{Stdout: []byte(stdout.String()), Stderr: []byte(stderr.String()), ExitCode: exit}
	}
	for _, mode := range []string{"dropped", "stored", "closed", "escape", "overwrite", "teardown", "sorted"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			for _, binary := range []string{sanitized, counted} {
				result := run(binary, mode, "10")
				if result.ExitCode != 0 {
					t.Fatalf("%s exit %d\n%s%s", filepath.Base(binary), result.ExitCode, result.Stdout, result.Stderr)
				}
			}
			if report := leakcheck.Unbalanced(run(counted, mode, "10")); report != "" {
				t.Fatal(report)
			}
		})
	}
	t.Run("dropped peak", func(t *testing.T) {
		t.Parallel()
		peak := func(loops string) int {
			result := run(counted, "dropped", loops)
			match := graphRegionOfOnePeak.FindSubmatch(result.Stderr)
			if result.ExitCode != 0 || match == nil {
				t.Fatalf("loops %s: exit %d\n%s", loops, result.ExitCode, result.Stderr)
			}
			value, err := strconv.Atoi(string(match[1]))
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		// One copy's worth at most, whatever the loop count. Merged at adoption, each dropped copy
		// was held until the graph died, and the peak grew by one copy a turn (#5nrs3q2).
		if few, many := peak("10"), peak("1000"); many != few {
			t.Fatalf("peak %d at 10 loops, %d at 1000: a dropped copy outlives its loop", few, many)
		}
	})
}

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
//   - adamic_graph_merge unites only its two endpoints (no between): "third", "diamond" and "taken"
//     keep a cycle of counted edges, which the counted build's acyclic check reports.
//   - between stops at the first region it finds that reaches the target: "diamond" leaves x2 out.
//   - between skips no search on a region with no edge in: "search cost" counts N squared for the
//     unrelated stores at N = 1000.
//   - between's post-order pass is the discovery-order fixed point again: "search cost" counts about N
//     squared edge reads for the chain at N = 1000.
//   - between reads the holder member instead of its region (passing it as a target with no
//     region): "member" leaks a graph.
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
// holder makes an adopted array of the given members, each counted, as a spread or literal of them is.
static adamic_array *holder(void *first, void *second) {
 adamic_array *array = adamic_array_make(0, adamic_ir_Union);
 adamic_array_push(array, (adamic_value){.reference = adamic_retain(first)});
 if (second != NULL) { adamic_array_push(array, (adamic_value){.reference = adamic_retain(second)}); }
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
 } else if (strcmp(mode, "third") == 0) {
  // graph_regions_entries' shape: a pair holds a, an array holds the pair, and b stores the array.
  // The store joins b's graph and the array; the pair between them, holding one and held by the
  // other, has to join too, or the three keep a cycle of counted edges.
  adamic_array *pair = copy(a, NULL);
  adamic_array *outer = holder(pair, NULL);
  adamic_release(pair);
  link(b, 1, outer);
  adamic_release(outer);
  label(((adamic_array *)((adamic_array *)b->slots[1].reference)->elements[0].reference)->elements[0].reference);
 } else if (strcmp(mode, "diamond") == 0) {
  // Language's diamond: one store closes two paths at once, through x1 and through x2.
  adamic_array *x1 = copy(a, NULL), *x2 = copy(b, NULL);
  adamic_array *top = holder(x1, x2);
  adamic_release(x1);
  adamic_release(x2);
  link(a, 1, top);
  adamic_release(top);
  label(b);
 } else if (strcmp(mode, "taken") == 0) {
  // take merges as hold does: a moved value closes the same cycle a held one would.
  adamic_array *pair = copy(a, NULL);
  adamic_array *outer = holder(pair, NULL);
  adamic_release(pair);
  void *old = b->slots[1].reference;
  b->slots[1].reference = adamic_graph_take(b, adamic_retain(outer));
  adamic_graph_drop(b, old);
  adamic_release(outer);
  label(a);
 } else if (strcmp(mode, "chain") == 0) {
  // The worst case for one store: a chain of regions, each holding the one before, closed by a
  // single store. The search visits every region in the chain once, and all of them merge.
  adamic_array *chain = copy(a, NULL);
  for (long i = 0; i < loops; i++) {
   adamic_array *next = holder(chain, NULL);
   adamic_release(chain);
   chain = next;
  }
  size_t before = adamic_counted.graph_searched;
  link(b, 1, chain);
  printf("searched %zu\n", adamic_counted.graph_searched - before);
  adamic_release(chain);
 } else if (strcmp(mode, "unrelated") == 0) {
  // The worst case for many stores: the same chain stored, again and again, into fresh nodes it
  // doesn't reach back to. Each search visits the chain and finds nothing to merge.
  adamic_array *chain = copy(a, NULL);
  for (long i = 0; i < loops; i++) {
   adamic_array *next = holder(chain, NULL);
   adamic_release(chain);
   chain = next;
  }
  size_t before = adamic_counted.graph_searched;
  for (long i = 0; i < loops; i++) {
   adamic_object *fresh = node(&name_c);
   link(fresh, 1, chain);
   adamic_release(fresh);
  }
  printf("searched %zu\n", adamic_counted.graph_searched - before);
  adamic_release(chain);
 } else if (strcmp(mode, "member") == 0) {
  // Language's pin: the copy's edge runs to b, the store is into a. a has no counted edge in of its
  // own, but its region does, through b, so the store must merge.
  adamic_array *held = copy(b, NULL);
  link(a, 1, held);
  adamic_release(held);
  label(a);
 } else if (strcmp(mode, "open") == 0) {
  // The open bound: a chain of N regions over c's graph, stored into a, whose region has an edge in
  // (from w) that the chain doesn't reach. The search visits what the chain reaches and merges none
  // of it. #fhcjc66's topological order is what bounds this per store.
  adamic_object *c = node(&name_c);
  adamic_array *w = copy(a, NULL);
  adamic_array *chain = copy(c, NULL);
  adamic_release(c);
  for (long i = 0; i < loops; i++) {
   adamic_array *next = holder(chain, NULL);
   adamic_release(chain);
   chain = next;
  }
  size_t before = adamic_counted.graph_searched;
  link(a, 1, chain);
  printf("searched %zu\n", adamic_counted.graph_searched - before);
  adamic_release(chain);
  adamic_release(w);
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
	for _, mode := range []string{"dropped", "stored", "closed", "escape", "overwrite", "teardown", "sorted", "third", "diamond", "taken", "chain", "unrelated", "member", "open"} {
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
	// The search's cost, as numbers (#5nrs3q2, Adamic's ruling): counted edges read in the counted
	// build, by the walk and by the post-order pass, so a pass that rereads the chain shows. A chain of
	// N regions closed by one store reads at most 3N + 4, each region merged once; the discovery-order
	// fixed point the pass replaced read about N squared there (Language's read), and is the mutant. The
	// chain stored into N fresh nodes reads none, since nothing points into a fresh node's region;
	// without the incoming check it was N squared, the other mutant. The open bound, a store into a
	// region with edges in that the value doesn't reach, reads what the value reaches, and is held to the
	// same 3N + 4 so it can't grow unseen; #fhcjc66 bounds it per store.
	t.Run("search cost", func(t *testing.T) {
		t.Parallel()
		searched := regexp.MustCompile(`searched (\d+)`)
		visits := func(mode string, loops int) int {
			result := run(counted, mode, strconv.Itoa(loops))
			match := searched.FindSubmatch(result.Stdout)
			if result.ExitCode != 0 || match == nil {
				t.Fatalf("%s %d: exit %d\n%s%s", mode, loops, result.ExitCode, result.Stdout, result.Stderr)
			}
			value, err := strconv.Atoi(string(match[1]))
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		for _, loops := range []int{10, 1000} {
			if got := visits("chain", loops); got > 3*loops+4 {
				t.Errorf("chain of %d: %d edges read, want at most %d", loops, got, 3*loops+4)
			} else {
				t.Logf("chain of %d closed by one store: %d edges read", loops, got)
			}
			if got := visits("unrelated", loops); got != 0 {
				t.Errorf("chain of %d stored into %d fresh nodes: %d edges read, want 0", loops, loops, got)
			}
			if got := visits("open", loops); got > 3*loops+4 {
				t.Errorf("open bound, chain of %d: %d edges read, want at most %d", loops, got, 3*loops+4)
			} else {
				t.Logf("open bound, chain of %d: %d edges read", loops, got)
			}
		}
	})
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

// A boundary edge into the Program region stays plain (#5nrs3q2, Language's third condition): a
// Program member is no graph member, so adoption never counts it, and the Program region is freed
// once at its end (rulings.md:330). A copy that holds a Program member, held back by that member,
// is a cycle the Program region's end lets go: clean by the counted build and LeakSanitizer. A copy
// still held past that end releases into freed memory, which is why every boundary edge must go
// first: "late" is the mutant, and ASan reports it.
const graphRegionProgramHarness = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include "graph_regions.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

static const char *const names[] = {"held"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};

int main(int count, char **arguments) {
 if (count != 2) { return 2; }
 adamic_object *program = adamic_object_new(&shape);
 program = adamic_program_adopt_owned(program, adamic_object_size(1));
 adamic_array *copy = adamic_array_make(0, adamic_ir_Union);
 adamic_array_push(copy, (adamic_value){.reference = adamic_retain(program)});
 copy = adamic_graph_adopt_owned(copy, sizeof *copy);
 program->slots[0].reference = adamic_retain(copy);
 if (strcmp(arguments[1], "cycle") == 0) {
  adamic_release(copy);
  puts("cycle");
  adamic_program_region_end();
  assert(adamic_counted.allocations == adamic_counted.frees + adamic_counted.regions);
  return 0;
 }
 adamic_program_region_end();
 adamic_release(copy);
 return 0;
}
`

func TestGraphRegionProgramEdge(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(graphRegionProgramHarness, binary, native.Options{Sanitize: true, Count: true, ProgramRegion: true}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"cycle", "late"} {
		command := exec.Command(binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		switch mode {
		case "cycle":
			if err != nil {
				t.Fatalf("cycle: %v\n%s%s", err, stdout.String(), stderr.String())
			}
			// The counts line is stderr's last; stdout is read apart from it.
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: []byte(stdout.String()), Stderr: []byte(stderr.String())}); report != "" {
				t.Fatalf("cycle: %s\n%s", report, stderr.String())
			}
		case "late":
			if err == nil || !strings.Contains(stderr.String(), "heap-use-after-free") {
				t.Fatalf("a copy released after the Program region's end isn't reported: %v\n%s", err, stderr.String())
			}
		}
	}
}

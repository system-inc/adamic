package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A record can be a graph member (#1vb0x5s, step 1): adamic_record_new_graph makes its table and its
// wrapper in graph storage, with no adoption move (#chncwr5), and merges them at birth, one region,
// every write takes its key and value into the table after the refusal check, and a refused write
// releases them plainly. No emitter calls it yet, so these drive the runtime alone, as
// graph_region_of_one_test.go's harness does, judged by ASan and by the counted build: allocations
// equal frees, and the graph region reports, one line for each region or lone member as its last
// release frees it, with its member count. LeakSanitizer alone isn't trusted here, since it misses
// small cycles.
//
// Mutants, each a rewrite of record.c by exact match, each red on the row it names:
//   - M1, adamic_record_new_graph adopts nothing and returns a plain record: "closed" leaks its cycle.
//   - M2, take_in takes nothing: "closed" leaks, the node held by a count inside its own region.
//   - M3, set_checked takes before its readonly check: "refused-readonly" releases the refused node
//     once too often, and the runtime panics "graph released without ownership".
//   - M4, let_go drops through adamic_graph_drop again: "refused-member-of-another-graph" takes a
//     counted edge off that the refused write never made, and the record's region panics as it frees.
//   - M5, adamic_record_new_graph skips its merge: "regions" finds two regions where one belongs.
//   - M6, adamic_record_new_graph adopts its table and wrapper by move again: "moves" counts two.
const recordGraphHarness = `#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *const names[] = {"held", "label"};
static const bool references[] = {true, true};
static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape shape = {2, names, references, NULL, kinds};

// text is a string made at run time, so every key and label is an allocation the counts see, and a
// freed label reads under ASan.
static adamic_string *text(const char *bytes) {
 size_t length = strlen(bytes);
 adamic_string *made = adamic_string_allocate(length);
 memcpy((char *)made->bytes, bytes, length);
 return made;
}
// node is a graph member with no region yet, adopted as graph_region_of_one_test.go's nodes are.
static adamic_object *node(const char *name) {
 adamic_object *object = adamic_object_new(&shape);
 object = adamic_graph_adopt(object, adamic_object_allocation_size(object));
 object->slots[1].reference = text(name);
 return object;
}
// store writes a node's held slot as the region harness's link does: hold before the slot is written.
static void store(adamic_object *from, void *to) {
 void *old = from->slots[0].reference;
 from->slots[0].reference = adamic_graph_hold(from, to);
 adamic_graph_drop(from, old);
}
static void label(const void *value) {
 const adamic_string *name = ((const adamic_object *)value)->slots[1].reference;
 printf("%.*s\n", (int)name->length, name->bytes);
}
// owned is a write's value: a reference the write consumes.
static adamic_value owned(void *value) {
 return (adamic_value){.reference = adamic_retain(value)};
}
// value_of is what record holds under name, borrowed.
static void *value_of(adamic_record *record, const char *name) {
 adamic_string *key = text(name);
 const adamic_value *slot = adamic_record_get_own(record, key);
 adamic_release(key);
 return slot == NULL ? NULL : slot->reference;
}

int main(int count, char **arguments) {
 if (count != 2) { return 2; }
 const char *mode = arguments[1];
 if (strcmp(mode, "closed") == 0) {
  // The record holds the node under "k" and the node holds the record back: a cycle the region frees.
  adamic_object *n = node("n");
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  adamic_record_define(record, text("k"), owned(n));
  store(n, record);
  label(value_of(record, "k"));
  adamic_release(record);
  adamic_release(n);
 } else if (strcmp(mode, "replace") == 0) {
  // The second set lets go of the first value, which stays a member until the region frees.
  adamic_object *one = node("one"), *two = node("two");
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  adamic_record_set(record, text("k"), owned(one));
  adamic_release(one);
  adamic_record_set(record, text("k"), owned(two));
  adamic_release(two);
  label(value_of(record, "k"));
  adamic_release(record);
 } else if (strcmp(mode, "delete") == 0) {
  adamic_object *n = node("n");
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  adamic_record_define(record, text("k"), owned(n));
  store(n, record);
  adamic_release(n);
  adamic_string *key = text("k");
  (void)adamic_record_delete(record, key);
  adamic_release(key);
  printf("%zu\n", adamic_record_size(record));
  adamic_release(record);
 } else if (strcmp(mode, "refused-readonly") == 0 || strcmp(mode, "refused-fixed") == 0) {
  // "k" is defined read-only and fixed; a set, or a define of another value, is refused, and the
  // refused node is the caller's again, alive until its own last release.
  adamic_object *one = node("one"), *two = node("two");
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  printf("%d\n", adamic_record_define_hidden(record, text("k"), owned(one), adamic_record_references));
  adamic_release(one);
  if (strcmp(mode, "refused-readonly") == 0) {
   printf("%d\n", adamic_record_set_checked(record, text("k"), owned(two)));
  } else {
   printf("%d\n", adamic_record_define_hidden(record, text("k"), owned(two), adamic_record_references));
  }
  label(two);
  label(value_of(record, "k"));
  adamic_release(two);
  adamic_release(record);
 } else if (strcmp(mode, "refused-member-of-another-graph") == 0) {
  // Runtime's condition: the refused value is a member of a second graph, x and y holding each other,
  // that outlives the record and frees once, at its own end.
  adamic_object *one = node("one");
  adamic_object *x = node("x"), *y = node("y");
  store(x, y);
  store(y, x);
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  printf("%d\n", adamic_record_define_hidden(record, text("k"), owned(one), adamic_record_references));
  adamic_release(one);
  printf("%d\n", adamic_record_set_checked(record, text("k"), owned(x)));
  adamic_release(record);
  label(x);
  label(x->slots[0].reference);
  adamic_release(x);
  adamic_release(y);
 } else if (strcmp(mode, "regions") == 0) {
  // The counted build counts each region made and each union. A birth makes two, the table's and the
  // wrapper's, and its merge unites them: what's left is made less merged, since a birth unites two
  // regions and never a region and a lone member.
  size_t made = adamic_counted.graph_regions, merged = adamic_counted.graph_merges;
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  fprintf(stderr, "record graph: regions made %zu merged %zu\n", (size_t)adamic_counted.graph_regions - made, (size_t)adamic_counted.graph_merges - merged);
  adamic_release(record);
 } else if (strcmp(mode, "moves") == 0) {
  // The counted build counts each value graph adoption moves, a copy and a free. Making and freeing a
  // graph record moves none: its table and wrapper are graph storage from birth.
  size_t moved = adamic_counted.graph_moves;
  adamic_record *record = adamic_record_new_graph(adamic_field_reference);
  adamic_release(record);
  fprintf(stderr, "record graph: moved %zu\n", (size_t)adamic_counted.graph_moves - moved);
 } else if (strcmp(mode, "plain") == 0) {
  // A record that's no graph member holds a graph node by a plain count, as before.
  adamic_object *n = node("n");
  adamic_record *record = adamic_record_new(adamic_field_reference);
  adamic_record_define(record, text("k"), owned(n));
  adamic_release(n);
  label(value_of(record, "k"));
  adamic_release(record);
 } else {
  return 3;
 }
 return 0;
}
`

// recordGraphMode is one row: what it prints, and the member count of each graph region report a
// counted run writes, in order, one for each region or lone member its last release frees.
type recordGraphMode struct {
	name, stdout string
	regions      []int
}

var recordGraphModes = []recordGraphMode{
	// The wrapper, the table and the node, one region.
	{"closed", "n\n", []int{3}},
	// The replaced node stays a member until the region frees, with the one that replaced it.
	{"replace", "two\n", []int{4}},
	{"delete", "0\n", []int{3}},
	// The refused node frees alone at the harness's release, before the record's region.
	{"refused-readonly", "1\n0\ntwo\none\n", []int{1, 3}},
	{"refused-fixed", "1\n0\ntwo\none\n", []int{1, 3}},
	// The record's region, then the second graph, once.
	{"refused-member-of-another-graph", "1\n0\nx\ny\n", []int{3, 2}},
	// The wrapper and the table, one region.
	{"regions", "", []int{2}},
	// The wrapper and the table, one region, and neither moved.
	{"moves", "", []int{2}},
	// The node alone, freed by the plain table's release.
	{"plain", "n\n", []int{1}},
}

var (
	recordGraphCounts = regexp.MustCompile(`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains \d+ releases \d+ peak \d+ regions (\d+)\n\z`)
	recordGraphReport = regexp.MustCompile(`(?m)^adamic: graph region: live (\d+) `)
	recordGraphBirth  = regexp.MustCompile(`(?m)^record graph: regions made (\d+) merged (\d+)$`)
	recordGraphMoves  = regexp.MustCompile(`(?m)^record graph: moved (\d+)$`)
)

type recordGraphRun struct {
	stdout, stderr string
	exit           int
}

func recordGraphExecute(t *testing.T, binary, mode string) recordGraphRun {
	t.Helper()
	command := exec.Command(binary, mode)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	// A binary that never ran says nothing about its row, and must not read as a red one.
	if command.ProcessState == nil {
		t.Fatalf("%s %s didn't run: %v", filepath.Base(binary), mode, err)
	}
	return recordGraphRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

// recordGraphRed says why a run of mode is red, or "" when it's green. A counted run is also held to
// its counts and its region reports.
func recordGraphRed(mode recordGraphMode, counted bool, run recordGraphRun) string {
	if run.exit != 0 {
		return fmt.Sprintf("exit %d", run.exit)
	}
	if run.stdout != mode.stdout {
		return fmt.Sprintf("stdout %q, want %q", run.stdout, mode.stdout)
	}
	if !counted {
		return ""
	}
	counts := recordGraphCounts.FindStringSubmatch(run.stderr)
	if counts == nil {
		return "no counts line"
	}
	var values [3]int
	for index, group := range []int{1, 2, 3} {
		value, err := strconv.Atoi(counts[group])
		if err != nil {
			return fmt.Sprintf("unreadable counts: %v", err)
		}
		values[index] = value
	}
	if allocations, frees, regions := values[0], values[1], values[2]; allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d, in regions %d)", allocations-frees-regions, allocations, frees, regions)
	}
	var died []int
	for _, report := range recordGraphReport.FindAllStringSubmatch(run.stderr, -1) {
		members, err := strconv.Atoi(report[1])
		if err != nil {
			return fmt.Sprintf("unreadable region report: %v", err)
		}
		died = append(died, members)
	}
	if !slices.Equal(died, mode.regions) {
		return fmt.Sprintf("graph regions freed with %v members, want %v", died, mode.regions)
	}
	if mode.name == "regions" {
		birth := recordGraphBirth.FindStringSubmatch(run.stderr)
		if birth == nil {
			return "no birth line"
		}
		made, err := strconv.Atoi(birth[1])
		if err != nil {
			return fmt.Sprintf("unreadable birth line: %v", err)
		}
		merged, err := strconv.Atoi(birth[2])
		if err != nil {
			return fmt.Sprintf("unreadable birth line: %v", err)
		}
		if made-merged != 1 {
			return fmt.Sprintf("a graph record's birth left %d regions (made %d, merged %d), want 1", made-merged, made, merged)
		}
	}
	if mode.name == "moves" {
		moves := recordGraphMoves.FindStringSubmatch(run.stderr)
		if moves == nil {
			return "no moves line"
		}
		moved, err := strconv.Atoi(moves[1])
		if err != nil {
			return fmt.Sprintf("unreadable moves line: %v", err)
		}
		if moved != 0 {
			return fmt.Sprintf("making and freeing a graph record moved %d values into graph storage, want 0", moved)
		}
	}
	return ""
}

func TestRecordGraph(t *testing.T) {
	t.Parallel()
	sanitized := filepath.Join(t.TempDir(), "sanitized")
	if err := Build(recordGraphHarness, sanitized, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	counted := filepath.Join(t.TempDir(), "counted")
	if err := Build(recordGraphHarness, counted, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range recordGraphModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			for _, binary := range []struct {
				path    string
				counted bool
			}{{sanitized, false}, {counted, true}} {
				run := recordGraphExecute(t, binary.path, mode.name)
				if red := recordGraphRed(mode, binary.counted, run); red != "" {
					t.Fatalf("%s: %s\nstdout:\n%s\nstderr:\n%s", filepath.Base(binary.path), red, run.stdout, run.stderr)
				}
			}
		})
	}
}

// recordGraphMutations are the mutants named at the top of this file: before, in record.c, becomes
// after, and the row reds names goes red.
var recordGraphMutations = []struct{ name, before, after, reds string }{
	// M1 reds closed.
	{"M1-no-adoption", "\tadamic_record *record = adamic_object_new_graph(&record_reference_shape);\n\tadamic_map *map = adamic_map_new_graph(true, true);\n\tadamic_graph_born(map);\n\trecord->slots[0].reference = map;\n\tadamic_graph_born(record);\n\tadamic_graph_merge(record, map);\n", "\tadamic_record *record = adamic_object_new(&record_reference_shape);\n\tadamic_map *map = adamic_map_new(true, true);\n\trecord->slots[0].reference = map;\n", "closed"},
	// M2 reds closed.
	{"M2-no-take", "\t\tadamic_graph_take(map, key);\n\t\tif (map->reference_values) { adamic_graph_take(map, value.reference); }\n", "", "closed"},
	// M3 reds refused-readonly.
	{"M3-take-before-readonly-check", "\tif (slot != NULL && entry_of(slot)->readonly) {\n\t\tlet_go(map, key, value);\n\t\treturn false;\n\t}\n\ttake_in(map, key, value);\n", "\ttake_in(map, key, value);\n\tif (slot != NULL && entry_of(slot)->readonly) {\n\t\tlet_go(map, key, value);\n\t\treturn false;\n\t}\n", "refused-readonly"},
	// M4 reds refused-member-of-another-graph.
	{"M4-let-go-drops", "\tadamic_release(key);\n\tif (map->reference_values) {\n\t\tadamic_release(value.reference);\n\t}\n", "\tif (adamic_graph_is(map)) { adamic_graph_drop(map, key); } else { adamic_release(key); }\n\tif (map->reference_values) {\n\t\tif (adamic_graph_is(map)) { adamic_graph_drop(map, value.reference); } else { adamic_release(value.reference); }\n\t}\n", "refused-member-of-another-graph"},
	// M5 reds regions.
	{"M5-no-merge", "\tadamic_graph_merge(record, map);\n", "", "regions"},
	// M6 reds moves.
	{"M6-adopted-by-move", "\tadamic_record *record = adamic_object_new_graph(&record_reference_shape);\n\tadamic_map *map = adamic_map_new_graph(true, true);\n\tadamic_graph_born(map);\n\trecord->slots[0].reference = map;\n\tadamic_graph_born(record);\n", "\tadamic_record *record = adamic_object_new(&record_reference_shape);\n\tadamic_map *map = adamic_map_new(true, true);\n\tmap = adamic_graph_adopt_owned(map, sizeof *map);\n\trecord->slots[0].reference = map;\n\trecord = adamic_graph_adopt_owned(record, adamic_object_allocation_size(record));\n", "moves"},
}

// recordGraphMutants is the record-graph-mutants product: the harness linked against each mutant's
// runtime, a binary named for the mutant, counted and sanitized. It's built as recordMutants builds
// record_test.go's: the runtime's unchanged translation units compile once, each mutant compiles only
// its record.c, and every object links directly.
func recordGraphMutants(t *testing.T) string {
	t.Helper()
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true, Count: true})
	inputs := buildcache.Inputs{
		Name:      "record-graph-mutants",
		Files:     []string{"internal/native_runtime/c", "internal/native/record_graph_test.go", "internal/native/split_units_test.go", "internal/native/runtime_mutants_test.go", "internal/native/native.go", "internal/native_runtime/options.go", "internal/native_runtime/library.go"},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.CompilerReport("clang")},
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		files, err := native_runtime.ReadRuntime(native_runtime.Sources, "c")
		if err != nil {
			return err
		}
		var sources []string
		for _, file := range files {
			if strings.HasSuffix(file.Name, ".c") {
				sources = append(sources, file.Name)
			}
		}
		scratch, err := os.MkdirTemp("", "adamic-record-graph-mutants-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		compile := func(within, source string) error {
			arguments := append(append(slices.Clone(flags), native_runtime.PrefixMaps(within, "/adamic-runtime")...), "-c", source, "-o", strings.TrimSuffix(source, ".c")+".o")
			if output, err := runtimeMutantClang(within, arguments...).CombinedOutput(); err != nil {
				return fmt.Errorf("compiling runtime %s: %w\n%s", source, err, output)
			}
			return nil
		}
		write := func(within string, files []native_runtime.RuntimeFile) error {
			if err := os.MkdirAll(within, 0o755); err != nil {
				return err
			}
			for _, file := range files {
				if err := os.WriteFile(filepath.Join(within, file.Name), file.Contents, 0o644); err != nil {
					return err
				}
			}
			return nil
		}
		unchanged := filepath.Join(scratch, "runtime")
		if err := write(unchanged, files); err != nil {
			return err
		}
		if err := eachInParallel(len(sources), func(index int) error { return compile(unchanged, sources[index]) }); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(recordGraphHarness), 0o644); err != nil {
			return err
		}
		return eachInParallel(len(recordGraphMutations), func(index int) error {
			mutant := recordGraphMutations[index]
			// The mutant's directory holds the runtime's headers and its record.c, so that file compiles as
			// it would among the rest.
			var changed []native_runtime.RuntimeFile
			for _, file := range files {
				if file.Name == "record.c" {
					contents := string(file.Contents)
					if found := strings.Count(contents, mutant.before); found != 1 {
						return fmt.Errorf("mutant %s: its text is in record.c %d times, want once", mutant.name, found)
					}
					changed = append(changed, native_runtime.RuntimeFile{Name: file.Name, Contents: []byte(strings.Replace(contents, mutant.before, mutant.after, 1))})
				} else if strings.HasSuffix(file.Name, ".h") {
					changed = append(changed, file)
				}
			}
			within := filepath.Join(scratch, "mutant-"+strconv.Itoa(index))
			if err := write(within, changed); err != nil {
				return err
			}
			if err := compile(within, "record.c"); err != nil {
				return fmt.Errorf("mutant %s: %w", mutant.name, err)
			}
			arguments := append(append(slices.Clone(flags), native_runtime.PrefixMaps(scratch, "/adamic-build")...), "-I", "runtime", "-o", filepath.Join(directory, mutant.name), "main.c")
			for _, source := range sources {
				object := strings.TrimSuffix(source, ".c") + ".o"
				if source == "record.c" {
					arguments = append(arguments, filepath.Join(filepath.Base(within), object))
				} else {
					arguments = append(arguments, filepath.Join("runtime", object))
				}
			}
			arguments = append(arguments, "-lm")
			if output, err := runtimeMutantClang(scratch, arguments...).CombinedOutput(); err != nil {
				return fmt.Errorf("mutant %s must compile: %w\n%s", mutant.name, err, output)
			}
			return nil
		})
	})
}

// TestProduct_RecordGraphMutants builds every record graph mutant, so the tree build builds them ahead.
func TestProduct_RecordGraphMutants(t *testing.T) {
	t.Parallel()
	recordGraphMutants(t)
}

func TestRecordGraphMutants(t *testing.T) {
	t.Parallel()
	source, err := native_runtime.Sources.ReadFile("c/record.c")
	if err != nil {
		t.Fatal(err)
	}
	product := recordGraphMutants(t)
	for _, mutant := range recordGraphMutations {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			if found := strings.Count(string(source), mutant.before); found != 1 {
				t.Fatalf("the mutant's text is in record.c %d times, want once", found)
			}
			index := slices.IndexFunc(recordGraphModes, func(mode recordGraphMode) bool { return mode.name == mutant.reds })
			if index < 0 {
				t.Fatalf("no row named %s", mutant.reds)
			}
			mode := recordGraphModes[index]
			run := recordGraphExecute(t, filepath.Join(product, mutant.name), mode.name)
			red := recordGraphRed(mode, true, run)
			if red == "" {
				t.Fatalf("%s left %s green", mutant.name, mode.name)
			}
			t.Logf("%s reds %s: %s", mutant.name, mode.name, red)
		})
	}
}

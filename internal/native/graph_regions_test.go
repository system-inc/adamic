package native_test

import (
	"fmt"
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

// The runtime is exercised independently of graph-type selection. A successful
// runtime test does not authorize removing a cycle refusal from the compiler.
const graphHarness = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

static const char *const names[] = {"next", "other", "label"};
static const bool references[] = {true, true, true};
static const adamic_field_kind shape_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference};
static const adamic_shape shape = {3, names, references, NULL, shape_kinds};
static adamic_object *node(void) {
 adamic_object *object = adamic_object_new(&shape);
 return adamic_graph_adopt(object, adamic_object_size(3));
}
static void link(adamic_object *from, size_t index, adamic_object *to) {
 void *old = from->slots[index].reference;
 from->slots[index].reference = adamic_graph_hold(from, to);
 adamic_graph_drop(from, old);
}
__attribute__((noinline)) static int run(int count, char **arguments) {
 (void)count;
 printf("node_bytes %zu\n", adamic_object_size(3));
 fflush(stdout);
 adamic_object *root = node(), *child = node(), *orphan = node();
 static adamic_string a = ADAMIC_STRING("made ");
 static adamic_string b = ADAMIC_STRING("at runtime");
 root->slots[2].reference = adamic_string_concat(2, (adamic_string *const[]){&a, &b});
 link(root, 0, child); link(child, 0, root);
 link(root, 1, orphan); link(root, 1, NULL);
 adamic_weak *weak = adamic_weak_of(orphan);
 adamic_release(orphan);
 if (strcmp(arguments[1], "shared") == 0) {
  adamic_graph_mark_shared(root);
  adamic_object *other = node();
  adamic_graph_merge(root, other);
  return 2;
 }
 if (strcmp(arguments[1], "escaping") == 0) {
  adamic_release(root);
  root = NULL;
  adamic_object *parent = child->slots[0].reference;
  adamic_string *label = parent->slots[2].reference;
  printf("%.*s\n", (int)label->length, label->bytes);
  if (adamic_weak_target(weak) != orphan) { return 3; }
  adamic_release(child);
 } else {
  adamic_release(child);
  if (strcmp(arguments[1], "leak") != 0) { adamic_release(root); }
 }
 if (strcmp(arguments[1], "later") == 0) {
  printf("%zu\n", root->shape->count);
 }
 if (strcmp(arguments[1], "leak") != 0 && adamic_weak_target(weak) != NULL) { return 4; }
 adamic_release(weak);
 return 0;
}
int main(int count, char **arguments) { return run(count, arguments); }
`

func TestGraphRegionsRuntime(test *testing.T) {
	test.Parallel()
	for _, counted := range []bool{false, true} {
		binary := filepath.Join(test.TempDir(), "regions")
		if buildError := native.Build(
			graphHarness,
			binary,
			native.Options{Sanitize: true, Count: counted},
		); buildError != nil {
			test.Fatal(buildError)
		}
		for _, mode := range []string{"escaping", "anchor", "later", "leak", "shared"} {
			command := exec.Command(binary, mode)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
			output, combinedOutputError := command.CombinedOutput()
			text := string(output)
			switch mode {
			case "anchor", "escaping":
				if report := leakcheck.Report(test, graphHarness, binary, mode); report != "" {
					test.Fatalf("%s leaked: %s", mode, report)
				}
				if combinedOutputError != nil {
					test.Fatalf("%s counted=%t: %v\n%s", mode, counted, combinedOutputError, text)
				}
				if mode == "escaping" && !strings.Contains(text, "made at runtime\n") {
					test.Fatal(text)
				}
				if counted {
					if !strings.Contains(text, fmt.Sprintf(
						"live 3 bytes %d reachable 2 bytes %d unreachable 1 bytes %d",
						3*graphMeasuredSize(test, output, "node_bytes"),
						2*graphMeasuredSize(test, output, "node_bytes"),
						graphMeasuredSize(test, output, "node_bytes"),
					)) {
						test.Fatal(text)
					}
					if !strings.Contains(text, "graph counts: regions 1 merges 2") {
						test.Fatal(text)
					}
					if !strings.Contains(text, "allocations 5 frees 5 retains 0 releases 4 peak 5 regions 0") {
						test.Fatal(text)
					}
				} else if strings.Contains(text, "graph region:") {
					test.Fatal("diagnostic tracing in release build")
				}
			case "later":
				if combinedOutputError == nil ||
					!strings.Contains(text, "ERROR: AddressSanitizer: heap-use-after-free") {
					test.Fatalf("later read not caught: %v\n%s", combinedOutputError, text)
				}
			case "leak":
				if combinedOutputError != nil {
					test.Fatalf("anchor mutant failed before leak check: %v\n%s", combinedOutputError, text)
				}
				report := leakcheck.Report(test, graphHarness, binary, mode)
				if report == "" {
					// LSan can conservatively find a stale pointer after run returns. The
					// ownership ledger must still catch the four deliberately held values.
					countedBinary := binary
					if !counted {
						countedBinary = filepath.Join(test.TempDir(), "anchor-counted")
						if buildError := native.Build(
							graphHarness,
							countedBinary,
							native.Options{Sanitize: true, Count: true},
						); buildError != nil {
							test.Fatal(buildError)
						}
					}
					command := exec.Command(countedBinary, mode)
					command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
					output, combinedOutputError := command.CombinedOutput()
					if combinedOutputError != nil {
						test.Fatalf(
							"anchor mutant failed before counted leak check: %v\n%s",
							combinedOutputError,
							output,
						)
					}
					report = leakcheck.Unbalanced(leakcheck.Run{Stderr: output})
					if !strings.Contains(report, "heap values leaked: 4") {
						test.Fatalf("unreleased anchor not caught by counted leak check: %s\n%s", report, output)
					}
				}
				test.Logf("unreleased anchor counted=%t: %s", counted, report)
			case "shared":
				if combinedOutputError == nil ||
					!strings.Contains(text, "merging shared graph regions is not yet supported") {
					test.Fatalf("shared merge not rejected: %v\n%s", combinedOutputError, text)
				}
			}
		}
	}
}

// Interior cells count their owning environment, including when that environment
// is a graph member. Closures and environments then form plain internal edges.
func TestGraphClosureEnvironment(test *testing.T) {
	test.Parallel()
	const harness = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include "graph_regions.h"
static adamic_value code(adamic_closure *self, adamic_value *arguments) {
 (void)arguments;
 return (adamic_value){.number = (double)self->count};
}
int main(void) {
 adamic_environment *environment = adamic_environment_new(1);
 environment = adamic_graph_adopt(environment, sizeof *environment + sizeof(adamic_cell));
 adamic_closure *closure = adamic_closure_new(code, 1);
 closure = adamic_graph_adopt(closure, sizeof *closure + sizeof(adamic_cell *));
 adamic_cell *cell = &environment->cells[0];
 cell->references = true; cell->ready = true;
 cell->value.reference = adamic_graph_hold(cell, closure);
 closure->cells[0] = adamic_graph_hold(closure, cell);
 adamic_retain(cell);
 adamic_release(environment); adamic_release(closure);
 if (((adamic_closure *)cell->value.reference)->count != 1) { return 1; }
 adamic_release(cell);
 return 0;
}
`
	binary := filepath.Join(test.TempDir(), "closure")
	if buildError := native.Build(harness, binary, native.Options{Sanitize: true, Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("%v\n%s", combinedOutputError, output)
	}
	if report := leakcheck.Report(test, harness, binary); report != "" {
		test.Fatal(report)
	}
	if !strings.Contains(string(output), "graph counts: regions 1 merges 1") {
		test.Fatal(string(output))
	}
	if !strings.Contains(string(output), "allocations 2 frees 2 retains 1 releases 3 peak 2 regions 0") {
		test.Fatal(string(output))
	}
	test.Logf("closure/environment counts:\n%s", output)
}

// This standalone runtime workload measures regions independently of the
// language fixtures and compiler allocation classification.
// Not parallel: two large resident sets are measured sequentially.
func TestGraphRegionsMillion(test *testing.T) {
	const cmain = `#include <sys/resource.h>
int main(void) {
 const size_t count = 1000000;
 adamic_object *root = node(), *previous = root;
 size_t reachable = 1;
 for (size_t i = 1; i < count; i++) {
  adamic_object *current = node();
  if (i % 20 == 0) {
   link(previous, 1, current);
   link(previous, 1, root);
   adamic_release(current);
  } else {
   link(previous, 0, current);
   link(current, 1, i % 7 == 0 ? previous : root);
   if (previous != root) { adamic_release(previous); }
   previous = current;
   reachable++;
  }
 }
 if (previous != root) { adamic_release(previous); }
 printf("members %zu reachable %zu\n", count, reachable);
 printf("node_bytes %zu\n", adamic_object_size(2));
 struct rusage usage;
 getrusage(RUSAGE_SELF, &usage);
 printf("rss_kib %ld\n", usage.ru_maxrss);
 adamic_release(root);
 return 0;
}
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + cmain
	source = strings.ReplaceAll(source, "shape = {3,", "shape = {2,")
	source = strings.ReplaceAll(source, "adamic_object_size(3)", "adamic_object_size(2)")
	directory := test.TempDir()
	binary := filepath.Join(directory, "million")
	if buildError := native.Build(source, binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(
		"/bin/sh",
		"-c",
		`"$@" & child=$!; wait "$child"`,
		"rss",
		binary,
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("native: %v\n%s", combinedOutputError, output)
	}
	size := regexp.MustCompile(`(?m)^node_bytes (\d+)$`).FindSubmatch(output)
	if size == nil {
		test.Fatal("missing node allocation size")
	}
	nodeBytes, parseError := strconv.ParseUint(string(size[1]), 10, 64)
	if parseError != nil || nodeBytes == 0 {
		test.Fatalf("invalid node allocation size: %s", size[1])
	}
	want := fmt.Sprintf(
		"live 1000000 bytes %d reachable 950001 bytes %d unreachable 49999 bytes %d metadata 16000064",
		1000000*nodeBytes,
		950001*nodeBytes,
		49999*nodeBytes,
	)
	if !strings.Contains(string(output), want) {
		test.Fatal(string(output))
	}
	test.Logf("native runtime foundation:\n%s", output)
	release := filepath.Join(directory, "million-release")
	if buildError := native.Build(source, release, native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	releaseOutput, combinedOutputError := exec.Command(
		"/bin/sh",
		"-c",
		`"$@" & child=$!; wait "$child"`,
		"rss",
		release,
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("release: %v\n%s", combinedOutputError, releaseOutput)
	}
	if strings.Contains(string(releaseOutput), "graph region:") {
		test.Fatal("diagnostics in release")
	}
	test.Logf("native release flags: %s", strings.Join(native.Flags(native.Options{}), " "))
	test.Logf("native release:\n%s", releaseOutput)
	sanitized := filepath.Join(directory, "million-sanitized")
	if buildError := native.Build(source, sanitized, native.Options{Count: true, Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	sanitizedOutput, combinedOutputError := exec.Command(sanitized).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("million-member sanitizer/leak check: %v\n%s", combinedOutputError, sanitizedOutput)
	}
	if !strings.Contains(string(sanitizedOutput), "allocations 1000000 frees 1000000") {
		test.Fatal(string(sanitizedOutput))
	}
	if report := leakcheck.Report(test, source, sanitized); report != "" {
		test.Fatal(report)
	}
	test.Log("million-member ASan, UBSan and shared leak check: clean")
	const js = `const count = 1000000;
const node = () => ({ next: undefined, other: undefined, label: undefined });
let root = node(), previous = root, reachable = 1;
for (let i = 1; i < count; i++) {
 const current = node();
 if (i % 20 === 0) {
  previous.other = current;
  previous.other = root;
 } else {
  previous.next = current;
  current.other = i % 7 === 0 ? previous : root;
  previous = current;
  reachable++;
 }
}
console.log('members ' + count + ' reachable ' + reachable);
console.log('heap_used_bytes ' + process.memoryUsage().heapUsed);
console.log('rss_kib ' + process.resourceUsage().maxRSS);
root = previous = undefined;
`
	path := filepath.Join(directory, "million.js")
	if writeError := os.WriteFile(path, []byte(js), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	output, combinedOutputError = exec.Command(
		"/bin/sh",
		"-c",
		`"$@" & child=$!; wait "$child"`,
		"rss",
		"node",
		path,
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v\n%s", combinedOutputError, output)
	}
	if !strings.Contains(string(output), "members 1000000 reachable 950001") {
		test.Fatal(string(output))
	}
	test.Logf("Node:\n%s", output)
}

func TestGraphContainerBoundary(test *testing.T) {
	test.Parallel()
	const main = `int main(void) {
 adamic_object *root = node(), *child = node();
 adamic_array *array = adamic_array_make(1, adamic_ir_Union);
 array = adamic_graph_adopt(array, sizeof *array);
 adamic_map *cache = adamic_map_new(true, true);
 cache = adamic_graph_adopt(cache, sizeof *cache);
 static adamic_string a = ADAMIC_STRING("cache ");
 static adamic_string b = ADAMIC_STRING("key");
 adamic_string *key = adamic_string_concat(2, (adamic_string *const[]){&a, &b});
 adamic_array_push(array, (adamic_value){.reference = adamic_graph_hold(array, child)});
 root->slots[0].reference = adamic_graph_hold(root, array);
 link(child, 0, root);
 adamic_map_set(cache, (adamic_value){.reference = key}, (adamic_value){.reference = adamic_graph_hold(cache, root)});
 printf("node_bytes %zu\narray_bytes %zu\nmap_bytes %zu\n", adamic_object_size(3), sizeof *array + array->capacity * sizeof(adamic_value), sizeof *cache + cache->capacity * sizeof(adamic_map_entry) + cache->bucket_count * sizeof(size_t));
 adamic_release(root); adamic_release(child); adamic_release(array);
 adamic_object *kept = adamic_map_get(cache, (adamic_value){.reference = key})->reference;
 if (((adamic_array *)kept->slots[0].reference)->length != 1) { return 1; }
 // A counted object is another outside keeper of the graph.
 adamic_object *holder = adamic_object_new(&shape);
 holder->slots[0].reference = adamic_graph_hold(holder, kept);
 adamic_release(cache);
 if (((adamic_object *)holder->slots[0].reference)->shape->count != 3) { return 2; }
 adamic_release(holder);
 return 0;
}
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + main
	binary := filepath.Join(test.TempDir(), "containers")
	if buildError := native.Build(source, binary, native.Options{Count: true, Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("%v\n%s", combinedOutputError, output)
	}
	if report := leakcheck.Report(test, source, binary); report != "" {
		test.Fatal(report)
	}
	if !strings.Contains(string(output), "graph counts: regions 1 merges 3") {
		test.Fatal(string(output))
	}
	if !strings.Contains(string(output), fmt.Sprintf(
		"live 4 bytes %d reachable 3 bytes %d unreachable 1 bytes %d",
		2*graphMeasuredSize(test, output, "node_bytes")+
			graphMeasuredSize(test, output, "array_bytes")+
			graphMeasuredSize(test, output, "map_bytes"),
		2*graphMeasuredSize(test, output, "node_bytes")+graphMeasuredSize(test, output, "array_bytes"),
		graphMeasuredSize(test, output, "map_bytes"),
	)) {
		test.Fatal(string(output))
	}
	if !strings.Contains(string(output), "allocations 6 frees 6 retains 1 releases 5 peak 6 regions 0") {
		test.Fatal(string(output))
	}
	test.Logf("container/boundary counts:\n%s", output)
}

func TestGraphLazyRegions(test *testing.T) {
	test.Parallel()
	const main = `int main(void) {
  printf("node_bytes %zu\n", adamic_object_size(3));
  if (sizeof(adamic_graph_header) != 16) { return 1; }
  adamic_object *a = node(), *b = node(), *c = node(), *d = node();
  if (adamic_graph_header_of(a)->region != NULL) { return 2; }
  adamic_retain(a); adamic_release(a);
  if (adamic_graph_header_of(a)->region != NULL || a->heap.references != 1) { return 3; }
  link(a, 0, b); link(c, 0, d);
  // Merge two established regions, each with two outside owners.
  link(b, 0, c); link(d, 0, a);
  adamic_release(b); adamic_release(c); adamic_release(d);
  if (((adamic_object *)((adamic_object *)a->slots[0].reference)->slots[0].reference)->shape->count != 3) { return 4; }
  adamic_release(a);
  adamic_object *lone = node();
  if (adamic_graph_header_of(lone)->region != NULL) { return 5; }
  adamic_release(lone);
  return 0;
 }
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + main
	binary := filepath.Join(test.TempDir(), "lazy")
	if buildError := native.Build(source, binary, native.Options{Sanitize: true, Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("%v\n%s", combinedOutputError, output)
	}
	if report := leakcheck.Report(test, source, binary); report != "" {
		test.Fatal(report)
	}
	if !strings.Contains(string(output), "graph counts: regions 2 merges 3") ||
		!strings.Contains(string(output), "allocations 5 frees 5") {
		test.Fatal(string(output))
	}
	if !strings.Contains(string(output), fmt.Sprintf(
		"live 1 bytes %d reachable 1 bytes %d unreachable 0 bytes 0 metadata 16",
		graphMeasuredSize(test, output, "node_bytes"),
		graphMeasuredSize(test, output, "node_bytes"),
	)) {
		test.Fatal(string(output))
	}
	test.Logf("lazy regions:\n%s", output)
}

// The same count predicate used on darwin must catch this mutant on every host,
// even when the allocator's chunks keep leaked values reachable to malloc tools.
func TestGraphUnreleasedAnchorCounted(test *testing.T) {
	test.Parallel()
	binary := filepath.Join(test.TempDir(), "anchor-counted")
	if buildError := native.Build(graphHarness, binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for _, mode := range []string{"anchor", "leak"} {
		output, combinedOutputError := exec.Command(binary, mode).CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("%s: %v\n%s", mode, combinedOutputError, output)
		}
		report := leakcheck.Unbalanced(leakcheck.Run{Stderr: output})
		if mode == "anchor" && report != "" {
			test.Fatal(report)
		}
		if mode == "leak" && !strings.Contains(report, "heap values leaked: 4") {
			test.Fatalf("unreleased anchor bypassed counted check: %s\n%s", report, output)
		}
		test.Logf("%s count predicate: %q", mode, report)
	}
}

func graphMeasuredSize(test *testing.T, output []byte, name string) uint64 {
	test.Helper()
	match := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + ` (\d+)$`).FindSubmatch(output)
	if match == nil {
		test.Fatalf("missing %s: %s", name, output)
	}
	size, parseError := strconv.ParseUint(string(match[1]), 10, 64)
	if parseError != nil || size == 0 {
		test.Fatalf("invalid %s: %s", name, match[1])
	}
	return size
}

// Adoption moves an allocation and frees the original, so it takes only the kinds it can re-aim
// (graph_regions.c's unadoptable) and refuses a value something still names by address. Each refusal
// is a compiler bug if a program reaches it, so it stops by name. The control moves a Weak's target
// without the check and shows the read that the check prevents.
func TestGraphAdoptionRefusals(test *testing.T) {
	test.Parallel()
	const main = `static adamic_value code(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments;
 return (adamic_value){.number = 1};
}
int main(int count, char **arguments) {
 (void)count;
 if (strcmp(arguments[1], "string") == 0) {
  static adamic_string a = ADAMIC_STRING("made ");
  static adamic_string b = ADAMIC_STRING("at runtime");
  adamic_string *made = adamic_string_concat(2, (adamic_string *const[]){&a, &b});
  adamic_graph_adopt(made, sizeof *made);
  return 1;
 }
 if (strcmp(arguments[1], "number") == 0) {
  adamic_heap *box = adamic_box_number(7);
  adamic_graph_adopt(box, sizeof *box);
  return 1;
 }
 if (strcmp(arguments[1], "canonical") == 0) {
  adamic_environment *environment = adamic_environment_new(1);
  adamic_closure *closure = adamic_closure_canonical(&environment->cells[0], code, 0, NULL);
  adamic_graph_adopt(closure, sizeof *closure);
  // Without the refusal, the cache still names the freed closure, and this lookup reads it.
  adamic_closure_canonical(&environment->cells[0], code, 0, NULL);
  return 1;
 }
 adamic_object *object = adamic_object_new(&shape);
 adamic_weak *weak = adamic_weak_of(object);
 if (strcmp(arguments[1], "weak") == 0) {
  adamic_graph_adopt(object, adamic_object_size(3));
  // Without the refusal, the Weak still names the freed original, and this read reaches it.
  printf("%zu\n", ((adamic_object *)adamic_weak_target(weak))->shape->count);
  return 1;
 }
 // control: the move the check refuses, made directly.
 object = adamic_heap_graph_storage(object, adamic_object_size(3));
 adamic_object *named = adamic_weak_target(weak);
 printf("%zu\n", named->shape->count);
 return 0;
}
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + main
	binary := filepath.Join(test.TempDir(), "refusals")
	if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for mode, want := range map[string]string{
		"string":    "compiler bug: a string can't be adopted into a graph region",
		"number":    "compiler bug: a number box can't be adopted into a graph region",
		"canonical": "compiler bug: a canonical closure can't be adopted into a graph region after it's cached",
		"weak":      "compiler bug: a value can't be adopted into a graph region while a Weak names it",
		"control":   "ERROR: AddressSanitizer: heap-use-after-free",
	} {
		command := exec.Command(binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError == nil || !strings.Contains(string(output), want) {
			test.Fatalf("%s: want %q: %v\n%s", mode, want, combinedOutputError, output)
		}
		if mode != "control" && strings.Contains(string(output), "Sanitizer") {
			test.Fatalf("%s stopped by a sanitizer, not by name:\n%s", mode, output)
		}
		test.Logf("%s: %s", mode, want)
	}
}

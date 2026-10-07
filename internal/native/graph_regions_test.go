package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The runtime is exercised independently of graph-type selection. A successful
// runtime test does not authorize removing a cycle refusal from the compiler.
const graphHarness = `#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

static const char *const names[] = {"next", "other", "label"};
static const bool references[] = {true, true, true};
static const adamic_shape shape = {3, names, references, NULL};
static adamic_object *node(void) {
 adamic_object *object = adamic_object_new(&shape);
 return adamic_graph_adopt(object, sizeof *object + 3 * sizeof(adamic_value));
}
static void link(adamic_object *from, size_t index, adamic_object *to) {
 void *old = from->slots[index].reference;
 from->slots[index].reference = adamic_graph_hold(from, to);
 adamic_graph_drop(from, old);
}
__attribute__((noinline)) static int run(int count, char **arguments) {
 (void)count;
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

func TestGraphRegionsRuntime(t *testing.T) {
	t.Parallel()
	for _, counted := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "regions")
		if err := Build(graphHarness, binary, Options{Sanitize: true, Count: counted}); err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"escaping", "anchor", "later", "leak", "shared"} {
			output, err := exec.Command(binary, mode).CombinedOutput()
			text := string(output)
			switch mode {
			case "anchor", "escaping":
				if err != nil {
					t.Fatalf("%s counted=%t: %v\n%s", mode, counted, err, text)
				}
				if mode == "escaping" && !strings.Contains(text, "made at runtime\n") {
					t.Fatal(text)
				}
				if counted {
					if !strings.Contains(text, "live 3 bytes 192 reachable 2 bytes 128 unreachable 1 bytes 64") {
						t.Fatal(text)
					}
					if !strings.Contains(text, "graph counts: regions 1 merges 2") {
						t.Fatal(text)
					}
					if !strings.Contains(text, "allocations 5 frees 5 retains 0 releases 4 peak 5 regions 0") {
						t.Fatal(text)
					}
				} else if strings.Contains(text, "graph region:") {
					t.Fatal("diagnostic tracing in release build")
				}
			case "later":
				if err == nil || !strings.Contains(text, "ERROR: AddressSanitizer: heap-use-after-free") {
					t.Fatalf("later read not caught: %v\n%s", err, text)
				}
			case "leak":
				if err == nil || !strings.Contains(text, "ERROR: LeakSanitizer:") {
					t.Fatalf("unreleased anchor not caught: %v\n%s", err, text)
				}
			case "shared":
				if err == nil || !strings.Contains(text, "merging shared graph regions is not yet supported") {
					t.Fatalf("shared merge not rejected: %v\n%s", err, text)
				}
			}
		}
	}
}

// Interior cells count their owning environment, including when that environment
// is a graph member. Closures and environments then form plain internal edges.
func TestGraphClosureEnvironment(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
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
	binary := filepath.Join(t.TempDir(), "closure")
	if err := Build(harness, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	if !strings.Contains(string(output), "graph counts: regions 1 merges 1") {
		t.Fatal(string(output))
	}
	if !strings.Contains(string(output), "allocations 2 frees 2 retains 1 releases 3 peak 2 regions 0") {
		t.Fatal(string(output))
	}
	t.Logf("closure/environment counts:\n%s", output)
}

// This measures the runtime foundation. It is not an Adamic compiler benchmark:
// graph-type selection has not yet been connected to allocations.
func TestGraphRegionsMillion(t *testing.T) {
	// Not parallel: two large resident sets are measured sequentially.
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
 struct rusage usage;
 getrusage(RUSAGE_SELF, &usage);
 printf("rss_kib %ld\n", usage.ru_maxrss);
 adamic_release(root);
 return 0;
}
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + cmain
	directory := t.TempDir()
	binary := filepath.Join(directory, "million")
	if err := Build(source, binary, Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "live 1000000 bytes 64000000 reachable 950001 bytes 60800064 unreachable 49999 bytes 3199936 metadata 16000064") {
		t.Fatal(string(output))
	}
	t.Logf("native runtime foundation:\n%s", output)
	release := filepath.Join(directory, "million-release")
	if err := Build(source, release, Options{}); err != nil {
		t.Fatal(err)
	}
	releaseOutput, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", release).CombinedOutput()
	if err != nil {
		t.Fatalf("release: %v\n%s", err, releaseOutput)
	}
	if strings.Contains(string(releaseOutput), "graph region:") {
		t.Fatal("diagnostics in release")
	}
	t.Logf("native release flags: %s", strings.Join(Flags(Options{}), " "))
	t.Logf("native release:\n%s", releaseOutput)
	sanitized := filepath.Join(directory, "million-sanitized")
	if err := Build(source, sanitized, Options{Count: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	sanitizedOutput, err := exec.Command(sanitized).CombinedOutput()
	if err != nil {
		t.Fatalf("million-member sanitizer/leak check: %v\n%s", err, sanitizedOutput)
	}
	if !strings.Contains(string(sanitizedOutput), "allocations 1000000 frees 1000000") {
		t.Fatal(string(sanitizedOutput))
	}
	t.Log("million-member ASan, UBSan and LeakSanitizer: clean")
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
	if err := os.WriteFile(path, []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err = exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", "node", path).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "members 1000000 reachable 950001") {
		t.Fatal(string(output))
	}
	t.Logf("Node:\n%s", output)
}

func TestGraphContainerBoundary(t *testing.T) {
	t.Parallel()
	const main = `int main(void) {
 adamic_object *root = node(), *child = node();
 adamic_array *array = adamic_array_new(1, true);
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
	binary := filepath.Join(t.TempDir(), "containers")
	if err := Build(source, binary, Options{Count: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	if !strings.Contains(string(output), "graph counts: regions 1 merges 3") {
		t.Fatal(string(output))
	}
	if !strings.Contains(string(output), "live 4 bytes 592 reachable 3 bytes 192 unreachable 1 bytes 400") {
		t.Fatal(string(output))
	}
	if !strings.Contains(string(output), "allocations 6 frees 6 retains 1 releases 5 peak 6 regions 0") {
		t.Fatal(string(output))
	}
	t.Logf("container/boundary counts:\n%s", output)
}

func TestGraphLazyRegions(t *testing.T) {
	t.Parallel()
	const main = `int main(void) {
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
	binary := filepath.Join(t.TempDir(), "lazy")
	if err := Build(source, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	if !strings.Contains(string(output), "graph counts: regions 2 merges 3") || !strings.Contains(string(output), "allocations 5 frees 5") {
		t.Fatal(string(output))
	}
	if !strings.Contains(string(output), "live 1 bytes 64 reachable 1 bytes 64 unreachable 0 bytes 0 metadata 16") {
		t.Fatal(string(output))
	}
	t.Logf("lazy regions:\n%s", output)
}

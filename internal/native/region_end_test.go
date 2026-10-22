package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The planner currently refuses an escaping Weak, even though its runtime read becomes absent.
// Exercise regional targets directly: multiple blocks, multiple targets, and an unrelated handle
// that must remain present. A missed forget is dereferenced after free, so ASan sees it.
func TestRegionEndWeakTargets(t *testing.T) {
	t.Parallel()
	const source = `
#include "adamic.h"
#include <stdio.h>
static const char *const names[] = {"number"};
static const bool references[] = {false};
static const adamic_shape shape = {1, names, references, NULL, NULL};
int main(void) {
	adamic_region region = ADAMIC_REGION;
	adamic_object *first = adamic_object_new_filled_in(&region, &shape);
	first->slots[0].number = 7;
	adamic_weak *one = adamic_weak_of(first);
	adamic_object *last = first;
	for (size_t index = 0; index < 10000; index++) {
		last = adamic_object_new_filled_in(&region, &shape);
		last->slots[0].number = (double)index;
	}
	adamic_weak *two = adamic_weak_of(last);
	adamic_object *outside = adamic_object_new(&shape);
	outside->slots[0].number = 11;
	adamic_weak *other = adamic_weak_of(outside);
	adamic_region_end(&region);
	adamic_object *left = adamic_weak_target(one);
	adamic_object *right = adamic_weak_target(two);
	if (left != NULL) printf("missed first %.0f\n", left->slots[0].number);
	if (right != NULL) printf("missed last %.0f\n", right->slots[0].number);
	adamic_object *kept = adamic_weak_target(other);
	if (kept == NULL || kept->slots[0].number != 11) return 1;
	adamic_release(outside);
	if (adamic_weak_target(other) != NULL) return 2;
	adamic_release(other);
	adamic_release(two);
	adamic_release(one);
	// Also remove the last table entry from a region: no unrelated handle to keep the table alive.
	adamic_object *only = adamic_object_new_filled_in(&region, &shape);
	only->slots[0].number = 13;
	adamic_weak *handle = adamic_weak_of(only);
	adamic_region_end(&region);
	adamic_object *target = adamic_weak_target(handle);
	if (target != NULL) printf("missed only %.0f\n", target->slots[0].number);
	adamic_release(handle);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "weak-region")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:malloc_fill_byte=239")
	output, err := command.CombinedOutput()
	if err != nil || len(output) != 0 {
		t.Fatalf("regional Weak targets: %v\n%s", err, output)
	}
}

// A throw while later fields are evaluated must find no partially filled unzeroed literal.
// Fill fresh malloc storage with an aligned nonpointer pattern: UBSan's alignment check must not
// mask the ASan invalid read that would occur if allocation moved ahead of a throwing field.
func TestRegionEndThrowInitialization(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/region_end.a")
	if err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "throw-initialization")
	if err := Build(C(program), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, truth)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:malloc_fill_byte=240")
	output, err := command.CombinedOutput()
	if err != nil || string(output) != string(truth) {
		t.Fatalf("throw initialization: %v\n%s\nNode: %s", err, output, truth)
	}
}

// Emission controls distinguish filled literals, outside children, and transitive writes.
func TestRegionEndEmission(t *testing.T) {
	t.Parallel()
	program := loopProgram(t, "../oracle/testdata/region_end.a")
	plan := planRegions(program)
	seen := 0
	for index, function := range program.Functions {
		want, check := map[string]bool{"build": false, "count": false, "replace": true, "through": true}[function.Name]
		if check {
			seen++
			if plan.referenceWrites[index] != want {
				t.Errorf("%s writes references: got %t, want %t", function.Name, plan.referenceWrites[index], want)
			}
		}
	}
	if seen != 4 {
		t.Fatalf("found %d summary controls, want 4", seen)
	}
	source := C(program)
	for _, want := range []string{"adamic_object_new_filled_in(region,", "adamic_region_hold(region,", ".holds_outside = true;"} {
		if !strings.Contains(source, want) {
			t.Errorf("missing %q", want)
		}
	}
}

// The same handle is read while its region ends on a worker. The reader never
// dereferences the returned target: the table lock protects absence, not ownership.
func TestRegionEndConcurrentWeak(t *testing.T) {
	t.Parallel()
	if goruntime.GOOS != "linux" {
		t.Skip("requires Linux ThreadSanitizer")
	}
	const source = `
#define _XOPEN_SOURCE 700
#include "adamic.h"
#include <pthread.h>
#include <stdatomic.h>
#include <stdlib.h>
#define N 256
static const adamic_shape shape = {0, NULL, NULL, NULL, NULL};
static adamic_region regions[N];
static adamic_weak *handles[N];
static pthread_barrier_t barrier;
static _Atomic int done;
static void *end_regions(void *unused) {
 (void)unused;
 pthread_barrier_wait(&barrier);
 for (size_t i = 0; i < N; i++) adamic_region_end(&regions[i]);
 atomic_store_explicit(&done, 1, memory_order_relaxed);
 return NULL;
}
int main(void) {
 for (size_t i = 0; i < N; i++) {
  adamic_object *object = adamic_object_new_filled_in(&regions[i], &shape);
  handles[i] = adamic_weak_of(object);
 }
 pthread_barrier_init(&barrier, NULL, 2);
 pthread_t worker;
 if (pthread_create(&worker, NULL, end_regions, NULL) != 0) return 1;
 pthread_barrier_wait(&barrier);
 do {
  for (size_t i = 0; i < N; i++) (void)adamic_weak_target(handles[i]);
 } while (!atomic_load_explicit(&done, memory_order_relaxed));
 pthread_join(worker, NULL);
 for (size_t i = 0; i < N; i++) {
  if (adamic_weak_target(handles[i]) != NULL) return 2;
  adamic_release(handles[i]);
 }
 pthread_barrier_destroy(&barrier);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "region-weak-tsan")
	if err := Build(source, binary, Options{ThreadSanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "TSAN_OPTIONS=halt_on_error=1")
	output, err := command.CombinedOutput()
	if err != nil || len(output) != 0 {
		t.Fatalf("concurrent region Weak: %v\n%s", err, output)
	}
}

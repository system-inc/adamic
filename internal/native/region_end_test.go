package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
static const adamic_shape shape = {1, names, references, NULL};
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
	command := exec.Command(binary)
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
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:malloc_fill_byte=240")
	output, err := command.CombinedOutput()
	if err != nil || string(output) != string(truth) {
		t.Fatalf("throw initialization: %v\n%s\nNode: %s", err, output, truth)
	}
}

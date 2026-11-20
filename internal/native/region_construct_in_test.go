package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A constructor's region version makes a class literal with an Absent field as a construction object
// in its region, adamic_object_construct_in, and the region's end walks past each one's write order
// to the next object (#kg38qe5). The harness drives region.c alone, under AddressSanitizer and the
// counted build: construction objects of one to nine slots, every padding of the order words, each
// with a heap child, then a complete object with one, then the end; a region of one construction
// object alone; and the heap's construction object for a `NULL` region. Every child is let go, so
// nothing is live after.
//
// Mutants, by hand on Loom: the walk stepping by object_size alone stays green, since
// adamic_object_size reserves the order words too and both round to the same place at every count;
// construct_in not setting holds_outside reds "alone" (its child is never let go, live 1).
const regionConstructInHarness = `#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <stdlib.h>

static const char *const names[] = {"a", "b", "c", "d", "e", "f", "g", "h", "i"};
static const bool references[] = {true, false, false, false, false, false, false, false, false};
static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number};
#define SHAPE(count) {count, names, references, NULL, kinds}
static const adamic_shape shapes[] = {SHAPE(1), SHAPE(2), SHAPE(3), SHAPE(4), SHAPE(5), SHAPE(6), SHAPE(7), SHAPE(8), SHAPE(9)};
static const adamic_shape leaf = {0, NULL, NULL, NULL, NULL};

// construct makes a construction object in a region and writes a heap child into its first slot, as
// an emitted constructor's write does.
static void construct(adamic_region *region, const adamic_shape *shape) {
 adamic_object *object = adamic_object_construct_in(region, shape);
 if (object->write_order == NULL || (object->heap.slab & ADAMIC_REGION_VALUE) == 0 || adamic_object_present(object, 0)) { exit(3); }
 adamic_object_publish(object, 0);
 adamic_object_initialized(object)[0] = 1;
 object->slots[0].reference = adamic_object_new(&leaf);
}

int main(void) {
 adamic_region region = ADAMIC_REGION;
 for (size_t index = 0; index < sizeof shapes / sizeof shapes[0]; index++) { construct(&region, &shapes[index]); }
 adamic_object *complete = adamic_object_new_in(&region, &shapes[0]);
 complete->slots[0].reference = adamic_object_new(&leaf);
 adamic_region_end(&region);
 printf("walk live %zu\n", (size_t)adamic_counted.live);
 adamic_region alone = ADAMIC_REGION;
 construct(&alone, &shapes[2]);
 adamic_region_end(&alone);
 printf("alone live %zu\n", (size_t)adamic_counted.live);
 adamic_object *heap = adamic_object_construct_in(NULL, &shapes[0]);
 if (heap->write_order == NULL || (heap->heap.slab & ADAMIC_REGION_VALUE) != 0) { return 4; }
 adamic_release(heap);
 printf("heap live %zu\n", (size_t)adamic_counted.live);
 return 0;
}
`

func TestRegionConstructInWalk(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "region-construct-in")
	if err := Build(regionConstructInHarness, binary, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if want := "walk live 0\nalone live 0\nheap live 0\n"; err != nil || stdout.String() != want || strings.Contains(stderr.String(), "Sanitizer") || strings.Contains(stderr.String(), "runtime error") {
		t.Fatalf("construction objects in a region: %v\nstdout %q, want %q\n%s", err, stdout.String(), want, stderr.String())
	}
}

// The emitter's half of #kg38qe5: a region-planned caller hands a class's constructor its region, and
// the region version's class literal, given an Absent field, is made by adamic_object_construct_in.
// On main no source gives a class literal an Absent field (a function constructor does, on Lower's
// branch), so the test lowers a class whose constructor is region-eligible, checks the plan and the
// complete layout's adamic_object_new_in first, then gives the literal an Absent number field, as
// construction.go gives a factory's. Built under AddressSanitizer, the caller's reads of the object
// after the constructor returns are clean.
//
// Mutant, by hand on Loom: the emitter's construct_in branch off (the heap's adamic_object_construct
// for a region version too) reds this, no construct_in in the C and a heap use after free when the
// caller reads the object its region version released.
func TestRegionConstructorAbsentField(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "region-constructor.a")
	const source = `class Point {
	label: string;
	x: number;
	constructor(x: number) {
		this.x = x;
		this.label = 'point' + String(x);
	}
}
function measure(point: Point): number {
	return point.x + point.label.length;
}
console.log('measure ' + String(measure(new Point(3))));
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	// The statement-region plan, independent of the process's Program mode opt-in.
	program, err := lower.LowerWithOptions(context.Background(), loaded, lower.Options{})
	if err != nil {
		t.Fatal(err)
	}
	plan := planRegions(program)
	if len(plan.classObjects) != 1 || len(plan.statements) == 0 {
		t.Fatalf("probe lost its region-eligible constructor: %d class allocators, %d region statements", len(plan.classObjects), len(plan.statements))
	}
	if code := C(program); !strings.Contains(code, "adamic_object_new_in(region, &") || strings.Contains(code, "adamic_object_construct") {
		t.Fatal("the complete layout's region version doesn't make its object by adamic_object_new_in")
	}
	constructor := -1
	for function := range plan.classObjects {
		constructor = function
	}
	declare := program.Functions[constructor].Body[0].(ir.Declare)
	literal := declare.Value.(ir.ObjectLiteral)
	spare := ir.Field{Name: "spare", Value: ir.NumberConstant{}, Absent: true, Uninitialized: true}
	literal.Fields = append(literal.Fields, spare)
	declare.Value = literal
	program.Functions[constructor].Body[0] = declare
	program.Classes[literal.Class-1].Fields = append(program.Classes[literal.Class-1].Fields, spare)
	code := C(program)
	if !strings.Contains(code, "adamic_object_construct_in(region, &") {
		t.Fatal("the region version makes its construction object on the heap, not by adamic_object_construct_in")
	}
	binary := filepath.Join(t.TempDir(), "region-constructor")
	if err := Build(code, binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	output, err := command.CombinedOutput()
	if err != nil || string(output) != "measure 9\n" {
		t.Fatalf("region-planned constructor with an Absent field: %v\n%s", err, output)
	}
}

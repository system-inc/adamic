package native_test

import (
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentMixedRegionStorage(t *testing.T) {
	t.Parallel()
	const code = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include <assert.h>
#include <stdio.h>
static const char *const names[] = {"label"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};
int main(void) {
 adamic_stack_thread_start();
 adamic_region region = ADAMIC_REGION;
 adamic_string *text = adamic_string_from_number(123);
 adamic_weak *weak = NULL;
 for (size_t i = 0; i < 80; i++) {
  adamic_object *object = adamic_object_new_in(&region, &shape);
  object->slots[0].reference = adamic_retain(text);
  adamic_environment *environment = adamic_environment_new_in(&region, 3);
  assert(environment->heap.references == 0);
  assert(environment->cells[0].owner == &environment->heap);
  environment->cells[0].references = true;
  environment->cells[0].ready = true;
  environment->cells[0].value.reference = adamic_retain(text);
  adamic_retain(&environment->cells[0]); adamic_release(&environment->cells[0]);
  if (i == 0) { weak = adamic_weak_of(&environment->cells[0]); }
 }
 adamic_release(text);
 adamic_region_end(&region);
 assert(adamic_weak_target(weak) == NULL);
 adamic_release(weak);
 puts("mixed region released");
 adamic_heap_end();
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "mixed-region")
	if err := native.Build(code, binary, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("mixed-region lifetime: %v %s", err, output)
	}
	if !strings.Contains(string(output), "mixed region released") {
		t.Fatal(string(output))
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
	t.Log(string(output))
}

// A heap environment freed by its count forgets its interior cells, as the graph free and the Program
// end forget their members' (#0h34h5w): a Weak of a cell reads nothing after, and no handle is left in
// the side table under the freed address for the next value made there to inherit. Mutant: free_one's
// cell loop taken out, and the target still reads the freed cell (the assertion on it fails), while the
// next weak made at that address may come back as the stale handle.
func TestEnvironmentHeapCellsForgotten(t *testing.T) {
	t.Parallel()
	const code = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include <assert.h>
#include <stdio.h>
int main(void) {
 adamic_stack_thread_start();
 adamic_string *text = adamic_string_from_number(123);
 adamic_environment *environment = adamic_environment_new(2);
 assert(environment->heap.references == 1);
 assert(environment->cells[1].owner == &environment->heap);
 environment->cells[1].references = true;
 environment->cells[1].ready = true;
 environment->cells[1].value.reference = adamic_retain(text);
 adamic_weak *weak = adamic_weak_of(&environment->cells[1]);
 assert(adamic_weak_target(weak) == &environment->cells[1]);
 adamic_release(environment);
 if (adamic_weak_target(weak) != NULL) {
  puts("a weak of a freed environment's cell still names it");
  return 3;
 }
 adamic_release(weak);
 adamic_release(text);
 puts("heap environment cells forgotten");
 adamic_heap_end();
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "heap-environment-cells")
	if err := native.Build(code, binary, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("heap environment cells: %v %s", err, output)
	}
	if !strings.Contains(string(output), "heap environment cells forgotten") {
		t.Fatal(string(output))
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
}

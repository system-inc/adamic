package native_test

import (
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
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
	if err := native.Build(code, binary, native.Options{Sanitize: true, Count: true}); err != nil {
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

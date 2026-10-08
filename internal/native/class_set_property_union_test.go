package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The fallback must reject metadata-free scalars and foreign slots instead of
// interpreting their bits as a heap pointer. Ordinary represented fields pass.
func TestUnionRuntimeFieldGuards(t *testing.T) {
	const harness = `#include "adamic.h"
#include <string.h>
const adamic_object *adamic_union_slot_owner(const adamic_object *, const adamic_value *);
adamic_heap *adamic_union_runtime_field(const adamic_object *, const adamic_value *);
int main(int argc, char **argv) {
 (void)argc;
 static const char *const names[] = {"unrepresented"};
 static const bool references[] = {false};
 static const adamic_shape shape = {1, names, references, NULL};
 adamic_object *object = adamic_object_new(&shape);
 object->slots[0].number = 42;
 if (strcmp(argv[1], "scalar") == 0) adamic_union_runtime_field(object, &object->slots[0]);
 if (strcmp(argv[1], "foreign") == 0) { adamic_value foreign = {.number = 42}; adamic_union_slot_owner(object, &foreign); }
 adamic_release(object);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "guard")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, message string }{
		{"scalar", "union field view of a runtime scalar without representation metadata"},
		{"foreign", "union field slot has no owning layout"},
		{"valid", ""},
	} {
		output, err := exec.Command(binary, probe.name).CombinedOutput()
		if probe.message == "" {
			if err != nil || len(output) != 0 {
				t.Fatalf("control: %v %s", err, output)
			}
			continue
		}
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(string(output), probe.message) {
			t.Fatalf("%s: want named panic 70, got %v %s", probe.name, err, output)
		}
	}
}

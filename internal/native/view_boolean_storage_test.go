package native_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The mixed-field snapshot entry point can consume packed storage even where
// source lowering still refuses the mixed-union read family.
func TestPackedBooleanViewSnapshot(t *testing.T) {
	t.Parallel()
	want, err := exec.Command("node", "-e", "for (const value of [false, undefined, true]) console.log(`${String(value)} ${String(value)}`);").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	source := `
#include "adamic.h"
#include "view_unions_mixed.h"
#include <stdio.h>
int main(void) {
    static const char *const names[] = {"value"};
    static const bool references[] = {false};
    static const adamic_shape shape = {1, names, references, NULL};
    adamic_object *object = adamic_object_new(&shape);
    adamic_object_field_types(object)[0] = 9;
    const adamic_maybe_boolean values[] = {{true,false},{false,false},{true,true}};
    adamic_slot_cache cache = {NULL,0};
    for (size_t i = 0; i < 3; i++) {
        object->slots[0].maybe_boolean = adamic_maybe_boolean_pack(values[i]);
        adamic_view_union_value snapshot = adamic_object_view_union_snapshot(object,"value",&cache,"actual.value","boolean | undefined",false);
        const char *observed = snapshot.kind == adamic_view_union_undefined ? "undefined" : snapshot.kind == adamic_view_union_boolean ? (snapshot.payload.boolean ? "true" : "false") : "wrong kind";
        adamic_value optional = adamic_object_optional_view(object,"value",&cache,9,"boolean | undefined","actual.value",true,false);
        const char *through_view = optional.reference == NULL ? "undefined" : ((adamic_boolean_box *)optional.reference)->boolean ? "true" : "false";
        printf("%s %s\n", observed, through_view);
        fflush(stdout);
    }
    adamic_release(object);
    return 0;
}
`
	for _, mode := range []struct {
		name    string
		options native.Options
	}{
		{"sanitized", native.Options{Sanitize: true}}, {"release", native.Options{}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "snapshot")
			if err := native.Build(source, binary, mode.options); err != nil {
				t.Fatal(err)
			}
			got, err := exec.Command(binary).CombinedOutput()
			if err != nil || string(got) != string(want) {
				t.Fatalf("got %q, %v; Node %q", got, err, want)
			}
			if mode.options.Sanitize {
				if report := leakcheck.Report(t, source, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

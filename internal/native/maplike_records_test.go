package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The compiler refuses scalar named views. Keep the erased runtime boundary
// loud as well, and prove its exit/diagnostic check can reject a real mutant.
func TestMapLikeScalarErasureGuard(t *testing.T) {
	const source = `#include "adamic.h"
#include <string.h>
int main(void) {
    adamic_object *record = adamic_record_new(false);
    adamic_string *key = adamic_string_allocate(5);
    memcpy((char *)key->bytes, "value", 5);
    adamic_record_define(record, key, (adamic_value){.number = 42});
    adamic_heap *value = adamic_dynamic_property((adamic_heap *)record, "value");
    adamic_release(value);
    adamic_release(record);
    return 0;
}
`
	binary := filepath.Join(t.TempDir(), "guard")
	if err := Build(source, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := recordRun(binary)
	const message = "adamic: panic: dynamic read of a record scalar without type metadata\n"
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || stdout != "" || !strings.HasPrefix(stderr, message) || !recordCounts.MatchString(strings.TrimPrefix(stderr, message)) {
		t.Fatalf("want exact scalar erasure stop: %v %q %q", err, stdout, stderr)
	}
	mutant := recordMutant(t, "union.c", `adamic_panic("dynamic read of a record scalar without type metadata", sizeof "dynamic read of a record scalar without type metadata" - 1);`, "return NULL;", source)
	stdout, stderr, err = recordRun(mutant)
	if err != nil || stdout != "" {
		t.Fatalf("mutant must finish cleanly: %v %q %q", err, stdout, stderr)
	}
	recordCheckCounts(t, stderr)
	t.Log("scalar erasure mutant caught by exact diagnostic and exit check; mutant exits 0 without sanitizer or leak failures")
}

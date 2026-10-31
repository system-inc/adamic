package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// adamic_object_assign stores each field as the target holds it. Lowering refuses a name held one
// way in the source and another in the target (library_object.go), so this drives the runtime
// directly: a mismatch in either direction stops by name, its 200-byte field name printed whole
// (the message is sized from the name, never a fixed buffer), and an agreeing reference field is
// retained by the copy and released once by each side. Released by the source's flag instead, the
// mismatch frees a number's bits as a pointer, and the named panic never comes.
func TestObjectAssignStoresByTheTargetsFlag(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <string.h>

#define LONG_NAME "f_" TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN "_end"
#define TEN "0123456789"
static const char *const names[] = {LONG_NAME};
static const bool counted[] = {true};
static const bool scalar[] = {false};
static const adamic_shape counted_shape = {1, names, counted, NULL};
static const adamic_shape scalar_shape = {1, names, scalar, NULL};

static adamic_object *with_string(void) {
 static adamic_string a = ADAMIC_STRING("made ");
 static adamic_string b = ADAMIC_STRING("at runtime");
 adamic_object *object = adamic_object_new(&counted_shape);
 object->slots[0].reference = adamic_string_concat(2, (adamic_string *const[]){&a, &b});
 return object;
}
static adamic_object *with_number(void) {
 adamic_object *object = adamic_object_new(&scalar_shape);
 object->slots[0].number = 1;
 return object;
}
int main(int count, char **arguments) {
 (void)count;
 if (strcmp(arguments[1], "mismatch") == 0) {
  adamic_object *target = with_number(), *source = with_string();
  adamic_object_assign(target, source);
  return 1;
 }
 if (strcmp(arguments[1], "reverse") == 0) {
  adamic_object *target = with_string(), *source = with_number();
  adamic_object_assign(target, source);
  return 1;
 }
 adamic_object *target = with_string(), *source = with_string();
 adamic_object_assign(target, source);
 if (target->slots[0].reference != source->slots[0].reference) { return 2; }
 adamic_release(source);
 adamic_string *kept = target->slots[0].reference;
 if (kept->length != 15) { return 3; }
 adamic_release(target);
 return 0;
}
`
	name := "f_" + strings.Repeat("0123456789", 19) + "_end"
	binary := filepath.Join(t.TempDir(), "assign")
	if err := native.Build(harness, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"mismatch", "reverse"} {
		command := exec.Command(binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := command.CombinedOutput()
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(string(output), "compiler bug: Object.assign of field "+name+", held as a reference on one side only\n") || strings.Contains(string(output), "Sanitizer") {
			t.Fatalf("%s: want the named panic and exit 70: %v\n%s", mode, err, output)
		}
	}
	command := exec.Command(binary, "agree")
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("agree: %v\n%s", err, output)
	}
	if report := leakcheck.Report(t, harness, binary, "agree"); report != "" {
		t.Fatalf("agree leaked: %s", report)
	}
}

package native_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// adamic_object_assign stores each field as the target holds it. Lowering refuses a name held one
// way in the source and another in the target (library_object.go), so this drives the runtime
// directly: a mismatch in either direction stops by name, its 200-byte field name printed whole
// (the message is sized from the name, never a fixed buffer), and an agreeing reference field is
// retained by the copy and released once by each side. Released by the source's flag instead, the
// mismatch frees a number's bits as a pointer, and the named panic never comes.
func TestObjectAssignStoresByTheTargetsFlag(test *testing.T) {
	test.Parallel()
	const harness = `#include "adamic.h"
#include <string.h>

#define LONG_NAME "f_" TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN TEN "_end"
#define TEN "0123456789"
static const char *const names[] = {LONG_NAME};
static const bool counted[] = {true};
static const bool scalar[] = {false};
static const adamic_field_kind counted_kinds[] = {adamic_field_reference};
static const adamic_field_kind scalar_kinds[] = {adamic_field_number};
static const adamic_shape counted_shape = {1, names, counted, NULL, counted_kinds};
static const adamic_shape scalar_shape = {1, names, scalar, NULL, scalar_kinds};

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
	binary := filepath.Join(test.TempDir(), "assign")
	if buildError := native.Build(harness, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	for _, mode := range []string{"mismatch", "reverse"} {
		command := exec.Command(binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		output, combinedOutputError := command.CombinedOutput()
		exit, ok := combinedOutputError.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(
			string(output),
			"compiler bug: Object.assign of field "+name+", held as a reference on one side only\n",
		) || strings.Contains(string(output), "Sanitizer") {
			test.Fatalf("%s: want the named panic and exit 70: %v\n%s", mode, combinedOutputError, output)
		}
	}
	command := exec.Command(binary, "agree")
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("agree: %v\n%s", combinedOutputError, output)
	}
	if report := leakcheck.Report(test, harness, binary, "agree"); report != "" {
		test.Fatalf("agree leaked: %s", report)
	}
}

// Object.assign onto a class never reaches the runtime: its target must be a plain literal or that
// literal's const binding (library_object.go), so a static class target is refused in lowering. That
// keeps adamic_object_write_field's static branch unreached from here, and this pins why.
func TestObjectAssignOntoAStaticClassIsRefusedInLowering(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := "class Base { static count = 1; }\nclass Child extends Base {}\nObject.assign(Child, { count: 2 });\nconsole.log(Base.count + ' ' + Child.count);\n"
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), checked)
	var notYet *lower.NotYet
	if !errors.As(lowerError, &notYet) ||
		notYet.What != "Object.assign on a shape not proven by a plain literal or its const binding" {
		test.Fatalf("want Object.assign onto a class refused in lowering, got %v", lowerError)
	}
}

package native_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// objectPrototypeHarness reads the key its argument names off an object by a run-time key. Given
// "program", the object's shape is a literal's, with no methods, so the read must give the same
// immortal closure twice, whose code it then calls directly. Given "host", the shape's methods are
// adamic_host_methods, as a Stats' are, and the read must stop.
const objectPrototypeHarness = `#include "adamic.h"
#include <assert.h>
#include <string.h>
static const char *const names[] = {"present"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};
static const adamic_shape host_shape = {1, names, references, &adamic_host_methods, kinds};
int main(int count, char **arguments) {
 assert(count == 3);
 adamic_stack_thread_start();
 static adamic_string here = ADAMIC_STRING("here");
 adamic_object *object = adamic_object_new(strcmp(arguments[2], "program") == 0 ? &shape : &host_shape);
 object->slots[0].reference = &here;
 adamic_string *key = adamic_decode_utf8((const unsigned char *)arguments[1], strlen(arguments[1]));
 adamic_heap *first = adamic_dynamic_property_keyed(&object->heap, key, NULL, 0);
 adamic_heap *second = adamic_dynamic_property_keyed(&object->heap, key, NULL, 0);
 assert(first != NULL && first == second && first->kind == adamic_kind_closure && first->references == 0);
 adamic_closure *closure = (adamic_closure *)first;
 closure->code(closure, NULL);
 return 0;
}
`

// The closures a run-time key read hands out for Object.prototype's methods and for Object are
// guards: no program can call one, since lowering refuses every call through an unknown narrowed to
// a function ("calling through the unchecked Function type", function_rulings.go). If that refusal
// ever loosens, the call must stop loudly, naming what wasn't called, never run as something else.
// Mutant: the code function returning undefined, so the harness exits 0. An object whose shape the
// runtime marks host, a Stats or a Date, has another prototype than Object's, so on one the read
// stops instead.
func TestObjectPrototypeClosuresStopWhenCalled(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "object-prototype")
	if err := native.Build(objectPrototypeHarness, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ key, object, message string }{
		{"toString", "program", "Object.prototype.toString isn't called yet"},
		{"__lookupSetter__", "program", "Object.prototype.__lookupSetter__ isn't called yet"},
		{"constructor", "program", "Object isn't called yet"},
		{"toString", "host", "this runtime value's properties aren't read by a run-time key yet (reading 'toString')"},
	} {
		t.Run(probe.object+"/"+probe.key, func(t *testing.T) {
			t.Parallel()
			command := exec.Command(binary, probe.key, probe.object)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stdout.Len() != 0 || stderr.String() != "adamic: panic: "+probe.message+"\n" || strings.Contains(stderr.String(), "Sanitizer") {
				t.Fatalf("want the loud stop and exit 70: %v\nstdout %q\nstderr %q", err, stdout.String(), stderr.String())
			}
		})
	}
}

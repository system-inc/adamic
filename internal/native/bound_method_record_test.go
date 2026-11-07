package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A record may change after a function-property narrowing. Binding must inspect
// the current value's kind before treating it as a closure.
func TestBoundMethodRecordNonfunction(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <string.h>

static adamic_string *text(const char *bytes) {
	size_t length = strlen(bytes);
	adamic_string *value = adamic_string_allocate(length);
	memcpy((char *)value->bytes, bytes, length);
	return value;
}

int main(void) {
	adamic_record *record = adamic_record_new(adamic_field_reference);
	adamic_record_define(record, text("m"), (adamic_value){.reference = text("not callable")});
	adamic_slot_cache cache = {0};
	adamic_method_entry method = NULL;
	adamic_closure *callee = adamic_object_callee(record, "m", &cache, &method);
	(void)adamic_closure_call(callee, NULL, 0);
	adamic_release(record);
	adamic_heap_end();
	return 0;
}
`
	flags := Flags(Options{Sanitize: true})
	compiler, library := runtimeWithFlags(t, flags)
	root := filepath.Dir(library)
	flags = append(flags, "-I", root)
	for _, mutant := range []bool{false, true} {
		name := "guarded"
		archive := library
		if mutant {
			name = "kind_check_removed"
			archive = boundMethodRecordKindMutant(t, compiler, library, flags)
		}
		t.Run(name, func(t *testing.T) {
			binary := buildHarnessWithLibrary(t, harness, flags, compiler, archive)
			stdout, stderr, err := recordRun(binary)
			exit, exited := err.(*exec.ExitError)
			matches := exited && exit.ExitCode() == 70 && stdout == "" && stderr == "adamic: panic: TypeError: record key 'm' is not a function\n"
			if !mutant {
				if !matches {
					t.Fatalf("want exit 70 naming key m, got %v\nstdout %q\nstderr %s", err, stdout, stderr)
				}
				return
			}
			report := firstSanitizerLine(stderr)
			if matches || err == nil || report == "" {
				t.Fatalf("kind-check mutant must compile and fail under the sanitizer, got %v\nstdout %q\nstderr %s", err, stdout, stderr)
			}
			t.Logf("kind-check mutant compiled; exit-70 TypeError pin rejected it: %s", report)
		})
	}
}

// Replace only object.c in the same sanitized runtime archive.
func boundMethodRecordKindMutant(t *testing.T, compiler, library string, flags []string) string {
	t.Helper()
	source, err := runtime.ReadFile("runtime/object.c")
	if err != nil {
		t.Fatal(err)
	}
	const guard = "\t\tconst adamic_heap *reference = slot->reference;\n\t\tif (reference->kind == adamic_kind_closure) { return slot->reference; }"
	if strings.Count(string(source), guard) != 1 {
		t.Fatal("record callee kind-check mutation site moved")
	}
	directory := t.TempDir()
	path, object := filepath.Join(directory, "object.c"), filepath.Join(directory, "object.o")
	if err := os.WriteFile(path, []byte(strings.Replace(string(source), guard, "\t\treturn slot->reference;", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-c", path, "-o", object)
	if output, err := productCompilerCommand(t, t.Context(), compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("kind-check mutant must compile: %v\n%s", err, output)
	}
	data, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(directory, "runtime.a")
	if err := os.WriteFile(archive, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(archiverName(compiler), "rcs", archive, object).CombinedOutput(); err != nil {
		t.Fatalf("kind-check mutant archive: %v\n%s", err, output)
	}
	return archive
}

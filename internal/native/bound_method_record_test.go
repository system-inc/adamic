package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
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
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true})
	compiler, library := runtimeWithFlags(t, flags)
	root := filepath.Dir(library)
	flags = append(flags, "-I", root)
	for _, mutant := range []bool{false, true} {
		name := "guarded"
		archive := library
		if mutant {
			name = "kind_check_removed"
			const guard = "\t\tconst adamic_heap *reference = references ? slot->reference : NULL;\n\t\tif (reference != NULL && reference->kind == adamic_kind_closure) { return slot->reference; }"
			archive = runtimeSourceMutant(t, compiler, library, flags, "object.c", guard, "\t\treturn slot->reference;")
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

// A record of numbers keeps them unboxed in its slots, so a call through one of its keys answers not
// callable in the reference record's words, before anything reads the slot as a pointer (#3tjjcjs).
// Mutant: the record's kind no longer asked, so the packed number is read as a pointer and the
// sanitizer stops it.
func TestBoundMethodRecordScalar(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <string.h>

int main(void) {
	adamic_record *record = adamic_record_new(adamic_field_number);
	adamic_string *key = adamic_string_allocate(1);
	memcpy((char *)key->bytes, "m", 1);
	adamic_record_define(record, key, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){.present = true, .number = 41})});
	adamic_slot_cache cache = {0};
	adamic_method_entry method = NULL;
	adamic_closure *callee = adamic_object_callee(record, "m", &cache, &method);
	(void)adamic_closure_call(callee, NULL, 0);
	adamic_release(record);
	adamic_heap_end();
	return 0;
}
`
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true})
	compiler, library := runtimeWithFlags(t, flags)
	root := filepath.Dir(library)
	flags = append(flags, "-I", root)
	for _, mutant := range []bool{false, true} {
		name := "kind_asked"
		archive := library
		if mutant {
			name = "kind_not_asked"
			archive = runtimeSourceMutant(t, compiler, library, flags, "object.c", "\t\tbool references = adamic_record_values(object) == adamic_field_reference;", "\t\tbool references = true;")
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
			if matches || err == nil {
				t.Fatalf("kind mutant must fail other than the TypeError, got %v\nstdout %q\nstderr %s", err, stdout, stderr)
			}
			t.Logf("kind mutant failed: %s", firstSanitizerLine(stderr))
		})
	}
}

// Checked Object.values over a record of numbers reads each value by the record's kind, never the slot
// as a pointer (#3tjjcjs). No program reaches this road yet: a number record seen as the fixed view a
// checked enumeration takes stops at compile time (fixed objects and records have different storage),
// so the runtime is driven directly. Mutant: the raw retain restored, and the sanitizer stops it.
func TestCheckedValuesRecordScalar(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <stdio.h>
#include <string.h>

adamic_array *adamic_object_values_checked(adamic_object *, int, const char *, bool, const adamic_value *, size_t);

int main(void) {
	adamic_record *record = adamic_record_new(adamic_field_number);
	adamic_string *key = adamic_string_allocate(1);
	memcpy((char *)key->bytes, "k", 1);
	adamic_record_define(record, key, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){.present = true, .number = 41})});
	adamic_array *values = adamic_object_values_checked(record, 1, "number", false, NULL, 0);
	printf("%zu %g\n", values->length, values->elements[0].number);
	adamic_release(values);
	adamic_release(record);
	adamic_heap_end();
	return 0;
}
`
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true})
	compiler, library := runtimeWithFlags(t, flags)
	root := filepath.Dir(library)
	flags = append(flags, "-I", root)
	for _, mutant := range []bool{false, true} {
		name := "kind_read"
		archive := library
		if mutant {
			name = "raw_retain"
			archive = runtimeSourceMutant(t, compiler, library, flags, "library_object.c", "observed = slot == NULL ? NULL : adamic_record_value_union(object, slot);", "observed = slot == NULL ? NULL : adamic_retain(slot->reference);")
		}
		t.Run(name, func(t *testing.T) {
			binary := buildHarnessWithLibrary(t, harness, flags, compiler, archive)
			stdout, stderr, err := recordRun(binary)
			if !mutant {
				if err != nil || stdout != "1 41\n" {
					t.Fatalf("want 1 41, got %v\nstdout %q\nstderr %s", err, stdout, stderr)
				}
				return
			}
			if err == nil && stdout == "1 41\n" {
				t.Fatalf("raw-retain mutant must fail, got stdout %q\nstderr %s", stdout, stderr)
			}
			t.Logf("raw-retain mutant failed: %v %s", err, firstSanitizerLine(stderr))
		})
	}
}

// runtimeSourceMutant is the sanitized runtime archive with only one runtime file replaced, its one site
// rewritten as replacement.
func runtimeSourceMutant(t *testing.T, compiler, library string, flags []string, file, site, replacement string) string {
	t.Helper()
	source, err := native_runtime.Sources.ReadFile("c/" + file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), site) != 1 {
		t.Fatalf("%s mutation site moved: %q", file, site)
	}
	directory := t.TempDir()
	path, object := filepath.Join(directory, file), filepath.Join(directory, strings.TrimSuffix(file, ".c")+".o")
	if err := os.WriteFile(path, []byte(strings.Replace(string(source), site, replacement, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-c", path, "-o", object)
	if output, err := productCompilerCommand(t, t.Context(), compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("%s mutant must compile: %v\n%s", file, err, output)
	}
	data, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(directory, "runtime.a")
	if err := os.WriteFile(archive, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(native_runtime.ArchiverName(compiler), "rcs", archive, object).CombinedOutput(); err != nil {
		t.Fatalf("%s mutant archive: %v\n%s", file, err, output)
	}
	return archive
}

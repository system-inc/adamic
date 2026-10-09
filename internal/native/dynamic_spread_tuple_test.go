package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The same dynamic copy entry implements spreads that reserve a new field and
// spreads of an already dynamic object. Both results must be plain objects.
const dynamicSpreadTupleHarness = `#include "adamic.h"
#include "view_representations.h"
#include <stdio.h>
#ifdef adamic_slot_index
#error c2 must use runtime's inline slot index
#endif
int main(void) {
    static const char *const names[] = {"value"};
    static const char *const extra_names[] = {"extra"};
    static const bool references[] = {false};
    static const adamic_shape base = {1, names, references, NULL};
    static const adamic_shape extra = {1, extra_names, references, NULL};
    adamic_object *source = adamic_object_new(&base);
    source->tuple = true;
    source->slots[0].number = 7;
    adamic_object_field_types(source)[0] = adamic_rep_number;
    adamic_object *first = adamic_object_copy_reserving_checked(source, &extra, NULL);
    adamic_object *second = adamic_object_copy_checked(first, NULL);
    bool tuple = first->tuple || second->tuple;
    bool dynamic = first->dynamic_shape && second->dynamic_shape;
    bool copied = first->slots[0].number == 7 && second->slots[0].number == 7;
    adamic_release(second);
    adamic_release(first);
    adamic_release(source);
    if (tuple) { puts("dynamic spread retained tuple flag"); return 1; }
    if (!dynamic || !copied) { return 2; }
    puts("dynamic spreads are plain");
    return 0;
}
`

func TestDynamicSpreadTupleInitialized(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "control")
	if err := Build(dynamicSpreadTupleHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil || string(output) != "dynamic spreads are plain\n" {
		t.Fatalf("sanitized dynamic spreads: %v\n%s", err, output)
	}
}

// Nonzero allocator bytes are legal. Seed only the tuple byte immediately after
// allocation in a private runtime, so the omission is deterministic on platforms
// without MemorySanitizer too. The control uses the identical seed.
func TestDynamicSpreadTupleInitializationMutant(t *testing.T) {
	t.Parallel()
	flags := Flags(Options{Sanitize: true})
	for _, omitted := range []bool{false, true} {
		binary := dynamicSpreadTupleRuntime(t, flags, true, omitted)
		output, err := exec.Command(binary).CombinedOutput()
		if !omitted && (err != nil || string(output) != "dynamic spreads are plain\n") {
			t.Fatalf("seeded control: %v\n%s", err, output)
		}
		if omitted && (err == nil || string(output) != "dynamic spread retained tuple flag\n") {
			t.Fatalf("initialization omission escaped: %v\n%s", err, output)
		}
		t.Logf("omitted=%v: %v, %s", omitted, err, output)
	}
}

func TestDynamicSpreadTupleMemorySanitizer(t *testing.T) {
	t.Parallel()
	flags := append(Flags(Options{Malloc: true}), "-g", "-fsanitize=memory", "-fsanitize-memory-track-origins=2")
	// Check the compiler/runtime capability rather than assuming the platform.
	directory := t.TempDir()
	probe := filepath.Join(directory, "probe.c")
	if err := os.WriteFile(probe, []byte("int main(void) { return 0; }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	probeBinary := filepath.Join(directory, "probe")
	if output, err := exec.Command("clang", append(append([]string{}, flags...), probe, "-o", probeBinary)...).CombinedOutput(); err != nil {
		t.Skipf("MemorySanitizer unavailable; deterministic sanitized mutant remains active: %v\n%s", err, output)
	}
	if output, err := exec.Command(probeBinary).CombinedOutput(); err != nil {
		t.Skipf("MemorySanitizer cannot run; deterministic sanitized mutant remains active: %v\n%s", err, output)
	}
	for _, omitted := range []bool{false, true} {
		binary := dynamicSpreadTupleRuntime(t, flags, false, omitted)
		output, err := exec.Command(binary).CombinedOutput()
		if !omitted && (err != nil || string(output) != "dynamic spreads are plain\n") {
			t.Fatalf("MemorySanitizer control: %v\n%s", err, output)
		}
		if omitted && (err == nil || !strings.Contains(string(output), "MemorySanitizer: use-of-uninitialized-value")) {
			t.Fatalf("initialization omission escaped MemorySanitizer: %v\n%s", err, output)
		}
		t.Logf("omitted=%v: %v, %s", omitted, err, output)
	}
}

func dynamicSpreadTupleRuntime(t *testing.T, flags []string, seeded, omitted bool) string {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for index := range files {
		if files[index].name != "object.c" {
			continue
		}
		before, body, found := strings.Cut(string(files[index].contents), "adamic_object *adamic_object_copy_reserving_checked(")
		if !found || strings.Count(body, "object->tuple = false;") != 1 {
			t.Fatal("dynamic tuple initialization seam changed")
		}
		if seeded {
			allocation := "adamic_object *object = adamic_allocate(size + sizeof(adamic_shape) + capacity * (sizeof(const char *) + sizeof(int) + sizeof(bool)), adamic_kind_object);"
			if strings.Count(body, allocation) != 1 {
				t.Fatal("dynamic allocation seam changed")
			}
			body = strings.Replace(body, allocation, allocation+"\n\tobject->tuple = true;", 1)
		}
		if omitted {
			body = strings.Replace(body, "\tobject->tuple = false;\n", "", 1)
		}
		files[index].contents = []byte(before + "adamic_object *adamic_object_copy_reserving_checked(" + body)
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	library, err := cachedRuntime(files, flags, compiler, string(version), filepath.Join(directory, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	main, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "harness")
	if err := os.WriteFile(main, []byte(dynamicSpreadTupleHarness), 0600); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), main, "-o", binary)
	arguments = append(append(arguments, RuntimeLinkFlags(library)...), "-lm")
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("harness must compile: %v\n%s", err, output)
	}
	return binary
}

package oracle

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"array_next", "string_next", "map_set_live", "tsc_helpers", "identity_close", "to_array"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/library_iterator_" + name + ".a", lowers: true})
	}
}

// Every runtime mutant must compile, exit 0 and stay sanitizer-clean. Only source Node
// decides whether its output is wrong; a native invariant failure does not kill these mutants.
func TestLibraryIteratorProtocolMutants(t *testing.T) {
	mutants := []struct{ name, fixture, file, before, after string }{
		{"map_keys", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 1 && set == false && iterator->next == 0) { iterator->next = 1; }"},
		{"map_values", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 2 && set == false && iterator->next == 0) { iterator->next = 1; }"},
		{"map_entries", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 3 && set == false && iterator->next == 0) { iterator->next = 1; }"},
		{"set_keys", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 1 && set == true && iterator->next == 0) { iterator->next = 1; }"},
		{"set_values", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 2 && set == true && iterator->next == 0) { iterator->next = 1; }"},
		{"set_entries", "map_set_live", "map_set.c", "bool set = state->slots[4].boolean;", "bool set = state->slots[4].boolean;\n\tif (part == 3 && set == true && iterator->next == 0) { iterator->next = 1; }"},
		{"array_keys", "array_next", "map_set.c", "size_t position = (size_t)state->slots[2].number;", "size_t position = (size_t)state->slots[2].number;\n\tif (part == 1 && position == 0 && source->kind == adamic_kind_array) { position = 1; }"},
		{"array_values", "array_next", "map_set.c", "size_t position = (size_t)state->slots[2].number;", "size_t position = (size_t)state->slots[2].number;\n\tif (part == 2 && position == 0 && source->kind == adamic_kind_array) { position = 1; }"},
		{"array_entries", "array_next", "map_set.c", "size_t position = (size_t)state->slots[2].number;", "size_t position = (size_t)state->slots[2].number;\n\tif (part == 3 && position == 0 && source->kind == adamic_kind_array) { position = 1; }"},
		{"string_code_units", "string_next", "map_set.c", "width = 2;", "width = 1;"},
		{"array_revived", "array_next", "map_set.c", "state->slots[4].boolean || position >= length", "position >= length"},
		{"live_bound", "map_set_live", "map.c", "iterator->next < iterator->map->used", "iterator->next < iterator->map->count"},
		{"done_flag", "map_set_live", "map_set.c", "result->slots[0].boolean = !present;", "result->slots[0].boolean = true;"},
		{"tsc_callback_order", "tsc_helpers", "map_set.c", "pair->slots[0] = key;\n\t\tpair->slots[1] = value;", "pair->slots[0] = value;\n\t\tpair->slots[1] = key;"},
		{"iterator_identity_copy", "identity_close", "map_set.c", "adamic_retain(self)", "adamic_object_copy(self)"},
		{"shared_next_skip", "identity_close", "map_set.c", "return next->code(next, arguments);", "adamic_value skipped = next->code(next, arguments);\n\tadamic_release(skipped.reference);\n\treturn next->code(next, arguments);"},
		{"to_array_skip", "to_array", "map_set.c", "state->slots[2].number = (double)(position + 1);", "state->slots[2].number = (double)(position + 2);"},
		{"completion_value", "map_set_live", "map_set.c", "(adamic_maybe_number){false, 0}", "(adamic_maybe_number){true, 0}"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_iterator_"+mutant.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			files, err := filepath.Glob(filepath.Join(repository, "internal/native/runtime/*"))
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			for _, path := range files {
				if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" {
					continue
				}
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if filepath.Base(path) == mutant.file {
					before := string(source)
					after := strings.Replace(before, mutant.before, mutant.after, 1)
					if before == after {
						t.Fatal("mutant changed no runtime code")
					}
					source = []byte(after)
				}
				if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			options := native.Options{Sanitize: true}
			library, err := native.RuntimeLibrary(directory, options)
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(directory, "program.c")
			if err := os.WriteFile(source, []byte(native.C(program)), 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "mutant")
			flags := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, source)
			flags = append(flags, native.RuntimeLinkFlags(library)...)
			flags = append(flags, "-lm")
			if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
				t.Fatalf("link: %v\n%s", err, output)
			}
			// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
			sanitizer := "ASAN_OPTIONS=halt_on_error=1"
			if runtime.GOOS == "linux" {
				sanitizer = "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"
			}
			result := executeWith(t, []string{sanitizer, "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("Node comparison did not catch mutant: %q", difference)
			}
			t.Logf("%s: clean exit 0, ASan/UBSan/LSan clean, caught only by Node stdout", mutant.name)
		})
	}
}

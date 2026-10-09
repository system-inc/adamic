package oracle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, kind := range []string{"objects", "classes", "numbers", "strings", "mixed"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_map_set_keys_" + kind + ".a", true, false})
	}
}

// Each mutant changes runtime C in a private snapshot. Both the Node source and
// the sanitized mutant must finish cleanly; only their stdout may reject it.
func TestMapSetAnyKeyMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, kind, before, after string }{
		{"object shape equality", "objects", "default:\n\t\treturn left == right;", "default:\n\t\treturn true;"},
		{"class shape equality", "classes", "default:\n\t\treturn left == right;", "default:\n\t\treturn true;"},
		{"number NaN unequal", "numbers", "return left.number == right.number || (isnan(left.number) && isnan(right.number));", "return left.number == right.number;"},
		{"string identity equality", "strings", "return adamic_string_equal(left.reference, right.reference);", "return left.reference == right.reference;"},
		{"mixed NaN unequal", "mixed", "return a == b || (isnan(a) && isnan(b));", "return a == b;"},
		{"mixed numeric identity hash", "mixed", "if (reference != NULL && reference->kind == adamic_kind_number) {", "if (false && reference != NULL && reference->kind == adamic_kind_number) {"},
		{"mixed string identity hash", "mixed", "if (reference != NULL && reference->kind == adamic_kind_string) {", "if (false && reference != NULL && reference->kind == adamic_kind_string) {"},
		{"mixed zero not normalized", "mixed", "adamic_heap *zero = adamic_box_number(0);", "adamic_heap *zero = adamic_box_number(number);"},
		{"readded entry order", "strings", "map->used++;\n\tmap->count++;", "map->used++;\n\tmap->count++;\n\tif (map->used > map->count) {\n\t\tsize_t first = 0, last = map->used - 1;\n\t\twhile (map->entries[first].deleted) first++;\n\t\tadamic_map_entry swap = map->entries[first];\n\t\tmap->entries[first] = map->entries[last];\n\t\tmap->entries[last] = swap;\n\t\tfor (size_t bucket = 0; bucket < map->bucket_count; bucket++) {\n\t\t\tif (map->buckets[bucket] == first + 1) map->buckets[bucket] = last + 1;\n\t\t\telse if (map->buckets[bucket] == last + 1) map->buckets[bucket] = first + 1;\n\t\t}\n\t}"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_map_set_keys_"+mutant.kind+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			snapshot := t.TempDir()
			directory := filepath.Join(repository, "internal/native/runtime")
			entries, err := os.ReadDir(directory)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for _, entry := range entries {
				if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
					continue
				}
				contents, err := os.ReadFile(filepath.Join(directory, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if entry.Name() == "map.c" {
					source := string(contents)
					if !strings.Contains(source, mutant.before) {
						t.Fatal("mutant input missing")
					}
					contents = []byte(strings.Replace(source, mutant.before, mutant.after, 1))
					changed = true
				}
				if err := os.WriteFile(filepath.Join(snapshot, entry.Name()), contents, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			options := native.Options{Sanitize: true}
			library, err := native.RuntimeLibrary(snapshot, options)
			if err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(snapshot, "main.c")
			if err := os.WriteFile(main, []byte(native.C(program)), 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(snapshot, "mutant")
			arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, main)
			arguments = append(arguments, native.RuntimeLinkFlags(library)...)
			arguments = append(arguments, "-lm", "-pthread")
			if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
				t.Fatalf("must compile: %v\n%s", err, output)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("failed outside comparison: %+v", got)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node failed: %+v", truth)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("want Node stdout to catch mutant, got %q", difference)
			}
			t.Log("caught only by Node stdout; clean exit, no sanitizer error or leak")
		})
	}
}

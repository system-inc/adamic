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
	for _, name := range []string{"numbers", "order", "live", "set_live"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/library_map_small_" + name + ".a", lowers: true})
	}
}

// Every runtime mutant must compile, exit 0 and stay sanitizer-clean. Only source Node
// decides whether its output is wrong; a native invariant failure does not kill these mutants.
func TestMapSmallMutants(t *testing.T) {
	for _, name := range []string{"nan", "zero", "promotion_order", "promotion_iterator"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := "numbers"
			if name == "promotion_order" {
				fixture = "order"
			}
			if name == "promotion_iterator" {
				fixture = "live"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_map_small_"+fixture+".a"))
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
				if filepath.Base(path) == "map.c" {
					before := string(source)
					var after string
					switch name {
					case "nan":
						after = strings.Replace(before, "same_key(map, entry->key, key)", "(!map->reference_keys && !map->boolean_keys && !map->maybe_number_keys ? entry->key.number == key.number : same_key(map, entry->key, key))", 1)
					case "zero":
						after = strings.Replace(before, "same_key(map, entry->key, key)", "(!map->reference_keys && !map->boolean_keys && !map->maybe_number_keys ? memcmp(&entry->key.number, &key.number, sizeof key.number) == 0 : same_key(map, entry->key, key))", 1)
					case "promotion_order":
						after = strings.Replace(before, "memcpy(grown, map->entries, map->used * sizeof *grown);", `memcpy(grown, map->entries, map->used * sizeof *grown);
            if (map->used > 1) { adamic_map_entry swap = grown[0]; grown[0] = grown[1]; grown[1] = swap; }`, 1)
					case "promotion_iterator":
						after = strings.Replace(before, "// used is read each time", "if (iterator->map->bucket_count != 0 && iterator->next == 1) { iterator->next++; }\n\t// used is read each time", 1)
					}
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
			t.Logf("%s: clean exit 0, ASan/UBSan/LSan clean, caught only by Node stdout", name)
		})
	}
}

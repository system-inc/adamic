package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Moving a cached address must fail at runtime, not merely at C compilation.
// Both constructor conventions use the same publish-after-adoption seam.
func TestCanonicalGraphCacheOrderMutant(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"canonical_graph.a", "canonical_graph_counted.a"} {
		t.Run(fixture, func(t *testing.T) {
			checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", fixture)})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			source := C(program)
			constructor := "adamic_closure_canonical_graph"
			if strings.Contains(fixture, "counted") {
				constructor = "adamic_counted_closure_canonical_graph"
			}
			if !strings.Contains(source, constructor+"(") {
				t.Fatalf("fixture does not exercise %s", constructor)
			}
			files, err := readRuntime(runtime, "runtime")
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			changed := false
			for _, file := range files {
				contents := file.contents
				if file.name == "closure.c" {
					original := string(contents)
					adoption := "\tif (graph) {\n\t\tclosure = adamic_graph_adopt_owned(closure, sizeof *closure + count * sizeof closure->cells[0]);\n\t}\n"
					publish := "\t*functions = closure;\n"
					if strings.Count(original, adoption) != 1 || strings.Count(original, publish) != 1 {
						t.Fatal("cache/adoption mutation seam moved")
					}
					mutant := strings.Replace(original, adoption, "", 1)
					mutant = strings.Replace(mutant, publish, publish+adoption, 1)
					contents = []byte(mutant)
					changed = true
				}
				if err := os.WriteFile(filepath.Join(directory, file.name), contents, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if !changed {
				t.Fatal("mutant did not change the runtime")
			}
			options := Options{Sanitize: true}
			library, err := RuntimeLibraryForSource(directory, source, options)
			if err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(directory, "main.c")
			if err := os.WriteFile(main, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "mutant")
			arguments := append(LinkFlags(options), "-I", filepath.Dir(library), "-o", binary, main)
			arguments = append(arguments, RuntimeLinkFlags(library)...)
			arguments = append(arguments, "-lm")
			if output, err := exec.Command(compilerName(options), arguments...).CombinedOutput(); err != nil {
				t.Fatalf("mutant must compile: %v\n%s", err, output)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "ERROR: AddressSanitizer: heap-use-after-free") {
				t.Fatalf("cache-before-adoption mutant escaped: %v\n%s", err, output)
			}
			t.Log("cache-before-adoption mutant compiled and was caught by ASan heap-use-after-free")
		})
	}
}

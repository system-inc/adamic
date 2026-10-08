package css

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These mutations affect only temporary generated C. The release mutant must
// preserve output and pass ASan/UBSan with leak detection off. Only LSan kills it.
// The access and arithmetic probes likewise preserve ordinary native output.
func TestComposedMemoryChecksCanFail(t *testing.T) {
	if runtime.GOOS != "linux" {
		// census: not-applicable Requires Linux sanitizer harness.
		t.Skip("Linux sanitizer harness proof")
	}
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, []byte(">Ca{color:red;margin:0.50px}\n>Sa{--x:{b:c};color:rgb(1,2,3)}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := printerAnswers(t, cases, "default")
	main, _ := filepath.Abs("print_main.ts")
	program := lowered(t, main)
	source := native.C(program)
	for _, mutation := range []struct{ name, body, report string }{
		{"address sanitizer", `static void css_memory_probe(int argc) {
    unsigned char *storage = malloc((size_t)argc + 16);
    if (storage == NULL) abort();
    volatile unsigned char *observed = storage;
    observed[0] = 65;
    free(storage);
    volatile unsigned char value = observed[0];
    (void)value;
  }
`, "AddressSanitizer: heap-use-after-free"},
		{"undefined behavior sanitizer", `static void css_memory_probe(int argc) {
    volatile int largest = 2147483647;
    volatile int wrapped = largest + argc;
    (void)wrapped;
  }
`, "runtime error: signed integer overflow"},
		{"leak sanitizer", "", "LeakSanitizer"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			code := source
			if mutation.body == "" {
				code = strings.ReplaceAll(code, "adamic_release(", "(void)(")
				if code == source {
					t.Fatal("release mutant changed nothing")
				}
			} else {
				anchor := "int main(int argc, char **argv) {\n"
				if strings.Count(code, anchor) != 1 {
					t.Fatal("main anchor changed")
				}
				code = "#include <stdlib.h>\n" + mutation.body + strings.Replace(code, anchor, anchor+"\tcss_memory_probe(argc);\n", 1)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			ordinary := execute(t, nil, binary, cases)
			if ordinary.exitCode != 0 || len(ordinary.stderr) != 0 || string(ordinary.stdout) != expected {
				t.Fatalf("ordinary output must survive: %d %s %s", ordinary.exitCode, ordinary.stderr, firstDifference(string(ordinary.stdout), expected))
			}
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			options := "ASAN_OPTIONS=detect_leaks=0"
			if mutation.body == "" {
				clean := execute(t, []string{options}, binary, cases)
				if clean.exitCode != 0 || len(clean.stderr) != 0 || string(clean.stdout) != expected {
					t.Fatalf("only leak detection should catch omitted releases: %d %s", clean.exitCode, clean.stderr)
				}
				options = "ASAN_OPTIONS=detect_leaks=1"
			}
			caught := execute(t, []string{options}, binary, cases)
			if caught.exitCode == 0 || !strings.Contains(string(caught.stderr), mutation.report) {
				t.Fatalf("memory mutant survived: %d %s", caught.exitCode, caught.stderr)
			}
			for _, line := range strings.Split(string(caught.stderr), "\n") {
				if strings.Contains(line, "SUMMARY:") || strings.Contains(line, "runtime error:") {
					t.Logf("only %s caught: %s", mutation.name, line)
				}
			}
		})
	}
}

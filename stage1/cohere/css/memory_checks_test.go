package css

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// These mutations affect only temporary generated C. The release mutant must
// preserve output and pass ASan/UBSan with leak detection off. Only LSan kills it.
// The access and arithmetic probes likewise preserve ordinary native output.
var composedMemoryMutations = []struct{ name, body, report string }{
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
}

// A memory mutant's two natives, ordinary and sanitized, built from the printer's emitted C with the mutation applied.
type composedMemoryNatives struct{ ordinary, sanitized string }

var composedMemoryOnce [3]sync.Once
var composedMemoryProducts [3]composedMemoryNatives

// composedMemoryProduct builds one mutation's natives ahead, as a product: the printer's lowering and four clang builds
// filled TestComposedMemoryChecksCanFail's 57.6 s alone on one core (Home), and each check now only runs them. Its key
// is what the printer's lowering reads, as cssPrinterProduct's is, with the mutation and both builds' flags.
func composedMemoryProduct(t *testing.T, index int) composedMemoryNatives {
	t.Helper()
	composedMemoryOnce[index].Do(func() {
		mutation := composedMemoryMutations[index]
		inputs := buildcache.Inputs{
			Name:      "css-composed-memory-mutant-" + strconv.Itoa(index),
			Files:     cssLoweringFiles,
			Flags:     append(append(append([]string{"mutation=" + mutation.name}, native.Flags(native.Options{})...), native.Flags(native.Options{Sanitize: true})...), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			main, err := filepath.Abs("print_main.ts")
			if err != nil {
				return err
			}
			checked, err := load.Load([]string{main})
			if err != nil {
				return err
			}
			program, err := lower.Lower(t.Context(), checked)
			if err != nil {
				return err
			}
			source := native.C(program)
			code := source
			if mutation.body == "" {
				code = strings.ReplaceAll(code, "adamic_release(", "(void)(")
				if code == source {
					return errors.New("release mutant changed nothing")
				}
			} else {
				anchor := "int main(int argc, char **argv) {\n"
				if strings.Count(code, anchor) != 1 {
					return errors.New("main anchor changed")
				}
				code = "#include <stdlib.h>\n" + mutation.body + strings.Replace(code, anchor, anchor+"\tcss_memory_probe(argc);\n", 1)
			}
			if err := native.Build(code, filepath.Join(dir, "ordinary"), native.Options{}); err != nil {
				return err
			}
			return native.Build(code, filepath.Join(dir, "sanitized"), native.Options{Sanitize: true})
		})
		composedMemoryProducts[index] = composedMemoryNatives{filepath.Join(directory, "ordinary"), filepath.Join(directory, "sanitized")}
	})
	if composedMemoryProducts[index].ordinary == "" {
		t.Fatalf("composed memory mutant %d unavailable: its build failed earlier in this process (see that test's log)", index)
	}
	return composedMemoryProducts[index]
}

// composedMemoryCheck runs one mutation's natives: the ordinary build keeps the printer's output, and only its
// sanitizer catches the sanitized build.
func composedMemoryCheck(t *testing.T, index int) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux sanitizer harness proof")
	}
	mutation := composedMemoryMutations[index]
	natives := composedMemoryProduct(t, index)
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, []byte(">Ca{color:red;margin:0.50px}\n>Sa{--x:{b:c};color:rgb(1,2,3)}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := printerAnswers(t, cases, "default")
	ordinary := execute(t, nil, natives.ordinary, cases)
	if ordinary.exitCode != 0 || len(ordinary.stderr) != 0 || string(ordinary.stdout) != expected {
		t.Fatalf("ordinary output must survive: %d %s %s", ordinary.exitCode, ordinary.stderr, firstDifference(string(ordinary.stdout), expected))
	}
	options := "ASAN_OPTIONS=detect_leaks=0"
	if mutation.body == "" {
		clean := execute(t, []string{options}, natives.sanitized, cases)
		if clean.exitCode != 0 || len(clean.stderr) != 0 || string(clean.stdout) != expected {
			t.Fatalf("only leak detection should catch omitted releases: %d %s", clean.exitCode, clean.stderr)
		}
		options = "ASAN_OPTIONS=detect_leaks=1"
	}
	caught := execute(t, []string{options}, natives.sanitized, cases)
	if caught.exitCode == 0 || !strings.Contains(string(caught.stderr), mutation.report) {
		t.Fatalf("memory mutant survived: %d %s", caught.exitCode, caught.stderr)
	}
	for _, line := range strings.Split(string(caught.stderr), "\n") {
		if strings.Contains(line, "SUMMARY:") || strings.Contains(line, "runtime error:") {
			t.Logf("only %s caught: %s", mutation.name, line)
		}
	}
}

func TestProduct_CSSComposedMemoryAddressMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux sanitizer harness proof")
	}
	composedMemoryProduct(t, 0)
}
func TestProduct_CSSComposedMemoryUndefinedBehaviorMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux sanitizer harness proof")
	}
	composedMemoryProduct(t, 1)
}
func TestProduct_CSSComposedMemoryLeakMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux sanitizer harness proof")
	}
	composedMemoryProduct(t, 2)
}

func TestComposedMemoryChecksCanFail_AddressSanitizer(t *testing.T) {
	t.Parallel()
	composedMemoryCheck(t, 0)
}
func TestComposedMemoryChecksCanFail_UndefinedBehaviorSanitizer(t *testing.T) {
	t.Parallel()
	composedMemoryCheck(t, 1)
}
func TestComposedMemoryChecksCanFail_LeakSanitizer(t *testing.T) {
	t.Parallel()
	composedMemoryCheck(t, 2)
}

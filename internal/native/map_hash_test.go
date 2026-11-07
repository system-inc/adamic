package native

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Compile the probe in map.c's own translation unit so it measures the private hash, the real
// insertion algorithm and the runtime's own growing table sizes, without exporting a test API.
func mapHashProbe(test *testing.T, options Options, mutation func(string) string) (string, error) {
	test.Helper()
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		test.Fatal(readRuntimeError)
	}
	harness, readError := os.ReadFile("testdata/map_hash.c")
	if readError != nil {
		test.Fatal(readError)
	}
	for index := range files {
		if files[index].name == "map.c" {
			files[index].contents = append(append([]byte{}, files[index].contents...), harness...)
		}
		if files[index].name == "map_set.c" && mutation != nil {
			source := string(files[index].contents)
			const signature = "uint64_t adamic_map_number_hash(double number) {"
			if strings.Count(source, signature) != 1 {
				test.Fatal("numeric hash definition must occur exactly once")
			}
			start := strings.Index(source, signature) + len(signature) - 1
			end, depth := start, 0
			for ; end < len(source); end++ {
				switch source[end] {
				case '{':
					depth++
				case '}':
					depth--
				}
				if depth == 0 {
					end++
					break
				}
			}
			if depth != 0 {
				test.Fatal("numeric hash definition has unmatched braces")
			}
			body := source[start:end]
			changed := mutation(body)
			if changed == body {
				test.Fatal("mutant changed no numeric hash code")
			}
			files[index].contents = []byte(source[:start] + changed + source[end:])
		}
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	flags := Flags(options)
	library, cachedRuntimeError := cachedRuntime(files, flags, compiler, "map hash probe", test.TempDir())
	if cachedRuntimeError != nil {
		test.Fatal(cachedRuntimeError)
	}
	binary := filepath.Join(test.TempDir(), "probe")
	arguments := append(append([]string{}, flags...), RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm", "-o", binary)
	if output, combinedOutputError := toolchain.Command(
		compiler,
		arguments...,
	).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("link probe: %v\n%s", combinedOutputError, output)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(timeoutContext, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	return stderr.String() + stdout.String(), runError
}

// The probe's two allocators, by name. Each probe builds and runs the runtime on its own, 10 s alone on one core, so
// each is its own top-level test below, and TestMapHashAllocatorsEachHaveOneTest holds the two in step.
var mapHashProbeAllocators = map[string]bool{"malloc": false, "slabs": true}

func TestMapHashProbeBoundMalloc(test *testing.T) {
	test.Parallel()
	runMapHashProbeBound(test, "malloc")
}

func TestMapHashProbeBoundSlabs(test *testing.T) {
	test.Parallel()
	runMapHashProbeBound(test, "slabs")
}

// An allocator added to the table without its test, or a test naming an allocator that isn't there, fails here
// without building anything.
func TestMapHashAllocatorsEachHaveOneTest(test *testing.T) {
	test.Parallel()
	checkNamedWrappers(
		test,
		"map_hash_test.go",
		"TestMapHashProbeBound",
		"runMapHashProbeBound",
		mapHashProbeAllocators,
	)
}

func runMapHashProbeBound(test *testing.T, name string) {
	test.Helper()
	slabs, ok := mapHashProbeAllocators[name]
	if !ok {
		test.Fatalf("no map hash probe allocator %q", name)
	}
	output, mapHashProbeError := mapHashProbe(test, Options{Sanitize: true, slabs: slabs}, nil)
	test.Log(output)
	if mapHashProbeError != nil {
		test.Fatalf("runtime map hash: %v", mapHashProbeError)
	}
}

// The numeric hash's mutants, by name. Each probe builds and runs the runtime on its own, as much
// as one test should carry alone, so each is its own top-level test below, and
// TestMapHashMutantsEachHaveOneTest holds the two in step.
var mapHashMutants = map[string]struct {
	before, after, caught string
}{
	// Restore the old finite-number hash, retaining SameValueZero normalization so only
	// the distribution check catches it. Empty before means replace the entire function body.
	"old hash": {"", `{
	if (number == 0) { number = 0; }
	if (isnan(number)) { return 0x7ff8000000000000ull; }
	uint64_t bits;
	memcpy(&bits, &number, sizeof bits);
	return (bits ^ (bits >> 29)) * 1099511628211ull;
}`, "probe bound 64 exceeded: integers"},
	// V8 normalizes -0 through its integer conversion. Send only -0 down the raw-double
	// path to remove that normalization without changing any other integer's hash.
	"zero normalization": {
		"if (number >= INT32_MIN",
		"if (!(number == 0 && signbit(number)) && number >= INT32_MIN",
		"zero hashes differ",
	},
	"NaN normalization": {"isnan(number)", "false", "NaN hashes differ"},
}

func TestMapHashProbeCatchesMutantsOldHash(test *testing.T) {
	test.Parallel()
	runMapHashMutant(test, "old hash")
}

func TestMapHashProbeCatchesMutantsZeroNormalization(test *testing.T) {
	test.Parallel()
	runMapHashMutant(test, "zero normalization")
}

func TestMapHashProbeCatchesMutantsNaNNormalization(test *testing.T) {
	test.Parallel()
	runMapHashMutant(test, "NaN normalization")
}

// A mutant added to the table without its test, or a test naming a mutant that isn't there,
// fails here without building anything.
func TestMapHashMutantsEachHaveOneTest(test *testing.T) {
	test.Parallel()
	checkNamedWrappers(test, "map_hash_test.go", "TestMapHashProbeCatchesMutants", "runMapHashMutant", mapHashMutants)
}

func runMapHashMutant(test *testing.T, name string) {
	test.Helper()
	mutant, ok := mapHashMutants[name]
	if !ok {
		test.Fatalf("no map hash mutant %q", name)
	}
	output, mapHashProbeError := mapHashProbe(test, Options{Sanitize: true}, func(source string) string {
		if mutant.before == "" {
			return mutant.after
		}
		if strings.Count(source, mutant.before) != 1 {
			test.Fatal("mutant must change exactly one site")
		}
		return strings.Replace(source, mutant.before, mutant.after, 1)
	})
	var exit *exec.ExitError
	if !errors.As(mapHashProbeError, &exit) ||
		exit.ExitCode() != 1 ||
		!strings.Contains(output, mutant.caught) ||
		strings.Contains(output, "ERROR: AddressSanitizer") ||
		strings.Contains(output, "runtime error:") {
		test.Fatalf("mutant escaped or failed for another reason: %v\n%s", mapHashProbeError, output)
	}
	for _, line := range strings.Split(output, "\n") {
		if name == "old hash" && strings.Contains(line, "entries=2048 buckets=") {
			test.Log(line)
		}
		if strings.Contains(line, mutant.caught) {
			test.Log(line)
		}
	}
}

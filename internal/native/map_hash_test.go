package native

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
)

// Compile the probe in map.c's own translation unit so it measures the private hash, the real
// insertion algorithm and the runtime's own growing table sizes, without exporting a test API.
func mapHashProbe(t *testing.T, options Options, mutation func(string) string) (string, error) {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	harness, err := os.ReadFile("testdata/map_hash.c")
	if err != nil {
		t.Fatal(err)
	}
	for index := range files {
		if files[index].name == "map.c" {
			files[index].contents = append(append([]byte{}, files[index].contents...), harness...)
		}
		if files[index].name == "map_set.c" && mutation != nil {
			source := string(files[index].contents)
			const signature = "uint64_t adamic_map_number_hash(double number) {"
			if strings.Count(source, signature) != 1 {
				t.Fatal("numeric hash definition must occur exactly once")
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
				t.Fatal("numeric hash definition has unmatched braces")
			}
			body := source[start:end]
			changed := mutation(body)
			if changed == body {
				t.Fatal("mutant changed no numeric hash code")
			}
			files[index].contents = []byte(source[:start] + changed + source[end:])
		}
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	flags := Flags(options)
	library, err := cachedRuntime(files, flags, compiler, "map hash probe", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "probe")
	arguments := append(append([]string{}, flags...), RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm", "-o", binary)
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("link probe: %v\n%s", err, output)
	}
	command := exec.Command(binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = childguard.Run(command, nativeGuardOptions)
	return stderr.String() + stdout.String(), err
}

func TestMapHashProbeBound(t *testing.T) {
	t.Parallel()
	for _, slabs := range []bool{false, true} {
		t.Run(map[bool]string{false: "malloc", true: "slabs"}[slabs], func(t *testing.T) {
			output, err := mapHashProbe(t, Options{Sanitize: true, slabs: slabs}, nil)
			t.Log(output)
			if err != nil {
				t.Fatalf("runtime map hash: %v", err)
			}
		})
	}
}

func TestMapHashProbeCatchesMutants(t *testing.T) {
	t.Parallel()
	mutants := []struct {
		name, before, after, caught string
	}{
		// Restore the old finite-number hash, retaining SameValueZero normalization so only
		// the distribution check catches it. Empty before means replace the entire function body.
		{"old hash", "", `{
	if (number == 0) { number = 0; }
	if (isnan(number)) { return 0x7ff8000000000000ull; }
	uint64_t bits;
	memcpy(&bits, &number, sizeof bits);
	return (bits ^ (bits >> 29)) * 1099511628211ull;
}`, "probe bound 64 exceeded: integers"},
		// V8 normalizes -0 through its integer conversion. Send only -0 down the raw-double
		// path to remove that normalization without changing any other integer's hash.
		{"zero normalization", "if (number >= INT32_MIN", "if (!(number == 0 && signbit(number)) && number >= INT32_MIN", "zero hashes differ"},
		{"NaN normalization", "isnan(number)", "false", "NaN hashes differ"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			output, err := mapHashProbe(t, Options{Sanitize: true}, func(source string) string {
				if mutant.before == "" {
					return mutant.after
				}
				if strings.Count(source, mutant.before) != 1 {
					t.Fatal("mutant must change exactly one site")
				}
				return strings.Replace(source, mutant.before, mutant.after, 1)
			})
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(output, mutant.caught) || strings.Contains(output, "ERROR: AddressSanitizer") || strings.Contains(output, "runtime error:") {
				t.Fatalf("mutant escaped or failed for another reason: %v\n%s", err, output)
			}
			for _, line := range strings.Split(output, "\n") {
				if mutant.name == "old hash" && strings.Contains(line, "entries=2048 buckets=") {
					t.Log(line)
				}
				if strings.Contains(line, mutant.caught) {
					t.Log(line)
				}
			}
		})
	}
}

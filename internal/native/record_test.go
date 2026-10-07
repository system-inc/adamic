package native

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const recordFixtures = "testdata/records"

func recordHarness(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(recordFixtures, "harness.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func recordRun(binary string, arguments ...string) (string, string, error) {
	command := exec.Command(binary, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=23", "UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	return stdout.String(), stderr.String(), err
}

func recordNode(t *testing.T, arguments ...string) string {
	t.Helper()
	arguments = append([]string{filepath.Join(recordFixtures, "oracle.js")}, arguments...)
	stdout, stderr, err := recordRun("node", arguments...)
	if err != nil || stderr != "" {
		t.Fatalf("Node: %v\n%s", err, stderr)
	}
	return stdout
}

var recordCounts = regexp.MustCompile(`^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`)

func recordCheckCounts(t *testing.T, stderr string) {
	t.Helper()
	counts := recordCounts.FindStringSubmatch(stderr)
	if counts == nil {
		t.Fatalf("want only a counts report, got:\n%s", stderr)
	}
	if counts[1] != counts[2] || counts[6] != "0" {
		t.Fatalf("heap values not balanced: %s", stderr)
	}
	t.Log(strings.TrimSpace(stderr))
}

// Every successful fixture is counted and run with ASan, UBSan and Linux LeakSanitizer enabled.
// The oracle uses Node's own Object operations and for...in, never the runtime's sort algorithm.
func TestRecordsAgainstNode(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "records")
	if err := Build(recordHarness(t), binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"semantics"}, {"prototypes"}, {"references"}, {"iteration"}, {"numeric"}, {"workload", "1000000"}, {"bench", "1000"}} {
		t.Run(strings.Join(arguments, "-"), func(t *testing.T) {
			stdout, stderr, err := recordRun(binary, arguments...)
			if err != nil {
				t.Fatalf("native: %v\n%s", err, stderr)
			}
			recordCheckCounts(t, stderr)
			want := recordNode(t, arguments...)
			if arguments[0] == "bench" {
				// Timing is observational. Only workload results are held byte for byte.
				stdout = strings.Split(stdout, "\n")[0]
				want = strings.Split(want, "\n")[0]
			}
			if stdout != want {
				t.Fatalf("Node comparison differs: native %d bytes, Node %d bytes; first difference at %d", len(stdout), len(want), recordDifference(stdout, want))
			}
		})
	}
	// A deliberate NotYet ends at a panic, like other runtime panics: leak checking at normal
	// exit does not apply. Check its exact message and exit, and reject sanitizer diagnostics.
	t.Run("proto-assignment", func(t *testing.T) {
		stdout, stderr, err := recordRun(binary, "proto-assignment")
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || stdout != "" || !strings.HasPrefix(stderr, "adamic: panic: NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property\n") || strings.Contains(stderr, "Sanitizer") {
			t.Fatalf("want explicit NotYet and exit 70: %v\n%s\n%s", err, stdout, stderr)
		}
	})
}

func recordDifference(a, b string) int {
	for index := 0; index < len(a) && index < len(b); index++ {
		if a[index] != b[index] {
			return index
		}
	}
	return min(len(a), len(b))
}

// Build a changed runtime in an isolated cache, without editing the working tree. Each mutant
// changes production C, links successfully, and is rejected by the named external check.
func recordMutant(t *testing.T, file, before, after string) string {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for index := range files {
		if files[index].name == file {
			contents := string(files[index].contents)
			if !strings.Contains(contents, before) {
				t.Fatalf("mutant input missing in %s", file)
			}
			files[index].contents = []byte(strings.Replace(contents, before, after, 1))
			found = true
		}
	}
	if !found {
		t.Fatalf("runtime file missing: %s", file)
	}
	directory := t.TempDir()
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	flags := Flags(Options{Sanitize: true, Count: true})
	library, err := cachedRuntime(files, flags, compiler, "record mutant", filepath.Join(directory, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	source, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "main")
	if err := os.WriteFile(source, []byte(recordHarness(t)), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append(flags, "-I", filepath.Dir(library), "-o", binary, source)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("mutant must compile: %v\n%s", err, output)
	}
	return binary
}

func TestRecordMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, file, before, after, mode, caught string
	}{
		{"indices-in-insertion-order", "record.c", "qsort(indices, count, sizeof *indices, compare_indices);", "(void)compare_indices;", "semantics", "Node"},
		{"uint32-max-as-index", "record.c", "number >= UINT32_MAX", "number > UINT32_MAX", "semantics", "Node"},
		{"deleted-key-iterated", "record.c", "if (slot != NULL) {\n\t\t\t*key = candidate;\n\t\t\t*value = *slot;", "if (true) {\n\t\t\t*key = candidate;\n\t\t\t*value = slot == NULL ? (adamic_value){.number = 0} : *slot;", "iteration", "Node"},
		{"overwrite-key-leaked", "map.c", "adamic_release(key.reference);", "(void)key;", "references", "LeakSanitizer"},
		{"stored-key-freed", "record.c", "adamic_map_set(table(record), (adamic_value){.reference = key}, value);", "adamic_map_set(table(record), (adamic_value){.reference = key}, value);\n\tadamic_release(key);", "semantics", "AddressSanitizer: heap-use-after-free"},
		{"own-slot-null-read", "record.c", "return adamic_map_get(table(record), (adamic_value){.reference = (void *)key});", "adamic_value *missing = NULL;\n\tvolatile double observed = missing->number;\n\t(void)observed;\n\treturn adamic_map_get(table(record), (adamic_value){.reference = (void *)key});", "prototypes", "runtime error: member access within null pointer"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			binary := recordMutant(t, mutant.file, mutant.before, mutant.after)
			stdout, stderr, err := recordRun(binary, mutant.mode)
			if mutant.caught == "Node" {
				if err != nil {
					t.Fatalf("order mutant must finish without sanitizer failure: %v\n%s", err, stderr)
				}
				recordCheckCounts(t, stderr)
				if want := recordNode(t, mutant.mode); stdout == want {
					t.Fatal("Node comparison did not catch mutant")
				}
			} else if err == nil || !strings.Contains(stderr, mutant.caught) {
				t.Fatalf("want %s to catch mutant: %v\n%s", mutant.caught, err, stderr)
			}
			t.Logf("caught by %s", mutant.caught)
		})
	}
}

// Opt in to observations: timing has no pass threshold on a shared worker. Each of five fresh
// process pairs executes identical operations and consumes the result; best per operation is logged.
func TestRecordBenchmark(t *testing.T) {
	if os.Getenv("ADAMIC_RECORD_BENCH") != "1" {
		t.Skip("set ADAMIC_RECORD_BENCH=1 for five-round Node comparisons")
	}
	binary := filepath.Join(t.TempDir(), "records")
	if err := Build(recordHarness(t), binary, Options{}); err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{1000, 10000, 100000, 1000000} {
		var best [2][5]float64
		for side := range best {
			for operation := range best[side] {
				best[side][operation] = math.Inf(1)
			}
		}
		for round := 0; round < 5; round++ {
			var results [2]string
			for step := 0; step < 2; step++ {
				side := (step + round) % 2
				var stderr string
				var err error
				if side == 0 {
					results[side], stderr, err = recordRun(binary, "bench", strconv.Itoa(size))
				} else {
					results[side], stderr, err = recordRun("node", filepath.Join(recordFixtures, "oracle.js"), "bench", strconv.Itoa(size))
				}
				if err != nil || stderr != "" {
					t.Fatalf("benchmark side %d: %v\n%s", side, err, stderr)
				}
				lines := strings.Split(strings.TrimSpace(results[side]), "\n")
				if len(lines) != 2 {
					t.Fatalf("benchmark output: %q", results[side])
				}
				fields := strings.Fields(lines[1])
				if len(fields) != 5 {
					t.Fatal("missing benchmark timings")
				}
				for operation, field := range fields {
					duration, err := strconv.ParseFloat(field, 64)
					if err != nil || duration < 0 || math.IsNaN(duration) {
						t.Fatalf("invalid duration %q", field)
					}
					best[side][operation] = math.Min(best[side][operation], duration)
				}
				t.Logf("size=%d round=%d side=%d %s", size, round+1, side, strings.TrimSpace(results[side]))
			}
			if strings.Split(results[0], "\n")[0] != strings.Split(results[1], "\n")[0] {
				t.Fatal("benchmark work differs from Node")
			}
		}
		for operation, name := range []string{"insert", "hit", "miss", "delete-half", "iterate"} {
			t.Log(fmt.Sprintf("best-of-five size=%d %s native=%.6fs Node=%.6fs ratio=%.3f", size, name, best[0][operation], best[1][operation], best[0][operation]/best[1][operation]))
		}
	}
}

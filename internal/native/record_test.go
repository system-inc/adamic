package native_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
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
	command.Env = append(os.Environ(), "UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
	if runtime.GOOS == "linux" {
		command.Env = append(command.Env, "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1")
	}
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

// Every successful fixture is counted and sanitized, then checked by the shared leak helper.
// The oracle uses Node's own Object operations and for...in, never the runtime's sort algorithm.
func TestRecordsAgainstNode(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "records")
	if err := native.Build(recordHarness(t), binary, native.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"semantics"}, {"prototypes"}, {"reads"}, {"references"}, {"iteration"}, {"numeric"}, {"workload", "1000000"}, {"bench", "1000"}} {
		t.Run(strings.Join(arguments, "-"), func(t *testing.T) {
			stdout, stderr, err := recordRun(binary, arguments...)
			if err != nil {
				t.Fatalf("native: %v\n%s", err, stderr)
			}
			recordCheckCounts(t, stderr)
			if report := leakcheck.Report(t, recordHarness(t), binary, arguments...); report != "" {
				t.Fatal(report)
			}
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
	// The authoritative member list is read from Node rather than copied into this check.
	// Each operation must stop for every missing member, naming exactly the requested key.
	for _, name := range recordPrototypeNames(t) {
		for _, operation := range []string{"missing-get", "missing-has"} {
			t.Run(operation+"/"+name, func(t *testing.T) {
				stdout, stderr, err := recordRun(binary, operation, name)
				if !recordMemberStop(stdout, stderr, err, name) {
					t.Fatalf("want exact own-only stop for %s %q: %v\nstdout %q\nstderr %s", operation, name, err, stdout, stderr)
				}
			})
		}
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

func recordPrototypeNames(t *testing.T) []string {
	t.Helper()
	stdout, stderr, err := recordRun("node", "-e", `console.log(JSON.stringify({version:process.version,members:Object.getOwnPropertyNames(Object.prototype)}))`)
	if err != nil || stderr != "" {
		t.Fatalf("Node prototype names: %v\n%s", err, stderr)
	}
	var observation struct {
		Version string
		Members []string
	}
	if err := json.Unmarshal([]byte(stdout), &observation); err != nil || observation.Version == "" || len(observation.Members) == 0 {
		t.Fatalf("Node prototype observation: %v, %q", err, stdout)
	}
	t.Logf("Object.getOwnPropertyNames(Object.prototype) on Node %s: %s", observation.Version, strings.Join(observation.Members, ", "))
	return observation.Members
}

func recordMemberMessage(name string) string {
	return fmt.Sprintf("adamic: panic: record member '%s' is missing; records hold own keys only\n", name)
}

// Intentional stops do not reach normal-exit leak checking. Require the complete diagnostic
// followed only by the counted report, so a sanitizer failure cannot masquerade as this stop.
func recordMemberStop(stdout, stderr string, err error, name string) bool {
	exit, ok := err.(*exec.ExitError)
	message := recordMemberMessage(name)
	return ok && exit.ExitCode() == 70 && stdout == "" && strings.HasPrefix(stderr, message) &&
		recordCounts.MatchString(strings.TrimPrefix(stderr, message))
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
type recordMutantProgram struct {
	binary, directory string
}

func recordMutant(t *testing.T, file, before, after string) recordMutantProgram {
	t.Helper()
	directory := t.TempDir()
	entries, err := os.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		contents, err := os.ReadFile(filepath.Join("runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if !strings.Contains(string(contents), before) {
				t.Fatalf("mutant input missing in %s", file)
			}
			contents = []byte(strings.Replace(string(contents), before, after, 1))
			found = true
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if !found {
		t.Fatalf("runtime file missing: %s", file)
	}
	program := recordMutantProgram{filepath.Join(t.TempDir(), "mutant"), directory}
	if err := program.build(recordHarness(t), program.binary, native.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatalf("mutant must compile: %v", err)
	}
	return program
}

func (program recordMutantProgram) build(code, binary string, options native.Options) error {
	library, err := native.RuntimeLibrary(program.directory, options)
	if err != nil {
		return err
	}
	source := binary + ".c"
	if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
		return err
	}
	arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, source)
	arguments = append(arguments, native.RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	output, err := exec.Command("clang", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile: %w\n%s", err, output)
	}
	return nil
}

func recordMutantLeaks(t *testing.T, program recordMutantProgram, arguments ...string) string {
	t.Helper()
	report, err := leakcheck.Check(leakcheck.Program{
		C: recordHarness(t), Sanitized: program.binary, Counted: filepath.Join(t.TempDir(), "counted"),
		BuildCounted: func(code, output string) error { return program.build(code, output, native.Options{Count: true}) },
		Arguments:    func() []string { return arguments },
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, name, arguments...)
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
			command.WaitDelay = 5 * time.Second
			command.Env = append(os.Environ(), environment...)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatal(err)
				}
			}
			return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

var recordMutationCases = []struct {
	name, file, before, after, mode, caught string
}{
	{"indices-in-insertion-order", "record.c", "qsort(indices, count, sizeof *indices, compare_indices);", "(void)compare_indices;", "semantics", "Node"},
	{"uint32-max-as-index", "record.c", "number >= UINT32_MAX", "number > UINT32_MAX", "semantics", "Node"},
	{"deleted-key-iterated", "record.c", "if (slot != NULL) {\n\t\t\t*key = candidate;\n\t\t\t*value = *slot;", "if (true) {\n\t\t\t*key = candidate;\n\t\t\t*value = slot == NULL ? (adamic_value){.number = 0} : *slot;", "iteration", "Node"},
	{"overwrite-key-leaked", "map.c", "adamic_release(key.reference);", "(void)key;", "references", "LeakSanitizer"},
	{"stored-key-freed", "record.c", "adamic_map_set(table(record), (adamic_value){.reference = key}, value);", "adamic_map_set(table(record), (adamic_value){.reference = key}, value);\n\tadamic_release(key);", "semantics", "AddressSanitizer: heap-use-after-free"},
	{"own-slot-null-read", "record.c", "return adamic_map_get(table(record), (adamic_value){.reference = (void *)key});", "adamic_value *missing = NULL;\n\tvolatile double observed = missing->number;\n\t(void)observed;\n\treturn adamic_map_get(table(record), (adamic_value){.reference = (void *)key});", "prototypes", "runtime error: member access within null pointer"},
}

func runRecordMutantUnit(t *testing.T, unit int) {
	t.Helper()
	for index, mutant := range recordMutationCases {
		if index%6 != unit {
			continue
		}
		t.Run(mutant.name, func(t *testing.T) {
			program := recordMutant(t, mutant.file, mutant.before, mutant.after)
			if mutant.caught == "LeakSanitizer" {
				report := recordMutantLeaks(t, program, mutant.mode)
				if !strings.Contains(report, "LeakSanitizer") && !strings.Contains(report, "heap values leaked") {
					t.Fatalf("leak mutant survived: %s", report)
				}
				counted := filepath.Join(t.TempDir(), "counted-proof")
				if err := program.build(recordHarness(t), counted, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				stdout, stderr, err := recordRun(counted, mutant.mode)
				if err != nil {
					t.Fatalf("counted mutant failed: %v\n%s", err, stderr)
				}
				proof := leakcheck.Unbalanced(leakcheck.Run{Stdout: []byte(stdout), Stderr: []byte(stderr)})
				if !strings.Contains(proof, "heap values leaked") {
					t.Fatalf("counted rule missed mutant: %s", proof)
				}
				t.Logf("shared check: %s; counted check: %s", report, proof)
				return
			}
			stdout, stderr, err := recordRun(program.binary, mutant.mode)
			if mutant.caught == "Node" {
				if err != nil {
					t.Fatalf("order mutant must finish without sanitizer failure: %v\n%s", err, stderr)
				}
				recordCheckCounts(t, stderr)
				if report := recordMutantLeaks(t, program, mutant.mode); report != "" {
					t.Fatal(report)
				}
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

func TestRecordReadMutants(t *testing.T) {
	t.Parallel()
	const guardedMiss = "if (value == NULL) {\n\t\tcheck_missing_member(key);\n\t}"
	for _, mutant := range []struct {
		name, before, after, operation string
	}{
		// Restore JavaScript's inherited in result, which this own-only contract forbids.
		{"prototype-membership-restored", "return adamic_record_get(record, key) != NULL;", "return adamic_record_get_own(record, key) != NULL || prototype_member(key);", "missing-has"},
		{"missing-read-silent", guardedMiss, "(void)key;", "missing-get"},
		{"own-read-checked-as-missing", guardedMiss, "check_missing_member(key);", "reads"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program := recordMutant(t, "record.c", mutant.before, mutant.after)
			binary := program.binary
			if mutant.operation == "reads" {
				stdout, stderr, err := recordRun(binary, "reads")
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 70 || !strings.HasPrefix(stderr, recordMemberMessage("toString")) || stdout == recordNode(t, "reads") {
					t.Fatalf("want own prototype-name hit wrongly stopped: %v\n%s\n%s", err, stdout, stderr)
				}
				t.Log("caught by the own-hit fixture")
				return
			}
			stdout, stderr, err := recordRun(binary, mutant.operation, "toString")
			if err != nil {
				t.Fatalf("fallback mutant must finish without sanitizer failures: %v\n%s", err, stderr)
			}
			recordCheckCounts(t, stderr)
			if report := recordMutantLeaks(t, program, mutant.operation, "toString"); report != "" {
				t.Fatal(report)
			}
			if recordMemberStop(stdout, stderr, err, "toString") {
				t.Fatal("exact stop contract did not catch mutant")
			}
			want := "0\n"
			if mutant.operation == "missing-has" {
				want = "1\n"
			}
			if stdout != want {
				t.Fatalf("want mutant's forbidden result %q, got %q", want, stdout)
			}
			t.Logf("caught by exact diagnostic and exit check; forbidden result %q", stdout)
		})
	}
}

// Opt in to observations: timing has no pass threshold on a shared worker. Each of five fresh
// process pairs executes identical operations and consumes the result; best per operation is logged.
// Not parallel: measures native and Node process timings, requiring exclusive execution.
func TestRecordBenchmark(t *testing.T) {
	if os.Getenv("ADAMIC_RECORD_BENCH") != "1" {
		t.Skip("set ADAMIC_RECORD_BENCH=1 for five-round Node comparisons")
	}
	binary := filepath.Join(t.TempDir(), "records")
	if err := native.Build(recordHarness(t), binary, native.Options{}); err != nil {
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

func TestRecordMutantsUnit00(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 0)
}

func TestRecordMutantsUnit01(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 1)
}

func TestRecordMutantsUnit02(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 2)
}

func TestRecordMutantsUnit03(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 3)
}

func TestRecordMutantsUnit04(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 4)
}

func TestRecordMutantsUnit05(t *testing.T) {
	t.Parallel()
	runRecordMutantUnit(t, 5)
}

func TestRecordMutationMembership(t *testing.T) {
	t.Parallel()
	t.Run("record_mutants", func(t *testing.T) {
		if len(recordMutationCases) != 6 {
			t.Fatalf("got %d record mutants, want 6", len(recordMutationCases))
		}
		seen := map[string]bool{}
		for _, mutant := range recordMutationCases {
			if seen[mutant.name] {
				t.Fatalf("duplicate record mutant %s", mutant.name)
			}
			seen[mutant.name] = true
		}
	})
}

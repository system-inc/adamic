package native

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
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
	return recordOutput(command)
}

// recordNodeRun is recordRun for the node toolchain resolves, under its environment.
func recordNodeRun(arguments ...string) (string, string, error) {
	return recordOutput(toolchain.Command("node", arguments...))
}

func recordOutput(command *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	return stdout.String(), stderr.String(), err
}

func recordNode(t *testing.T, arguments ...string) string {
	t.Helper()
	arguments = append([]string{filepath.Join(recordFixtures, "oracle.js")}, arguments...)
	stdout, stderr, err := recordNodeRun(arguments...)
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

// recordPinnedCounts holds the workload's cost exactly, so a change that adds an allocation, a retain or
// a release per operation fails here and names the count, even when every answer and leak check holds.
// Linux is the gate of record: macOS's runtime takes other paths for the same fixture, so its rows aren't
// pinned. A count that moves on purpose is re-pinned from the Linux failure, which prints all four.
// Measured on Loom (Linux, internal/native#8/16) at 2b22b31c.
var recordPinnedCounts = map[string]struct{ allocations, retains, releases, peak uint64 }{
	"workload-1000000": {allocations: 3500004, retains: 500001, releases: 3000002, peak: 1000003},
}

func recordCheckPin(t *testing.T, name, stderr string) {
	t.Helper()
	pin, ok := recordPinnedCounts[name]
	if !ok {
		return
	}
	if goruntime.GOOS != "linux" {
		t.Logf("%s's counts are pinned on Linux only", name)
		return
	}
	counts := recordCounts.FindStringSubmatch(stderr)
	if counts == nil {
		t.Fatalf("want only a counts report, got:\n%s", stderr)
	}
	measured := make([]uint64, 6)
	for index := range measured {
		value, err := strconv.ParseUint(counts[index+1], 10, 64)
		if err != nil {
			t.Fatalf("count %d: %v", index, err)
		}
		measured[index] = value
	}
	for _, check := range []struct {
		name          string
		measured, pin uint64
	}{{"allocations", measured[0], pin.allocations}, {"retains", measured[2], pin.retains}, {"releases", measured[3], pin.releases}, {"peak", measured[4], pin.peak}} {
		if check.measured != check.pin {
			t.Errorf("%s: %s %d, pinned %d", name, check.name, check.measured, check.pin)
		}
	}
	if t.Failed() {
		t.Fatalf("%s's counts moved: allocations %d retains %d releases %d peak %d; a rise is a cost regression, an intended change re-pins these from Linux", name, measured[0], measured[2], measured[3], measured[4])
	}
}

// Every successful fixture is counted and run with ASan, UBSan and Linux LeakSanitizer enabled.
// The oracle uses Node's own Object operations and for...in, never the runtime's sort algorithm.
func TestRecordsAgainstNode(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "records")
	if err := Build(recordHarness(t), binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"semantics"}, {"prototypes"}, {"reads"}, {"references"}, {"iteration"}, {"numeric"}, {"two-indices"}, {"workload", "1000000"}, {"bench", "1000"}} {
		t.Run(strings.Join(arguments, "-"), func(t *testing.T) {
			stdout, stderr, err := recordRun(binary, arguments...)
			if err != nil {
				t.Fatalf("native: %v\n%s", err, stderr)
			}
			recordCheckCounts(t, stderr)
			recordCheckPin(t, strings.Join(arguments, "-"), stderr)
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
	// A record reaching a fixed-object read is a lowering bug, and must stop loudly rather than read
	// undefined or its own table: through an optional field, through a field named like its slot, and
	// through a uniform field slot's unchecked load.
	// The uniform slot load is checked because this harness builds sanitized and counted.
	for _, mode := range []struct{ name, field string }{{"fixed-optional-read", "total"}, {"fixed-read", "0"}, {"uniform-read", "total"}} {
		t.Run(mode.name, func(t *testing.T) {
			stdout, stderr, err := recordRun(binary, mode.name)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stdout != "" || !strings.HasPrefix(stderr, "adamic: panic: compiler bug: a record reached a fixed-object read of "+mode.field+"\n") || strings.Contains(stderr, "Sanitizer") {
				t.Fatalf("want a loud stop and exit 70: %v\nstdout %q\nstderr %s", err, stdout, stderr)
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

func recordPrototypeNames(t *testing.T) []string {
	t.Helper()
	stdout, stderr, err := recordNodeRun("-e", `console.log(JSON.stringify({version:process.version,members:Object.getOwnPropertyNames(Object.prototype)}))`)
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

// recordMutant is the harness built against a changed runtime, without editing the working tree: the
// mutant of that name in recordMutations, from the record-mutants product. Each mutant changes
// production C, links successfully, and is rejected by the named external check.
func recordMutant(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(recordMutants(t), name)
}

// recordMutation is one registered mutant's change to the runtime: before, in file, becomes after.
type recordMutation struct{ name, file, before, after string }

// recordMutations is every mutant a record test builds, from both tables: recordMutationCases, the
// mutant units', and recordReadMutationCases, TestRecordReadMutants'. A test reaches its binary by the
// mutant's name, and TestProduct_RecordMutants builds every one ahead under the key the tests fetch, so
// a mutant in neither table has no binary.
func recordMutations(t testing.TB) []recordMutation {
	t.Helper()
	var mutations []recordMutation
	for _, mutant := range recordMutationCases {
		mutations = append(mutations, recordMutation{mutant.name, mutant.file, mutant.before, mutant.after})
	}
	for _, mutant := range recordReadMutationCases {
		mutations = append(mutations, recordMutation{mutant.name, "record.c", mutant.before, mutant.after})
	}
	seen := map[string]bool{}
	for _, mutation := range mutations {
		if seen[mutation.name] {
			t.Fatalf("two record mutants are named %s, and each names its binary", mutation.name)
		}
		seen[mutation.name] = true
	}
	return mutations
}

// recordMutants is the record-mutants product, a product built ahead by TestProduct_RecordMutants: the
// harness linked against each registered mutant's runtime, a binary named for the mutant, built as
// cachedRuntime builds a runtime (Flags with Sanitize and Count, every translation unit). Every test
// used to build its own mutant's whole runtime into a cache under the home directory and link it
// through a cache keyed by the test binary's hash, which no product declares, so a runner that reads
// products never found it. The runtime's unchanged translation units are the same in every mutant, so
// they compile once; each mutant compiles only its changed file, and links every object directly,
// which is what the whole-archive link of cachedRuntime's archive did.
func recordMutants(t *testing.T) string {
	t.Helper()
	mutations := recordMutations(t)
	flags := Flags(Options{Sanitize: true, Count: true})
	inputs := buildcache.Inputs{
		Name:      "record-mutants",
		Files:     []string{"internal/native/runtime", "internal/native/testdata/records/harness.c", "internal/native/record_test.go", "internal/native/split_units_test.go", "internal/native/native.go", "internal/native/library.go"},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	for _, mutation := range mutations {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutant %s %s %x %x", mutation.name, mutation.file, sha256.Sum256([]byte(mutation.before)), sha256.Sum256([]byte(mutation.after))))
	}
	harness := recordHarness(t)
	return buildcache.Product(t, inputs, func(directory string) error {
		files, err := readRuntime(runtime, "runtime")
		if err != nil {
			return err
		}
		var sources []string
		for _, file := range files {
			if strings.HasSuffix(file.name, ".c") {
				sources = append(sources, file.name)
			}
		}
		scratch, err := os.MkdirTemp("", "adamic-record-mutants-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		// compile builds a directory's translation unit from inside it, mapped, so no object names the
		// scratch directory (cachedRuntime's way).
		compile := func(within, source string) error {
			arguments := append(append(slices.Clone(flags), prefixMaps(within, "/adamic-runtime")...), "-c", source, "-o", strings.TrimSuffix(source, ".c")+".o")
			command := toolchain.CommandIn(within, "clang", arguments...)
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("compiling runtime %s: %w\n%s", source, err, output)
			}
			return nil
		}
		write := func(within string, files []runtimeFile) error {
			if err := os.MkdirAll(within, 0o755); err != nil {
				return err
			}
			for _, file := range files {
				if err := os.WriteFile(filepath.Join(within, file.name), file.contents, 0o644); err != nil {
					return err
				}
			}
			return nil
		}
		unchanged := filepath.Join(scratch, "runtime")
		if err := write(unchanged, files); err != nil {
			return err
		}
		if err := eachInParallel(len(sources), func(index int) error { return compile(unchanged, sources[index]) }); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(harness), 0o644); err != nil {
			return err
		}
		return eachInParallel(len(mutations), func(index int) error {
			mutation := mutations[index]
			// The mutant's directory holds the runtime's headers and its changed file, so that file compiles
			// as it would among the rest.
			var changed []runtimeFile
			found := false
			for _, file := range files {
				if file.name == mutation.file {
					contents := string(file.contents)
					if !strings.Contains(contents, mutation.before) {
						return fmt.Errorf("mutant %s: input missing in %s", mutation.name, mutation.file)
					}
					changed = append(changed, runtimeFile{file.name, []byte(strings.Replace(contents, mutation.before, mutation.after, 1))})
					found = true
				} else if strings.HasSuffix(file.name, ".h") {
					changed = append(changed, file)
				}
			}
			if !found || !strings.HasSuffix(mutation.file, ".c") {
				return fmt.Errorf("mutant %s: runtime file missing: %s", mutation.name, mutation.file)
			}
			within := filepath.Join(scratch, "mutant-"+strconv.Itoa(index))
			if err := write(within, changed); err != nil {
				return err
			}
			if err := compile(within, mutation.file); err != nil {
				return fmt.Errorf("mutant %s: %w", mutation.name, err)
			}
			// Linked from the scratch directory by relative names, mapped, so the binary names neither it
			// nor the product.
			arguments := append(append(slices.Clone(flags), prefixMaps(scratch, "/adamic-build")...), "-I", "runtime", "-o", filepath.Join(directory, mutation.name), "main.c")
			for _, source := range sources {
				object := strings.TrimSuffix(source, ".c") + ".o"
				if source == mutation.file {
					arguments = append(arguments, filepath.Join(filepath.Base(within), object))
				} else {
					arguments = append(arguments, filepath.Join("runtime", object))
				}
			}
			arguments = append(arguments, "-lm")
			command := toolchain.CommandIn(scratch, "clang", arguments...)
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("mutant %s must compile: %w\n%s", mutation.name, err, output)
			}
			return nil
		})
	})
}

// TestProduct_RecordMutants builds every registered record mutant, so the tree build builds them ahead.
func TestProduct_RecordMutants(t *testing.T) {
	t.Parallel()
	recordMutants(t)
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
	shard := testShard{unit, 6}
	for index, mutant := range recordMutationCases {
		if !shard.owns(index) {
			continue
		}
		t.Run(mutant.name, func(t *testing.T) {
			binary := recordMutant(t, mutant.name)
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

const recordGuardedMiss = "if (value == NULL) {\n\t\tcheck_missing_member(key);\n\t}"

// recordReadMutationCases are TestRecordReadMutants' mutants of record.c, registered with
// recordMutationCases in recordMutations so TestProduct_RecordMutants builds them too.
var recordReadMutationCases = []struct {
	name, before, after, operation string
}{
	// Restore JavaScript's inherited in result, which this own-only contract forbids.
	{"prototype-membership-restored", "return adamic_record_get(record, key) != NULL;", "return adamic_record_get_own(record, key) != NULL || prototype_member(key);", "missing-has"},
	{"missing-read-silent", recordGuardedMiss, "(void)key;", "missing-get"},
	{"own-read-checked-as-missing", recordGuardedMiss, "check_missing_member(key);", "reads"},
}

func runRecordReadMutant(t *testing.T, unit int) {
	t.Helper()
	for index, mutant := range recordReadMutationCases {
		if index != unit {
			continue
		}
		t.Run(mutant.name, func(t *testing.T) {
			binary := recordMutant(t, mutant.name)
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
					results[side], stderr, err = recordNodeRun(filepath.Join(recordFixtures, "oracle.js"), "bench", strconv.Itoa(size))
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

func TestRecordReadMutantsUnit00(t *testing.T) {
	t.Parallel()
	runRecordReadMutant(t, 0)
}
func TestRecordReadMutantsUnit01(t *testing.T) {
	t.Parallel()
	runRecordReadMutant(t, 1)
}
func TestRecordReadMutantsUnit02(t *testing.T) {
	t.Parallel()
	runRecordReadMutant(t, 2)
}

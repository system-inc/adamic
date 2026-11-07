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
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

const recordFixtures = "testdata/records"

func recordHarness(test *testing.T) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join(recordFixtures, "harness.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func recordRun(binary string, arguments ...string) (string, string, error) {
	command := exec.Command(binary, arguments...)
	command.Env = append(
		os.Environ(),
		"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1",
		"LSAN_OPTIONS=exitcode=23",
		"UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1",
	)
	return recordOutput(command)
}

// recordNodeRun is recordRun for the node toolchain resolves, under its environment.
func recordNodeRun(arguments ...string) (string, string, error) {
	return recordOutput(toolchain.Command("node", arguments...))
}

func recordOutput(command *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	return stdout.String(), stderr.String(), runError
}

func recordNode(test *testing.T, arguments ...string) string {
	test.Helper()
	arguments = append([]string{filepath.Join(recordFixtures, "oracle.js")}, arguments...)
	stdout, stderr, recordNodeRunError := recordNodeRun(arguments...)
	if recordNodeRunError != nil || stderr != "" {
		test.Fatalf("Node: %v\n%s", recordNodeRunError, stderr)
	}
	return stdout
}

var recordCounts = regexp.MustCompile(
	`^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`,
)

func recordCheckCounts(test *testing.T, stderr string) {
	test.Helper()
	counts := recordCounts.FindStringSubmatch(stderr)
	if counts == nil {
		test.Fatalf("want only a counts report, got:\n%s", stderr)
	}
	if counts[1] != counts[2] || counts[6] != "0" {
		test.Fatalf("heap values not balanced: %s", stderr)
	}
	test.Log(strings.TrimSpace(stderr))
}

// recordPinnedCounts holds the workload's cost exactly, so a change that adds an allocation, a retain or
// a release per operation fails here and names the count, even when every answer and leak check holds.
// Linux is the gate of record: macOS's runtime takes other paths for the same fixture, so its rows aren't
// pinned. A count that moves on purpose is re-pinned from the Linux failure, which prints all four.
// Measured on Loom (Linux, internal/native#8/16) at 2b22b31c.
var recordPinnedCounts = map[string]struct{ allocations, retains, releases, peak uint64 }{
	"workload-1000000": {allocations: 3500004, retains: 500001, releases: 3000002, peak: 1000003},
}

func recordCheckPin(test *testing.T, name, stderr string) {
	test.Helper()
	pin, ok := recordPinnedCounts[name]
	if !ok {
		return
	}
	if goruntime.GOOS != "linux" {
		test.Logf("%s's counts are pinned on Linux only", name)
		return
	}
	counts := recordCounts.FindStringSubmatch(stderr)
	if counts == nil {
		test.Fatalf("want only a counts report, got:\n%s", stderr)
	}
	measured := make([]uint64, 6)
	for index := range measured {
		value, parseError := strconv.ParseUint(counts[index+1], 10, 64)
		if parseError != nil {
			test.Fatalf("count %d: %v", index, parseError)
		}
		measured[index] = value
	}
	for _, check := range []struct {
		name          string
		measured, pin uint64
	}{
		{"allocations", measured[0], pin.allocations},
		{"retains", measured[2], pin.retains},
		{"releases", measured[3], pin.releases},
		{"peak", measured[4], pin.peak},
	} {
		if check.measured != check.pin {
			test.Errorf("%s: %s %d, pinned %d", name, check.name, check.measured, check.pin)
		}
	}
	if test.Failed() {
		test.Fatalf(
			"%s's counts moved: allocations %d retains %d releases %d peak %d; a rise is a cost regression, an intended change re-pins these from Linux",
			name,
			measured[0],
			measured[2],
			measured[3],
			measured[4],
		)
	}
}

// recordHarnessBinary is the counted, sanitized harness every record comparison runs, built once for
// all of them in this process, in a directory of its own, since a test's own would go when that test ends.
var recordHarnessBinary = sync.OnceValues(func() (string, error) {
	source, readError := os.ReadFile(filepath.Join(recordFixtures, "harness.c"))
	if readError != nil {
		return "", readError
	}
	directory, temporaryDirectoryError := os.MkdirTemp("", "adamic-records-")
	if temporaryDirectoryError != nil {
		return "", temporaryDirectoryError
	}
	binary := filepath.Join(directory, "records")
	return binary, Build(string(source), binary, Options{Sanitize: true, Count: true})
})

func recordBinary(test *testing.T) string {
	test.Helper()
	binary, recordHarnessBinaryError := recordHarnessBinary()
	if recordHarnessBinaryError != nil {
		test.Fatal(recordHarnessBinaryError)
	}
	return binary
}

// recordAgainstNode runs one fixture natively and holds it to the oracle. Node runs beside the native
// binary rather than after it, so the million-key workload costs the slower of the two, not their sum.
func recordAgainstNode(test *testing.T, arguments ...string) {
	test.Helper()
	binary := recordBinary(test)
	var oracle struct {
		stdout, stderr string
		err            error
	}
	oracleDone := make(chan struct{})
	go func() {
		defer close(oracleDone)
		oracle.stdout, oracle.stderr, oracle.err = recordNodeRun(
			append([]string{filepath.Join(recordFixtures, "oracle.js")}, arguments...)...,
		)
	}()
	// A native failure still waits for Node, so no oracle outlives its test.
	defer func() { <-oracleDone }()
	stdout, stderr, recordRunError := recordRun(binary, arguments...)
	if recordRunError != nil {
		test.Fatalf("native: %v\n%s", recordRunError, stderr)
	}
	recordCheckCounts(test, stderr)
	recordCheckPin(test, strings.Join(arguments, "-"), stderr)
	<-oracleDone
	if oracle.err != nil || oracle.stderr != "" {
		test.Fatalf("Node: %v\n%s", oracle.err, oracle.stderr)
	}
	want := oracle.stdout
	if arguments[0] == "bench" {
		// Timing is observational. Only workload results are held byte for byte.
		stdout = strings.Split(stdout, "\n")[0]
		want = strings.Split(want, "\n")[0]
	}
	if stdout != want {
		test.Fatalf(
			"Node comparison differs: native %d bytes, Node %d bytes; first difference at %d",
			len(stdout),
			len(want),
			recordDifference(stdout, want),
		)
	}
}

// Every successful fixture is counted and run with ASan, UBSan and Linux LeakSanitizer enabled.
// The oracle uses Node's own Object operations and for...in, never the runtime's sort algorithm.
// The million-key workload is TestRecordWorkloadAgainstNode's, a test of its own for its time.
func TestRecordFixturesAgainstNode(test *testing.T) {
	test.Parallel()
	for _, arguments := range [][]string{
		{"semantics"},
		{"prototypes"},
		{"reads"},
		{"references"},
		{"iteration"},
		{"numeric"},
		{"two-indices"},
		{"hidden"},
		{"bench", "1000"},
	} {
		test.Run(strings.Join(arguments, "-"), func(test *testing.T) {
			recordAgainstNode(test, arguments...)
		})
	}
}

// The million-key workload, counted, sanitized, pinned on Linux and held to Node byte for byte.
func TestRecordWorkloadAgainstNode(test *testing.T) {
	test.Parallel()
	recordAgainstNode(test, "workload", "1000000")
}

// The authoritative member list is read from Node rather than copied into this check.
// Each operation must stop for every missing member, naming exactly the requested key.
func TestRecordMissingMembersStop(test *testing.T) {
	test.Parallel()
	binary := recordBinary(test)
	for _, name := range recordPrototypeNames(test) {
		for _, operation := range []string{"missing-get", "missing-has"} {
			test.Run(operation+"/"+name, func(test *testing.T) {
				stdout, stderr, recordRunError := recordRun(binary, operation, name)
				if !recordMemberStop(stdout, stderr, recordRunError, name) {
					test.Fatalf(
						"want exact own-only stop for %s %q: %v\nstdout %q\nstderr %s",
						operation,
						name,
						recordRunError,
						stdout,
						stderr,
					)
				}
			})
		}
	}
}

// A deliberate NotYet ends at a panic, like other runtime panics: leak checking at normal
// exit does not apply. Check its exact message and exit, and reject sanitizer diagnostics.
func TestRecordProtoAssignmentStops(test *testing.T) {
	test.Parallel()
	stdout, stderr, recordRunError := recordRun(recordBinary(test), "proto-assignment")
	exit, ok := recordRunError.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || stdout != "" || !strings.HasPrefix(
		stderr,
		"adamic: panic: NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property\n",
	) || strings.Contains(stderr, "Sanitizer") {
		test.Fatalf("want explicit NotYet and exit 70: %v\n%s\n%s", recordRunError, stdout, stderr)
	}
}

// The three record shapes (record.c) differ by address alone, and a run-time key read boxes a record's
// value by the kind its shape names, so a record made with each kind must read that kind back. A linker
// that folded identical constants (--icf=all, -fmerge-all-constants) would make them one shape, and this
// goes red (Runtime's should on #zhgxzjg).
func TestRecordShapesKeepTheirValueKinds(test *testing.T) {
	test.Parallel()
	stdout, stderr, recordRunError := recordRun(recordBinary(test), "value-kinds")
	if recordRunError != nil || stdout != "number boolean reference\n" {
		test.Fatalf("want each record shape's own kind: %v\nstdout %q\nstderr %s", recordRunError, stdout, stderr)
	}
}

// A record reaching a fixed-object read is a lowering bug, and must stop loudly rather than read
// undefined or its own table: through an optional field, through a field named like its slot, and
// through a uniform field slot's unchecked load, checked because this harness builds sanitized and counted.
func TestRecordFixedReadsStop(test *testing.T) {
	test.Parallel()
	binary := recordBinary(test)
	for _, mode := range []struct{ name, field string }{
		{"fixed-optional-read", "total"},
		{"fixed-read", "0"},
		{"uniform-read", "total"},
	} {
		test.Run(mode.name, func(test *testing.T) {
			stdout, stderr, recordRunError := recordRun(binary, mode.name)
			exit, ok := recordRunError.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stdout != "" || !strings.HasPrefix(
				stderr,
				"adamic: panic: compiler bug: a record reached a fixed-object read of "+mode.field+"\n",
			) || strings.Contains(stderr, "Sanitizer") {
				test.Fatalf("want a loud stop and exit 70: %v\nstdout %q\nstderr %s", recordRunError, stdout, stderr)
			}
		})
	}
}

func recordPrototypeNames(test *testing.T) []string {
	test.Helper()
	stdout, stderr, recordNodeRunError := recordNodeRun(
		"-e",
		`console.log(JSON.stringify({version:process.version,members:Object.getOwnPropertyNames(Object.prototype)}))`,
	)
	if recordNodeRunError != nil || stderr != "" {
		test.Fatalf("Node prototype names: %v\n%s", recordNodeRunError, stderr)
	}
	var observation struct {
		Version string
		Members []string
	}
	if decodeError := json.Unmarshal(
		[]byte(stdout),
		&observation,
	); decodeError != nil || observation.Version == "" || len(observation.Members) == 0 {
		test.Fatalf("Node prototype observation: %v, %q", decodeError, stdout)
	}
	test.Logf(
		"Object.getOwnPropertyNames(Object.prototype) on Node %s: %s",
		observation.Version,
		strings.Join(observation.Members, ", "),
	)
	return observation.Members
}

func recordMemberMessage(name string) string {
	return fmt.Sprintf("adamic: panic: record member '%s' is missing; records hold own keys only\n", name)
}

// Intentional stops do not reach normal-exit leak checking. Require the complete diagnostic
// followed only by the counted report, so a sanitizer failure cannot masquerade as this stop.
func recordMemberStop(stdout, stderr string, recordRunError error, name string) bool {
	exit, ok := recordRunError.(*exec.ExitError)
	message := recordMemberMessage(name)
	return ok && exit.ExitCode() == 70 && stdout == "" && strings.HasPrefix(stderr, message) &&
		recordCounts.MatchString(strings.TrimPrefix(stderr, message))
}

func recordDifference(left, right string) int {
	for index := 0; index < len(left) && index < len(right); index++ {
		if left[index] != right[index] {
			return index
		}
	}
	return min(len(left), len(right))
}

// recordMutant is the harness built against a changed runtime, without editing the working tree: the
// mutant of that name in recordMutations, from the record-mutants product. Each mutant changes
// production C, links successfully, and is rejected by the named external check.
func recordMutant(test *testing.T, name string) string {
	test.Helper()
	return filepath.Join(recordMutants(test), name)
}

// recordMutation is one registered mutant's change to the runtime: before, in file, becomes after.
type recordMutation struct{ name, file, before, after string }

// recordMutations is every mutant a record test builds, from both tables: recordMutationCases, the
// mutant units', and recordReadMutationCases, TestRecordReadMutants'. A test reaches its binary by the
// mutant's name, and TestProduct_RecordMutants builds every one ahead under the key the tests fetch, so
// a mutant in neither table has no binary.
func recordMutations(test testing.TB) []recordMutation {
	test.Helper()
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
			test.Fatalf("two record mutants are named %s, and each names its binary", mutation.name)
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
func recordMutants(test *testing.T) string {
	test.Helper()
	mutations := recordMutations(test)
	flags := Flags(Options{Sanitize: true, Count: true})
	inputs := buildcache.Inputs{
		Name: "record-mutants",
		Files: []string{
			"internal/native/runtime",
			"internal/native/testdata/records/harness.c",
			"internal/native/record_test.go",
			"internal/native/split_units_test.go",
			"internal/native/runtime_mutants_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
		},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	for _, mutation := range mutations {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf(
			"mutant %s %s %x %x",
			mutation.name,
			mutation.file,
			sha256.Sum256([]byte(mutation.before)),
			sha256.Sum256([]byte(mutation.after)),
		))
	}
	harness := recordHarness(test)
	return buildcache.Product(test, inputs, func(directory string) error {
		files, readRuntimeError := readRuntime(runtime, "runtime")
		if readRuntimeError != nil {
			return readRuntimeError
		}
		var sources []string
		for _, file := range files {
			if strings.HasSuffix(file.name, ".c") {
				sources = append(sources, file.name)
			}
		}
		scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-record-mutants-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		// compile builds a directory's translation unit from inside it, mapped, so no object names the
		// scratch directory (cachedRuntime's way).
		compile := func(within, source string) error {
			arguments := append(
				append(slices.Clone(flags), prefixMaps(within, "/adamic-runtime")...),
				"-c",
				source,
				"-o",
				strings.TrimSuffix(source, ".c")+".o",
			)
			command := runtimeMutantClang(within, arguments...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("compiling runtime %s: %w\n%s", source, combinedOutputError, output)
			}
			return nil
		}
		write := func(within string, files []runtimeFile) error {
			if createDirectoryError := os.MkdirAll(within, 0o755); createDirectoryError != nil {
				return createDirectoryError
			}
			for _, file := range files {
				if writeError := os.WriteFile(
					filepath.Join(within, file.name),
					file.contents,
					0o644,
				); writeError != nil {
					return writeError
				}
			}
			return nil
		}
		unchanged := filepath.Join(scratch, "runtime")
		if writeError := write(unchanged, files); writeError != nil {
			return writeError
		}
		if eachInParallelError := eachInParallel(
			len(sources),
			func(index int) error { return compile(unchanged, sources[index]) },
		); eachInParallelError != nil {
			return eachInParallelError
		}
		if writeError := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(harness), 0o644); writeError != nil {
			return writeError
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
					changed = append(
						changed,
						runtimeFile{file.name, []byte(strings.Replace(contents, mutation.before, mutation.after, 1))},
					)
					found = true
				} else if strings.HasSuffix(file.name, ".h") {
					changed = append(changed, file)
				}
			}
			if !found || !strings.HasSuffix(mutation.file, ".c") {
				return fmt.Errorf("mutant %s: runtime file missing: %s", mutation.name, mutation.file)
			}
			within := filepath.Join(scratch, "mutant-"+strconv.Itoa(index))
			if writeError := write(within, changed); writeError != nil {
				return writeError
			}
			if compileError := compile(within, mutation.file); compileError != nil {
				return fmt.Errorf("mutant %s: %w", mutation.name, compileError)
			}
			// Linked from the scratch directory by relative names, mapped, so the binary names neither it
			// nor the product.
			arguments := append(
				append(slices.Clone(flags), prefixMaps(scratch, "/adamic-build")...),
				"-I",
				"runtime",
				"-o",
				filepath.Join(directory, mutation.name),
				"main.c",
			)
			for _, source := range sources {
				object := strings.TrimSuffix(source, ".c") + ".o"
				if source == mutation.file {
					arguments = append(arguments, filepath.Join(filepath.Base(within), object))
				} else {
					arguments = append(arguments, filepath.Join("runtime", object))
				}
			}
			arguments = append(arguments, "-lm")
			command := runtimeMutantClang(scratch, arguments...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("mutant %s must compile: %w\n%s", mutation.name, combinedOutputError, output)
			}
			return nil
		})
	})
}

// TestProduct_RecordMutants builds every registered record mutant, so the tree build builds them ahead.
func TestProduct_RecordMutants(test *testing.T) {
	test.Parallel()
	recordMutants(test)
}

var recordMutationCases = []struct {
	name, file, before, after, mode, caught string
}{
	{
		"indices-in-insertion-order",
		"record.c",
		"qsort(indices, count, sizeof *indices, compare_indices);",
		"(void)compare_indices;",
		"semantics",
		"Node",
	},
	{"uint32-max-as-index", "record.c", "number >= UINT32_MAX", "number > UINT32_MAX", "semantics", "Node"},
	{
		"deleted-key-iterated",
		"record.c",
		"if (slot != NULL) {\n\t\t\t*key = candidate;\n\t\t\t*value = *slot;",
		"if (true) {\n\t\t\t*key = candidate;\n\t\t\t*value = slot == NULL ? (adamic_value){.number = 0} : *slot;",
		"iteration",
		"Node",
	},
	{"overwrite-key-leaked", "map.c", "adamic_release(key.reference);", "(void)key;", "references", "LeakSanitizer"},
	{
		"stored-key-freed",
		"record.c",
		"adamic_map_set(table(record), (adamic_value){.reference = key}, value);",
		"adamic_map_set(table(record), (adamic_value){.reference = key}, value);\n\tadamic_release(key);",
		"semantics",
		"AddressSanitizer: heap-use-after-free",
	},
	{
		"own-slot-null-read",
		"record.c",
		"return adamic_map_get(table(record), (adamic_value){.reference = (void *)key});",
		"adamic_value *missing = NULL;\n\tvolatile double observed = missing->number;\n\t(void)observed;\n\treturn adamic_map_get(table(record), (adamic_value){.reference = (void *)key});",
		"prototypes",
		"runtime error: member access within null pointer",
	},
	// A new entry keeps whatever its slot held: stale attributes in a slot compaction freed, and in a fresh one
	// the sanitized allocator's fill, which UBSan stops at the first attribute read, before the reuse case.
	{
		"attributes-not-reset",
		"map.c",
		"\t// The slot may be one compaction freed, a record key's with attributes among them.\n\tentry->hidden = false;\n\tentry->readonly = false;\n\tentry->fixed = false;\n",
		"",
		"hidden",
		"runtime error: load of value",
	},
	{
		"hidden-keys-listed",
		"record.c",
		"if (!entry->deleted && !entry->hidden && !array_index(entry->key.reference, &number)) {",
		"if (!entry->deleted && !array_index(entry->key.reference, &number)) {",
		"hidden",
		"Node",
	},
}

func runRecordMutantUnit(test *testing.T, unit int) {
	test.Helper()
	shard := testShard{unit, 6}
	for index, mutant := range recordMutationCases {
		if !shard.owns(index) {
			continue
		}
		test.Run(mutant.name, func(test *testing.T) {
			binary := recordMutant(test, mutant.name)
			stdout, stderr, recordRunError := recordRun(binary, mutant.mode)
			if mutant.caught == "Node" {
				if recordRunError != nil {
					test.Fatalf("order mutant must finish without sanitizer failure: %v\n%s", recordRunError, stderr)
				}
				recordCheckCounts(test, stderr)
				if want := recordNode(test, mutant.mode); stdout == want {
					test.Fatal("Node comparison did not catch mutant")
				}
			} else if recordRunError == nil || !strings.Contains(stderr, mutant.caught) {
				test.Fatalf("want %s to catch mutant: %v\n%s", mutant.caught, recordRunError, stderr)
			}
			test.Logf("caught by %s", mutant.caught)
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
	{
		"prototype-membership-restored",
		"return adamic_record_get(record, key) != NULL;",
		"return adamic_record_get_own(record, key) != NULL || prototype_member(key);",
		"missing-has",
	},
	{"missing-read-silent", recordGuardedMiss, "(void)key;", "missing-get"},
	{"own-read-checked-as-missing", recordGuardedMiss, "check_missing_member(key);", "reads"},
}

func runRecordReadMutant(test *testing.T, unit int) {
	test.Helper()
	for index, mutant := range recordReadMutationCases {
		if index != unit {
			continue
		}
		test.Run(mutant.name, func(test *testing.T) {
			binary := recordMutant(test, mutant.name)
			if mutant.operation == "reads" {
				stdout, stderr, recordRunError := recordRun(binary, "reads")
				exit, ok := recordRunError.(*exec.ExitError)
				if !ok ||
					exit.ExitCode() != 70 ||
					!strings.HasPrefix(stderr, recordMemberMessage("toString")) ||
					stdout == recordNode(test, "reads") {
					test.Fatalf(
						"want own prototype-name hit wrongly stopped: %v\n%s\n%s",
						recordRunError,
						stdout,
						stderr,
					)
				}
				test.Log("caught by the own-hit fixture")
				return
			}
			stdout, stderr, recordRunError := recordRun(binary, mutant.operation, "toString")
			if recordRunError != nil {
				test.Fatalf("fallback mutant must finish without sanitizer failures: %v\n%s", recordRunError, stderr)
			}
			recordCheckCounts(test, stderr)
			if recordMemberStop(stdout, stderr, recordRunError, "toString") {
				test.Fatal("exact stop contract did not catch mutant")
			}
			want := "0\n"
			if mutant.operation == "missing-has" {
				want = "1\n"
			}
			if stdout != want {
				test.Fatalf("want mutant's forbidden result %q, got %q", want, stdout)
			}
			test.Logf("caught by exact diagnostic and exit check; forbidden result %q", stdout)
		})
	}
}

// recordDuration is one timed operation as either side prints it, nine decimals of seconds. A reading
// at or below the last printed digit, 1 ns, measured nothing (Perry's report gave 0.0 ms in every
// cell), so it's refused rather than kept as the best of five.
func recordDuration(field string) (float64, error) {
	duration, parseError := strconv.ParseFloat(field, 64)
	if parseError != nil {
		return 0, fmt.Errorf("invalid duration %q", field)
	}
	if !(duration > 1e-9) {
		return 0, fmt.Errorf("measured nothing: %q is at or below the 1 ns both sides print to", field)
	}
	return duration, nil
}

// A timing kept as the best of five is one the clock could see, and a non-number is refused.
func TestRecordDurationRefusesWhatMeasuredNothing(test *testing.T) {
	test.Parallel()
	if duration, recordDurationError := recordDuration(
		"0.000154732",
	); recordDurationError != nil || duration != 0.000154732 {
		test.Fatalf("a real timing: %v %v", duration, recordDurationError)
	}
	for _, field := range []string{"0.000000000", "0.000000001", "-0.000000500", "NaN"} {
		if _, recordDurationError := recordDuration(
			field,
		); recordDurationError == nil || !strings.HasPrefix(recordDurationError.Error(), "measured nothing: ") {
			test.Fatalf("%q was kept: %v", field, recordDurationError)
		}
	}
	if _, recordDurationError := recordDuration("fast"); recordDurationError == nil {
		test.Fatal("a non-number was kept")
	}
}

// Opt in to observations: timing has no pass threshold on a shared worker. Each of five fresh
// process pairs executes identical operations and consumes the result; best per operation is logged.
// Not parallel: measures native and Node process timings, requiring exclusive execution.
func TestRecordBenchmark(test *testing.T) {
	if os.Getenv("ADAMIC_RECORD_BENCH") != "1" {
		test.Skip("set ADAMIC_RECORD_BENCH=1 for five-round Node comparisons")
	}
	binary := filepath.Join(test.TempDir(), "records")
	if buildError := Build(recordHarness(test), binary, Options{}); buildError != nil {
		test.Fatal(buildError)
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
					results[side], stderr, err = recordNodeRun(
						filepath.Join(recordFixtures, "oracle.js"),
						"bench",
						strconv.Itoa(size),
					)
				}
				if err != nil || stderr != "" {
					test.Fatalf("benchmark side %d: %v\n%s", side, err, stderr)
				}
				lines := strings.Split(strings.TrimSpace(results[side]), "\n")
				if len(lines) != 2 {
					test.Fatalf("benchmark output: %q", results[side])
				}
				fields := strings.Fields(lines[1])
				if len(fields) != 5 {
					test.Fatal("missing benchmark timings")
				}
				for operation, field := range fields {
					duration, recordDurationError := recordDuration(field)
					if recordDurationError != nil {
						test.Fatalf("side %d operation %d: %v", side, operation, recordDurationError)
					}
					best[side][operation] = math.Min(best[side][operation], duration)
				}
				test.Logf("size=%d round=%d side=%d %s", size, round+1, side, strings.TrimSpace(results[side]))
			}
			if strings.Split(results[0], "\n")[0] != strings.Split(results[1], "\n")[0] {
				test.Fatal("benchmark work differs from Node")
			}
		}
		for operation, name := range []string{"insert", "hit", "miss", "delete-half", "iterate"} {
			test.Log(fmt.Sprintf(
				"best-of-five size=%d %s native=%.6fs Node=%.6fs ratio=%.3f",
				size,
				name,
				best[0][operation],
				best[1][operation],
				best[0][operation]/best[1][operation],
			))
		}
	}
}

func TestRecordMutantsUnit00(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 0)
}

func TestRecordMutantsUnit01(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 1)
}

func TestRecordMutantsUnit02(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 2)
}

func TestRecordMutantsUnit03(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 3)
}

func TestRecordMutantsUnit04(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 4)
}

func TestRecordMutantsUnit05(test *testing.T) {
	test.Parallel()
	runRecordMutantUnit(test, 5)
}

func TestRecordReadMutantsUnit00(test *testing.T) {
	test.Parallel()
	runRecordReadMutant(test, 0)
}
func TestRecordReadMutantsUnit01(test *testing.T) {
	test.Parallel()
	runRecordReadMutant(test, 1)
}
func TestRecordReadMutantsUnit02(test *testing.T) {
	test.Parallel()
	runRecordReadMutant(test, 2)
}

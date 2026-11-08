package estree

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type recoverySetup struct {
	start  time.Time
	builds time.Duration
}

func beginRecoverySetup(test *testing.T) *recoverySetup {
	setup := &recoverySetup{start: time.Now()}
	cpu := recoveryCpu(test)
	test.Cleanup(func() { test.Logf("TEST totalCPU=%.6fs", (recoveryCpu(test) - cpu).Seconds()) })
	return setup
}
func (setup *recoverySetup) report(test *testing.T) {
	wall := time.Since(setup.start)
	test.Logf(
		"setup wall=%.6fs builds=%.6fs without-builds=%.6fs",
		wall.Seconds(),
		setup.builds.Seconds(),
		(wall - setup.builds).Seconds(),
	)
}
func recoveryProduct(test *testing.T, setup *recoverySetup, in buildcache.Inputs, build func(string) error) string {
	test.Helper()
	start := time.Now()
	dir := buildcache.Product(test, in, func(dir string) error {
		cold := time.Now()
		buildError := build(dir)
		test.Logf("BUILD %s cold wall=%.6fs", in.Name, time.Since(cold).Seconds())
		return buildError
	})
	setup.builds += time.Since(start)
	return dir
}

func recoveryBuild(test *testing.T, path string, sanitize bool) (string, string) {
	test.Helper()
	if !sanitize {
		test.Fatal("recovery native products require sanitizers")
	}
	setup := beginRecoverySetup(test)
	defer setup.report(test)
	return recoveryPort(test, setup, path)
}
func recoveryPort(test *testing.T, setup *recoverySetup, path string) (string, string) {
	test.Helper()
	inputs := recoveryCacheInputs(test, path)
	lowered := recoveryProduct(test, setup, inputs, func(dir string) error {
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(ir)), 0644); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	code, readError := os.ReadFile(filepath.Join(lowered, "port.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	inputs = buildcache.Inputs{
		Name: "estree-sanitized-native",
		Files: []string{
			"internal/native/runtime",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/clang_in.go",
			"stage1/cohere/estree/recovery_cache_test.go",
		},
		Flags: append(
			native.Flags(native.Options{Sanitize: true}),
			fmt.Sprintf("source-sha256=%x", sha256.Sum256(code)),
			"relative-source-names",
			"canonical-debug-prefix=/adamic-estree",
		),
		Toolchain: []string{
			runtime.GOOS,
			runtime.GOARCH,
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("getconf", "GNU_LIBC_VERSION"),
		},
	}
	binary := recoveryProduct(
		test,
		setup,
		inputs,
		func(dir string) error { return recoveryStableNative(root(test), dir, string(code)) },
	)
	return filepath.Join(binary, "port"), filepath.Join(lowered, "port.mjs")
}

type recoveryCase struct {
	id     string
	source string
}

func recoveryCases(prefix string, sources []string) []recoveryCase {
	out := make([]recoveryCase, len(sources))
	for index, source := range sources {
		out[index] = recoveryCase{fmt.Sprintf("%s/%03d", prefix, index), source}
	}
	return out
}

// Partition the entire enumeration before applying ADAMIC_TEST_SHARD. Verify
// both cardinality and exact IDs, including duplicate IDs in the original list.
func partitionRecovery(test *testing.T, cases []recoveryCase, count int) [][]recoveryCase {
	test.Helper()
	if count < 1 || count > len(cases) {
		test.Fatalf("declared shard count %d for %d cases", count, len(cases))
	}
	shards := make([][]recoveryCase, count)
	want := map[string]bool{}
	for index, recoveryCase := range cases {
		if want[recoveryCase.id] {
			test.Fatalf("repeated case ID %s", recoveryCase.id)
		}
		want[recoveryCase.id] = true
		shards[index%count] = append(shards[index%count], recoveryCase)
	}
	seen := map[string]bool{}
	total := 0
	for _, shard := range shards {
		if len(shard) == 0 {
			test.Fatal("empty enumerated shard")
		}
		for _, recoveryCase := range shard {
			if !want[recoveryCase.id] || seen[recoveryCase.id] {
				test.Fatalf("missing or repeated case ID %s", recoveryCase.id)
			}
			seen[recoveryCase.id] = true
			total++
		}
	}
	if len(shards) != count || total != len(cases) || len(seen) != len(want) {
		test.Fatalf(
			"union mismatch: shards=%d/%d cases=%d/%d IDs=%d/%d",
			len(shards),
			count,
			total,
			len(cases),
			len(seen),
			len(want),
		)
	}
	for id := range want {
		if !seen[id] {
			test.Fatalf("missing case ID %s", id)
		}
	}
	test.Logf("union: %d cases, %d unique IDs, %d shards", total, len(seen), len(shards))
	return shards
}
func recoveryShardSelection(test *testing.T) (int, int) {
	test.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return 0, 1
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	index, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || n < 1 || index < 0 || index >= n {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index, n
}
func runRecoveryShards(test *testing.T, shards [][]recoveryCase, run func(*testing.T, int, []recoveryCase)) {
	test.Helper()
	selected, total := recoveryShardSelection(test)
	for index, cases := range shards {
		if index%total != selected {
			continue
		}
		// Not parallel within the parent: a parallel subtest queues behind every waiting test in the unit, and the
		// parent's own time counts that wait (TestRecoveryLoweredRecipe, one shard: 105.6 s alone on one core, 4 s of
		// work).
		test.Run(fmt.Sprintf("shard-%03d", index), func(test *testing.T) { run(test, index, cases) })
	}
}
func recoveryComparison(want, got []byte, mutant bool) string {
	diff := firstDifference(want, got)
	if mutant {
		if diff == "" {
			return "mutant survived"
		}
		return ""
	}
	return diff
}

func recoveryCount(got []byte, needle string, want int) string {
	count := strings.Count(string(got), needle)
	if count != want {
		return fmt.Sprintf("%q count=%d want=%d", needle, count, want)
	}
	return ""
}

type recoveryProof struct {
	Ids            []string `json:"IDs"`
	Count, Planted int
	Mode           string
}

// The subprocess deliberately fails. Its parent requires exactly the owning
// shard to fail, using the same dispatcher and verdict checks as the real tests.
func proveRecoveryShard(test *testing.T, cases []recoveryCase, count int, mutant bool) {
	mode := "agreement"
	if mutant {
		mode = "mutant"
	}
	proveRecoveryVerdict(test, cases, count, mode)
}
func proveRecoveryVerdict(test *testing.T, cases []recoveryCase, count int, mode string) {
	test.Helper()
	partitionRecovery(test, cases, count)
	planted := len(cases)/2 + 1
	if planted >= len(cases) {
		planted = len(cases) - 1
	}
	configuration := recoveryProof{Count: count, Planted: planted, Mode: mode}
	for _, recoveryCase := range cases {
		configuration.Ids = append(configuration.Ids, recoveryCase.id)
	}
	data, encodeError := json.Marshal(configuration)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command := exec.Command(executable, "-test.v", "-test.run=^TestRecoveryShardProofChild$")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_ESTREE_SHARD_PROOF=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_SHARD_PROOF="+string(data))
	output, combinedOutputError := command.CombinedOutput()
	owner := fmt.Sprintf("shard-%03d", planted%count)
	prefix := "--- FAIL: TestRecoveryShardProofChild/"
	if combinedOutputError == nil ||
		strings.Count(string(output), prefix) != 1 ||
		!strings.Contains(string(output), prefix+owner+" ") {
		test.Fatalf("planted %s must fail only %s: %v\n%s", cases[planted].id, owner, combinedOutputError, output)
	}
	test.Logf("planted %s %s failed only %s", mode, cases[planted].id, owner)
}
func TestRecoveryShardProofChild(test *testing.T) {
	test.Parallel()
	data := os.Getenv("ADAMIC_ESTREE_SHARD_PROOF")
	if data == "" {
		return
	}
	var configuration recoveryProof
	if decodeError := json.Unmarshal([]byte(data), &configuration); decodeError != nil {
		test.Fatal(decodeError)
	}
	var cases []recoveryCase
	for _, id := range configuration.Ids {
		cases = append(cases, recoveryCase{id: id})
	}
	shards := partitionRecovery(test, cases, configuration.Count)
	planted := cases[configuration.Planted].id
	runRecoveryShards(test, shards, func(test *testing.T, index int, cases []recoveryCase) {
		var want, got strings.Builder
		for _, recoveryCase := range cases {
			if configuration.Mode == "refusal" {
				script := "printf 'ESTree unattached decorator' >&2; exit 1"
				if recoveryCase.id == planted {
					script = "printf 'accepted'"
				}
				// A shell script, with no port to load, so no fixed cost to calibrate: the parser's own budget.
				refusedWithin(
					test,
					[]string{"/bin/sh", "-c", script},
					"ESTree unattached decorator",
					estreeParserBudget,
				)
				continue
			}
			want.WriteString("0 Program " + recoveryCase.id + "\n")
			bad := recoveryCase.id == planted
			if configuration.Mode == "mutant" {
				bad = index != configuration.Planted%configuration.Count
			}
			if bad {
				got.WriteString("0 WrongProgram " + recoveryCase.id + "\n")
			} else {
				got.WriteString("0 Program " + recoveryCase.id + "\n")
			}
		}
		var failure string
		if configuration.Mode == "control" {
			failure = recoveryCount([]byte(got.String()), "0 Program ", len(cases))
		} else if configuration.Mode != "refusal" {
			failure = recoveryComparison([]byte(want.String()), []byte(got.String()), configuration.Mode == "mutant")
		}
		if failure != "" {
			test.Fatal(failure)
		}
	})
}

func recoveryCpu(test *testing.T) time.Duration {
	test.Helper()
	var self, children syscall.Rusage
	if getrusageError := syscall.Getrusage(syscall.RUSAGE_SELF, &self); getrusageError != nil {
		test.Fatal(getrusageError)
	}
	if getrusageError := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children); getrusageError != nil {
		test.Fatal(getrusageError)
	}
	seconds := self.Utime.Sec + self.Stime.Sec + children.Utime.Sec + children.Stime.Sec
	micros := self.Utime.Usec + self.Stime.Usec + children.Utime.Usec + children.Stime.Usec
	return time.Duration(seconds)*time.Second + time.Duration(micros)*time.Microsecond
}

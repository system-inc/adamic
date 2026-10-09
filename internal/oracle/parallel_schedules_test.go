package oracle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Each explicit top-level unit is a gate shard. Observations always execute afresh.
type parallelScheduleUnit struct {
	name, path  string
	tsan        bool
	threads     []string
	first, last int
}

var parallelScheduleUnits = []parallelScheduleUnit{
	{"TestParallelSchedulesParallelFilesTSanLow01", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 1, 2},
	{"TestParallelSchedulesParallelFilesTSanLow02", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 3, 4},
	{"TestParallelSchedulesParallelFilesTSanLow03", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 5, 6},
	{"TestParallelSchedulesParallelFilesTSanLow04", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 7, 8},
	{"TestParallelSchedulesParallelFilesTSanLow05", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 9, 10},
	{"TestParallelSchedulesParallelFilesTSanLow06", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 11, 12},
	{"TestParallelSchedulesParallelFilesTSanLow07", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 13, 14},
	{"TestParallelSchedulesParallelFilesTSanLow08", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 15, 16},
	{"TestParallelSchedulesParallelFilesTSanLow09", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 17, 18},
	{"TestParallelSchedulesParallelFilesTSanLow10", "bench/parallel_files.a", true, []string{"1", "2", "3"}, 19, 20},
	{"TestParallelSchedulesParallelFilesTSanHigh01", "bench/parallel_files.a", true, []string{"7", "64", ""}, 1, 2},
	{"TestParallelSchedulesParallelFilesTSanHigh02", "bench/parallel_files.a", true, []string{"7", "64", ""}, 3, 4},
	{"TestParallelSchedulesParallelFilesTSanHigh03", "bench/parallel_files.a", true, []string{"7", "64", ""}, 5, 6},
	{"TestParallelSchedulesParallelFilesTSanHigh04", "bench/parallel_files.a", true, []string{"7", "64", ""}, 7, 8},
	{"TestParallelSchedulesParallelFilesTSanHigh05", "bench/parallel_files.a", true, []string{"7", "64", ""}, 9, 10},
	{"TestParallelSchedulesParallelFilesTSanHigh06", "bench/parallel_files.a", true, []string{"7", "64", ""}, 11, 12},
	{"TestParallelSchedulesParallelFilesTSanHigh07", "bench/parallel_files.a", true, []string{"7", "64", ""}, 13, 14},
	{"TestParallelSchedulesParallelFilesTSanHigh08", "bench/parallel_files.a", true, []string{"7", "64", ""}, 15, 16},
	{"TestParallelSchedulesParallelFilesTSanHigh09", "bench/parallel_files.a", true, []string{"7", "64", ""}, 17, 18},
	{"TestParallelSchedulesParallelFilesTSanHigh10", "bench/parallel_files.a", true, []string{"7", "64", ""}, 19, 20},

	{"TestParallelSchedulesFreshRelease", "internal/oracle/testdata/concurrency/accepted/fresh.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesFreshTSan", "internal/oracle/testdata/concurrency/accepted/fresh.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesGlobalCaptureRelease", "internal/oracle/testdata/concurrency/accepted/global_capture.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesGlobalCaptureTSan", "internal/oracle/testdata/concurrency/accepted/global_capture.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesIdentityRelease", "internal/oracle/testdata/concurrency/accepted/identity.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesIdentityTSan", "internal/oracle/testdata/concurrency/accepted/identity.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesLargeRelease", "internal/oracle/testdata/concurrency/accepted/large.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesLargeTSan", "internal/oracle/testdata/concurrency/accepted/large.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesMapCaptureRelease", "internal/oracle/testdata/concurrency/accepted/map_capture.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesMapCaptureTSan", "internal/oracle/testdata/concurrency/accepted/map_capture.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesNestedRelease", "internal/oracle/testdata/concurrency/accepted/nested.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesNestedTSan", "internal/oracle/testdata/concurrency/accepted/nested.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesNumbersRelease", "internal/oracle/testdata/concurrency/accepted/numbers.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesNumbersTSan", "internal/oracle/testdata/concurrency/accepted/numbers.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesReadonlyFormsRelease", "internal/oracle/testdata/concurrency/accepted/readonly_forms.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesReadonlyFormsTSan", "internal/oracle/testdata/concurrency/accepted/readonly_forms.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesReadonlyFunctionRelease", "internal/oracle/testdata/concurrency/accepted/readonly_function.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesReadonlyFunctionTSan", "internal/oracle/testdata/concurrency/accepted/readonly_function.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesRecordsRelease", "internal/oracle/testdata/concurrency/accepted/records.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesRecordsTSan", "internal/oracle/testdata/concurrency/accepted/records.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesRecursiveTreeRelease", "internal/oracle/testdata/concurrency/accepted/recursive_tree.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesRecursiveTreeTSan", "internal/oracle/testdata/concurrency/accepted/recursive_tree.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesStringsRelease", "internal/oracle/testdata/concurrency/accepted/strings.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesStringsTSan", "internal/oracle/testdata/concurrency/accepted/strings.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesThrowMiddleRelease", "internal/oracle/testdata/concurrency/accepted/throw_middle.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesThrowMiddleTSan", "internal/oracle/testdata/concurrency/accepted/throw_middle.a", true, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesParallelFilesRelease", "bench/parallel_files.a", false, []string{"1", "2", "3", "7", "64", ""}, 1, 20},
	{"TestParallelSchedulesMoveObjectsReleaseLow", "internal/oracle/testdata/moves/accepted/objects.a", false, []string{"1", "2", "3"}, 1, 20},
	{"TestParallelSchedulesMoveObjectsReleaseHigh", "internal/oracle/testdata/moves/accepted/objects.a", false, []string{"7", "64", ""}, 1, 20},
	{"TestParallelSchedulesMoveObjectsTSanLow01", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"1", "2", "3"}, 1, 5},
	{"TestParallelSchedulesMoveObjectsTSanLow02", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"1", "2", "3"}, 6, 10},
	{"TestParallelSchedulesMoveObjectsTSanLow03", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"1", "2", "3"}, 11, 15},
	{"TestParallelSchedulesMoveObjectsTSanLow04", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"1", "2", "3"}, 16, 20},
	{"TestParallelSchedulesMoveObjectsTSanHigh01", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"7", "64", ""}, 1, 5},
	{"TestParallelSchedulesMoveObjectsTSanHigh02", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"7", "64", ""}, 6, 10},
	{"TestParallelSchedulesMoveObjectsTSanHigh03", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"7", "64", ""}, 11, 15},
	{"TestParallelSchedulesMoveObjectsTSanHigh04", "internal/oracle/testdata/moves/accepted/objects.a", true, []string{"7", "64", ""}, 16, 20},
}

// Verify the explicit inventory against exactly the old sweep's selection, including
// both build variants and every thread count. A newly registered fixture must get units.
func TestParallelSchedulesUnion(t *testing.T) {
	t.Parallel()
	expected := map[string]bool{}
	for _, fixture := range fixtures {
		if strings.HasPrefix(fixture.path, "internal/oracle/testdata/concurrency/accepted/") || fixture.path == "bench/parallel_files.a" {
			expected[fixture.path] = true
		}
	}
	expected["internal/oracle/testdata/moves/accepted/objects.a"] = true
	seen := map[string]bool{}
	names := map[string]bool{}
	for _, unit := range parallelScheduleUnits {
		if names[unit.name] {
			t.Errorf("duplicate unit %s", unit.name)
		}
		names[unit.name] = true
		if !expected[unit.path] {
			t.Errorf("unexpected fixture %s", unit.path)
		}
		if unit.first < 1 || unit.last > 20 || unit.first > unit.last {
			t.Errorf("invalid repetition range for %s", unit.name)
		}
		for _, threads := range unit.threads {
			if !slices.Contains([]string{"1", "2", "3", "7", "64", ""}, threads) {
				t.Errorf("unexpected thread count %q", threads)
			}
			for repetition := unit.first; repetition <= unit.last; repetition++ {
				if threads == "1" && repetition > 1 {
					continue
				}
				key := fmt.Sprintf("%s/%t/%s/%d", unit.path, unit.tsan, threads, repetition)
				if seen[key] {
					t.Errorf("duplicate schedule %s", key)
				}
				seen[key] = true
			}
		}
	}
	for path := range expected {
		for _, tsan := range []bool{false, true} {
			for _, threads := range []string{"1", "2", "3", "7", "64", ""} {
				repetitions := 20
				if threads == "1" {
					repetitions = 1
				}
				for repetition := 1; repetition <= repetitions; repetition++ {
					key := fmt.Sprintf("%s/%t/%s/%d", path, tsan, threads, repetition)
					if !seen[key] {
						t.Errorf("missing schedule %s", key)
					}
				}
			}
		}
	}
}

// Preparations are shared by fixture, and builds by fixture/variant, even when a
// slow pair is split again. Keep their directory alive until TestMain finishes.
type parallelScheduleFixture struct {
	once   sync.Once
	source string
	oracle run
	ready  bool
	builds sync.Map
}
type parallelScheduleBuild struct {
	once   sync.Once
	binary string
	err    error
}

var parallelScheduleFixtures sync.Map
var parallelScheduleDirectory struct {
	once sync.Once
	path string
	err  error
}

func parallelSchedulePrepared(t *testing.T, unit parallelScheduleUnit) (run, string) {
	t.Helper()
	value, _ := parallelScheduleFixtures.LoadOrStore(unit.path, new(parallelScheduleFixture))
	fixture := value.(*parallelScheduleFixture)
	fixture.once.Do(func() {
		path := filepath.Join(repository, unit.path)
		started := time.Now()
		program, err := lowered(t, path)
		t.Logf("schedule lowering: %s", time.Since(started))
		if err != nil {
			t.Fatal(err)
		}
		fixture.source = native.C(program)
		if !strings.Contains(fixture.source, "adamic_parallel_map(") && !strings.Contains(fixture.source, "adamic_parallel_map_move(") {
			t.Fatal("schedule fixture emits no parallel map")
		}
		started = time.Now()
		fixture.oracle = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
		t.Logf("schedule Node: %s", time.Since(started))
		fixture.ready = true
	})
	if !fixture.ready {
		t.Fatal("fixture preparation failed in an earlier unit")
	}
	parallelScheduleDirectory.once.Do(func() {
		parallelScheduleDirectory.path, parallelScheduleDirectory.err = os.MkdirTemp("", "parallel-schedules-")
	})
	if parallelScheduleDirectory.err != nil {
		t.Fatal(parallelScheduleDirectory.err)
	}
	value, _ = fixture.builds.LoadOrStore(unit.tsan, new(parallelScheduleBuild))
	build := value.(*parallelScheduleBuild)
	build.once.Do(func() {
		directory, err := os.MkdirTemp(parallelScheduleDirectory.path, "build-")
		if err != nil {
			build.err = err
			return
		}
		build.binary = filepath.Join(directory, "program")
		started := time.Now()
		build.err = native.Build(fixture.source, build.binary, native.Options{ThreadSanitize: unit.tsan})
		t.Logf("schedule build: %s", time.Since(started))
	})
	if build.err != nil {
		t.Fatal(build.err)
	}
	return fixture.oracle, build.binary
}

func parallelSchedules(t *testing.T) {
	t.Helper()
	index := slices.IndexFunc(parallelScheduleUnits, func(unit parallelScheduleUnit) bool { return unit.name == t.Name() })
	if index < 0 {
		t.Fatal("top-level unit missing from schedule inventory")
	}
	unit := parallelScheduleUnits[index]
	if unit.tsan && runtime.GOOS != "linux" {
		t.Skip("TSan schedule proof requires Linux")
	}
	oracle, binary := parallelSchedulePrepared(t, unit)
	completed := 0
	defer func() { t.Logf("schedule unit: %d native executions", completed) }()
	for _, threads := range unit.threads {
		name, repetitions := threads, 20
		if threads == "1" {
			repetitions = 1
		}
		if unit.first > repetitions {
			continue
		}
		if threads == "" {
			name = "unset"
		}
		t.Run(name, func(t *testing.T) {
			for repetition := unit.first; repetition <= min(repetitions, unit.last); repetition++ {
				// Fresh processes on the shared binary; never cache a schedule observation.
				deadline := time.Minute
				if unit.tsan {
					deadline = 5 * time.Minute
				}
				observed := executeParallelDeadline(t, threads, false, deadline, binary)
				completed++
				t.Logf("execution=%d threads=%s repetition=%d/%d exit=%d", completed, name, repetition, repetitions, observed.exitCode)
				if bytes.Contains(observed.stderr, []byte("ThreadSanitizer")) {
					t.Fatalf("threads=%s repetition=%d: ThreadSanitizer report:\n%s", name, repetition, observed.stderr)
				}
				if difference := disagreement(oracle, observed); difference != "" {
					t.Fatalf("threads=%s repetition=%d: %s: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", name, repetition, difference, oracle.exitCode, oracle.stdout, oracle.stderr, observed.exitCode, observed.stdout, observed.stderr)
				}
			}
		})
	}
}

func TestParallelSchedulesFreshRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesFreshTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesGlobalCaptureRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesGlobalCaptureTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesIdentityRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesIdentityTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesLargeRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesLargeTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMapCaptureRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMapCaptureTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesNestedRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesNestedTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesNumbersRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesNumbersTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesReadonlyFormsRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesReadonlyFormsTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesReadonlyFunctionRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesReadonlyFunctionTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesRecordsRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesRecordsTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesRecursiveTreeRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesRecursiveTreeTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesStringsRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesStringsTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesThrowMiddleRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesThrowMiddleTSan(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesRelease(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsReleaseLow(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow01(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow02(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow03(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow04(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow05(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow06(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow07(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow08(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow09(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanLow10(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh01(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh02(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh03(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh04(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh05(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh06(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh07(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh08(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh09(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesParallelFilesTSanHigh10(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsReleaseHigh(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanLow01(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanLow02(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanLow03(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanLow04(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanHigh01(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanHigh02(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanHigh03(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

func TestParallelSchedulesMoveObjectsTSanHigh04(t *testing.T) {
	// Not parallel: each process can start 64 threads; keep fixture pools from competing.
	parallelSchedules(t)
}

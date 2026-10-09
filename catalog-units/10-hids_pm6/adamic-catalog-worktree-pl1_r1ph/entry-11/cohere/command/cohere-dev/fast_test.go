package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Every entry the fast tier leaves to the landing gate was recorded above fastBudgetCPUSeconds of CPU, and
// still costs more than fastEvictionCPUSeconds measured alone now. One that got cheap belongs back in the
// edit loop, and one that names nothing is refused. An entry recorded at or under the budget is refused
// too, so one cannot be added unmeasured or below the bar; that half judges a number in this file, which no
// load moves. The live half judges CPU, user plus system time, against a bar a third lower, so the
// scatter between runs cannot flip it; the wall is logged beside it and judges nothing. An entry whose cost
// is a corpus skips naming the corpus's variable when it is unset, since its tests then skip and it would
// read as free.
//
// It is itself the landing gate's: under the fast tier it skips, since measuring three entries costs
// about 30s of CPU in an edit loop meant to be instant.
//
// Not parallel: the package's other tests beside it would compete for the cores and stretch what it logs.
func TestEveryLandingGateOnlyEntryStillEarnsItsPlace(t *testing.T) {
	if os.Getenv(fastTierVariable) != "" {
		t.Skip("measuring the landing gate's entries is the landing gate's job, left to it by cohere-dev test --fast")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := modulePackagesIn(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range landingGateOnly {
		// Not parallel: each entry's measurement would compete with the others' for the cores.
		t.Run(entry.pattern, func(t *testing.T) {
			if !packages[entry.pattern] {
				t.Fatalf("%s is not a package of this module", entry.pattern)
			}
			if entry.wallSeconds <= 0 || entry.cpuSeconds <= 0 || entry.reason == "" {
				t.Fatalf("%s is listed without its measured wall, CPU or reason", entry.pattern)
			}
			if entry.cpuSeconds <= fastBudgetCPUSeconds {
				t.Fatalf("%s is recorded at %.1fs of CPU, at or under the fast budget of %.0fs, so it belongs in the fast tier",
					entry.pattern, entry.cpuSeconds, float64(fastBudgetCPUSeconds))
			}
			if entry.reads.Name != "" {
				entry.reads.Root(t)
			}
			wall, cpu := measureTests(t, root, entry.pattern, nil)
			if entry.skippedBy != "" {
				fastWall, fastCPU := measureTests(t, root, entry.pattern, []string{entry.skippedBy + "=1"})
				wall, cpu = wall-fastWall, cpu-fastCPU
			}
			t.Logf("%s: %.1fs wall, %.1fs CPU now (recorded %.1fs, %.1fs)", entry.pattern, wall, cpu, entry.wallSeconds, entry.cpuSeconds)
			if cpu <= fastEvictionCPUSeconds {
				t.Errorf("%s costs %.1fs of CPU alone (%.1fs of wall), at or under the eviction bar of %.0fs of CPU, so "+
					"it belongs back in the fast tier: remove it from landingGateOnly", entry.pattern, cpu, wall,
					float64(fastEvictionCPUSeconds))
			}
		})
	}
}

// measureTests runs one package's tests with -count=1 and returns the wall and the CPU, user plus system,
// of go test and everything it waited for. The fast tier's variable is never inherited, so a measurement
// in full is in full whatever ran this test; environment adds it where a caller wants it.
//
// The test binary is built first, untimed, so the CPU is the tests' and never the compiler's: otherwise
// whichever of an entry's two measurements came first would pay for the build, and the difference between
// them would read a compile as the entry's cost.
func measureTests(t *testing.T, root string, pattern string, environment []string) (wall float64, cpu float64) {
	t.Helper()
	build := exec.Command("go", "test", "-count=1", "-run", "^$", pattern)
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building %s's tests failed, so it measured nothing: %v\n%s", pattern, err, output)
	}
	command := exec.Command("go", "test", "-count=1", pattern)
	command.Dir = root
	for _, variable := range os.Environ() {
		if !strings.HasPrefix(variable, fastTierVariable+"=") {
			command.Env = append(command.Env, variable)
		}
	}
	command.Env = append(command.Env, environment...)
	start := time.Now()
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go test -count=1 %s failed, so it measured nothing: %v\n%s", pattern, err, output)
	}
	return time.Since(start).Seconds(), (command.ProcessState.UserTime() + command.ProcessState.SystemTime()).Seconds()
}

// The fast tier refuses -count=1 as a whole-module run does, before it runs anything: it would throw away
// the cache the fast tier rests on, across nearly every package, with no slot.
func TestTheFastTierRefusesCountOnce(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{{"--fast", "-count=1"}, {"--fast", "-count", "1"}, {"--fast", "-v", "--count=1"}} {
		// --full-priority, so the refusal is all this asks of run: nothing nices this test process.
		if code := run("test", append([]string{fullPriorityFlag}, arguments...)); code != 2 {
			t.Errorf("cohere-dev test %v exited %d, want 2, the refusal", arguments, code)
		}
	}
}

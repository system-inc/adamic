package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: changes runtime.GOMAXPROCS and saturates all CPUs in the saturated subtest.
func TestProgramCPUDeadline(test *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		test.Skip("CPU limits require Linux or macOS")
	}
	// clock() measures this child's CPU time, independent of scheduling delays.
	directory := test.TempDir()
	source := filepath.Join(directory, "cpu.c")
	binary := filepath.Join(directory, "cpu")
	if writeError := os.WriteFile(source, []byte(`#include <stdio.h>
#include <time.h>
int main(int argc, char **argv) {
    clock_t start = clock();
    for (;;) {
        if (argc > 1 && clock() - start >= CLOCKS_PER_SEC) {
            puts("completed");
            return 0;
        }
    }
}
`), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	if output, combinedOutputError := toolchain.Command(
		"clang",
		"-O2",
		"-o",
		binary,
		source,
	).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("build CPU helper: %v\n%s", combinedOutputError, output)
	}
	test.Run("spin", func(test *testing.T) {
		test.Parallel()
		start := time.Now()
		result := runProgram(time.Second, nil, binary)
		if !result.TimedOut || result.Exit != -1 || result.Signal != "" {
			test.Fatalf("spinning child: %+v, want TimedOut with Exit -1", result)
		}
		if time.Since(start) >= time.Minute {
			test.Fatal("spinning child reached the wall backstop instead of the CPU limit")
		}
	})
	// Not parallel: changes runtime.GOMAXPROCS and saturates all the unit's CPUs (GOMAXPROCS, never the host's cores, so a
	// runner's unit saturates its own share and not 128 threads' worth of a 64-core box).
	test.Run("saturated", func(test *testing.T) {
		processors := runtime.GOMAXPROCS(0)
		previous := runtime.GOMAXPROCS(2 * processors)
		defer runtime.GOMAXPROCS(previous)
		var stop atomic.Bool
		var ready, workers sync.WaitGroup
		for range 2 * processors {
			ready.Add(1)
			workers.Go(func() {
				ready.Done()
				for !stop.Load() {
				}
			})
		}
		defer func() { stop.Store(true); workers.Wait() }()
		ready.Wait()
		const budget = 2 * time.Second
		start := time.Now()
		result := runProgram(budget, nil, binary, "one-second")
		elapsed := time.Since(start)
		if result.TimedOut || result.Exit != 0 || result.Stdout != "completed\n" {
			test.Fatalf("1 CPU second under load: %+v after %s; want successful completion", result, elapsed)
		}
		// Whether the load stretched wall time past the budget depends on the box (about 1 run in 10
		// it doesn't), so it's reported, not asserted: the assertion is that CPU, not wall time,
		// decided the verdict, which holds either way.
		test.Logf(
			"1 CPU second completed in %s with a %s CPU budget (wall past the budget: %t)",
			elapsed,
			budget,
			elapsed > budget,
		)
	})
}

func TestRunFilterAttemptLimit(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	e, prepareModeError := prepareMode("../..", "testdata/mini", test.TempDir(), nil, nil, true)
	if prepareModeError != nil {
		test.Fatal(prepareModeError)
	}
	e.cache = &resultCache{directory: test.TempDir()}
	e.log = io.Discard
	for _, jobs := range []int{1, 4} {
		e.jobs = jobs
		report, runFilterError := e.runFilter("", 1, false)
		if runFilterError != nil {
			test.Fatal(runFilterError)
		}
		attempted := report.Pass + report.Fail + report.Refused + report.Crashed
		if report.Total != 5 || report.Skipped != 1 || report.Unrun != 3 || attempted != 1 {
			test.Fatalf("jobs=%d: total=%d skipped=%d unrun=%d attempted=%d; want 5, 1, 3, 1",
				jobs, report.Total, report.Skipped, report.Unrun, attempted)
		}
	}
}

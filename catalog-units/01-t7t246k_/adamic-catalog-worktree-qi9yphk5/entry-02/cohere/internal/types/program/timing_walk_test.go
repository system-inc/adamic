//go:build darwin || linux

package program_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// spinSink keeps the compiler from deleting the spin as dead code.
var spinSink atomic.Uint64

// spin burns a fixed amount of CPU. Fixed work rather than spinning on the clock, because a goroutine
// descheduled while it watched the clock would spend wall time without CPU, which is the very
// confusion under test.
func spin(iterations int) {
	value := uint64(1)
	for range iterations {
		value = value*6364136223846793005 + 1442695040888963407
	}
	spinSink.Add(value)
}

// threadCPU is the calling thread's CPU, by the kernel's thread clock: the same clock the meter reads, read
// here by the planted rule about itself, on its own thread at its own moment. It panics rather than taking
// a *testing.T, because the planted rules call it from the walk's workers, where t.Fatal may not be
// called; a rule's panic is contained and shows up below as a rule that ran on fewer files.
func threadCPU() time.Duration {
	var spec unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_THREAD_CPUTIME_ID, &spec); err != nil {
		panic(fmt.Sprintf("reading the thread CPU clock: %v", err))
	}
	return time.Duration(spec.Nano())
}

// TestATimedWalkBillsCPUNotWaiting is the control the instrument is trusted on (#8qyzmxw).
//
// The table it replaced charged each listener call its wall time, so a rule descheduled mid-call was
// billed for the time it spent off the CPU, and the better-tailwindcss family read about 1.2s when its
// real cost was about 0.05s. Two planted rules tell the two instruments apart. One sleeps in every file:
// it waits and spends almost no CPU, so it must be billed a small fraction of its sleep, where a wall
// clock bills all of it. The other spins a fixed amount of work in every file and measures its own CPU
// as it goes: it must be billed what it measured.
//
// Both bounds are relative, so they hold at any load (#8qyzmxw's flake: at load 111 the sleeper was
// billed 2.39ms after sleeping 7.69s, past an absolute 2ms bound, at 0.03% of its sleep). And neither
// reads process-wide CPU, which a parallel test beside this one would add to. Shown able to fail: with the
// meter reading the wall clock in place of the thread clock, the sleeper is billed nearly all of its sleep.
func TestATimedWalkBillsCPUNotWaiting(t *testing.T) {
	t.Parallel()
	const fileCount = 80
	const sleepPerFile = 5 * time.Millisecond

	// Size one spin call to about 5ms of CPU, on this thread's own clock.
	runtime.LockOSThread()
	iterations := 1_000_000
	for {
		before := threadCPU()
		spin(iterations)
		if threadCPU()-before >= 4*time.Millisecond {
			break
		}
		iterations *= 2
	}
	runtime.UnlockOSThread()

	files := map[string]string{"tsconfig.json": minimalConfig}
	for index := range fileCount {
		files[fmt.Sprintf("file%02d.ts", index)] = fmt.Sprintf("export const value%d = %d;\n", index, index)
	}
	root := writeProject(t, files)
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// The walk locks each worker to its thread for the walk, so a listener's own reads of the thread
	// clock are of the thread the meter reads.
	var sleptTotal, spunTotal atomic.Int64
	sleeper := rule.Rule{
		Name: "planted-sleeper",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
				start := time.Now()
				time.Sleep(sleepPerFile)
				sleptTotal.Add(int64(time.Since(start)))
			}}
		},
	}
	spinner := rule.Rule{
		Name: "planted-spinner",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
				start := threadCPU()
				spin(iterations)
				spunTotal.Add(int64(threadCPU() - start))
			}}
		},
	}

	graph.CollectTimings = true
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{sleeper, spinner})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if result.Timings.Account.Unavailable != "" {
		t.Fatalf("the walk was not measured, so this proves nothing: %s", result.Timings.Account.Unavailable)
	}

	byName := map[string]program.RuleTiming{}
	for _, timing := range result.Timings.Sorted() {
		byName[timing.Name] = timing
	}
	slept, spun := byName["planted-sleeper"], byName["planted-spinner"]

	// Both rules ran in every file, or a zero below would be a rule that never ran.
	if slept.NodesOffered != fileCount || spun.NodesOffered != fileCount {
		t.Fatalf("the planted rules ran on %d and %d files, want %d each", slept.NodesOffered, spun.NodesOffered, fileCount)
	}
	sleptWall := time.Duration(sleptTotal.Load())
	if sleptWall < fileCount*sleepPerFile {
		t.Fatalf("the sleeper slept %v, under the %v it was planted with, so it cannot show a wall clock's error",
			sleptWall, fileCount*sleepPerFile)
	}

	// The sleeper's CPU is the few instructions around each sleep: about 28µs per 50ms slept on
	// darwin/arm64, a twentieth of a percent. A twentieth of the sleep is a hundred times that, and a wall
	// clock bills twenty times more again.
	if slept.TotalCPU() > sleptWall/20 {
		t.Fatalf("a rule that only sleeps was billed %v of CPU, having slept %v", slept.TotalCPU(), sleptWall)
	}

	// The spinner's own reads bracket its work inside the meter's, so the meter bills it what it measured,
	// less the calibrated cost of measuring a call, plus the few instructions between the two pairs of reads.
	spunSelf := time.Duration(spunTotal.Load())
	if got := spun.TotalCPU(); got < spunSelf*9/10 || got > spunSelf*11/10+2*time.Millisecond {
		t.Fatalf("a rule that measured %v of its own CPU was billed %v", spunSelf, got)
	}
	t.Logf("spinner billed %v against %v it measured itself; sleeper billed %v after %v asleep",
		spun.TotalCPU(), spunSelf, slept.TotalCPU(), sleptWall)
}

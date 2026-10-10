package fuzz

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// The subprocess measures CPU rather than elapsed time, so contention cannot
// shorten the finite workload. Keep it single-threaded like a generated loop.
// Not parallel: changes process-wide runtime.GOMAXPROCS in the CPU helper subprocess.
func TestExecuteCPUHelper(t *testing.T) {
	mode := os.Getenv("ADAMIC_FUZZ_CPU_HELPER")
	if mode == "" {
		return
	}
	runtime.GOMAXPROCS(1)
	cpu := func() time.Duration {
		var usage syscall.Rusage
		if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
			panic(err)
		}
		return time.Duration(usage.Utime.Sec+usage.Stime.Sec)*time.Second + time.Duration(usage.Utime.Usec+usage.Stime.Usec)*time.Microsecond
	}
	start := cpu()
	for mode == "spin" || cpu()-start < time.Second {
		for i := 0; i < 100000; i++ {
		}
	}
	os.Exit(0)
}

func TestExecuteCPULimit(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("infinite", func(t *testing.T) {
		t.Parallel()
		start := time.Now()
		run := execute(t.TempDir(), []string{"ADAMIC_FUZZ_CPU_HELPER=spin"}, time.Second, executable, "-test.run=^TestExecuteCPUHelper$")
		if !run.TimedOut {
			t.Fatalf("infinite CPU loop was not timed out: %+v", run)
		}
		if elapsed := time.Since(start); elapsed >= 2*time.Minute {
			t.Fatalf("infinite CPU loop reached the wall backstop: %s", elapsed)
		}
	})
	t.Run("SIGXCPU", func(t *testing.T) {
		t.Parallel()
		run := execute(t.TempDir(), nil, time.Second, "/bin/sh", "-c", "kill -XCPU $$")
		if !run.TimedOut {
			t.Fatalf("SIGXCPU death was not timed out: %+v", run)
		}
	})
	t.Run("saturated", func(t *testing.T) {
		t.Parallel()
		// Separate single-threaded processes keep all workers runnable even when
		// the Go scheduler would preempt busy goroutines in this test process.
		for i := 0; i < 2*runtime.NumCPU(); i++ {
			worker := exec.Command("/bin/sh", "-c", "while :; do :; done")
			if err := worker.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = worker.Process.Kill()
				_ = worker.Wait()
			})
		}
		const limit = 2 * time.Second
		start := time.Now()
		run := execute(t.TempDir(), []string{"ADAMIC_FUZZ_CPU_HELPER=finite"}, limit, executable, "-test.run=^TestExecuteCPUHelper$")
		elapsed := time.Since(start)
		if run.TimedOut || run.ExitCode != 0 {
			t.Fatalf("1 s CPU workload under saturation: TimedOut=%v ExitCode=%d elapsed=%s stderr=%s", run.TimedOut, run.ExitCode, elapsed, run.Stderr)
		}
		// Whether the load stretched wall time past the limit depends on the box (about 1 run in 10
		// it doesn't), so it's reported, not asserted: the assertion is that CPU, not wall time,
		// decided the verdict, which holds either way.
		t.Logf("1 s CPU workload completed in %s under a %s CPU limit (wall past the limit: %t)", elapsed, limit, elapsed > limit)
	})
}

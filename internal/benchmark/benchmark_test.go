package benchmark

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// megabytes is a peak as the tests' messages print it.
func megabytes(bytes int64) string {
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
}

// On Linux, exec folds the starting process's high-water resident set into the program's peak, so a program started
// directly read at least the measuring process's own (J-wme5jp2). Through the reaper, a program that only prints reads
// its own few megabytes while this process holds 256 MB. Started directly, the same program reads the 256 MB, which
// pins the fold itself: if the kernel ever stops it, this says so, and the reaper can go.
// Not parallel: what it reads depends on what this process holds.
func TestPeakIsTheProgramsOwn(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the reaper is built and run on Linux only, where exec carries the starting process's peak into the program")
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("no clang on PATH to build internal/benchmark/reaper/reaper.c")
	}
	built, err := BuildReaper(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	held := make([]byte, 256<<20)
	for index := 0; index < len(held); index += 4096 {
		held[index] = 1
	}
	t.Run("ThroughTheReaper", func(t *testing.T) {
		result := Run(built, []string{"/bin/echo", "hi"}, "", 20*time.Second, Environment())
		if result.Failure != "" {
			t.Fatalf("/bin/echo through the reaper: %s", result.Failure)
		}
		if result.Peak >= 10<<20 {
			t.Fatalf("through the reaper, /bin/echo's peak read %s while this process held 256 MB; want under 10 MB, its own", megabytes(result.Peak))
		}
		t.Logf("through the reaper, /bin/echo's peak read %s while this process held 256 MB", megabytes(result.Peak))
	})
	t.Run("Directly", func(t *testing.T) {
		result := Run("", []string{"/bin/echo", "hi"}, "", 20*time.Second, Environment())
		if result.Failure != "" {
			t.Fatalf("/bin/echo started directly: %s", result.Failure)
		}
		if result.Peak <= 200<<20 {
			t.Fatalf("started directly, /bin/echo's peak read %s while this process held 256 MB; want over 200 MB. Linux's exec no longer carries the starting process's peak into the program, so internal/benchmark/reaper/reaper.c cures nothing and can go", megabytes(result.Peak))
		}
		t.Logf("started directly, /bin/echo's peak read %s while this process held 256 MB, the fold the reaper cures", megabytes(result.Peak))
	})
	runtime.KeepAlive(held)
}

// The reaper's line reads as the program's user and system time and its peak in bytes; a line missing, cut short, with
// a word or a negative in it, or without its newline is an error, never a zero. Mutants: the peak left in kilobytes;
// the system time dropped, or read from the user field; the newline not required, so a line the reaper never finished
// reads as numbers.
func TestReapedReadsTheReapersLine(t *testing.T) {
	t.Parallel()
	user, system, peak, err := reaped("1250000 30000 2048\n")
	if err != nil || user != 1250*time.Millisecond || system != 30*time.Millisecond || peak != 2048*1024 {
		t.Fatalf("read user %s, system %s, peak %d, %v; want 1.25s, 30ms and 2 MiB", user, system, peak, err)
	}
	for _, malformed := range []string{"", "1250000 30000\n", "1250000 30000 2048", "1250000 soon 2048\n", "1250000 30000 -1\n", "1 2 3 4\n"} {
		if _, _, _, err := reaped(malformed); err == nil {
			t.Errorf("%q read without an error", malformed)
		}
	}
}

// A large output is kept as its first line, its length and its sha256, so two runs that printed the same bytes still
// compare equal, one byte apart compares unequal, and nothing near the output's size is held per run.
func TestALargeOutputIsKeptAsItsDigest(t *testing.T) {
	t.Parallel()
	large := []byte("first line\n" + strings.Repeat("x", outputKeptWhole))
	other := slices.Clone(large)
	other[len(other)-1] = 'y'
	if keptOutput(large) != keptOutput(slices.Clone(large)) {
		t.Fatal("the same large output kept two different ways")
	}
	if keptOutput(large) == keptOutput(other) {
		t.Fatal("two large outputs one byte apart kept the same way")
	}
	if kept := keptOutput(large); len(kept) > 200 || firstLine(kept) != "first line" {
		t.Fatalf("a large output kept as %d bytes with first line %q; want a short digest under its own first line", len(kept), firstLine(kept))
	}
	if small := "42\n"; keptOutput([]byte(small)) != small {
		t.Fatalf("a small output kept as %q; want it whole", keptOutput([]byte(small)))
	}
}

// A run with an input reads that file on its standard input, as the Benchmarks Game hands a program its input, and one
// whose input isn't there is a failure naming it, never a run reading nothing. Mutants: the input left off standard
// input, and cat prints nothing; a missing input started anyway.
func TestARunReadsItsInputOnStandardInput(t *testing.T) {
	t.Parallel()
	input := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(input, []byte(">ONE Homo sapiens alu\nGGCCGGGCGCGG\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Run("", []string{"/bin/cat"}, input, 20*time.Second, Environment())
	if result.Failure != "" || result.Stdout != ">ONE Homo sapiens alu\nGGCCGGGCGCGG\n" {
		t.Fatalf("cat with an input printed %q, failure %q; want the input back", result.Stdout, result.Failure)
	}
	missing := Run("", []string{"/bin/cat"}, filepath.Join(t.TempDir(), "missing.txt"), 20*time.Second, Environment())
	if !strings.HasPrefix(missing.Failure, "its input can't be opened: ") {
		t.Fatalf("cat with a missing input read as %+v; want its input named as the failure", missing)
	}
}

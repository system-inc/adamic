package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// benchFixture writes a program to run with Node, which the harness times as a whole process.
func benchFixture(t *testing.T, name, source string) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return []string{"node", path}
}

// A benchmark whose work vanished, the way a loop the optimizer deleted does, runs no slower than
// its runtime starting with nothing to do. The harness must refuse its number, and still report one
// for a program that does real work. Each round times the floor, the empty loop and the busy loop in
// turn, so load on a shared machine falls on all three alike.
// Not parallel: it times whole processes against a floor, which a sibling test beside it would load.
func TestRefusesARunThatOnlyMeasuredItsRuntimeStarting(t *testing.T) {
	floorProgram := benchFixture(t, "floor.mjs", "console.log(0);\n")
	empty := benchFixture(t, "empty.mjs", "for (let index = 0; index < 1000; index++) {}\nconsole.log(0);\n")
	busy := benchFixture(t, "busy.mjs", "let sum = 0;\nfor (let index = 0; index < 300000000; index++) { sum = (sum + index) % 1000003; }\nconsole.log(sum);\n")
	var floors, empties, busies []measurement
	for range 5 {
		floors = append(floors, run(floorProgram, 20*time.Second))
		empties = append(empties, run(empty, 20*time.Second))
		busies = append(busies, run(busy, 20*time.Second))
	}
	floor := slowestOf(floors)
	if floor.failure != "" {
		t.Fatalf("the floor must run: %s", floor.failure)
	}
	resolution := clockResolution()
	if judged := judge(bestOf(empties), floor, resolution); !strings.HasPrefix(judged.failure, "measured nothing: ") {
		t.Fatalf("an empty loop was reported as %s against a floor of up to %s; want it refused", short(judged), seconds(floor.elapsed))
	}
	if judged := judge(bestOf(busies), floor, resolution); judged.failure != "" {
		t.Fatalf("real work was refused against a floor of up to %s: %s", seconds(floor.elapsed), judged.failure)
	}
}

// Perry's report gave 0.0 ms in every cell: a reading at or below the clock's resolution measured
// nothing, whatever the floor says.
func TestRefusesAReadingAtTheClocksResolution(t *testing.T) {
	t.Parallel()
	resolution := clockResolution()
	if resolution <= 0 {
		t.Fatalf("clock resolution %s", resolution)
	}
	for _, elapsed := range []time.Duration{0, resolution} {
		judged := judge(measurement{elapsed: elapsed, stdout: "0\n"}, measurement{}, resolution)
		if !strings.HasPrefix(judged.failure, "measured nothing: ") {
			t.Fatalf("a reading of %s at a resolution of %s was reported as %s; want it refused", elapsed, resolution, short(judged))
		}
	}
}

// Every program prints a checksum, so a run that printed nothing did no work anyone can check.
func TestRefusesARunThatPrintedNothing(t *testing.T) {
	t.Parallel()
	result := run(benchFixture(t, "silent.mjs", "let sum = 0;\nfor (let index = 0; index < 1000; index++) { sum += index; }\n"), 20*time.Second)
	if result.failure != "printed nothing" {
		t.Fatalf("a silent program was reported as %s; want it refused", short(result))
	}
}

// A run that exited nonzero failed, whatever it printed first.
func TestRefusesARunThatExitedNonzero(t *testing.T) {
	t.Parallel()
	result := run(benchFixture(t, "exits.mjs", "console.log(0);\nprocess.exit(3);\n"), 20*time.Second)
	if !strings.HasPrefix(result.failure, "failed: ") {
		t.Fatalf("a program exiting 3 was reported as %s; want it refused", short(result))
	}
}

// On Linux, exec folds the starting process's high-water resident set into the program's peak, so a
// program bench started itself read at least bench's own (J-wme5jp2). Through the reaper, a program
// that only prints reads its own few megabytes while this process holds 256 MB. Started directly, the
// same program reads the 256 MB, which pins the fold itself: if the kernel ever stops it, this says so,
// and the reaper can go.
// Not parallel: what it reads depends on what this process holds, and it sets the package's reaper.
func TestPeakIsTheProgramsOwn(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the reaper is built and run on Linux only, where exec carries the starting process's peak into the program")
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("no clang on PATH to build bench/reaper/reaper.c")
	}
	// buildReaper finds bench/ from the repository's root and moves there; this puts the directory back.
	t.Chdir(".")
	built, err := buildReaper(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	previous := reaper
	t.Cleanup(func() { reaper = previous })
	held := make([]byte, 256<<20)
	for index := 0; index < len(held); index += 4096 {
		held[index] = 1
	}
	t.Run("ThroughTheReaper", func(t *testing.T) {
		reaper = built
		result := run([]string{"/bin/echo", "hi"}, 20*time.Second)
		if result.failure != "" {
			t.Fatalf("/bin/echo through the reaper: %s", result.failure)
		}
		if result.peak >= 10<<20 {
			t.Fatalf("through the reaper, /bin/echo's peak read %s while this process held 256 MB; want under 10 MB, its own", megabytes(result.peak))
		}
		t.Logf("through the reaper, /bin/echo's peak read %s while this process held 256 MB", megabytes(result.peak))
	})
	t.Run("Directly", func(t *testing.T) {
		reaper = ""
		result := run([]string{"/bin/echo", "hi"}, 20*time.Second)
		if result.failure != "" {
			t.Fatalf("/bin/echo started directly: %s", result.failure)
		}
		if result.peak <= 200<<20 {
			t.Fatalf("started directly, /bin/echo's peak read %s while this process held 256 MB; want over 200 MB. Linux's exec no longer carries the starting process's peak into the program, so bench/reaper/reaper.c cures nothing and can go", megabytes(result.peak))
		}
		t.Logf("started directly, /bin/echo's peak read %s while this process held 256 MB, the fold the reaper cures", megabytes(result.peak))
	})
	runtime.KeepAlive(held)
}

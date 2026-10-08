//go:build darwin

// Command apple times what a call into AppKit costs from Adamic, beside the same loops in Swift and
// Objective-C, and what a frame of SwiftUI costs beside Swift: go run ./bench/apple [-rounds 5].
// Each loop is timed as the wall time of the process at N iterations less its time at 0, so starting
// up and making the window cancel out; the runs are interleaved, and each cell is the best of the
// rounds. A counted build gives Adamic's retains, releases and allocations per iteration. State the
// machine's load with the numbers.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func main() {
	rounds := flag.Int("rounds", 5, "how many times to run each loop, keeping the best")
	flag.Parse()
	directory, err := os.MkdirTemp("", "adamic-bench-apple-")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(directory)

	// Each program's binaries, by implementation; counted is Adamic's counted build.
	binaries := map[string]map[string]string{}
	for _, name := range []string{"crossing", "frame"} {
		binaries[name] = map[string]string{
			"adamic":  filepath.Join(directory, name+"-adamic"),
			"counted": filepath.Join(directory, name+"-counted"),
			"swift":   filepath.Join(directory, name+"-swift"),
		}
		program, err := load.Load([]string{"bench/apple/testdata/" + name + ".a"})
		if err != nil {
			fail(err)
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			fail(err)
		}
		if err := native.BuildApple(native.C(lowered), binaries[name]["adamic"], native.Options{}); err != nil {
			fail(err)
		}
		if err := native.BuildApple(native.C(lowered), binaries[name]["counted"], native.Options{Count: true}); err != nil {
			fail(err)
		}
		command("xcrun", "swiftc", "-O", "bench/apple/testdata/"+name+".swift", "-o", binaries[name]["swift"])
	}
	binaries["crossing"]["objc"] = filepath.Join(directory, "crossing-objc")
	command("clang", "-fobjc-arc", "-O2", "bench/apple/testdata/crossing.m", "-framework", "AppKit", "-o", binaries["crossing"]["objc"])

	// A frame's loop takes its count as its only argument; a crossing's names the loop first.
	type loop struct {
		program, name   string
		times           int
		implementations []string
	}
	loops := []loop{
		{"crossing", "scalar", 2_000_000, []string{"adamic", "swift", "objc"}},
		{"crossing", "object", 2_000_000, []string{"adamic", "swift", "objc"}},
		{"crossing", "string", 200_000, []string{"adamic", "swift", "objc"}},
		{"frame", "frame", 300, []string{"adamic", "swift"}},
	}
	arguments := func(l loop, times int) []string {
		if l.program == "frame" {
			return []string{fmt.Sprint(times)}
		}
		return []string{l.name, fmt.Sprint(times)}
	}
	best := map[string]time.Duration{}
	for round := 0; round < *rounds; round++ {
		for _, l := range loops {
			for _, implementation := range l.implementations {
				for _, times := range []int{0, l.times} {
					key := fmt.Sprintf("%s %s %d", implementation, l.name, times)
					start := time.Now()
					command(binaries[l.program][implementation], arguments(l, times)...)
					if elapsed := time.Since(start); best[key] == 0 || elapsed < best[key] {
						best[key] = elapsed
					}
				}
			}
		}
	}
	fmt.Printf("nanoseconds per iteration (a frame's in microseconds), best of %d, interleaved\n", *rounds)
	for _, l := range loops {
		fmt.Printf("%-7s", l.name)
		unit := 1.0
		if l.program == "frame" {
			unit = 1000
		}
		for _, implementation := range l.implementations {
			full := best[fmt.Sprintf("%s %s %d", implementation, l.name, l.times)]
			empty := best[fmt.Sprintf("%s %s 0", implementation, l.name)]
			fmt.Printf("  %s %9.1f", implementation, float64(full-empty)/float64(l.times)/unit)
		}
		fmt.Println()
	}
	fmt.Println("Adamic's counts per iteration (counted build, N less 0)")
	for _, l := range loops {
		full := counts(binaries[l.program]["counted"], arguments(l, l.times))
		empty := counts(binaries[l.program]["counted"], arguments(l, 0))
		per := func(index int) float64 { return float64(full[index]-empty[index]) / float64(l.times) }
		fmt.Printf("%-7s  allocations %6.2f  retains %6.2f  releases %6.2f\n", l.name, per(0), per(2), per(3))
	}
}

// counts runs a counted build and returns its allocations, frees, retains and releases.
func counts(binary string, arguments []string) [4]int {
	output, err := exec.Command(binary, arguments...).CombinedOutput()
	if err != nil {
		fail(fmt.Errorf("%s: %w\n%s", binary, err, output))
	}
	text := string(output)
	if at := strings.LastIndex(text, "adamic: counts:"); at >= 0 {
		text = text[at:]
	}
	var result [4]int
	var peak, regions int
	if _, err := fmt.Sscanf(text, "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &result[0], &result[1], &result[2], &result[3], &peak, &regions); err != nil {
		fail(fmt.Errorf("%s: no counts: %w\n%s", binary, err, output))
	}
	return result
}

func command(name string, arguments ...string) {
	if output, err := exec.Command(name, arguments...).CombinedOutput(); err != nil {
		fail(fmt.Errorf("%s: %w\n%s", name, err, output))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bench/apple:", err)
	os.Exit(1)
}

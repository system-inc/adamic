//go:build darwin

// Command apple times what a call into AppKit costs from Adamic, beside the same loops in Swift and
// Objective-C: go run ./bench/apple [-rounds 5]. Each loop is timed as the wall time of the process
// at N iterations less its time at 0, so starting up and making the window cancel out; the runs are
// interleaved, and each cell is the best of the rounds. State the machine's load with the numbers.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	binaries := map[string]string{}
	binaries["adamic"] = filepath.Join(directory, "adamic")
	program, err := load.Load([]string{"bench/apple/testdata/crossing.a"})
	if err != nil {
		fail(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fail(err)
	}
	if err := native.BuildApple(native.C(lowered), binaries["adamic"], native.Options{}); err != nil {
		fail(err)
	}
	binaries["swift"] = filepath.Join(directory, "swift")
	command("xcrun", "swiftc", "-O", "bench/apple/testdata/crossing.swift", "-o", binaries["swift"])
	binaries["objc"] = filepath.Join(directory, "objc")
	command("clang", "-fobjc-arc", "-O2", "bench/apple/testdata/crossing.m", "-framework", "AppKit", "-o", binaries["objc"])

	loops := []struct {
		name  string
		times int
	}{{"scalar", 2_000_000}, {"object", 2_000_000}, {"string", 200_000}}
	implementations := []string{"adamic", "swift", "objc"}
	best := map[string]time.Duration{}
	for round := 0; round < *rounds; round++ {
		for _, loop := range loops {
			for _, implementation := range implementations {
				for _, times := range []int{0, loop.times} {
					key := fmt.Sprintf("%s %s %d", implementation, loop.name, times)
					start := time.Now()
					command(binaries[implementation], loop.name, fmt.Sprint(times))
					if elapsed := time.Since(start); best[key] == 0 || elapsed < best[key] {
						best[key] = elapsed
					}
				}
			}
		}
	}
	fmt.Printf("nanoseconds per iteration, best of %d, interleaved\n", *rounds)
	for _, loop := range loops {
		fmt.Printf("%-7s", loop.name)
		for _, implementation := range implementations {
			full := best[fmt.Sprintf("%s %s %d", implementation, loop.name, loop.times)]
			empty := best[fmt.Sprintf("%s %s 0", implementation, loop.name)]
			fmt.Printf("  %s %7.1f", implementation, float64(full-empty)/float64(loop.times))
		}
		fmt.Println()
	}
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

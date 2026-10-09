//go:build ignore

// Run this file with go run inside the integrated compiler checkout.
// It measures the public native.Build API after emission and runtime preparation.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: build-metrics GENERATED_C NEW_OUTPUT_DIRECTORY JOBS")
		os.Exit(2)
	}
	var jobs int
	if _, err := fmt.Sscan(os.Args[3], &jobs); err != nil || jobs < 1 {
		fmt.Fprintln(os.Stderr, "jobs must be positive")
		os.Exit(2)
	}
	source, err := os.ReadFile(os.Args[1])
	if err != nil || len(source) == 0 {
		fmt.Fprintln(os.Stderr, "generated C must exist and be nonempty:", err)
		os.Exit(1)
	}
	out, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err := os.Mkdir(out, 0o755); err != nil {
		panic(err) // Never replace earlier measurements or their cache.
	}
	if err := os.Setenv("XDG_CACHE_HOME", filepath.Join(out, "cache")); err != nil {
		panic(err)
	}
	os.Unsetenv("ADAMIC_NATIVE_SPLIT")
	os.Unsetenv("ADAMIC_GATE_UNCACHED")
	report := map[string]any{
		"generated_c_bytes": len(source), "generated_c_lines": strings.Count(string(source), "\n"),
		"jobs": jobs, "flags": native.Flags(native.Options{}),
		"timing_scope": "native.Build wall time, runtime prewarmed; includes splitting, preprocessing, cache checks, compilation and linking; excludes C emission and Go startup",
	}
	start := time.Now()
	if _, err := native.RuntimeLibrary("", native.Options{}); err != nil {
		panic(err)
	}
	report["runtime_preparation_seconds"] = time.Since(start).Seconds()
	failed := false
	for _, mode := range []string{"unsplit", "split-cold", "split-warm"} {
		binary := filepath.Join(out, mode)
		options := native.Options{Split: mode != "unsplit", Jobs: jobs}
		start := time.Now()
		err := native.Build(string(source), binary, options)
		row := map[string]any{"build_wall_seconds": time.Since(start).Seconds(), "binary": binary}
		if err != nil {
			row["error"] = err.Error()
			failed = true
		} else if info, err := os.Stat(binary); err == nil {
			row["binary_bytes"] = info.Size()
		}
		report[mode] = row
		if failed {
			break // A failed cold compile has no warm-cache measurement.
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(out, "report.json"), data, 0o644); err != nil {
		panic(err)
	}
	fmt.Print(string(data))
	if failed {
		os.Exit(1)
	}
}

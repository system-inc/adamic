//go:build stage3_split_metrics

// Build already emitted C, timing native.Build separately from checking and lowering.
package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func main() {
	if len(os.Args) != 6 {
		panic("usage: build-metrics C OUTPUT MODE JOBS REPORT")
	}
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	jobs, err := strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}
	options := native.Options{Split: os.Args[3] != "unsplit", Jobs: jobs}
	started := time.Now()
	buildErr := native.Build(string(source), os.Args[2], options)
	result := map[string]any{"mode": os.Args[3], "jobs": jobs,
		"native_build_wall_seconds": time.Since(started).Seconds(),
		"timing_scope":              "native.Build only: runtime preparation, splitting, preprocessing, object cache, clang compilation and linking; excludes checking, lowering and C emission", "flags": native.Flags(options)}
	if buildErr != nil {
		result["error"] = buildErr.Error()
	} else {
		stat, err := os.Stat(os.Args[2])
		if err != nil {
			panic(err)
		}
		result["binary_bytes"] = stat.Size()
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[5], append(data, '\n'), 0644); err != nil {
		panic(err)
	}
	if buildErr != nil {
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Optional process-wide profiling starts after load and stops at the last release.
// It never changes fact bytes. CPU samples include native leaf PCs and Go GC.
var profile struct {
	file           *os.File
	before         runtime.MemStats
	inspect, parts time.Duration
}

func startProfile() {
	path := os.Getenv("ADAMIC_TSGO_PROFILE")
	if path == "" || profile.file != nil {
		return
	}
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	runtime.ReadMemStats(&profile.before)
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		panic(err)
	}
	profile.file = file
}
func queryStarted() time.Time {
	if profile.file == nil {
		return time.Time{}
	}
	return time.Now()
}
func queryFinished(start time.Time, parts bool) {
	if start.IsZero() {
		return
	}
	if parts {
		profile.parts += time.Since(start)
	} else {
		profile.inspect += time.Since(start)
	}
}
func stopProfile() {
	if profile.file == nil {
		return
	}
	pprof.StopCPUProfile()
	profile.file.Close()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	fmt.Fprintf(os.Stderr, "tsgo_go_profile: inspect_ns=%d parts_ns=%d allocated_bytes=%d allocated_objects=%d collections=%d\n", profile.inspect.Nanoseconds(), profile.parts.Nanoseconds(), after.TotalAlloc-profile.before.TotalAlloc, after.Mallocs-profile.before.Mallocs, after.NumGC-profile.before.NumGC)
	profile.file = nil
	profile.inspect = 0
	profile.parts = 0
}

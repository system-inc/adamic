package main

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
)

// cohere runs and dies, so the collector mostly buys nothing: memory it frees is handed back to an
// operating system that would reclaim all of it at exit a few seconds later. Measured on a cold ahra run
// (#nvv3bjy, 2026-10-02, interleaved at the same load): the default collector spends about 27% of the
// run's CPU, between the collections themselves (6% by gctrace) and the mark assists and write barriers
// it puts on the allocating goroutines. With collection off the run allocates about 11 GB and keeps it.
//
// So collection is off, under a ceiling at half the memory available when the run starts. Available
// rather than physical, because several cohere runs share one machine (one per worktree, one per agent),
// and ten cold ahra runs would hold 110 GB. Half rather than all, so a run leaves room for whatever starts
// beside it. On a quiet 128 GB machine the ceiling is never reached and nothing is collected; on a 16 GB
// runner it is about 7 GB and the collector wakes late; on a 7 GB runner it collects about as often as the
// default does, and the run never swaps.
//
// `GOGC` or `GOMEMLIMIT` in the environment wins: the runtime has already read it, and the policy steps
// aside. It is not a CohereSettings.json key, because it is a property of the machine, not the project.
//
// A test binary keeps the default. It is not a run that dies seconds later: it builds graph after graph for
// minutes, and with collection off it would hold every one of them up to the ceiling.
var activeMemoryPolicy = func() memoryPolicy {
	if testing.Testing() {
		return memoryPolicy{testBinary: true}
	}
	return decideMemoryPolicy(os.LookupEnv, availableMemory)
}()

func init() {
	activeMemoryPolicy.apply()
}

// memoryPolicy is what this process decided about its collector, kept so the summary can say so.
type memoryPolicy struct {
	// environment names the variables that set the collector, when any did, and then nothing else applies.
	environment string
	// available is the memory free for this run when it started, ceiling the limit set from it.
	available uint64
	ceiling   uint64
	// unreadable is why available memory could not be read, which leaves the runtime's defaults.
	unreadable error
	// testBinary is a `go test` process, which keeps the runtime's defaults.
	testBinary bool
}

// memoryCeilingFraction is the share of available memory one run may take before collecting.
const memoryCeilingFraction = 0.5

func decideMemoryPolicy(lookupEnvironment func(string) (string, bool), available func() (uint64, error)) memoryPolicy {
	var set []string
	for _, name := range []string{"GOGC", "GOMEMLIMIT"} {
		if value, ok := lookupEnvironment(name); ok {
			set = append(set, name+"="+value)
		}
	}
	if len(set) > 0 {
		return memoryPolicy{environment: strings.Join(set, " ")}
	}
	bytes, err := available()
	if err != nil {
		return memoryPolicy{unreadable: err}
	}
	if bytes == 0 {
		return memoryPolicy{unreadable: fmt.Errorf("it read as zero")}
	}
	return memoryPolicy{available: bytes, ceiling: uint64(float64(bytes) * memoryCeilingFraction)}
}

func (p memoryPolicy) apply() {
	if p.ceiling == 0 {
		return
	}
	// The limit first: with collection off and no limit, nothing would bound the heap for the instant
	// between the two calls.
	debug.SetMemoryLimit(int64(min(p.ceiling, math.MaxInt64)))
	debug.SetGCPercent(-1)
}

// line is the summary's account of the collector, so a run's memory use is never a surprise.
func (p memoryPolicy) line() string {
	switch {
	case p.testBinary:
		return "memory: the collector runs at Go's default, since this is a test binary"
	case p.environment != "":
		return "memory: the environment sets the collector (" + p.environment + ")"
	case p.unreadable != nil:
		return fmt.Sprintf("memory: the collector runs at Go's default, since available memory could not be read (%v)", p.unreadable)
	}
	return fmt.Sprintf("memory: collector off, ceiling %s (half of %s available)", gigabytes(p.ceiling), gigabytes(p.available))
}

// darwinAvailableMemory is free pages, plus file-backed pages (the cache the kernel drops before anything
// else), plus purgeable ones, which is what Activity Monitor counts as not in use. vm_stat's free plus
// inactive would need host_statistics64, out of reach of a binary built without cgo, and it would count
// inactive anonymous pages too, which come back only by compressing or swapping them.
func darwinAvailableMemory(pageSize, free, fileBacked, purgeable uint64) uint64 {
	return pageSize * (free + fileBacked + purgeable)
}

// linuxCgroupMemory is what a cgroup says about this process's memory, each file's raw bytes, any of
// them missing outside a container or under the other cgroup version.
type linuxCgroupMemory struct {
	v2Max, v2Current []byte // memory.max, memory.current
	v1Limit, v1Usage []byte // memory/memory.limit_in_bytes, memory/memory.usage_in_bytes
}

// linuxAvailableMemory is MemAvailable from /proc/meminfo, lowered to the room left under the cgroup's
// limit when one applies, since Go reads no cgroup limit for memory and a container is killed at its
// limit long before the host runs short.
func linuxAvailableMemory(meminfo []byte, cgroup linuxCgroupMemory) (uint64, error) {
	var available uint64
	found := false
	for line := range strings.SplitSeq(string(meminfo), "\n") {
		rest, ok := strings.CutPrefix(line, "MemAvailable:")
		if !ok {
			continue
		}
		kilobytes, err := strconv.ParseUint(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(rest), "kB")), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("reading MemAvailable from /proc/meminfo: %w", err)
		}
		available, found = kilobytes*1024, true
		break
	}
	if !found {
		return 0, fmt.Errorf("/proc/meminfo has no MemAvailable line")
	}
	for _, pair := range [][2][]byte{{cgroup.v2Max, cgroup.v2Current}, {cgroup.v1Limit, cgroup.v1Usage}} {
		// memory.max reads "max" when there is no limit, which fails the parse and so applies none.
		limit, err := strconv.ParseUint(strings.TrimSpace(string(pair[0])), 10, 64)
		if err != nil {
			continue
		}
		room := limit
		if used, err := strconv.ParseUint(strings.TrimSpace(string(pair[1])), 10, 64); err == nil {
			room = 0
			if used < limit {
				room = limit - used
			}
		}
		available = min(available, room)
	}
	return available, nil
}

func gigabytes(bytes uint64) string {
	value := float64(bytes) / 1e9
	if value < 10 {
		return fmt.Sprintf("%.1f GB", value)
	}
	return fmt.Sprintf("%.0f GB", value)
}

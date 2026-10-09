package typeaware

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const typeAwareUnitBudget = 60 * time.Second
const typeAwareUnitKill = 75 * time.Second

// Setup is split into individual immutable product builds. Every setup unit and
// every check shard has its own watchdog; one slow unit cannot hold the worker.
// Exit 124 identifies cooked work separately from an assertion failure.
func typeAwareDeadline(t *testing.T, name string) func() {
	t.Helper()
	started := time.Now()
	cooked := func() {
		fmt.Fprintf(os.Stdout, "cooked unit=%s elapsed_s=%.6f budget_s=60 kill_s=75\n", name, time.Since(started).Seconds())
		typeAwareKillChildren(os.Getpid())
		os.Exit(124)
	}
	timer := time.AfterFunc(typeAwareUnitKill, cooked)
	return func() {
		timer.Stop()
		elapsed := time.Since(started)
		if elapsed > typeAwareUnitBudget {
			cooked()
		}
		t.Logf("unit-budget name=%s elapsed_s=%.6f cooked=false", name, elapsed.Seconds())
	}
}

// The watchdog kills this test binary's subprocess tree before exiting, so a
// timed-out Go/clang build is not left running after its worker calls it cooked.
func typeAwareKillChildren(parent int) {
	entries, _ := filepath.Glob("/proc/[0-9]*/stat")
	for _, entry := range entries {
		data, err := os.ReadFile(entry)
		if err != nil {
			continue
		}
		_, tail, ok := strings.Cut(string(data), ") ")
		fields := strings.Fields(tail)
		if !ok || len(fields) < 2 {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil || ppid != parent {
			continue
		}
		pid, err := strconv.Atoi(filepath.Base(filepath.Dir(entry)))
		if err != nil {
			continue
		}
		typeAwareKillChildren(pid)
		if child, err := os.FindProcess(pid); err == nil {
			_ = child.Kill()
		}
	}
}

func typeAwareSetup(h *harness, name string, build func() string) string {
	h.t.Helper()
	defer typeAwareDeadline(h.t, "setup/"+name)()
	return build()
}
func typeAwareStage0(h *harness) string {
	return typeAwareSetup(h, "stage0", h.stage0)
}
func typeAwareArchive(h *harness, name, overlay string, sanitize bool) string {
	return typeAwareSetup(h, name, func() string { return h.archive(name, overlay, sanitize) })
}
func typeAwareBuild(h *harness, stage0, name, entry, archive string, sanitize bool) string {
	return typeAwareSetup(h, name, func() string { return h.build(stage0, name, entry, archive, sanitize) })
}
func typeAwareProduct(h *harness, name string, command *exec.Cmd) string {
	return typeAwareSetup(h, name, func() string { return h.sixBuildProduct(name, command) })
}
func typeAwareMutant(h *harness, stage0, archive, entry, name, relative, from, to string, imports ...string) string {
	return typeAwareSetup(h, name, func() string { return suiteTSMutant(h, stage0, archive, entry, name, relative, from, to, imports...) })
}
func typeAwareRunShards(t *testing.T, h *harness, expected []string, shards []sixShard, value string, required int) {
	t.Helper()
	// Keep the original scheduler, static declarations and both live union checks.
	bounded := append([]sixShard(nil), shards...)
	for i := range bounded {
		run := bounded[i].run
		bounded[i].run = func(h *harness) {
			defer typeAwareDeadline(h.t, h.t.Name())()
			run(h)
		}
	}
	sixRunShards(t, h, expected, bounded, value, required)
}

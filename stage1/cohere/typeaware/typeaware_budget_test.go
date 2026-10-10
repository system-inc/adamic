package typeaware

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const typeAwareUnitBudget = 90 * time.Second
const typeAwareUnitKill = 90 * time.Second

// Shared Six setup is observed separately; only case work has a watchdog.
// The outer gate still bounds the whole process, including cold builds.
// Crossing the deadline kills active command groups and fails the worker.
func typeAwareDeadline(t *testing.T, name string) func() {
	return typeAwareDeadlineWithin(t, name, typeAwareUnitBudget, typeAwareUnitKill)
}

func typeAwareDeadlineWithin(t *testing.T, name string, budget, kill time.Duration) func() {
	t.Helper()
	started := time.Now()
	cooked := func() {
		fmt.Fprintf(os.Stdout, "unit deadline exceeded name=%s elapsed_s=%.6f limit_s=%.6f\n", name, time.Since(started).Seconds(), kill.Seconds())
		typeAwareKillCommandGroups()
		os.Exit(1)
	}
	timer := time.AfterFunc(kill, cooked)
	return func() {
		timer.Stop()
		elapsed := time.Since(started)
		if elapsed > budget {
			cooked()
		}
		t.Logf("unit-budget name=%s elapsed_s=%.6f cooked=false", name, elapsed.Seconds())
	}
}

func typeAwareSetup(h *harness, name string, build func() string) string {
	h.t.Helper()
	// Shared setup has no deadline of its own; Loom bounds the whole unit.
	// A shard starts its unchanged watchdog only after its products are ready.
	started := time.Now()
	defer func() { h.t.Logf("shared-setup product=%s elapsed_s=%.6f", name, time.Since(started).Seconds()) }()
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
	return typeAwareSetup(h, name, func() string {
		if filepath.Base(entry) == "sharded_suite.ts" {
			data, err := os.ReadFile(entry)
			if err != nil {
				h.t.Fatal(err)
			}
			driver := strings.ReplaceAll(string(data), "../../../typescript", filepath.Join(h.repository, "stage1/typescript"))
			for _, file := range []string{"facts.ts", "diagnostic.ts", "unary_minus.ts", "rules.ts", "flags.ts", "frames.ts", "types.ts", "type_fact.ts", "parameters.ts"} {
				driver = strings.ReplaceAll(driver, "../"+file, "./"+file)
			}
			entry = h.write(name+"-driver.ts", driver)
		}
		return suiteTSMutant(h, stage0, archive, entry, name, relative, from, to, imports...)
	})
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

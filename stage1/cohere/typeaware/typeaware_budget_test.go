package typeaware

import (
	"fmt"
	"os"
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
func typeAwareDeadline(test *testing.T, name string) func() {
	return typeAwareDeadlineWithin(test, name, typeAwareUnitBudget, typeAwareUnitKill)
}

func typeAwareDeadlineWithin(test *testing.T, name string, budget, kill time.Duration) func() {
	test.Helper()
	started := time.Now()
	cooked := func() {
		fmt.Fprintf(
			os.Stdout,
			"unit deadline exceeded name=%s elapsed_s=%.6f limit_s=%.6f\n",
			name,
			time.Since(started).Seconds(),
			kill.Seconds(),
		)
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
		test.Logf("unit-budget name=%s elapsed_s=%.6f cooked=false", name, elapsed.Seconds())
	}
}

func typeAwareSetup(harness *harness, name string, build func() string) string {
	harness.t.Helper()
	// Shared setup has no deadline of its own; Loom bounds the whole unit.
	// A shard starts its unchanged watchdog only after its products are ready.
	started := time.Now()
	defer func() { harness.t.Logf("shared-setup product=%s elapsed_s=%.6f", name, time.Since(started).Seconds()) }()
	return build()
}
func typeAwareStage0(harness *harness) string {
	return typeAwareSetup(harness, "stage0", harness.stage0)
}
func typeAwareArchive(harness *harness, name string, sanitize bool) string {
	return typeAwareSetup(harness, name, func() string { return harness.archive(sanitize) })
}
func typeAwareMutantArchive(harness *harness, name string) string {
	return typeAwareSetup(harness, name+"-checker", func() string { return checkerMutantArchive(harness.t, name) })
}
func typeAwareBuild(harness *harness, stage0, name, entry, archive string, sanitize bool) string {
	return typeAwareSetup(harness, name, func() string { return harness.build(stage0, name, entry, archive, sanitize) })
}

// typeAwareMutantNative is typeAwareBuild for a mutant's native, made as a product (mutantNative) so its TestProduct_
// builds it ahead and its check only fetches it.
func typeAwareMutantNative(harness *harness, name, entry, archive string, sanitize bool) string {
	return typeAwareSetup(harness, name, func() string { return harness.mutantNative(name, entry, archive, sanitize) })
}
func typeAwareMutant(harness *harness, archive, entry, name, relative, from, to string, imports ...string) string {
	return typeAwareSetup(harness, name, func() string {
		if filepath.Base(entry) == "sharded_suite.ts" {
			data, readError := os.ReadFile(entry)
			if readError != nil {
				harness.t.Fatal(readError)
			}
			driver := strings.ReplaceAll(
				string(data),
				"../../../typescript",
				filepath.Join(harness.repository, "stage1/typescript"),
			)
			for _, file := range []string{
				"facts.ts",
				"diagnostic.ts",
				"unary_minus.ts",
				"rules.ts",
				"flags.ts",
				"frames.ts",
				"types.ts",
				"type_fact.ts",
				"parameters.ts",
			} {
				driver = strings.ReplaceAll(driver, "../"+file, "./"+file)
			}
			entry = harness.write(name+"-driver.ts", driver)
		}
		return suiteTSMutant(harness, archive, entry, name, relative, from, to, imports...)
	})
}

func typeAwareRunShards(test *testing.T, h *harness, expected []string, shards []sixShard, value string, required int) {
	test.Helper()
	// Keep the original scheduler, static declarations and both live union checks.
	bounded := append([]sixShard(nil), shards...)
	for index := range bounded {
		run := bounded[index].run
		bounded[index].run = func(harness *harness) {
			defer typeAwareDeadline(harness.t, harness.t.Name())()
			run(harness)
		}
	}
	sixRunShards(test, h, expected, bounded, value, required)
}

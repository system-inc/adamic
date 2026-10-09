package typeaware

import (
	"bytes"
	"context"
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

// Setup is split into individual immutable product builds. Every setup unit and
// every check shard has its own watchdog; one slow unit cannot hold the worker.
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

// A real nested subprocess must be killed before it can publish its marker.
// This exercises the same watchdog as the setup products and check shards.
// Not parallel: deadline probe shares the package command-group registry.
func TestTypeAwareUnitDeadline(t *testing.T) {
	mode := os.Getenv("ADAMIC_TYPEAWARE_DEADLINE_CHILD")
	marker := os.Getenv("ADAMIC_TYPEAWARE_DEADLINE_MARKER")
	if mode == "marker" {
		time.Sleep(time.Second)
		if err := os.WriteFile(marker, []byte("survived"), 0600); err != nil {
			t.Fatal(err)
		}
		return
	}
	if mode != "" {
		defer typeAwareDeadlineWithin(t, "watchdog-probe", 100*time.Millisecond, 200*time.Millisecond)()
		if mode == "quick" {
			return
		}
		command := exec.Command(os.Args[0], "-test.run=^TestTypeAwareUnitDeadline$")
		command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_DEADLINE_CHILD=marker")
		if err := typeAwareRunCommand(command); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, mode := range []string{"quick", "slow"} {
		t.Run(mode, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), "marker")
			ctx, cancel := context.WithTimeout(context.Background(), typeAwareChildLimit)
			defer cancel()
			command := typeAwareContextCommand(ctx, os.Args[0], "-test.run=^TestTypeAwareUnitDeadline$", "-test.v")
			command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_DEADLINE_CHILD="+mode, "ADAMIC_TYPEAWARE_DEADLINE_MARKER="+marker)
			output, err := command.CombinedOutput()
			if mode == "quick" {
				if err != nil || !bytes.Contains(output, []byte("cooked=false")) {
					t.Fatalf("quick unit: %v %s", err, output)
				}
				return
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !bytes.Contains(output, []byte("unit deadline exceeded name=watchdog-probe")) {
				t.Fatalf("slow unit was not cooked: %v %s", err, output)
			}
			time.Sleep(time.Second)
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("cooked unit left its subprocess alive: %v", err)
			}
		})
	}
}

package typeaware

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

type typeAwarePlan struct {
	h        *harness
	expected []string
	shards   []sixShard
}

type typeAwarePlanEntry struct {
	once sync.Once
	plan *typeAwarePlan
}

var typeAwarePlans sync.Map

func newTypeAwarePlan(test *testing.T, h *harness, expected []string, shards []sixShard, required int) *typeAwarePlan {
	test.Helper()
	if len(shards) != required {
		test.Fatalf("enumerated shard count %d differs from declared %d", len(shards), required)
	}
	if sixUnionError := sixUnion(expected, shards); sixUnionError != nil {
		test.Fatal(sixUnionError)
	}
	test.Logf("six-union count=%d unique=%d shards=%d", len(expected), len(expected), len(shards))
	test.Logf("split-setup elapsed_s=%.6f builds=shared-once", time.Since(h.setupStarted).Seconds())
	return &typeAwarePlan{h: h, expected: expected, shards: shards}
}
func typeAwareTopPlan(test *testing.T, key string) *typeAwarePlan {
	test.Helper()
	stored, _ := typeAwarePlans.LoadOrStore(key, &typeAwarePlanEntry{})
	entry := stored.(*typeAwarePlanEntry)
	entry.once.Do(func() {
		// Any selected shard prepares this plan once per process, without a
		// setup watchdog. Its own-work deadline starts after preparation.
		if key == "planted" {
			entry.plan = typeAwarePlantedPlan(test)
		} else if key == "six" {
			entry.plan = prepareSixRuleAgreementAndMutants(test)
		} else {
			entry.plan = prepareTypeAwareAgreementAndMutants(test)
		}
	})
	if entry.plan == nil {
		test.Fatal("shared shard setup did not complete")
	}
	return entry.plan
}
func typeAwareTopShard(test *testing.T, key string, index int) {
	test.Helper()
	plan := typeAwareTopPlan(test, key)
	defer typeAwareDeadline(test, test.Name())()
	shard := plan.shards[index]
	directory := filepath.Join(productDirectory, "top-shards", test.Name())
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	h := &harness{t: test, repository: plan.h.repository, directory: directory, parallel: true}
	started := time.Now()
	defer func() {
		test.Logf(
			"six-shard name=shard-%03d content=%s cases=%d elapsed_s=%.6f",
			index,
			shard.name,
			len(shard.ids),
			time.Since(started).Seconds(),
		)
	}()
	shard.run(h)
}
func typeAwareTopUnion(test *testing.T, key string) {
	test.Helper()
	plan := typeAwareTopPlan(test, key)
	defer typeAwareDeadline(test, test.Name())()
	if len(plan.expected) == 0 {
		test.Fatal("empty shard corpus")
	}
	if sixUnionError := sixUnion(plan.expected, plan.shards); sixUnionError != nil {
		test.Fatal(sixUnionError)
	}
	test.Logf("six-union count=%d unique=%d shards=%d", len(plan.expected), len(plan.expected), len(plan.shards))
}

// Corrupt one case and run the actual selected top-level leaf in a child.
// Precisely the owning leaf must fail; another slice must remain green.
func typeAwarePlantedPlan(test *testing.T) *typeAwarePlan {
	var shards []sixShard
	for i := range 2 {
		id := fmt.Sprintf("file.ts/%d", i)
		shards = append(shards, sixShard{name: fmt.Sprintf("planted-%d", i), ids: []string{id}, run: func(h *harness) {
			want := []byte("file\t" + id + "\n")
			got := append([]byte(nil), want...)
			if i == 1 && os.Getenv("ADAMIC_TYPEAWARE_TOP_PLANTED") == "1" {
				got[0] = 'X'
			}
			if err := sixOutputError(h.t.Name(), want, got, nil); err != nil {
				h.t.Fatal(err)
			}
		}})
	}
	return newTypeAwarePlan(test, &harness{setupStarted: time.Now()}, []string{"file.ts/0", "file.ts/1"}, shards, 2)
}
func TestTypeAwareTopPlanted_000(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "planted", 0)
}
func TestTypeAwareTopPlanted_001(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "planted", 1)
}
func TestTypeAwareTopShardPlantedFailure(test *testing.T) {
	test.Parallel()
	for i := range 2 {
		ctx, cancel := context.WithTimeout(context.Background(), typeAwareChildLimit)
		name := fmt.Sprintf("TestTypeAwareTopPlanted_%03d", i)
		command := typeAwareContextCommand(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_TOP_PLANTED=1")
		// Read through childguard.Capture, to its end with no WaitDelay deciding when, so a nonholding leaf that passed
		// but was slow to drain under load never reads as failed, and the holding leaf's message is never cut short.
		stdout, stderr, captureError := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		cancel()
		if i == 0 && captureError != nil {
			test.Fatalf("nonholding leaf failed: %v %s", captureError, output)
		}
		if i == 1 && (captureError == nil || !bytes.Contains(output, []byte("shard "+name+" mismatch"))) {
			test.Fatalf("holding leaf did not catch planted failure: %v %s", captureError, output)
		}
	}
}

func TestTypeAwareAgreementAndMutantsUnion(test *testing.T) {
	test.Parallel()
	typeAwareTopUnion(test, "typeaware")
}
func TestTypeAwareAgreementAndMutants_000(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 0)
}
func TestTypeAwareAgreementAndMutants_001(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 1)
}
func TestTypeAwareAgreementAndMutants_002(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 2)
}
func TestTypeAwareAgreementAndMutants_003(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 3)
}
func TestTypeAwareAgreementAndMutants_004(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 4)
}
func TestTypeAwareAgreementAndMutants_005(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 5)
}
func TestTypeAwareAgreementAndMutants_006(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 6)
}
func TestTypeAwareAgreementAndMutants_007(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 7)
}
func TestTypeAwareAgreementAndMutants_008(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 8)
}
func TestTypeAwareAgreementAndMutants_009(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 9)
}
func TestTypeAwareAgreementAndMutants_010(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 10)
}
func TestTypeAwareAgreementAndMutants_011(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 11)
}
func TestTypeAwareAgreementAndMutants_012(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 12)
}
func TestTypeAwareAgreementAndMutants_013(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 13)
}
func TestTypeAwareAgreementAndMutants_014(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 14)
}
func TestTypeAwareAgreementAndMutants_015(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 15)
}
func TestTypeAwareAgreementAndMutants_016(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 16)
}
func TestTypeAwareAgreementAndMutants_017(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 17)
}
func TestTypeAwareAgreementAndMutants_018(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 18)
}
func TestTypeAwareAgreementAndMutants_019(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 19)
}
func TestTypeAwareAgreementAndMutants_020(test *testing.T) {
	test.Parallel()
	typeAwareTopShard(test, "typeaware", 20)
}

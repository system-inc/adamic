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

func newTypeAwarePlan(t *testing.T, h *harness, expected []string, shards []sixShard, required int) *typeAwarePlan {
	t.Helper()
	if len(shards) != required {
		t.Fatalf("enumerated shard count %d differs from declared %d", len(shards), required)
	}
	if err := sixUnion(expected, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("six-union count=%d unique=%d shards=%d", len(expected), len(expected), len(shards))
	t.Logf("split-setup elapsed_s=%.6f builds=shared-once", time.Since(h.setupStarted).Seconds())
	return &typeAwarePlan{h: h, expected: expected, shards: shards}
}
func typeAwareTopPlan(t *testing.T, key string) *typeAwarePlan {
	t.Helper()
	stored, _ := typeAwarePlans.LoadOrStore(key, &typeAwarePlanEntry{})
	entry := stored.(*typeAwarePlanEntry)
	entry.once.Do(func() {
		// Shared setup has no deadline of its own. Loom bounds the whole unit;
		// the shard's clock starts after this once-per-process preparation.
		if key == "planted" {
			entry.plan = typeAwarePlantedPlan(t)
		} else if key == "six" {
			entry.plan = prepareSixRuleAgreementAndMutants(t)
		} else {
			entry.plan = prepareTypeAwareAgreementAndMutants(t)
		}
	})
	if entry.plan == nil {
		t.Fatal("shared shard setup did not complete")
	}
	return entry.plan
}
func typeAwareTopShard(t *testing.T, key string, index int) {
	t.Helper()
	plan := typeAwareTopPlan(t, key)
	defer typeAwareDeadline(t, t.Name())()
	shard := plan.shards[index]
	directory := filepath.Join(productDirectory, "top-shards", t.Name())
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: plan.h.repository, directory: directory, parallel: true}
	started := time.Now()
	defer func() {
		t.Logf("six-shard name=shard-%03d content=%s cases=%d elapsed_s=%.6f", index, shard.name, len(shard.ids), time.Since(started).Seconds())
	}()
	shard.run(h)
}
func typeAwareTopUnion(t *testing.T, key string) {
	t.Helper()
	plan := typeAwareTopPlan(t, key)
	defer typeAwareDeadline(t, t.Name())()
	if len(plan.expected) == 0 {
		t.Fatal("empty shard corpus")
	}
	if err := sixUnion(plan.expected, plan.shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("six-union count=%d unique=%d shards=%d", len(plan.expected), len(plan.expected), len(plan.shards))
}

// Corrupt one case and run the actual selected top-level leaf in a child.
// Precisely the owning leaf must fail; another slice must remain green.
func typeAwarePlantedPlan(t *testing.T) *typeAwarePlan {
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
	return newTypeAwarePlan(t, &harness{setupStarted: time.Now()}, []string{"file.ts/0", "file.ts/1"}, shards, 2)
}
func TestTypeAwareTopPlanted_000(t *testing.T) { t.Parallel(); typeAwareTopShard(t, "planted", 0) }
func TestTypeAwareTopPlanted_001(t *testing.T) { t.Parallel(); typeAwareTopShard(t, "planted", 1) }
func TestTypeAwareTopShardPlantedFailure(t *testing.T) {
	t.Parallel()
	for i := range 2 {
		ctx, cancel := context.WithTimeout(context.Background(), typeAwareChildLimit)
		name := fmt.Sprintf("TestTypeAwareTopPlanted_%03d", i)
		command := typeAwareContextCommand(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_TOP_PLANTED=1")
		output, err := command.CombinedOutput()
		cancel()
		if i == 0 && err != nil {
			t.Fatalf("nonholding leaf failed: %v %s", err, output)
		}
		if i == 1 && (err == nil || !bytes.Contains(output, []byte("shard "+name+" mismatch"))) {
			t.Fatalf("holding leaf did not catch planted failure: %v %s", err, output)
		}
	}
}

func TestTypeAwareAgreementAndMutantsUnion(t *testing.T) {
	t.Parallel()
	typeAwareTopUnion(t, "typeaware")
}
func TestTypeAwareAgreementAndMutants_000(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 0)
}
func TestTypeAwareAgreementAndMutants_001(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 1)
}
func TestTypeAwareAgreementAndMutants_002(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 2)
}
func TestTypeAwareAgreementAndMutants_003(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 3)
}
func TestTypeAwareAgreementAndMutants_004(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 4)
}
func TestTypeAwareAgreementAndMutants_005(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 5)
}
func TestTypeAwareAgreementAndMutants_006(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 6)
}
func TestTypeAwareAgreementAndMutants_007(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 7)
}
func TestTypeAwareAgreementAndMutants_008(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 8)
}
func TestTypeAwareAgreementAndMutants_009(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 9)
}
func TestTypeAwareAgreementAndMutants_010(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 10)
}
func TestTypeAwareAgreementAndMutants_011(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 11)
}
func TestTypeAwareAgreementAndMutants_012(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 12)
}
func TestTypeAwareAgreementAndMutants_013(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 13)
}
func TestTypeAwareAgreementAndMutants_014(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 14)
}
func TestTypeAwareAgreementAndMutants_015(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 15)
}
func TestTypeAwareAgreementAndMutants_016(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 16)
}
func TestTypeAwareAgreementAndMutants_017(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 17)
}
func TestTypeAwareAgreementAndMutants_018(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 18)
}
func TestTypeAwareAgreementAndMutants_019(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 19)
}
func TestTypeAwareAgreementAndMutants_020(t *testing.T) {
	t.Parallel()
	typeAwareTopShard(t, "typeaware", 20)
}

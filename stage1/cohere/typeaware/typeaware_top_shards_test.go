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
		defer typeAwareDeadline(t, "setup/"+key)()
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

func TestSixRuleAgreementAndMutantsUnion(t *testing.T) { t.Parallel(); typeAwareTopUnion(t, "six") }
func TestSixRuleAgreementAndMutants_000(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 0) }
func TestSixRuleAgreementAndMutants_001(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 1) }
func TestSixRuleAgreementAndMutants_002(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 2) }
func TestSixRuleAgreementAndMutants_003(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 3) }
func TestSixRuleAgreementAndMutants_004(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 4) }
func TestSixRuleAgreementAndMutants_005(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 5) }
func TestSixRuleAgreementAndMutants_006(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 6) }
func TestSixRuleAgreementAndMutants_007(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 7) }
func TestSixRuleAgreementAndMutants_008(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 8) }
func TestSixRuleAgreementAndMutants_009(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 9) }
func TestSixRuleAgreementAndMutants_010(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 10) }
func TestSixRuleAgreementAndMutants_011(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 11) }
func TestSixRuleAgreementAndMutants_012(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 12) }
func TestSixRuleAgreementAndMutants_013(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 13) }
func TestSixRuleAgreementAndMutants_014(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 14) }
func TestSixRuleAgreementAndMutants_015(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 15) }
func TestSixRuleAgreementAndMutants_016(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 16) }
func TestSixRuleAgreementAndMutants_017(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 17) }
func TestSixRuleAgreementAndMutants_018(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 18) }
func TestSixRuleAgreementAndMutants_019(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 19) }
func TestSixRuleAgreementAndMutants_020(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 20) }
func TestSixRuleAgreementAndMutants_021(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 21) }
func TestSixRuleAgreementAndMutants_022(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 22) }
func TestSixRuleAgreementAndMutants_023(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 23) }
func TestSixRuleAgreementAndMutants_024(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 24) }
func TestSixRuleAgreementAndMutants_025(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 25) }
func TestSixRuleAgreementAndMutants_026(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 26) }
func TestSixRuleAgreementAndMutants_027(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 27) }
func TestSixRuleAgreementAndMutants_028(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 28) }
func TestSixRuleAgreementAndMutants_029(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 29) }
func TestSixRuleAgreementAndMutants_030(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 30) }
func TestSixRuleAgreementAndMutants_031(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 31) }
func TestSixRuleAgreementAndMutants_032(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 32) }
func TestSixRuleAgreementAndMutants_033(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 33) }
func TestSixRuleAgreementAndMutants_034(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 34) }
func TestSixRuleAgreementAndMutants_035(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 35) }
func TestSixRuleAgreementAndMutants_036(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 36) }
func TestSixRuleAgreementAndMutants_037(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 37) }
func TestSixRuleAgreementAndMutants_038(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 38) }
func TestSixRuleAgreementAndMutants_039(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 39) }
func TestSixRuleAgreementAndMutants_040(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 40) }
func TestSixRuleAgreementAndMutants_041(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 41) }
func TestSixRuleAgreementAndMutants_042(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 42) }
func TestSixRuleAgreementAndMutants_043(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 43) }
func TestSixRuleAgreementAndMutants_044(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 44) }
func TestSixRuleAgreementAndMutants_045(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 45) }
func TestSixRuleAgreementAndMutants_046(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 46) }
func TestSixRuleAgreementAndMutants_047(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 47) }
func TestSixRuleAgreementAndMutants_048(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 48) }
func TestSixRuleAgreementAndMutants_049(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 49) }
func TestSixRuleAgreementAndMutants_050(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 50) }
func TestSixRuleAgreementAndMutants_051(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 51) }
func TestSixRuleAgreementAndMutants_052(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 52) }
func TestSixRuleAgreementAndMutants_053(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 53) }
func TestSixRuleAgreementAndMutants_054(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 54) }
func TestSixRuleAgreementAndMutants_055(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 55) }
func TestSixRuleAgreementAndMutants_056(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 56) }
func TestSixRuleAgreementAndMutants_057(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 57) }
func TestSixRuleAgreementAndMutants_058(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 58) }
func TestSixRuleAgreementAndMutants_059(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 59) }
func TestSixRuleAgreementAndMutants_060(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 60) }
func TestSixRuleAgreementAndMutants_061(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 61) }
func TestSixRuleAgreementAndMutants_062(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 62) }
func TestSixRuleAgreementAndMutants_063(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 63) }
func TestSixRuleAgreementAndMutants_064(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 64) }
func TestSixRuleAgreementAndMutants_065(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 65) }
func TestSixRuleAgreementAndMutants_066(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 66) }
func TestSixRuleAgreementAndMutants_067(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 67) }
func TestSixRuleAgreementAndMutants_068(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 68) }
func TestSixRuleAgreementAndMutants_069(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 69) }
func TestSixRuleAgreementAndMutants_070(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 70) }
func TestSixRuleAgreementAndMutants_071(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 71) }
func TestSixRuleAgreementAndMutants_072(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 72) }
func TestSixRuleAgreementAndMutants_073(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 73) }
func TestSixRuleAgreementAndMutants_074(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 74) }
func TestSixRuleAgreementAndMutants_075(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 75) }
func TestSixRuleAgreementAndMutants_076(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 76) }
func TestSixRuleAgreementAndMutants_077(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 77) }
func TestSixRuleAgreementAndMutants_078(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 78) }
func TestSixRuleAgreementAndMutants_079(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 79) }
func TestSixRuleAgreementAndMutants_080(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 80) }
func TestSixRuleAgreementAndMutants_081(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 81) }
func TestSixRuleAgreementAndMutants_082(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 82) }
func TestSixRuleAgreementAndMutants_083(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 83) }
func TestSixRuleAgreementAndMutants_084(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 84) }
func TestSixRuleAgreementAndMutants_085(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 85) }
func TestSixRuleAgreementAndMutants_086(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 86) }
func TestSixRuleAgreementAndMutants_087(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 87) }
func TestSixRuleAgreementAndMutants_088(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 88) }
func TestSixRuleAgreementAndMutants_089(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 89) }
func TestSixRuleAgreementAndMutants_090(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 90) }
func TestSixRuleAgreementAndMutants_091(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 91) }
func TestSixRuleAgreementAndMutants_092(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 92) }
func TestSixRuleAgreementAndMutants_093(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 93) }
func TestSixRuleAgreementAndMutants_094(t *testing.T)  { t.Parallel(); typeAwareTopShard(t, "six", 94) }

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

// Not parallel: publishes the shared Six plan before parallel case leaves run.
func TestSixRuleAgreementAndMutants_Setup(t *testing.T) {
	typeAwareTopPlan(t, "six")
}

// Not parallel: publishes the shared TypeAware plan before parallel case leaves run.
func TestTypeAwareAgreementAndMutants_Setup(t *testing.T) {
	typeAwareTopPlan(t, "typeaware")
}

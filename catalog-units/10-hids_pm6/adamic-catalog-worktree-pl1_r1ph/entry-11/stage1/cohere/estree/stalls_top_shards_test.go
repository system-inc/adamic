package estree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These fixed top-level units are visible to go test -list without gate changes.
// ADAMIC_TEST_SHARD=i/n selects unit indexes modulo n; unset runs every unit.
// Each unit retains every implementation/check for its stable hash-assigned cases.
func TestBoundedPortParser_000(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 0) }
func TestBoundedPortParser_001(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 1) }
func TestBoundedPortParser_002(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 2) }
func TestBoundedPortParser_003(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 3) }
func TestBoundedPortParser_004(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 4) }
func TestBoundedPortParser_005(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 5) }
func TestBoundedPortParser_006(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 6) }
func TestBoundedPortParser_007(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 7) }
func TestBoundedPortParser_008(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 8) }
func TestBoundedPortParser_009(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 9) }
func TestBoundedPortParser_010(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 10) }
func TestBoundedPortParser_011(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 11) }
func TestBoundedPortParser_012(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 12) }
func TestBoundedPortParser_013(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 13) }
func TestBoundedPortParser_014(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 14) }
func TestBoundedPortParser_015(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 15) }
func TestBoundedPortParser_016(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 16) }
func TestBoundedPortParser_017(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 17) }
func TestBoundedPortParser_018(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 18) }
func TestBoundedPortParser_019(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 19) }
func TestBoundedPortParser_020(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 20) }
func TestBoundedPortParser_021(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 21) }
func TestBoundedPortParser_022(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 22) }
func TestBoundedPortParser_023(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 23) }
func TestBoundedPortParser_024(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 24) }
func TestBoundedPortParser_025(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 25) }
func TestBoundedPortParser_026(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 26) }
func TestBoundedPortParser_027(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 27) }
func TestBoundedPortParser_028(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 28) }
func TestBoundedPortParser_029(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 29) }
func TestBoundedPortParser_030(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 30) }
func TestBoundedPortParser_031(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 31) }
func TestBoundedPortParser_032(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 32) }
func TestBoundedPortParser_033(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 33) }
func TestBoundedPortParser_034(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 34) }
func TestBoundedPortParser_035(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 35) }
func TestBoundedPortParser_036(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 36) }
func TestBoundedPortParser_037(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 37) }
func TestBoundedPortParser_038(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 38) }
func TestBoundedPortParser_039(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 39) }
func TestBoundedPortParser_040(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 40) }
func TestBoundedPortParser_041(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 41) }
func TestBoundedPortParser_042(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 42) }
func TestBoundedPortParser_043(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 43) }
func TestBoundedPortParser_044(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 44) }
func TestBoundedPortParser_045(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 45) }
func TestBoundedPortParser_046(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 46) }
func TestBoundedPortParser_047(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 47) }
func TestBoundedPortParser_048(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 48) }
func TestBoundedPortParser_049(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 49) }
func TestBoundedPortParser_050(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 50) }
func TestBoundedPortParser_051(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 51) }
func TestBoundedPortParser_052(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 52) }
func TestBoundedPortParser_053(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 53) }
func TestBoundedPortParser_054(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 54) }
func TestBoundedPortParser_055(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 55) }
func TestBoundedPortParser_056(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 56) }
func TestBoundedPortParser_057(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 57) }
func TestBoundedPortParser_058(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 58) }
func TestBoundedPortParser_059(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 59) }
func TestBoundedPortParser_060(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 60) }
func TestBoundedPortParser_061(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 61) }
func TestBoundedPortParser_062(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 62) }
func TestBoundedPortParser_063(t *testing.T) { t.Parallel(); boundedPortParserShard(t, 63) }
func TestPortStallControl_000(t *testing.T)  { t.Parallel(); portStallControlShard(t, 0) }

func TestBoundedPortParserUnion(t *testing.T) {
	t.Parallel()
	estreeTopUnion(t, "TestBoundedPortParser", testBoundedPortParserShards, boundedPortIDs(boundedPortCases(t)))
}
func TestPortStallControlUnion(t *testing.T) {
	t.Parallel()
	estreeTopUnion(t, "TestPortStallControl", testPortStallControlShards, []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"})
}

func estreeScopedIndexes(t *testing.T, count int, ids []string, shard int) []int {
	t.Helper()
	plan, err := miscPlan(ids, count)
	if err != nil {
		t.Fatal(err)
	}
	if shard < 0 || shard >= len(plan) {
		t.Fatalf("invalid shard index %d", shard)
	}
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		selected, err := strconv.Atoi(parts[0])
		boxes, other := strconv.Atoi(parts[1])
		if err != nil || other != nil || boxes < 1 || boxes > count || selected < 0 || selected >= boxes {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		if shard%boxes != selected {
			t.Skip("unit is assigned to another instance")
		}
	}
	t.Logf("union: %d cases, %d unique ids, %d top-level shards", len(ids), len(ids), len(plan))
	return plan[shard]
}

func estreeTopUnion(t *testing.T, root string, count int, ids []string) {
	t.Helper()
	// miscPlan checks every live case is covered exactly once. Listing the actual
	// executable also catches a missing, repeated or extra top-level shard function.
	if _, err := miscPlan(ids, count); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, os.Args[0], "-test.list=^"+root+"_[0-9]{3}$").CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("cooked: listing top-level units exceeded 75 seconds")
	}
	if err != nil {
		t.Fatalf("list units: %v\n%s", err, output)
	}
	names := strings.Fields(string(output))
	if len(names) != count {
		t.Fatalf("enumerated %d top-level units, want %d", len(names), count)
	}
	seen := make(map[string]bool, count)
	for _, name := range names {
		if seen[name] {
			t.Fatalf("repeated top-level unit %s", name)
		}
		seen[name] = true
	}
	for i := range count {
		if name := fmt.Sprintf("%s_%03d", root, i); !seen[name] {
			t.Fatalf("missing top-level unit %s", name)
		}
	}
	t.Logf("union: %d live cases, exactly one assignment each, %d listed top-level units", len(ids), count)
}

// The proof drives the actual top-level wrappers and their hash planner with one
// checker-level disagreement/survivor. Every other top-level unit must pass.
func estreeScopedProof(t *testing.T, root string, ids []string, indexes []int) bool {
	t.Helper()
	if os.Getenv("ADAMIC_ESTREE_TOP_PROOF") != root {
		return false
	}
	for _, i := range indexes {
		planted := i == len(ids)-1
		var err error
		if root == "TestPortStallControl" {
			err = portStallControlResult(!planted, nil)
		} else {
			got := []byte("agree")
			if planted {
				got = []byte("disagree")
			}
			err = miscCompare([]byte("agree"), got, false)
		}
		if err != nil {
			t.Fatalf("%s: %v", ids[i], err)
		}
	}
	return true
}

func estreeTopPlantedProof(t *testing.T, root string, count int, ids []string) {
	t.Helper()
	if _, err := miscPlan(ids, count); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+root+"_[0-9]{3}$", "-test.v", "-test.timeout=75s")
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, "ADAMIC_ESTREE_TOP_PROOF=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_TOP_PROOF="+root)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("cooked: planted proof exceeded 75 seconds")
	}
	name := fmt.Sprintf("%s_%03d", root, miscShard(ids[len(ids)-1], count))
	if err == nil || bytes.Count(output, []byte("--- FAIL: "+root+"_")) != 1 || !bytes.Contains(output, []byte("--- FAIL: "+name+" ")) {
		t.Fatalf("planted failure must be caught only by %s: exit=%v\n%s", name, err, output)
	}
	t.Logf("planted case %s caught only by %s", ids[len(ids)-1], name)
}

package estree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// These fixed top-level units are visible to go test -list without gate changes.
// ADAMIC_TEST_SHARD=i/n selects unit indexes modulo n; unset runs every unit.
// Each unit retains every implementation/check for its stable hash-assigned cases.
func TestBoundedPortParser_000(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 0) }
func TestBoundedPortParser_001(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 1) }
func TestBoundedPortParser_002(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 2) }
func TestBoundedPortParser_003(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 3) }
func TestBoundedPortParser_004(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 4) }
func TestBoundedPortParser_005(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 5) }
func TestBoundedPortParser_006(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 6) }
func TestBoundedPortParser_007(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 7) }
func TestBoundedPortParser_008(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 8) }
func TestBoundedPortParser_009(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 9) }
func TestBoundedPortParser_010(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 10) }
func TestBoundedPortParser_011(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 11) }
func TestBoundedPortParser_012(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 12) }
func TestBoundedPortParser_013(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 13) }
func TestBoundedPortParser_014(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 14) }
func TestBoundedPortParser_015(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 15) }
func TestBoundedPortParser_016(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 16) }
func TestBoundedPortParser_017(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 17) }
func TestBoundedPortParser_018(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 18) }
func TestBoundedPortParser_019(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 19) }
func TestBoundedPortParser_020(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 20) }
func TestBoundedPortParser_021(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 21) }
func TestBoundedPortParser_022(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 22) }
func TestBoundedPortParser_023(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 23) }
func TestBoundedPortParser_024(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 24) }
func TestBoundedPortParser_025(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 25) }
func TestBoundedPortParser_026(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 26) }
func TestBoundedPortParser_027(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 27) }
func TestBoundedPortParser_028(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 28) }
func TestBoundedPortParser_029(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 29) }
func TestBoundedPortParser_030(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 30) }
func TestBoundedPortParser_031(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 31) }
func TestBoundedPortParser_032(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 32) }
func TestBoundedPortParser_033(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 33) }
func TestBoundedPortParser_034(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 34) }
func TestBoundedPortParser_035(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 35) }
func TestBoundedPortParser_036(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 36) }
func TestBoundedPortParser_037(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 37) }
func TestBoundedPortParser_038(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 38) }
func TestBoundedPortParser_039(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 39) }
func TestBoundedPortParser_040(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 40) }
func TestBoundedPortParser_041(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 41) }
func TestBoundedPortParser_042(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 42) }
func TestBoundedPortParser_043(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 43) }
func TestBoundedPortParser_044(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 44) }
func TestBoundedPortParser_045(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 45) }
func TestBoundedPortParser_046(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 46) }
func TestBoundedPortParser_047(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 47) }
func TestBoundedPortParser_048(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 48) }
func TestBoundedPortParser_049(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 49) }
func TestBoundedPortParser_050(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 50) }
func TestBoundedPortParser_051(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 51) }
func TestBoundedPortParser_052(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 52) }
func TestBoundedPortParser_053(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 53) }
func TestBoundedPortParser_054(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 54) }
func TestBoundedPortParser_055(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 55) }
func TestBoundedPortParser_056(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 56) }
func TestBoundedPortParser_057(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 57) }
func TestBoundedPortParser_058(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 58) }
func TestBoundedPortParser_059(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 59) }
func TestBoundedPortParser_060(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 60) }
func TestBoundedPortParser_061(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 61) }
func TestBoundedPortParser_062(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 62) }
func TestBoundedPortParser_063(test *testing.T) { test.Parallel(); boundedPortParserShard(test, 63) }
func TestPortStallControl_000(test *testing.T)  { test.Parallel(); portStallControlShard(test, 0) }

func TestBoundedPortParserUnion(test *testing.T) {
	test.Parallel()
	estreeTopUnion(test, "TestBoundedPortParser", testBoundedPortParserShards, boundedPortIds(boundedPortCases(test)))
}
func TestPortStallControlUnion(test *testing.T) {
	test.Parallel()
	estreeTopUnion(
		test,
		"TestPortStallControl",
		testPortStallControlShards,
		[]string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"},
	)
}

func estreeScopedIndexes(test *testing.T, count int, ids []string, shard int) []int {
	test.Helper()
	plan, miscPlanError := miscPlan(ids, count)
	if miscPlanError != nil {
		test.Fatal(miscPlanError)
	}
	if shard < 0 || shard >= len(plan) {
		test.Fatalf("invalid shard index %d", shard)
	}
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		selected, parseError := strconv.Atoi(parts[0])
		boxes, other := strconv.Atoi(parts[1])
		if parseError != nil || other != nil || boxes < 1 || boxes > count || selected < 0 || selected >= boxes {
			test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		if shard%boxes != selected {
			test.Skip("unit is assigned to another instance")
		}
	}
	test.Logf("union: %d cases, %d unique ids, %d top-level shards", len(ids), len(ids), len(plan))
	return plan[shard]
}

func estreeTopUnion(test *testing.T, root string, count int, ids []string) {
	test.Helper()
	// miscPlan checks every live case is covered exactly once. Listing the actual
	// executable also catches a missing, repeated or extra top-level shard function.
	if _, miscPlanError := miscPlan(ids, count); miscPlanError != nil {
		test.Fatal(miscPlanError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Read through childguard.Capture for estreeScopedCommand's reason, stdout then stderr as one listing.
	stdout, stderr, captureError := childguard.Capture(
		timeoutContext,
		estreeScopedCommand(timeoutContext, os.Args[0], "-test.list=^"+root+"_[0-9]{3}$"),
	)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatal("cooked: listing top-level units exceeded 90 seconds")
	}
	if captureError != nil {
		test.Fatalf("list units: %v\n%s", captureError, output)
	}
	names := strings.Fields(string(output))
	if len(names) != count {
		test.Fatalf("enumerated %d top-level units, want %d", len(names), count)
	}
	seen := make(map[string]bool, count)
	for _, name := range names {
		if seen[name] {
			test.Fatalf("repeated top-level unit %s", name)
		}
		seen[name] = true
	}
	for index := range count {
		if name := fmt.Sprintf("%s_%03d", root, index); !seen[name] {
			test.Fatalf("missing top-level unit %s", name)
		}
	}
	test.Logf("union: %d live cases, exactly one assignment each, %d listed top-level units", len(ids), count)
}

// The proof drives the actual top-level wrappers and their hash planner with one
// checker-level disagreement/survivor. Every other top-level unit must pass.
func estreeScopedProof(test *testing.T, root string, ids []string, indexes []int) bool {
	test.Helper()
	if os.Getenv("ADAMIC_ESTREE_TOP_PROOF") != root {
		return false
	}
	for _, index := range indexes {
		planted := index == len(ids)-1
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
			test.Fatalf("%s: %v", ids[index], err)
		}
	}
	return true
}

func estreeTopPlantedProof(test *testing.T, root string, count int, ids []string) {
	test.Helper()
	if _, miscPlanError := miscPlan(ids, count); miscPlanError != nil {
		test.Fatal(miscPlanError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := estreeScopedCommand(
		timeoutContext,
		os.Args[0],
		"-test.run=^"+root+"_[0-9]{3}$",
		"-test.v",
		"-test.timeout=75s",
	)
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, "ADAMIC_ESTREE_TOP_PROOF=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_TOP_PROOF="+root)
	// Read through childguard.Capture for estreeScopedCommand's reason, stdout then stderr.
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatal("cooked: planted proof exceeded 90 seconds")
	}
	name := fmt.Sprintf("%s_%03d", root, miscShard(ids[len(ids)-1], count))
	if captureError == nil ||
		bytes.Count(output, []byte("--- FAIL: "+root+"_")) != 1 ||
		!bytes.Contains(output, []byte("--- FAIL: "+name+" ")) {
		test.Fatalf("planted failure must be caught only by %s: exit=%v\n%s", name, captureError, output)
	}
	test.Logf("planted case %s caught only by %s", ids[len(ids)-1], name)
}

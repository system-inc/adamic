package oracle

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
	"time"
)

// Fixed, name-sorted round-robin partition. Registration order is irrelevant.
// Sixteen shards measured 63.639s with 16.729s shared setup. Doubling
// membership partitions leaves margin beneath the 45-second cold target.
const nativeOracleBaseShardCount = 32
const nativeOracleShardCount = 34

// Only measured heavy buckets are subdivided; other fixture ownership stays put.
func nativeOracleParts(base int) int {
	if base == 3 || base == 19 {
		return 2
	}
	return 1
}

func nativeOracleShardNames() []string {
	var names []string
	for base := 0; base < nativeOracleBaseShardCount; base++ {
		for part := 0; part < nativeOracleParts(base); part++ {
			name := fmt.Sprintf("shard-%03d", base)
			if nativeOracleParts(base) > 1 {
				name += fmt.Sprintf("-%d", part)
			}
			names = append(names, name)
		}
	}
	return names
}

func nativeOracleShards() [][]int {
	rows := make([]int, len(fixtures))
	for i := range rows {
		rows[i] = i
	}
	sort.Slice(rows, func(i, j int) bool { return fixtures[rows[i]].path < fixtures[rows[j]].path })
	buckets := make([][]int, nativeOracleBaseShardCount)
	for ordinal, row := range rows {
		buckets[ordinal%len(buckets)] = append(buckets[ordinal%len(buckets)], row)
	}
	var shards [][]int
	for base, rows := range buckets {
		parts := make([][]int, nativeOracleParts(base))
		for ordinal, row := range rows {
			parts[ordinal%len(parts)] = append(parts[ordinal%len(parts)], row)
		}
		shards = append(shards, parts...)
	}
	return shards
}

func nativeOracleUnion(shards [][]int) error {
	seen := make([]bool, len(fixtures))
	names := make(map[string]bool)
	for shard, rows := range shards {
		if len(rows) == 0 {
			return fmt.Errorf("empty shard %03d", shard)
		}
		for _, row := range rows {
			if row < 0 || row >= len(fixtures) {
				return fmt.Errorf("unexpected fixture %d", row)
			}
			if seen[row] || names[fixtures[row].path] {
				return fmt.Errorf("duplicate fixture %s", fixtures[row].path)
			}
			seen[row] = true
			names[fixtures[row].path] = true
		}
	}
	for row, present := range seen {
		if !present {
			return fmt.Errorf("missing fixture %s", fixtures[row].path)
		}
	}
	return nil
}

func TestNativeOracleShardUnion(t *testing.T) {
	shards := nativeOracleShards()
	if len(shards) != nativeOracleShardCount {
		t.Fatalf("declared %d shards, got %d", nativeOracleShardCount, len(shards))
	}
	if err := nativeOracleUnion(shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d registered fixtures, exactly once, %d shards", len(fixtures), len(shards))
	if err := nativeOracleUnion(shards[1:]); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("omitted shard accepted: %v", err)
	}
	duplicate := append(append([][]int(nil), shards...), shards[0])
	if err := nativeOracleUnion(duplicate); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate shard accepted: %v", err)
	}
	// Prove that a change in registration order cannot change path ownership.
	for shard, rows := range shards {
		for _, row := range rows {
			rank := 0
			for _, fixture := range fixtures {
				if fixture.path < fixtures[row].path {
					rank++
				}
			}
			base := rank % nativeOracleBaseShardCount
			expected := (rank / nativeOracleBaseShardCount) % nativeOracleParts(base)
			for previous := 0; previous < base; previous++ {
				expected += nativeOracleParts(previous)
			}
			if expected != shard {
				t.Fatalf("unstable ownership for %s", fixtures[row].path)
			}
		}
	}
}

// Run the owning shard with a mutated lowered dedication, through precisely the
// live fixture comparison. The subprocess must fail and identify that fixture.
func TestNativeOracleShardPlantedFailure(t *testing.T) {
	planted, owner := -1, -1
	shards := nativeOracleShards()
	for shard, rows := range shards {
		for _, row := range rows {
			if fixtures[row].path == "dedication/dedication.a" {
				planted, owner = row, shard
			}
		}
	}
	if planted < 0 {
		t.Fatal("planted fixture missing from registration")
	}
	if os.Getenv("ADAMIC_ORACLE_SHARD_PLANTED_CHILD") == "1" {
		t.Run(nativeOracleShardNames()[owner], func(t *testing.T) { runNativeOracleShard(t, shards[owner], planted) })
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestNativeOracleShardPlantedFailure$", "-test.parallel=4", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_ORACLE_SHARD_PLANTED_CHILD=1", "ADAMIC_GATE_UNCACHED=1")
	// The outer test's timeout is also bounded by the measurement harness.
	command.WaitDelay = time.Second
	output, err := command.CombinedOutput()
	want := "--- FAIL: TestNativeOracleShardPlantedFailure/" + nativeOracleShardNames()[owner] + "/dedication/dedication.a"
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), want) || !strings.Contains(string(output), "stdout differs") {
		t.Fatalf("planted failure lost (exit %v), want %s:\n%s", err, want, output)
	}
	t.Logf("owning %s caught planted fixture %s through live comparison", nativeOracleShardNames()[owner], fixtures[planted].path)
}

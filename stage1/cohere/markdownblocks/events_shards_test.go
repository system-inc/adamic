package markdownblocks

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Bound both transport size and case count. A document is indivisible: a large
// document gets its own unit, preserving tokenizer state across the whole file.
func tokenizerEventShards(inputs [][]uint16) [][]int {
	const maxUnits = 100000
	const maxCases = 2048
	var shards [][]int
	var rows []int
	units := 0
	for i, input := range inputs {
		if len(rows) > 0 && (units+len(input) > maxUnits || len(rows) == maxCases) {
			shards = append(shards, rows)
			rows, units = nil, 0
		}
		rows = append(rows, i)
		units += len(input)
	}
	if len(rows) > 0 {
		shards = append(shards, rows)
	}
	// Smaller corpus modes use the same static gate grid. Split contiguous
	// ranges deterministically; larger corpora that need more units fail the
	// declared-count assertion rather than silently changing gate membership.
	for len(shards) < testTokenizerEventsShards-3 {
		longest := -1
		for i, rows := range shards {
			if len(rows) > 1 && (longest < 0 || len(rows) > len(shards[longest])) {
				longest = i
			}
		}
		if longest < 0 {
			break
		}
		rows := shards[longest]
		middle := len(rows) / 2
		shards = append(shards, nil)
		copy(shards[longest+2:], shards[longest+1:])
		shards[longest], shards[longest+1] = rows[:middle], rows[middle:]
	}
	return shards
}

func tokenizerEventUnion(total int, shards [][]int) error {
	seen := make([]bool, total)
	count := 0
	for shard, rows := range shards {
		if len(rows) == 0 {
			return fmt.Errorf("empty shard %d", shard)
		}
		for _, row := range rows {
			if row < 0 || row >= total {
				return fmt.Errorf("shard %d: unexpected case id %d", shard, row)
			}
			if seen[row] {
				return fmt.Errorf("shard %d: repeated case id %d", shard, row)
			}
			seen[row] = true
			count++
		}
	}
	for row, present := range seen {
		if !present {
			return fmt.Errorf("missing case id %d", row)
		}
	}
	if count != total {
		return fmt.Errorf("union count %d, want %d", count, total)
	}
	return nil
}

func validateTokenizerEventUnion(t *testing.T, total int, shards [][]int) {
	t.Helper()
	if err := tokenizerEventUnion(total, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d/%d cases; exact id set; no repeats; %d shards", total, total, len(shards))
}

func tokenizerEventSelection(t *testing.T) func(int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return func(int) bool { return true }
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatalf("ADAMIC_TEST_SHARD=%q: expected zero-based i/n", value)
	}
	i, errI := strconv.Atoi(parts[0])
	n, errN := strconv.Atoi(parts[1])
	if errI != nil || errN != nil || n <= 0 || i < 0 || i >= n {
		t.Fatalf("ADAMIC_TEST_SHARD=%q: require 0 <= i < n", value)
	}
	return func(ordinal int) bool { return ordinal%n == i }
}

// The ordinal is the case id from the unsplit enumeration. It distinguishes
// even two corpus entries with identical names or identical text.
func tokenizerEventDifference(shard, side string, want, got []byte, rows []int) error {
	if bytes.Equal(want, got) {
		return nil
	}
	expected := bytes.Split(bytes.TrimSuffix(want, []byte("\n")), []byte("\n"))
	observed := bytes.Split(bytes.TrimSuffix(got, []byte("\n")), []byte("\n"))
	for i, row := range rows {
		if i >= len(expected) || i >= len(observed) || !bytes.Equal(expected[i], observed[i]) {
			return fmt.Errorf("%s: %s disagreement in case id %d", shard, side, row)
		}
	}
	return fmt.Errorf("%s: %s disagreement in output framing or extra rows", shard, side)
}

// The shared buildcache helper is absent on this base. This is deliberately a
// single-run builder, with no package cache. Each caller declares its inputs
// beside a func(dir string) error that writes its product inside dir.
type tokenizerEventBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func tokenizerEventProduct(t *testing.T, inputs tokenizerEventBuildInputs, dir string, build func(string) error) {
	t.Helper()
	started := time.Now()
	if err := build(dir); err != nil {
		t.Fatal(err)
	}
	t.Logf("build cold %s: %.3fs; toolchain %s", inputs.Name, time.Since(started).Seconds(), inputs.Toolchain)
}

func TestTokenizerEventShardUnion(t *testing.T) {
	inputs := make([][]uint16, 5000)
	inputs[2300] = make([]uint16, 200000)
	shards := tokenizerEventShards(inputs)
	validateTokenizerEventUnion(t, len(inputs), shards)
	for _, bad := range [][][]int{
		{{0}, {0, 1}}, // repeated
		{{0}},         // missing
		{{0}, {1, 2}}, // unexpected
	} {
		if err := tokenizerEventUnion(2, bad); err == nil {
			t.Fatal("invalid union accepted")
		}
	}
}

// Plant a disagreement in one case and use the same comparison as the live
// oracle. Exactly its owning shard must reject it, and name itself and the id.
func TestTokenizerEventShardPlantedDisagreement(t *testing.T) {
	inputs := make([][]uint16, 5000)
	shards := tokenizerEventShards(inputs)
	const planted = 2345
	caught := 0
	for ordinal, rows := range shards {
		shard := fmt.Sprintf("shard-%03d", ordinal)
		want := bytes.Repeat([]byte("agreed\n"), len(rows))
		got := bytes.Clone(want)
		owner := false
		for i, row := range rows {
			if row == planted {
				got[i*len("agreed\n")] = 'X'
				owner = true
			}
		}
		err := tokenizerEventDifference(shard, "planted oracle", want, got, rows)
		if owner {
			if err == nil || !strings.Contains(err.Error(), shard) || !strings.Contains(err.Error(), "2345") {
				t.Fatalf("planted disagreement lost: %v", err)
			}
			t.Logf("caught by %s: %v", shard, err)
			caught++
		} else if err != nil {
			t.Fatalf("unrelated shard failed: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted case caught by %d shards, want exactly one", caught)
	}
}

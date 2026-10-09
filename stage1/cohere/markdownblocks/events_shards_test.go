package markdownblocks

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A case's index is local to its path/generated name, never its list position.
// Repository files are mutable; the fixed bucket grid does not follow corpus size.
func tokenizerEventKeys(names []string) []string {
	keys := make([]string, len(names))
	occurrences := make(map[string]int)
	for i, name := range names {
		keys[i] = fmt.Sprintf("%s\x00case=%d", name, occurrences[name])
		occurrences[name]++
	}
	return keys
}

func tokenizerEventShard(key string) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(key))
	return int(hash.Sum64() % uint64(testTokenizerEventsShards))
}

func tokenizerEventShards(keys []string) [][]int {
	shards := make([][]int, testTokenizerEventsShards)
	for row, key := range keys {
		shard := tokenizerEventShard(key)
		shards[shard] = append(shards[shard], row)
	}
	return shards
}

func tokenizerEventUnion(total int, shards [][]int) error {
	seen := make([]bool, total)
	count := 0
	for shard, rows := range shards {
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
	t.Parallel()
	keys := tokenizerEventFixtureKeys()
	shards := tokenizerEventShards(keys)
	validateTokenizerEventUnion(t, len(keys), shards)
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
	t.Parallel()
	keys := tokenizerEventFixtureKeys()
	shards := tokenizerEventShards(keys)
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

func tokenizerEventFixtureKeys() []string {
	names := make([]string, 5000)
	for i := range names {
		names[i] = fmt.Sprintf("fixture/file-%d.md", i)
	}
	return tokenizerEventKeys(names)
}

func TestTokenizerEventShardGrowth(t *testing.T) {
	t.Parallel()
	names := []string{"a.md", "a.md", "b.md", "generated/units/0"}
	before := tokenizerEventKeys(names)
	// Inserting an earlier-sorted file and another file with two modes changes
	// row positions but must not move any existing stable case key.
	after := tokenizerEventKeys(append([]string{"0-new.md", "new.md", "new.md"}, names...))
	shards := tokenizerEventShards(after)
	if len(shards) != testTokenizerEventsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testTokenizerEventsShards)
	}
	validateTokenizerEventUnion(t, len(after), shards)
	owners := make(map[string]int)
	for shard, rows := range shards {
		for _, row := range rows {
			if _, repeated := owners[after[row]]; repeated {
				t.Fatalf("repeated stable key %q", after[row])
			}
			owners[after[row]] = shard
		}
	}
	for i, key := range before {
		if after[i+3] != key || owners[key] != tokenizerEventShard(key) {
			t.Fatalf("existing case moved: %q", key)
		}
	}
}

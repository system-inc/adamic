package native

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const randomRegexCases = 10000
const randomRegexUnitSize = 250
const decodePrefixes = 92014
const decodePrefixUnitSize = 2048
const normalizePointUnitSize = 65536

var splitCacheModes = []string{"0", "1"}
var decodeTargets = []string{"native", "wasi"}

type testShard struct{ index, count int }

func parseTestShard(value string) (testShard, error) {
	if value == "" {
		return testShard{0, 1}, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return testShard{}, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n with 0 <= i < n: %q", value)
	}
	index, firstError := strconv.Atoi(parts[0])
	count, secondError := strconv.Atoi(parts[1])
	if firstError != nil || secondError != nil || count < 1 || index < 0 || index >= count {
		return testShard{}, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n with 0 <= i < n: %q", value)
	}
	return testShard{index, count}, nil
}
func currentTestShard(t *testing.T) testShard {
	t.Helper()
	shard, err := parseTestShard(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	return shard
}
func (s testShard) owns(piece int) bool { return piece%s.count == s.index }

func TestShardSelection(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"-1/2", "0/0", "2/2", "one/2", "0/2/3", "/2"} {
		if _, err := parseTestShard(value); err == nil {
			t.Errorf("invalid shard accepted: %q", value)
		}
	}
	if shard, err := parseTestShard(""); err != nil || shard != (testShard{0, 1}) {
		t.Fatalf("default shard: %+v %v", shard, err)
	}
}

type testUnitRange struct{ first, last int }

func (piece testUnitRange) name() string {
	return fmt.Sprintf("cases_%06d_%06d", piece.first, piece.last)
}
func unitRanges(total, width int) []testUnitRange {
	var pieces []testUnitRange
	for first := 0; first < total; first += width {
		pieces = append(pieces, testUnitRange{first, min(first+width, total)})
	}
	return pieces
}
func runRegexPartitions(t *testing.T, cases []regexCase) {
	t.Helper()
	if len(cases) != randomRegexCases {
		t.Fatalf("random regex corpus: got %d, want %d", len(cases), randomRegexCases)
	}
	shard := currentTestShard(t)
	for index, piece := range unitRanges(len(cases), randomRegexUnitSize) {
		if !shard.owns(index) {
			continue
		}
		t.Run(piece.name(), func(t *testing.T) { runRegexCases(t, cases[piece.first:piece.last]) })
	}
}

// Frozen counts describe the old workloads. A missing, duplicated or overlapping
// piece fails independently of whether all the remaining observations agree.
func TestSplitUnitCoverage(t *testing.T) {
	t.Parallel()
	for _, plan := range []struct {
		name                string
		total, width, units int
	}{
		{"random_regex", 10000, randomRegexUnitSize, 40},
		{"decode_prefixes", 92014, decodePrefixUnitSize, 45},
		{"normalize_points", 0x110000, normalizePointUnitSize, 17},
	} {
		t.Run(plan.name, func(t *testing.T) {
			pieces := unitRanges(plan.total, plan.width)
			if len(pieces) != plan.units {
				t.Fatalf("got %d units, want %d", len(pieces), plan.units)
			}
			next := 0
			for _, piece := range pieces {
				if piece.first != next || piece.last <= piece.first || piece.last > plan.total {
					t.Fatalf("invalid coverage at %+v after %d", piece, next)
				}
				next = piece.last
			}
			if next != plan.total {
				t.Fatalf("covered %d pieces, want %d", next, plan.total)
			}
		})
	}
	t.Run("shard_union", func(t *testing.T) {
		for _, pieces := range []int{90, 40, 18, 36, 2, 6} {
			for count := 1; count <= 100; count++ {
				seen := make([]int, pieces)
				for index := 0; index < count; index++ {
					shard, err := parseTestShard(fmt.Sprintf("%d/%d", index, count))
					if err != nil {
						t.Fatal(err)
					}
					for piece := 0; piece < pieces; piece++ {
						if shard.owns(piece) {
							seen[piece]++
						}
					}
				}
				for piece, owners := range seen {
					if owners != 1 {
						t.Fatalf("piece %d of %d has %d owners across %d shards", piece, pieces, owners, count)
					}
				}
			}
		}
	})
	t.Run("decode_targets", func(t *testing.T) {
		if strings.Join(decodeTargets, ",") != "native,wasi" {
			t.Fatal("native and WASI decoder targets must both run")
		}
	})
	t.Run("WASI", func(t *testing.T) {
		if len(wasiFixtures) != 35 {
			t.Fatalf("got %d WASI fixtures, want 35", len(wasiFixtures))
		}
		want := "a7207f91f36f405d152fc3e2be706f49eecd3dbd916c0c7f5f0ae3303a56497a"
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(wasiFixtures, "\n")))); got != want {
			t.Fatalf("WASI fixture membership changed: %s", got)
		}
	})
	t.Run("split_cache", func(t *testing.T) {
		if strings.Join(splitCacheModes, ",") != "0,1" {
			t.Fatal("cached and bypass builds must both run")
		}
	})
	t.Run("record_mutants", func(t *testing.T) {
		if len(recordMutationCases) != 6 {
			t.Fatalf("got %d record mutants, want 6", len(recordMutationCases))
		}
		seen := map[string]bool{}
		for _, mutant := range recordMutationCases {
			if seen[mutant.name] {
				t.Fatalf("duplicate record mutant %s", mutant.name)
			}
			seen[mutant.name] = true
		}
	})
}

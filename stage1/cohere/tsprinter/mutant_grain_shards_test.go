package tsprinter

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// The explicitly generated mutant set has one bounded unit per mutation.
// Repository corpus growth does not change these units.
func TestMutantsUnion(test *testing.T) {
	test.Parallel()
	if len(oracles.Mutations) != oracles.MutantsShards {
		test.Fatalf(
			"%d mutants, want %d declared product pairs and leaves",
			len(oracles.Mutations),
			oracles.MutantsShards,
		)
	}
	if len(mutantShardFunctions) != oracles.MutantsShards {
		test.Fatalf("%d top-level shards, want %d", len(mutantShardFunctions), oracles.MutantsShards)
	}
	seen := make([]int, len(oracles.Mutations))
	for shard := 0; shard < oracles.MutantsShards; shard++ {
		for _, index := range mutantShardIndices(shard) {
			if index < 0 || index >= len(seen) {
				test.Fatalf("shard %d: invalid mutant %d", shard, index)
			}
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("mutant %q covered %d times", oracles.Mutations[index].Name, count)
		}
	}
	if len(seen) == 0 {
		test.Fatal("empty mutant corpus")
	}
}

func mutantShardIndices(shard int) []int {
	var indices []int
	for index := shard; index < len(oracles.Mutations); index += oracles.MutantsShards {
		indices = append(indices, index)
	}
	return indices
}

func runMutantShard(test *testing.T, shard int) {
	test.Helper()
	for _, index := range mutantShardIndices(shard) {
		testMutant(test, index)
	}
}

func TestMutants_000(test *testing.T) { test.Parallel(); runMutantShard(test, 0) }
func TestMutants_001(test *testing.T) { test.Parallel(); runMutantShard(test, 1) }
func TestMutants_002(test *testing.T) { test.Parallel(); runMutantShard(test, 2) }
func TestMutants_003(test *testing.T) { test.Parallel(); runMutantShard(test, 3) }
func TestMutants_004(test *testing.T) { test.Parallel(); runMutantShard(test, 4) }
func TestMutants_005(test *testing.T) { test.Parallel(); runMutantShard(test, 5) }
func TestMutants_006(test *testing.T) { test.Parallel(); runMutantShard(test, 6) }
func TestMutants_007(test *testing.T) { test.Parallel(); runMutantShard(test, 7) }
func TestMutants_008(test *testing.T) { test.Parallel(); runMutantShard(test, 8) }
func TestMutants_009(test *testing.T) { test.Parallel(); runMutantShard(test, 9) }
func TestMutants_010(test *testing.T) { test.Parallel(); runMutantShard(test, 10) }
func TestMutants_011(test *testing.T) { test.Parallel(); runMutantShard(test, 11) }
func TestMutants_012(test *testing.T) { test.Parallel(); runMutantShard(test, 12) }
func TestMutants_013(test *testing.T) { test.Parallel(); runMutantShard(test, 13) }
func TestMutants_014(test *testing.T) { test.Parallel(); runMutantShard(test, 14) }
func TestMutants_015(test *testing.T) { test.Parallel(); runMutantShard(test, 15) }
func TestMutants_016(test *testing.T) { test.Parallel(); runMutantShard(test, 16) }
func TestMutants_017(test *testing.T) { test.Parallel(); runMutantShard(test, 17) }
func TestMutants_018(test *testing.T) { test.Parallel(); runMutantShard(test, 18) }
func TestMutants_019(test *testing.T) { test.Parallel(); runMutantShard(test, 19) }
func TestMutants_020(test *testing.T) { test.Parallel(); runMutantShard(test, 20) }
func TestMutants_021(test *testing.T) { test.Parallel(); runMutantShard(test, 21) }
func TestMutants_022(test *testing.T) { test.Parallel(); runMutantShard(test, 22) }
func TestMutants_023(test *testing.T) { test.Parallel(); runMutantShard(test, 23) }
func TestMutants_024(test *testing.T) { test.Parallel(); runMutantShard(test, 24) }
func TestMutants_025(test *testing.T) { test.Parallel(); runMutantShard(test, 25) }
func TestMutants_026(test *testing.T) { test.Parallel(); runMutantShard(test, 26) }
func TestMutants_027(test *testing.T) { test.Parallel(); runMutantShard(test, 27) }
func TestMutants_028(test *testing.T) { test.Parallel(); runMutantShard(test, 28) }

// Compile references and check the declared count against the top-level units.
var mutantShardFunctions = [...]func(*testing.T){
	TestMutants_000,
	TestMutants_001,
	TestMutants_002,
	TestMutants_003,
	TestMutants_004,
	TestMutants_005,
	TestMutants_006,
	TestMutants_007,
	TestMutants_008,
	TestMutants_009,
	TestMutants_010,
	TestMutants_011,
	TestMutants_012,
	TestMutants_013,
	TestMutants_014,
	TestMutants_015,
	TestMutants_016,
	TestMutants_017,
	TestMutants_018,
	TestMutants_019,
	TestMutants_020,
	TestMutants_021,
	TestMutants_022,
	TestMutants_023,
	TestMutants_024,
	TestMutants_025,
	TestMutants_026,
	TestMutants_027,
	TestMutants_028,
}

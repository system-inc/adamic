package tsprinter

import "testing"

// The explicitly generated mutant set has one bounded unit per mutation.
// Repository corpus growth does not change these units.
func TestMutantsUnion(t *testing.T) {
	t.Parallel()
	if len(mutantShardFunctions) != testMutantsShards {
		t.Fatalf("%d top-level shards, want %d", len(mutantShardFunctions), testMutantsShards)
	}
	seen := make([]int, len(mutations))
	for shard := 0; shard < testMutantsShards; shard++ {
		for _, index := range mutantShardIndices(shard) {
			if index < 0 || index >= len(seen) {
				t.Fatalf("shard %d: invalid mutant %d", shard, index)
			}
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Fatalf("mutant %q covered %d times", mutations[index].name, count)
		}
	}
	if len(seen) == 0 {
		t.Fatal("empty mutant corpus")
	}
}

func mutantShardIndices(shard int) []int {
	var indices []int
	for index := shard; index < len(mutations); index += testMutantsShards {
		indices = append(indices, index)
	}
	return indices
}

func runMutantShard(t *testing.T, shard int) {
	t.Helper()
	for _, index := range mutantShardIndices(shard) {
		testMutant(t, index)
	}
}

func TestMutants_000(t *testing.T) { t.Parallel(); runMutantShard(t, 0) }
func TestMutants_001(t *testing.T) { t.Parallel(); runMutantShard(t, 1) }
func TestMutants_002(t *testing.T) { t.Parallel(); runMutantShard(t, 2) }
func TestMutants_003(t *testing.T) { t.Parallel(); runMutantShard(t, 3) }
func TestMutants_004(t *testing.T) { t.Parallel(); runMutantShard(t, 4) }
func TestMutants_005(t *testing.T) { t.Parallel(); runMutantShard(t, 5) }
func TestMutants_006(t *testing.T) { t.Parallel(); runMutantShard(t, 6) }
func TestMutants_007(t *testing.T) { t.Parallel(); runMutantShard(t, 7) }
func TestMutants_008(t *testing.T) { t.Parallel(); runMutantShard(t, 8) }
func TestMutants_009(t *testing.T) { t.Parallel(); runMutantShard(t, 9) }
func TestMutants_010(t *testing.T) { t.Parallel(); runMutantShard(t, 10) }
func TestMutants_011(t *testing.T) { t.Parallel(); runMutantShard(t, 11) }
func TestMutants_012(t *testing.T) { t.Parallel(); runMutantShard(t, 12) }
func TestMutants_013(t *testing.T) { t.Parallel(); runMutantShard(t, 13) }
func TestMutants_014(t *testing.T) { t.Parallel(); runMutantShard(t, 14) }
func TestMutants_015(t *testing.T) { t.Parallel(); runMutantShard(t, 15) }
func TestMutants_016(t *testing.T) { t.Parallel(); runMutantShard(t, 16) }
func TestMutants_017(t *testing.T) { t.Parallel(); runMutantShard(t, 17) }
func TestMutants_018(t *testing.T) { t.Parallel(); runMutantShard(t, 18) }
func TestMutants_019(t *testing.T) { t.Parallel(); runMutantShard(t, 19) }
func TestMutants_020(t *testing.T) { t.Parallel(); runMutantShard(t, 20) }
func TestMutants_021(t *testing.T) { t.Parallel(); runMutantShard(t, 21) }
func TestMutants_022(t *testing.T) { t.Parallel(); runMutantShard(t, 22) }
func TestMutants_023(t *testing.T) { t.Parallel(); runMutantShard(t, 23) }
func TestMutants_024(t *testing.T) { t.Parallel(); runMutantShard(t, 24) }
func TestMutants_025(t *testing.T) { t.Parallel(); runMutantShard(t, 25) }
func TestMutants_026(t *testing.T) { t.Parallel(); runMutantShard(t, 26) }
func TestMutants_027(t *testing.T) { t.Parallel(); runMutantShard(t, 27) }
func TestMutants_028(t *testing.T) { t.Parallel(); runMutantShard(t, 28) }

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

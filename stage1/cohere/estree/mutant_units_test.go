package estree

import (
	"os"
	"strconv"
	"testing"
)

// Each top-level shard is a gate leaf, including its hashed product fetches.
// ADAMIC_TEST_SHARD=i/n selects indices modulo n; unset runs every shard.
func runTopLevelMutantShard(t *testing.T, cases []string, mutations []portMutation, count, index int) {
	t.Helper()
	if !selectedShard(t, index) {
		t.Skip("ADAMIC_TEST_SHARD selected another shard")
	}
	t.Parallel()
	shards := mutantShardPlan(t, cases, mutations, count)
	shard := shards[index]
	m := shard.cases[0].mutant
	if value := os.Getenv("ADAMIC_ESTREE_PLANTED_MUTANT"); value != "" {
		planted, err := strconv.Atoi(value)
		if err != nil || planted < 0 || planted >= len(mutations) {
			t.Fatal("invalid planted mutant")
		}
		for _, pair := range shard.cases {
			want, source, native := []byte("oracle"), []byte("oracle"), []byte("oracle")
			if pair.witness {
				source, native = []byte("mutant"), []byte("mutant")
			}
			if pair.mutant == planted && pair.witness {
				native = want
			}
			if err := checkMutantCase(pair, want, source, native); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	oracle := goOracle(t)
	sourcePath, binary := mutantProduct(t, mutations[m])
	texts := make([]string, len(shard.cases))
	for i, pair := range shard.cases {
		texts[i] = cases[pair.index]
	}
	list := manifest(t, texts)
	want := canonicalRecords(t, execute(t, "", oracle, "--manifest", list), len(texts))
	source := canonicalRecords(t, onNode(t, sourcePath, "--manifest", list), len(texts))
	native := canonicalRecords(t, execute(t, "", binary, "--manifest", list), len(texts))
	for i, pair := range shard.cases {
		if err := checkMutantCase(pair, want[i], source[i], native[i]); err != nil {
			t.Fatalf("%s %s: %v", shard.name, mutations[m].name, err)
		}
	}
	t.Logf("%s: %s cases [%d,%d), %d pairs; Go/source Node/sanitized native checked", shard.name, mutations[m].name, shard.cases[0].index, shard.cases[len(shard.cases)-1].index+1, len(texts))
}

func TestThreePortMutants_000(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 0)
}

func TestThreePortMutants_001(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 1)
}

func TestThreePortMutants_002(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 2)
}

func TestThreePortMutants_003(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 3)
}

func TestThreePortMutants_004(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 4)
}

func TestThreePortMutants_005(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 5)
}

func TestThreePortMutants_006(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 6)
}

func TestThreePortMutants_007(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 7)
}

func TestThreePortMutants_008(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 8)
}

func TestThreePortMutants_009(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 9)
}

func TestThreePortMutants_010(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 10)
}

func TestThreePortMutants_011(t *testing.T) {
	runTopLevelMutantShard(t, generated(), threePortMutations(), testThreePortMutantsShards, 11)
}

func TestAcceptanceMutants_000(t *testing.T) {
	runTopLevelMutantShard(t, acceptanceGrammar(), acceptanceMutations(), testAcceptanceMutantsShards, 0)
}

func TestAcceptanceMutants_001(t *testing.T) {
	runTopLevelMutantShard(t, acceptanceGrammar(), acceptanceMutations(), testAcceptanceMutantsShards, 1)
}

func TestDeepMutants_000(t *testing.T) {
	runTopLevelMutantShard(t, deepMutantCases(), deepMutations(), testDeepMutantsShards, 0)
}

func TestDeepMutants_001(t *testing.T) {
	runTopLevelMutantShard(t, deepMutantCases(), deepMutations(), testDeepMutantsShards, 1)
}

func TestDeepMutants_002(t *testing.T) {
	runTopLevelMutantShard(t, deepMutantCases(), deepMutations(), testDeepMutantsShards, 2)
}

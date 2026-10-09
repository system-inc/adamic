package tsprinter

import "testing"

// Each shard is a gate-discoverable top-level leaf. ADAMIC_TEST_SHARD=i/n
// selects shard numbers modulo n; unset runs all sixteen fixed hash slices.
func statementTopLevelShard(t *testing.T, number int) {
	t.Helper()
	t.Parallel()
	if statementShardProofEnabled() {
		statementRunShardProof(t, number)
		return
	}
	statementAgainstGoAndPrettierShard(t, number)
}

func TestStatementsAgainstGoAndPrettier_000(t *testing.T) { statementTopLevelShard(t, 0) }
func TestStatementsAgainstGoAndPrettier_001(t *testing.T) { statementTopLevelShard(t, 1) }
func TestStatementsAgainstGoAndPrettier_002(t *testing.T) { statementTopLevelShard(t, 2) }
func TestStatementsAgainstGoAndPrettier_003(t *testing.T) { statementTopLevelShard(t, 3) }
func TestStatementsAgainstGoAndPrettier_004(t *testing.T) { statementTopLevelShard(t, 4) }
func TestStatementsAgainstGoAndPrettier_005(t *testing.T) { statementTopLevelShard(t, 5) }
func TestStatementsAgainstGoAndPrettier_006(t *testing.T) { statementTopLevelShard(t, 6) }
func TestStatementsAgainstGoAndPrettier_007(t *testing.T) { statementTopLevelShard(t, 7) }
func TestStatementsAgainstGoAndPrettier_008(t *testing.T) { statementTopLevelShard(t, 8) }
func TestStatementsAgainstGoAndPrettier_009(t *testing.T) { statementTopLevelShard(t, 9) }
func TestStatementsAgainstGoAndPrettier_010(t *testing.T) { statementTopLevelShard(t, 10) }
func TestStatementsAgainstGoAndPrettier_011(t *testing.T) { statementTopLevelShard(t, 11) }
func TestStatementsAgainstGoAndPrettier_012(t *testing.T) { statementTopLevelShard(t, 12) }
func TestStatementsAgainstGoAndPrettier_013(t *testing.T) { statementTopLevelShard(t, 13) }
func TestStatementsAgainstGoAndPrettier_014(t *testing.T) { statementTopLevelShard(t, 14) }
func TestStatementsAgainstGoAndPrettier_015(t *testing.T) { statementTopLevelShard(t, 15) }

// Enumerate the live corpus and validate all slices before any box selection.
func TestStatementsAgainstGoAndPrettierUnion(t *testing.T) {
	t.Parallel()
	cases, want, specs := statementCorpus(t)
	shards := statementShards(t, cases, want, specs)
	if len(shards) != testStatementsAgainstGoAndPrettierShards {
		t.Fatalf("enumerated %d shards", len(shards))
	}
	total := 0
	for _, shard := range shards {
		total += len(shard.indices)
	}
	t.Logf("live union: %d unique cases in %d shards", total, len(shards))
}

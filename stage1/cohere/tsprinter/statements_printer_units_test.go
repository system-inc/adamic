package tsprinter

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// Each shard is a gate-discoverable top-level leaf. ADAMIC_TEST_SHARD=i/n
// selects shard numbers modulo n; unset runs all sixteen fixed hash slices.
func statementTopLevelShard(test *testing.T, number int) {
	test.Helper()
	if statementShardProofEnabled() {
		statementRunShardProof(test, number)
		return
	}
	statementAgainstGoAndPrettierShard(test, number)
}

func TestStatementsAgainstGoAndPrettier_000(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 0)
}
func TestStatementsAgainstGoAndPrettier_001(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 1)
}
func TestStatementsAgainstGoAndPrettier_002(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 2)
}
func TestStatementsAgainstGoAndPrettier_003(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 3)
}
func TestStatementsAgainstGoAndPrettier_004(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 4)
}
func TestStatementsAgainstGoAndPrettier_005(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 5)
}
func TestStatementsAgainstGoAndPrettier_006(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 6)
}
func TestStatementsAgainstGoAndPrettier_007(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 7)
}
func TestStatementsAgainstGoAndPrettier_008(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 8)
}
func TestStatementsAgainstGoAndPrettier_009(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 9)
}
func TestStatementsAgainstGoAndPrettier_010(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 10)
}
func TestStatementsAgainstGoAndPrettier_011(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 11)
}
func TestStatementsAgainstGoAndPrettier_012(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 12)
}
func TestStatementsAgainstGoAndPrettier_013(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 13)
}
func TestStatementsAgainstGoAndPrettier_014(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 14)
}
func TestStatementsAgainstGoAndPrettier_015(test *testing.T) {
	test.Parallel()
	statementTopLevelShard(test, 15)
}

// Enumerate the live corpus and validate all slices before any box selection.
func TestStatementsAgainstGoAndPrettierUnion(test *testing.T) {
	test.Parallel()
	products := statementReadyShared(test)
	cases, want, specs := products.Cases, products.Want, products.Specs
	shards := oracles.StatementShards(test, cases, want, specs)
	if len(shards) != oracles.StatementsAgainstGoAndPrettierShards {
		test.Fatalf("enumerated %d shards", len(shards))
	}
	total := 0
	for _, shard := range shards {
		total += len(shard.Indices)
	}
	test.Logf("live union: %d unique cases in %d shards", total, len(shards))
}

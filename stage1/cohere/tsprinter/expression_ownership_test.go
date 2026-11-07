package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func TestExpressionOwnershipStable(test *testing.T) {
	test.Parallel()
	root, _ := filepath.Abs(oracles.Repository)
	cases := []oracles.PrinterCase{
		{Label: root + "/stage1/a.ts:1:Identifier"},
		{Label: root + "/stage1/a.ts:2:Identifier"},
		{Label: "generated"},
		{Label: "generated"},
	}
	keys := oracles.ExpressionGateKeys(test, cases)
	grown := append([]oracles.PrinterCase{{Label: root + "/bench/new.ts:0:Identifier"}}, cases...)
	next := oracles.ExpressionGateKeys(test, grown)
	for index, key := range keys {
		if next[index+1] != key {
			test.Fatalf("growth moved %s to %s", key, next[index+1])
		}
	}
	old, added := oracles.ExpressionGateOwners(keys, 64), oracles.ExpressionGateOwners(next, 64)
	for owner, ids := range old {
		for _, id := range ids {
			found := false
			for _, candidate := range added[owner] {
				found = found || candidate == id+1
			}
			if !found {
				test.Fatalf("case %d changed owner", id)
			}
		}
	}
}

func expressionGateProofEnabled() bool { return os.Getenv("ADAMIC_EXPRESSIONS_SHARD_PROOF") == "1" }
func expressionGateProofKeys() []string {
	keys := make([]string, 256)
	for index := range keys {
		keys[index] = fmt.Sprintf("stage1/cohere/tsprinter/testdata/expressions_side_test.go:proof:%d", index)
	}
	return keys
}
func expressionGateProofShard(test *testing.T, number int) {
	test.Helper()
	if number == oracles.ExpressionsAgainstGoAndPrettierShards-1 {
		return
	}
	ids := oracles.ExpressionGateOwners(
		expressionGateProofKeys(),
		oracles.ExpressionsAgainstGoAndPrettierShards-1,
	)[number]
	plan := &oracles.CorpusPlan{
		Labels: make([]string, 256),
		Shards: make([]oracles.CorpusShard, oracles.ExpressionsAgainstGoAndPrettierShards-1),
	}
	plan.Shards[number].Indices = ids
	var input, want strings.Builder
	local := -1
	for position, id := range ids {
		fmt.Fprintf(&input, ">%d;\n", id)
		fmt.Fprintf(&want, "ok\t%d;\\n\n", id)
		plan.Labels[id] = fmt.Sprintf("planted case %d", id)
		if id == 17 {
			local = position
		}
	}
	path := filepath.Join(test.TempDir(), "proof.txt")
	if writeError := os.WriteFile(path, []byte(input.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	port, _ := filepath.Abs("main.ts")
	result := onNode(test, port, "--cases", path, "80")
	if expressionDisagreementError := expressionDisagreement(
		plan,
		number,
		[]byte(want.String()),
		result,
	); expressionDisagreementError != nil {
		test.Fatal(expressionDisagreementError)
	}
	if local >= 0 {
		rows := strings.Split(string(result.Stdout), "\n")
		rows[local] = "ok\tplanted disagreement"
		result.Stdout = []byte(strings.Join(rows, "\n"))
	}
	if expressionDisagreementError := expressionDisagreement(
		plan,
		number,
		[]byte(want.String()),
		result,
	); expressionDisagreementError != nil {
		test.Fatal(expressionDisagreementError)
	}
}
func TestExpressionsShardPlantedDisagreement(test *testing.T) {
	test.Parallel()
	target := -1
	for owner, ids := range oracles.ExpressionGateOwners(
		expressionGateProofKeys(),
		oracles.ExpressionsAgainstGoAndPrettierShards-1,
	) {
		for _, id := range ids {
			if id == 17 {
				target = owner
			}
		}
	}
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	result := oracles.StatementExecute(
		test,
		[]string{"ADAMIC_EXPRESSIONS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="},
		binary,
		"-test.run=^TestExpressionsAgainstGoAndPrettier_[0-9]{3}$",
		"-test.v",
		"-test.parallel=4",
		"-test.timeout=75s",
	)
	name := fmt.Sprintf("TestExpressionsAgainstGoAndPrettier_%03d", target)
	if result.ExitCode != 1 ||
		len(result.Stderr) != 0 ||
		strings.Count(string(result.Stdout), "--- FAIL: TestExpressionsAgainstGoAndPrettier_") != 1 ||
		!strings.Contains(string(result.Stdout), "--- FAIL: "+name) ||
		!strings.Contains(string(result.Stdout), "planted case 17") {
		test.Fatalf(
			"expected exactly %s to catch planted case: exit %d stdout %s stderr %s",
			name,
			result.ExitCode,
			result.Stdout,
			result.Stderr,
		)
	}
	test.Logf("planted case 17 caught by exactly %s", name)
}

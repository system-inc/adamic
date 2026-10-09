package tsprinter

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expressionGateKeys(t *testing.T, cases []printerCase) []string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := statementCaseKeys(cases, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	// Generated expression cases are owned by their actual generator file.
	for i, key := range keys {
		keys[i] = strings.Replace(key, "testdata/statements_side_test.go:", "testdata/expressions_side_test.go:", 1)
	}
	return keys
}

func expressionGateOwners(keys []string, count int) [][]int {
	shards := make([][]int, count)
	for id, key := range keys {
		digest := sha256.Sum256([]byte(key))
		var value uint64
		for _, b := range digest[:8] {
			value = value<<8 | uint64(b)
		}
		owner := int(value % uint64(count))
		shards[owner] = append(shards[owner], id)
	}
	return shards
}

func expressionGatePartition(t *testing.T, cases []printerCase, count int) [][]int {
	t.Helper()
	keys := expressionGateKeys(t, cases)
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			t.Fatalf("repeated stable expression key %q", key)
		}
		seen[key] = true
	}
	return expressionGateOwners(keys, count)
}

func TestExpressionOwnershipStable(t *testing.T) {
	t.Parallel()
	root, _ := filepath.Abs(repository)
	cases := []printerCase{{Label: root + "/stage1/a.ts:1:Identifier"}, {Label: root + "/stage1/a.ts:2:Identifier"}, {Label: "generated"}, {Label: "generated"}}
	keys := expressionGateKeys(t, cases)
	grown := append([]printerCase{{Label: root + "/bench/new.ts:0:Identifier"}}, cases...)
	next := expressionGateKeys(t, grown)
	for i, key := range keys {
		if next[i+1] != key {
			t.Fatalf("growth moved %s to %s", key, next[i+1])
		}
	}
	old, added := expressionGateOwners(keys, 64), expressionGateOwners(next, 64)
	for owner, ids := range old {
		for _, id := range ids {
			found := false
			for _, candidate := range added[owner] {
				found = found || candidate == id+1
			}
			if !found {
				t.Fatalf("case %d changed owner", id)
			}
		}
	}
}

func expressionGateProofEnabled() bool { return os.Getenv("ADAMIC_EXPRESSIONS_SHARD_PROOF") == "1" }
func expressionGateProofKeys() []string {
	keys := make([]string, 256)
	for i := range keys {
		keys[i] = fmt.Sprintf("stage1/cohere/tsprinter/testdata/expressions_side_test.go:proof:%d", i)
	}
	return keys
}
func expressionGateProofShard(t *testing.T, number int) {
	t.Helper()
	if number == testExpressionsAgainstGoAndPrettierShards-1 {
		return
	}
	ids := expressionGateOwners(expressionGateProofKeys(), testExpressionsAgainstGoAndPrettierShards-1)[number]
	plan := &corpusPlan{labels: make([]string, 256), shards: make([]corpusShard, testExpressionsAgainstGoAndPrettierShards-1)}
	plan.shards[number].indices = ids
	var input, want strings.Builder
	local := -1
	for position, id := range ids {
		fmt.Fprintf(&input, ">%d;\n", id)
		fmt.Fprintf(&want, "ok\t%d;\\n\n", id)
		plan.labels[id] = fmt.Sprintf("planted case %d", id)
		if id == 17 {
			local = position
		}
	}
	path := filepath.Join(t.TempDir(), "proof.txt")
	if err := os.WriteFile(path, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	port, _ := filepath.Abs("main.ts")
	result := onNode(t, port, "--cases", path, "80")
	if err := expressionDisagreement(plan, number, []byte(want.String()), result); err != nil {
		t.Fatal(err)
	}
	if local >= 0 {
		rows := strings.Split(string(result.stdout), "\n")
		rows[local] = "ok\tplanted disagreement"
		result.stdout = []byte(strings.Join(rows, "\n"))
	}
	if err := expressionDisagreement(plan, number, []byte(want.String()), result); err != nil {
		t.Fatal(err)
	}
}
func TestExpressionsShardPlantedDisagreement(t *testing.T) {
	t.Parallel()
	target := -1
	for owner, ids := range expressionGateOwners(expressionGateProofKeys(), testExpressionsAgainstGoAndPrettierShards-1) {
		for _, id := range ids {
			if id == 17 {
				target = owner
			}
		}
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result := statementExecute(t, []string{"ADAMIC_EXPRESSIONS_SHARD_PROOF=1", "ADAMIC_TEST_SHARD="}, binary, "-test.run=^TestExpressionsAgainstGoAndPrettier_[0-9]{3}$", "-test.v", "-test.parallel=4", "-test.timeout=75s")
	name := fmt.Sprintf("TestExpressionsAgainstGoAndPrettier_%03d", target)
	if result.exitCode != 1 || len(result.stderr) != 0 || strings.Count(string(result.stdout), "--- FAIL: TestExpressionsAgainstGoAndPrettier_") != 1 || !strings.Contains(string(result.stdout), "--- FAIL: "+name) || !strings.Contains(string(result.stdout), "planted case 17") {
		t.Fatalf("expected exactly %s to catch planted case: exit %d stdout %s stderr %s", name, result.exitCode, result.stdout, result.stderr)
	}
	t.Logf("planted case 17 caught by exactly %s", name)
}

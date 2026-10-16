package estree

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// refusedBeforeDeadline holds a refusal of argv (a program, then its input path) to its calibrated budget.
func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	refusedWithin(t, argv, diagnostic, estreeRefusalBudget(t, argv))
}

func refusedWithin(t *testing.T, argv []string, diagnostic string, budget time.Duration) {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	err, timedOut, cpu := runWithCPUBudget(t, argv, output, &stderr, budget)
	output.Close()
	info, _ := os.Stat(output.Name())
	if timedOut || err == nil || info.Size() != 0 || !strings.Contains(stderr.String(), diagnostic) {
		t.Fatalf("%v: timeout=%v cpu=%s budget=%s exit=%v stdout=%d stderr=%s", argv, timedOut, cpu, budget, err, info.Size(), &stderr)
	}
}

const testBoundedPortParserShards = 64

type boundedPortCase struct {
	id, text, filename string
	audit              bool
}

func boundedPortCases(t *testing.T) []boundedPortCase {
	inputs := oracles.BoundedCases(t)
	cases := make([]boundedPortCase, len(inputs))
	for i, item := range inputs {
		cases[i] = boundedPortCase{item.ID, item.Text, item.Filename, item.Audit}
	}
	return cases
}

func boundedPortIDs(cases []boundedPortCase) []string {
	ids := make([]string, len(cases))
	for i := range cases {
		ids[i] = cases[i].id
	}
	return ids
}

// 64 fixed hash shards leave headroom for the live repository corpus. Keys are
// repository-relative fixture paths plus case index/mode; new files move no cases.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every live fixture and EOF case.
func boundedPortParserShard(t *testing.T, shard int) {
	finishSetup := miscStart(t)
	cases := boundedPortCases(t)
	ids := boundedPortIDs(cases)
	indexes := estreeScopedIndexes(t, testBoundedPortParserShards, ids, shard)
	if estreeScopedProof(t, "TestBoundedPortParser", ids, indexes) || len(indexes) == 0 {
		finishSetup()
		return
	}
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := miscBuild(t, main)
	oracle := miscOracle(t)
	audited := estreeBoundedAudits(t)
	answers := miscAnswers(t, oracle, audited)
	byID := make(map[string]miscAnswer, len(audited))
	for i, item := range audited {
		byID[item.id] = answers[i]
	}
	finishSetup()
	for _, i := range indexes {
		item := cases[i]
		path := filepath.Join(t.TempDir(), item.filename)
		if err := os.WriteFile(path, []byte(item.text), 0644); err != nil {
			t.Fatal(err)
		}
		accepted := item.audit && byID[item.id].Status == "ok"
		if accepted {
			want := byID[item.id].Data
			for name, got := range map[string][]byte{"Node": estreeScopedNode(t, main, path), "native": estreeScopedExecute(t, binary, path), "emitted": estreeScopedNode(t, script, path)} {
				if err := miscCompare(want, got, false); err != nil {
					t.Fatalf("%s %s: %v", item.id, name, err)
				}
			}
		} else {
			for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
				estreeScopedRefusal(t, argv, "ESTree parser")
			}
		}
	}
}

func TestBoundedPortParserPlantedDisagreement(t *testing.T) {
	t.Parallel()
	estreeTopPlantedProof(t, "TestBoundedPortParser", testBoundedPortParserShards, boundedPortIDs(boundedPortCases(t)))
}

const testPortStallControlShards = 1

func portStallControlResult(timedOut bool, err error) error {
	if !timedOut {
		return fmt.Errorf("stall control must hit deadline, got %v", err)
	}
	return nil
}

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs the single guard-disabled
// mutant case. Both source Node and sanitized native stay together in its leaf.
func portStallControlShard(t *testing.T, shard int) {
	finishSetup := miscStart(t)
	ids := []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"}
	indexes := estreeScopedIndexes(t, testPortStallControlShards, ids, shard)
	if estreeScopedProof(t, "TestPortStallControl", ids, indexes) || len(indexes) == 0 {
		finishSetup()
		return
	}
	main := estreeStallMutantPath(t)
	binary, _ := miscBuild(t, main)
	finishSetup()
	for range indexes {
		path := filepath.Join(t.TempDir(), "input.ts")
		if err := os.WriteFile(path, []byte("class C { ) }"), 0644); err != nil {
			t.Fatal(err)
		}
		// The guard-disabled port held to the calibrated budget a refusal gets: it must run until the budget stops it. A
		// 500 ms wall deadline caught it before, and Node's startup alone outlasts that whether the guard holds or not.
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}} {
			budget := estreeRefusalBudget(t, argv)
			output, err := os.CreateTemp(t.TempDir(), "control-log")
			if err != nil {
				t.Fatal(err)
			}
			err, timedOut, cpu := runWithCPUBudget(t, argv, output, output, budget)
			output.Close()
			if err := portStallControlResult(timedOut, err); err != nil {
				t.Fatal(err)
			}
			t.Logf("%s guard-disabled control ran to its budget: %s of CPU against a budget of %s", argv[0], cpu, budget)
		}
	}
}

func TestPortStallControlPlantedSurvivor(t *testing.T) {
	t.Parallel()
	estreeTopPlantedProof(t, "TestPortStallControl", testPortStallControlShards, []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"})
}

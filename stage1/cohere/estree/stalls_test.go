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
func refusedBeforeDeadline(test *testing.T, argv []string, diagnostic string) {
	test.Helper()
	refusedWithin(test, argv, diagnostic, estreeRefusalBudget(test, argv))
}

func refusedWithin(test *testing.T, argv []string, diagnostic string, budget time.Duration) {
	test.Helper()
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	var stderr bytes.Buffer
	runWithCpuBudgetError, timedOut, cpu := runWithCpuBudget(test, argv, output, &stderr, budget)
	output.Close()
	info, _ := os.Stat(output.Name())
	if timedOut || runWithCpuBudgetError == nil || info.Size() != 0 || !strings.Contains(stderr.String(), diagnostic) {
		test.Fatalf(
			"%v: timeout=%v cpu=%s budget=%s exit=%v stdout=%d stderr=%s",
			argv,
			timedOut,
			cpu,
			budget,
			runWithCpuBudgetError,
			info.Size(),
			&stderr,
		)
	}
}

const testBoundedPortParserShards = 64

type boundedPortCase struct {
	id, text, filename string
	audit              bool
}

func boundedPortCases(test *testing.T) []boundedPortCase {
	inputs := oracles.BoundedCases(test)
	cases := make([]boundedPortCase, len(inputs))
	for index, item := range inputs {
		cases[index] = boundedPortCase{item.ID, item.Text, item.Filename, item.Audit}
	}
	return cases
}

func boundedPortIds(cases []boundedPortCase) []string {
	ids := make([]string, len(cases))
	for index := range cases {
		ids[index] = cases[index].id
	}
	return ids
}

// 64 fixed hash shards leave headroom for the live repository corpus. Keys are
// repository-relative fixture paths plus case index/mode; new files move no cases.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every live fixture and EOF case.
func boundedPortParserShard(test *testing.T, shard int) {
	finishSetup := miscStart(test)
	cases := boundedPortCases(test)
	ids := boundedPortIds(cases)
	indexes := estreeScopedIndexes(test, testBoundedPortParserShards, ids, shard)
	if estreeScopedProof(test, "TestBoundedPortParser", ids, indexes) || len(indexes) == 0 {
		finishSetup()
		return
	}
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary, script := miscBuild(test, main)
	oracle := miscOracle(test)
	audited := estreeBoundedAudits(test)
	answers := miscAnswers(test, oracle, audited)
	byId := make(map[string]miscAnswer, len(audited))
	for index, item := range audited {
		byId[item.id] = answers[index]
	}
	finishSetup()
	for _, index := range indexes {
		item := cases[index]
		path := filepath.Join(test.TempDir(), item.filename)
		if writeError := os.WriteFile(path, []byte(item.text), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		accepted := item.audit && byId[item.id].Status == "ok"
		if accepted {
			want := byId[item.id].Data
			for name, got := range map[string][]byte{
				"Node":    estreeScopedNode(test, main, path),
				"native":  estreeScopedExecute(test, binary, path),
				"emitted": estreeScopedNode(test, script, path),
			} {
				if miscCompareError := miscCompare(want, got, false); miscCompareError != nil {
					test.Fatalf("%s %s: %v", item.id, name, miscCompareError)
				}
			}
		} else {
			for _, argv := range [][]string{
				{
					"node",
					"--disable-warning=ExperimentalWarning",
					filepath.Join(root(test), "oracle/node.mjs"),
					main,
					path,
				},
				{binary, path},
				{
					"node",
					"--disable-warning=ExperimentalWarning",
					filepath.Join(root(test), "oracle/node.mjs"),
					script,
					path,
				},
			} {
				estreeScopedRefusal(test, argv, "ESTree parser")
			}
		}
	}
}

func TestBoundedPortParserPlantedDisagreement(test *testing.T) {
	test.Parallel()
	estreeTopPlantedProof(
		test,
		"TestBoundedPortParser",
		testBoundedPortParserShards,
		boundedPortIds(boundedPortCases(test)),
	)
}

const testPortStallControlShards = 1

func portStallControlResult(timedOut bool, sourceError error) error {
	if !timedOut {
		return fmt.Errorf("stall control must hit deadline, got %v", sourceError)
	}
	return nil
}

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs the single guard-disabled
// mutant case. Both source Node and sanitized native stay together in its leaf.
func portStallControlShard(test *testing.T, shard int) {
	finishSetup := miscStart(test)
	ids := []string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"}
	indexes := estreeScopedIndexes(test, testPortStallControlShards, ids, shard)
	if estreeScopedProof(test, "TestPortStallControl", ids, indexes) || len(indexes) == 0 {
		finishSetup()
		return
	}
	main := estreeStallMutantPath(test)
	binary, _ := miscBuild(test, main)
	finishSetup()
	for range indexes {
		path := filepath.Join(test.TempDir(), "input.ts")
		if writeError := os.WriteFile(path, []byte("class C { ) }"), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		// The guard-disabled port held to the calibrated budget a refusal gets: it must run until the budget stops it.
		// A 500 ms wall deadline caught it before, and Node's startup alone outlasts that whether the guard holds or
		// not.
		for _, argv := range [][]string{
			{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(test), "oracle/node.mjs"), main, path},
			{binary, path},
		} {
			budget := estreeRefusalBudget(test, argv)
			output, temporaryFileError := os.CreateTemp(test.TempDir(), "control-log")
			if temporaryFileError != nil {
				test.Fatal(temporaryFileError)
			}
			runWithCpuBudgetError, timedOut, cpu := runWithCpuBudget(test, argv, output, output, budget)
			output.Close()
			if portStallControlResultError := portStallControlResult(
				timedOut,
				runWithCpuBudgetError,
			); portStallControlResultError != nil {
				test.Fatal(portStallControlResultError)
			}
			test.Logf(
				"%s guard-disabled control ran to its budget: %s of CPU against a budget of %s",
				argv[0],
				cpu,
				budget,
			)
		}
	}
}

func TestPortStallControlPlantedSurvivor(test *testing.T) {
	test.Parallel()
	estreeTopPlantedProof(
		test,
		"TestPortStallControl",
		testPortStallControlShards,
		[]string{"stage1/cohere/estree/stalls_test.go:0:guard-disabled"},
	)
}

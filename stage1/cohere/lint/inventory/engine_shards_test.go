package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testInventoryEngineShards = 8

// Keep headroom independent of the live case count. Test names are stable modes
// within this repository-relative file; enumeration order never enters the key.
const inventoryEngineCaseFile = "stage1/cohere/lint/inventory/testdata/engine_test.go"

// inventoryEngineBuild is the engine's test binary, built ahead from the same two complete files as the unsplit suite
// laid over cohere (buildcache.GoTest), and the directory it runs in. With mutant, the unknown-frequency mutant is
// planted into testdata/engine.go inside the product (buildcache.GoTestMutated). Neither changes the checkout, and
// both are Go builds keyed by go list, never by hand-listed inputs.
func inventoryEngineBuild(test *testing.T, mutant bool) (string, string) {
	test.Helper()
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	cohere := filepath.Join(root, "cohere")
	virtual := filepath.Join(cohere, "adamic_inventory.go")
	virtualTest := filepath.Join(cohere, "adamic_inventory_test.go")
	overlay, encodeError := json.Marshal(map[string]any{"Replace": map[string]string{
		virtual:     filepath.Join(root, inventoryUnknownFrequencyMutant.File),
		virtualTest: filepath.Join(root, inventoryEngineCaseFile),
	}})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	path := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	arguments := []string{"-overlay=" + path}
	if mutant {
		return buildcache.GoTestMutated(
			test,
			"cohere",
			"engine.test",
			virtual+" "+virtualTest,
			arguments,
			[]buildcache.Mutation{inventoryUnknownFrequencyMutant},
		), cohere
	}
	return buildcache.GoTest(test, "cohere", "engine.test", virtual+" "+virtualTest, arguments), cohere
}

// inventoryUnknownFrequencyMutant is a real production mutant: unknown frequency becomes measured zero.
var inventoryUnknownFrequencyMutant = buildcache.Mutation{
	File:   "stage1/cohere/lint/inventory/testdata/engine.go",
	Before: "func countText(f frequency) string {\n\tif f.Count == nil {\n\t\treturn \"unknown\"",
	After:  "func countText(f frequency) string {\n\tif f.Count == nil {\n\t\treturn \"0\"",
}

func TestProduct_InventoryEngine(test *testing.T) {
	test.Parallel()
	inventoryEngineBuild(test, false)
}

func TestProduct_InventoryEngineUnknownFrequencyMutant(test *testing.T) {
	test.Parallel()
	inventoryEngineBuild(test, true)
}

func inventoryEngineShard(name string) int {
	hash := fnv.New64a()
	hash.Write([]byte(inventoryEngineCaseFile + "#" + name))
	return int(hash.Sum64() % uint64(testInventoryEngineShards))
}

func inventoryEnginePartition(cases []string) ([][]string, error) {
	if len(cases) == 0 {
		return nil, fmt.Errorf("inventory engine corpus is empty")
	}
	shards := make([][]string, testInventoryEngineShards)
	seen := map[string]bool{}
	for _, name := range cases {
		if seen[name] {
			return nil, fmt.Errorf("repeated case %s", name)
		}
		seen[name] = true
		index := inventoryEngineShard(name)
		shards[index] = append(shards[index], name)
	}
	return shards, nil
}

func inventoryEngineUnion(test *testing.T, binary, directory string) [][]string {
	test.Helper()
	command := inventoryEngineSetupCommand(binary, "-test.list=^Test")
	command.Dir = directory
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("enumerate engine: %v\n%s", combinedOutputError, output)
	}
	unsplit := strings.Fields(string(output))
	shards, inventoryEnginePartitionError := inventoryEnginePartition(unsplit)
	if inventoryEnginePartitionError != nil {
		test.Fatal(inventoryEnginePartitionError)
	}
	if len(shards) != testInventoryEngineShards {
		test.Fatalf("enumerated %d shards, want %d", len(shards), testInventoryEngineShards)
	}
	union := []string{}
	seen := map[string]bool{}
	for index, cases := range shards {
		for _, name := range cases {
			if seen[name] {
				test.Fatalf("repeated case %s in shard union", name)
			}
			if inventoryEngineShard(name) != index {
				test.Fatalf("case %s assigned to wrong shard", name)
			}
			seen[name] = true
			union = append(union, name)
		}
	}
	slices.Sort(union)
	slices.Sort(unsplit)
	if !slices.Equal(union, unsplit) {
		test.Fatalf("shard union %v differs from live enumeration %v", union, unsplit)
	}
	test.Logf(
		"union: %d unique cases equal the complete live enumeration across %d fixed shards",
		len(union),
		len(shards),
	)
	return shards
}

// Adding or reordering cases must not move an existing case; empty and repeated
// corpora must fail rather than silently produce a passing set of shards.
func TestInventoryEngineShardAssignment(test *testing.T) {
	test.Parallel()
	original := []string{"TestExistingA", "TestExistingB", "TestExistingC"}
	before, inventoryEnginePartitionError := inventoryEnginePartition(original)
	if inventoryEnginePartitionError != nil {
		test.Fatal(inventoryEnginePartitionError)
	}
	grown := append([]string{"TestAdded"}, original...)
	slices.Reverse(grown)
	after, inventoryEnginePartitionError := inventoryEnginePartition(grown)
	if inventoryEnginePartitionError != nil {
		test.Fatal(inventoryEnginePartitionError)
	}
	if len(before) != testInventoryEngineShards || len(after) != testInventoryEngineShards {
		test.Fatal("growth changed shard count")
	}
	for index, cases := range before {
		for _, name := range cases {
			if !slices.Contains(after[index], name) {
				test.Fatalf("growth moved case %s out of shard-%03d", name, index)
			}
		}
	}
	total := 0
	for _, cases := range after {
		total += len(cases)
	}
	if total != len(grown) {
		test.Fatalf("growth lost cases: %d, want live total %d", total, len(grown))
	}
	for _, invalid := range [][]string{nil, {"TestRepeated", "TestRepeated"}} {
		if _, inventoryEnginePartitionError := inventoryEnginePartition(invalid); inventoryEnginePartitionError == nil {
			test.Fatalf("invalid corpus accepted: %v", invalid)
		}
	}
}

func inventoryEngineRun(test *testing.T, binary, directory, name string) ([]byte, error) {
	command, ctx, cancel := inventoryEngineTestCommand(
		test,
		binary,
		"-test.run=^"+regexp.QuoteMeta(name)+"$",
		"-test.count=1",
		"-test.v",
		"-test.timeout=75s",
	)
	defer cancel()
	command.Dir = directory
	return inventoryEngineOutput(ctx, command)
}

func inventoryEngineSelector(value string) (func(int) bool, error) {
	if value == "" {
		return func(int) bool { return true }, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n, got %q", value)
	}
	index, indexError := strconv.Atoi(parts[0])
	count, countError := strconv.Atoi(parts[1])
	if indexError != nil || countError != nil || count < 1 || index < 0 || index >= count {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return func(shard int) bool { return shard%count == index }, nil
}

// A real production mutant: unknown frequency becomes measured zero. Run every
// live case against it and require exactly its stable hash shard to reject it.
func TestInventoryEngineShardMutant(test *testing.T) {
	test.Parallel()
	binary, cohere := inventoryEngineBuild(test, true)
	shards := inventoryEngineUnion(test, binary, cohere)
	caught := []string{}
	for index, cases := range shards {
		for _, name := range cases {
			output, inventoryEngineRunError := inventoryEngineRun(test, binary, cohere, name)
			if inventoryEngineRunError == nil {
				continue
			}
			if name != "TestUnknownFrequencyIsNotZero" ||
				!strings.Contains(string(output), "unmeasured count printed as zero") {
				test.Fatalf(
					"unexpected mutant failure in shard-%03d (%s): %v\n%s",
					index,
					name,
					inventoryEngineRunError,
					output,
				)
			}
			caught = append(caught, fmt.Sprintf("shard-%03d", index))
		}
	}
	expected := fmt.Sprintf("shard-%03d", inventoryEngineShard("TestUnknownFrequencyIsNotZero"))
	if !slices.Equal(caught, []string{expected}) {
		test.Fatalf("unknown-frequency mutant caught by %v, want exactly %s", caught, expected)
	}
	test.Logf("unknown-frequency mutant caught by exactly %s (TestUnknownFrequencyIsNotZero)", expected)
}

// The engine's build is shared by every top-level shard, and its union is validated once.
var inventoryEngineSharedBuild struct {
	once   sync.Once
	binary string
	cohere string
	shards [][]string
}

func inventoryEngineShared(test *testing.T) (string, string, [][]string) {
	test.Helper()
	shared := &inventoryEngineSharedBuild
	shared.once.Do(func() {
		shared.binary, shared.cohere = inventoryEngineBuild(test, false)
		shared.shards = inventoryEngineUnion(test, shared.binary, shared.cohere)
	})
	if shared.binary == "" || len(shared.shards) != testInventoryEngineShards {
		test.Fatal("shared inventory build or union validation did not complete")
	}
	return shared.binary, shared.cohere, shared.shards
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
// Assignment hashes the repository-relative case file plus the test name.
func inventoryEngineTopLevelShard(test *testing.T, index int) {
	test.Helper()
	selector, inventoryEngineSelectorError := inventoryEngineSelector(os.Getenv("ADAMIC_TEST_SHARD"))
	if inventoryEngineSelectorError != nil {
		test.Fatal(inventoryEngineSelectorError)
	}
	if !selector(index) {
		test.Skip("outside ADAMIC_TEST_SHARD selection")
	}
	binary, directory, shards := inventoryEngineShared(test)
	inventoryEngineDeadline(test, binary, directory, shards[index])
}

func TestInventoryEngine_000(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 0) }
func TestInventoryEngine_001(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 1) }
func TestInventoryEngine_002(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 2) }
func TestInventoryEngine_003(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 3) }
func TestInventoryEngine_004(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 4) }
func TestInventoryEngine_005(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 5) }
func TestInventoryEngine_006(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 6) }
func TestInventoryEngine_007(test *testing.T) { test.Parallel(); inventoryEngineTopLevelShard(test, 7) }

var inventoryEngineTopLevelTests = [...]func(*testing.T){
	TestInventoryEngine_000,
	TestInventoryEngine_001,
	TestInventoryEngine_002,
	TestInventoryEngine_003,
	TestInventoryEngine_004,
	TestInventoryEngine_005,
	TestInventoryEngine_006,
	TestInventoryEngine_007,
}

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
	"time"
)

const testInventoryEngineShards = 8

// Keep headroom independent of the live case count. Test names are stable modes
// within this repository-relative file; enumeration order never enters the key.
const inventoryEngineCaseFile = "stage1/cohere/lint/inventory/testdata/engine_test.go"

// Build from the same two complete files as the unsplit suite. All output goes
// into directory; the overlay and mutant never change the checkout or a product.
func inventoryEngineBuild(t *testing.T, mutant bool) (string, string, time.Duration) {
	t.Helper()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := ""
	if mutant {
		directory = t.TempDir()
	} else {
		directory, err = os.MkdirTemp("", "inventory-engine-shared-")
		if err != nil {
			t.Fatal(err)
		}
		inventoryEngineScratch = directory
	}
	cohere := filepath.Join(root, "cohere")
	virtual := filepath.Join(cohere, "adamic_inventory.go")
	virtualTest := filepath.Join(cohere, "adamic_inventory_test.go")
	source := filepath.Join(root, "stage1/cohere/lint/inventory/testdata/engine.go")
	testSource := filepath.Join(root, "stage1/cohere/lint/inventory/testdata/engine_test.go")
	name := "inventory-engine"
	if mutant {
		text, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		before := "func countText(f frequency) string {\n\tif f.Count == nil {\n\t\treturn \"unknown\""
		after := "func countText(f frequency) string {\n\tif f.Count == nil {\n\t\treturn \"0\""
		if strings.Count(string(text), before) != 1 {
			t.Fatal("unknown-frequency mutant has no unique target")
		}
		source = filepath.Join(directory, "mutant.go")
		if err := os.WriteFile(source, []byte(strings.Replace(string(text), before, after, 1)), 0644); err != nil {
			t.Fatal(err)
		}
		name += "-unknown-frequency-mutant"
	}
	binary := filepath.Join(directory, "engine.test")
	build := func(directory string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, virtualTest: testSource}})
		if err != nil {
			return err
		}
		path := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command, cancel := inventoryEngineTestCommand(t, "go", "test", "-c", "-overlay="+path, "-o", binary, virtual, virtualTest)
		defer cancel()
		command.Dir = cohere
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w\n%s", err, output)
		}
		return nil
	}
	started := time.Now()
	if err := build(directory); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(started)
	t.Logf("build %s private-overlay %.6fs", name, elapsed.Seconds())
	return binary, cohere, elapsed
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

func inventoryEngineUnion(t *testing.T, binary, directory string) [][]string {
	t.Helper()
	command, cancel := inventoryEngineTestCommand(t, binary, "-test.list=^Test")
	defer cancel()
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("enumerate engine: %v\n%s", err, output)
	}
	unsplit := strings.Fields(string(output))
	shards, err := inventoryEnginePartition(unsplit)
	if err != nil {
		t.Fatal(err)
	}
	if len(shards) != testInventoryEngineShards {
		t.Fatalf("enumerated %d shards, want %d", len(shards), testInventoryEngineShards)
	}
	union := []string{}
	seen := map[string]bool{}
	for index, cases := range shards {
		for _, name := range cases {
			if seen[name] {
				t.Fatalf("repeated case %s in shard union", name)
			}
			if inventoryEngineShard(name) != index {
				t.Fatalf("case %s assigned to wrong shard", name)
			}
			seen[name] = true
			union = append(union, name)
		}
	}
	slices.Sort(union)
	slices.Sort(unsplit)
	if !slices.Equal(union, unsplit) {
		t.Fatalf("shard union %v differs from live enumeration %v", union, unsplit)
	}
	t.Logf("union: %d unique cases equal the complete live enumeration across %d fixed shards", len(union), len(shards))
	return shards
}

// Adding or reordering cases must not move an existing case; empty and repeated
// corpora must fail rather than silently produce a passing set of shards.
func TestInventoryEngineShardAssignment(t *testing.T) {
	t.Parallel()
	original := []string{"TestExistingA", "TestExistingB", "TestExistingC"}
	before, err := inventoryEnginePartition(original)
	if err != nil {
		t.Fatal(err)
	}
	grown := append([]string{"TestAdded"}, original...)
	slices.Reverse(grown)
	after, err := inventoryEnginePartition(grown)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != testInventoryEngineShards || len(after) != testInventoryEngineShards {
		t.Fatal("growth changed shard count")
	}
	for index, cases := range before {
		for _, name := range cases {
			if !slices.Contains(after[index], name) {
				t.Fatalf("growth moved case %s out of shard-%03d", name, index)
			}
		}
	}
	total := 0
	for _, cases := range after {
		total += len(cases)
	}
	if total != len(grown) {
		t.Fatalf("growth lost cases: %d, want live total %d", total, len(grown))
	}
	for _, invalid := range [][]string{nil, {"TestRepeated", "TestRepeated"}} {
		if _, err := inventoryEnginePartition(invalid); err == nil {
			t.Fatalf("invalid corpus accepted: %v", invalid)
		}
	}
}

func inventoryEngineRun(t *testing.T, binary, directory, name string) ([]byte, error) {
	command, cancel := inventoryEngineTestCommand(t, binary, "-test.run=^"+regexp.QuoteMeta(name)+"$", "-test.count=1", "-test.v", "-test.timeout=75s")
	defer cancel()
	command.Dir = directory
	return command.CombinedOutput()
}

func inventoryEngineSelector(value string) (func(int) bool, error) {
	if value == "" {
		return func(int) bool { return true }, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n, got %q", value)
	}
	index, indexErr := strconv.Atoi(parts[0])
	count, countErr := strconv.Atoi(parts[1])
	if indexErr != nil || countErr != nil || count < 1 || index < 0 || index >= count {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return func(shard int) bool { return shard%count == index }, nil
}

// A real production mutant: unknown frequency becomes measured zero. Run every
// live case against it and require exactly its stable hash shard to reject it.
func TestInventoryEngineShardMutant(t *testing.T) {
	t.Parallel()
	binary, cohere, _ := inventoryEngineBuild(t, true)
	shards := inventoryEngineUnion(t, binary, cohere)
	caught := []string{}
	for index, cases := range shards {
		for _, name := range cases {
			output, err := inventoryEngineRun(t, binary, cohere, name)
			if err == nil {
				continue
			}
			if name != "TestUnknownFrequencyIsNotZero" || !strings.Contains(string(output), "unmeasured count printed as zero") {
				t.Fatalf("unexpected mutant failure in shard-%03d (%s): %v\n%s", index, name, err, output)
			}
			caught = append(caught, fmt.Sprintf("shard-%03d", index))
		}
	}
	expected := fmt.Sprintf("shard-%03d", inventoryEngineShard("TestUnknownFrequencyIsNotZero"))
	if !slices.Equal(caught, []string{expected}) {
		t.Fatalf("unknown-frequency mutant caught by %v, want exactly %s", caught, expected)
	}
	t.Logf("unknown-frequency mutant caught by exactly %s (TestUnknownFrequencyIsNotZero)", expected)
}

// A per-run private overlay build is shared until TestMain cleans it up after
// every top-level shard. It is a Go build: do not hand-list Product inputs.
var inventoryEngineScratch string
var inventoryEngineSharedBuild struct {
	once   sync.Once
	binary string
	cohere string
	shards [][]string
}

func TestMain(m *testing.M) {
	code := m.Run()
	if inventoryEngineScratch != "" {
		os.RemoveAll(inventoryEngineScratch)
	}
	os.Exit(code)
}

func inventoryEngineShared(t *testing.T) (string, string, [][]string) {
	t.Helper()
	shared := &inventoryEngineSharedBuild
	shared.once.Do(func() {
		shared.binary, shared.cohere, _ = inventoryEngineBuild(t, false)
		shared.shards = inventoryEngineUnion(t, shared.binary, shared.cohere)
	})
	if shared.binary == "" || len(shared.shards) != testInventoryEngineShards {
		t.Fatal("shared inventory build or union validation did not complete")
	}
	return shared.binary, shared.cohere, shared.shards
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
// Assignment hashes the repository-relative case file plus the test name.
func inventoryEngineTopLevelShard(t *testing.T, index int) {
	t.Helper()
	selector, err := inventoryEngineSelector(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	if !selector(index) {
		t.Skip("outside ADAMIC_TEST_SHARD selection")
	}
	binary, directory, shards := inventoryEngineShared(t)
	inventoryEngineDeadline(t, binary, directory, shards[index])
}

func TestInventoryEngine_000(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 0) }
func TestInventoryEngine_001(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 1) }
func TestInventoryEngine_002(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 2) }
func TestInventoryEngine_003(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 3) }
func TestInventoryEngine_004(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 4) }
func TestInventoryEngine_005(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 5) }
func TestInventoryEngine_006(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 6) }
func TestInventoryEngine_007(t *testing.T) { t.Parallel(); inventoryEngineTopLevelShard(t, 7) }

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

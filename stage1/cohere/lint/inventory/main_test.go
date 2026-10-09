package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testInventoryEngineShards = 8

// Keep headroom independent of the live case count. Test names are stable modes
// within this repository-relative file; enumeration order never enters the key.
const inventoryEngineCaseFile = "stage1/cohere/lint/inventory/testdata/engine_test.go"

// ADAMIC_TEST_SHARD=i/n runs shards whose index modulo n is i; unset runs all.
// The overlay engine is built once before parallel shards. Overlay builds remain
// private: internal/buildcache.GoBuild refuses overlays.
func TestInventoryEngine(t *testing.T) {
	started := time.Now()
	binary, cohere, buildTime := inventoryEngineBuild(t, false)
	shards := inventoryEngineUnion(t, binary, cohere)
	selector, err := inventoryEngineSelector(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	for index, cases := range shards {
		if !selector(index) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, name := range cases {
				output, err := inventoryEngineRun(binary, cohere, name)
				t.Logf("%s", output)
				if err != nil {
					t.Fatalf("case %s: %v", name, err)
				}
			}
		})
	}
	setup := time.Since(started)
	t.Logf("setup with builds %.6fs; without builds %.6fs", setup.Seconds(), (setup - buildTime).Seconds())
}

// Build from the same two complete files as the unsplit suite. All output goes
// into directory; the overlay and mutant never change the checkout or a product.
func inventoryEngineBuild(t *testing.T, mutant bool) (string, string, time.Duration) {
	t.Helper()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
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
		command := exec.Command("go", "test", "-c", "-overlay="+path, "-o", binary, virtual, virtualTest)
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
	command := exec.Command(binary, "-test.list=^Test")
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

func inventoryEngineRun(binary, directory, name string) ([]byte, error) {
	command := exec.Command(binary, "-test.run=^"+regexp.QuoteMeta(name)+"$", "-test.count=1", "-test.v", "-test.timeout=3h")
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
	binary, cohere, _ := inventoryEngineBuild(t, true)
	shards := inventoryEngineUnion(t, binary, cohere)
	caught := []string{}
	for index, cases := range shards {
		for _, name := range cases {
			output, err := inventoryEngineRun(binary, cohere, name)
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

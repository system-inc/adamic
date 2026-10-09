package main

import (
	"encoding/json"
	"fmt"
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

var inventoryEngineCases = []string{
	"TestTransitiveSiblingAndMethodDependencies",
	"TestHelperRankingCountsRulesOnce",
	"TestUnknownFrequencyIsNotZero",
	"TestRegisteredCorpusControl",
	"TestBranchEvidenceRequiresExecutableSelector",
	"TestAdamicCorpusUsesCompleteProgram",
	"TestFailedFamilyCannotBeMarkedPassed",
	"TestInventoryFileNameBytes",
}

// ADAMIC_TEST_SHARD=i/n runs shards whose index modulo n is i; unset runs all.
// The overlay engine is built once before parallel shards. Overlay builds remain
// private: internal/buildcache.GoBuild refuses overlays.
func TestInventoryEngine(t *testing.T) {
	started := time.Now()
	binary, cohere, buildTime := inventoryEngineBuild(t, false)
	inventoryEngineUnion(t, binary, cohere)
	selector, err := inventoryEngineSelector(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	for index, name := range inventoryEngineCases {
		if !selector(index) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			output, err := inventoryEngineRun(binary, cohere, name)
			t.Logf("%s", output)
			if err != nil {
				t.Fatalf("case %s: %v", name, err)
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

func inventoryEngineUnion(t *testing.T, binary, directory string) {
	t.Helper()
	command := exec.Command(binary, "-test.list=^Test")
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("enumerate engine: %v\n%s", err, output)
	}
	unsplit := strings.Fields(string(output))
	if len(unsplit) != testInventoryEngineShards || len(inventoryEngineCases) != testInventoryEngineShards {
		t.Fatalf("enumerated %d cases, planned %d, want %d shards", len(unsplit), len(inventoryEngineCases), testInventoryEngineShards)
	}
	seen := map[string]bool{}
	for _, name := range inventoryEngineCases {
		if seen[name] {
			t.Fatalf("repeated case %s", name)
		}
		seen[name] = true
	}
	union := slices.Clone(inventoryEngineCases)
	slices.Sort(union)
	slices.Sort(unsplit)
	if !slices.Equal(union, unsplit) {
		t.Fatalf("shard union %v differs from unsplit enumeration %v", union, unsplit)
	}
	t.Logf("union: %d unique cases equal the complete unsplit enumeration", len(union))
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
// original case against it and require exactly shard-002 to reject it.
func TestInventoryEngineShardMutant(t *testing.T) {
	binary, cohere, _ := inventoryEngineBuild(t, true)
	inventoryEngineUnion(t, binary, cohere)
	caught := []string{}
	for index, name := range inventoryEngineCases {
		output, err := inventoryEngineRun(binary, cohere, name)
		if err == nil {
			continue
		}
		if name != "TestUnknownFrequencyIsNotZero" || !strings.Contains(string(output), "unmeasured count printed as zero") {
			t.Fatalf("unexpected mutant failure in shard-%03d (%s): %v\n%s", index, name, err, output)
		}
		caught = append(caught, fmt.Sprintf("shard-%03d", index))
	}
	if !slices.Equal(caught, []string{"shard-002"}) {
		t.Fatalf("unknown-frequency mutant caught by %v, want exactly shard-002", caught)
	}
	t.Log("unknown-frequency mutant caught by exactly shard-002 (TestUnknownFrequencyIsNotZero)")
}

package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryTypeNodeKindAgrees(t *testing.T) {
	oracle := goOracle(t)
	expected := execute(t, "", oracle, "--type-kinds").output
	manifest := wholeManifest(t, wholeCases())
	trees := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	docs := wholeManifest(t, []string{`/** @type {...T} */ let x;`, `/** @type {T=} */ let x;`})
	want := execute(t, "", oracle, "--manifest", docs, "--doc-types")
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	got := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", docs, "--doc-types")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node doc types: %s", diff)
	}
	got = execute(t, "", buildPort(t, directory, true), "--manifest", docs, "--doc-types")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native doc types: %s", diff)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(trees)+string(want.output), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 2 {
			seen[fields[1]] = true
		}
	}
	kinds := strings.Fields(string(expected))
	for _, kind := range kinds {
		if !seen[kind] {
			t.Errorf("Go IsTypeNodeKind: missing coverage for %s", kind)
		}
	}
	t.Logf("Go IsTypeNodeKind inventory: %d kinds, each covered; %d identical doc-type bytes", len(kinds), len(want.output))
}

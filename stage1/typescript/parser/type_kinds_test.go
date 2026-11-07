package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryTypeNodeKindAgrees(test *testing.T) {
	test.Parallel()
	oracle := goOracle(test)
	expected := execute(test, "", oracle, "--type-kinds").output
	manifest := wholeManifest(test, wholeCases())
	trees := execute(test, "", oracle, "--manifest", manifest, "--whole").output
	docs := wholeManifest(test, []string{`/** @type {...T} */ let x;`, `/** @type {T=} */ let x;`})
	want := execute(test, "", oracle, "--manifest", docs, "--doc-types")
	directory, err := filepath.Abs(".")
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	got := execute(
		test,
		"",
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		docs,
		"--doc-types",
	)
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("Node doc types: %s", diff)
	}
	got = execute(test, "", buildPort(test, directory, true), "--manifest", docs, "--doc-types")
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("native doc types: %s", diff)
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
			test.Errorf("Go IsTypeNodeKind: missing coverage for %s", kind)
		}
	}
	test.Logf(
		"Go IsTypeNodeKind inventory: %d kinds, each covered; %d identical doc-type bytes",
		len(kinds),
		len(want.output),
	)
}

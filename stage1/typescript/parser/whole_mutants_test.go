package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testWholeMutantsShards = 3

// ADAMIC_TEST_SHARD=i/n selects shard indices congruent to i modulo n;
// unset runs all three top-level TestWholeMutants_NNN leaves. Each leaf
// fetches hash-addressed inputs, then retains both sides and sanitizers.
func wholeMutantsShard(t *testing.T, index int) {
	started := time.Now()
	if !wholeMutantShardSelected(t, index) {
		t.Skip("assigned to another ADAMIC_TEST_SHARD")
	}
	mutations := wholeMutantCases()
	if len(mutations) != testWholeMutantsShards {
		t.Fatalf("enumerated %d mutants, declared %d", len(mutations), testWholeMutantsShards)
	}
	if index < 0 || index >= len(mutations) {
		t.Fatalf("invalid mutant shard %d", index)
	}
	mutation := mutations[index]
	manifest := wholeManifest(t, []string{`for (const x of xs) f(x); import type {X} from "x"; type T = keyof X;`})
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	productStarted := time.Now()
	oracle := wholeMutantOracleProduct(t)
	binary := wholeMutantPortProduct(t, directory, true)
	mutant := copyPort(t, mutation.file, mutation.from, mutation.to)
	mutantBinary := wholeMutantPortProduct(t, mutant, true)
	products := time.Since(productStarted)
	wholeMutantSetup(t, started, products)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	for _, got := range []execution{wholeNode(t, directory, manifest, false), execute(t, "", binary, "--manifest", manifest, "--whole")} {
		if diff := difference(got.output, want); diff != "" {
			t.Fatal(diff)
		}
	}
	for _, side := range []struct {
		name string
		got  execution
	}{
		{"Node", wholeNode(t, mutant, manifest, false)},
		{"native", execute(t, "", mutantBinary, "--manifest", manifest, "--whole")},
	} {
		got := side.got.output
		// After the real sanitized run, emulate a survivor in exactly one case.
		planted := os.Getenv("ADAMIC_WHOLE_MUTANT_SURVIVOR") == mutation.name && side.name == "native"
		if planted {
			got = want
		}
		diff := difference(got, want)
		if diff == "" {
			if planted {
				t.Fatalf("%s planted mutant survived: %s", side.name, mutation.name)
			}
			t.Fatalf("%s mutant survived: %s", side.name, mutation.name)
		}
		t.Logf("%s %s caught: %s", side.name, mutation.name, strings.Split(diff, "\n")[0])
	}
}

func TestWholeCountCheckCatchesMutant(t *testing.T) {
	manifest := wholeManifest(t, []string{`const x = 1; type T = keyof X;`})
	oracle := goOracle(t)
	wantTree := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	wantCount := execute(t, "", oracle, "--manifest", manifest, "--whole", "--count").output
	from := "export function countTree(nodes: readonly ParseNode[], root: number): number {\n    const node = nodes[root] ?? panic('missing parse node');\n    let count = 1;"
	mutant := copyPort(t, "nodes.ts", from, strings.Replace(from, "let count = 1;", "let count = 0;", 1))
	binary := buildPort(t, mutant, true)
	for _, side := range []struct {
		name        string
		tree, count execution
	}{
		{"Node", wholeNode(t, mutant, manifest, false), wholeNode(t, mutant, manifest, true)},
		{"native", execute(t, "", binary, "--manifest", manifest, "--whole"), execute(t, "", binary, "--manifest", manifest, "--whole", "--count")},
	} {
		if diff := difference(side.tree.output, wantTree); diff != "" {
			t.Fatalf("%s counter mutant changed trees: %s", side.name, diff)
		}
		if string(side.count.output) == string(wantCount) {
			t.Fatalf("%s counter mutant survived", side.name)
		}
		t.Logf("%s: identical whole-tree bytes, mutant count %q caught against Go %q", side.name, side.count.output, wantCount)
	}
}

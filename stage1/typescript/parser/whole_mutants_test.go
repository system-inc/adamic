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
func wholeMutantsShard(test *testing.T, index int) {
	started := time.Now()
	if !wholeMutantShardSelected(test, index) {
		test.Skip("assigned to another ADAMIC_TEST_SHARD")
	}
	mutations := wholeMutantCases()
	if len(mutations) != testWholeMutantsShards {
		test.Fatalf("enumerated %d mutants, declared %d", len(mutations), testWholeMutantsShards)
	}
	if index < 0 || index >= len(mutations) {
		test.Fatalf("invalid mutant shard %d", index)
	}
	mutation := mutations[index]
	manifest := wholeManifest(test, []string{`for (const x of xs) f(x); import type {X} from "x"; type T = keyof X;`})
	directory, err := filepath.Abs(".")
	if err != nil {
		test.Fatal(err)
	}
	productStarted := time.Now()
	oracle := wholeMutantOracleProduct(test)
	binary := wholeMutantPortProduct(test, directory, true)
	mutant := copyPort(test, mutation.file, mutation.from, mutation.to)
	mutantBinary := wholeMutantPortProduct(test, mutant, true)
	products := time.Since(productStarted)
	wholeMutantSetup(test, started, products)
	want := execute(test, "", oracle, "--manifest", manifest, "--whole").output
	for _, got := range []execution{
		wholeNode(test, directory, manifest, false),
		execute(test, "", binary, "--manifest", manifest, "--whole"),
	} {
		if diff := difference(got.output, want); diff != "" {
			test.Fatal(diff)
		}
	}
	for _, side := range []struct {
		name string
		got  execution
	}{
		{"Node", wholeNode(test, mutant, manifest, false)},
		{"native", execute(test, "", mutantBinary, "--manifest", manifest, "--whole")},
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
				test.Fatalf("%s planted mutant survived: %s", side.name, mutation.name)
			}
			test.Fatalf("%s mutant survived: %s", side.name, mutation.name)
		}
		test.Logf("%s %s caught: %s", side.name, mutation.name, strings.Split(diff, "\n")[0])
	}
}

func TestWholeCountCheckCatchesMutant(test *testing.T) {
	test.Parallel()
	manifest := wholeManifest(test, []string{`const x = 1; type T = keyof X;`})
	oracle := goOracle(test)
	wantTree := execute(test, "", oracle, "--manifest", manifest, "--whole").output
	wantCount := execute(test, "", oracle, "--manifest", manifest, "--whole", "--count").output
	mutant := countTreeMutant(test)
	binary := buildPort(test, mutant, true)
	for _, side := range []struct {
		name        string
		tree, count execution
	}{
		{"Node", wholeNode(test, mutant, manifest, false), wholeNode(test, mutant, manifest, true)},
		{
			"native",
			execute(test, "", binary, "--manifest", manifest, "--whole"),
			execute(test, "", binary, "--manifest", manifest, "--whole", "--count"),
		},
	} {
		if diff := difference(side.tree.output, wantTree); diff != "" {
			test.Fatalf("%s counter mutant changed trees: %s", side.name, diff)
		}
		if string(side.count.output) == string(wantCount) {
			test.Fatalf("%s counter mutant survived", side.name)
		}
		test.Logf(
			"%s: identical whole-tree bytes, mutant count %q caught against Go %q",
			side.name,
			side.count.output,
			wantCount,
		)
	}
}

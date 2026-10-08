package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWholeMutants(t *testing.T) {
	t.Parallel()
	manifest := wholeManifest(t, []string{`for (const x of xs) f(x); import type {X} from "x"; type T = keyof X;`})
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []execution{wholeNode(t, directory, manifest, false), execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")} {
		if diff := difference(got.output, want); diff != "" {
			t.Fatal(diff)
		}
	}
	mutations := []struct{ name, file, from, to string }{
		{"for-of becomes for-in", "statements.ts", "of ? 'ForOfStatement' : 'ForInStatement'", "of ? 'ForInStatement' : 'ForInStatement'"},
		{"type-only import phase lost", "statements.ts", "this.parser.node(clause).semantic = phase;", "this.parser.node(clause).semantic = 'Unknown';"},
		{"keyof becomes readonly", "parser.ts", "this.node(left).operator = operator;", "this.node(left).operator = operator === 'KeyOfKeyword' ? 'ReadonlyKeyword' : operator;"},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutant := copyPort(t, mutation.file, mutation.from, mutation.to)
			for _, side := range []struct {
				name string
				got  execution
			}{{"Node", wholeNode(t, mutant, manifest, false)}, {"native", execute(t, "", buildPort(t, mutant, true), "--manifest", manifest, "--whole")}} {
				diff := difference(side.got.output, want)
				if diff == "" {
					t.Fatalf("%s mutant survived", side.name)
				}
				t.Logf("%s caught: %s", side.name, strings.Split(diff, "\n")[0])
			}
		})
	}
}

func TestWholeCountCheckCatchesMutant(t *testing.T) {
	t.Parallel()
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

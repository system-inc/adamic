package estree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func syntaxGrammar() []string {
	return []string{
		"type T=ReturnType<<U>(x:U)=>number>; const b=foo<<U>(x:U)=>number>(()=>1);",
		"export declare const f: import('m').Modifier<<T>(x:T)=>T>;",
		"type A={get foo(){return 0} set foo(v:any){}}; interface I {x:number=5;}",
		"type T={ [P in A]?: B; model:'hour'|'day'; }; type U={ [P in A]\nmodel: B; };",
		"interface I { f(): typeof a\n<T>():void; foo(): I\nis():boolean; }",
		"interface I { get p():asserts this is string; set p(x:asserts this is string); }",
		"/// <reference path='a.ts' />\nx;",
		"x;\n/// <reference path='broken.ts />",
		"/* /// <reference path='broken.ts /> */ x; const s=\"/// <reference path='broken.ts />\";",
	}
}

const testSyntaxGrammarShards = 9

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
func TestSyntaxGrammar(t *testing.T) {
	estreeAgreementShards(t, testSyntaxGrammarShards, syntaxGrammar())
}
func TestSyntaxGrammarShardFailure(t *testing.T) {
	sources := syntaxGrammar()
	estreeShardFailure(t, testSyntaxGrammarShards, len(sources), estreeSingles(len(sources)), "agreement", 4)
}

const testSyntaxMutantsShards = 3

func syntaxMutantGroups() ([][]int, int) {
	groups := make([][]int, 3)
	id := 0
	for i, count := range []int{len(syntaxGrammar()), 1, 1} {
		for range count {
			groups[i] = append(groups[i], id)
			id++
		}
	}
	return groups, id
}

// ADAMIC_TEST_SHARD=i/n runs the shards whose index modulo n is i; unset runs all.
func TestSyntaxMutants(t *testing.T) {
	estreeAccounting(t)
	started := time.Now()
	groups, cases := syntaxMutantGroups()
	selected := estreeShardPlan(t, testSyntaxMutantsShards, cases, groups)
	oracle := estreeTimedOracle(t)
	type mutant struct {
		name, file, from, to string
		sources              []string
		main, binary         string
	}
	mutants := []mutant{
		{name: "mapped-constraint", file: "convert.ts", from: "this.set(result, 'constraint', this.converted(this.child(parameter, 1)));", to: "this.set(result, 'constraint', absent());", sources: syntaxGrammar()},
		{name: "erasure-precedence", file: "sourceBinary.ts", from: "if(nextRank > lastRank ||", to: "if(false && nextRank > lastRank ||", sources: []string{"1+1 as number *2;"}},
		{name: "reference-pragma", file: "pipeline.ts", from: "if(reference !== '')", to: "if(false)", sources: []string{"/// <reference path='missingquote.ts />\nx;"}},
	}
	if len(mutants) != len(groups) {
		t.Fatal("mutant enumeration differs from shard plan")
	}
	for i, m := range mutants {
		if len(m.sources) != len(groups[i]) {
			t.Fatalf("shard-%03d has %d inputs, plan has %d IDs", i, len(m.sources), len(groups[i]))
		}
	}
	for i := range mutants {
		if !selected[i] {
			continue
		}
		m := &mutants[i]
		m.main = mutantPort(t, m.file, m.from, m.to)
		m.binary, _ = estreeTimedBuild(t, m.main, true)
	}
	t.Logf("setup including builds: %.3fs; union: %d cases", time.Since(started).Seconds(), cases)
	for i, m := range mutants {
		if !selected[i] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			list := manifest(t, m.sources)
			if i == 0 {
				want := estreeOracleOutput(t, oracle, "--manifest", list)
				for name, got := range map[string][]byte{"Node": onNode(t, m.main, "--manifest", list), "native": execute(t, "", m.binary, "--manifest", list)} {
					if err := estreeMutantVerdict(want, got); err != nil {
						t.Fatalf("%s %s: %v", m.name, name, err)
					}
				}
			} else {
				statuses := string(estreeOracleOutput(t, oracle, "--audit", list, t.TempDir()))
				if !strings.Contains(statuses, `"status":"error"`) {
					t.Fatal(statuses)
				}
				for name, got := range map[string][]byte{"Node": onNode(t, m.main, "--manifest", list), "native": execute(t, "", m.binary, "--manifest", list)} {
					if !strings.Contains(string(got), "0 Program ") {
						t.Fatal(name + " control did not accept")
					}
					t.Log(m.name + " " + name + ": disabled check accepts Go-refused input; acceptance oracle catches it")
				}
			}
		})
	}
}

const testSyntaxRefusalsShards = 3

func syntaxRefusalCases() []string {
	return []string{"1+1 as number *2;", "/// <reference path='missingquote.ts />\nx;", "/// <reference types='m' resolution-mode='invalid' />\nx;"}
}

// ADAMIC_TEST_SHARD=i/n selects shard indices modulo n; unset runs all.
func TestSyntaxRefusals(t *testing.T) {
	estreeAccounting(t)
	started := time.Now()
	sources := syntaxRefusalCases()
	selected := estreeShardPlan(t, testSyntaxRefusalsShards, len(sources), estreeSingles(len(sources)))
	oracle := estreeTimedOracle(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := estreeTimedBuild(t, main, true)
	t.Logf("setup including builds: %.3fs", time.Since(started).Seconds())
	for i, source := range sources {
		if !selected[i] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			list := manifest(t, []string{source})
			statuses := string(estreeOracleOutput(t, oracle, "--audit", list, t.TempDir()))
			if strings.Count(statuses, `"status":"error"`) != 1 {
				t.Fatal(statuses)
			}
			paths, err := os.ReadFile(list)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range strings.Fields(string(paths)) {
				for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
					estreeRefused(t, argv, "ESTree parser")
				}
			}
		})
	}
}
func TestSyntaxRefusalsShardFailure(t *testing.T) {
	sources := syntaxRefusalCases()
	estreeShardFailure(t, testSyntaxRefusalsShards, len(sources), estreeSingles(len(sources)), "refusal", 1)
}

func TestSyntaxLibraries(t *testing.T) {
	if os.Getenv("ADAMIC_ESTREE_LIBRARY") == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	samples := syntaxGrammar()
	estreeCheckOriginalLibraries(t, []string{samples[0], samples[1], samples[4], samples[6]}, 0)
}
func TestTypeMemberLibraryGap(t *testing.T) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	source := "interface I {x:number=5;}"
	list := manifest(t, []string{source})
	estreeOracleOutput(t, estreeOracleProduct(t), "--manifest", list)
	code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);try{lib.parse(process.argv[2],{warnOnUnsupportedTypeScriptVersion:false});console.log('accepted')}catch(e){console.log('refused')}`
	if got := string(execute(t, "", "node", "--input-type=module", "-e", code, library, source)); got != "refused\n" {
		t.Fatal(got)
	}
	t.Log("Go accepts an interface property initializer, pinned typescript-estree refuses it; the port follows Go")
}

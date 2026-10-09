package estree

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func recoveredGrammar() []string {
	return []string{
		"class C { readonly!: number; static #x=1; 1n=2; }",
		"class C { private public x: number; protected public m(){} }",
		"class C { constructor(static x: number, readonly y: number){} }",
		"type X = { static x: number; new: string; #x: number; [a,b]: number; };",
		"({ x=1, y?, z?:2 });",
		"f<>(); new C<>(); class C<> {} function f<>(){}",
		"class C { constructor<>(){} }",
		"class C implements string, number {}",
		"class C extends A extends B implements X implements Y {}",
		"export {}; await(x); await 1; function f(){await(x);} namespace N {await(x);}",
		"export default async(x); export default async () => 1; export default async;",
		"declare\nmodule\n'x'\n{}",
		"let obj = /** @satisfies {{ f(s:string):void }} */ ({f(s){}});",
		"const expected = /** @type {{name:string}} */ (null);",
		"a ? (b) : c => (d) : e => f;",
		"class C { static { await(x); class D { x=await(x); } } }",
	}
}
func TestRecoveredGrammar(t *testing.T) {
	list := manifest(t, recoveredGrammar())
	want := recoveryAnswer(t, goOracle(t), list, "--manifest")
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatal(name + ": " + diff)
		}
	}
	t.Logf("%d recovered grammar cases, %d identical bytes in all three port builds", len(recoveredGrammar()), len(want))
}

const testRecoveryMutantsShards = 3

var recoveryMutations = []struct{ name, file, from, to string }{
	{"first-accessibility", "modifiers.ts", "return stringValue(kind.slice(0, -7).toLowerCase());", "return stringValue('public');"},
	{"empty-type-list-range", "typeLists.ts", "arena.node(result).set('params', listValue([]));", "arena.node(result).end -= 1; arena.node(result).set('params', listValue([]));"},
	{"module-await", "pipeline.ts", "&& externalModule(parser.nodes, root)", "&& false && externalModule(parser.nodes, root)"},
}

func recoveryMutantCases() []recoveryCase {
	var cases []recoveryCase
	// Interleave mutants so partition i%3 assigns one complete corpus per mutant.
	for i, source := range recoveredGrammar() {
		for _, m := range recoveryMutations {
			cases = append(cases, recoveryCase{fmt.Sprintf("%s/%03d", m.name, i), source})
		}
	}
	return cases
}

// ADAMIC_TEST_SHARD=i/n runs shards whose index modulo n is i; unset runs all.
// Each mutant retains the entire grammar corpus on both Node and sanitized native.
func TestRecoveryMutants(t *testing.T) {
	// Compatibility enumeration only; execution lives in the top-level shards.
	checkRecoveryMutantUnion(t)
}

func TestRecoveryMutantsShardSurvivor(t *testing.T) {
	proveRecoveryShard(t, recoveryMutantCases(), testRecoveryMutantsShards, true)
}
func TestRecoveryLibraryGaps(t *testing.T) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	for _, source := range []string{"f<>();", "class C<> {}"} {
		list := manifest(t, []string{source})
		recoveryAnswer(t, goOracle(t), list, "--manifest")
		code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);try{lib.parse(process.argv[2],{warnOnUnsupportedTypeScriptVersion:false});console.log('accepted')}catch(e){console.log('refused')}`
		if got := string(execute(t, "", "node", "--input-type=module", "-e", code, library, source)); got != "refused\n" {
			t.Fatalf("library gap changed for %q: %s", source, got)
		}
		t.Logf("Go accepts %q, pinned typescript-estree refuses", source)
	}
	checkOriginalLibraries(t, []string{"class C implements string {}", "export {}; await(x);", "export default async(x);", "let obj = /** @satisfies {{ f(s:string):void }} */ ({f(s){}});"}, 0)
}

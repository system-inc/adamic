package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestSyntaxGrammar(test *testing.T) {
	list := manifest(test, syntaxGrammar())
	want := execute(test, "", goOracle(test), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, main, "--manifest", list),
		"native":  execute(test, "", binary, "--manifest", list),
		"emitted": onNode(test, script, "--manifest", list),
	} {
		if diff := firstDifference(want, got); diff != "" {
			test.Fatal(name + ": " + diff)
		}
	}
	test.Logf("%d syntax cases, %d identical canonical bytes", len(syntaxGrammar()), len(want))
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestSyntaxRefusals(test *testing.T) {
	sources := []string{
		"1+1 as number *2;",
		"/// <reference path='missingquote.ts />\nx;",
		"/// <reference types='m' resolution-mode='invalid' />\nx;",
	}
	list := manifest(test, sources)
	statuses := string(execute(test, "", goOracle(test), "--audit", list, test.TempDir()))
	if strings.Count(statuses, `"status":"error"`) != len(sources) {
		test.Fatal(statuses)
	}
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	paths, readError := os.ReadFile(list)
	if readError != nil {
		test.Fatal(readError)
	}
	for _, path := range strings.Fields(string(paths)) {
		for _, argv := range [][]string{
			{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(test), "oracle/node.mjs"), main, path},
			{binary, path},
			{
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(root(test), "oracle/node.mjs"),
				script,
				path,
			},
		} {
			refusedBeforeDeadline(test, argv, "ESTree parser")
		}
	}
	test.Log("three Go refusals explicitly refused with empty stdout on all builds")
}
func TestSyntaxLibraries(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ESTREE_LIBRARY") == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	samples := syntaxGrammar()
	checkOriginalLibraries(test, []string{samples[0], samples[1], samples[4], samples[6]}, 0)
}
func TestTypeMemberLibraryGap(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	source := "interface I {x:number=5;}"
	list := manifest(test, []string{source})
	execute(test, "", goOracle(test), "--manifest", list)
	code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);try{lib.parse(process.argv[2],{warnOnUnsupportedTypeScriptVersion:false});console.log('accepted')}catch(e){console.log('refused')}`
	if got := string(
		execute(test, "", "node", "--input-type=module", "-e", code, library, source),
	); got != "refused\n" {
		test.Fatal(got)
	}
	test.Log("Go accepts an interface property initializer, pinned typescript-estree refuses it; the port follows Go")
}

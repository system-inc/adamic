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
func TestSyntaxGrammar(t *testing.T) {
	t.Parallel()
	list := manifest(t, syntaxGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
	t.Logf("%d syntax cases, %d identical canonical bytes", len(syntaxGrammar()), len(want))
}
func TestSyntaxMutants(t *testing.T) {
	t.Parallel()
	list := manifest(t, syntaxGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	t.Run("mapped-constraint", func(t *testing.T) {
		t.Parallel()
		main := mutantPort(t, "convert.ts", "this.set(result, 'constraint', this.converted(this.child(parameter, 1)));", "this.set(result, 'constraint', absent());")
		checkPort(t, main, []string{"--manifest", list}, want, true, false)
	})
	for _, item := range []struct{ name, file, from, to, source string }{
		{"erasure-precedence", "sourceBinary.ts", "if(nextRank > lastRank ||", "if(false && nextRank > lastRank ||", "1+1 as number *2;"},
		{"reference-pragma", "pipeline.ts", "if(reference !== '')", "if(false)", "/// <reference path='missingquote.ts />\nx;"},
	} {
		t.Run(item.name, func(t *testing.T) {
			t.Parallel()
			list := manifest(t, []string{item.source})
			statuses := string(execute(t, "", goOracle(t), "--audit", list, t.TempDir()))
			if !strings.Contains(statuses, `"status":"error"`) {
				t.Fatal(statuses)
			}
			main := mutantPort(t, item.file, item.from, item.to)
			checkAcceptanceControl(t, main, list, 1)
		})
	}
}
func TestSyntaxRefusals(t *testing.T) {
	t.Parallel()
	sources := []string{"1+1 as number *2;", "/// <reference path='missingquote.ts />\nx;", "/// <reference types='m' resolution-mode='invalid' />\nx;"}
	list := manifest(t, sources)
	statuses := string(execute(t, "", goOracle(t), "--audit", list, t.TempDir()))
	if strings.Count(statuses, `"status":"error"`) != len(sources) {
		t.Fatal(statuses)
	}
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	paths, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range strings.Fields(string(paths)) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			checkRefusalModes(t, main, binary, script, path, "ESTree parser")
		})
	}
	t.Log("three Go refusals explicitly refused with empty stdout on all builds")
}
func TestSyntaxLibraries(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ESTREE_LIBRARY") == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	samples := syntaxGrammar()
	checkOriginalLibraries(t, []string{samples[0], samples[1], samples[4], samples[6]}, 0)
}
func TestTypeMemberLibraryGap(t *testing.T) {
	t.Parallel()
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	source := "interface I {x:number=5;}"
	list := manifest(t, []string{source})
	execute(t, "", goOracle(t), "--manifest", list)
	code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);try{lib.parse(process.argv[2],{warnOnUnsupportedTypeScriptVersion:false});console.log('accepted')}catch(e){console.log('refused')}`
	if got := string(execute(t, "", "node", "--input-type=module", "-e", code, library, source)); got != "refused\n" {
		t.Fatal(got)
	}
	t.Log("Go accepts an interface property initializer, pinned typescript-estree refuses it; the port follows Go")
}

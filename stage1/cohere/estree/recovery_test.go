package estree

import (
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"path/filepath"
	"testing"
)

func recoveredGrammar() []string { return oracles.RecoveredGrammar() }

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveredGrammar(test *testing.T) {
	list := manifest(test, recoveredGrammar())
	want := recoveryAnswer(test, goOracle(test), list, "--manifest")
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(test, main, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, main, "--manifest", list),
		"native":  execute(test, "", binary, "--manifest", list),
		"emitted": onNode(test, script, "--manifest", list),
	} {
		if diff := firstDifference(want, got); diff != "" {
			test.Fatal(name + ": " + diff)
		}
	}
	test.Logf(
		"%d recovered grammar cases, %d identical bytes in all three port builds",
		len(recoveredGrammar()),
		len(want),
	)
}

// recoveryLibraryGaps are sources Go accepts and the pinned typescript-estree refuses.
func recoveryLibraryGaps() []string { return oracles.RecoveryLibraryGaps() }

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveryLibraryGaps(test *testing.T) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	for _, source := range recoveryLibraryGaps() {
		list := manifest(test, []string{source})
		recoveryAnswer(test, goOracle(test), list, "--manifest")
		code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);try{lib.parse(process.argv[2],{warnOnUnsupportedTypeScriptVersion:false});console.log('accepted')}catch(e){console.log('refused')}`
		if got := string(
			execute(test, "", "node", "--input-type=module", "-e", code, library, source),
		); got != "refused\n" {
			test.Fatalf("library gap changed for %q: %s", source, got)
		}
		test.Logf("Go accepts %q, pinned typescript-estree refuses", source)
	}
	checkOriginalLibraries(test, []string{
		"class C implements string {}",
		"export {}; await(x);",
		"export default async(x);",
		"let obj = /** @satisfies {{ f(s:string):void }} */ ({f(s){}});",
	}, 0)
}

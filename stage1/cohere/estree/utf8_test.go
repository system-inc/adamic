package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cookedSurrogates() []string {
	return []string{
		"'\\ud800a\\udc00';",
		"'\\ud800';",
		"'\\udfff';",
		"'\\ud800\\udc00';",
		"'\\ud800\\ud800';",
		"`a\\ud800b`;",
		"tag`a\\ud800b`;",
		"const key = {'\\ud800': 1};",
	}
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestCookedSurrogates(test *testing.T) {
	list := manifest(test, cookedSurrogates())
	want := execute(test, "", goOracle(test), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, main, "--manifest", list),
		"native":  execute(test, "", binary, "--manifest", list),
		"emitted": onNode(test, script, "--manifest", list),
	} {
		if d := firstDifference(want, got); d != "" {
			test.Fatal(name + ": " + d)
		}
	}
	test.Logf("%d cooked-surrogate files, %d bytes match Go", len(cookedSurrogates()), len(want))
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestCookedSurrogateMutant(test *testing.T) {
	list := manifest(test, cookedSurrogates())
	want := execute(test, "", goOracle(test), "--manifest", list)
	path := estreeCookedSurrogateMutantPath(test)
	binary, _ := build(test, path, true)
	for name, got := range map[string][]byte{
		"Node":   onNode(test, path, "--manifest", list),
		"native": execute(test, "", binary, "--manifest", list),
	} {
		if d := firstDifference(want, got); d == "" {
			test.Fatal(name + " mutant survived")
		} else {
			test.Log(name + ": " + d)
		}
	}
}
func TestCookedSurrogateLibraryGap(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	list := manifest(test, []string{cookedSurrogates()[0]})
	goAnswer := string(execute(test, "", goOracle(test), "--manifest", list))
	if !strings.Contains(goAnswer, `.value string \ufffd\ufffd\ufffda\ufffd\ufffd\ufffd`) {
		test.Fatal("Go WTF-8 gap changed")
	}
	code := `import {pathToFileURL} from 'node:url';const lib=await import(pathToFileURL(process.argv[1]+'/node_modules/@typescript-eslint/typescript-estree/dist/index.js').href);const ast=lib.parse("'\\ud800a\\udc00';",{warnOnUnsupportedTypeScriptVersion:false});console.log(JSON.stringify(ast.body[0].expression.value));`
	got := string(execute(test, "", "node", "--input-type=module", "-e", code, library))
	if got != `"\ud800a\udc00"`+"\n" {
		test.Fatalf("library surrogate gap changed: %q", got)
	}
	test.Log("Go serializes each unpaired WTF-8 surrogate as three U+FFFD; original keeps UTF-16 surrogates")
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestLossyInputRefusal(test *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	for _, body := range [][]byte{{47, 47, 240, 144, 128, 10, 120, 59}, {47, 47, 239, 191, 189, 10, 120, 59}} {
		path := filepath.Join(test.TempDir(), "input.ts")
		os.WriteFile(path, body, 0644)
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
			refusedBeforeDeadline(test, argv, "cannot recover original UTF-8 bytes")
		}
	}
	test.Log("equal-size malformed and valid replacement inputs both explicitly refuse; raw-byte API remains required")
}

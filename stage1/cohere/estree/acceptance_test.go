package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func acceptanceGrammar() []string {
	return []string{
		"public abstract class C {}",
		"({public x:1, async y:2, readonly z:3});",
		"function* f(){(yield);[yield];f(yield);}",
		"class implements {}",
		"try {} catch(e=1){}",
		"function f(){throw\nx;}",
		"tag?.`hello`;",
		"class C extends (A) implements (B) {}",
		"type T=typeof obj.#x;",
		"interface I extends A extends B {}",
		"for(using of=0;;){}",
	}
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestAcceptanceGrammar(test *testing.T) {
	list := manifest(test, acceptanceGrammar())
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
	test.Logf("%d acceptance grammar cases, %d identical canonical bytes", len(acceptanceGrammar()), len(want))
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestAcceptanceDiagnostics(test *testing.T) {
	sources := []string{
		"++await 42;",
		"++delete foo.bar;",
		"--ANY1--;",
		"type T = A | () => B;",
		"new obj?.member();",
		"import { 'a' } from 'm';",
		"import A from 'm' assert {type:'json'};",
		"super<T>();",
	}
	list := manifest(test, sources)
	statuses := string(execute(test, "", goOracle(test), "--audit", list, test.TempDir()))
	if strings.Count(statuses, `"status":"error"`) != len(sources) {
		test.Fatal(statuses)
	}
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	paths, err := os.ReadFile(list)
	if err != nil {
		test.Fatal(err)
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
	test.Logf("%d Go refusals explicitly refused before deadline on all builds", len(sources))
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestAcceptanceDiagnosticControl(test *testing.T) {
	list := manifest(test, []string{"++await 42;"})
	statuses := string(execute(test, "", goOracle(test), "--audit", list, test.TempDir()))
	if !strings.Contains(statuses, `"status":"error"`) {
		test.Fatal(statuses)
	}
	main := estreeDiagnosticControlPath(test)
	binary, _ := build(test, main, true)
	for name, got := range map[string][]byte{
		"Node":   onNode(test, main, "--manifest", list),
		"native": execute(test, "", binary, "--manifest", list),
	} {
		if !strings.Contains(string(got), "0 Program ") {
			test.Fatal(name + " control did not accept")
		}
		test.Log(name + ": disabled syntax validation accepts Go-refused update operand; acceptance check catches it")
	}
}

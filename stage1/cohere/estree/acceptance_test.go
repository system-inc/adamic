package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func acceptanceGrammar() []string {
	return []string{
		"public abstract class C {}", "({public x:1, async y:2, readonly z:3});", "function* f(){(yield);[yield];f(yield);}", "class implements {}", "try {} catch(e=1){}", "function f(){throw\nx;}", "tag?.`hello`;", "class C extends (A) implements (B) {}", "type T=typeof obj.#x;", "interface I extends A extends B {}", "for(using of=0;;){}",
	}
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestAcceptanceGrammar(t *testing.T) {
	list := manifest(t, acceptanceGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatal(name + ": " + diff)
		}
	}
	t.Logf("%d acceptance grammar cases, %d identical canonical bytes", len(acceptanceGrammar()), len(want))
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestAcceptanceDiagnostics(t *testing.T) {
	sources := []string{"++await 42;", "++delete foo.bar;", "--ANY1--;", "type T = A | () => B;", "new obj?.member();", "import { 'a' } from 'm';", "import A from 'm' assert {type:'json'};", "super<T>();"}
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
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree parser")
		}
	}
	t.Logf("%d Go refusals explicitly refused before deadline on all builds", len(sources))
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestAcceptanceDiagnosticControl(t *testing.T) {
	list := manifest(t, []string{"++await 42;"})
	statuses := string(execute(t, "", goOracle(t), "--audit", list, t.TempDir()))
	if !strings.Contains(statuses, `"status":"error"`) {
		t.Fatal(statuses)
	}
	main := mutantPort(t, "pipeline.ts", "if(syntax !== '')", "if(false)")
	binary, _ := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
		if !strings.Contains(string(got), "0 Program ") {
			t.Fatal(name + " control did not accept")
		}
		t.Log(name + ": disabled syntax validation accepts Go-refused update operand; acceptance check catches it")
	}
}

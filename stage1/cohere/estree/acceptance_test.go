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
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
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
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			checkRefusalModes(t, main, binary, script, path, "ESTree parser")
		})
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
	checkAcceptanceControl(t, main, list, 1)
}

// Not parallel: acceptance mutant products compile native runtime cache entries.
func TestAcceptanceMutants(t *testing.T) {
	for shard := 0; shard < testAcceptanceMutantsShards; shard++ {
		runAcceptanceMutantShard(t, shard)
	}
}

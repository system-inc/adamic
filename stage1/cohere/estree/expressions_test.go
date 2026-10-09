package estree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func recoveredExpressions() []string {
	return []string{"const C = @d class {};", "const C = @d @e(1) class Named extends Base { static x=1; };", "(@d class { m(){return 1;} });", "function f(){ await 'x'; await 1; await true; await this; }", "function f(){ yield 'x'; yield 1; yield true; }", "await\nx; yield\nx;", "await(x); yield(x);"}
}

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveredExpressions(t *testing.T) {
	list := manifest(t, recoveredExpressions())
	want := recoveryAnswer(t, goOracle(t), list, "--manifest")
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if d := firstDifference(want, got); d != "" {
			t.Fatal(name + ": " + d)
		}
	}
	t.Logf("%d decorated class / await / yield cases, %d bytes match Go", len(recoveredExpressions()), len(want))
}

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveredExpressionMutant(t *testing.T) {
	list := manifest(t, recoveredExpressions())
	want := recoveryAnswer(t, goOracle(t), list, "--manifest")
	path := mutantPort(t, "sourceLookahead.ts", "(scanner.flags & 1) === 0 &&\n        (token", "(scanner.flags & 1) >= 0 &&\n        (token")
	binary, _ := recoveryBuild(t, path, true)
	for name, got := range map[string][]byte{"Node": onNode(t, path, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
		if d := firstDifference(want, got); d == "" {
			t.Fatal(name + " mutant survived")
		} else {
			t.Log(name + ": " + d)
		}
	}
}

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestUnattachedDecorator(t *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(t, main, true)
	for _, body := range []string{"@dec\nawait 1", "@dec\nx"} {
		path := filepath.Join(t.TempDir(), "input.ts")
		os.WriteFile(path, []byte(body), 0644)
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree unattached decorator")
		}
	}
}

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestUnattachedDecoratorControl(t *testing.T) {
	list := manifest(t, []string{"@dec\nawait 1", "@dec\nx"})
	statuses := string(recoveryAnswer(t, goOracle(t), list, "--audit"))
	if strings.Count(statuses, `"status":"error"`) != 2 {
		t.Fatalf("Go decorator refusals changed: %s", statuses)
	}
	main := mutantPort(t, "pipeline.ts", `parser.nodes[id]?.kind === 'Decorator'`, `parser.nodes[id]?.kind === 'UnusedDecoratorControl'`)
	binary, _ := recoveryBuild(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
		if strings.Count(string(got), "0 Program ") != 2 {
			t.Fatal(name + " decorator control did not finish with two incorrect acceptances")
		}
		t.Log(name + ": disabled orphan guard accepts both Go-refused files; acceptance check catches it")
	}
}

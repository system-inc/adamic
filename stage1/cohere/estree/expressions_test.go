package estree

import (
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"path/filepath"
	"testing"
)

func recoveredExpressions() []string { return oracles.RecoveredExpressions() }

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveredExpressions(test *testing.T) {
	list := manifest(test, recoveredExpressions())
	want := recoveryAnswer(test, goOracle(test), list, "--manifest")
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(test, main, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, main, "--manifest", list),
		"native":  execute(test, "", binary, "--manifest", list),
		"emitted": onNode(test, script, "--manifest", list),
	} {
		if d := firstDifference(want, got); d != "" {
			test.Fatal(name + ": " + d)
		}
	}
	test.Logf("%d decorated class / await / yield cases, %d bytes match Go", len(recoveredExpressions()), len(want))
}

// recoveredExpressionMutantPath is the port with sourceLookahead's flag test widened, so recovery admits what Go
// refuses.
func recoveredExpressionMutantPath(test *testing.T) string {
	test.Helper()
	return mutantPort(
		test,
		"sourceLookahead.ts",
		"(scanner.flags & 1) === 0 &&\n        (token",
		"(scanner.flags & 1) >= 0 &&\n        (token",
	)
}

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestRecoveredExpressionMutant(test *testing.T) {
	list := manifest(test, recoveredExpressions())
	want := recoveryAnswer(test, goOracle(test), list, "--manifest")
	path := recoveredExpressionMutantPath(test)
	binary, _ := recoveryBuild(test, path, true)
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

// Not parallel: recovery helpers write the shared cache directory adamic-build
func TestUnattachedDecorator(test *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := recoveryBuild(test, main, true)
	for _, body := range []string{"@dec\nawait 1", "@dec\nx"} {
		path := filepath.Join(test.TempDir(), "input.ts")
		os.WriteFile(path, []byte(body), 0644)
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
			refusedBeforeDeadline(test, argv, "ESTree unattached decorator")
		}
	}
}

package lane

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The two Python suites phase:stage3 ran beside the upstream suite are Go tests of their own, so Planner places each
// like any other test and a red names it: apply's tests (stage3/test_apply.py, whose ordinary-apply case adapts
// 00-setup alone) and the lane's checker and table tests. The lane's normalize-api.cjs reads stage3/api's install,
// which a tree holds, so neither needs the phase's npm ci. The lane's other suites, which only hand-run prove_*.py
// scripts ever ran, run beside them (#mpesmv7): a suite no gate runs can't fail.

func TestApplyPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_apply")
}

func TestLanePython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_check", "test_table", "test_artifact", "test_case_observer", "test_measure_plan",
		"test_profile", "test_shards", "test_worker_environment")
}

// stage3's other suites run from their own directories, which their imports name (caches, verdict's run and census,
// the scanner driver's runner), so each is a test of its own here (#mpesmv7).

func TestCachesPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_caches")
}

func TestVerdictPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/verdict", "test_verdict")
}

func TestScannerScratchRunnerPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/drivers/scanner", "test_scratch_runner")
}

// any-returns' known-answer fixture is a script, not a unittest suite: it inspects a three-callee fixture with stock
// TypeScript, whose typescript and @types/node are stage3/api's install, and asserts its own lib-to-tsc mutant is
// caught. Run without --evidence it writes nothing into the tree (#mpesmv7).
func TestAnyReturnsFixture(t *testing.T) {
	t.Parallel()
	runPythonCommand(t, "stage3/census/any-returns", []string{"NODE_PATH=" + filepath.Join(repositoryRoot(t), "stage3/api/node_modules")}, "test_fixture.py")
}

// runPython runs modules' unittest suites from directory, with a stage3 cache of the test's own, and fails with
// unittest's output when any of them fails.
func runPython(t *testing.T, directory string, modules ...string) {
	t.Helper()
	runPythonCommand(t, directory, nil, append([]string{"-m", "unittest"}, modules...)...)
}

// runPythonCommand runs python3 with arguments from directory, environment added to the toolchain's, and fails with
// its output when it fails.
func runPythonCommand(t *testing.T, directory string, environment []string, arguments ...string) {
	t.Helper()
	root := repositoryRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	command := boundedCommand(ctx, filepath.Join(root, directory), "python3", append([]string{"-E", "-S", "-B"}, arguments...)...)
	// -B is this python's; the suites' own python children read PYTHONDONTWRITEBYTECODE, so none writes a
	// __pycache__ into the tree (stage3/oracle's, on Cloud, before).
	command.Env = append(command.Env, "STAGE3_CACHE="+t.TempDir(), "PYTHONDONTWRITEBYTECODE=1")
	command.Env = append(command.Env, environment...)
	// Ordinary apply clones the pinned TypeScript from the gate input, which the toolchain's environment doesn't carry.
	if source, ok := os.LookupEnv("ADAMIC_TYPESCRIPT_SOURCE"); ok {
		command.Env = append(command.Env, "ADAMIC_TYPESCRIPT_SOURCE="+source)
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s in %s: %v\n%s", command, directory, err, output)
	}
}

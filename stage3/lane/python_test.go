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

// runPython runs modules' unittest suites from directory, with a stage3 cache of the test's own, and fails with
// unittest's output when any of them fails.
func runPython(t *testing.T, directory string, modules ...string) {
	t.Helper()
	root := repositoryRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	command := boundedCommand(ctx, filepath.Join(root, directory), "python3", append([]string{"-E", "-S", "-B", "-m", "unittest"}, modules...)...)
	// -B is this python's; the suites' own python children read PYTHONDONTWRITEBYTECODE, so none writes a
	// __pycache__ into the tree (stage3/oracle's, on Cloud, before).
	command.Env = append(command.Env, "STAGE3_CACHE="+t.TempDir(), "PYTHONDONTWRITEBYTECODE=1")
	// Ordinary apply clones the pinned TypeScript from the gate input, which the toolchain's environment doesn't carry.
	if source, ok := os.LookupEnv("ADAMIC_TYPESCRIPT_SOURCE"); ok {
		command.Env = append(command.Env, "ADAMIC_TYPESCRIPT_SOURCE="+source)
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s in %s: %v\n%s", command, directory, err, output)
	}
}

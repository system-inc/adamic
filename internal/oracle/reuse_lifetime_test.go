package oracle

import (
	"path/filepath"
	"testing"
)

// Each witness runs independently, including the release build and the leak check.
func checkRuntimeWitness(test *testing.T, fixture string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	got, leaked, comparable := nativelyCheckingLeaks(test, program, want.exitCode == 0)
	if leaked != "" {
		test.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	for _, result := range []run{got, released(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Errorf(
				"%s: Node exit %d stdout %q stderr %q; got exit %d stdout %q stderr %q",
				difference,
				want.exitCode,
				want.stdout,
				want.stderr,
				result.exitCode,
				result.stdout,
				result.stderr,
			)
		}
	}
	test.Logf("Node exit %d stdout %q stderr %q", want.exitCode, want.stdout, want.stderr)
}

func TestReuseLifetimeAsideThrow(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "aside_throw.a")
}

func TestReuseLifetimeReuseComparatorSpread(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "reuse_comparator_spread.a")
}

func TestReuseLifetimeReuseHandoverThrow(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "reuse_handover_throw.a")
}

func TestReuseLifetimeReuseMoveBeforeReady(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "reuse_move_before_ready.a")
}

package oracle

import (
	"path/filepath"
	"testing"
)

// Each witness runs independently, including the release build and the leak check.
func checkRuntimeWitness(t *testing.T, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, leaked, comparable := nativelyCheckingLeaks(t, program, want.exitCode == 0)
	if leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	for _, result := range []run{got, released(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: Node exit %d stdout %q stderr %q; got exit %d stdout %q stderr %q", difference, want.exitCode, want.stdout, want.stderr, result.exitCode, result.stdout, result.stderr)
		}
	}
	t.Logf("Node exit %d stdout %q stderr %q", want.exitCode, want.stdout, want.stderr)
}

func TestReuseLifetimeAsideThrow(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "aside_throw.a")
}

func TestReuseLifetimeReuseComparatorSpread(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "reuse_comparator_spread.a")
}

func TestReuseLifetimeReuseHandoverThrow(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "reuse_handover_throw.a")
}

func TestReuseLifetimeReuseMoveBeforeReady(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "reuse_move_before_ready.a")
}

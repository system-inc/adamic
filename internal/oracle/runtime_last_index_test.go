package oracle

import (
	"path/filepath"
	"testing"
)

func TestRuntimeLastIndexOfMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/runtime_last_index_of.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	sanitized, binary := natively(t, program)
	for name, got := range map[string]run{"sanitized": sanitized, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s; Node stdout %q, native stdout %q, stderr %q", name, diff, want.stdout, got.stdout, got.stderr)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	t.Logf("Node, JavaScript, release and sanitizers agree on %d output bytes", len(want.stdout))
}

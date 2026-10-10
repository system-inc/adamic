package oracle

import (
	"path/filepath"
	"testing"
)

func TestCallTargetThrowAgreesWithNode(t *testing.T) {
	t.Parallel()
	callTargetFixtureAgrees(t, "statements_small_throw.a", "true/Error/stored stored \nfinally\ninnerinner\n")
}

func TestDirectClosureCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	callTargetFixtureAgrees(t, "nested_mutual.a", "true false 22\nfalse true 24\n")
}

// Reuse registered fixtures, whose allocation counts are already recorded.
func callTargetFixtureAgrees(t *testing.T, name, stdout string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != stdout {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, observation := range []struct {
		name string
		got  run
	}{
		{"JavaScript", onJavaScriptBackend(t, program)},
		{"native release", releasedUncached(t, program)},
	} {
		if difference := disagreement(want, observation.got); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", observation.name, difference, observation.got.exitCode, observation.got.stdout, observation.got.stderr)
		}
	}
	sanitized, binary := nativelyUncached(t, program)
	if difference := disagreement(want, sanitized); difference != "" {
		t.Errorf("native sanitized: %s: exit %d, stdout %q, stderr %q", difference, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Errorf("native leaks: %s", report)
	}
}

package oracle

import (
	"path/filepath"
	"testing"
)

func TestCallTargetThrowAgreesWithNode(test *testing.T) {
	test.Parallel()
	callTargetFixtureAgrees(test, "statements_small_throw.a", "true/Error/stored stored \nfinally\ninnerinner\n")
}

func TestDirectClosureCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	callTargetFixtureAgrees(test, "nested_mutual.a", "true false 22\nfalse true 24\n")
}

// Reuse registered fixtures, whose allocation counts are already recorded.
func callTargetFixtureAgrees(test *testing.T, name, stdout string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != stdout {
		test.Fatalf("Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	for _, observation := range []struct {
		name string
		got  run
	}{
		{"JavaScript", onJavaScriptBackend(test, program)},
		{"native release", releasedUncached(test, program)},
	} {
		if difference := disagreement(want, observation.got); difference != "" {
			test.Errorf(
				"%s: %s: exit %d, stdout %q, stderr %q",
				observation.name,
				difference,
				observation.got.exitCode,
				observation.got.stdout,
				observation.got.stderr,
			)
		}
	}
	sanitized, binary := nativelyUncached(test, program)
	if difference := disagreement(want, sanitized); difference != "" {
		test.Errorf(
			"native sanitized: %s: exit %d, stdout %q, stderr %q",
			difference,
			sanitized.exitCode,
			sanitized.stdout,
			sanitized.stderr,
		)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Errorf("native leaks: %s", report)
	}
}

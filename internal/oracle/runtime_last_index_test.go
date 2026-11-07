package oracle

import (
	"path/filepath"
	"testing"
)

func TestRuntimeLastIndexOfMatchesNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/runtime_last_index_of.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	sanitized, binary := natively(test, program)
	for name, got := range map[string]run{
		"sanitized":  sanitized,
		"release":    released(test, program),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if diff := disagreement(want, got); diff != "" {
			test.Errorf(
				"%s: %s; Node stdout %q, native stdout %q, stderr %q",
				name,
				diff,
				want.stdout,
				got.stdout,
				got.stderr,
			)
		}
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Errorf("leaks: %s", leaked)
	}
	test.Logf("Node, JavaScript, release and sanitizers agree on %d output bytes", len(want.stdout))
}

package oracle

import (
	"path/filepath"
	"testing"
)

// A class's private x isn't reached by the string "#x" on Node, and natively in judges a class instance by
// its public shape, so the program lowers and agrees on sanitized and release native and the JavaScript
// backend (probed on Golem on main 80c2d328, Nov 2). "#message" in an Error is held to Node beside the
// Error's other keys in error_message_presence_test.go. The fixtures sit in private_key_in, since the flow
// and fresh corpora lower every program at the top of testdata.
func TestPrivateKeyInClassAgrees(test *testing.T) {
	test.Parallel()
	for _, file := range []string{"private_hash.a", "private_hash.ts"} {
		test.Run(file, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/private_key_in", file),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if disagreement(run{stdout: []byte("1 false\n")}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("Lower: %v", loweredError)
			}
			native, _ := natively(test, program)
			for name, got := range map[string]run{
				"native":     native,
				"release":    released(test, program),
				"javascript": onJavaScriptBackend(test, program),
			} {
				if problem := disagreement(truth, got); problem != "" {
					test.Errorf(
						"%s: %s: exit %d, stdout %q, stderr %q",
						name,
						problem,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
		})
	}
}

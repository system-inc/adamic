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
func TestPrivateKeyInClassAgrees(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"private_hash.a", "private_hash.ts"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/private_key_in", file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte("1 false\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

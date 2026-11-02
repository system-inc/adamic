package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// "#message" in an Error made without a message is false on Node, since the message is inherited, but
// natively that Error keeps its inherited "" in a slot named #message, which adamic_has_property finds on a
// classless object. in with exactly that key is NotYet at the key (unknown.go's inProperty), in .a and .ts.
// Any other # key stays an ordinary string key, and a class's privates are judged by its public shape, so
// private_hash ("#x" in an instance with a private x) lowers and agrees with Node, as
// unknown_narrowing.a's '#hidden' and '#ordinary' lines do. The fixtures sit in private_key_in, since the
// flow and fresh corpora lower every program at the top of testdata. Mutant: the check dropped from
// inProperty, and message_hash lowers again and prints true natively.
func TestPrivateKeyInNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, stdout string }{
		{"message_hash.a", "message_hash.a:4:48", "false\n"},
		{"message_hash.ts", "message_hash.ts:4:48", "false\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/private_key_in", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), "in with the key #message") {
				t.Fatalf("want in with the #message key NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}

// A class's private x isn't reached by the string "#x" on Node, and natively in judges a class instance by
// its public shape, so the program lowers and agrees on sanitized and release native and the JavaScript
// backend (probed on Golem on main 80c2d328, Nov 2).
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

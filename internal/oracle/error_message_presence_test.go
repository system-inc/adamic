package oracle

import (
	"path/filepath"
	"testing"
)

// An Error made without a message has no own message on Node, but Error.prototype supplies one, "", so
// 'message' in e is true and e.message reads "". Natively such an Error has the absent-message shape, which
// keeps that "" in #message (exceptions.c), and the in test and the dynamic read (union.c) looked for
// message by its own name, so on main no_message and no_message_unguarded printed "none" natively where
// Node prints "". Each is held to Node on sanitized and release native and the JavaScript backend, with an
// Error that has a message beside them. The fixtures sit in error_message_presence, since the flow and fresh
// corpora lower every program at the top of testdata. Mutants: adamic_has_property without
// adamic_error_field_name turns no_message red natively; adamic_dynamic_property without it turns
// no_message_unguarded red, since its in test passes and String(e.message) is a dynamic read.
func TestErrorMessagePresenceAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"no_message.a", "\"\"\n"},
		{"no_message.ts", "\"\"\n"},
		{"message.a", "\"x\"\n"},
		{"message.ts", "\"x\"\n"},
		{"no_message_unguarded.a", "\"\"\n"},
		{"no_message_unguarded.ts", "\"\"\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/error_message_presence", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
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

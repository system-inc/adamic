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
//
// hash_message asks '#message' and 'message' of an Error made without a message and one made with 'x',
// then '#ordinary' of a plain object: false, false, true, true, true on Node. The #message slot made the
// first true natively, so adamic_has_property answers false for a # key on the absent-message shape, and
// only the plain key "message", mapped by adamic_error_field_name, reaches that slot. Mutant: that guard
// removed from union.c, and hash_message's first line prints true natively and goes red.
func TestErrorMessagePresenceAgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"no_message.a", "\"\"\n"},
		{"no_message.ts", "\"\"\n"},
		{"message.a", "\"x\"\n"},
		{"message.ts", "\"x\"\n"},
		{"no_message_unguarded.a", "\"\"\n"},
		{"no_message_unguarded.ts", "\"\"\n"},
		{"hash_message.a", "false\nfalse\ntrue\ntrue\ntrue\n"},
		{"hash_message.ts", "false\nfalse\ntrue\ntrue\ntrue\n"},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/error_message_presence", probe.file),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
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

package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A ?. read that lowers to a read by run-time key, a record that may be null or undefined (Seats' program) or a
// field declared unknown, tests its receiver for null and undefined before the key read, as Node does
// (lower's optionalDynamicProperty, #0q0cv0m). On main the key read took the nullish receiver and stopped: natively
// in union.c's dynamic_property, and on the JavaScript backend with a TypeError. Each program here agrees with
// Node natively, sanitized with its leak check, and on the JavaScript backend, and Node's output is pinned so a
// program can't pass by printing nothing. present_control.ts is the same record read with no ?., which agreed on
// main. The mutants are in testdata/optional_dynamic_receiver/mutants.md.
func TestOptionalDynamicReceiverAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string]string{
		"null_receiver.ts":         "missing\nset\nmissing\n",
		"null_only.ts":             "missing\nset\n",
		"unknown_field.ts":         "missing\nset\n",
		"unknown_field_nullish.ts": "missing\nset\nmissing\n",
		"present_control.ts":       "set\nmissing\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionalDynamicReceiverPath(t, name)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != printed {
				t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Errorf("%s: %s: Node %q; %s exit %d stdout %q stderr %q", backend, difference, node.stdout, backend, got.exitCode, got.stdout, got.stderr)
				}
			}
			if leaked := leaks(t, program, sanitized); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
}

// A record that may be undefined but never null isn't held as a union, so its ?. read never reaches the key read:
// it stops at compile time in records.go's recordAccess, as it did on main, in today's words.
func TestOptionalDynamicReceiverUndefinedRecordStops(t *testing.T) {
	t.Parallel()
	path := optionalDynamicReceiverPath(t, "undefined_only.ts")
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "optional record property access" || notYet.Where != path+":6:17" {
		t.Fatalf("want NotYet %q at :6:17, got %v", "optional record property access", err)
	}
}

// optionalDynamicReceiverPath is a fixture of optional_dynamic_receiver, absolute, as a stop's Where names it.
func optionalDynamicReceiverPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_dynamic_receiver", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

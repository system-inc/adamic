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

// A chain that goes on past the ?. read, value?.label.length, is optional_chain.go's, never this head's helper: its
// guard holds a union receiver only where nullableObjectUnion takes it, and a record that may be null or undefined
// isn't one, so the chain stops at compile time in the words it stopped with on main, while Node runs it.
func TestOptionalDynamicReceiverChainStops(t *testing.T) {
	t.Parallel()
	path := optionalDynamicReceiverPath(t, "chain.ts")
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "undefined\n3\nundefined\n" {
		t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "an optional chain longer than one step" || notYet.Where != path+":6:24" {
		t.Fatalf("want NotYet %q at :6:24, got %v", "an optional chain longer than one step", err)
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

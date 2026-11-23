package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An object literal's own field named like an Object.prototype method, read through a view that hides it and
// called, stops at compile time as a class's method or field of that name already does (#jm2ba55):
// prototypeHazard relates each allocation to the view, and an object literal is related by its regular widened
// type, so its freshness can't make it skip the walk. On main the literal's field was passed over: native
// answered Object.prototype's hasOwnProperty, true, where Node calls the field and prints false, and the
// JavaScript backend exited 70. Each stop is lowering's, so it holds both backends. Node's own answer is pinned
// beside it, so a fixture that stopped hiding its override shows here. Mutant (scratch-jm): the literal related by
// its fresh type again turns every literal row red.
func TestPrototypeOverrideFieldsThroughAViewStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		file, words, node string
		nodeExits         bool
	}{
		{"has_own_field.ts", "a user-defined hasOwnProperty may be hidden by the view", "false\n", false},
		{"has_own_number_field.ts", "a user-defined hasOwnProperty may be hidden by the view", "", true},
		{"property_is_enumerable_field.ts", "a user-defined propertyIsEnumerable may be hidden by the view", "false\n", false},
		{"to_string_field.ts", "a user-defined toString may be hidden by the view", "mine\n", false},
		{"to_locale_string_field.ts", "a user-defined toLocaleString may be hidden by the view", "mine\n", false},
		// Object.prototype.toString.call reads the tag and never the own toString, so Node prints [object Object];
		// it stops all the same, as it does for a class's toString, since the view walk can't yet tell the roads
		// apart.
		{"tag_call_field.ts", "a user-defined toString may be hidden by the view", "[object Object]\n", false},
		// The class method's stop, which main already had.
		{"has_own_method.ts", "a user-defined hasOwnProperty may be hidden by the view", "false\n", false},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/prototype_override_fields", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if probe.nodeExits {
				if truth.exitCode == 0 {
					t.Fatalf("source Node: exit 0, stdout %q, want the TypeError calling a number", truth.stdout)
				}
			} else if truth.exitCode != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q, want %q", truth.exitCode, truth.stdout, truth.stderr, probe.node)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.file) || !strings.Contains(err.Error(), probe.words) {
				t.Fatalf("want NotYet with %q, got %v", probe.words, err)
			}
		})
	}
}

// A literal the view hides nothing of still lowers and agrees with Node on every backend, sanitized and leaking
// nothing: one with only the view's field, and one with a field the view lacks that names no Object.prototype
// method, which the relation by regular type now walks.
func TestPrototypeOverrideFieldsControlsAgree(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, node string }{
		{"no_override.ts", "true\n"},
		{"unrelated_extra_field.ts", "true\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/prototype_override_fields/" + probe.file
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: exit %d, stdout %q, want %q", truth.exitCode, truth.stdout, probe.node)
			}
			generatorAgrees(t, relative)
		})
	}
}

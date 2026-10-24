package lower

import (
	"errors"
	"strings"
	"testing"
)

// Union values fit a Map slot, but that says nothing about unproven types,
// two-word booleans, or the representation and comparison of Map keys. A brand whose
// member can hold a value is refused at its declaration, before any Map reads it.
func TestMapUnionValueBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, reason string
		refused              bool
	}{
		{"null_and_undefined", `const map = new Map<string, string | null | undefined>();`, "a Map of", false},
		{"boolean_and_undefined", `const map = new Map<string, boolean | undefined>();`, "a Map of", false},
		{"unknown_value", `const map = new Map<string, unknown>();`, "a Map of", false},
		{"branded_string", `type Path = string & { readonly _path: never }; const map = new Map<string, Path>();`, "a primitive brand member _path whose type is not void", true},
		{"intersection_reference_field", `interface Node { readonly child: { readonly value: number }; } type Value = Node & { readonly isTypeOnly: true }; const map = new Map<string, Value>();`, "a Map of", false},
		{"union_keys", `const map = new Map<string | number, string>();`, "a Map whose keys", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if probe.refused {
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
					t.Fatalf("want Refused containing %q, got %v", probe.reason, err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want NotYet containing %q, got %v", probe.reason, err)
			}
		})
	}
}

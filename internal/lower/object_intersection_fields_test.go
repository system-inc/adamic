package lower

import (
	"errors"
	"strings"
	"testing"
)

// An object intersection is held as an object whatever its fields hold, so long as a record would
// hold each one: an object field and one that holds the intersection again lower. A Weak field and
// a two-word optional boolean have no such storage (recordStorageType), so each stays NotYet.
func TestObjectIntersectionFields(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		admitted     bool
	}{
		{"object field", "interface Label { readonly text: string; }\ntype Labeled = { readonly id: number } & { readonly label: Label };\nconst value: Labeled = { id: 1, label: { text: 'built' + '-label' } };\nconsole.log(value.label.text);\n", true},
		{"itself", "type Chain = { readonly value: number } & { readonly next: Chain | undefined };\nconst chain: Chain = { value: 1, next: { value: 2, next: undefined } };\nconsole.log(`${chain.value + (chain.next === undefined ? 0 : chain.next.value)}`);\n", true},
		{"weak field", "import type { Weak } from 'adamic';\ninterface Label { readonly text: string; }\ntype Held = { readonly id: number } & { readonly label: Weak<Label> };\nconst value: Held = { id: 1, label: undefined };\nconsole.log(`${value.id}`);\n", false},
		{"optional boolean field", "type Flagged = { readonly id: number } & { readonly flag?: boolean };\nconst value: Flagged = { id: 1, flag: true };\nconsole.log(`${value.id}`);\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.admitted {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var stop *NotYet
			if !errors.As(err, &stop) {
				t.Fatalf("want NotYet, got %v", err)
			}
		})
	}
}

// adamic_object_assign retains and releases by the source slot's reference flag, which is safe only
// while a name is held alike on both sides. A field that's a string in the target and a boxed
// string | number in the source is refused, and so is assigning an object with an object field to
// itself: Object.assign copies only number, string and boolean fields.
func TestObjectAssignKeepsFieldsHeldAlike(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"string and boxed union":      "const target = { value: 'built' + '-text' };\nconst source = { value: 1 as string | number };\nObject.assign(target, source);\nconsole.log(target.value);\n",
		"itself with an object field": "const target = { count: 1, label: { text: 'built' + '-label' } };\nconst same = Object.assign(target, target);\nconsole.log(`${same.count} ${same.label.text}`);\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "intersection result") {
				t.Fatalf("want the intersection result refusal, got %v", err)
			}
		})
	}
}

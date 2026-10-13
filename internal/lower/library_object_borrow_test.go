package lower

import (
	"errors"
	"strings"
	"testing"
)

const scout22OwnSignature = `const hasOwnProperty: (this: object, key: string) => boolean = Object.prototype.hasOwnProperty; `

func TestBorrowedObjectOwnProperty(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		scout22OwnSignature + `const value = { n: 1 }; console.log(String(hasOwnProperty.call(value, 'n')));`,
		scout22OwnSignature + `const value = { hasOwnProperty: (key: string): boolean => false }; hasOwnProperty.call(value, 'hasOwnProperty');`,
		scout22OwnSignature + `const value = { n: 1, hidden: undefined }; for (const key in value) { console.log(String(hasOwnProperty.call(value,key))); }`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBorrowedObjectOwnPropertyRefusesUnprovenUses(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{scout22OwnSignature + `const other = hasOwnProperty;`, "escaping"},
		{scout22OwnSignature + `console.log(typeof hasOwnProperty);`, "escaping"},
		{scout22OwnSignature + `export { hasOwnProperty };`, "escaping"},
		{scout22OwnSignature + `const value: { n: number } = { n: 1 }; hasOwnProperty.call(value, 'n');`, "complete plain literal shape"},
		{scout22OwnSignature + `function own(value: { n: number }): boolean { return hasOwnProperty.call(value, 'n'); } console.log(String(own({n:1})));`, "complete plain literal shape"},
	} {
		t.Run(probe.source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want NotYet %q, got %v", probe.reason, err)
			}
		})
	}
}

func TestBorrowedObjectOwnPropertyDoesNotTrustMutableOrShadowedAliases(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`let own: (this: object, key: string) => boolean = Object.prototype.hasOwnProperty; console.log(String(own.call({n:1},'n')));`,
		`const Object = { prototype: { hasOwnProperty: (key: string): boolean => false } }; console.log(String(Object.prototype.hasOwnProperty.call({},'n')));`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("unproven intrinsic alias lowered")
			}
		})
	}
}

// Preserve the grouped probe when prototype observation becomes represented.
func admittedPrototypeProbe(t *testing.T, source string, err error) bool {
	t.Helper()
	if source != `Object.getPrototypeOf({value:1});` {
		return false
	}
	if err != nil {
		t.Fatalf("proven own shape must expose its prototype: %v", err)
	}
	return true
}

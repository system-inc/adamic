package lower

import (
	"errors"
	"testing"
)

func TestClockGenericReturnsT02Shapes(t *testing.T) {
	for _, probe := range []struct {
		name, shape string
		supported   bool
	}{
		{"nested readonly fields", `{ body: { readonly statements: readonly string[] } }`, true},
		{"additional readonly fields", `{ body: { readonly statements: readonly string[] }; readonly extra: { readonly text: string; readonly flag: boolean } }`, true},
		{"unsupported nested array", `{ body: { readonly statements: readonly string[] }; readonly numbers: readonly number[] }`, false},
		{"callable", `{ body: { readonly statements: readonly string[] }; (): string; readonly left: { readonly text: string } }`, false},
		{"constructable", `{ body: { readonly statements: readonly string[] }; new (): { text: string }; readonly left: { readonly text: string } }`, false},
		{"indexed", `{ body: { readonly statements: readonly string[] }; readonly [key: string]: unknown; readonly left: { readonly text: string } }`, false},
		{"container intersection", `readonly string[]`, false},
		{"nested callable", `{ body: { readonly statements: readonly string[] }; readonly field: () => string }`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := `interface Base { readonly name: string; readonly body?: { readonly statements: readonly string[] }; } function make(): (Base & ` + probe.shape + `) | undefined { return undefined; } console.log(typeof make());`
			_, err := lowerSource(t, source)
			if probe.supported {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var stop *NotYet
			var refusal *Refused
			if !errors.As(err, &stop) && !errors.As(err, &refusal) {
				t.Fatalf("want unsupported shape stop, got %v", err)
			}
		})
	}
	t.Run("null remains distinct", func(t *testing.T) {
		_, err := lowerSource(t, `interface Base { readonly name: string; readonly body?: { readonly statements: readonly string[] }; }
function make(): (Base & { body: { readonly statements: readonly string[] } }) | undefined | null { return null; }
console.log(typeof make());`)
		var stop *NotYet
		if !errors.As(err, &stop) {
			t.Fatalf("want a distinct null/undefined representation stop, got %v", err)
		}
	})

}

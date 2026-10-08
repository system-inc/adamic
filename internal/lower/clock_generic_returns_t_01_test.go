package lower

import (
	"errors"
	"testing"
)

func TestClockGenericReturnsT01Shapes(t *testing.T) {
	for _, probe := range []struct {
		name, shape string
		supported   bool
	}{
		{"nested readonly fields", `{ readonly left: { readonly text: string; readonly generated: boolean }; readonly children: readonly string[] }`, true},
		{"callable", `{ (): string; readonly left: { readonly text: string } }`, false},
		{"constructable", `{ new (): { text: string }; readonly left: { readonly text: string } }`, false},
		{"indexed", `{ readonly [key: string]: unknown; readonly left: { readonly text: string } }`, false},
		{"container intersection", `readonly string[]`, false},
		{"nested callable", `{ readonly field: () => string }`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := `interface Base<T> { readonly token: T; } function make(): (Base<"="> & ` + probe.shape + `) | undefined { return undefined; } console.log(typeof make());`
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
		_, err := lowerSource(t, `interface Base<T> { readonly token: T; }
function make(): (Base<"="> & { readonly left: { readonly text: string } }) | undefined | null { return null; }
console.log(typeof make());`)
		var stop *NotYet
		if !errors.As(err, &stop) {
			t.Fatalf("want a distinct null/undefined representation stop, got %v", err)
		}
	})

}

package lower

import (
	"errors"
	"strings"
	"testing"
)

// fieldFitHeld is #6gwzmjx's interface: a readonly field held as a union over a number.
const fieldFitHeld = "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\n"

// An object whose field is stored as a number goes where the slot reads that field as a union: a plain read
// through the slot works the number as a box it never was, and native faulted (ASan SEGV) where Node
// printed the field (#6gwzmjx, live at main d87e6963). It stops where the value goes into the slot,
// whatever reads it after: an argument, a class field and its getter, and a typeof. Mutant: fieldsSeenAs'
// scalar field check dropped, and each row lowers.
func TestFieldStoredOtherwiseThanItsSlotReadsItStops(t *testing.T) {
	t.Parallel()
	wrapper := "class Wrapper {\n\treadonly held: Held;\n\tconstructor(base: Held) { this.held = base; }\n\tget isOne(): boolean { return this.check(); }\n\tcheck(): boolean { return typeof this.held.amount === \"number\"; }\n}\n"
	raw := "const raw = { kind: 'Holder' as const, amount: 1 };\n"
	for _, row := range []struct{ name, source, at string }{
		{"argument read", fieldFitHeld + "function value(held: Held): number | string { return held.amount; }\n" + raw + "console.log(String(value(raw)));\n", "main.ts:4:26"},
		{"argument typeof", fieldFitHeld + "function check(held: Held): boolean { return typeof held.amount === \"number\"; }\n" + raw + "console.log(String(check(raw)));\n", "main.ts:4:26"},
		{"class field method", fieldFitHeld + wrapper + raw + "console.log(String(new Wrapper(raw).check()));\n", "main.ts:9:32"},
		{"class field getter", fieldFitHeld + wrapper + raw + "console.log(String(new Wrapper(raw).isOne));\n", "main.ts:9:32"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var notYet *NotYet
			// The union's members print in the checker's order, so the words are read around it.
			const stored, held = "its field amount, a number, is stored as number and read there as ", ", held as a boxed union"
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, stored) || !strings.Contains(notYet.What, held) || !strings.HasSuffix(notYet.Where, row.at) {
				t.Fatalf("want NotYet %q...%q at %s, got %v", stored, held, row.at, err)
			}
		})
	}
}

// The same object made as the slot's type stores the field as the slot reads it, so it lowers and agrees.
func TestFieldStoredAsItsSlotReadsItLowers(t *testing.T) {
	t.Parallel()
	source := fieldFitHeld + "function value(held: Held): number | string { return held.amount; }\nconst raw: Held = { kind: 'Holder', amount: 1 };\nconsole.log(String(value(raw)));\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "1\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 1, what Node prints", code, stdout, stderr)
	}
}

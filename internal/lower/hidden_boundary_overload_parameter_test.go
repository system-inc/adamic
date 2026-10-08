package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestHiddenBoundaryOverloadLayoutStops(t *testing.T) {
	for _, probe := range []struct{ name, source, diagnostic string }{
		{"alias", `function widen(elem: { readonly value: number }): number;
function widen(elem: { readonly value: number | string }): number { return Number(elem.value); }
const elem = { value: 7 };
console.log(String(widen(elem)));`, "an overload argument requiring another field representation"},
		{"nested", `function outer(): number {
function widen(elem: { readonly value: number }): number;
function widen(elem: { readonly value: number | string }): number { return Number(elem.value); }
return widen({ value: 7 });
}
console.log(String(outer()));`, "with a nested or rest parameter requiring another field representation"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.diagnostic) {
				t.Fatalf("expected layout stop %q, got %v", probe.diagnostic, err)
			}
		})
	}
}

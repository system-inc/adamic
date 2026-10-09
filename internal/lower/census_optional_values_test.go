package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalFunctionValueRelation(t *testing.T) {
	for _, source := range []string{
		`const narrow = { run(x: number): number { return x + 1; } }; const wider: { run(x?: number): number } = narrow; console.log(String(wider.run()));`,
		`const narrow = { run: (x: number): number => x + 1 }; const wider: { run(x?: number): number } = narrow; console.log(String(wider.run()));`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "a function taking number seen as one taking number | undefined") {
				t.Fatalf("got %v, want strict optional-parameter refusal", err)
			}
		})
	}
}

func TestOverloadedShorthandFunctionValueStaysNotYet(t *testing.T) {
	_, err := lowerSource(t, `function f(x: number): number; function f(x: number): number { return x; } const o = { f }; console.log(String(o.f(1)));`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "an overloaded function as a value") {
		t.Fatalf("got %v, want overload-value diagnostic", err)
	}
}

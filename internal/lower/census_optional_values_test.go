package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalFunctionValueRelation(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const narrow = { run(x: number): number { return x + 1; } }; const wider: { run(x?: number): number } = narrow; console.log(String(wider.run()));`,
		`const narrow = { run: (x: number): number => x + 1 }; const wider: { run(x?: number): number } = narrow; console.log(String(wider.run()));`,
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, source)
			var refused *Refused
			if !errors.As(err, &refused) ||
				!strings.Contains(refused.What, "a function taking number seen as one taking number | undefined") {
				test.Fatalf("got %v, want strict optional-parameter refusal", err)
			}
		})
	}
}

func TestOverloadedShorthandFunctionValueStaysNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`function f(x: number): number; function f(x: number): number { return x; } const o = { f }; console.log(String(o.f(1)));`,
	)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "an overloaded function as a value") {
		test.Fatalf("got %v, want overload-value diagnostic", err)
	}
}

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

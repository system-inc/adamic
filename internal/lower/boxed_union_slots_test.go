package lower

import (
	"errors"
	"testing"
)

// These paths carry fitted boxed values. General field/view admission and
// optional boolean ABIs remain governed by their existing rules.
func TestBoxedUnionSlotPaths(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ source, refusal string }{
		{`const mixed: (string | number)[] = ['a', 1];`, ""},
		// FX5 keeps a union array built from a narrower literal refused.
		{`const numbers: (string | number)[] = [1, 2];`, "union-typed array built from a narrower literal"},
		{"const show = (value: string | number): string => `${value}`;", ""},
		{`const made = Array.from({ length: 2 }, (value: string | number | undefined, index: number) => index);`, ""},
	} {
		test.Run(probe.source, func(test *testing.T) {
			_, lowerSourceError := lowerSource(test, probe.source)
			if probe.refusal != "" {
				var refused *Refused
				if !errors.As(lowerSourceError, &refused) || refused.What != probe.refusal {
					test.Fatalf("want FX5 refusal %q, got %v", probe.refusal, lowerSourceError)
				}
				return
			}
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
		})
	}
}

package lower

import (
	"errors"
	"testing"
)

// These paths carry fitted boxed values. General field/view admission and
// optional boolean ABIs remain governed by their existing rules.
func TestBoxedUnionSlotPaths(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, refusal string }{
		{`const mixed: (string | number)[] = ['a', 1];`, ""},
		// FX5 keeps a union array built from a narrower literal refused.
		{`const numbers: (string | number)[] = [1, 2];`, "union-typed array built from a narrower literal"},
		{"const show = (value: string | number): string => `${value}`;", ""},
		{`const made = Array.from({ length: 2 }, (value: string | number | undefined, index: number) => index);`, ""},
	} {
		t.Run(probe.source, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if probe.refusal != "" {
				var refused *Refused
				if !errors.As(err, &refused) || refused.What != probe.refusal {
					t.Fatalf("want FX5 refusal %q, got %v", probe.refusal, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

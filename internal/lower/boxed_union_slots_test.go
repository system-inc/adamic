package lower

import "testing"

// These were blanket Union refusals. Their writers already fit to the declared
// representation; optional booleans and general object fields remain refused.
func TestBoxedUnionSlotPaths(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const mixed: (string | number)[] = ['a', 1];`,
		`const numbers: (string | number)[] = [1, 2];`,
		"const show = (value: string | number): string => `${value}`;",
		`const made = Array.from({ length: 2 }, (value: string | number | undefined, index: number) => index);`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

package lower

import "testing"

// These paths carry fitted boxed values. General field/view admission and
// optional boolean ABIs remain governed by their existing rules.
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

package lower

import (
	"testing"
)

func TestStringLastIndexOfPositionLower(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"console.log(`${'ababa'.lastIndexOf('a', NaN)}`);",
		"console.log(`${'ababa'.lastIndexOf('a', -1.9)}`);",
		"console.log(`${'ababa'.lastIndexOf('a', undefined)}`);",
		"function f(p: number | undefined): number { return 'ababa'.lastIndexOf('a', p); } console.log(`${f(undefined)}`);",
		"console.log(`${String.prototype.lastIndexOf.call('ababa', 'a', 1.9)}`);",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("%s: %v", source, err)
		}
	}
}

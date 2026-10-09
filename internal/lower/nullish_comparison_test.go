package lower

import (
	"errors"
	"strings"
	"testing"
)

// Each read is isolated: an earlier map refusal must not hide the array boundary.
func TestNullableReferenceComparisonNeedsTag(t *testing.T) {
	t.Parallel()
	for name, read := range map[string]string{
		"map":   "new Map<string, RegExpExecArray | null>().get('absent')",
		"array": "([] as (RegExpExecArray | null)[])[3]",
	} {
		for _, literal := range []string{"null", "undefined"} {
			for _, operator := range []string{"===", "!=="} {
				for _, reverse := range []bool{false, true} {
					expression := read + operator + literal
					if reverse {
						expression = literal + operator + read
					}
					t.Run(name+"/"+expression, func(t *testing.T) {
						_, err := lowerSource(t, "console.log(`${"+expression+"}`);")
						var notYet *NotYet
						if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "null and undefined comparison without a tagged reference slot") {
							t.Fatalf("want tagged-slot boundary, got %v", err)
						}
					})
				}
			}
		}
	}
}

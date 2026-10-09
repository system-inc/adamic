package lower

import (
	"errors"
	"testing"
)

func TestObjectSealUnsupportedShapesRemainNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Object.freeze(new RegExp('a', 'g'));`,
		`const map = Object.seal(new Map());`,
		`const set = Object.seal(new Set());`,
		`Object.seal(new Error());`,
		`function seal(date: Date): void { Object.seal(date); }`,
		`function seal(map: Map<string,number>): void { Object.seal(map); }`,
		`Object.seal([1,2]);`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			var refused *NotYet
			if !errors.As(err, &refused) {
				t.Fatalf("want NotYet, got %v", err)
			}
		})
	}
}

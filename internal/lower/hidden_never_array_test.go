package lower

import (
	"errors"
	"testing"
)

func TestHiddenNeverArrayRejectsWritableWidening(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const items: never[] = []; const writable: number[] = items; writable.push(1);`,
		`const items: never[] = []; const writable: string[] = items; writable.push('text');`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("want writable widening refused, got %v", err)
		}
	}
}

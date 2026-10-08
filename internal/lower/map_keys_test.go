package lower

import (
	"errors"
	"testing"
)

func TestMapKeyBrandGuards(t *testing.T) {
	for _, source := range []string{
		`type Key = string & { __brand: number }; const map = new Map<Key, number>(); console.log("done");`,
		`type Key = string & { __brand: never }; const map = new Map<Key, number>(); console.log("done");`,
		`type Key = string & { toString: void }; const map = new Map<Key, number>(); console.log("done");`,
		`type Key = string & { "0": void }; const map = new Map<Key, number>(); console.log("done");`,
		`type Key = string & { (): void }; const map = new Map<Key, number>(); console.log("done");`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Fatalf("unproven key brand must remain unsupported: %v", err)
		}
	}
}

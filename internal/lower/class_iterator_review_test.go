package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestClassIteratorPolymorphicReceiverCannotHideReturn(t *testing.T) {
	_, err := lowerSource(t, `interface Step { readonly value: number; readonly done: boolean }
class Lines {
 [Symbol.iterator](): Lines { return this; }
 next(): Step { return { value: 0, done: false }; }
 first(): void { for (const value of this) { console.log("value"); break; } }
}
class Closing extends Lines {
 return(): Step { console.log('close'); return { value: 0, done: true }; }
}
new Closing().first();`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "hide a return") || !strings.Contains(gap.What, "retain the concrete subclass type") {
		t.Fatalf("want a hidden-return refusal with a concrete-type repair, got %v", err)
	}
}

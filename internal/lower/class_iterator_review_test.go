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

func TestClassIteratorKeysKeepExplicitStringKeysDistinct(t *testing.T) {
	_, err := lowerSource(t, `const source = { value: 1, [Symbol.iterator]() { return { next() { return { value: 0, done: true }; } }; } };
class Named { readonly value = 2; readonly __adamic_symbol_iterator = 'visible'; }
function keys(view: { readonly value: number }): string { return Object.keys(view).join(','); }
console.log(keys(source)); console.log(keys(new Named()));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "explicit __adamic_symbol_iterator string key") || !strings.Contains(refused.Fix, "rename the explicit") {
		t.Fatalf("want a reserved string-key ambiguity refusal with a repair, got %v", err)
	}
}

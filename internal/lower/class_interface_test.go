package lower

import (
	"strings"
	"testing"
)

func TestClassInterfaceRuling(t *testing.T) {
	t.Run("a literal a class and an interface both take", func(t *testing.T) {
		_, err := lowerSource(t, `class A { readonly tag: string = 'a'; bark(): string { return 'woof'; } }
interface Named { bark(): string }
const either: A | Named = { tag: 'literal', bark: () => 'literal' };
function read(value: A | Named): string { if ('tag' in value) { return value.bark(); } return value.bark(); }
console.log(read(either)); console.log(read(new A()));`)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("a tag does not prove different generic arguments", func(t *testing.T) {
		_, err := lowerSource(t, `class Box<T> { readonly item: T; constructor(item: T) { this.item = item; } }
interface Named { readonly item: string }
function cast(value: Box<'a'> | Box<string> | Named): Box<'a'> { return value as Box<'a'>; }`)
		if err == nil || (!strings.Contains(err.Error(), "no-unchecked-cast") && !strings.Contains(err.Error(), "nominal-class")) {
			t.Fatalf("want generic argument refusal, got %v", err)
		}
	})
}

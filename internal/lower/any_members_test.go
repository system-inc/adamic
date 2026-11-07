package lower

import (
	"errors"
	"strings"
	"testing"
)

// A member the program declares any is refused where it's declared, with its fix (#tmdhz1d, step 2,
// #bw3xg7c): written on an interface's property or a type literal's, through an alias of any, on a class
// field, or inferred from a class field's initializer. Language's v2, the program that read text as a
// string natively, now stops at Loose's declaration. Mutant: refuseAnyMember unregistered, under which
// every program here reaches lowering, and v2 stops only at step 1's relation NotYet.
func TestAnyMemberIsRefusedWhereDeclared(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"v2", `interface Loose { readonly kind: number; readonly pos: number; readonly text: any; }
interface HasText { readonly text: string; }
function show(n: Loose): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (text, declared)"},
		{"a type literal's property", `function show(n: { readonly text: any }): void { console.log(typeof n.text); }
show({ text: 5 });`, "a member typed any (text, declared)"},
		{"through an alias of any", `type Anything = any;
interface Bag { readonly item: Anything; }
const b: Bag = { item: 1 }; console.log(typeof b.item);`, "a member typed any (item, declared)"},
		{"a class field", `class Holder { value: any = 1; }
console.log(typeof new Holder().value);`, "a member typed any (value, declared)"},
		{"a class field inferred", `class Holder { value = JSON.parse("1"); }
console.log(typeof new Holder().value);`, "a member typed any (value, inferred from its initializer)"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerTypeScriptAssertionSource(test, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) ||
				!strings.Contains(refused.What, probe.want) ||
				!strings.Contains(refused.Fix, "unknown") {
				test.Fatalf("want Refused with %q and the unknown fix, got %v", probe.want, err)
			}
		})
	}
}

// A member typed unknown, or one whose any is nested (items: any[]), isn't this rule's: the first lowers
// and agrees with Node, and the second is step 1's relation to judge where it's read as a typed member.
func TestAnyMemberLeavesUnknownAndNested(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeAs(test, `interface Opaque { readonly text: unknown; }
const o: Opaque = { text: 5 };
console.log(typeof o.text);`, ".ts")
	_, err := lowerTypeScriptAssertionSource(test, `interface Bag { readonly items: any[]; }
interface Names { readonly items: readonly string[]; }
function show(b: Bag): void { const n: Names = b; console.log(String(n.items[0])); }
show({ items: [5] });`)
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a member typed any (items[]) related to string") {
		test.Fatalf("a nested any: want step 1's relation NotYet, got %v", err)
	}
}

package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A cast from a union of plain objects no literal tag tells apart, to one member or narrower (#axqgdck,
// untaggedUnionArm), lowers in .ts to a checked view of the target with no test at the cast (rulings.md:18):
// no CheckedCast, and its view marked untagged. A source holding undefined beside two or more such members
// takes absentArm's presence check first, whose message is the only cast message. tsc's shapes: a member of
// PackageJsonInfo | MissingPackageJsonInfo, Node out of FileReference | Node, and an Identifier out of
// Declaration | Expression, whose kind is open on both. Mutant: untaggedUnionArm returning false reds every row.
func TestUntaggedUnionCastLowers(t *testing.T) {
	t.Parallel()
	members := "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: string } "
	for _, probe := range []struct {
		name, source string
		absent       bool
	}{
		{"a member", members + "function f(x: A | B): A { return x as A; } console.log(String(f({ a: 1, shared: 's' }).a));", false},
		{"a member out of three", members + "interface C { readonly c: boolean } function f(x: A | B | C): B { return x as B; } console.log(f({ b: 'b', shared: 's' }).b);", false},
		{"narrower than a member", "interface Declaration { readonly kind: number; readonly pos: number } interface Expression { readonly kind: number; readonly end: number } interface Identifier extends Expression { readonly text: string } function f(x: Declaration | Expression): Identifier { return x as Identifier; } console.log(f({ kind: 80, end: 1, text: 't' } as Identifier).text);", false},
		{"a member beside undefined", members + "function f(x: A | B | undefined): A { return x as A; } console.log(String(f({ a: 1, shared: 's' }).a));", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowered, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, function := range lowered.Functions {
				walk(function.Body, func(node any) bool {
					if cast, ok := node.(ir.CheckedCast); ok {
						t.Fatalf("a CheckedCast on %s: an untagged union's cast is a view with no test at the cast", cast.Field)
					}
					return true
				})
			}
			messages := 0
			for _, text := range lowered.Strings {
				if strings.HasPrefix(text, "cast failed:") {
					messages++
				}
			}
			if want := map[bool]int{false: 0, true: 1}[probe.absent]; messages != want {
				t.Fatalf("%d cast messages, want %d: only absentArm's presence check stops at the cast", messages, want)
			}
		})
	}
}

// What the arm leaves as it was, each by its own words: a member that isn't a plain object (a primitive, a
// class instance, a dictionary), a union target, and a union a tag does tell apart, whose payload refinement
// stays refused (rulings.md, the object-union member entry). A union whose tags collide (duplicate enum
// aliases, mixed tag representations) stays refused too, pinned by TestUncheckableCastsStayRefused. In .a the
// arm's cast is refused as every downcast whose checks no proof erases. tsc refuses a sideways cast whose
// types don't overlap itself (TS2352), so no row holds one. Mutants: plainViewObject returning true reds the
// primitive, class and dictionary rows; unionHoldsLiterals returning false reds "a tagged union's refinement"
// and TestUncheckableCastsStayRefused's colliding-tag rows; keptCheckDowncastRefusal without proof.untagged
// reds "in .a".
func TestUntaggedUnionCastsStayRefused(t *testing.T) {
	t.Parallel()
	members := "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: string } "
	for _, probe := range []struct{ name, source, extension, want string }{
		{"a primitive member", members + "function f(x: A | string): A { return x as A; }", ".ts", "a cast the runtime can't check"},
		{"a class instance member", members + "class K { readonly k = 1; } function f(x: A | K): A { return x as A; }", ".ts", "a cast the runtime can't check"},
		{"a dictionary member", members + "function f(x: A | Record<string, number>): A { return x as A; }", ".ts", "a cast the runtime can't check"},
		{"a union target", members + "interface C { readonly c: boolean } function f(x: A | B | C): A | B { return x as A | B; }", ".ts", "a cast the runtime can't check"},
		{"a tagged union's refinement", "interface P { readonly kind: 'p'; readonly v: string | number } interface Q { readonly kind: 'q' } interface PN { readonly kind: 'p'; readonly v: number } function f(x: P | Q): PN { return x as PN; }", ".ts", "a cast the runtime can't check"},
		{"in .a", members + "function f(x: A | B): A { return x as A; }", ".a", "in .a whose checks no proof erases"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want a refusal %q, got %v", probe.want, err)
			}
		})
	}
}

// In a graph program the untagged view pairs each arm of its source with its target (Memory's item 4 on
// #axqgdck): a B's inner is read and written as A's through the view, so a Y the finder never relates to X
// stops the program NotYet, naming the arm, whatever road a later read or write takes (Language's x05 and
// x06: a parameter and an alias typed A). C has no inner, so the union holds none in common and only the arm
// loop pairs it. The same union whose B holds an X lowers. Mutant: untaggedUnionCast not marking its view
// untagged reds "an arm the finder doesn't relate".
func TestUntaggedUnionCastInAGraphProgram(t *testing.T) {
	t.Parallel()
	// X and Y hold leaves of their own, so the view's contract never reaches the cycle's Link.
	declarations := "interface XLeaf { readonly v: number } interface YLeaf { readonly w: string } interface X { readonly x: XLeaf } interface Y { readonly y: YLeaf } interface A { readonly a: number; readonly inner: X } interface C { readonly c: number } "
	use := "function f(u: A | B | C): X { return (u as A).inner; } console.log(String(f({ a: 1, inner: { x: { v: 4 } } }).x.v) + String(link.next === link));"
	for _, probe := range []struct{ name, b, want string }{
		{"an arm the finder doesn't relate", "interface B { readonly b: number; readonly inner: Y } ", "through an arm no test tells (B as A) through its field inner (Y as X)"},
		{"an arm whose fields pair", "interface B { readonly b: number; readonly inner: X } ", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, graphReviewCycle+declarations+probe.b+use, ".ts")
			if probe.want == "" {
				if err != nil {
					t.Fatalf("an untagged view whose arms pair, in a graph program: %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
			}
		})
	}
}

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
		// An optional literal field is no tag (fieldLiteral answers nil for kind?: 'a'), so the union takes this arm,
		// held by the checked reads; recorded here so a ruling either way moves this row (Lower's should).
		{"an optional literal field", "interface A { readonly kind?: 'a'; readonly a: number } interface B { readonly kind?: 'b'; readonly b: string } function f(x: A | B): A { return x as A; } console.log(String(f({ kind: 'a', a: 1 }).a));", false},
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

// What the arm leaves to the stops it found, each by its own words since the cast causes (#b5w3ycg): a member
// that isn't a plain object (a primitive, a class instance, a dictionary), a sub-union target, and a union a tag
// does tell apart, whose payload refinement stays refused (rulings.md, the object-union member entry). A union
// whose tags collide (duplicate enum aliases, mixed tag representations) stays refused too, pinned by
// TestUncheckableCastsStayRefused. In .a the arm's cast is refused as every downcast whose checks no proof
// erases. tsc refuses a sideways cast whose types don't overlap itself (TS2352), so no row holds one. A row
// with no words holds only that the arm leaves it stopped. Mutants: plainViewObject returning true reds the
// class row (a primitive member and a dictionary member are stopped before the arm, J-7j0yf17); unionHoldsLiterals
// returning false reds "a tagged union's refinement" and TestUncheckableCastsStayRefused's colliding-tag rows;
// keptCheckDowncastRefusal without proof.untagged reds "in .a".
func TestUntaggedUnionCastsKeepTheirStops(t *testing.T) {
	t.Parallel()
	members := "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: string } "
	for _, probe := range []struct {
		name, source, extension string
		refused                 bool
		want                    string
	}{
		{"a primitive member", members + "function f(x: A | string): A { return x as A; }", ".ts", false, "which holds primitives"},
		{"a class instance member", members + "class K { readonly k = 1; } function f(x: A | K): A { return x as A; }", ".ts", false, "whose members no literal tag tells apart"},
		{"a dictionary member", members + "function f(x: A | Record<string, number>): A { return x as A; }", ".ts", false, ""},
		// binder.ts:1416's shape: a CallExpression, whose kind is a literal, out of a union of nodes whose kind is
		// open. A view with no test can't prove the literal, so it keeps the untagged stop until its tag road is
		// built (Diagnostics' must). Mutant: declaresLiteral returning false, and the enum-literal rule refuses it.
		{"a target with a literal kind", "enum SyntaxKind { Identifier = 80, CallExpression = 213, Block = 241 } interface Expression { readonly kind: SyntaxKind; readonly end: number } interface Statement { readonly kind: SyntaxKind; readonly pos: number } interface CallExpression extends Expression { readonly kind: SyntaxKind.CallExpression; readonly argumentCount: number } function f(x: Expression | Statement): CallExpression { return x as CallExpression; }", ".ts", false, "a checked view of CallExpression over the union Expression | Statement, whose members no literal tag tells apart"},
		{"a sub-union target", members + "interface C { readonly c: boolean } function f(x: A | B | C): A | B { return x as A | B; }", ".ts", false, "a checked view of A | B over the union A | B | C, whose members no literal tag tells apart"},
		// Runtime's two questions on #axqgdck: a writable field whose type differs by arm is refused at the cast by
		// castUnionWrites (its union type X | Y fits neither arm's slot), so no X is ever written into a B and no
		// cycle the finder can't see closes through one. One nested in a readonly field, which castUnionWrites
		// doesn't walk, is stopped too (overlapViewCast's per-member widening).
		{"a writable field whose type differs by arm", "interface X { readonly x: number } interface Y { readonly y: string } interface A { note: X } interface B { note: Y } function f(u: A | B): A { return u as A; }", ".ts", true, "whose writable field note some member can't hold every value of"},
		{"a writable field nested in a readonly one", "interface X { readonly x: number } interface Y { readonly y: string } interface CX { v: X } interface CY { v: Y } interface A { readonly child: CX } interface B { readonly child: CY } function f(u: A | B): A { return u as A; }", ".ts", false, ""},
		{"a tagged union's refinement", "interface P { readonly kind: 'p'; readonly v: string | number } interface Q { readonly kind: 'q' } interface PN { readonly kind: 'p'; readonly v: number } function f(x: P | Q): PN { return x as PN; }", ".ts", true, "which is narrower than a member of P | Q"},
		{"in .a", members + "function f(x: A | B): A { return x as A; }", ".a", true, "in .a whose checks no proof erases"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatal("lowered: the arm must leave this cast stopped")
			case probe.want == "":
				if !errors.As(err, &refused) && !errors.As(err, &notYet) {
					t.Fatalf("want a stop, got %v", err)
				}
			case probe.refused && (!errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want a refusal holding %q, got %v", probe.want, err)
			case !probe.refused && (!errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
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

package lower

import (
	"strings"
	"testing"
)

// Language's t00 to t03 (#s8gy89r, J-6ng4s5e), the pins of rulings.md:142 as amended (#tvw1zkp). Base's
// kind is the whole enum K, and isSecond's body tests only the tag.
const tagDomainTypes = "enum K { A = 1, B = 2 }\ninterface Base { readonly kind: K }\ninterface First extends Base { readonly kind: K.A; readonly a: string }\ninterface Second extends Base { readonly kind: K.B; readonly b: number }\n"

const tagDomainUse = "if (isSecond(v)) { console.log(String(v.b + 1)); } else { console.log('not'); }\n"

// In .a a tag-only predicate over an open supertype is refused at the predicate, naming :142, the
// supertype and the fixes: t00 builds a Base at K.B with no b, and t03 makes one from a parameter. In .ts
// the same body stays :144's checked predicate, which stops loudly at the read as main does.
// Mutant: predicateTagDomainRefusal answering nil lowers both .a rows.
func TestTagOnlyPredicateOverAnOpenSupertypeIsRefusedInA(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"t00 a Base built at K.B", "function isSecond(n: Base): n is Second { return n.kind === K.B; }\nconst v: Base = { kind: K.B };\n"},
		{"t03 a Base made from a parameter", "function isSecond(n: Base): n is Second { return n.kind === K.B; }\nfunction make(k: K): Base { return { kind: k }; }\nconst v = make(K.B);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := tagDomainTypes + probe.source + tagDomainUse
			_, err := lowerSourceAs(t, source, ".a")
			if err == nil || !strings.Contains(err.Error(), "rulings.md:142") || !strings.Contains(err.Error(), "over Base") || !strings.Contains(err.Error(), "take a closed union of the members as the parameter in place of Base") {
				t.Fatalf(".a: want the predicate refused under rulings.md:142 over Base, got %v", err)
			}
			program, err := lowerSourceAs(t, source, ".ts")
			if err != nil {
				t.Fatalf(".ts: want the checked predicate lowered, got %v", err)
			}
			if code, _, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, "is not initialized") {
				t.Fatalf(".ts native: exit %d, stderr %q; want the checked read's stop, exit 70", code, stderr)
			}
		})
	}
}

// t01, an honest Second seen as a Base, and t02, the parameter a closed union, are proven in both, and agree
// with Node on the JavaScript backend and sanitized natively: every source of K.B in K(Base) is a Second.
func TestTagOnlyPredicateOverAClosedTypeIsProven(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"t01 an honest Second", tagDomainTypes + "function isSecond(n: Base): n is Second { return n.kind === K.B; }\nconst s: Second = { kind: K.B, b: 4 };\nconst v: Base = s;\n"},
		{"t02 a closed union", "enum K { A = 1, B = 2 }\ninterface First { readonly kind: K.A; readonly a: string }\ninterface Second { readonly kind: K.B; readonly b: number }\ntype N = First | Second;\nfunction isSecond(n: N): n is Second { return n.kind === K.B; }\nconst s: Second = { kind: K.B, b: 4 };\nconst v: N = s;\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				program := lowersAndAgreesWithNodeAs(t, probe.source+tagDomainUse, extension)
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "5\n" {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "5\n")
				}
			})
		}
	}
}

// Adamic's pair (#s8gy89r, Nov 21): Seats' program 02, Initializer = Expression | List, where Expression's
// kind is the whole enum and the closed-kind rule closes it on Text's K.A, so a value carrying K.B is a List.
// In .a its tag-only predicate is proven, and the program lowers and agrees with Node; on main the predicate
// was refused as an open hazard. One direct Expression build at K.B puts K.B in K(Expression) through a
// source that isn't a List, and the predicate is refused at :142. 02's own body adds && node.enabled, a field
// read the body prover reads as opaque, on main as here, so this pins its tag test alone.
// Mutants: enumTagHazardOf in place of enumTagWalkHazardOf in .a refuses the proven rows; predicateTagDomainRefusal
// answering nil lowers the refused one.
func TestTagOnlyPredicateOverAConstructionClosedSupertype(t *testing.T) {
	t.Parallel()
	const types = "enum K { A, B, C }\ninterface Expression { readonly kind: K; readonly brand: undefined; readonly enabled: boolean }\ninterface Text extends Expression { readonly kind: K.A }\ninterface List { readonly kind: K.B; readonly enabled: boolean; readonly label: string }\ntype Initializer = Expression | List;\n"
	const exact = "function isList(node: Initializer): node is List { return node.kind === K.B; }\n"
	const use = "function show(node: Initializer): string { if (!isList(node)) { return 'other'; } return node.label; }\nconst list: List = { kind: K.B, enabled: true, label: 'before' };\nconst text: Text = { kind: K.A, brand: undefined, enabled: false };\nconsole.log(show(list));\nconsole.log(show(text));\n"
	t.Run("proven", func(t *testing.T) {
		t.Parallel()
		program := lowersAndAgreesWithNodeAs(t, types+exact+use, ".a")
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "before\nother\n" {
			t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "before\nother\n")
		}
	})
	t.Run("an Expression built at K.B", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, types+exact+use+"const lie: Expression = { kind: K.B, brand: undefined, enabled: true };\nconsole.log(show(lie));\n", ".a")
		if err == nil || !strings.Contains(err.Error(), "rulings.md:142") || !strings.Contains(err.Error(), "over Expression") {
			t.Fatalf("want the predicate refused under rulings.md:142 over Expression, got %v", err)
		}
	})
}

// A member whose kind is two one-member enums holds any number and has no kind domain, so it's never closed:
// in .a a tag-only predicate over it is refused at the predicate as t00 is, and in .ts it stops at the
// checked read (Lower's question on #tvw1zkp, J-gt8dx1e). Mutant: the member skipped unless its tag is the
// whole open enum (openNumericEnumType) lowers the .a row to the read's stop.
func TestTagOnlyPredicateOverAOneMemberEnumTagIsRefusedInA(t *testing.T) {
	t.Parallel()
	const source = "enum One { X = 2 }\nenum Two { Y = 3 }\ninterface Other { readonly kind: One | Two }\ninterface OtherB extends Other { readonly kind: One.X; readonly b: number }\nfunction isB(n: Other): n is OtherB { return n.kind === One.X; }\nconst o: Other = { kind: One.X };\nif (isB(o)) { console.log(String(o.b + 1)); } else { console.log('not'); }\n"
	if _, err := lowerSourceAs(t, source, ".a"); err == nil || !strings.Contains(err.Error(), "rulings.md:142") || !strings.Contains(err.Error(), "over Other") {
		t.Fatalf(".a: want the predicate refused under rulings.md:142 over Other, got %v", err)
	}
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf(".ts: want the checked predicate lowered, got %v", err)
	}
	if code, _, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, "is not initialized") {
		t.Fatalf(".ts native: exit %d, stderr %q; want the checked read's stop, exit 70", code, stderr)
	}
}

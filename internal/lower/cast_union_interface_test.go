package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A represented union cast to one interface that only overlaps it (#b5w3ycg 2d, overlapTarget): tsc's
// DeclarationDiagnosticProducing as NamedDeclaration. Its claim is a shape, so it lowers in .ts to a checked view
// of the interface with no test at the cast (Adamic's ruling under rulings.md:18): no CheckedCast and no cast
// message, whatever the members' tags. Language's width roads (a value of the member the interface doesn't name,
// holding the interface's field anyway) type-check and lower the same way.
func TestUnionToInterfaceCastLowers(t *testing.T) {
	t.Parallel()
	members := "enum K { A = 1, B = 2, C = 3, D = 4 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface B extends Named { readonly kind: K.B; readonly b: number } interface C extends Named { readonly kind: K.C } interface D { readonly kind: K.D; readonly d: number } type U = A | B | C | D; "
	for _, probe := range []struct{ name, source string }{
		{"member", members + "function f(x: U): Named { return x as Named; }"},
		{"width through a subtype", members + "interface D2 extends D { readonly name: string } function g(u: U): string { return (u as Named).name; } function f(d: D): string { return g(d); } console.log(f({ kind: K.D, d: 1, name: 'hidden' } as D2));"},
		{"duplicate tags", "enum K { A = 1, D = 1 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface D { readonly kind: K.D; readonly d: number } function f(x: A | D): Named { return x as Named; }"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					if cast, ok := node.(ir.CheckedCast); ok {
						t.Fatalf("a CheckedCast on %s: the overlap cast is a view with no test at the cast", cast.Field)
					}
					return true
				})
			}
			for _, text := range program.Strings {
				if strings.HasPrefix(text, "cast failed:") {
					t.Fatalf("a cast message %q: the overlap cast never stops at the cast", text)
				}
			}
		})
	}
}

// What overlapTarget leaves refused, each by its own reason: a target no member upcasts to, and the same cast in
// .a, refused as every downcast whose checks no proof erases. Then the members TypeScript relates to the target that
// Adamic's relation doesn't (a readonly field seen as writable, a method's parameter, a missing optional field):
// rules that run before castProof refuse each today, and this says which moved if one lifts.
func TestUnionToInterfaceCastsStayRefused(t *testing.T) {
	t.Parallel()
	members := "enum K { A = 1, B = 2, D = 4 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface D { readonly kind: K.D; readonly d: number } "
	for _, probe := range []struct{ name, source, extension, want string }{
		{"no member upcasts", members + "interface Wider { readonly kind: K.A; readonly name: string; readonly other: number } function f(x: A | D): Wider { return x as Wider; }", ".ts", "a cast the runtime can't check"},
		{"in .a", members + "function f(x: A | D): Named { return x as Named; }", ".a", "in .a whose checks no proof erases"},
		// A member the view could write a field it can't hold: B's box holds v: 'b', and seen as Holder a write of any
		// string to box.v would break it. overlapViewCast's per-member widened loop would make it NotYet, but no shape
		// reaches that loop today: main's view-site rule (a cast never launders a view its slot would refuse) refuses
		// the same widening at the cast first, by its own words. C, which has no box, keeps the union as a whole from
		// relating to Holder, and it's still refused there. If that rule lifts, this goes red and the loop is the stop.
		{"a member the view could write past", "enum K { A = 1, B = 2, C = 3 } interface Holder { readonly box: { v: string } } interface A { readonly kind: K.A; readonly box: { v: string } } interface B { readonly kind: K.B; readonly box: { v: 'b' } } interface C { readonly kind: K.C; readonly c: number } function f(x: A | B | C): Holder { return x as Holder; }", ".ts", "seen as Holder, which can write string where \"b\" is read"},
		{"tsc alone relates: readonly seen as writable", "enum K { A = 1, B = 2 } interface Named { name: string } interface A { readonly kind: K.A; name: string } interface B { readonly kind: K.B; readonly name: string } function f(x: A | B): Named { return x as Named; }", ".ts", ""},
		{"tsc alone relates: a method's parameter", "enum K { A = 1, B = 2 } interface Handler { handle(x: string | number): void } interface A { readonly kind: K.A; handle(x: string | number): void } interface B { readonly kind: K.B; handle(x: string): void } function f(x: A | B): Handler { return x as Handler; }", ".ts", ""},
		{"tsc alone relates: a missing optional field", "enum K { A = 1, B = 2 } interface Named { readonly name: string; readonly extra?: number } interface A { readonly kind: K.A; readonly name: string; readonly extra?: number } interface B { readonly kind: K.B; readonly name: string } function f(x: A | B): Named { return x as Named; }", ".ts", ""},
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

package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A represented union cast to one interface some of its members satisfy (#b5w3ycg 2d, memberUpcastTarget): tsc's
// DeclarationDiagnosticProducing as NamedDeclaration. It lowers in .ts to one CheckedCast whose allowed tags are
// exactly the members Adamic's relation upcasts to the interface, so the member that isn't one fails the tag test.
func TestUnionToInterfaceCastLowers(t *testing.T) {
	t.Parallel()
	source := "enum K { A = 1, B = 2, C = 3, D = 4 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface B extends Named { readonly kind: K.B; readonly b: number } interface C extends Named { readonly kind: K.C } interface D { readonly kind: K.D; readonly d: number } type U = A | B | C | D; function f(x: U): Named { return x as Named; }"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	casts := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			cast, ok := node.(ir.CheckedCast)
			if !ok {
				return true
			}
			casts++
			allowed := []float64{}
			for _, value := range cast.Allowed {
				allowed = append(allowed, value.(ir.NumberConstant).Value)
			}
			if cast.Field != "kind" || len(allowed) != 3 || allowed[0] != 1 || allowed[1] != 2 || allowed[2] != 3 {
				t.Fatalf("checked %s for %v, want kind for A, B and C (1, 2, 3)", cast.Field, allowed)
			}
			if cast.Message != "cast failed: this U is not a Named" {
				t.Fatalf("message %q", cast.Message)
			}
			return true
		})
	}
	if casts != 1 {
		t.Fatalf("got %d checked casts, want 1", casts)
	}
}

// What memberUpcastTarget leaves refused: a target no member upcasts to, duplicate tags, and the same cast in .a,
// refused as every member cast is, each by its own reason. Then the members TypeScript relates to the target that
// Adamic's relation doesn't: a readonly field seen as writable, a method whose parameter only tsc's bivariance
// relates, and an optional field the member lacks. Each is refused today by a rule that runs before castProof
// (erasedWritableViewCast, the method-variance refusal, the optional-relation refusal), so memberUpcastTarget's
// own guard on them is defense in depth with no mutant that can go red: if one of those rules lifts, the guard
// keeps the cast refused, and this test says which shape moved.
func TestUnionToInterfaceCastsStayRefused(t *testing.T) {
	t.Parallel()
	members := "enum K { A = 1, B = 2, D = 4 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface D { readonly kind: K.D; readonly d: number } "
	const unchecked = "a cast the runtime can't check"
	for _, probe := range []struct{ name, source, extension, want string }{
		{"no member upcasts", members + "interface Wider { readonly kind: K.A; readonly name: string; readonly other: number } function f(x: A | D): Wider { return x as Wider; }", ".ts", unchecked},
		{"duplicate tags", "enum K { A = 1, D = 1 } interface Named { readonly name: string } interface A extends Named { readonly kind: K.A } interface D { readonly kind: K.D; readonly d: number } function f(x: A | D): Named { return x as Named; }", ".ts", unchecked},
		{"in .a", members + "function f(x: A | D): Named { return x as Named; }", ".a", "in .a whose checks no proof erases"},
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

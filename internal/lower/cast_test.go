package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestUncheckableCastsStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"unknown interface", "interface User { readonly name: string } function read(x: unknown): User { return x as User; }"},
		{"callback signature", "function cast(x: (n: number | string) => number | string): (n: number) => number { return x as (n: number) => number; }"},
		{"mutable widening", "function cast(x: { value: 'dog' }): { value: string } { return x as { value: string }; }"},
		{"mutable union tag", "type A = { kind: 'a'; readonly n: number }; type B = { kind: 'b'; readonly n: number }; function f(x: A | B): A { return x as A; }"},
		{"mutable union payload", "type A = { readonly kind: 'a'; n: 1 }; type B = { readonly kind: 'b'; n: 2 }; function f(x: A | B): A { return x as A; }"},
		{"double assertion", "const n = ('wrong' as unknown) as number;"},
		{"related double assertion", "function f(n: number): number { return (n as unknown) as number; }"},
		{"primitive union", "function f(x: number | string | boolean): number { return x as number; }"},
		{"tuple length", "function f(x: readonly [1] | readonly [1, 2]): readonly [1] { return x as readonly [1]; }"},
		{"unrelated class", "class A { readonly n = 1; } class B { readonly n = 1; } function f(x: A): B { return x as B; }"},
		{"structural class", "class A { readonly n = 1; } function f(x: { readonly n: number }): A { return x as A; }"},
		{"payload refinement", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'b'; readonly n: number }; function f(x: A | B): { readonly kind: 'a'; readonly n: 1 } { return x as { readonly kind: 'a'; readonly n: 1 }; }"},
		{"duplicate literal tags", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'a'; readonly s: string }; function f(x: A | B): A { return x as A; }"},
		{"duplicate enum values", "enum K { A = 1, B = 1 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly s: string }; function f(x: A | B): A { return x as A; }"},
		{"different enums same value", "enum K { A = 1 } enum L { B = 1 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: L.B; readonly s: string }; function f(x: A | B): A { return x as A; }"},
		{"mixed tag representations", "enum K { A = 1, B = 'b' } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly s: string }; function f(x: A | B): A { return x as A; }"},
		{"erased generic argument", "class Base {} class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } } function f(x: Base): Box<number> { return x as Box<number>; }"},
		{"incompatible generic arguments", "class Box<T> { readonly value: T; constructor(value: T) { this.value = value; } } function f(x: Box<'a'>): Box<string> { return x as Box<string>; }"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/no-unchecked-cast") || !strings.Contains(err.Error(), "enum tags") || !strings.Contains(err.Error(), "nominal class ancestry") {
				t.Fatalf("want unchecked-cast refusal and accepted forms, got %v", err)
			}
		})
	}
}

func TestCheckedCastProofAndElision(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		checks       int
	}{
		{"enum member", "enum K { A, B } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly n: number }; function f(x: A | B): A { return x as A; }", 1},
		{"invariant mutable payload", "type A = { readonly kind: 'a'; n: number }; type B = { readonly kind: 'b'; n: number }; function f(x: A | B): A { return x as A; }", 1},
		{"sub-union", "enum K { A, B, C } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; type C = { readonly kind: K.C }; function f(x: A | B | C): A | C { return x as A | C; }", 1},
		{"class downcast", "class A { readonly n = 1; } class B extends A { readonly b = 2; } function f(x: A): B { return x as B; }", 1},
		{"empty subclass", "class A {} class B extends A {} function f(x: A): B { return x as B; }", 1},
		{"generic ancestry", "class Base<T> { readonly n: T; constructor(n: T) { this.n = n; } } class Derived<T> extends Base<T> {} function f(x: Base<number>): Derived<number> { return x as Derived<number>; }", 1},
		{"class upcast", "class A { readonly n = 1; } class B extends A { readonly b = 2; } function f(x: B): A { return x as A; }", 0},
		{"flow narrowing", "type A = { readonly kind: 'a' }; type B = { readonly kind: 'b' }; function f(x: A | B): void { if (x.kind === 'a') { const a = x as A; } }", 0},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			checks := 0
			for _, function := range program.Functions {
				if function.Name == "checked_class_cast" {
					checks++
				}
				walk(function.Body, func(node any) bool {
					if _, ok := node.(ir.CheckedCast); ok {
						checks++
					}
					return true
				})
			}
			if checks != probe.checks {
				t.Fatalf("got %d checks, want %d", checks, probe.checks)
			}
		})
	}
}

// Optional widening now installs the shared checked view, so this formerly
// refused upcast is allowed only while the hidden field retains a read check.
func TestOptionalCastKeepsHiddenFieldCheck(t *testing.T) {
	program, err := lowerSource(t, "function cast(x: { readonly n: number }): { readonly n: number; readonly hidden?: number } { return x as { readonly n: number; readonly hidden?: number }; }")
	if err != nil {
		t.Fatal(err)
	}
	if !program.CheckedFields["hidden"] {
		t.Fatal("hidden optional cast lost its checked read")
	}
}

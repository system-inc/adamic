package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestEnumSlotViews(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"variable", "enum E { A, B } function open(n: number): E { return n; }"},
		{"initializer", "enum E { A, B } function run(n: number): void { const value: E = n; }"},
		{"member initializer", "enum E { A, B } const value: E.A = 0;"},
		{"argument", "enum E { A, B } function use(e: E): void {} function run(n: number): void { use(n); }"},
		{"assignment", "enum E { A, B } let value: E = E.A; function run(n: number): void { value = n; }"},
		{"literal property", "enum E { A, B } function run(n: number): void { const box: { readonly value: E } = { value: n }; }"},
		{"array literal", "enum E { A, B } function run(n: number): void { const values: readonly E[] = [n]; }"},
		{"increment", "enum E { A, B } let value: E = E.A; value++;"},
		{"compound", "enum E { A, B } let value: E = E.A; value += 7;"},
		{"array alias", "enum E { A, B } const values: E[] = [E.A]; const numbers: number[] = values; numbers.push(77);"},
		{"readonly into enum array", "enum E { A, B } function run(values: readonly number[]): void { const members: readonly E[] = values; }"},
		{"function view", "enum E { A, B } const open = (n: number): number => n; const closed: (e: E) => E = open;"},
		{"object assign", "enum E { A, B } function run(n: number): void { Object.assign(E, { A: E.A }); }"},
		{"override", "enum E { A, B } class Base { read(n: number): void {} } class Derived extends Base { override read(n: E): void {} }"},
		{"class field", "enum E { A, B } class Box { value: E; constructor(n: number) { this.value = n; } }"},
		{"optional union", "enum E { A, B } function run(n: number): void { const value: E | undefined = n; }"},
		{"map alias", "enum E { A, B } const members = new Map<string, E>(); const numbers: Map<string, number> = members; numbers.set('x', 77);"},
		{"mutable field alias", "enum E { A, B } const box: { value: E } = { value: E.A }; const wide: { value: number } = box; wide.value = 77;"},
		{"generic class invariant", "enum E { A, B } class Box<T> { readonly value: T; constructor(value: T) { this.value = value; } } const box = new Box<E>(E.A); const wide: Box<number> = box;"},
		{"structural numeric enum object", "enum E { A, B } const copy: typeof E = { A: E.A, B: E.B };"},
		{"structural string enum object", "enum E { A = 'a', B = 'b' } const copy = { A: E.A, B: E.B, hidden: 99 } as const; const view: typeof E = copy;"},
		{"structural enum object array", "enum E { A, B } const copies: readonly { readonly A: E.A; readonly B: E.B }[] = [{ A: E.A, B: E.B }]; const views: readonly (typeof E)[] = copies;"},
		{"enum object writable view", "enum E { A, B } const view: { A: number } = E; view.A = 9;"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			// Open numeric enums do not erase generic class argument identity.
			closed := strings.Contains(probe.name, "enum object") || probe.name == "object assign" || probe.name == "generic class invariant"
			if !closed {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("got %v, want Refused", err)
			}
			if probe.name == "generic class invariant" && !strings.Contains(err.Error(), "adamic/nominal-class") {
				t.Fatalf("want nominal generic argument refusal, got %v", err)
			}
			if !strings.Contains(err.Error(), "enum") && !strings.Contains(err.Error(), "readonly") && !strings.Contains(err.Error(), "invariant-mutable") && !strings.Contains(err.Error(), "override") && !strings.Contains(err.Error(), "Object.assign") && !strings.Contains(err.Error(), "nominal-class") {
				t.Fatalf("missing reason: %v", err)
			}
		})
	}
}

func TestEnumSwitchExhaustiveness(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"enum E { A = 'a', B = 'b' } function show(e: E): void { switch (e) { case E.A: break; } }",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "non-exhaustive enum switch") {
			t.Fatalf("got %v", err)
		}
	}
	for _, source := range []string{
		"enum E { A, B } function show(e: E): void { switch (e) { case E.A: break; } }",
		"enum E { A, B } function show(e: E): void { switch (e) { case E.A: break; default: break; } }",
		"enum E { A, Alias = A, B } function show(e: E): void { switch (e) { case E.A: break; case E.B: break; } }",
		"enum E { A, B } function show(e: E.A): void { switch (e) { case E.A: break; } }",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnumLimitsStayLoud(t *testing.T) {
	t.Parallel()
	t.Fatal("deliberate failure: the fast gate's fail-fast probe")
	for _, source := range []string{
		"enum E { A = Math.floor(2) }",
		"enum E { A } enum E { B = 1 }",
		"declare enum E { A }",
		"declare enum E { A = 0 }",
		"enum E { A = 1 / 0 }",
		"function read(): number { return E.A; } read(); enum E { A }",
		"function read(): number { return E.A; } function call(): number { return read(); } call(); enum E { A }",
		"[1].map(() => E.A); enum E { A }",
		"function run(): void { enum E { A } }",
		"function read(): number { return E.A; } const alias = read; alias(); enum E { A }",
		"const box = { read: (): number => E.A }; box.read(); enum E { A }",
		"function read(e: E = E.A): void {} read(); enum E { A }",
		"function read(): number { return E.A; } let alias = (): number => 1; alias = read; alias(); enum E { A }",
		"function read(): number { return E.A; } let alias = (): number => 1; alias = read; [1].map(alias); enum E { A }",
		"function read(): number { return E.A; } class C { static readonly value = read(); } enum E { A }",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Fatalf("got %v, want NotYet", err)
		}
	}
	for _, source := range []string{
		"namespace N { export const value = 1; }",
		"class Box { constructor(public value: number) {} }",
		"enum E { '__proto__' = 'bad' }",
		"enum E { NaN = 0 }",
		"enum E { Infinity = 0 }",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("got %v, want Refused", err)
		}
	}
}

func TestConstEnumErasesRuntimeObject(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "const enum E { A = 1 << 2 } const value: E.A = E.A; console.log(`${value}`);")
	if err != nil {
		t.Fatal(err)
	}
	for _, local := range program.Locals {
		if local.Name == "E" {
			t.Fatal("const enum made a runtime binding")
		}
	}
	walk(program.Main, func(node any) bool {
		if _, ok := node.(ir.ObjectLiteral); ok {
			t.Fatal("const enum allocated an object")
		}
		return true
	})
}

func TestEnumIdentityAcrossModules(t *testing.T) {
	t.Parallel()
	main := filepath.Join(t.TempDir(), "main.a")
	other := filepath.Join(filepath.Dir(main), "other.a")
	for name, source := range map[string]string{main: "import { E as Other } from './other.a'; enum E { A = 'a', B = 'b' } const value: E = Other.A;", other: "export enum E { A = 'a', B = 'b' }"} {
		if err := os.WriteFile(name, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{main})
	if err != nil {
		t.Fatalf("the checker must admit the probe: %v", err)
	}
	_, err = Lower(context.Background(), program)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "enum-members") {
		t.Fatalf("got %v, want closed enum identity refusal", err)
	}
}

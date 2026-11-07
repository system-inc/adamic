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

func TestEnumSlotViews(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"variable", "enum E { A, B } function open(n: number): E { return n; }"},
		{
			"initializer",
			"enum E { A, B } function run(n: number): void { const value: E = n; console.log(`${value}`); }",
		},
		{"member initializer", "enum E { A, B } const value: E.A = 0;"},
		{
			"argument",
			"enum E { A, B } function use(e: E): void { console.log(`${e}`); } function run(n: number): void { use(n); }",
		},
		{"assignment", "enum E { A, B } let value: E = E.A; function run(n: number): void { value = n; }"},
		{
			"literal property",
			"enum E { A, B } function run(n: number): void { const box: { readonly value: E } = { value: n }; console.log(`${box.value}`); }",
		},
		{
			"array literal",
			"enum E { A, B } function run(n: number): void { const values: readonly E[] = [n]; console.log(values.join(',')); }",
		},
		{"increment", "enum E { A, B } let value: E = E.A; value++;"},
		{"compound", "enum E { A, B } let value: E = E.A; value += 7;"},
		{
			"array alias",
			"enum E { A, B } const values: E[] = [E.A]; const numbers: number[] = values; numbers.push(77);",
		},
		{
			"readonly into enum array",
			"enum E { A, B } function run(values: readonly number[]): void { const members: readonly E[] = values; console.log(members.join(',')); }",
		},
		{"function view", "enum E { A, B } const open = (n: number): number => n; const closed: (e: E) => E = open;"},
		{"object assign", "enum E { A, B } function run(n: number): void { Object.assign(E, { A: E.A }); }"},
		{
			"override",
			"enum E { A, B } class Base { read(n: number): void {} } class Derived extends Base { override read(n: E): void { console.log(`${n}`); } }",
		},
		{"class field", "enum E { A, B } class Box { value: E; constructor(n: number) { this.value = n; } }"},
		{
			"optional union",
			"enum E { A, B } function run(n: number): void { const value: E | undefined = n; console.log(`${value}`); }",
		},
		{
			"map alias",
			"enum E { A, B } const members = new Map<string, E>(); const numbers: Map<string, number> = members; numbers.set('x', 77);",
		},
		{
			"mutable field alias",
			"enum E { A, B } const box: { value: E } = { value: E.A }; const wide: { value: number } = box; wide.value = 77;",
		},
		{
			"generic class invariant",
			"enum E { A, B } class Box<T> { readonly value: T; constructor(value: T) { this.value = value; } } const box = new Box<E>(E.A); const wide: Box<number> = box;",
		},
		{"structural numeric enum object", "enum E { A, B } const copy: typeof E = { A: E.A, B: E.B };"},
		{
			"structural string enum object",
			"enum E { A = 'a', B = 'b' } const copy = { A: E.A, B: E.B, hidden: 99 } as const; const view: typeof E = copy;",
		},
		{
			"structural enum object array",
			"enum E { A, B } const copies: readonly { readonly A: E.A; readonly B: E.B }[] = [{ A: E.A, B: E.B }]; const views: readonly (typeof E)[] = copies;",
		},
		{"enum object writable view", "enum E { A, B } const view: { A: number } = E; view.A = 9;"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			// Open numeric enums do not erase generic class argument identity.
			closed := strings.Contains(probe.name, "enum object") ||
				probe.name == "object assign" ||
				probe.name == "generic class invariant"
			if !closed {
				observations := map[string]string{
					"variable":                 "console.log(`${open(7)}`);",
					"initializer":              "run(7);",
					"member initializer":       "console.log(`${value}`);",
					"argument":                 "run(7);",
					"assignment":               "run(7); console.log(`${value}`);",
					"literal property":         "run(7);",
					"array literal":            "run(7);",
					"increment":                "console.log(`${value}`);",
					"compound":                 "console.log(`${value}`);",
					"array alias":              "console.log(`${values.join(',')}`);",
					"readonly into enum array": "run([7]);",
					"function view":            "console.log(`${closed(E.B)}`);",
					"override":                 "new Derived().read(E.B);",
					"class field":              "console.log(`${new Box(7).value}`);",
					"optional union":           "run(7);",
					"map alias":                "console.log(`${members.get('x')}`);",
					"mutable field alias":      "console.log(`${box.value}`);",
				}
				lowersAndAgreesWithNode(test, probe.source+observations[probe.name])
				return
			}
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) {
				test.Fatalf("got %v, want Refused", lowerSourceError)
			}
			if probe.name == "generic class invariant" &&
				!strings.Contains(lowerSourceError.Error(), "adamic/nominal-class") {
				test.Fatalf("want nominal generic argument refusal, got %v", lowerSourceError)
			}
			if !strings.Contains(lowerSourceError.Error(), "enum") &&
				!strings.Contains(lowerSourceError.Error(), "readonly") &&
				!strings.Contains(lowerSourceError.Error(), "invariant-mutable") &&
				!strings.Contains(lowerSourceError.Error(), "override") &&
				!strings.Contains(lowerSourceError.Error(), "Object.assign") &&
				!strings.Contains(lowerSourceError.Error(), "nominal-class") {
				test.Fatalf("missing reason: %v", lowerSourceError)
			}
		})
	}
}

func TestEnumSwitchExhaustiveness(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"enum E { A = 'a', B = 'b' } function show(e: E): void { switch (e) { case E.A: break; } }",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) ||
			!strings.Contains(lowerSourceError.Error(), "non-exhaustive enum switch") {
			test.Fatalf("got %v", lowerSourceError)
		}
	}
	for _, source := range []string{
		"enum E { A, B } function show(e: E): void { switch (e) { case E.A: console.log('a'); break; } }",
		"enum E { A, B } function show(e: E): void { switch (e) { case E.A: console.log('a'); break; default: console.log('other'); break; } }",
		"enum E { A, Alias = A, B } function show(e: E): void { switch (e) { case E.A: console.log('a'); break; case E.B: console.log('b'); break; } }",
		"enum E { A, B } function show(e: E.A): void { switch (e) { case E.A: console.log('a'); break; } }",
	} {
		lowersAndAgreesWithNode(test, source+"show(E.A); console.log(`${E.A}/${E.B}`);")
	}
}

func TestEnumLimitsStayLoud(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"enum E { A = Math.floor(2) }",
		"enum E { A } enum E { B = 1 }",
		"enum E { A = 1 / 0 }",
		"function read(): number { return E.A; } read(); enum E { A }",
		"function read(): number { return E.A; } function call(): number { return read(); } call(); enum E { A }",
		"[1].map(() => E.A); enum E { A }",
		"function run(): void { enum E { A } }",
		"function read(): number { return E.A; } const alias = read; alias(); enum E { A }",
		"function read(e: E = E.A): void {} read(); enum E { A }",
		"function read(): number { return E.A; } class C { static readonly value = read(); } enum E { A }",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(lowerSourceError, &notYet) {
			test.Fatalf("got %v, want NotYet", lowerSourceError)
		}
	}
	for _, source := range []string{
		"declare enum E { A }",
		"declare enum E { A = 0 }",
		"enum E { '__proto__' = 'bad' }",
		"enum E { NaN = 0 }",
		"enum E { Infinity = 0 }",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) {
			test.Fatalf("got %v, want Refused", lowerSourceError)
		}
	}
}

func TestConstEnumErasesRuntimeObject(test *testing.T) {
	test.Parallel()
	// Behavior cannot observe an unused const-enum runtime allocation.
	program := lowersAndAgreesWithNode(
		test,
		"const enum E { A = 1 << 2 } const value: E.A = E.A; console.log(`${value}`);",
	)
	for _, local := range program.Locals {
		if local.Name == "E" {
			test.Fatal("const enum made a runtime binding")
		}
	}
	walk(program.Main, func(node any) bool {
		if _, ok := node.(ir.ObjectLiteral); ok {
			test.Fatal("const enum allocated an object")
		}
		return true
	})
}

func TestEnumIdentityAcrossModules(test *testing.T) {
	test.Parallel()
	main := filepath.Join(test.TempDir(), "main.a")
	other := filepath.Join(filepath.Dir(main), "other.a")
	for name, source := range map[string]string{
		main:  "import { E as Other } from './other.a'; enum E { A = 'a', B = 'b' } const value: E = Other.A;",
		other: "export enum E { A = 'a', B = 'b' }",
	} {
		if writeError := os.WriteFile(name, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loadError := load.Load([]string{main})
	if loadError != nil {
		test.Fatalf("the checker must admit the probe: %v", loadError)
	}
	_, lowerError := Lower(context.Background(), program)
	var refused *Refused
	if !errors.As(lowerError, &refused) || !strings.Contains(lowerError.Error(), "enum-members") {
		test.Fatalf("got %v, want closed enum identity refusal", lowerError)
	}
}

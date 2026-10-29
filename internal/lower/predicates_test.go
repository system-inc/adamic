package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestUnprovenPredicateReturnsAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"unconditional true", `function isNumber(x: number | string): x is number { return true; }`, "true return"},
		{"different variable", `function isNumber(x: number | string, other: number | string): x is number { return typeof other === 'number'; }`, "return expression"},
		{"wider narrowing", `function isOne(x: number | string): x is 1 { return typeof x === 'number'; }`, "true return narrows to number, not 1"},
		{"partial true set", `function isNumber(x: number | string): x is number { if (x === 1) { return true; } throw new Error("rest"); }`, "does not match"},
		{"narrower narrowing", `function isNumber(x: number | string): x is number { return x === 1; }`, "false return"},
		{"false branch lie", `function isNumber(x: number | string): x is number { if (typeof x === 'number') { return false; } return false; }`, "false return"},
		{"positive number", `function isNumber(x: number | string): x is number { return typeof x === 'number' && x > 0; }`, "return expression"},
		{"opaque call", `function check(x: number | string): boolean { return true; } function isNumber(x: number | string): x is number { return check(x); }`, "return expression"},
		{"modified parameter", `function isNumber(x: number | string): x is number { x = 1; return typeof x === 'number'; }`, "parameter is assigned"},
		{"captured write", `function isNumber(x: number | string): x is number { const change = (): void => { x = 1; }; change(); return typeof x === 'number'; }`, "parameter is assigned"},
		{"aliased discriminant call", `type A = { kind: 'a'; n: number }; type B = { kind: 'b'; n: number }; function change(x: A | B): void {} function isA(x: A | B): x is A { if (x.kind === 'a') { change(x); return true; } return false; }`, "true return"},
		{"aliased discriminant write", `type A = { kind: 'a'; n: number }; type B = { kind: 'b'; n: number }; function isA(x: A | B): x is A { const alias: { kind: 'a' | 'b'; n: number } = x; if (x.kind === 'a') { alias.kind = 'b'; return true; } return false; }`, "true return"},
		{"aliased discriminant getter", `type A = { kind: 'a' }; type B = { kind: 'b' }; class Reader { get value(): number { return 1; } } function isA(x: A | B, reader: Reader): x is A { if (x.kind === 'a') { reader.value; return true; } return false; }`, "true return"},
		{"nominal guard mismatch", `class A { readonly n = 1; } class B { readonly n = 1; } function isA(x: A | B): x is A { if (x instanceof B) { return true; } throw new Error('other'); }`, "true return"},
		{"nominal assertion mismatch", `class A { readonly n = 1; } class B { readonly n = 1; } function assertA(x: A | B): asserts x is A { if (!(x instanceof B)) { throw new Error('other'); } }`, "normal return"},
		{"assertion empty", `function assertNumber(x: number | string): asserts x is number {}`, "normal return"},
		{"assertion early return", `function assertNumber(x: number | string): asserts x is number { return; }`, "normal return"},
		{"assertion wrong branch", `function assertNumber(x: number | string): asserts x is number { if (typeof x === 'number') { throw new Error('wrong'); } }`, "normal return"},
		{"assertion call invalidation", `type A = { kind: 'a' }; type B = { kind: 'b' }; function change(x: A | B): void {} function assertA(x: A | B): asserts x is A { if (x.kind !== 'a') { throw new Error('bad'); } change(x); }`, "normal return"},
		{"assert cond empty", `function assertCondition(cond: boolean): asserts cond {}`, "normal return"},
		{"assert cond false", `function assertCondition(cond: boolean): asserts cond { if (cond) { throw new Error('bad'); } }`, "normal return"},
		{"assert cond nonboolean", `function assertCondition(cond: string): asserts cond {}`, "boolean parameter"},
		{"bodyless signature", `type Guard = (x: number | string) => x is number;`, "no body"},
		{"default argument", `function isNumber(x: number | string = 1): x is number { return typeof x === 'number'; }`, "unchanged plain parameter"},
		{"unknown control flow", `function isNumber(x: number | string): x is number { while (typeof x === 'number') { return true; } return false; }`, "return paths"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "return is not proven") || !strings.Contains(refusal.What, probe.reason) || !strings.Contains(refusal.Fix, "return a discriminant comparison") {
				t.Fatalf("got %v, want a predicate refusal naming %q and its rewrite", err, probe.reason)
			}
		})
	}
}

func TestPredicateBodiesAreProven(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function isNumber(x: number | string): x is number { return typeof x === 'number'; }`,
		`function isNumber(x: number | string): x is number { if (typeof x !== 'number') { return false; } return true; }`,
		`function isText(x: string | undefined): x is string { return x !== undefined; }`,
		`function isA(x: 'a' | 'b'): x is 'a' { return 'a' === x; }`,
		`function isAB(x: 'a' | 'b' | 'c'): x is 'a' | 'b' { return x === 'a' || x === 'b'; }`,
		`function isAB(x: 'a' | 'b' | 'c'): x is 'a' | 'b' { if (x === 'a') { return true; } return x === 'b'; }`,
		`function assertNumber(x: number | string): asserts x is number { if (typeof x !== 'number') { throw new Error('bad'); } }`,
		`function assertNumber(x: number | string): asserts x is number { if (typeof x === 'number') { return; } throw new Error('bad'); }`,
		`function assertCondition(cond: boolean): asserts cond { if (!cond) { throw new Error('bad'); } }`,
		`import { panic } from 'adamic'; function assertNumber(x: number | string): asserts x is number { if (typeof x !== 'number') { panic('bad'); } }`,
		`const isText = (x: string | undefined): x is string => x !== undefined;`,
	} {
		_, err := lowerSource(t, source)
		if err != nil {
			t.Errorf("%s: %v", source, err)
		}
	}
}

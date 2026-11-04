package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// The walk of enum_tag_flow.go takes a direct call to an exact kind predicate as the comparison its
// body makes (#00rd54t, Adamic's ruling, Nov 3, extending rulings.md:213), judged like that comparison
// and only where the checker's narrowing at the call is the comparison's own. Each refinement here
// needs the walk: F is a one-member enum, so A's flag is an open enum tag B lacks, and the union N
// has a hazard; kind tells A and B apart, so a comparison of kind is proven.
const kindPredicateTypes = `enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
type N = A | B;
`

const kindPredicateRefusal = "an object refinement using an open numeric enum as a literal tag"

// Each exact predicate's call lowers and agrees with Node, on the JavaScript backend and sanitized
// natively, on the side each read sits: true, false by `!`, by an early return, by `?:`, and by a
// comparison with a boolean literal. Mutant: kindPredicateStep refusing every call, so each goes back
// to the open-enum refusal (TestKindPredicateWalkRefusals' controls hold the refusal's own words).
func TestKindPredicateCallIsItsComparison(t *testing.T) {
	t.Parallel()
	const isA = "function isA(n: N): n is A { return n.kind === K.A; }\n"
	const values = "const a: A = { kind: K.A, flag: F.X, a: 1 };\nconst b: B = { kind: K.B, b: 'x' };\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"true and false", isA + "function show(n: N): string { if (isA(n)) { return `a ${n.a + 1}`; } return `b ${n.b}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 2\nb x\n"},
		{"negated", isA + "function show(n: N): string { if (!isA(n)) { return `b ${n.b}`; } return `a ${n.a}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 1\nb x\n"},
		{"conditional", isA + "function show(n: N): string { return isA(n) ? `a ${n.a}` : `b ${n.b}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 1\nb x\n"},
		{"early return", isA + "function show(n: N): string { if (isA(n)) return 'a'; return `b ${n.b}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a\nb x\n"},
		{"=== false", isA + "function show(n: N): string { if (isA(n) === false) { return `b ${n.b}`; } return `a ${n.a}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 1\nb x\n"},
		{"!== true", isA + "function show(n: N): string { if (isA(n) !== true) { return `b ${n.b}`; } return `a ${n.a}`; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 1\nb x\n"},
		{"const arrow", "const isA = (n: N): n is A => n.kind === K.A;\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\n" + values + "console.log(show(a));\nconsole.log(show(b));\n", "a 1\nb\n"},
		// Compiler's filter pair: the predicate called in a walked branch and handed to filter as a value,
		// where it stays R-T80's checked closure (this head drops no check anywhere).
		{"walked call beside filter", isA + "function show(n: N): string { return isA(n) ? `a ${n.a}` : `b ${n.b}`; }\n" + values + "console.log(show(a));\nconsole.log([a, b].filter(isA).map((n) => `${n.a}`).join(','));\n", "a 1\n1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := kindPredicateTypes + probe.source
			for _, extension := range []string{".ts", ".a"} {
				if _, err := lowerSourceAs(t, source, extension); err != nil {
					t.Fatalf("%s: want it lowered, got %v", extension, err)
				}
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// kindPredicateStep models the comparison's narrowing by hand (a member stays on the true side when
// E.M is assignable to its tag, and on the false side unless its tag is E.M itself), so these hold the
// model to the checker's own narrowing of the written comparison where they could drift (Compiler's
// should on 5d222aba): a member whose tag is a union of literals, and a const enum's member. Each
// program narrows by the predicate and by the written comparison, and both must lower and agree with
// Node; a predicate whose declared type is narrower than what its body proves is refused, fail
// closed, while the written comparison lowers. The union-of-literals row is R-T80's: its kept check
// refuses the call over that tag ("membership over an open tag domain") before the walk's model is
// asked, so the model is unreached there and fails closed; the pin moves to agreement when it lifts.
func TestKindPredicateMatchesTheWrittenComparison(t *testing.T) {
	t.Parallel()
	const unionTag = `enum K { A = 1, B = 2, C = 3 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
interface C { readonly kind: K.B | K.C; readonly c: string }
type N = A | B | C;
`
	const constEnum = `const enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
type N = A | B;
`
	both := func(predicate, otherwise string) string {
		return "function viaCall(n: N): string { if (isA(n)) { return `a ${n.a}`; } return " + otherwise + "; }\n" +
			"function viaComparison(n: N): string { if (n.kind === K.A) { return `a ${n.a}`; } return " + otherwise + "; }\n" + predicate
	}
	for _, probe := range []struct{ name, source, stdout, refusal string }{
		{"union-of-literals tag", unionTag + "function isA(n: N): n is A { return n.kind === K.A; }\nfunction viaCall(n: N): string { return isA(n) ? `a ${n.a}` : 'other'; }\nfunction viaComparison(n: N): string { return n.kind === K.A ? `a ${n.a}` : 'other'; }\n" +
			"const values: N[] = [{ kind: K.A, flag: F.X, a: 1 }, { kind: K.B, b: 'x' }, { kind: K.C, c: 'y' }];\nfor (const n of values) { console.log(viaCall(n) + ' ' + viaComparison(n)); }\n", "", "membership over an open tag domain"},
		{"const enum tag", constEnum + "function isA(n: N): n is A { return n.kind === K.A; }\n" + both("", "`b ${n.b}`") +
			"const values: N[] = [{ kind: K.A, flag: F.X, a: 1 }, { kind: K.B, b: 'x' }];\nfor (const n of values) { console.log(viaCall(n) + ' ' + viaComparison(n)); }\n", "a 1 a 1\nb x b x\n", ""},
		// isBC's body keeps B and C, its declared type only C: the call is refused, the comparison lowers.
		{"declared type narrower than the body", unionTag + "function isBC(n: N): n is C { return n.kind === K.B; }\nfunction show(n: N): string { return isBC(n) ? `c ${n.c}` : 'other'; }\nconsole.log(show({ kind: K.C, c: 'y' }));\n", "", kindPredicateRefusal},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if probe.refusal != "" {
				if _, err := lowerSourceAs(t, probe.source, ".ts"); err == nil || !strings.Contains(err.Error(), probe.refusal) {
					t.Fatalf("want %q, got %v", probe.refusal, err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// What the walk still refuses, each with the reason it gives: a call that isn't an exact predicate's,
// a predicate whose declared type disagrees with its body, and a comparison whose own hazard check
// fails, which a predicate's call meets just as the comparison does. Mutants, each named in the
// commit: condition 2's identity check dropped (the lying T and the twin layouts lower), and the
// hazard check skipped on the predicate path only (the shared value and the wide kind lower). A
// let arrow and a method are predicate values, so R-T80 refuses them first, as indirect calls.
func TestKindPredicateWalkRefusals(t *testing.T) {
	t.Parallel()
	const show = "function show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"
	for _, probe := range []struct{ name, types, source, refusal string }{
		// Controls: today's comparisons, as the walk judged them before this change.
		{"comparison control", kindPredicateTypes, "function show(n: N): string { return n.kind === K.A ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", ""},
		// The comparison's own hazard, met by the predicate's call.
		{"wide kind call", "enum K { A = 1, B = 2 }\nenum F { X = 1 }\ninterface A { readonly kind: K.A; readonly flag: F; readonly a: number }\ninterface B { readonly kind: K; readonly b: string }\ntype N = A | B;\n", "function isA(n: N): n is A { return n.kind === K.A; }\n" + show, kindPredicateRefusal},
		{"shared value call", "enum K { A = 1, Z = 5 }\nenum L { B = 1, W = 6 }\ninterface A { readonly kind: K.A; readonly a: number }\ninterface B { readonly kind: L.B; readonly b: string }\ntype N = A | B;\n", "function isA(n: N): n is A { return n.kind === K.A; }\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, a: 1 }));\n", "share a value"},
		{"shared value comparison", "enum K { A = 1, Z = 5 }\nenum L { B = 1, W = 6 }\ninterface A { readonly kind: K.A; readonly a: number }\ninterface B { readonly kind: L.B; readonly b: string }\ntype N = A | B;\n", "function show(n: N): string { return n.kind === K.A ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, a: 1 }));\n", "share a value"},
		// Condition 2: the checker narrows by the declared type, so one that disagrees proves nothing.
		{"lying declared type", kindPredicateTypes, "function isA(n: N): n is A { return n.kind === K.B; }\n" + show, kindPredicateRefusal},
		{"twin layouts", "enum K { A = 1, B = 2 }\nenum F { X = 1 }\ninterface A { readonly kind: K.A; readonly flag: F; readonly a: number }\ninterface B { readonly kind: K.A; readonly b: string }\ntype N = A | B;\n", "function isA(n: N): n is A { return n.kind === K.A; }\n" + show, kindPredicateRefusal},
		// A const whose annotation is wider than its arrow narrows by its annotation, N here, so its call
		// isn't exact and narrows nothing: a step that changes nothing, and the comparison after it proves A.
		{"annotated const wider than its arrow", kindPredicateTypes, "const isA: (n: N) => n is N = (n: N): n is A => n.kind === K.A;\nfunction show(n: N): string { return isA(n) && n.kind === K.A ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", ""},
		// A generic `n is T & A` is refused by R-T80 first, before the walk sees it: it fails closed.
		{"generic, reduced", kindPredicateTypes, "function isA<T extends N>(n: T): n is T & A { return n.kind === K.A; }\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : `b ${n.b}`; }\nconst a: A = { kind: K.A, flag: F.X, a: 1 };\nconst b: B = { kind: K.B, b: 'x' };\nconsole.log(show(a));\nconsole.log(show(b));\n", "checked predicate overload target never"},
		// A `this: void` predicate stops in signature() first (#j8mt4pr, which closed main's shifted arguments).
		// exactKindCall reads the first parameter after `this`, as the ruling says, for the day it lowers.
		{"this parameter", kindPredicateTypes, "function isA(this: void, n: N): n is A { return n.kind === K.A; }\n" + show, "a function or method with a this parameter followed by another parameter"},
		// Not exact, so the call is any other step.
		{"inferred predicate", kindPredicateTypes, "function isA(n: N) { return n.kind === K.A; }\n" + show, kindPredicateRefusal},
		{"overloaded", kindPredicateTypes, "function isA(n: A): n is A;\nfunction isA(n: N): n is A;\nfunction isA(n: N): n is A { return n.kind === K.A; }\n" + show, kindPredicateRefusal},
		{"method", kindPredicateTypes, "const kinds = { isA(n: N): n is A { return n.kind === K.A; } };\nfunction show(n: N): string { return kinds.isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", "an indirect call of a checked predicate"},
		{"value alias", kindPredicateTypes, "function isAlpha(n: N): n is A { return n.kind === K.A; }\nconst isA = isAlpha;\n" + show, kindPredicateRefusal},
		{"let arrow", kindPredicateTypes, "let isA = (n: N): n is A => n.kind === K.A;\n" + show, "an indirect call of a checked predicate"},
		// A default or an optional mark: isA(undefined) would test the default's kind, not the argument's
		// (Language's attack on 5d222aba).
		{"parameter with a default", kindPredicateTypes, "const fallback: N = { kind: K.A, flag: F.X, a: 0 };\nfunction isA(n: N = fallback): n is A { return n.kind === K.A; }\nfunction show(n: N | undefined): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", kindPredicateRefusal},
		{"optional parameter", kindPredicateTypes, "function isA(n?: N): n is A { return n !== undefined && n.kind === K.A; }\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", kindPredicateRefusal},
		{"two statements", kindPredicateTypes, "function isA(n: N): n is A { const k = n.kind; return k === K.A; }\n" + show, kindPredicateRefusal},
		{"asserts", kindPredicateTypes, "function assertA(n: N): asserts n is A { if (n.kind !== K.A) { throw new Error('not A'); } }\nfunction show(n: N): string { assertA(n); return `a ${n.a}`; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", kindPredicateRefusal},
		{"switch (true)", kindPredicateTypes, "function isA(n: N): n is A { return n.kind === K.A; }\nfunction show(n: N): string { switch (true) { case isA(n): return `a ${n.a}`; } return 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", kindPredicateRefusal},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.types+probe.source, ".ts")
			if probe.refusal == "" {
				if err != nil {
					t.Fatalf("want it lowered, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("want %q, got %v", probe.refusal, err)
			}
		})
	}
}

// An import under another name and a re-export under another name both resolve to the one
// declaration, so they count; a call through a namespace import is a property access and doesn't.
func TestKindPredicateThroughModules(t *testing.T) {
	t.Parallel()
	library := kindPredicateTypes + "export function isA(n: N): n is A { return n.kind === K.A; }\nexport { K, F };\nexport type { A, B, N };\n"
	for _, probe := range []struct{ name, main, refusal string }{
		{"import as", "import { isA as isAlpha, K, F, type N } from './library';\nfunction show(n: N): string { return isAlpha(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", ""},
		{"re-export as", "import { isBeta, K, F, type N } from './relay';\nfunction show(n: N): string { return isBeta(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n", ""},
		{"namespace import", "import * as library from './library';\nfunction show(n: library.N): string { return library.isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: library.K.A, flag: library.F.X, a: 1 }));\n", kindPredicateRefusal},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for name, source := range map[string]string{
				"library.ts": library,
				"relay.ts":   "export { isA as isBeta, K, F } from './library';\nexport type { N } from './library';\n",
				"main.ts":    probe.main,
			} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), program)
			if probe.refusal == "" {
				if err != nil {
					t.Fatalf("want it lowered, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("want %q, got %v", probe.refusal, err)
			}
		})
	}
}

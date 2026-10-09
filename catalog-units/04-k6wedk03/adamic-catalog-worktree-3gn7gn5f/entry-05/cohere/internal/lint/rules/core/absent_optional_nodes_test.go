package core

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every rule in this package is run over shapes where a node it reaches for is legitimately absent.
//
// Ported from the structure package's guard, which exists because a shipped rule crashed on a nil
// that review had read past twice. `ast.SkipParentheses` dereferences its argument, so a nil check
// placed after the call rather than before it panics, and the shapes that produce a nil are
// ordinary rather than exotic.
//
// A crash is worse than a wrong finding. A wrong finding is visible and arguable; a panic takes
// down the walk for the whole file, so every other rule's verdict on that file is lost with it, and
// on a 3,407-file tree the file that crashed is one line in a log nobody reads. It fails in the
// direction of silence.
//
// This package earns its own copy: it holds eleven `SkipParentheses` call sites across five rules,
// more than anywhere else in the tree, and its rules reach for initializers, arguments, discriminants
// and catch clauses that need not exist.
//
// # The one call site that is safe only by an invariant
//
// `prefer_spread.go` passes `memberAccessObject(callee)` straight into `SkipParentheses`, and that
// helper returns nil for a callee that is neither a property nor an element access. It cannot fire,
// because an `isApplyMemberAccess` gate fifteen lines earlier has already excluded every other kind.
// Widening that gate reintroduces a panic at a line the change does not touch, and this guard is
// what would catch it. The comment there says the same thing at the site.
//
// # What a pass here does and does not mean
//
// The shape list is the measurement and it is short. A green run means every listed rule survived
// these shapes. It is not a claim that the package cannot panic, and the two are indistinguishable
// from here.
//
// A probe that does not reach a line proves nothing about the line, only about the inputs. When a
// new rule reaches for something optional in a shape not listed, the fix is to add the shape rather
// than to trust the green.
func TestNoRuleCrashesOnAbsentOptionalNodes(t *testing.T) {
	t.Parallel()

	sources := []string{
		// Declarations with no initializer, which is where a rule reaching for one finds nil.
		"declare const absent: unknown;\nexport const value = 1;\n",
		"let uninitialized;\nexport const value = uninitialized;\n",
		"declare function apply(a: unknown, b: unknown[]): void;\nexport const r = apply(null, []);\n",
		// Calls with no arguments, for rules indexing into an argument list.
		"declare const f: { apply(): void };\nexport const r = f.apply();\n",
		"export const r = [].length;\n",
		// An object shorthand property, which has no explicit value node.
		"declare const shorthand: unknown;\nexport const o = { shorthand };\n",
		// A switch with no clauses, and a case with no statements.
		"declare const d: number;\nexport function f() { switch (d) {} }\n",
		"declare const d: number;\nexport function f() { switch (d) { case 1: } }\n",
		// try/catch/finally with empty and parameterless forms.
		"export function f() { try {} catch {} }\n",
		"export function f() { try { g(); } finally {} }\ndeclare function g(): void;\n",
		// A generator and an async function with empty bodies.
		"export function* gen() {}\n",
		"export async function run() {}\n",
		// Sparse and empty array and object patterns.
		"export const [, , third] = [1, 2, 3];\n",
		"export function f({}: Record<string, unknown>) { return 1; }\n",
		// A delete on a bare identifier and a typeof against a non-literal.
		"declare const o: any;\nexport function f() { return delete o; }\n",
		"declare const v: unknown;\ndeclare const s: string;\nexport const r = typeof v === s;\n",
	}

	rules := []rule.Rule{
		NoCaseDeclarations,
		NoCompareNegZero,
		NoDebugger,
		NoDeleteVar,
		NoDuplicateCase,
		NoEmpty,
		NoEmptyPattern,
		NoEmptyStaticBlock,
		NoExAssign,
		NoSparseArrays,
		NoUnsafeFinally,
		NoVar,
		NoUselessCatch,
		PreferSpread,
		RequireYield,
		UseIsNaN,
		ValidTypeof,
	}

	for _, currentRule := range rules {
		for index, source := range sources {
			t.Run(fmt.Sprintf("%s/shape-%d", currentRule.Name, index), func(t *testing.T) {
				t.Parallel()
				// A panic fails the test. The assertion is only that this returns at all, so
				// findings are deliberately not checked: what each rule concludes about these
				// shapes belongs in its own pair, and asserting it here would make this guard
				// fail for reasons that are not crashes.
				rule_testing.Run(t, currentRule, "shapes.ts", source)
			})
		}
	}
}

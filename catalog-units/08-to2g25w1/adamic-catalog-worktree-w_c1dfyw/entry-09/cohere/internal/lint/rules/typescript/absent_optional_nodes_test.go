package typescript

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every rule in this package is run over shapes where a node it reaches for is legitimately absent,
// or where the construct it keys on appears in an unusual position.
//
// A crash is worse than a wrong finding. A wrong finding is visible and arguable; a panic takes down
// the walk for the whole file, so every other rule's verdict on that file is lost with it, and on a
// 3,407-file tree the file that crashed is one line in a log nobody reads. It fails in the direction
// of silence.
//
// # Why this package earns its own copy
//
// The non-null family walks *up* from the node it visits rather than only down. Three of the guards
// here read `node.Parent`, one walks a chain of parents through parentheses, and `isAssignmentTarget`
// recurses upward through five node kinds. Upward walks fail differently from downward ones: a
// missing child is a nil the rule can see coming, while a missing parent is the top of the file
// arriving earlier than the walk expected.
//
// The shapes below are chosen for that: a non-null assertion at the outermost position of a
// statement, inside a decorator, inside a default parameter, at the top of an expression body. Each
// puts an assertion somewhere its parent chain terminates in a kind the guards were not written
// against.
//
// Also included: `ast.SkipParentheses` dereferences its argument, so a nil check placed after the
// call rather than before it panics. `no-non-null-asserted-optional-chain` calls it, and the shapes
// with an empty or unusual inner expression are what would reach a missing one.
//
// # What a pass here does and does not mean
//
// The shape list is the measurement and it is short. A green run means every listed rule survived
// these shapes. It is not a claim that the package cannot panic, and the two are indistinguishable
// from here. When a new rule reaches for something optional in a shape not listed, the fix is to add
// the shape rather than to trust the green.
func TestNoRuleCrashesOnAbsentOptionalNodes(t *testing.T) {
	t.Parallel()

	sources := []string{
		// An assertion as an entire expression statement, so its parent chain is one link deep.
		"declare const a: any;\na!;\n",
		"declare const a: any;\n(a!);\n",
		"declare const a: any;\n((a!));\n",
		// An assertion at the top of a return and an arrow body, where the parent is not an
		// expression at all.
		"declare const a: any;\nexport function get() {\n    return a!;\n}\n",
		"declare const a: any;\nexport const get = () => a!;\n",
		// An assertion in a default parameter value and an initializer, both positions where the
		// parent is a declaration rather than an expression.
		"declare const a: any;\nexport function f(x = a!) {\n    return x;\n}\n",
		"declare const a: any;\nexport const value = a!;\n",
		// Assertions inside constructs whose parents are clauses rather than expressions.
		"declare const a: any;\nexport function f() {\n    if (a!) { return 1; }\n    return 0;\n}\n",
		"declare const a: any;\nexport function f() {\n    switch (a!) { }\n    return 0;\n}\n",
		"declare const a: any;\nexport function f() {\n    for (const x of a!) { void x; }\n}\n",
		"declare const a: any;\nexport function f() {\n    try { void a!; } catch {}\n}\n",
		// Optional chains with nothing following, and chains whose links are empty of arguments.
		"declare const a: any;\nexport const r = a?.b;\n",
		"declare const a: any;\nexport const r = a?.();\n",
		"declare const a: any;\nexport const r = a?.()!;\n",
		"declare const a: any;\nexport const r = (a?.())!;\n",
		// Deeply parenthesized assertions, which is what the upward paren walk is for. If it were
		// written as a single step rather than a loop, these are what would reach past it.
		"declare const a: any;\nexport const r = (((a!)))!;\n",
		"declare const a: any;\nexport const r = (((a?.b)))!;\n",
		// The assignment-target recursion, at shapes where it terminates in something unusual.
		"declare const a: any;\n[a!.b] = [0];\n",
		"declare const a: any;\n[, a!.b] = [0, 1];\n",
		"declare const a: any;\n({ ...a!.b } = { });\n",
		"declare const a: any;\n(a!.b as unknown as number)++;\n",
		// A definite assignment with no initializer, which carries the same token and no node.
		"export let value!: number;\n",
		"export class Thing {\n    value!: number;\n}\n",
		// An assertion in a type position's neighborhood, and a bare optional chain in a template.
		"declare const a: any;\nexport const r = `${a!}`;\n",
		"declare const a: any;\nexport const r = `${a?.b}`;\n",
	}

	rules := []rule.Rule{
		NoExtraNonNullAssertion,
		NoNonNullAssertedOptionalChain,
		NoNonNullAssertion,
		NoUnsafeFunctionType,
	}

	for _, currentRule := range rules {
		for index, source := range sources {
			t.Run(fmt.Sprintf("%s/shape-%d", currentRule.Name, index), func(t *testing.T) {
				t.Parallel()
				// A panic fails the test. The assertion is only that this returns at all, so
				// findings are deliberately not checked: what each rule concludes about these
				// shapes belongs in its own pair, and asserting it here would make this guard fail
				// for reasons that are not crashes.
				rule_testing.Run(t, currentRule, "shapes.ts", source)
			})
		}
	}
}

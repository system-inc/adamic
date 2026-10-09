package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// prototypeBuiltinsFile is where the fixtures pretend to live.
const prototypeBuiltinsFile = "/repository/source/PrototypeBuiltins.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_prototype_builtins.rs`:
// 27 pass and 20 fail in a single tester block, and the snapshot records 20 diagnostics from those
// 20 fail inputs, so one finding per input is measured here rather than assumed. Copied because a
// fixture a porter invents encodes the same belief as the port, and the case that catches a bug is
// the one nobody would think to write.
func TestNoPrototypeBuiltinsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// One per disallowed name, in the plain dotted spelling. Covering a subset of the three
		// silently is the failure shape this rule invites, so each name is pinned in both
		// directions rather than trusting one to stand for the others.
		{"hasOwnProperty on a receiver", "foo.hasOwnProperty('bar')"},
		{"isPrototypeOf on a receiver", "foo.isPrototypeOf('bar')"},
		{"propertyIsEnumerable on a receiver", "foo.propertyIsEnumerable('bar')"},
		// The receiver may be any expression, not just an identifier.
		{"a nested receiver", "foo.bar.hasOwnProperty('bar')"},
		{"a twice nested receiver", "foo.bar.baz.isPrototypeOf('bar')"},
		// Subscript spellings resolve to the same property, so they report the same.
		{"a string subscript", "foo['hasOwnProperty']('bar')"},
		{"a template subscript, result then read", "foo[`isPrototypeOf`]('bar').baz"},
		{"a string subscript on a nested receiver", `foo.bar["propertyIsEnumerable"]('baz')`},
		// Shadowing `Object` changes nothing about the judgment: the rule never reads `Object`.
		// It only mattered upstream for whether the suggestion could be offered, and we offer none.
		{"inside a function shadowing Object", "(function(Object) {return foo.hasOwnProperty('bar');})"},
		{"with the Object global disabled", "foo.hasOwnProperty('bar')"},
		// The four optional spellings. These are the cases that hurt the neighbouring rules, and
		// each reaches the callee by a different route, so all four are kept separately.
		{"an optional receiver", "foo?.hasOwnProperty('bar')"},
		{"an optional link earlier in the chain", "foo?.bar.hasOwnProperty('baz')"},
		{"an optional call", "foo.hasOwnProperty?.('bar')"},
		{"an optional receiver, result then read", "foo?.hasOwnProperty('bar').baz"},
		{"an optional read of the result", "foo.hasOwnProperty('bar')?.baz"},
		// A sequence expression as receiver, which is why parentheses have to be seen through.
		{"a sequence expression receiver", "(a,b).hasOwnProperty('bar')"},
		// ESLint calls these ChainExpression callees and declines to repair them. We report them
		// like any other call; `ast.SkipParentheses` is what reaches the member expression.
		{"a parenthesized optional callee", "(foo?.hasOwnProperty)('bar')"},
		{"a parenthesized optional callee, optionally called", "(foo?.hasOwnProperty)?.('bar')"},
		{"an optional string subscript", "foo?.['hasOwnProperty']('bar')"},
		{"a parenthesized optional template subscript", "(foo?.[`hasOwnProperty`])('bar')"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile, testCase.sourceText),
				"noPrototypeBuiltins")
		})
	}
}

// The clean cases are the whole discrimination, and they fail in five distinct ways.
//
// Calling through the prototype is the fix the rule asks for, so it must stay silent. An access
// with no call is not a call. A dynamic key is not knowably this property. A key that merely
// resembles one of the three names is a different property. And a private field spelling the same
// word cannot reach the builtin at all.
func TestNoPrototypeBuiltinsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The repair the rule is asking for, one per name, called and applied.
		{"hasOwnProperty through the prototype", "Object.prototype.hasOwnProperty.call(foo, 'bar')"},
		{"isPrototypeOf through the prototype", "Object.prototype.isPrototypeOf.call(foo, 'bar')"},
		{"propertyIsEnumerable through the prototype", "Object.prototype.propertyIsEnumerable.call(foo, 'bar')"},
		{"hasOwnProperty applied through the prototype", "Object.prototype.hasOwnProperty.apply(foo, ['bar'])"},
		{"isPrototypeOf applied through the prototype", "Object.prototype.isPrototypeOf.apply(foo, ['bar'])"},
		{"propertyIsEnumerable applied through the prototype", "Object.prototype.propertyIsEnumerable.apply(foo, ['bar'])"},
		// An access is not a call, and the call that matters is the outermost property.
		{"a bare access with no call", "foo.hasOwnProperty"},
		{"a call on a different property beyond it", "foo.hasOwnProperty.bar()"},
		// The name appearing as a value or as a callee is not a property access at all.
		{"the name passed as an argument", "foo(hasOwnProperty)"},
		{"hasOwnProperty called as a free function", "hasOwnProperty(foo, 'bar')"},
		{"isPrototypeOf called as a free function", "isPrototypeOf(foo, 'bar')"},
		{"propertyIsEnumerable called as a free function", "propertyIsEnumerable(foo, 'bar')"},
		// The empty-object-literal spelling of the same repair.
		{"hasOwnProperty through an object literal", "({}.hasOwnProperty.call(foo, 'bar'))"},
		{"isPrototypeOf through an object literal", "({}.isPrototypeOf.call(foo, 'bar'))"},
		{"propertyIsEnumerable through an object literal", "({}.propertyIsEnumerable.call(foo, 'bar'))"},
		{"hasOwnProperty applied through an object literal", "({}.hasOwnProperty.apply(foo, ['bar']))"},
		{"isPrototypeOf applied through an object literal", "({}.isPrototypeOf.apply(foo, ['bar']))"},
		{"propertyIsEnumerable applied through an object literal", "({}.propertyIsEnumerable.apply(foo, ['bar']))"},
		// A computed key read from a variable is not knowable.
		{"a subscript through a variable", "foo[hasOwnProperty]('bar')"},
		// Casing and spelling: near misses are different properties.
		{"a differently cased key", "foo['HasOwnProperty']('bar')"},
		{"a template key with a trailing letter", "foo[`isPrototypeOff`]('bar')"},
		{"an optional subscript with a truncated key", "foo?.['propertyIsEnumerabl']('bar')"},
		// Non-string keys resolve to names that are not in the set.
		{"a numeric subscript", "foo[1]('bar')"},
		{"a null subscript", "foo[null]('bar')"},
		// A private field is a different property that merely spells the same word.
		{"a private field spelling the same name", "class C { #hasOwnProperty; foo() { obj.#hasOwnProperty('bar'); } }"},
		// A key computed at runtime is not statically knowable, even when its value would match.
		{"a concatenated key", "foo['hasOwn' + 'Property']('bar')"},
		{"a substituting template key", "foo[`hasOwnProperty${''}`]('bar')"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile, testCase.sourceText))
		})
	}
}

// Cases written from reading our code rather than upstream's.
//
// The imported corpus exercises no TypeScript at all, and this rule runs over a TypeScript tree.
// A non-null assertion and a type assertion both wrap the receiver in a node ESTree has no analogue
// for, and a port reaching for the callee without seeing through the wrapper would miss the first
// and report on the wrong thing for the rest. `this` and a call-returned receiver are the shapes
// our own code most plausibly writes.
func TestNoPrototypeBuiltinsHandlesTypeScriptReceivers(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
	}{
		{"a non-null asserted receiver", "declare const foo: object | null;\nfoo!.hasOwnProperty('bar');\n"},
		{"an as-asserted receiver", "declare const foo: unknown;\n(foo as object).hasOwnProperty('bar');\n"},
		{"a this receiver", "class C { foo() { return this.hasOwnProperty('bar'); } }"},
		{"a call-returned receiver", "declare function make(): object;\nmake().isPrototypeOf(bar);\n"},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile, testCase.sourceText),
				"noPrototypeBuiltins")
		})
	}

	// A method *definition* named for one of the three is a declaration, not a call through a
	// receiver, and defining `hasOwnProperty` on your own class is how you would legitimately
	// provide it. Nothing in the imported corpus covers a class body.
	t.Run("a class declaring the method itself", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile,
			"class C { hasOwnProperty(key: string) { return false; } }"))
	})

	// Two receivers in one file report twice. Nothing upstream carries more than one finding per
	// input, so the rule's per-call behavior is otherwise unmeasured.
	t.Run("two calls in one file report twice", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile,
			"foo.hasOwnProperty('a');\nbar.isPrototypeOf(baz);\n"),
			"noPrototypeBuiltins", "noPrototypeBuiltins")
	})
}

// The span, which the message-id fixtures above cannot see.
//
// This rule carries no repair, and that is exactly when a span defect ships unnoticed: every
// assertion above checks which message fired and none checks where it points. A rule reporting the
// whole call, or the receiver, or the argument would pass all 47 verbatim fixtures while pointing
// somewhere useless. Sliced out of the source with the finding's own range so the assertion cannot
// agree with the code by sharing its arithmetic.
func TestNoPrototypeBuiltinsPointsAtTheMemberExpression(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		// The member expression, which excludes the argument list and includes the receiver.
		{"a dotted call", "foo.hasOwnProperty('bar')", "foo.hasOwnProperty"},
		// A nested receiver is part of the member expression, so the span grows to the left.
		{"a nested receiver", "foo.bar.baz.isPrototypeOf('bar')", "foo.bar.baz.isPrototypeOf"},
		// A subscript's brackets and quotes are inside the span.
		{"a string subscript", "foo['hasOwnProperty']('bar')", "foo['hasOwnProperty']"},
		// The `?.` belongs to the member expression and is inside the span.
		{"an optional receiver", "foo?.hasOwnProperty('bar')", "foo?.hasOwnProperty"},
		// An optional *call* puts its `?.` after the member expression, so the span stops short of
		// it. This is the pair that distinguishes reporting the callee from reporting the call.
		{"an optional call", "foo.hasOwnProperty?.('bar')", "foo.hasOwnProperty"},
		// Trailing reads of the result are outside the span entirely.
		{"a read of the result", "foo.hasOwnProperty('bar')?.baz", "foo.hasOwnProperty"},
		// Parentheses are skipped to find the callee, so the span is the inner member expression
		// and excludes the parentheses themselves.
		{"a parenthesized optional callee", "(foo?.hasOwnProperty)('bar')", "foo?.hasOwnProperty"},
		// A sequence receiver keeps its own parentheses, which are inside the member expression.
		{"a sequence receiver", "(a,b).hasOwnProperty('bar')", "(a,b).hasOwnProperty"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Fatalf("finding pointed at %q, wanted %q", reported, testCase.wantText)
			}
		})
	}
}

// The rule declares no type checker, so the untyped harness must be enough.
//
// If a later change reaches for `ctx.TypeChecker`, the rule goes silent under `rule_testing.Run` and
// every clean case above starts passing vacuously. This asserts the plain harness still fires.
func TestNoPrototypeBuiltinsNeedsNoTypeChecker(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoPrototypeBuiltins, prototypeBuiltinsFile, "foo.hasOwnProperty('bar')"),
		"noPrototypeBuiltins")
}

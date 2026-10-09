package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// objectConstructorFile is where the fixtures pretend to live.
const objectConstructorFile = "/repository/source/ObjectConstructor.ts"

// The corpus is ESLint's own, extracted rather than retyped.
//
// `eslint/tests/lib/rules/no-object-constructor.js` at eslint 10.8.1 carries 6 valid and 50 invalid.
// Extracted by executing that file against a stub RuleTester, which matters more here than anywhere
// else in this batch: 44 of the 50 invalid cases are BUILT by a `.map()` over a list of code
// snippets, with the expected `output` computed by a regular expression at test-definition time. No
// reader could transcribe those correctly and none of them appear as literals in the source file.
//
// Each case carries its own suggestion messageId and its own output, so the split between the two
// suggestion arms is upstream's assertion rather than something recovered. 38 take `useLiteral` and
// 12 take `useLiteralAfterSemicolon`.
//
// All 56 were additionally driven through the installed build and agreed with the file.
//
// One valid case is NOT imported: `new Object()` with `globals: { Object: "off" }`. That case turns
// a global off through ESLint's own configuration surface, which has no analogue here, and it is
// clean upstream for a reason that is not about the rule. It is recorded here rather than silently
// dropped.
func TestNoObjectConstructorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		suggestionId string
		wantAfterFix string
	}{
		{"new Object", "useLiteral", "({})"},
		{"Object()", "useLiteral", "({})"},
		{"const fn = () => Object();", "useLiteral", "const fn = () => ({});"},
		{"Object() instanceof Object;", "useLiteral", "({}) instanceof Object;"},
		{"const obj = Object?.();", "useLiteral", "const obj = {};"},
		{"(new Object() instanceof Object);", "useLiteral", "({} instanceof Object);"},
		{"\n                foo\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                foo\n                ;({})\n                "},
		{"\n                foo()\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                foo()\n                ;({})\n                "},
		{"\n                new foo\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                new foo\n                ;({})\n                "},
		{"\n                (a++)\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                (a++)\n                ;({})\n                "},
		{"\n                ++a\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                ++a\n                ;({})\n                "},
		{"\n                const foo = function() {}\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                const foo = function() {}\n                ;({})\n                "},
		{"\n                const foo = class {}\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                const foo = class {}\n                ;({})\n                "},
		{"\n                foo = this.return\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                foo = this.return\n                ;({})\n                "},
		{"\n                var yield = bar.yield\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                var yield = bar.yield\n                ;({})\n                "},
		{"\n                var foo = { bar: baz }\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                var foo = { bar: baz }\n                ;({})\n                "},
		{"\n                <foo />\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                <foo />\n                ;({})\n                "},
		{"\n                <foo></foo>\n                Object()\n                ", "useLiteralAfterSemicolon", "\n                <foo></foo>\n                ;({})\n                "},
		{"\n                {}\n                Object()\n                ", "useLiteral", "\n                {}\n                ({})\n                "},
		{"\n                function foo() {}\n                Object()\n                ", "useLiteral", "\n                function foo() {}\n                ({})\n                "},
		{"\n                class Foo {}\n                Object()\n                ", "useLiteral", "\n                class Foo {}\n                ({})\n                "},
		{"foo: Object();", "useLiteral", "foo: ({});"},
		{"foo();Object();", "useLiteral", "foo();({});"},
		{"{ Object(); }", "useLiteral", "{ ({}); }"},
		{"if (a) Object();", "useLiteral", "if (a) ({});"},
		{"if (a); else Object();", "useLiteral", "if (a); else ({});"},
		{"while (a) Object();", "useLiteral", "while (a) ({});"},
		{"\n                do Object();\n                while (a);\n                ", "useLiteral", "\n                do ({});\n                while (a);\n                "},
		{"for (let i = 0; i < 10; i++) Object();", "useLiteral", "for (let i = 0; i < 10; i++) ({});"},
		{"for (const prop in obj) Object();", "useLiteral", "for (const prop in obj) ({});"},
		{"for (const element of iterable) Object();", "useLiteral", "for (const element of iterable) ({});"},
		{"\n                const foo = () => {}\n                Object()\n                ", "useLiteral", "\n                const foo = () => {}\n                ({})\n                "},
		{"\n                a++\n                Object()\n                ", "useLiteral", "\n                a++\n                ({})\n                "},
		{"\n                a--\n                Object()\n                ", "useLiteral", "\n                a--\n                ({})\n                "},
		{"\n                function foo() {\n                    return\n                    Object();\n                }\n                ", "useLiteral", "\n                function foo() {\n                    return\n                    ({});\n                }\n                "},
		{"\n                function * foo() {\n                    yield\n                    Object();\n                }\n                ", "useLiteral", "\n                function * foo() {\n                    yield\n                    ({});\n                }\n                "},
		{"\n                do {}\n                while (a)\n                Object()\n                ", "useLiteral", "\n                do {}\n                while (a)\n                ({})\n                "},
		{"\n                debugger\n                Object()\n                ", "useLiteral", "\n                debugger\n                ({})\n                "},
		{"\n                for (;;) {\n                    break\n                    Object()\n                }\n                ", "useLiteral", "\n                for (;;) {\n                    break\n                    ({})\n                }\n                "},
		{"\n                for (;;) {\n                    continue\n                    Object()\n                }\n                ", "useLiteral", "\n                for (;;) {\n                    continue\n                    ({})\n                }\n                "},
		{"\n                foo: break foo\n                Object()\n                ", "useLiteral", "\n                foo: break foo\n                ({})\n                "},
		{"\n                foo: while (true) continue foo\n                Object()\n                ", "useLiteral", "\n                foo: while (true) continue foo\n                ({})\n                "},
		{"\n                const foo = bar\n                export { foo }\n                Object()\n                ", "useLiteral", "\n                const foo = bar\n                export { foo }\n                ({})\n                "},
		{"\n                export { foo } from 'bar'\n                Object()\n                ", "useLiteral", "\n                export { foo } from 'bar'\n                ({})\n                "},
		{"\n                export * as foo from 'bar'\n                Object()\n                ", "useLiteral", "\n                export * as foo from 'bar'\n                ({})\n                "},
		{"\n                import foo from 'bar'\n                Object()\n                ", "useLiteral", "\n                import foo from 'bar'\n                ({})\n                "},
		{"\n                var yield = 5;\n\n                yield: while (foo) {\n                    if (bar)\n                        break yield\n                    new Object();\n                }\n                ", "useLiteral", "\n                var yield = 5;\n\n                yield: while (foo) {\n                    if (bar)\n                        break yield\n                    ({});\n                }\n                "},
		{"\n                var foo\n                Object()\n                ", "useLiteral", "\n                var foo\n                ({})\n                "},
		{"\n                let bar\n                Object()\n                ", "useLiteral", "\n                let bar\n                ({})\n                "},
	}

	for _, testCase := range cases {
		t.Run(strings.Join(strings.Fields(testCase.sourceText), " "), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile,
				testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if diagnostic.Message.Id != "preferLiteral" {
				t.Fatalf("message id was %q", diagnostic.Message.Id)
			}

			// Which suggestion arm fired. The two differ only by a leading semicolon in the
			// replacement, and applying the wrong one produces source that parses as a call.
			if len(diagnostic.Suggestions) != 1 {
				t.Fatalf("wanted one suggestion, got %d", len(diagnostic.Suggestions))
			}
			suggestion := diagnostic.Suggestions[0]
			if suggestion.Message.Id != testCase.suggestionId {
				t.Fatalf("suggestion id was %q, wanted %q", suggestion.Message.Id, testCase.suggestionId)
			}

			// The repair itself, asserted by APPLYING it and comparing the whole resulting source
			// rather than by comparing the fix text: a fix writing the right string over the wrong
			// span passes a text comparison and is a real defect.
			if len(suggestion.Fixes) != 1 {
				t.Fatalf("wanted one fix, got %d", len(suggestion.Fixes))
			}
			fix := suggestion.Fixes[0]
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			rewritten := onDisk[:fix.Range.Pos()] + fix.Text + onDisk[fix.Range.End():]
			wantOnDisk := strings.TrimSpace(testCase.wantAfterFix) + "\n"
			if rewritten != wantOnDisk {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, wantOnDisk)
			}
		})
	}
}

// The clean cases carry the whole judgment.
//
// `Object(x)` and `new Object(x)` are clean because the one-argument form is a different operation:
// it returns its argument boxed rather than a new object. `new globalThis.Object` is clean because
// the callee is a member access rather than the identifier, which is narrower than "evaluates to
// Object" and is upstream's own reading. The last two are a shadowed `Object`, in a parameter and in
// a `var`.
func TestNoObjectConstructorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"new Object(x)"},
		{"Object(x)"},
		{"new globalThis.Object"},
		{"const createObject = Object => new Object()"},
		{"var Object; new Object;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjectConstructor,
				objectConstructorFile, testCase.sourceText))
		})
	}
}

// A parenthesized callee is CLEAN, which is upstream's reading and looks like an omission.
//
// Upstream tests `node.callee.type !== "Identifier"` with no parenthesis skipping, and a
// parenthesized callee is a different node in ESTree as well. Measured on eslint 10.8.1: `(Object)()`
// reports nothing. Every other rule in this batch skips parentheses, so this is written down rather
// than left for the next reader to "fix".
func TestNoObjectConstructorDeclinesAParenthesizedCallee(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile,
		"(Object)();"))
}

// The typed harness is required, and a revert to the plain one must fail loudly.
func TestNoObjectConstructorNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "Object();"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile,
		source), "preferLiteral")

	if got := len(rule_testing.Run(t, NoObjectConstructor, objectConstructorFile,
		source).Diagnostics); got != 0 {
		t.Fatalf("the untyped harness produced %d diagnostics, so the nil-checker guard has moved "+
			"and this test no longer measures what it claims", got)
	}
}

// The message, asserted against literals typed here rather than the rule's own constants.
func TestNoObjectConstructorReportsWhyItMatters(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile, "Object();")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	const wantPrefix = "This calls the `Object` constructor with no argument"
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("description was %q", got)
	}
}

// The one divergence from upstream in this rule, kept visible rather than dropped.
//
// `with (obj) Object();` is an upstream REPORTING case and is silent here. The checker returns no
// symbol for `Object` inside a `with` block, because whether that name reaches the global genuinely
// depends on what `obj` holds at runtime, so `resolvesToAGlobal` declines and the rule declines with
// it. See the rule's doc comment for the measurement.
//
// Asserted as silence so that a later change making it report fails loudly and gets re-argued,
// rather than sliding in as an improvement nobody measured.
func TestNoObjectConstructorDeclinesInsideAWithBlock(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile,
		"with (obj) Object();"))
}

// Where the FINDING points, which the fix assertions above cannot see.
//
// Every reporting fixture in this file asserts the applied repair, and a repair is anchored on its
// own range rather than on the diagnostic's, so a rule reporting the callee while fixing the call
// passes all 49 of them. A mutation doing exactly that survived, which is how this test came to
// exist.
//
// The span is the whole call, measured on eslint 10.8.1. The optional row is the one that separates
// the two readings most clearly: `const obj = Object?.();` reports columns 13 through 23, which
// covers `Object?.()` and not the six characters of the callee.
func TestNoObjectConstructorReportsTheWholeCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPos    int
		wantEnd    int
	}{
		{"new Object()", 0, 12},
		{"Object();", 0, 8},
		{"new Object", 0, 10},
		{"const obj = Object?.();", 12, 22},
		{"(new Object() instanceof Object);", 1, 13},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoObjectConstructor, objectConstructorFile,
				testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			got := result.Diagnostics[0].Range
			if got.Pos() != testCase.wantPos || got.End() != testCase.wantEnd {
				t.Fatalf("reported [%d:%d) which is %q, wanted [%d:%d) which is %q",
					got.Pos(), got.End(), onDisk[got.Pos():got.End()],
					testCase.wantPos, testCase.wantEnd, onDisk[testCase.wantPos:testCase.wantEnd])
			}
		})
	}
}

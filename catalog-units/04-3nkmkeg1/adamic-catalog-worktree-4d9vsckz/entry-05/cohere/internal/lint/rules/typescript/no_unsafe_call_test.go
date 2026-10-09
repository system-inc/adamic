package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeCallFile = "/repository/source/Calling.ts"

func noUnsafeCallCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnsafeCallStaysSilent is upstream's fifteen passing cases verbatim.
//
// Each was re-measured one file per program against the installed 8.67.0 build under THIS
// harness's compiler options rather than upstream's, because upstream runs this rule's whole
// corpus under `tsconfig.noImplicitThis.json` and our harness pins `strict: true` with no way for
// a fixture to override it. All fifteen are clean under both settings, so nothing here depends on
// which one is in force. The two cases where the setting DOES decide the verdict are in the
// firing test, at the ids upstream produces under our configuration.
func TestNoUnsafeCallStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function foo(x: () => void) {\n  x();\n}\n",
		"function foo(x?: { a: () => void }) {\n  x?.a();\n}\n",
		"function foo(x: { a?: () => void }) {\n  x.a?.();\n}\n",
		"new Map();\n",
		"String.raw`foo`;\n",
		"const x = import('./foo');\n",
		"let foo: any = 23;\nString(foo); // ERROR: Unsafe call of an any typed value\n",
		"function foo<T extends any>(x: T) {\n  x();\n}\n",
		"// create a scope since it's illegal to declare a duplicate identifier\n// 'Function' in the global script scope.\n{\n  type Function = () => void;\n  const notGlobalFunctionType: Function = (() => {}) as Function;\n  notGlobalFunctionType();\n}\n",
		"interface SurprisinglySafe extends Function {\n  (): string;\n}\ndeclare const safe: SurprisinglySafe;\nsafe();\n",
		"interface CallGoodConstructBad extends Function {\n  (): void;\n}\ndeclare const safe: CallGoodConstructBad;\nsafe();\n",
		"interface ConstructSignatureMakesSafe extends Function {\n  new (): ConstructSignatureMakesSafe;\n}\ndeclare const safe: ConstructSignatureMakesSafe;\nnew safe();\n",
		"interface SafeWithNonVoidCallSignature extends Function {\n  (): void;\n  (x: string): string;\n}\ndeclare const safe: SafeWithNonVoidCallSignature;\nsafe();\n",
		"new Function('lol');\n",
		"Function('lol');\n",

		// Upstream's invalid case 11 belongs here under our compiler options and nowhere in
		// upstream's own lists. It asserts `unsafeCallThis` twice, which requires `noImplicitThis`
		// off; with it on, `this` in an object literal is typed as the literal rather than `any`
		// and the rule is correctly silent. Measured on the installed build under both settings:
		// two findings with the option off, zero with it on. Pinned here as silent because that is
		// the verdict this harness can observe, with the other half recorded in the rule's doc
		// comment so the branch is not read as unported.
		"const methods = {\n  methodA() {\n    return this.methodB()\n  },\n  methodB() {\n    return true\n  },\n  methodC() {\n    return this()\n  }\n};\n",

		// Measured boundary cases upstream does not write.

		// `unknown` is not `any`, and narrowing it to something other than a function keeps it so.
		"declare const value: unknown;\nif (typeof value === 'object' && value !== null) {\n  String(value);\n}\n",

		// A construct signature makes a Function subtype safe to CALL as well, which is the half
		// of the asymmetry the corpus exercises only in the construction direction.
		"interface ConstructOnly extends Function {\n  new (): ConstructOnly;\n}\ndeclare const safe: ConstructOnly;\nsafe();\n",

		// A tagged template whose tag is an ordinary function type.
		"declare function tag(strings: TemplateStringsArray): string;\ntag`hello`;\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnsafeCallCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeCall, noUnsafeCallFile, sourceText))
		})
	}
}

// noUnsafeCallFinding is one expected finding with every layer it can be wrong at.
//
// The rule has EIGHT message ids across three anchors, and the id is the whole judgment: which
// site fired, and whether the type was an `any` the author wrote or one the checker could not
// resolve. The rendered text carries a second judgment the id cannot, since `{{type}}` is filled
// with either "an `any`" or "a `Function`" and the two describe different defects.
type noUnsafeCallFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestNoUnsafeCallFires is upstream's reporting cases verbatim, at the verdicts this harness can
// observe.
//
// Twenty-three of upstream's twenty-four reporting cases are here; the twenty-fourth is in the
// silent test above for the compiler-option reason documented there. Case 23 reports under both
// settings and its IDS MOVE between them, from `errorCallThis` to `errorCall`, so the row asserts
// the latter and the rule's doc comment records the former.
//
// Every row asserts the span, the id, and the whole rendered message. The span matters more here
// than on most rules because the three anchors point at different things: a call and a tagged
// template report on the CALLEE, so `x.a.b.c.d.e.f.g()` underlines everything but the parentheses,
// while `new` reports on the WHOLE expression. Collapsing those would satisfy every id.
func TestNoUnsafeCallFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noUnsafeCallFinding
	}{
		{
			sourceText: "function foo(x: any) {\n  x();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: any) {\n  x?.();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: any) {\n  x.a.b.c.d.e.f.g();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x.a.b.c.d.e.f.g",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: any) {\n  x.a.b.c.d.e.f.g?.();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x.a.b.c.d.e.f.g",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: { a: any }) {\n  x.a();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x.a",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: { a: any }) {\n  x?.a();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x?.a",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: { a: any }) {\n  x.a?.();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x.a",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: any) {\n  new x();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new x()",
					wantId:      "unsafeNew",
					wantMessage: "Unsafe construction of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: { a: any }) {\n  new x.a();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new x.a()",
					wantId:      "unsafeNew",
					wantMessage: "Unsafe construction of an `any` typed value.",
				},
			},
		},
		{
			sourceText: "function foo(x: any) {\n  x`foo`;\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x",
					wantId:      "unsafeTemplateTag",
					wantMessage: "Unsafe use of an `any` typed template tag.",
				},
			},
		},
		{
			sourceText: "function foo(x: { tag: any }) {\n  x.tag`foo`;\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x.tag",
					wantId:      "unsafeTemplateTag",
					wantMessage: "Unsafe use of an `any` typed template tag.",
				},
			},
		},
		{
			sourceText: "const t: Function = () => {};\nt();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "t",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "const f: Function = () => {};\nf`oo`;\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "f",
					wantId:      "unsafeTemplateTag",
					wantMessage: "Unsafe use of a `Function` typed template tag.",
				},
			},
		},
		{
			sourceText: "declare const maybeFunction: unknown;\nif (typeof maybeFunction === 'function') {\n  maybeFunction('call', 'with', 'any', 'args');\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "maybeFunction",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "interface Unsafe extends Function {}\ndeclare const unsafe: Unsafe;\nunsafe();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "unsafe",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "interface Unsafe extends Function {}\ndeclare const unsafe: Unsafe;\nunsafe`bad`;\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "unsafe",
					wantId:      "unsafeTemplateTag",
					wantMessage: "Unsafe use of a `Function` typed template tag.",
				},
			},
		},
		{
			sourceText: "interface Unsafe extends Function {}\ndeclare const unsafe: Unsafe;\nnew unsafe();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new unsafe()",
					wantId:      "unsafeNew",
					wantMessage: "Unsafe construction of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "interface UnsafeToConstruct extends Function {\n  (): void;\n}\ndeclare const unsafe: UnsafeToConstruct;\nnew unsafe();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new unsafe()",
					wantId:      "unsafeNew",
					wantMessage: "Unsafe construction of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "interface StillUnsafe extends Function {\n  property: string;\n}\ndeclare const unsafe: StillUnsafe;\nunsafe();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "unsafe",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of a `Function` typed value.",
				},
			},
		},
		{
			sourceText: "let value: NotKnown;\nvalue();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "value",
					wantId:      "errorCall",
					wantMessage: "Unsafe call of a type that could not be resolved.",
				},
			},
		},
		{
			sourceText: "let value: NotKnown;\nvalue``;\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "value",
					wantId:      "errorTemplateTag",
					wantMessage: "Unsafe use of a template tag whose type could not be resolved.",
				},
			},
		},
		{
			sourceText: "let value: NotKnown;\nnew value();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new value()",
					wantId:      "errorNew",
					wantMessage: "Unsafe construction of a type that could not be resolved.",
				},
			},
		},
		{
			sourceText: "function callThis(this: NotKnown) {\n  this();\n  this.method();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "this",
					wantId:      "errorCall",
					wantMessage: "Unsafe call of a type that could not be resolved.",
				},
				{
					wantSpan:    "this.method",
					wantId:      "errorCall",
					wantMessage: "Unsafe call of a type that could not be resolved.",
				},
			},
		},

		// Measured additions covering discriminations the corpus leaves untested.
		{
			// A Function subtype with only a void-returning call signature. The corpus writes this
			// exact interface in its PASSING list, called rather than constructed, so this row is
			// the other half of that pair: the same type is safe to call and unsafe to construct,
			// which is the asymmetry the rule reproduces from upstream.
			sourceText: "interface CallGoodConstructBad extends Function {\n  (): void;\n}\ndeclare const safe: CallGoodConstructBad;\nnew safe();\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "new safe()",
					wantId:      "unsafeNew",
					wantMessage: "Unsafe construction of a `Function` typed value.",
				},
			},
		},
		{
			// A parenthesized callee. The corpus writes none, so this row is measured rather than
			// imported, and the first version of it was WRONG in a way only the measurement caught:
			// the checker answers the same type for `(x)` and for `x`, so the finding appears
			// either way and only the span moves. Upstream reports at columns 4-5, which is `x`,
			// because estree has no parenthesized node for its selector to bind to. Without the
			// skip this rule underlines `(x)` and every message-id assertion still passes.
			sourceText: "function foo(x: any) {\n  (x)();\n}\n",
			wantFindings: []noUnsafeCallFinding{
				{
					wantSpan:    "x",
					wantId:      "unsafeCall",
					wantMessage: "Unsafe call of an `any` typed value.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeCallCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeCall, noUnsafeCallFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so the span slices are against that text
			// rather than the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Id != want.wantId {
					t.Fatalf("finding %d id: expected %q, got %q", position, want.wantId,
						diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}
			}
		})
	}
}

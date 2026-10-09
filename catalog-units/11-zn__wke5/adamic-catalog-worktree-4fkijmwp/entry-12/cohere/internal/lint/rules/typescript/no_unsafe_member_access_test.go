package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeMemberAccessFile = "/repository/source/Members.ts"

func noUnsafeMemberAccessCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// noUnsafeMemberAccessDecoded routes a fixture's options through the rule's own exported decoder
// rather than building the options struct directly.
//
// The options text is the bare object rather than upstream's one-element array, because cohere's
// config layer unwraps the severity tuple before a decoder ever sees it. Going through the decoder
// is what puts the wire key name and the absent-versus-explicit-false handling under test.
func noUnsafeMemberAccessDecoded(t *testing.T, optionsJson string) any {
	t.Helper()
	decoded, err := DecodeNoUnsafeMemberAccessOptions([]byte(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestNoUnsafeMemberAccessStaysSilent is upstream's passing cases verbatim, plus the one reporting
// case that is silent under our compiler options.
//
// Upstream runs this corpus with `noImplicitThis` off and our harness pins `strict: true`, so every
// case was measured against the installed 8.67.0 build under BOTH settings. Thirty-five of
// thirty-six are identical either way. The exception is upstream's object-literal `this` case,
// which reports three findings under its configuration and is COMPLETELY silent under ours, because
// `this` in an object literal is typed as the literal rather than `any` once the option is on. It
// is pinned here at the verdict this harness can observe rather than deleted, with the other column
// recorded in the rule's doc comment.
func TestNoUnsafeMemberAccessStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{
			sourceText:  "function foo(x: { a: number }, y: any) {\n  x[y++];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x: { a: number }) {\n  x.a;\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x?: { a: number }) {\n  x?.a;\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x: { a: number }) {\n  x['a'];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x?: { a: number }) {\n  x?.['a'];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x: { a: number }, y: string) {\n  x[y];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x?: { a: number }, y: string) {\n  x?.[y];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x: string[]) {\n  x[1];\n}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "class B implements FG.A {}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "interface B extends FG.A {}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "class B implements F.S.T.A {}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "interface B extends F.S.T.A {}\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "function foo(x?: { a: number }) {\n  x?.a;\n}\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
		},
		{
			sourceText:  "function foo(x?: { a: number }, y: string) {\n  x?.[y];\n}\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
		},
		{
			sourceText:  "function foo(x: { a: number }, y: 'a') {\n  x?.[y];\n}\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
		},
		{
			sourceText:  "function foo(x: { a: number }, y: NotKnown) {\n  x?.[y];\n}\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
		},
		{
			sourceText:  "const methods = {\n  methodA() {\n    return this.methodB()\n  },\n  methodB() {\n    const getProperty = () => Math.random() > 0.5 ? 'methodB' : 'methodC'\n    return this[getProperty()]()\n  },\n  methodC() {\n    return true\n  },\n  methodD() {\n    return (this?.methodA)?.()\n  }\n};\n",
			optionsJson: "{}",
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeMemberAccessCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoUnsafeMemberAccess,
				noUnsafeMemberAccessFile, testCase.sourceText,
				noUnsafeMemberAccessDecoded(t, testCase.optionsJson)))
		})
	}
}

// noUnsafeMemberAccessFinding is one expected finding with all three layers it can be wrong at.
//
// Six message ids describe six different judgments across two anchors, and every one interpolates
// the property. The rendered text is where a port bracketing a computed key wrongly shows itself,
// and the span separates a finding on the property from one on a computed key at the same site.
type noUnsafeMemberAccessFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestNoUnsafeMemberAccessFires is upstream's reporting cases verbatim, at the verdicts this harness
// can observe, with every id, rendered text, and span taken from the installed 8.67.0 build.
//
// The span assertions carry more weight here than on most rules. A chain reports ONCE and on the
// innermost property, so `x.a.b.c` underlines `a`; a port without the memoized recursion reports
// three times, and one recursing the wrong way underlines `c`. Both satisfy every id assertion.
func TestNoUnsafeMemberAccessFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		optionsJson  string
		wantFindings []noUnsafeMemberAccessFinding
	}{
		{
			sourceText:  "function foo(x: any) {\n  x.a;\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "a",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .a on an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: any) {\n  x.a.b.c.d.e.f.g;\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "a",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .a on an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: any }) {\n  x.a.b.c.d.e.f.g;\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "b",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .b on an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: any) {\n  x['a'];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "'a'",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access ['a'] on an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: any) {\n  x['a']['b']['c'];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "'a'",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access ['a'] on an `any` value.",
				},
			},
		},
		{
			sourceText:  "let value: NotKnown;\n\nvalue.property;\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "property",
					wantId:      "errorMemberExpression",
					wantMessage: "Unsafe member access .property on a type that cannot be resolved.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: any) {\n  x[y];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [y] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x?: { a: number }, y: any) {\n  x?.[y];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [y] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: any) {\n  x[(y += 1)];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y += 1",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [y += 1] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: any) {\n  x[1 as any];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [1 as any] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: any) {\n  x[y()];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y()",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [y()] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: string[], y: any) {\n  x[y];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y",
					wantId:      "unsafeComputedMemberAccess",
					wantMessage: "Computed name [y] resolves to an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: NotKnown) {\n  x[y];\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y",
					wantId:      "errorComputedMemberAccess",
					wantMessage: "The type of computed name [y] cannot be resolved.",
				},
			},
		},
		{
			// Upstream's corpus asserts TWO findings here and this harness produces a third, on
			// `console.log`. That is not a rule difference: probed with a control, `console` does
			// not resolve in the program this harness builds, so its type is the checker's error
			// type and `errorMemberExpression` is the correct verdict for it. Declaring `console`
			// makes the case identical to upstream's, and that variant is the row below.
			//
			// Pinned at what this harness produces rather than weakened, because the extra finding
			// is right about the program it was given.
			sourceText:  "class C {\n  getObs$: any;\n  getPopularDepartments(): void {\n    this.getObs$.pipe().subscribe(res => {\n      console.log(res);\n    });\n  }\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "pipe",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .pipe on an `any` value.",
				},
				{
					wantSpan:    "subscribe",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .subscribe on an `any` value.",
				},
				{
					wantSpan:    "log",
					wantId:      "errorMemberExpression",
					wantMessage: "Unsafe member access .log on a type that cannot be resolved.",
				},
			},
		},
		{
			// The same case with `console` declared, which is upstream's corpus verdict exactly:
			// two findings, `pipe` before `subscribe`. Measured on the installed 8.67.0 build both
			// ways, which reports the same two either way because its harness resolves `console`.
			//
			// This row is also what pins the ORDER through a call. The object of `.subscribe` is a
			// CALL rather than a member access, so a recursion descending only through member
			// accesses never reaches `.pipe` and emits the two backwards. Every message id
			// assertion passes over that.
			sourceText:  "declare const console: { log(m: unknown): void };\nclass C {\n  getObs$: any;\n  getPopularDepartments(): void {\n    this.getObs$.pipe().subscribe(res => {\n      console.log(res);\n    });\n  }\n}\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "pipe",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .pipe on an `any` value.",
				},
				{
					wantSpan:    "subscribe",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .subscribe on an `any` value.",
				},
			},
		},
		{
			// Measured, not imported: the minimal shape of the same ordering question. The corpus
			// writes no access through a call outside the class case above.
			sourceText:  "declare const x: any;\nx.a().b;\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "a",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .a on an `any` value.",
				},
				{
					wantSpan:    "b",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .b on an `any` value.",
				},
			},
		},
		{
			// One link further, which pins that the call does NOT make the chain unsafe by
			// association: `.c` is suppressed by `.b` being unsafe, exactly as in a plain chain.
			sourceText:  "declare const x: any;\nx.a().b.c;\n",
			optionsJson: "{}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "a",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .a on an `any` value.",
				},
				{
					wantSpan:    "b",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .b on an `any` value.",
				},
			},
		},
		{
			sourceText:  "let value: any;\n\nvalue?.middle.inner;\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "inner",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .inner on an `any` value.",
				},
			},
		},
		{
			sourceText:  "let value: any;\n\nvalue?.outer.middle.inner;\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "middle",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .middle on an `any` value.",
				},
			},
		},
		{
			sourceText:  "let value: any;\n\nvalue.outer?.middle.inner;\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "outer",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .outer on an `any` value.",
				},
				{
					wantSpan:    "inner",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .inner on an `any` value.",
				},
			},
		},
		{
			sourceText:  "let value: any;\n\nvalue.outer.middle?.inner;\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "outer",
					wantId:      "unsafeMemberExpression",
					wantMessage: "Unsafe member access .outer on an `any` value.",
				},
			},
		},
		{
			sourceText:  "function foo(x: { a: number }, y: NotKnown) {\n  x[y];\n}\n",
			optionsJson: "{\"allowOptionalChaining\": true}",
			wantFindings: []noUnsafeMemberAccessFinding{
				{
					wantSpan:    "y",
					wantId:      "errorComputedMemberAccess",
					wantMessage: "The type of computed name [y] cannot be resolved.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeMemberAccessCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoUnsafeMemberAccess,
				noUnsafeMemberAccessFile, testCase.sourceText,
				noUnsafeMemberAccessDecoded(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

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

// TestNoUnsafeMemberAccessRequiresTheTypedHarness pins the nil-checker guard.
//
// It exists because a mutant removing that guard survives every other fixture, and the reason is
// structural rather than a coverage gap: the guard prevents a PANIC, and no ExpectFindings assertion
// can see one. This is also not hypothetical for this rule. Its first version was tested through
// `rule_testing.RunWithOptions`, the UNTYPED harness, which hands the rule a nil checker; every
// silent case passed vacuously and every reporting case failed, which reads like a broken rule
// rather than a broken harness.
//
// So this asserts the two halves separately. Under the untyped harness the rule must be silent and
// must not panic, and under the typed one the same source must report. Without the guard the first
// half takes the whole run down.
func TestNoUnsafeMemberAccessRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const x: any;\nx.a;\n"

	// The untyped harness gives the rule no checker. Silence here is the guard working; a panic is
	// what its absence produces, and a panic in one rule costs every rule its findings for the
	// whole file.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnsafeMemberAccess,
		noUnsafeMemberAccessFile, sourceText, noUnsafeMemberAccessDecoded(t, "{}")))

	// The control. Without it the row above is satisfied by a rule that can never report at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoUnsafeMemberAccess,
		noUnsafeMemberAccessFile, sourceText, noUnsafeMemberAccessDecoded(t, "{}")),
		"unsafeMemberExpression")
}

package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferFunctionTypeFile = "/repository/source/FunctionTypes.ts"

func preferFunctionTypeCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestPreferFunctionTypeStaysSilent is upstream's five passing cases verbatim, re-measured against
// the installed 8.67.0 build with the file written exactly as this harness writes it.
//
// Each one is a distinct reason to decline: a second member, a supertype that is not Function, a
// construct signature beside a call signature, and a call signature with no annotated return type.
// Skipping any of them ships that class of false positive.
func TestPreferFunctionTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"interface Foo {\n  (): void;\n  bar: number;\n}\n",
		"type Foo = {\n  (): void;\n  bar: number;\n};\n",
		"function foo(bar: { (): string; baz: number }): string {\n  return bar();\n}\n",
		"interface Foo {\n  bar: string;\n}\ninterface Bar extends Foo {\n  (): void;\n}\n",
		"interface Foo {\n  bar: string;\n}\ninterface Bar extends Function, Foo {\n  (): void;\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(preferFunctionTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, PreferFunctionType,
				preferFunctionTypeFile, sourceText))
		})
	}
}

// preferFunctionTypeFinding is one expected finding with every layer it can be wrong at.
//
// The two message ids describe different judgments, and the rendered text of each carries a second
// one: the ordinary message names whether the subject was an interface or a type literal, and the
// `this` message names the interface. The span separates a finding on the member from one on the
// `this` inside it.
type preferFunctionTypeFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string

	// wantPosition is the byte offset the finding starts at, asserted only where the span TEXT
	// cannot separate two candidates. The `this` message points at one of several `this` tokens and
	// every one of them slices to the same four characters, so a port reporting the last instead of
	// the first satisfies every span assertion. Zero means "not asserted", which is safe because no
	// finding in this rule can start at offset zero: it always points inside a declaration.
	wantPosition int

	// wantNoFix marks the findings upstream reports without a repair.
	wantNoFix bool
}

// TestPreferFunctionTypeFires is upstream's eighteen reporting cases verbatim, with every finding's
// id, rendered text, span, and repaired source taken from the installed 8.67.0 build.
//
// The repair is half this rule and the corpus asserts it, so `wantOutput` is the specification
// rather than a convenience. Three findings carry no repair and say so.
func TestPreferFunctionTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []preferFunctionTypeFinding
		wantOutput   string
	}{
		{
			sourceText: "interface Foo {\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = () => string;\n",
		},
		{
			sourceText: "export default interface Foo {\n  /** comment */\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "export default interface Foo {\n  /** comment */\n  (): string;\n}\n",
		},
		{
			sourceText: "interface Foo {\n  // comment\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "// comment\ntype Foo = () => string;\n",
		},
		{
			sourceText: "export interface Foo {\n  /** comment */\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "/** comment */\nexport type Foo = () => string;\n",
		},
		{
			sourceText: "export interface Foo {\n  // comment\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "// comment\nexport type Foo = () => string;\n",
		},
		{
			sourceText: "function foo(bar: { /* comment */ (s: string): number } | undefined): number {\n  return bar('hello');\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(s: string): number",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "function foo(bar: /* comment */ ((s: string) => number) | undefined): number {\n  return bar('hello');\n}\n",
		},
		{
			sourceText: "type Foo = {\n  (): string;\n};\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = () => string;\n",
		},
		{
			sourceText: "function foo(bar: { (s: string): number }): number {\n  return bar('hello');\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(s: string): number",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "function foo(bar: (s: string) => number): number {\n  return bar('hello');\n}\n",
		},
		{
			sourceText: "function foo(bar: { (s: string): number } | undefined): number {\n  return bar('hello');\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(s: string): number",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "function foo(bar: ((s: string) => number) | undefined): number {\n  return bar('hello');\n}\n",
		},
		{
			sourceText: "interface Foo extends Function {\n  (): void;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): void;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = () => void;\n",
		},
		{
			sourceText: "interface Foo<T> {\n  (bar: T): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(bar: T): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo<T> = (bar: T) => string;\n",
		},
		{
			sourceText: "interface Foo<T> {\n  (this: T): void;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(this: T): void;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo<T> = (this: T) => void;\n",
		},
		{
			sourceText: "type Foo<T> = { (this: string): T };\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(this: string): T",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo<T> = (this: string) => T;\n",
		},
		{
			sourceText: "interface Foo {\n  (arg: this): void;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "this",
					wantId:      "unexpectedThisOnFunctionOnlyInterface",
					wantMessage: "`this` refers to the function type 'Foo', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?",
					wantNoFix:   true,
				},
			},
			wantOutput: "interface Foo {\n  (arg: this): void;\n}\n",
		},
		{
			sourceText: "interface Foo {\n  (arg: number): this | undefined;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "this",
					wantId:      "unexpectedThisOnFunctionOnlyInterface",
					wantMessage: "`this` refers to the function type 'Foo', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?",
					wantNoFix:   true,
				},
			},
			wantOutput: "interface Foo {\n  (arg: number): this | undefined;\n}\n",
		},
		{
			sourceText: "// isn't actually valid ts but want to not give message saying it refers to Foo.\ninterface Foo {\n  (): {\n    a: {\n      nested: this;\n    };\n    between: this;\n    b: {\n      nested: string;\n    };\n  };\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): {\n    a: {\n      nested: this;\n    };\n    between: this;\n    b: {\n      nested: string;\n    };\n  };",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "// isn't actually valid ts but want to not give message saying it refers to Foo.\ntype Foo = () => {\n    a: {\n      nested: this;\n    };\n    between: this;\n    b: {\n      nested: string;\n    };\n  };\n",
		},
		{
			sourceText: "type X = {} | { (): void; }\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): void;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type X = {} | (() => void)\n",
		},
		{
			sourceText: "type X = {} & { (): void; };\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): void;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type X = {} & (() => void);\n",
		},

		{
			// Measured, not imported. A `declare` modifier does not change the judgment or the repair, and the rebuilt alias correctly
			// drops it, which is what upstream produces.
			sourceText: "declare interface Foo {\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = () => string;\n",
		},
		{
			// Measured, not imported. A comment ABOVE the interface is not a comment on the member and must not be moved. It survives
			// here because it sits outside the declaration's token range; an earlier version of this rule
			// filtered on `node.Pos()`, which reaches back over leading trivia, and duplicated it.
			sourceText: "/** docs */\ninterface Foo {\n  (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "/** docs */\ntype Foo = () => string;\n",
		},
		{
			// Measured, not imported. Export, generics, and a moved comment at once. The comment goes ABOVE the `export` because it
			// cannot sit between `export` and `type`, and the type parameter list survives with its closing
			// angle bracket, which sits one byte past what the parameter list's own End reports.
			sourceText: "export interface Foo<T> {\n  // c\n  (a: T): T;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(a: T): T;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "// c\nexport type Foo<T> = (a: T) => T;\n",
		},
		{
			// Measured, not imported. A construct signature alone is reported and rewritten to a constructor type. Upstream's corpus
			// only writes a construct signature BESIDE a call signature, where it is a passing case, so
			// nothing imported exercises the reporting half.
			sourceText: "interface Foo {\n  new (): string;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "new (): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = new () => string;\n",
		},
		{
			// Measured, not imported. An array type wraps, like a union and an intersection. The corpus writes the other two and not
			// this one, so the third arm of the wrapping test was unexercised.
			sourceText: "type X = { (): void }[];\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): void",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type X = (() => void)[];\n",
		},
		{
			// Measured, not imported. The call signature carries its OWN type parameters while the interface has none, which is a
			// different path through the alias rebuild than an interface with type parameters.
			sourceText: "interface Foo {\n  <T>(a: T): T;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "<T>(a: T): T;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = <T>(a: T) => T;\n",
		},
		{
			// Measured, not imported. No separator at all after the member. The semicolon strip has nothing to remove and the alias
			// still ends correctly.
			sourceText: "type Foo = {\n  (): string\n};\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "type Foo = () => string;\n",
		},
		{
			// Measured, not imported. An exported interface nested in a namespace, which is where most exported interfaces in this
			// tree actually live.
			sourceText: "namespace N {\n  export interface Foo {\n    (): string;\n  }\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string;",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "namespace N {\n  export type Foo = () => string;\n}\n",
		},
		{
			// Measured, not imported. The shape this rule actually finds in the tree, taken from the audit's first violation site. A
			// construct signature returning a type literal, exported. Every piece of the fixer is exercised
			// at once and the output is upstream's byte for byte.
			sourceText: "export interface UseEyeDropperOptions {\n  new (): {\n    open(): Promise<{ sRGBHex: string }>;\n  };\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "new (): {\n    open(): Promise<{ sRGBHex: string }>;\n  };",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Interface only has a call signature, you should use a function type instead.",
				},
			},
			wantOutput: "export type UseEyeDropperOptions = new () => {\n    open(): Promise<{ sRGBHex: string }>;\n  };\n",
		},
		{
			// Measured, and the one place this port deliberately differs from upstream. A member
			// separated by a comma is legal TypeScript, and upstream's fixer strips only a
			// semicolon, so it writes `type Foo = () => string,;`. Handed to the TypeScript parser
			// that is rejected with "';' expected", and a fix is applied with nobody watching.
			//
			// The finding is upstream's and reproduced; the repair is withheld. Rewriting the comma
			// to a semicolon would be the obvious improvement and is not taken, because it would be
			// a repair upstream never writes.
			sourceText: "type Foo = {\n  (): string,\n};\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:    "(): string,",
					wantId:      "functionTypeOverCallableType",
					wantMessage: "Type literal only has a call signature, you should use a function type instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "type Foo = {\n  (): string,\n};\n",
		},

		{
			// Measured, and it exists because a mutant reporting the LAST `this` instead of the
			// first survived every fixture: both tokens slice to the same four characters, so the
			// span text cannot tell them apart and only the offset can. Driven on the installed
			// 8.67.0 build, which reports at line 2 column 9, the first one.
			//
			// Upstream picks the first deliberately and says so at the line: several `this` types
			// in one signature would otherwise produce a pile of identical messages.
			sourceText: "interface Foo {\n  (arg: this, other: this): void;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:     "this",
					wantPosition: 24,
					wantId:       "unexpectedThisOnFunctionOnlyInterface",
					wantMessage:  "`this` refers to the function type 'Foo', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?",
					wantNoFix:    true,
				},
			},
			wantOutput: "interface Foo {\n  (arg: this, other: this): void;\n}\n",
		},
		{
			// The same discrimination with the two `this` types in different positions, one in a
			// parameter and one in the return annotation.
			sourceText: "interface Foo {\n  (arg: this): this;\n}\n",
			wantFindings: []preferFunctionTypeFinding{
				{
					wantSpan:     "this",
					wantPosition: 24,
					wantId:       "unexpectedThisOnFunctionOnlyInterface",
					wantMessage:  "`this` refers to the function type 'Foo', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?",
					wantNoFix:    true,
				},
			},
			wantOutput: "interface Foo {\n  (arg: this): this;\n}\n",
		},
	}
	for index, testCase := range cases {
		t.Run(preferFunctionTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferFunctionType, preferFunctionTypeFile,
				testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			anyFixExpected := false
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if want.wantPosition != 0 && diagnostic.Range.Pos() != want.wantPosition {
					t.Fatalf("finding %d position: expected %d, got %d", position, want.wantPosition,
						diagnostic.Range.Pos())
				}
				if diagnostic.Message.Id != want.wantId {
					t.Fatalf("finding %d id: expected %q, got %q", position, want.wantId,
						diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}

				// Whether a finding carries a repair is its own judgment, and the repaired source
				// cannot see it: a declined repair and one writing the original bytes look alike.
				wantFixes := 1
				if want.wantNoFix {
					wantFixes = 0
				} else {
					anyFixExpected = true
				}
				if len(diagnostic.Fixes) != wantFixes {
					t.Fatalf("finding %d fixes: expected %d, got %d", position, wantFixes,
						len(diagnostic.Fixes))
				}
			}

			if anyFixExpected {
				rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantOutput)+"\n")
			} else if strings.TrimSpace(testCase.wantOutput)+"\n" != onDisk {
				t.Fatalf("a case expecting no repair must expect the source unchanged")
			}
		})
	}
}

// TestPreferFunctionTypeKeepsTheExport pins the modifier this repair rebuilds rather than copies.
//
// Swept after `consistent-indexed-object-style` shipped an interface-to-alias repair that replaced
// the declaration from its first modifier and rebuilt only `type <name> = `, unexporting six
// interfaces. This one replaces the same span and survives because it writes `export ` back
// explicitly. `declare` is dropped, as upstream's installed build drops it; a type alias emits
// nothing either way, so that is a spelling and not a meaning. Both outputs are byte-identical to
// upstream's.
func TestPreferFunctionTypeKeepsTheExport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		want       string
	}{
		{"export declare interface Foo<T> {\n  (a: T): void;\n}\n", "export type Foo<T> = (a: T) => void;\n"},
		{"/** Doc. */\nexport interface Foo {\n  (): void;\n}\n", "/** Doc. */\nexport type Foo = () => void;\n"},
	}
	for _, testCase := range cases {
		result := rule_testing.Run(t, PreferFunctionType, preferFunctionTypeFile, testCase.sourceText)
		rule_testing.ExpectFindings(t, result, "functionTypeOverCallableType")
		rule_testing.ExpectFixedSource(t, result, testCase.want)
	}
}

// TestPreferFunctionTypeKeepsATypeParameterListWrittenOnePerLine is trpc's mutationOptions.ts,
// reduced. A list written one parameter per line ends in a trailing comma and a newline before its
// `>`, and the alias used to be cut one byte past the last parameter, at the comma, so the fix read
// `type Options<TDef, TFlags extends object = {}, = ...` and did not parse (TS1139, #kq9vtva).
// Upstream slices from the name through the list's own closing bracket, so the comma and the newline
// survive into the alias, as they do here.
func TestPreferFunctionTypeKeepsATypeParameterListWrittenOnePerLine(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		want       string
	}{
		{
			"export interface Options<\n  TDef,\n  TFlags extends object = {},\n> {\n  <TContext = unknown>(opts?: TDef): TContext;\n}\n",
			"export type Options<\n  TDef,\n  TFlags extends object = {},\n> = <TContext = unknown>(opts?: TDef) => TContext;\n",
		},
		{
			"interface Options<T,> {\n  (value: T): void;\n}\n",
			"type Options<T,> = (value: T) => void;\n",
		},
	}
	for _, testCase := range cases {
		result := rule_testing.Run(t, PreferFunctionType, preferFunctionTypeFile, testCase.sourceText)
		rule_testing.ExpectFindings(t, result, "functionTypeOverCallableType")
		rule_testing.ExpectFixedSource(t, result, testCase.want)
	}
}

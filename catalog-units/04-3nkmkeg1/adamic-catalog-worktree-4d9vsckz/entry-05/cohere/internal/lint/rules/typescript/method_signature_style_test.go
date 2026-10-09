package typescript

import (
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const methodSignatureStyleFile = "/repository/source/Thing.ts"

// methodSignatureStyleCaseName numbers a row so a failure names which one, since many rows differ
// only in the option above them or in one modifier.
func methodSignatureStyleCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodeMethodSignatureStyleOptions runs a configuration through the real decoder rather than
// building the options struct directly.
//
// The decoder holds the only default in this rule and reads a bare JSON string rather than an
// object, which is the one line with no upstream counterpart. An empty configuration is the bare
// `"error"` case, which the config layer turns into nil options, and it reaches the rule the same
// way here.
func decodeMethodSignatureStyleOptions(t *testing.T, configuration string) any {
	t.Helper()

	if configuration == "" {
		// A rule configured as a bare `"error"` is handed nil, not an empty struct. Passing nil
		// here is what puts the rule's own fallback under test rather than the decoder's.
		return nil
	}
	options, err := DecodeMethodSignatureStyleOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestMethodSignatureStyleStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All twenty-four of upstream's passing inputs with the option each carries, extracted from the
// clone's test file by parsing it with the TypeScript compiler rather than by reading it, so no
// escape sequence passed through a shell or a keyboard on the way here. Every one was additionally
// run through the installed 8.67.0 build, which reported nothing on all twenty-four.
func TestMethodSignatureStyleStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
	}{
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f: (a: string) => number;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  ['f']: (a: boolean) => void;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f: <T>(a: T) => T;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  ['f']: <T extends {}>(a: T, b: T) => T;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  'f!': </* a */ T>(/* b */ x: any /* c */) => void;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  get f(): number;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  set f(value: number): void;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "type Test = { readonly f: (a: string) => number };",
		},
		{
			configuration: "",
			sourceText:    "type Test = { ['f']?: (a: boolean) => void };",
		},
		{
			configuration: "",
			sourceText:    "type Test = { readonly f?: <T>(a?: T) => T };",
		},
		{
			configuration: "",
			sourceText:    "type Test = { readonly ['f']?: <T>(a: T, b: T) => T };",
		},
		{
			configuration: "",
			sourceText:    "type Test = { get f(): number };",
		},
		{
			configuration: "",
			sourceText:    "type Test = { set f(value: number): void };",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  f(a: string): number;\n}\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  ['f'](a: boolean): void;\n}\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  f<T>(a: T): T;\n}\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  ['f']<T extends {}>(a: T, b: T): T;\n}\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  'f!'</* a */ T>(/* b */ x: any /* c */): void;\n}\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { f(a: string): number };\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { ['f']?(a: boolean): void };\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { f?<T>(a?: T): T };\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { ['f']?<T>(a: T, b: T): T };\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { get f(): number };\n      ",
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { set f(value: number): void };\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(methodSignatureStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, MethodSignatureStyle,
				methodSignatureStyleFile, testCase.sourceText,
				decodeMethodSignatureStyleOptions(t, testCase.configuration)))
		})
	}
}

// TestMethodSignatureStyleFiresOnUpstreamFailCases is the imported reporting corpus.
//
// Thirty-eight inputs carrying fifty-nine findings between them, of which ten produce no unattended
// repair: four are `readonly` properties offered as suggestions instead, five have a `this` return
// type that a function type cannot express, and one sits inside a module declaration.
//
// Four things are asserted per row. The ids say which direction ran. The spans say where the finding
// points, which is the whole member. The applied source says what the edit engine will write
// unattended. And the suggestion, where there is one, says what a human accepting it would get,
// which is the only way to see a repair the engine never applies.
func TestMethodSignatureStyleFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantSpans     []string
		wantFixed     string
		wantSuggested []string
	}{
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f(a: string): number;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(a: string): number;"},
			wantFixed:     "\ninterface Test {\n  f: (a: string) => number;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  ['f'](a: boolean): void;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"['f'](a: boolean): void;"},
			wantFixed:     "\ninterface Test {\n  ['f']: (a: boolean) => void;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f<T>(a: T): T;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f<T>(a: T): T;"},
			wantFixed:     "\ninterface Test {\n  f: <T>(a: T) => T;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  ['f']<T extends {}>(a: T, b: T): T;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"['f']<T extends {}>(a: T, b: T): T;"},
			wantFixed:     "\ninterface Test {\n  ['f']: <T extends {}>(a: T, b: T) => T;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  'f!'</* a */ T>(/* b */ x: any /* c */): void;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"'f!'</* a */ T>(/* b */ x: any /* c */): void;"},
			wantFixed:     "\ninterface Test {\n  'f!': </* a */ T>(/* b */ x: any /* c */) => void;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ntype Test = { f(a: string): number };\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(a: string): number"},
			wantFixed:     "\ntype Test = { f: (a: string) => number };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ntype Test = { ['f']?(a: boolean): void };\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"['f']?(a: boolean): void"},
			wantFixed:     "\ntype Test = { ['f']?: (a: boolean) => void };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ntype Test = { f?<T>(a?: T): T };\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f?<T>(a?: T): T"},
			wantFixed:     "\ntype Test = { f?: <T>(a?: T) => T };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ntype Test = { ['f']?<T>(a: T, b: T): T };\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"['f']?<T>(a: T, b: T): T"},
			wantFixed:     "\ntype Test = { ['f']?: <T>(a: T, b: T) => T };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  f: (a: string) => number;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"f: (a: string) => number;"},
			wantFixed:     "\ninterface Test {\n  f(a: string): number;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  ['f']: (a: boolean) => void;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"['f']: (a: boolean) => void;"},
			wantFixed:     "\ninterface Test {\n  ['f'](a: boolean): void;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  f: <T>(a: T) => T;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"f: <T>(a: T) => T;"},
			wantFixed:     "\ninterface Test {\n  f<T>(a: T): T;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  ['f']: <T extends {}>(a: T, b: T) => T;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"['f']: <T extends {}>(a: T, b: T) => T;"},
			wantFixed:     "\ninterface Test {\n  ['f']<T extends {}>(a: T, b: T): T;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  'f!': </* a */ T>(/* b */ x: any /* c */) => void;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"'f!': </* a */ T>(/* b */ x: any /* c */) => void;"},
			wantFixed:     "\ninterface Test {\n  'f!'</* a */ T>(/* b */ x: any /* c */): void;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { f: (a: string) => number };\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"f: (a: string) => number"},
			wantFixed:     "\ntype Test = { f(a: string): number };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { ['f']?: (a: boolean) => void };\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"['f']?: (a: boolean) => void"},
			wantFixed:     "\ntype Test = { ['f']?(a: boolean): void };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { f?: <T>(a?: T) => T };\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"f?: <T>(a?: T) => T"},
			wantFixed:     "\ntype Test = { f?<T>(a?: T): T };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ntype Test = { ['f']?: <T>(a: T, b: T) => T };\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"['f']?: <T>(a: T, b: T) => T"},
			wantFixed:     "\ntype Test = { ['f']?<T>(a: T, b: T): T };\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "type Test = { readonly f: (a: string) => number };",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"readonly f: (a: string) => number"},
			wantFixed:     "",
			wantSuggested: []string{"type Test = { f(a: string): number };"},
		},
		{
			configuration: "\"method\"",
			sourceText:    "type Test = { readonly f?: <T>(a?: T) => T };",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"readonly f?: <T>(a?: T) => T"},
			wantFixed:     "",
			wantSuggested: []string{"type Test = { f?<T>(a?: T): T };"},
		},
		{
			configuration: "\"method\"",
			sourceText:    "type Test = { readonly ['f']?: <T>(a: T, b: T) => T };",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"readonly ['f']?: <T>(a: T, b: T) => T"},
			wantFixed:     "",
			wantSuggested: []string{"type Test = { ['f']?<T>(a: T, b: T): T };"},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Test {\n  readonly f: (a: string) => number;\n}\n      ",
			wantIds:       []string{"errorProperty"},
			wantSpans:     []string{"readonly f: (a: string) => number;"},
			wantFixed:     "",
			wantSuggested: []string{"\ninterface Test {\n  f(a: string): number;\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  semi(arg: string): void;\n  comma(arg: string): void,\n  none(arg: string): void\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"semi(arg: string): void;", "comma(arg: string): void,", "none(arg: string): void"},
			wantFixed:     "\ninterface Foo {\n  semi: (arg: string) => void;\n  comma: (arg: string) => void,\n  none: (arg: string) => void\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "\"method\"",
			sourceText:    "\ninterface Foo {\n  semi: (arg: string) => void;\n  comma: (arg: string) => void,\n  none: (arg: string) => void\n}\n      ",
			wantIds:       []string{"errorProperty", "errorProperty", "errorProperty"},
			wantSpans:     []string{"semi: (arg: string) => void;", "comma: (arg: string) => void,", "none: (arg: string) => void"},
			wantFixed:     "\ninterface Foo {\n  semi(arg: string): void;\n  comma(arg: string): void,\n  none(arg: string): void\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  x(\n    args: Pick<\n      Bar,\n      'one' | 'two' | 'three'\n    >,\n  ): Baz;\n  y(\n    foo: string,\n    bar: number,\n  ): void;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod"},
			wantSpans:     []string{"x(\n    args: Pick<\n      Bar,\n      'one' | 'two' | 'three'\n    >,\n  ): Baz;", "y(\n    foo: string,\n    bar: number,\n  ): void;"},
			wantFixed:     "\ninterface Foo {\n  x: (\n    args: Pick<\n      Bar,\n      'one' | 'two' | 'three'\n    >,\n  ) => Baz;\n  y: (\n    foo: string,\n    bar: number,\n  ) => void;\n}\n      ",
			wantSuggested: []string{"", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"foo(): one;", "foo(): two;", "foo(): three;"},
			wantFixed:     "\ninterface Foo {\n  foo: (() => one) & (() => two) & (() => three);\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  foo(bar: string): one;\n  foo(bar: number, baz: string): two;\n  foo(): three;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"foo(bar: string): one;", "foo(bar: number, baz: string): two;", "foo(): three;"},
			wantFixed:     "\ninterface Foo {\n  foo: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  [foo](bar: string): one;\n  [foo](bar: number, baz: string): two;\n  [foo](): three;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"[foo](bar: string): one;", "[foo](bar: number, baz: string): two;", "[foo](): three;"},
			wantFixed:     "\ninterface Foo {\n  [foo]: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Foo {\n  [foo](bar: string): one;\n  [foo](bar: number, baz: string): two;\n  [foo](): three;\n  bar(arg: string): void;\n  bar(baz: number): Foo;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"[foo](bar: string): one;", "[foo](bar: number, baz: string): two;", "[foo](): three;", "bar(arg: string): void;", "bar(baz: number): Foo;"},
			wantFixed:     "\ninterface Foo {\n  [foo]: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n  bar: ((arg: string) => void) & ((baz: number) => Foo);\n}\n      ",
			wantSuggested: []string{"", "", "", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\n        declare global {\n          namespace jest {\n            interface Matchers<R, T> {\n              // Add overloads specific to the DOM\n              toHaveProp<K extends keyof DomPropsOf<T>>(name: K, value?: DomPropsOf<T>[K]): R;\n              toHaveProps(props: Partial<DomPropsOf<T>>): R;\n            }\n          }\n        }\n      ",
			wantIds:       []string{"errorMethod", "errorMethod"},
			wantSpans:     []string{"toHaveProp<K extends keyof DomPropsOf<T>>(name: K, value?: DomPropsOf<T>[K]): R;", "toHaveProps(props: Partial<DomPropsOf<T>>): R;"},
			wantFixed:     "",
			wantSuggested: []string{"", ""},
		},
		{
			configuration: "",
			sourceText:    "\ntype Foo = {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"foo(): one;", "foo(): two;", "foo(): three;"},
			wantFixed:     "\ntype Foo = {\n  foo: (() => one) & (() => two) & (() => three);\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ndeclare const Foo: {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod", "errorMethod"},
			wantSpans:     []string{"foo(): one;", "foo(): two;", "foo(): three;"},
			wantFixed:     "\ndeclare const Foo: {\n  foo: (() => one) & (() => two) & (() => three);\n}\n      ",
			wantSuggested: []string{"", "", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface MyInterface {\n  methodReturningImplicitAny();\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"methodReturningImplicitAny();"},
			wantFixed:     "\ninterface MyInterface {\n  methodReturningImplicitAny: () => any;\n}\n      ",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f(value: number): this;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(value: number): this;"},
			wantFixed:     "",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  foo(): this;\n  foo(): Promise<this>;\n}\n      ",
			wantIds:       []string{"errorMethod", "errorMethod"},
			wantSpans:     []string{"foo(): this;", "foo(): Promise<this>;"},
			wantFixed:     "",
			wantSuggested: []string{"", ""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f(value: number): this | undefined;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(value: number): this | undefined;"},
			wantFixed:     "",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f(value: number): Promise<this>;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(value: number): Promise<this>;"},
			wantFixed:     "",
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\ninterface Test {\n  f(value: number): Promise<this | undefined>;\n}\n      ",
			wantIds:       []string{"errorMethod"},
			wantSpans:     []string{"f(value: number): Promise<this | undefined>;"},
			wantFixed:     "",
			wantSuggested: []string{""},
		},
	}
	for index, testCase := range cases {
		t.Run(methodSignatureStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MethodSignatureStyle,
				methodSignatureStyleFile, testCase.sourceText,
				decodeMethodSignatureStyleOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			for index, wantSuggested := range testCase.wantSuggested {
				diagnostic := result.Diagnostics[index]
				if wantSuggested == "" {
					if len(diagnostic.Suggestions) != 0 {
						t.Errorf("finding %d offers %d suggestions, want none",
							index, len(diagnostic.Suggestions))
					}
					continue
				}
				if len(diagnostic.Suggestions) != 1 || len(diagnostic.Suggestions[0].Fixes) != 1 {
					t.Fatalf("finding %d does not carry exactly one single-fix suggestion", index)
				}
				fix := diagnostic.Suggestions[0].Fixes[0]
				rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
					testCase.sourceText[fix.Range.End():]
				if rewritten != wantSuggested {
					t.Errorf("finding %d's suggestion produces:\n  %q\nwant:\n  %q",
						index, rewritten, wantSuggested)
				}
			}

			if testCase.wantFixed == "" {
				for index, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d carries %d fixes, want none because upstream withholds the repair",
							index, len(diagnostic.Fixes))
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestMethodSignatureStylePreservesEverythingInsideTheSignature is the type-safety fixture.
//
// This rule rebuilds a signature, which is the shape that lost type information twice in this
// project: one rewrite stranded a type annotation and widened eight declarations, and another
// dropped a return annotation it could not see from the parameter list.
//
// It is safe because it copies rather than re-renders. The parameter list is spliced as raw source
// from its opening parenthesis to its closing one, and the type parameter list is copied whole, so
// nothing inside either has to be enumerated to survive. These rows are what proves it: a generic
// constraint, an optional parameter, a rest parameter, a default, and a comment written between the
// parentheses all reach the rewritten source unchanged.
//
// Every expectation is what the installed 8.67.0 build produced, and each rewrite was additionally
// checked through the TypeScript compiler to confirm the member's type is identical before and
// after.
func TestMethodSignatureStylePreservesEverythingInsideTheSignature(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantFixed     string
		reason        string
	}{
		{
			configuration: "",
			sourceText:    "interface I { m<T extends object>(a: T): T; }",
			wantFixed:     "interface I { m: <T extends object>(a: T) => T; }",
			reason:        "a generic parameter with its constraint",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(a?: string): void; }",
			wantFixed:     "interface I { m: (a?: string) => void; }",
			reason:        "an optional parameter",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(...a: string[]): void; }",
			wantFixed:     "interface I { m: (...a: string[]) => void; }",
			reason:        "a rest parameter",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(a: string, b?: number): void; }",
			wantFixed:     "interface I { m: (a: string, b?: number) => void; }",
			reason:        "a mix of required and optional",
		},
		{
			configuration: "",
			sourceText:    "interface I { m?(a: string): void; }",
			wantFixed:     "interface I { m?: (a: string) => void; }",
			reason:        "an optional MEMBER, which moves onto the key",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(/* c */ a: string): void; }",
			wantFixed:     "interface I { m: (/* c */ a: string) => void; }",
			reason:        "a comment written inside the parameter list",
		},
		{
			configuration: "",
			sourceText:    "interface I { [k](a: string): void; }",
			wantFixed:     "interface I { [k]: (a: string) => void; }",
			reason:        "a computed key",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(a: string); }",
			wantFixed:     "interface I { m: (a: string) => any; }",
			reason:        "an absent return type becomes an explicit any",
		},
		{
			configuration: "",
			sourceText:    "type T = { m(a: string): void, n: number };",
			wantFixed:     "type T = { m: (a: string) => void, n: number };",
			reason:        "a comma delimiter, which a semicolon would break",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(a: { b: string }): void; }",
			wantFixed:     "interface I { m: (a: { b: string }) => void; }",
			reason:        "an inline object type in a parameter",
		},
		{
			configuration: "",
			sourceText:    "interface I { m<T, U>(a: T, b: U): [T, U]; }",
			wantFixed:     "interface I { m: <T, U>(a: T, b: U) => [T, U]; }",
			reason:        "two type parameters and a tuple return",
		},
	}
	for index, testCase := range cases {
		t.Run(methodSignatureStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MethodSignatureStyle,
				methodSignatureStyleFile, testCase.sourceText,
				decodeMethodSignatureStyleOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, "errorMethod")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestMethodSignatureStyleDiscriminatesOnCasesUpstreamDoesNotWrite covers the rest.
//
// Each row was run through the installed 8.67.0 build and carries the verdict that build produced,
// so a row asserting silence asserts upstream's silence rather than this port's.
func TestMethodSignatureStyleDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		reason        string
	}{
		{
			configuration: "",
			sourceText:    "interface I { m(): this; }",
			wantIds:       []string{"errorMethod"},
			reason:        "a this return has no property spelling",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(): Promise<this>; }",
			wantIds:       []string{"errorMethod"},
			reason:        "and neither does one nested inside a generic",
		},
		{
			configuration: "",
			sourceText:    "interface I { m(): void; }",
			wantIds:       []string{"errorMethod"},
			reason:        "a method with no parameters still reports",
		},
		{
			configuration: "\"method\"",
			sourceText:    "interface I { m: (a: string) => void; }",
			wantIds:       []string{"errorProperty"},
			reason:        "the method direction reports a function property",
		},
		{
			configuration: "\"method\"",
			sourceText:    "interface I { m: string; }",
			wantIds:       nil,
			reason:        "a non-function property is not this rule's business",
		},
		{
			configuration: "\"method\"",
			sourceText:    "interface I { readonly m: (a: string) => void; }",
			wantIds:       []string{"errorProperty"},
			reason:        "a readonly property reports with a suggestion, not a fix",
		},
		{
			configuration: "\"method\"",
			sourceText:    "interface I { m?: (a: string) => void; }",
			wantIds:       []string{"errorProperty"},
			reason:        "an optional function property reports",
		},
		{
			configuration: "",
			sourceText:    "declare module \"m\" { interface I { m(): void; } }",
			wantIds:       []string{"errorMethod"},
			reason:        "a member inside a module reports with no repair",
		},
		{
			configuration: "",
			sourceText:    "interface I { get m(): string; }",
			wantIds:       nil,
			reason:        "a getter signature is a different node kind",
		},
		{
			configuration: "",
			sourceText:    "interface I { new (a: string): I; }",
			wantIds:       nil,
			reason:        "a construct signature is not a method signature",
		},
		{
			configuration: "",
			sourceText:    "interface I { (a: string): void; }",
			wantIds:       nil,
			reason:        "nor is a call signature",
		},
	}
	for index, testCase := range cases {
		t.Run(methodSignatureStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, MethodSignatureStyle,
				methodSignatureStyleFile, testCase.sourceText,
				decodeMethodSignatureStyleOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestMethodSignatureStyleDecoderResolvesTheStyle puts the one line with no upstream counterpart
// under test.
func TestMethodSignatureStyleDecoderResolvesTheStyle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		wantStyle     MethodSignatureStyleSetting
	}{
		{"\"property\"", MethodSignatureStyleProperty},
		{"\"method\"", MethodSignatureStyleMethod},
		// An unrecognized spelling keeps the default rather than producing an empty style that
		// would match neither arm and silence the rule.
		{"\"nonsense\"", MethodSignatureStyleProperty},
	}
	for index, testCase := range cases {
		t.Run(methodSignatureStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeMethodSignatureStyleOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.configuration, err)
			}
			options, ok := decoded.(MethodSignatureStyleOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than MethodSignatureStyleOptions", decoded)
			}
			if options.Style != testCase.wantStyle {
				t.Errorf("style resolved to %q, want %q", options.Style, testCase.wantStyle)
			}
		})
	}
}

// TestMethodSignatureStyleNilOptionsFallsBackToUpstreamDefault bypasses the decoder entirely.
func TestMethodSignatureStyleNilOptionsFallsBackToUpstreamDefault(t *testing.T) {
	t.Parallel()

	// `property` is the default, so a shorthand method reports and a function property does not.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, MethodSignatureStyle,
		methodSignatureStyleFile, "interface I { m(a: string): void; }", nil), "errorMethod")
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, MethodSignatureStyle,
		methodSignatureStyleFile, "interface I { m: (a: string) => void; }", nil))
}

// TestMethodSignatureStyleSurvivesMalformedSignatures is a crash fixture.
//
// The repair scans the raw source backwards for a parameter list's opening parenthesis and forwards
// for its closing one, and error recovery produces signatures with neither. The walk recovers per
// FILE rather than per rule, so one panic would cost every rule that file rather than one finding.
//
// There is no finding to assert; the assertion is that the run completes.
func TestMethodSignatureStyleSurvivesMalformedSignatures(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"noParens":    "interface I { m: void; }",
		"unclosed":    "interface I { m(a: string): void;",
		"noName":      "interface I { (a: string): void; }",
		"emptyIface":  "interface I {}",
		"strayAngle":  "interface I { m<(a: string): void; }",
		"noBody":      "interface I",
		"emptyLit":    "type T = {};",
		"bareArrow":   "type T = { m: => void };",
		"halfGeneric": "interface I { m<T(a: T): T; }",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.RunWithOptions(t, MethodSignatureStyle, methodSignatureStyleFile,
				sourceText, decodeMethodSignatureStyleOptions(t, "\"property\""))
			rule_testing.RunWithOptions(t, MethodSignatureStyle, methodSignatureStyleFile,
				sourceText, decodeMethodSignatureStyleOptions(t, "\"method\""))
		})
	}
}

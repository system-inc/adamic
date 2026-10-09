package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const classLiteralPropertyStyleFile = "/repository/source/Thing.ts"

// classLiteralPropertyStyleCaseName numbers a row so a failure names which one, since many rows
// differ only in a modifier or in the option above them.
func classLiteralPropertyStyleCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodeClassLiteralPropertyStyleOptions runs a configuration through the real decoder rather than
// building the options struct directly.
//
// The decoder holds the only default in this rule and reads a bare JSON string rather than an
// object, which is the one line with no upstream counterpart. A fixture constructing the options
// struct by hand would leave it unexercised. An empty configuration is the bare `"error"` case,
// which the config layer turns into nil options, and it reaches the rule the same way here.
func decodeClassLiteralPropertyStyleOptions(t *testing.T, configuration string) any {
	t.Helper()

	if configuration == "" {
		// A rule configured as a bare `"error"` is handed nil, not an empty struct. Passing nil
		// here is what puts the rule's own fallback under test rather than the decoder's.
		return nil
	}
	options, err := DecodeClassLiteralPropertyStyleOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestClassLiteralPropertyStyleStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All thirty-one of upstream's passing inputs with the option each carries, extracted from the
// clone's test file by parsing it with the TypeScript compiler rather than by reading it, so no
// escape sequence passed through a shell or a keyboard on the way here. Every one was additionally
// run through the installed 8.67.0 build driven by the ESLint 10.8.1 Linter API, which reported
// nothing on all thirty-one.
//
// The same source appears under both options in several places, which is the point: this rule's
// verdict lives above the code as often as in it.
func TestClassLiteralPropertyStyleStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
	}{
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  declare readonly p1 = 1;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  readonly p1 = 'hello world';\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  p1 = 'hello world';\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  static p1 = 'hello world';\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  p1: string;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get p1();\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get p1() {}\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nabstract class Mx {\n  abstract get p1(): string;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get mySetting() {\n    if (this._aValue) {\n      return 'on';\n    }\n\n    return 'off';\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get mySetting() {\n    return `build-${process.env.build}`;\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  getMySetting() {\n    if (this._aValue) {\n      return 'on';\n    }\n\n    return 'off';\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public readonly myButton = styled.button`\n    color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n  `;\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  set p1(val) {}\n  get p1() {\n    return '';\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nlet p1 = 'p1';\nclass Mx {\n  set [p1](val) {}\n  get [p1]() {\n    return '';\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nlet p1 = 'p1';\nclass Mx {\n  set [/* before set */ p1 /* after set */](val) {}\n  get [/* before get */ p1 /* after get */]() {\n    return '';\n  }\n}\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  set ['foo'](val) {}\n  get foo() {\n    return '';\n  }\n  set bar(val) {}\n  get ['bar']() {\n    return '';\n  }\n  set ['baz'](val) {}\n  get baz() {\n    return '';\n  }\n}\n    ",
		},
		{
			configuration: "\"fields\"",
			sourceText:    "\nclass Mx {\n  public get myButton() {\n    return styled.button`\n      color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n    `;\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  declare public readonly foo = 1;\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  get p1() {\n    return 'hello world';\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  p1 = 'hello world';\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  p1: string;\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  readonly p1 = [1, 2, 3];\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  static p1: string;\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  static get p1() {\n    return 'hello world';\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  public readonly myButton = styled.button`\n    color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n  `;\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  public get myButton() {\n    return styled.button`\n      color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n    `;\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly foo: string = 'bar';\n  constructor(foo: string) {\n    this.foo = foo;\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly foo: string = 'bar';\n  constructor(foo: string) {\n    this['foo'] = foo;\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly foo: string = 'bar';\n  constructor(foo: string) {\n    const bar = new (class {\n      private readonly foo: string = 'baz';\n      constructor() {\n        this.foo = 'qux';\n      }\n    })();\n    this['foo'] = foo;\n  }\n}\n      ",
		},
		{
			configuration: "",
			sourceText:    "\ndeclare abstract class BaseClass {\n  get cursor(): string;\n}\n\nclass ChildClass extends BaseClass {\n  override get cursor() {\n    return 'overridden value';\n  }\n}\n      ",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\ndeclare abstract class BaseClass {\n  protected readonly foo: string;\n}\n\nclass ChildClass extends BaseClass {\n  protected override readonly foo = 'bar';\n}\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(classLiteralPropertyStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, testCase.sourceText,
				decodeClassLiteralPropertyStyleOptions(t, testCase.configuration)))
		})
	}
}

// TestClassLiteralPropertyStyleFiresOnUpstreamFailCases is the imported reporting corpus.
//
// Twenty-two inputs carrying twenty-two diagnostics and twenty-one suggestions between them. The
// count differing by one is the decorated getter, which upstream reports while deliberately
// withholding the repair, and that row is the reason the suggestion count is asserted per row
// rather than assumed equal to the finding count.
//
// Four things are asserted on each row and each catches a different defect. The ids say which of the
// two directions ran. The spans say where the finding points, which is the member's KEY rather than
// the member, so a reader is shown the name while the repair rewrites the whole declaration. The
// suggestion count says whether a repair was offered at all. And the applied source says what a
// human accepting that suggestion would get, which no message-id assertion can see.
func TestClassLiteralPropertyStyleFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantSpans     []string
		wantSuggested []string
	}{
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get p1() {\n    return 'hello world';\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  readonly p1 = 'hello world';\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  get p1() {\n    return `hello world`;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  readonly p1 = `hello world`;\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  static get p1() {\n    return 'hello world';\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  static readonly p1 = 'hello world';\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public static get foo() {\n    return 1;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"foo"},
			wantSuggested: []string{"\nclass Mx {\n  public static readonly foo = 1;\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public static get n(): 1 | 2 {\n    return 1;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"n"},
			wantSuggested: []string{"\nclass Mx {\n  public static readonly n: 1 | 2 = 1;\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public static get n() /* before */ : 1 | 2 /* after */ {\n    return 1;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"n"},
			wantSuggested: []string{"\nclass Mx {\n  public static readonly n /* before */ : 1 | 2 /* after */ = 1;\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  @logAccess\n  public static get foo(): number {\n    return 1;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"foo"},
			wantSuggested: []string{""},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public get [myValue]() {\n    return 'a literal value';\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"myValue"},
			wantSuggested: []string{"\nclass Mx {\n  public readonly [myValue] = 'a literal value';\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public get [myValue]() {\n    return 12345n;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"myValue"},
			wantSuggested: []string{"\nclass Mx {\n  public readonly [myValue] = 12345n;\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  public readonly [myValue] = 'a literal value';\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"myValue"},
			wantSuggested: []string{"\nclass Mx {\n  public get [myValue]() { return 'a literal value'; }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  readonly p1 = 'hello world';\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  get p1() { return 'hello world'; }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  readonly p1 = `hello world`;\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  get p1() { return `hello world`; }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  static readonly p1 = 'hello world';\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  static get p1() { return 'hello world'; }\n}\n      "},
		},
		{
			configuration: "\"fields\"",
			sourceText:    "\nclass Mx {\n  protected get p1() {\n    return 'hello world';\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  protected readonly p1 = 'hello world';\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  protected readonly p1 = 'hello world';\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  protected get p1() { return 'hello world'; }\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public static get p1() {\n    return 'hello world';\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  public static readonly p1 = 'hello world';\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  public static readonly p1 = 'hello world';\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"p1"},
			wantSuggested: []string{"\nclass Mx {\n  public static get p1() { return 'hello world'; }\n}\n      "},
		},
		{
			configuration: "",
			sourceText:    "\nclass Mx {\n  public get myValue() {\n    return gql`\n      {\n        user(id: 5) {\n          firstName\n          lastName\n        }\n      }\n    `;\n  }\n}\n      ",
			wantIds:       []string{"preferFieldStyle"},
			wantSpans:     []string{"myValue"},
			wantSuggested: []string{"\nclass Mx {\n  public readonly myValue = gql`\n      {\n        user(id: 5) {\n          firstName\n          lastName\n        }\n      }\n    `;\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass Mx {\n  public readonly myValue = gql`\n    {\n      user(id: 5) {\n        firstName\n        lastName\n      }\n    }\n  `;\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"myValue"},
			wantSuggested: []string{"\nclass Mx {\n  public get myValue() { return gql`\n    {\n      user(id: 5) {\n        firstName\n        lastName\n      }\n    }\n  `; }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly foo: string = 'bar';\n  constructor(foo: string) {\n    const bar = new (class {\n      private readonly foo: string = 'baz';\n      constructor() {\n        this.foo = 'qux';\n      }\n    })();\n  }\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"foo"},
			wantSuggested: []string{"\nclass A {\n  private get foo() { return 'bar'; }\n  constructor(foo: string) {\n    const bar = new (class {\n      private readonly foo: string = 'baz';\n      constructor() {\n        this.foo = 'qux';\n      }\n    })();\n  }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly ['foo']: string = 'bar';\n  constructor(foo: string) {\n    const bar = new (class {\n      private readonly foo: string = 'baz';\n      constructor() {}\n    })();\n\n    if (bar) {\n      this.foo = 'baz';\n    }\n  }\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"foo"},
			wantSuggested: []string{"\nclass A {\n  private readonly ['foo']: string = 'bar';\n  constructor(foo: string) {\n    const bar = new (class {\n      private get foo() { return 'baz'; }\n      constructor() {}\n    })();\n\n    if (bar) {\n      this.foo = 'baz';\n    }\n  }\n}\n      "},
		},
		{
			configuration: "\"getters\"",
			sourceText:    "\nclass A {\n  private readonly foo: string = 'bar';\n  constructor(foo: string) {\n    function func() {\n      this.foo = 'aa';\n    }\n  }\n}\n      ",
			wantIds:       []string{"preferGetterStyle"},
			wantSpans:     []string{"foo"},
			wantSuggested: []string{"\nclass A {\n  private get foo() { return 'bar'; }\n  constructor(foo: string) {\n    function func() {\n      this.foo = 'aa';\n    }\n  }\n}\n      "},
		},
	}
	for index, testCase := range cases {
		t.Run(classLiteralPropertyStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, testCase.sourceText,
				decodeClassLiteralPropertyStyleOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			// One entry per finding: the source a human would get by accepting that finding's
			// suggestion, or the empty string where upstream offers none.
			for index, wantSuggested := range testCase.wantSuggested {
				diagnostic := result.Diagnostics[index]
				if wantSuggested == "" {
					if len(diagnostic.Suggestions) != 0 {
						t.Errorf("finding %d offers %d suggestions, want none",
							index, len(diagnostic.Suggestions))
					}
					continue
				}
				if len(diagnostic.Suggestions) != 1 {
					t.Fatalf("finding %d offers %d suggestions, want one",
						index, len(diagnostic.Suggestions))
				}
				suggestion := diagnostic.Suggestions[0]
				if len(suggestion.Fixes) != 1 {
					t.Fatalf("finding %d's suggestion carries %d fixes, want one",
						index, len(suggestion.Fixes))
				}
				fix := suggestion.Fixes[0]
				rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
					testCase.sourceText[fix.Range.End():]
				if rewritten != wantSuggested {
					t.Errorf("finding %d's suggestion produces:\n  %q\nwant:\n  %q",
						index, rewritten, wantSuggested)
				}
			}
		})
	}
}

// TestClassLiteralPropertyStylePreservesTheReturnAnnotation is the type-safety fixture.
//
// A repair that rebuilds a member signature loses whatever it does not re-render, and two rules
// shipped that way tonight: one widened eight declarations to `any` by stranding a type annotation,
// and one dropped a `: void` return type it could not see from the parameter list.
//
// This rule does not rebuild the signature. It splices the RAW SOURCE between the parameter list's
// closing parenthesis and the body's opening brace, so anything living there survives byte for byte
// without being enumerated. These rows are what proves that: the annotation, its unusual spacing,
// and a comment written in the same span all reach the rewritten source unchanged.
//
// Upstream's corpus is TypeScript so it does carry annotated cases, but none with the spacing or the
// comment, and those are the two that would survive an enumerating fixer while telling you nothing.
// Every expectation is what the installed 8.67.0 build produced for that exact input.
func TestClassLiteralPropertyStylePreservesTheReturnAnnotation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText    string
		wantSuggested string
		reason        string
	}{
		{
			sourceText:    "class C { get x(): number { return 1; } }",
			wantSuggested: "class C { readonly x: number = 1; }",
			reason:        "a return annotation survives",
		},
		{
			sourceText:    "class C { public static get x(): 1 { return 1; } }",
			wantSuggested: "class C { public static readonly x: 1 = 1; }",
			reason:        "and so do the modifiers beside it",
		},
		{
			sourceText:    "class C { protected get x(): \"a\" { return \"a\"; } }",
			wantSuggested: "class C { protected readonly x: \"a\" = \"a\"; }",
			reason:        "a literal-typed annotation survives intact",
		},
		{
			sourceText:    "class C { get x()   :   number   { return 1; } }",
			wantSuggested: "class C { readonly x   :   number   = 1; }",
			reason:        "the spacing inside the span is copied, not re-rendered",
		},
		{
			sourceText:    "class C { get x() /* c */ { return 1; } }",
			wantSuggested: "class C { readonly x /* c */ = 1; }",
			reason:        "and so is a comment written there",
		},
		{
			sourceText:    "class C { get x(): readonly string[] { return `a`; } }",
			wantSuggested: "class C { readonly x: readonly string[] = `a`; }",
			reason:        "an annotation containing its own brackets survives",
		},
	}
	for index, testCase := range cases {
		t.Run(classLiteralPropertyStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, testCase.sourceText,
				decodeClassLiteralPropertyStyleOptions(t, "\"fields\""))
			rule_testing.ExpectFindings(t, result, "preferFieldStyle")

			suggestion := result.Diagnostics[0].Suggestions[0]
			fix := suggestion.Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSuggested {
				t.Errorf("the suggestion produces:\n  %q\nwant:\n  %q",
					rewritten, testCase.wantSuggested)
			}
		})
	}
}

// TestClassLiteralPropertyStyleDiscriminatesOnCasesUpstreamDoesNotWrite covers the parser folds.
//
// TSESTree folds a string, number, boolean, null, regex and bigint into one `Literal` node type and
// typescript-go does not, so upstream's one-line `case Literal: return true` becomes an enumerated
// list here and every arm needs its own row. The corpus writes only some of them.
//
// Each row was run through the installed 8.67.0 build and carries the verdict that build produced,
// so a row asserting silence asserts upstream's silence rather than this port's.
func TestClassLiteralPropertyStyleDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantCount     int
		reason        string
	}{
		{
			configuration: "",
			sourceText:    "class C { get x() { return true; } }",
			wantCount:     1,
			reason:        "a boolean is a Literal upstream and a keyword node here",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return false; } }",
			wantCount:     1,
			reason:        "and so is false",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return null; } }",
			wantCount:     1,
			reason:        "and null",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return 1n; } }",
			wantCount:     1,
			reason:        "a bigint literal",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return /re/; } }",
			wantCount:     1,
			reason:        "a regex literal",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return `a`; } }",
			wantCount:     1,
			reason:        "a template with no substitution",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return `a${b}`; } }",
			wantCount:     0,
			reason:        "a template WITH a substitution is not a literal",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return tag`a`; } }",
			wantCount:     1,
			reason:        "a tagged template with no substitution is",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return tag`a${b}`; } }",
			wantCount:     0,
			reason:        "and one with a substitution is not",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return undefined; } }",
			wantCount:     0,
			reason:        "undefined is an identifier, not a literal",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return -1; } }",
			wantCount:     0,
			reason:        "a negated numeric literal is a unary expression",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { const a = 1; return 1; } }",
			wantCount:     0,
			reason:        "only the FIRST statement is read",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() {} }",
			wantCount:     0,
			reason:        "an empty body has no first statement",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return 1; } set x(v) {} }",
			wantCount:     0,
			reason:        "a paired setter exempts the getter",
		},
		{
			configuration: "",
			sourceText:    "class C { get x() { return 1; } set [\"x\"](v) {} }",
			wantCount:     0,
			reason:        "and the pairing is on the key value, not its spelling",
		},
		{
			configuration: "",
			sourceText:    "class C extends B { override get x() { return 1; } }",
			wantCount:     0,
			reason:        "an override cannot become a field",
		},
		{
			configuration: "",
			sourceText:    "class C { @dec get x() { return 1; } }",
			wantCount:     1,
			reason:        "a decorated getter reports with no repair offered",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { readonly x = 1; }",
			wantCount:     1,
			reason:        "the getters direction reports a readonly field",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { x = 1; }",
			wantCount:     0,
			reason:        "a mutable field is not this rule's business",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { declare readonly x: 1; }",
			wantCount:     0,
			reason:        "a declare field has no runtime initializer",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C extends B { override readonly x = 1; }",
			wantCount:     0,
			reason:        "an override cannot become a getter",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { readonly x = 1; constructor() { this.x = 2; } }",
			wantCount:     0,
			reason:        "a constructor assignment exempts the field",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { readonly x = 1; constructor() { function f() { this.x = 2; } } }",
			wantCount:     1,
			reason:        "but not one inside a nested function",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { readonly x = 1; method() { this.x = 2; } }",
			wantCount:     1,
			reason:        "nor one outside the constructor",
		},
		{
			configuration: "\"getters\"",
			sourceText:    "class C { readonly x = 1; readonly y = 2; }",
			wantCount:     2,
			reason:        "two fields report twice",
		},
	}
	for index, testCase := range cases {
		t.Run(classLiteralPropertyStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, testCase.sourceText,
				decodeClassLiteralPropertyStyleOptions(t, testCase.configuration))
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				if testCase.configuration == "\"getters\"" {
					wantIds[index] = "preferGetterStyle"
				} else {
					wantIds[index] = "preferFieldStyle"
				}
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestClassLiteralPropertyStyleKeepsWhatTheMemberCarries pins the repair's treatment of everything
// in front of the member's name.
//
// The first row was broken: a decorated readonly field was offered a getter suggestion that replaced
// the member from its decorator on and rebuilt only accessibility and `static`, deleting `@dec`.
// Upstream's installed build offers the identical suggestion. It now reports with no suggestion,
// which is what upstream itself does in the opposite direction for a decorated getter, for the same
// reason: a field decorator and an accessor decorator do not receive the same thing.
//
// The other rows are what the suggestion keeps in both directions: the JSDoc above the member, which
// sits outside the replaced span, and the accessibility and `static` modifiers, which are rebuilt.
func TestClassLiteralPropertyStyleKeepsWhatTheMemberCarries(t *testing.T) {
	t.Parallel()

	t.Run("a decorated field is offered nothing", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle, classLiteralPropertyStyleFile,
			"declare const dec: any;\nclass C {\n  @dec\n  public static readonly foo = 'x';\n}\n",
			decodeClassLiteralPropertyStyleOptions(t, "\"getters\""))
		rule_testing.ExpectFindings(t, result, "preferGetterStyle")
		if suggestions := result.Diagnostics[0].Suggestions; len(suggestions) != 0 {
			t.Fatalf("a decorated field was offered %d suggestion(s), which would delete the decorator", len(suggestions))
		}
	})

	cases := []struct {
		name          string
		configuration string
		sourceText    string
		wantApplied   string
	}{
		{
			"a documented static field becomes a getter",
			"\"getters\"",
			"class C {\n  /** Doc. */\n  protected static readonly foo = 'x';\n}\n",
			"class C {\n  /** Doc. */\n  protected static get foo() { return 'x'; }\n}\n",
		},
		{
			"a documented static getter becomes a field",
			"\"fields\"",
			"class C {\n  /** Doc. */\n  protected static get foo() {\n    return 'x';\n  }\n}\n",
			"class C {\n  /** Doc. */\n  protected static readonly foo = 'x';\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle, classLiteralPropertyStyleFile,
				testCase.sourceText, decodeClassLiteralPropertyStyleOptions(t, testCase.configuration))
			if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].Suggestions) != 1 {
				t.Fatalf("expected one finding carrying one suggestion, got %d findings", len(result.Diagnostics))
			}
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected a suggestion only, got %d fixes", len(result.Diagnostics[0].Fixes))
			}
			applied := testCase.sourceText
			for _, fix := range result.Diagnostics[0].Suggestions[0].Fixes {
				applied = applied[:fix.Range.Pos()] + fix.Text + applied[fix.Range.End():]
			}
			if applied != testCase.wantApplied {
				t.Fatalf("suggestion applied:\n  %q\nwant:\n  %q", applied, testCase.wantApplied)
			}
		})
	}
}

// TestClassLiteralPropertyStyleDecoderResolvesTheStyle puts the one line with no upstream
// counterpart under test.
//
// The option is a bare JSON string rather than an object, and its default is `fields` rather than a
// Go zero value. Both are invisible to any fixture that builds the options struct directly, which
// is why every other test in this file routes through the decoder and why this one asserts the
// decoder's own output.
func TestClassLiteralPropertyStyleDecoderResolvesTheStyle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		wantStyle     ClassLiteralPropertyStyleSetting
	}{
		{"\"fields\"", ClassLiteralPropertyStyleFields},
		{"\"getters\"", ClassLiteralPropertyStyleGetters},
	}
	for index, testCase := range cases {
		t.Run(classLiteralPropertyStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeClassLiteralPropertyStyleOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.configuration, err)
			}
			options, ok := decoded.(ClassLiteralPropertyStyleOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than ClassLiteralPropertyStyleOptions", decoded)
			}
			if options.Style != testCase.wantStyle {
				t.Errorf("style resolved to %q, want %q", options.Style, testCase.wantStyle)
			}
		})
	}
}

// TestClassLiteralPropertyStyleNilOptionsFallsBackToUpstreamDefault bypasses the decoder entirely.
//
// A rule configured as a bare `"error"` is handed nil, and a bare type assertion on nil yields the
// zero value, which for this rule is an empty style matching neither arm. That shape registers on
// every file and reports nothing while every decoder-routed fixture stays green, which is the exact
// failure this project has shipped before.
func TestClassLiteralPropertyStyleNilOptionsFallsBackToUpstreamDefault(t *testing.T) {
	t.Parallel()

	// `fields` is the default, so a literal getter reports and a readonly field does not.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
		classLiteralPropertyStyleFile, "class C { get x() { return 1; } }", nil), "preferFieldStyle")
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
		classLiteralPropertyStyleFile, "class C { readonly x = 1; }", nil))
}

// TestClassLiteralPropertyStyleSurvivesMalformedClassMembers is a crash fixture.
//
// The field direction scans backwards from the body for the parameter list's closing parenthesis,
// and error recovery produces members that have a body and no parenthesis at all. The walk recovers
// per FILE rather than per rule, so one panic would cost every rule that file rather than one
// finding.
//
// There is no finding to assert; the assertion is that the run completes. Upstream cannot reach most
// of these, since its parser rejects what ours recovers from, so asserting a verdict would be
// inventing one.
func TestClassLiteralPropertyStyleSurvivesMalformedClassMembers(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"noParens":     "class C { get x { return 1; } }",
		"noBody":       "class C { get x() }",
		"noName":       "class C { get () { return 1; } }",
		"unclosed":     "class C { get x() { return 1; }",
		"emptyClass":   "class C {}",
		"strayGet":     "class C { get }",
		"getOnly":      "class C { get",
		"readonlyOnly": "class C { readonly }",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, sourceText,
				decodeClassLiteralPropertyStyleOptions(t, "\"fields\""))
			rule_testing.RunWithOptions(t, ClassLiteralPropertyStyle,
				classLiteralPropertyStyleFile, sourceText,
				decodeClassLiteralPropertyStyleOptions(t, "\"getters\""))
		})
	}
}

// TestClassLiteralPropertyStyleDeclinesAConversionThatCannotCompile is where cohere is deliberately
// quieter than ESLint.
//
// Upstream skips a member written `override`, because converting an overriding member changes what
// it overrides, but it only sees the KEYWORD. Without `noImplicitOverride` an override carries no
// keyword, and upstream reports a conversion the compiler rejects: a property overriding a concrete
// base ACCESSOR is TS2610, and an accessor overriding a concrete base PROPERTY is TS2611 (the
// vendored checker's checkKindsOfPropertyMemberOverrides). cohere asks the checker for the base member
// and stays silent exactly there.
//
// The first silent row is the real site: `nexus/source/validation/schema/StringSchema.ts:37` in ahra,
// a `typeDefault` getter returning the empty string, overriding `BaseSchema`'s concrete getter.
// ESLint reports it and its suggested field does not compile. The same class's `typeName`, which
// implements an ABSTRACT getter, is a true positive (a field may implement an abstract accessor) and
// stays in the reporting rows, as do overrides of abstract and interface-merged members, a member
// with no base counterpart, and a static override, which the checker does not compare.
func TestClassLiteralPropertyStyleDeclinesAConversionThatCannotCompile(t *testing.T) {
	t.Parallel()

	const schemaBase = "abstract class BaseSchema<TInput = unknown, TOutput = TInput> {\n" +
		"    abstract get typeName(): string;\n" +
		"    get typeDefault(): TOutput | undefined { return undefined; }\n" +
		"}\n"

	cases := []struct {
		name          string
		configuration string
		sourceText    string
		wantIds       []string
	}{
		{"StringSchema.ts:37, a getter over a concrete base getter, beside a true positive", "",
			schemaBase + "class StringSchema extends BaseSchema<string> {\n" +
				"    get typeName(): string { return 'string'; }\n" +
				"    get typeDefault(): string { return ''; }\n" +
				"}\n",
			[]string{"preferFieldStyle"}},
		{"a getter over a concrete getter-and-setter pair", "",
			"class A { get x(): number { return Math.PI; } set x(value: number) {} }\nclass B extends A { get x(): number { return 1; } }\n",
			nil},
		{"a getter over a concrete getter two classes up", "",
			"class A { get x(): number { return Math.PI; } }\nclass B extends A {}\nclass C extends B { get x(): number { return 1; } }\n",
			nil},
		{"a field over a concrete base property, getters style", "\"getters\"",
			"class A { readonly x: number = Math.PI; }\nclass B extends A { readonly x = 1; }\n",
			nil},
		{"a getter over an abstract getter", "",
			"abstract class A { abstract get x(): number; }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a getter over an abstract property", "",
			"abstract class A { abstract readonly x: number; }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a getter over a base property, which the field restores", "",
			"class A { readonly x: number = Math.PI; }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a getter with no base counterpart", "",
			"class A { get y(): number { return Math.PI; } }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a static getter over a static base getter", "",
			"class A { static get x(): number { return Math.PI; } }\nclass B extends A { static get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a field over an abstract property, getters style", "\"getters\"",
			"abstract class A { abstract readonly x: number; }\nclass B extends A { readonly x = 1; }\n",
			[]string{"preferGetterStyle"}},
		{"a field over a base getter, which the getter restores, getters style", "\"getters\"",
			"class A { get x(): number { return Math.PI; } }\nclass B extends A { readonly x = 1; }\n",
			[]string{"preferGetterStyle"}},
		// A static member is compared only against the base's static side, which the checker never
		// checks for kind, so an instance getter of the same name in the base is no obstacle.
		{"a static getter beside a base instance getter of the same name", "",
			"class A { get x(): number { return Math.PI; } }\nclass B extends A { static get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		// Members merged in from an interface are exempt from the kind check, accessor or not.
		{"a getter over an interface-merged accessor", "",
			"interface A { get x(): number; }\nclass A {}\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a field over an interface-merged property, getters style", "\"getters\"",
			"interface A { readonly x: number; }\nclass A {}\nclass B extends A { readonly x = 1; }\n",
			[]string{"preferGetterStyle"}},
		// The three below are already compile errors before any conversion (a getter over a method,
		// and a private on either side of a public), and the checker skips them when it compares
		// kinds, so the conversion adds no error and upstream's report stands.
		{"a getter over a base method", "",
			"class A { x(): number { return Math.PI; } }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		{"a private getter over a public base getter", "",
			"class A { get x(): number { return Math.PI; } }\nclass B extends A { private get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
		// The checker skips a base property that came from a mapped type, so the getter compiles.
		{"a field over a mapped-type base property, getters style", "\"getters\"",
			"type Shape = { [K in 'x']: number };\ndeclare const Base: new () => Shape;\nclass B extends Base { readonly x = 1; }\n",
			[]string{"preferGetterStyle"}},
		// An intersection base property is exempt when ANY of its declarations is abstract or from an
		// interface, where an ordinary one needs all of them; the checker accepts the getter.
		{"a field over an intersection property, one side an interface, getters style", "\"getters\"",
			"interface I { readonly x: number; }\ndeclare class C { readonly x: number; }\ndeclare const Base: new () => I & C;\nclass D extends Base { readonly x = 1; }\n",
			[]string{"preferGetterStyle"}},
		{"a getter over a private base getter", "",
			"class A { private get x(): number { return Math.PI; } }\nclass B extends A { get x(): number { return 1; } }\n",
			[]string{"preferFieldStyle"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ClassLiteralPropertyStyle, classLiteralPropertyStyleFile,
				testCase.sourceText, decodeClassLiteralPropertyStyleOptions(t, testCase.configuration))
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The decoder reads empty input and `null` as upstream's default and each of upstream's two styles as
// itself, and refuses anything else, naming the value (#rfbha44). Each refused value used to decode to
// the default, so a config naming one loaded clean and ran a style nobody wrote; upstream's schema is a
// string enum and refuses every one at load.
func TestClassLiteralPropertyStyleDecoderRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, accepted := range []struct {
		wire string
		want ClassLiteralPropertyStyleSetting
	}{
		{"", ClassLiteralPropertyStyleFields},
		{"null", ClassLiteralPropertyStyleFields},
		{`"fields"`, ClassLiteralPropertyStyleFields},
		{`"getters"`, ClassLiteralPropertyStyleGetters},
	} {
		decoded, err := DecodeClassLiteralPropertyStyleOptions([]byte(accepted.wire))
		if err != nil {
			t.Errorf("%s was refused: %v", accepted.wire, err)
			continue
		}
		if got := decoded.(ClassLiteralPropertyStyleOptions).Style; got != accepted.want {
			t.Errorf("%s decoded to %s, want %s", accepted.wire, got, accepted.want)
		}
	}
	for _, refused := range []struct{ wire, named string }{
		{`"getter"`, "getter"},
		{`"Fields"`, "Fields"},
		{`""`, `""`},
		{`{"style": "getters"}`, "style"},
		{`1`, "1"},
	} {
		_, err := DecodeClassLiteralPropertyStyleOptions([]byte(refused.wire))
		if err == nil {
			t.Errorf("%s decoded; it must be refused", refused.wire)
			continue
		}
		if !strings.Contains(err.Error(), refused.named) {
			t.Errorf("%s was refused without naming %s: %v", refused.wire, refused.named, err)
		}
	}
}

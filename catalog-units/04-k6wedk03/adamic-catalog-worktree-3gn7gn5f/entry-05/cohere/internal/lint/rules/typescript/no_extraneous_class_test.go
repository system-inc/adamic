package typescript

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noExtraneousClassFile names the fixture file.
//
// The rule reads no path and gates on no extension: upstream registers a bare ClassBody visitor.
// It ends in .ts rather than .tsx because several cases use a decorator, and a decorator on a class
// is unambiguous in either, while keeping one extension across the file means a failure is never
// about which one a case landed in.
const noExtraneousClassFile = "/repository/source/Classes.ts"

// noExtraneousClassCaseName numbers a row so a failure names which one.
func noExtraneousClassCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodeNoExtraneousClassOptionsForTest routes a fixture through the rule's own decoder.
//
// All four of this rule's keys default to false, so a decoder that dropped every one of them would
// still produce the right answer for an absent key and every fixture built from a struct would pass.
// Going through the decoder is what puts the wire shape under test: cohere strips ESLint's
// [severity, options] tuple, so what arrives is the bare object rather than a one-element array.
func decodeNoExtraneousClassOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoExtraneousClassOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// runNoExtraneousClass runs one case, with or without options.
func runNoExtraneousClass(t *testing.T, sourceText string, options string) rule_testing.Result {
	t.Helper()
	if options == "" {
		return rule_testing.Run(t, NoExtraneousClass, noExtraneousClassFile, sourceText)
	}
	return rule_testing.RunWithOptions(t, NoExtraneousClass, noExtraneousClassFile, sourceText,
		decodeNoExtraneousClassOptionsForTest(t, options))
}

// TestNoExtraneousClassStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All sixteen of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it, then byte verified. Every one was replayed
// through the installed 8.x build, which reported nothing on all sixteen.
//
// Six of these are the shapes our parser spells differently from upstream's and they are the reason
// this list earns its place. Upstream has a distinct node for an auto-accessor and separate nodes for
// every abstract member; ours gives all of them KindPropertyDeclaration or KindMethodDeclaration and
// puts the difference in a modifier, so the arm that keeps a class with only an accessor or only an
// abstract member from reporting has to read modifiers where upstream reads a node type.
func TestNoExtraneousClassStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{sourceText: "\nclass Foo {\n  public prop = 1;\n  constructor() {}\n}\n    ", options: ""},
		{sourceText: "\nexport class CClass extends BaseClass {\n  public static helper(): void {}\n  private static privateHelper(): boolean {\n    return true;\n  }\n  constructor() {}\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  constructor(public bar: string) {}\n}\n    ", options: ""},
		{sourceText: "class Foo {}", options: "{\"allowEmpty\": true}"},
		{sourceText: "\nclass Foo {\n  constructor() {}\n}\n      ", options: "{\"allowConstructorOnly\": true}"},
		{sourceText: "\nexport class Bar {\n  public static helper(): void {}\n  private static privateHelper(): boolean {\n    return true;\n  }\n}\n      ", options: "{\"allowStaticOnly\": true}"},
		{sourceText: "\nexport default class {\n  hello() {\n    return 'I am foo!';\n  }\n}\n    ", options: ""},
		{sourceText: "\n@FooDecorator\nclass Foo {}\n      ", options: "{\"allowWithDecorator\": true}"},
		{sourceText: "\n@FooDecorator\nclass Foo {\n  constructor(foo: Foo) {\n    foo.subscribe(a => {\n      console.log(a);\n    });\n  }\n}\n      ", options: "{\"allowWithDecorator\": true}"},
		{sourceText: "\nabstract class Foo {\n  abstract property: string;\n}\n    ", options: ""},
		{sourceText: "\nabstract class Foo {\n  abstract method(): string;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  accessor prop: string;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  accessor prop = 'bar';\n  static bar() {\n    return false;\n  }\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  [key: string]: string;\n}\n    ", options: ""},
		{sourceText: "\nabstract class Foo {\n  accessor prop: string;\n}\n    ", options: ""},
		{sourceText: "\nabstract class Foo {\n  abstract accessor prop: string;\n}\n    ", options: ""},
	}
	for index, testCase := range cases {
		t.Run(noExtraneousClassCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoExtraneousClass(t, testCase.sourceText, testCase.options))
		})
	}
}

// TestNoExtraneousClassFiresOnUpstreamFailCases is the imported failing corpus, verbatim, with the
// span of every finding asserted.
//
// The span is doing real work on this rule. Upstream reports on the class NAME when there is one and
// on the whole class otherwise, so an anonymous default export is reported across its entire body
// while a named class beside it is reported on three characters. Both shapes are in this table, and a
// port anchoring uniformly either way satisfies every message id here and points somewhere upstream
// never points.
//
// Two rows report twice, which is how a nested class is handled: the visitor fires per class body, so
// a class declared inside a constructor is judged on its own and reported independently of the class
// holding it.
func TestNoExtraneousClassFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "class Foo {}",
			options:    "",
			wantIds:    []string{"empty"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nclass Foo {\n  public prop = 1;\n  constructor() {\n    class Bar {\n      static PROP = 2;\n    }\n  }\n}\nexport class Bar {\n  public static helper(): void {}\n  private static privateHelper(): boolean {\n    return true;\n  }\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic", "onlyStatic"},
			wantSpans:  []string{"Bar", "Bar"},
		},
		{
			sourceText: "\nclass Foo {\n  constructor() {}\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyConstructor"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nexport class AClass {\n  public static helper(): void {}\n  private static privateHelper(): boolean {\n    return true;\n  }\n  constructor() {\n    class nestedClass {}\n  }\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic", "empty"},
			wantSpans:  []string{"AClass", "nestedClass"},
		},
		{
			sourceText: "\nexport default class {\n  static hello() {}\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"class {\n  static hello() {}\n}"},
		},
		{
			sourceText: "\n@FooDecorator\nclass Foo {}\n      ",
			options:    "{\"allowWithDecorator\": false}",
			wantIds:    []string{"empty"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\n@FooDecorator\nclass Foo {\n  constructor(foo: Foo) {\n    foo.subscribe(a => {\n      console.log(a);\n    });\n  }\n}\n      ",
			options:    "{\"allowWithDecorator\": false}",
			wantIds:    []string{"onlyConstructor"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nabstract class Foo {}\n      ",
			options:    "",
			wantIds:    []string{"empty"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nabstract class Foo {\n  static property: string;\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nabstract class Foo {\n  constructor() {}\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyConstructor"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nclass Foo {\n  static accessor prop: string;\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			sourceText: "\nabstract class Foo {\n  static accessor prop: string;\n}\n      ",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
	}
	for index, testCase := range cases {
		t.Run(noExtraneousClassCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := runNoExtraneousClass(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestNoExtraneousClassOnShapesUpstreamsCorpusDoesNotWrite covers the classifications upstream's
// twenty-eight cases leave unexercised.
//
// Upstream reads ESTree node types to classify a member and a class; ours reads kinds plus
// modifiers, and the two do not line up one to one. So the arms most likely to be wrong here are
// exactly the ones the imported corpus cannot reach: it writes no implements clause, no class
// expression, no static block, no private field, no getter, no decorated parameter, and no named
// class expression.
//
// Every verdict and every span below was measured against the installed 8.x build on that exact
// source, with a reporting control in the same run. Two of them changed this port: the named class
// expression, and the decorated parameter.
func TestNoExtraneousClassOnShapesUpstreamsCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		options    string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "implements-does-not-exempt",
			why:        "an implements clause is NOT a superclass and does not exempt; our parser puts both in one heritage list separated only by a token, so a check that merely asked whether a heritage list existed would silently exempt every class implementing an interface, and upstream's corpus writes no implements clause",
			sourceText: "class Foo implements Bar {\n  static helper(): void {}\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "extends-exempts",
			why:        "the extends half, so the token test is measured in both directions",
			sourceText: "class Foo extends Bar {\n  static helper(): void {}\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "extends-and-implements",
			why:        "both clauses at once, which is the arrangement where a loop over the list has to keep looking rather than judging the first entry",
			sourceText: "class Foo extends Bar implements Baz {\n  static helper(): void {}\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "class-expression",
			why:        "an anonymous class expression, reported across the whole expression",
			sourceText: "const Foo = class {\n  static helper(): void {}\n};\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"class {\n  static helper(): void {}\n}"},
		},
		{
			name:       "named-class-expression",
			why:        "a NAMED class expression, which upstream still reports across the whole node because its anchor test requires the node be a declaration before it consults the name. Upstream writes no named class expression anywhere, so nothing imported can see a port that reads the name here",
			sourceText: "const Foo = class Named {\n  static helper(): void {}\n};\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"class Named {\n  static helper(): void {}\n}"},
		},
		{
			name:       "static-block",
			why:        "a static initialization block, which is neither static-clearing nor a constructor, so the class stays static-only and reports",
			sourceText: "class Foo {\n  static {\n    console.log(1);\n  }\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "static-private-field",
			why:        "a static private field, which upstream's static test reads the same as any other static member",
			sourceText: "class Foo {\n  static #secret = 1;\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "instance-private-field",
			why:        "the instance half, which clears static-only and stays silent",
			sourceText: "class Foo {\n  #secret = 1;\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "static-getter",
			why:        "a static getter, one of the two accessor kinds our parser gives its own node kind while upstream files them under MethodDefinition",
			sourceText: "class Foo {\n  static get value() {\n    return 1;\n  }\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "instance-getter",
			why:        "the instance half of the row above",
			sourceText: "class Foo {\n  get value() {\n    return 1;\n  }\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "decorated-empty-without-option",
			why:        "a decorated class WITHOUT the option, which reports; the exemption is behind the option rather than automatic",
			sourceText: "@Dec\nclass Foo {}\n",
			options:    "",
			wantIds:    []string{"empty"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "decorated-static-with-option",
			why:        "the same class with the option on, which is exempt whatever its members hold",
			sourceText: "@Dec\nclass Foo {\n  static helper(): void {}\n}\n",
			options:    "{\"allowWithDecorator\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "decorated-with-superclass",
			why:        "a decorated subclass with the option off, silent on the superclass rather than on the decorator, which is the row separating the two exemptions",
			sourceText: "@Dec\nclass Foo extends Bar {}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "constructor-with-readonly-parameter",
			why:        "a readonly parameter property, which our parser spells as a modifier on a plain parameter rather than as its own node type, and which declares real state so the class is not extraneous",
			sourceText: "class Foo {\n  constructor(readonly bar: string) {}\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "constructor-with-decorated-parameter",
			why:        "a DECORATED parameter, which is not a parameter property and does not exempt; this is why the modifier test names the accessibility keywords rather than asking whether the modifier list is non-empty",
			sourceText: "class Foo {\n  constructor(@Inject() bar: string) {}\n}\n",
			options:    "",
			wantIds:    []string{"onlyConstructor"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "constructor-with-plain-parameter",
			why:        "a plain parameter, the baseline the two rows above are read against",
			sourceText: "class Foo {\n  constructor(bar: string) {}\n}\n",
			options:    "",
			wantIds:    []string{"onlyConstructor"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "static-index-signature",
			why:        "a static index signature, which upstream clears static-only for unconditionally, so the class is neither static-only nor constructor-only and stays silent. Confirmed in the upstream parser's own output that it is still an index signature node when written static",
			sourceText: "class Foo {\n  static [key: string]: string;\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "declare-static-property",
			why:        "a declare static property, which is a static member for this rule's purposes",
			sourceText: "class Foo {\n  declare static bar: number;\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "all-four-allowed-empty",
			why:        "all four options set at once on the shape the first of them covers, which is the only row where more than one key is present in the wire object",
			sourceText: "class Foo {}\n",
			options:    "{\"allowEmpty\": true, \"allowStaticOnly\": true, \"allowConstructorOnly\": true, \"allowWithDecorator\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-nec",
			why:        "the control: the plainest reporting shape, so a rule that had stopped reporting would still be visible in this table",
			sourceText: "class Foo {\n  static helper(): void {}\n}\n",
			options:    "",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoExtraneousClass(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestNoExtraneousClassDecoderKeepsAnAbsentKeyDistinctFromAnExplicitFalse pins the four lines of the
// decoder that have no upstream counterpart.
//
// Every one of this rule's keys defaults to false, so a decoder that dropped all four would produce
// the right answer for every absent key and every fixture above would still pass. This is the only
// thing in the file that can tell a decoder that reads its input from one that does not.
func TestNoExtraneousClassDecoderKeepsAnAbsentKeyDistinctFromAnExplicitFalse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want NoExtraneousClassOptions
	}{
		{raw: `{}`, want: NoExtraneousClassOptions{}},
		{raw: `{"allowEmpty": true}`, want: NoExtraneousClassOptions{AllowEmpty: true}},
		{raw: `{"allowEmpty": false}`, want: NoExtraneousClassOptions{}},
		{raw: `{"allowStaticOnly": true}`, want: NoExtraneousClassOptions{AllowStaticOnly: true}},
		{raw: `{"allowConstructorOnly": true}`, want: NoExtraneousClassOptions{AllowConstructorOnly: true}},
		{raw: `{"allowWithDecorator": true}`, want: NoExtraneousClassOptions{AllowWithDecorator: true}},
		{
			raw: `{"allowEmpty": true, "allowStaticOnly": true, "allowConstructorOnly": true, "allowWithDecorator": true}`,
			want: NoExtraneousClassOptions{
				AllowConstructorOnly: true,
				AllowEmpty:           true,
				AllowStaticOnly:      true,
				AllowWithDecorator:   true,
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded := decodeNoExtraneousClassOptionsForTest(t, testCase.raw)
			settings, isSettings := decoded.(NoExtraneousClassOptions)
			if !isSettings {
				t.Fatalf("the decoder returned %T rather than the rule's own options type", decoded)
			}
			if settings != testCase.want {
				t.Errorf("decoded to %+v, wanted %+v", settings, testCase.want)
			}
		})
	}
}

// TestNoExtraneousClassFallsBackWhenHandedNilOptions covers the path a rule configured as a bare
// severity string takes.
//
// A rule named as "error" with no object is handed nil, which the type assertion in Run cannot
// satisfy, so the fallback is the only thing between that configuration and a zero value. Every
// other fixture here reaches the rule through the decoder and none can see this line.
func TestNoExtraneousClassFallsBackWhenHandedNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoExtraneousClass, noExtraneousClassFile,
		"class Foo {}\n", nil)
	rule_testing.ExpectFindings(t, result, "empty")
}

// TestNoExtraneousClassRendersUpstreamsMessageText asserts what a reader is told, for all three.
//
// rule.Message is {Id, Description} with no interpolation, so there is nothing to render; what there
// is to get wrong is the text, and ExpectFindings compares ids and count and nothing else. The
// wanted strings are typed as literals rather than read from the rule's own constants, because a
// comparison against the constant moves with any mutation of it.
func TestNoExtraneousClassRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantId      string
		wantMessage string
	}{
		{
			sourceText:  "class Foo {}\n",
			wantId:      "empty",
			wantMessage: "Unexpected empty class.",
		},
		{
			sourceText:  "class Foo {\n  constructor() {}\n}\n",
			wantId:      "onlyConstructor",
			wantMessage: "Unexpected class with only a constructor.",
		},
		{
			sourceText:  "class Foo {\n  static helper(): void {}\n}\n",
			wantId:      "onlyStatic",
			wantMessage: "Unexpected class with only static properties.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.wantId, func(t *testing.T) {
			t.Parallel()
			result := runNoExtraneousClass(t, testCase.sourceText, "")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0]
			if reported.Message.Id != testCase.wantId {
				t.Errorf("reported id %q, wanted %q", reported.Message.Id, testCase.wantId)
			}
			if reported.Message.Description != testCase.wantMessage {
				t.Errorf("reported message %q, wanted %q",
					reported.Message.Description, testCase.wantMessage)
			}
		})
	}
}

// TestNoExtraneousClassOnIllegalStaticAbstractMembers pins the arm that reproduces upstream's
// separate abstract node types.
//
// That arm reads as dead code. An abstract member cannot be static in TypeScript, upstream says so
// at the line and cites the compiler issue, and our own static test would therefore cover every
// abstract shape on its own. Neutralizing the arm survived the whole fixture set, which is what sent
// this to a parse probe rather than to an equivalence argument.
//
// The parser recovers. `static abstract x: number` comes back as a property carrying BOTH modifiers,
// in either written order, for properties, methods and accessors alike. So the arm is reachable, and
// it is load bearing: upstream is silent on all four of these and reports on the legal control in
// the same run, so a port without the arm reports where upstream does not. The grammar says one
// thing and the parser does another, and only the parser decides what a rule sees.
func TestNoExtraneousClassOnIllegalStaticAbstractMembers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "static-abstract-property",
			why:        "static abstract is not legal TypeScript, and our parser RECOVERS from it rather than refusing the member: the property comes back carrying both modifiers, which makes upstream's separate abstract node types reachable here through error recovery rather than dead. Without the abstract arm this reports and upstream does not",
			sourceText: "abstract class Foo {\n  static abstract x: number;\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "abstract-static-property",
			why:        "the same two modifiers in the other order, since a port keying on which came first would pass the row above and fail this one",
			sourceText: "abstract class Foo {\n  abstract static x: number;\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "static-abstract-method",
			why:        "the method form, which is a different node kind here and reaches the same arm",
			sourceText: "abstract class Foo {\n  static abstract m(): void;\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "static-abstract-accessor",
			why:        "the accessor form, which our parser files as a property with two extra modifiers",
			sourceText: "abstract class Foo {\n  static abstract accessor x: number;\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-abstract",
			why:        "the control: the same class with a legal static member, which reports",
			sourceText: "abstract class Foo {\n  static x: number;\n}\n",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoExtraneousClass(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestNoExtraneousClassSeparatesADecoratorFromOtherModifiers pins the decorator search.
//
// ESTree gives a class its decorators in their own array, so upstream reads one field and cannot
// confuse a decorator with anything else. Our parser puts decorators, `export`, `default`, `declare`
// and `abstract` in ONE modifier list, so the search has to name the kind. A search that merely
// asked whether the list was non-empty survives every imported case, because upstream pairs the
// option only with classes that really are decorated.
//
// Measured against the installed 8.x build with a control in the same run.
func TestNoExtraneousClassSeparatesADecoratorFromOtherModifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "exported-class-with-decorator-option",
			why:        "an EXPORTED class with the decorator option on, which still reports. Our parser puts export in the same modifier list as a decorator, so a decorator search asking only whether the list is non-empty exempts this and upstream does not; upstream's corpus pairs the option only with genuinely decorated classes, so nothing imported can see it",
			sourceText: "export class Foo {\n  static helper(): void {}\n}\n",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "abstract-class-with-decorator-option",
			why:        "the same through the abstract modifier, which is the other keyword our parser files beside a decorator",
			sourceText: "abstract class Foo {\n  static helper(): void {}\n}\n",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
		{
			name:       "decorated-class-with-decorator-option",
			why:        "a class carrying BOTH a decorator and export, which is exempt; the row that keeps the two rows above from being satisfied by a search that never matches anything",
			sourceText: "@Dec\nexport class Foo {\n  static helper(): void {}\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-e24",
			why:        "the control: no modifiers at all with the option on, which reports",
			sourceText: "class Foo {\n  static helper(): void {}\n}\n",
			wantIds:    []string{"onlyStatic"},
			wantSpans:  []string{"Foo"},
		},
	}
	const options = `{"allowWithDecorator": true}`
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoExtraneousClass(t, testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestNoExtraneousClassExemptsALoadBearingEmptyClass pins cohere's divergence from typescript-eslint, by
// @system_cohere's ruling of 2026-10-03 (#ynneze5, Base's tests): an empty class is not extraneous when
// another class extends it, or when every value use of it goes where a constructor is expected. Each
// silent row has a reported neighbour, so a mutant dropping or widening either shape fails here.
func TestNoExtraneousClassExemptsALoadBearingEmptyClass(t *testing.T) {
	t.Parallel()

	slots := "type Constructor = new (...args: any[]) => object;\n" +
		"declare function register(target: Constructor): void;\n" +
		"declare function key<Target extends new () => object>(target: Target): void;\n" +
		"declare const scope: ReadonlySet<new () => object>;\n"
	silent := []struct {
		name   string
		source string
	}{
		{"a base another class extends", "class Base {}\nclass Child extends Base {\n  value = 1;\n}\nnew Child();\n"},
		{"a constructor-typed parameter", slots + "class Token {}\nregister(Token);\n"},
		{"a set of constructors", slots + "class Unrelated {}\nscope.has(Unrelated);\n"},
		{"every one of several uses is a constructor slot", slots + "class Token {}\nregister(Token);\nscope.has(Token);\n"},
		{"a use as a type does not count against it", slots + "class Token {}\nregister(Token);\ndeclare const typed: Token;\n"},
		{"a class expression inline where a constructor is expected", "declare function snapshot(entry: { target: new () => object }): void;\nsnapshot({ target: class Target {} });\n"},
		{"a const-bound class expression handed only to constructor slots", slots + "const token = class Token {};\nregister(token);\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtraneousClass, noExtraneousClassFile, testCase.source, nil))
		})
	}

	reported := []struct {
		name   string
		source string
	}{
		{"nothing references it", "class Foo {}\n"},
		{"a value use that wants no constructor", "class Token {}\nconsole.log(Token);\n"},
		{"one use in a constructor slot and one that is not", slots + "class Token {}\nregister(Token);\nconst alias = Token;\nconsole.log(alias);\n"},
		{"instantiated, which needs no class", "class Foo {}\nconsole.log(new Foo());\n"},
		{"exported, so uses elsewhere decide it", slots + "export class Token {}\nregister(Token);\n"},
		{"exported by a specifier", slots + "class Token {}\nregister(Token);\nexport { Token };\n"},
		{"extended by an interface, which uses it only as a type", "class Shape {}\ninterface Square extends Shape {\n  side: number;\n}\ndeclare const square: Square;\nconsole.log(square);\n"},
		{"a let-bound class expression that is reassigned, since the write wants no constructor", slots + "let token = class Token {};\nregister(token);\ntoken = class Other {};\n"},
		{"a const-bound class expression also used otherwise", slots + "const token = class {};\nregister(token);\nconsole.log(token);\n"},
		{"a class expression inline where no constructor is expected", "console.log(class {});\n"},
		{"an exported const-bound class expression", slots + "export const token = class {};\nregister(token);\n"},
		{"a generic parameter, whose contextual type is inferred from the class itself", slots + "declare function rename<Target>(target: Target, name: string): Target;\nclass Token {}\nrename(Token, 'renamed');\n"},
		{"a generic constrained to a constructor, which inference can't tell apart", slots + "class Token {}\nkey(Token);\n"},
		{"implemented rather than extended", "class Shape {}\nclass Square implements Shape {\n  side = 1;\n}\nnew Square();\n"},
	}
	for _, testCase := range reported {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoExtraneousClass, noExtraneousClassFile, testCase.source, nil), "empty")
		})
	}
}

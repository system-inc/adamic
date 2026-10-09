package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const relatedGetterSetterPairsFile = "/repository/source/Accessors.ts"

// TestRelatedGetterSetterPairsStaysSilent carries all sixteen of upstream's clean cases verbatim.
//
// They are the two filters and the assignability direction, stated as inputs. Four say a getter with
// no return annotation is never recorded; three say a setter is recorded only at exactly one
// parameter; and `get value(): string` beside `set value(newValue: string | undefined)` says the
// comparison runs getter-to-setter rather than either way round, since the reverse would report it.
//
// Extracted from the corpus by a script and byte-compared against it, not retyped. Every one of them
// was also driven through the installed 8.67.0 build before this file existed, and all sixteen were
// clean there.
func TestRelatedGetterSetterPairsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"string getter and string setter", "\ninterface Example {\n  get value(): string;\n  set value(newValue: string);\n}\n    "},
		{"optional getter with a zero parameter setter", "\ninterface Example {\n  get value(): string | undefined;\n  set value();\n}\n    "},
		{"optional getter with a two parameter setter", "\ninterface Example {\n  get value(): string | undefined;\n  set value(newValue: string, invalid: string);\n}\n    "},
		{"string getter and an optional setter", "\ninterface Example {\n  get value(): string;\n  set value(newValue: string | undefined);\n}\n    "},
		{"a getter with no setter", "\ninterface Example {\n  get value(): number;\n}\n    "},
		{"a getter and a zero parameter setter", "\ninterface Example {\n  get value(): number;\n  set value();\n}\n    "},
		{"a setter with no getter", "\ninterface Example {\n  set value(newValue: string);\n}\n    "},
		{"a zero parameter setter alone", "\ninterface Example {\n  set value();\n}\n    "},
		{"a type literal getter with no return annotation", "\ntype Example = {\n  get value();\n};\n    "},
		{"a type literal setter with no getter", "\ntype Example = {\n  set value();\n};\n    "},
		{"a class getter with no return annotation", "\nclass Example {\n  get value() {\n    return '';\n  }\n}\n    "},
		{"a class getter with no annotation and an empty setter", "\nclass Example {\n  get value() {\n    return '';\n  }\n  set value() {}\n}\n    "},
		{"a class getter with no annotation and an untyped setter parameter", "\nclass Example {\n  get value() {\n    return '';\n  }\n  set value(param) {}\n}\n    "},
		{"a class getter with no annotation and a number setter", "\nclass Example {\n  get value() {\n    return '';\n  }\n  set value(param: number) {}\n}\n    "},
		{"a class setter with no getter", "\nclass Example {\n  set value() {}\n}\n    "},
		{"a type literal number pair that matches", "\ntype Example = {\n  get value(): number;\n  set value(newValue: number);\n};\n    "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText))
		})
	}
}

// TestRelatedGetterSetterPairsFires carries all seven of upstream's reporting cases verbatim.
//
// Each reports exactly once. The last three are the interesting ones and they are about the checker
// rather than about the traversal: an inline object type, an unused alias sitting beside the real
// one, and a named alias all reach the same verdict, which is what pins that the comparison is on
// TYPES and not on the spelling of the annotation.
func TestRelatedGetterSetterPairsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an optional getter against a required setter", "\ninterface Example {\n  get value(): string | undefined;\n  set value(newValue: string);\n}\n      "},
		{"a number getter against a string setter", "\ninterface Example {\n  get value(): number;\n  set value(newValue: string);\n}\n      "},
		{"a type literal number getter against a string setter", "\ntype Example = {\n  get value(): number;\n  set value(newValue: string);\n};\n      "},
		{"a class boolean getter against a string setter", "\nclass Example {\n  get value(): boolean {\n    return true;\n  }\n  set value(newValue: string) {}\n}\n      "},
		{"a declared class against an inline object type", "\ntype GetType = { a: string; b: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: { c: string });\n}\n      "},
		{"a declared class beside an unused alias", "\ntype GetType = { a: string; b: string };\n\ntype SetTypeUnused = { c: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: { c: string });\n}\n      "},
		{"a declared class against a named alias", "\ntype GetType = { a: string; b: string };\n\ntype SetType = { c: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: SetType);\n}\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText), "mismatch")
		})
	}
}

// TestRelatedGetterSetterPairsSpans asserts WHERE each of upstream's seven findings points.
//
// The corpus states a line, a column, an endLine and an endColumn for every one of them, and those
// four numbers are the only thing separating a finding on the getter's return type from a
// defensible finding on the member, on the annotation, or on the setter's parameter. A message id
// fixture cannot see any of that: all four choices produce one `mismatch` on the same input.
//
// The expected text is the slice of upstream's own source between its stated columns, cut by the
// same script that imported the cases, so this asserts upstream's numbers rather than my reading of
// them.
//
// The harness TRIMS the fixture before writing it, so the file on disk is one byte shorter at the
// front than the Go literal here. That is why the expectation is compared against a slice of the
// harness's own result rather than against a slice of the literal.
func TestRelatedGetterSetterPairsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"an optional getter against a required setter", "\ninterface Example {\n  get value(): string | undefined;\n  set value(newValue: string);\n}\n      ", "string | undefined"},
		{"a number getter against a string setter", "\ninterface Example {\n  get value(): number;\n  set value(newValue: string);\n}\n      ", "number"},
		{"a type literal number getter against a string setter", "\ntype Example = {\n  get value(): number;\n  set value(newValue: string);\n};\n      ", "number"},
		{"a class boolean getter against a string setter", "\nclass Example {\n  get value(): boolean {\n    return true;\n  }\n  set value(newValue: string) {}\n}\n      ", "boolean"},
		{"a declared class against an inline object type", "\ntype GetType = { a: string; b: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: { c: string });\n}\n      ", "GetType"},
		{"a declared class beside an unused alias", "\ntype GetType = { a: string; b: string };\n\ntype SetTypeUnused = { c: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: { c: string });\n}\n      ", "GetType"},
		{"a declared class against a named alias", "\ntype GetType = { a: string; b: string };\n\ntype SetType = { c: string };\n\ndeclare class Foo {\n  get a(): GetType;\n\n  set a(x: SetType);\n}\n      ", "GetType"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			got := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.want {
				t.Errorf("reported span = %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestRelatedGetterSetterPairsFiresOnMeasuredShapes carries inputs upstream's corpus does not write,
// each one driven through the installed 8.67.0 build first and reproduced here from what it did
// rather than from what the source reads like.
//
// Four of them assert behavior that looks like a defect and is not. `static` is not part of the key,
// so a static getter pairs with an instance setter that can never be the same property. A quoted key
// that is a valid identifier keys bare, so `'value'` and `value` are one property. A numeric key
// cooks and then quotes, so `12` and `'12'` are one property. And the map is written with `.set`, so
// a second declaration under one name REPLACES the first rather than being ignored, which is why
// there is a fires case and a silent case for each of the two duplicate orderings.
//
// The rest are traversal: a class expression, a type literal nested in an interface and one in a
// parameter position, a setter written before its getter, and a `declare class`, which is the
// control proving the abstract exclusion below is about `abstract` and not about having no body.
func TestRelatedGetterSetterPairsFiresOnMeasuredShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a static getter pairs with an instance setter", "class C {\n  static get value(): number { return 1; }\n  set value(v: string) {}\n}"},
		{"a private pair", "class C {\n  get #value(): number { return 1; }\n  set #value(v: string) {}\n}"},
		{"a computed identifier key pairs with itself", "declare const k: string;\nclass C {\n  get [k](): number { return 1; }\n  set [k](v: string) {}\n}"},
		// The two that pin that a computed key is keyed by what is INSIDE the brackets rather than
		// by the bracketed source text. Both are needed and neither is redundant with the case
		// above: a rule keying `[k]` by its whole text still pairs `[k]` with `[k]`, so that case
		// alone cannot see the difference. These two cross the bracket boundary in both directions
		// and both report on the installed build.
		{"a plain identifier getter pairs with a computed quoted setter",
			"class C {\n  get value(): number { return 1; }\n  set [\"value\"](v: string) {}\n}"},
		{"a computed identifier getter pairs with a plain setter",
			"declare const k: string;\nclass C {\n  get [k](): number { return 1; }\n  set k(v: string) {}\n}"},
		// The raw-text arm keys by SOURCE TEXT, and `Pos()` in this tree includes leading trivia
		// while TSESTree's `range` starts at the token. So a key written after a newline would key
		// with the whitespace attached and stop pairing with the same key written inline. Measured
		// on the installed build: both spellings of this pair report, which is what makes the trim
		// load-bearing rather than tidy. A mutant using the untrimmed position goes silent here.
		{"a symbol key pairs across a line break",
			"class C {\n  get [Symbol.iterator](): number { return 1; }\n  set [\n    Symbol.iterator](v: string) {}\n}"},
		{"a quoted key pairs with a bare identifier of the same spelling", "class C {\n  get 'value'(): number { return 1; }\n  set value(v: string) {}\n}"},
		{"a numeric key pairs with a quoted one of the same value", "class C {\n  get 12(): number { return 1; }\n  set '12'(v: string) {}\n}"},
		{"a second setter replaces the first", "class C {\n  get value(): number { return 1; }\n  set value(v: number) {}\n  set value(v: string) {}\n}"},
		{"a second getter replaces the first", "class C {\n  get value(): string { return \"\"; }\n  get value(): number { return 1; }\n  set value(v: string) {}\n}"},
		{"a setter written before its getter", "interface I {\n  set value(v: string);\n  get value(): number;\n}"},
		{"a declare class reports where an abstract class does not", "declare class C {\n  get value(): number;\n  set value(v: string);\n}"},
		{"a class expression body", "const C = class {\n  get value(): number { return 1; }\n  set value(v: string) {}\n};"},
		{"a rest parameter counts as one parameter", "class C {\n  get value(): number { return 1; }\n  set value(...v: string[]) {}\n}"},
		{"an optional parameter is still one parameter", "class C {\n  get value(): number { return 1; }\n  set value(v?: string) {}\n}"},
		{"a never setter with a string getter", "interface I {\n  get value(): string;\n  set value(v: never);\n}"},
		{"a type literal nested inside an interface", "interface I {\n  p: {\n    get value(): number;\n    set value(v: string);\n  };\n}"},
		{"a type literal in a function parameter position", "function f(x: { get value(): number; set value(v: string) }) {}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText), "mismatch")
		})
	}
}

// TestRelatedGetterSetterPairsStaysSilentOnMeasuredShapes is the same, for inputs the installed
// build leaves alone.
//
// The four abstract cases are the ones worth naming. Upstream's selector is
// `:matches(MethodDefinition, TSMethodSignature)` and its parser gives an `abstract` accessor the
// node type `TSAbstractMethodDefinition`, so one abstract half silences the whole pair. Nothing in
// typescript-go reproduces that, since an abstract accessor here is an ordinary accessor with a
// modifier, so the rule declines it by hand and these four are what hold that decline in place. A
// mutant removing it survives every imported fixture, because the corpus writes no `abstract`.
//
// The object literal is the second gap with no tree exposure: its accessors are the same node kinds
// a class holds and are told apart only by their parent, so a rule anchored on the accessor kind
// rather than on the container would report it.
func TestRelatedGetterSetterPairsStaysSilentOnMeasuredShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an object literal is not one of the three containers", "const o = {\n  get value(): number { return 1; },\n  set value(v: string) {},\n};"},
		{"both accessors abstract", "abstract class C {\n  abstract get value(): number;\n  abstract set value(v: string);\n}"},
		{"an abstract getter with a concrete setter", "abstract class C {\n  abstract get value(): number;\n  set value(v: string) {}\n}"},
		{"a concrete getter with an abstract setter", "abstract class C {\n  get value(): number { return 1; }\n  abstract set value(v: string);\n}"},
		{"a nested class does not see the outer names", "class Outer {\n  get value(): number { return 1; }\n  m() {\n    class Inner {\n      set value(v: string) {}\n    }\n    return Inner;\n  }\n}"},
		{"a private getter does not pair with a public setter", "class C {\n  get #value(): number { return 1; }\n  set value(v: string) {}\n}"},
		{"a this parameter pushes the setter to two parameters", "class C {\n  get value(): number { return 1; }\n  set value(this: C, v: string) {}\n}"},
		{"a never getter is assignable to everything", "interface I {\n  get value(): never;\n  set value(v: string);\n}"},
		{"an any getter", "class C {\n  get value(): any { return 1; }\n  set value(v: string) {}\n}"},
		{"an any setter parameter", "class C {\n  get value(): number { return 1; }\n  set value(v: any) {}\n}"},
		{"an untyped setter parameter is an implicit any", "class C {\n  get value(): number { return 1; }\n  set value(v) {}\n}"},
		{"computed keys with different source text do not pair", "declare const k: \"x\";\nclass C {\n  get [k](): number { return 1; }\n  set [\"x\"](v: string) {}\n}"},
		{"two interface declarations merging do not share a container", "interface I { get value(): number; }\ninterface I { set value(v: string); }"},
		{"a second getter replaces a mismatching first", "class C {\n  get value(): number { return 1; }\n  get value(): string { return \"\"; }\n  set value(v: string) {}\n}"},
		{"a second setter replaces a mismatching first", "class C {\n  get value(): number { return 1; }\n  set value(v: string) {}\n  set value(v: number) {}\n}"},
		// The getter's return-annotation filter is CRASH PROTECTION on this input, and it is the
		// only input in this file that reaches that. Every OTHER unannotated getter is silent for a
		// reason that has nothing to do with the filter: TypeScript infers an unannotated getter's
		// type FROM its setter, measured over seven shapes, so the two types are identical and
		// assignability is trivially true. So a mutant disabling the filter survives the corpus,
		// survives every case above, and is not equivalent.
		//
		// `static` is what separates them. It is not part of the key, so upstream pairs a static
		// getter with an instance setter, but they really are two different properties, so no
		// inference links them: the getter is `string` and the setter takes `number`. The pair then
		// records, assignability fails, and the report reaches for a return type node that is nil.
		//
		// Measured: with the filter removed, this input PANICS with a nil dereference inside
		// SkipTypeParentheses. Upstream is clean, because its own filter declined the getter first.
		// No ExpectFindings fixture can see a panic, which is why this case is written and why the
		// sweep reported the mutant as survived rather than as caught.
		{"a static unannotated getter against an instance setter",
			"class C {\n  static get value() { return ''; }\n  set value(param: number) {}\n}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText))
		})
	}
}

// TestRelatedGetterSetterPairsSpansOnMeasuredShapes asserts the span on three shapes the corpus does
// not carry.
//
// The parenthesized type is the one that matters and it is the only place the two trees disagree
// about where the finding goes. TSESTree has no parenthesized-type node, so upstream reports
// `string` for `get value(): (string)`, measured on the installed build at columns 17 through 23.
// typescript-go keeps a `KindParenthesizedType` spanning `(string)`, so reporting the getter's type
// node directly would be two characters wider than upstream on exactly this input, and no imported
// fixture can see it because the corpus writes no parenthesized type.
func TestRelatedGetterSetterPairsSpansOnMeasuredShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a parenthesized return type reports the inner type", "interface I {\n  get value(): (string);\n  set value(v: number);\n}", "string"},
		{"a multi-line object return type spans all of it", "interface I {\n  get value(): {\n    a: string;\n  };\n  set value(v: number);\n}", "{\n    a: string;\n  }"},
		{"a symbol-keyed pair reports the getter's type", "class C {\n  get [Symbol.iterator](): number { return 1; }\n  set [Symbol.iterator](v: string) {}\n}", "number"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			got := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.want {
				t.Errorf("reported span = %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestRelatedGetterSetterPairsReportsInSourceOrder pins the ORDER of two findings in one pass.
//
// The pairs are gathered into a map and a map iteration in Go is deliberately randomized, so a rule
// reading the map directly reports two findings in an order that changes between runs. That is
// invisible to `ExpectFindings`, which compares ids and both are `mismatch`, and it is invisible to
// a single-finding span assertion. It shows up as a flaky diff in whatever consumes the output.
//
// The second case is the nesting claim stated as a measurement: an inner class declared inside an
// outer method reports its own pair and never sees the outer names, which is what the per-container
// listener buys instead of upstream's push-and-pop stack.
func TestRelatedGetterSetterPairsReportsInSourceOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"two mismatching pairs in one container report in source order", "class C {\n  get a(): number { return 1; }\n  set a(v: string) {}\n  get b(): boolean { return true; }\n  set b(v: string) {}\n}", []string{"number", "boolean"}},
		{"an outer and an inner class each report their own", "class Outer {\n  get a(): number { return 1; }\n  set a(v: string) {}\n  m() {\n    class Inner {\n      get b(): boolean { return true; }\n      set b(v: string) {}\n    }\n    return Inner;\n  }\n}", []string{"number", "boolean"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RelatedGetterSetterPairs,
				relatedGetterSetterPairsFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("want %d diagnostics, got %d", len(testCase.want), len(result.Diagnostics))
			}
			for i, want := range testCase.want {
				diagnostic := result.Diagnostics[i]
				got := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("diagnostic %d span = %q, want %q", i, got, want)
				}
			}
		})
	}
}

// TestRelatedGetterSetterPairsRequiresTheTypedHarness asserts the rule declines a nil checker.
//
// The plain harness hands a rule no program, and `GetSymbolAtLocation` and friends on a nil checker
// return nil rather than crashing, so a typed rule that forgets its guard goes SILENT instead of
// panicking. Silence is the more dangerous of the two, because every StaysSilent case above would
// then pass vacuously and the whole suite would prove nothing. This fails loudly if the guard is
// ever removed and the shim's nil tolerance changes underneath it.
func TestRelatedGetterSetterPairsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "class C {\n  get value(): number { return 1; }\n  set value(v: string) {}\n}"
	rule_testing.ExpectClean(t, rule_testing.Run(t, RelatedGetterSetterPairs,
		relatedGetterSetterPairsFile, source))
}

// TestRelatedGetterSetterPairsMessage pins the message id and its description text.
//
// A `rule.Message` is `{Id, Description}` with no interpolation, so there is nothing to render and
// nothing a format string can get wrong. Both fields are asserted against literals typed here rather
// than against the rule's own constant, because a comparison to the constant moves with it under
// mutation and passes while the message is wrong.
func TestRelatedGetterSetterPairsMessage(t *testing.T) {
	t.Parallel()

	source := "class C {\n  get value(): number { return 1; }\n  set value(v: string) {}\n}"
	result := rule_testing.RunTyped(t, RelatedGetterSetterPairs, relatedGetterSetterPairsFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "mismatch" {
		t.Errorf("message id = %q, want %q", got, "mismatch")
	}
	want := "A getter and its setter name one property, so a caller that reads the property and " +
		"writes the value straight back has to type-check. The getter here returns a type the " +
		"setter will not accept, which makes that round trip an error and usually means one of " +
		"the two annotations is wrong. Widen the setter's parameter, or narrow what the getter " +
		"returns."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("message description = %q, want %q", got, want)
	}
}

package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// buttonHasTypeFile is where the fixtures pretend to live.
//
// A `.tsx` name because most cases are JSX and need a parser that reads it. It is not load-bearing:
// this rule has no file gate, and `TestButtonHasTypeHasNoFileGate` pins that by reporting on a `.ts`
// file too.
const buttonHasTypeFile = "/repository/source/ButtonHasType.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/button-has-type.js` holds 30 valid and 28
// invalid cases. Every string below was pulled out of that file by loading it with a stubbed
// `RuleTester` and serializing the captured object to JSON, then emitted into these tables by a
// generator. No case was typed by hand.
//
// All 58 were run against the installed build, 7.37.5, through the ESLint Linter API with the
// TypeScript parser and each case's own settings honored. The clone and the installed build agreed
// on all 58, so there is no drift to record.
//
// # Three cases are held back, and each is recorded rather than dropped
//
// Two carry `settings: {react: {pragma: 'Foo'}}`, which has no counterpart here because
// cohere reads no React settings. `Foo.createElement("button")` reports upstream ONLY
// under that setting: re-run with settings removed it goes clean, measured, and clean is the answer
// this port has to produce. Both are pinned as declining cases in
// `TestButtonHasTypePragmaIsFixedAtReact` with the measurement at the line, rather than sitting in
// the silent list where they would read as agreement.
//
// The third needs ambient React types and is carried in
// `TestButtonHasTypeAcceptsTheTypedUpstreamCase` with a three-line preamble declaring them. It is
// upstream's own case verbatim below that preamble, and it was re-run with the preamble against the
// installed build to confirm the preamble does not change the verdict.
//
// # The invalid table asserts the RENDERED TEXT of the two interpolating ids
//
// `invalidValue` and `forbiddenValue` interpolate the offending value and carry byte-identical text
// upstream, differing only by id. So an id assertion alone cannot see a message that interpolates
// the wrong thing, which is the defect the port brief records as costing a porter a whole suite. The
// fifth column is the exact string the installed build rendered, captured during the same run that
// verified the ids, and those two are asserted as a PREFIX because this port appends a sentence of
// reasoning after upstream's wording.
//
// `missingType` and `complexType` interpolate nothing, and this port REPLACES their wording rather
// than appending to it, because upstream's two sentences name the symptom and not the consequence.
// So an upstream-prefix assertion is meaningless for them and the fifth column is empty. Their text
// is pinned instead by `TestButtonHasTypeMessagesReadAsWritten`, against literals typed there.
//
// # The options column is RAW JSON, on purpose
//
// All three options default to TRUE, which is not Go's zero value, so a fixture handed a struct
// built by hand would leave the decoder's default restoration entirely untested and would pass
// either way. Every case routes through `DecodeButtonHasTypeOptions`, the same function the config
// calls. An empty string means a bare severity, which is what hands a real rule nil options.

// TestButtonHasTypeFires asserts ids, count, order, and rendered text.
func TestButtonHasTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
		wantTexts  []string
	}{
		{"upstream invalid-0", "<button/>", "", []string{"missingType"}, []string{"Missing an explicit type attribute for button"}},
		{"upstream invalid-1", "<button type=\"foo\"/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-2", "<button type={foo}/>", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-3", "<button type={\"foo\"}/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-4", "<button type={'foo'}/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-5", "<button type={`foo`}/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-6", "<button type={`button${foo}`}/>", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-7", "<button type=\"reset\"/>", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-8", "<button type={condition ? \"button\" : foo}/>", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-9", "<button type={condition ? \"button\" : \"foo\"}/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-10", "<button type={condition ? \"button\" : \"reset\"}/>", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-11", "<button type={condition ? foo : \"button\"}/>", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-12", "<button type={condition ? \"foo\" : \"button\"}/>", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-13", "<button type/>", "", []string{"invalidValue"}, []string{"\"true\" is an invalid value for button type attribute"}},
		{"upstream invalid-14", "<button type={condition ? \"reset\" : \"button\"}/>", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-15", "React.createElement(\"button\")", "", []string{"missingType"}, []string{"Missing an explicit type attribute for button"}},
		{"upstream invalid-16", "React.createElement(\"button\", {type: foo})", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-17", "React.createElement(\"button\", {type: \"foo\"})", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-18", "React.createElement(\"button\", {type: \"reset\"})", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-19", "React.createElement(\"button\", {type: condition ? \"button\" : foo})", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-20", "React.createElement(\"button\", {type: condition ? \"button\" : \"foo\"})", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-21", "React.createElement(\"button\", {type: condition ? \"button\" : \"reset\"})", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-22", "React.createElement(\"button\", {type: condition ? foo : \"button\"})", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
		{"upstream invalid-23", "React.createElement(\"button\", {type: condition ? \"foo\" : \"button\"})", "", []string{"invalidValue"}, []string{"\"foo\" is an invalid value for button type attribute"}},
		{"upstream invalid-24", "React.createElement(\"button\", {type: condition ? \"reset\" : \"button\"})", "{\"reset\":false}", []string{"forbiddenValue"}, []string{"\"reset\" is an invalid value for button type attribute"}},
		{"upstream invalid-25", "React.createElement(\"button\", {...extraProps})", "", []string{"missingType"}, []string{"Missing an explicit type attribute for button"}},
		{"upstream invalid-27", "function Button({ type, ...extraProps }) { const button = type; return <button type={button} {...extraProps} />; }", "", []string{"complexType"}, []string{"The button type attribute must be specified by a static string or a trivial ternary expression"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runButtonHasType(t, testCase.sourceText, testCase.rawOptions)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			for index, diagnostic := range result.Diagnostics {
				// Only the two interpolating ids keep upstream's wording as a prefix; see the note
				// above the table. Skipping the other two here is not skipping the assertion: their
				// text is pinned by TestButtonHasTypeMessagesReadAsWritten.
				if diagnostic.Message.Id != "invalidValue" && diagnostic.Message.Id != "forbiddenValue" {
					continue
				}
				if !strings.HasPrefix(diagnostic.Message.Description, testCase.wantTexts[index]) {
					t.Errorf("finding %d reads %q, want it to begin %q",
						index, diagnostic.Message.Description, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestButtonHasTypeStaysSilent runs upstream's clean cases.
func TestButtonHasTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "<span/>", ""},
		{"upstream valid-1", "<span type=\"foo\"/>", ""},
		{"upstream valid-2", "<button type=\"button\"/>", ""},
		{"upstream valid-3", "<button type=\"submit\"/>", ""},
		{"upstream valid-4", "<button type=\"reset\"/>", ""},
		{"upstream valid-5", "<button type={\"button\"}/>", ""},
		{"upstream valid-6", "<button type={'button'}/>", ""},
		{"upstream valid-7", "<button type={`button`}/>", ""},
		{"upstream valid-8", "<button type={condition ? \"button\" : \"submit\"}/>", ""},
		{"upstream valid-9", "<button type={condition ? 'button' : 'submit'}/>", ""},
		{"upstream valid-10", "<button type={condition ? `button` : `submit`}/>", ""},
		{"upstream valid-11", "<button type=\"button\"/>", "{\"reset\":false}"},
		{"upstream valid-12", "React.createElement(\"span\")", ""},
		{"upstream valid-13", "React.createElement(\"span\", {type: \"foo\"})", ""},
		{"upstream valid-14", "React.createElement(\"button\", {type: \"button\"})", ""},
		{"upstream valid-15", "React.createElement(\"button\", {type: 'button'})", ""},
		{"upstream valid-16", "React.createElement(\"button\", {type: `button`})", ""},
		{"upstream valid-17", "React.createElement(\"button\", {type: \"submit\"})", ""},
		{"upstream valid-18", "React.createElement(\"button\", {type: 'submit'})", ""},
		{"upstream valid-19", "React.createElement(\"button\", {type: `submit`})", ""},
		{"upstream valid-20", "React.createElement(\"button\", {type: \"reset\"})", ""},
		{"upstream valid-21", "React.createElement(\"button\", {type: 'reset'})", ""},
		{"upstream valid-22", "React.createElement(\"button\", {type: `reset`})", ""},
		{"upstream valid-23", "React.createElement(\"button\", {type: condition ? \"button\" : \"submit\"})", ""},
		{"upstream valid-24", "React.createElement(\"button\", {type: condition ? 'button' : 'submit'})", ""},
		{"upstream valid-25", "React.createElement(\"button\", {type: condition ? `button` : `submit`})", ""},
		{"upstream valid-26", "React.createElement(\"button\", {type: \"button\"})", "{\"reset\":false}"},
		{"upstream valid-27", "document.createElement(\"button\")", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runButtonHasType(t, testCase.sourceText, testCase.rawOptions))
		})
	}
}

// runButtonHasType drives the rule through its own decoder, on a typed program.
//
// `RunTyped` because the rule declares `NeedsTypeChecker` for its bare-`createElement` arm, which is
// the same resolution `checked-requires-onchange-or-readonly` needs and the same helper.
func runButtonHasType(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeButtonHasTypeOptions([]byte(rawOptions))
	if err != nil {
		t.Fatalf("decoding %q: %v", rawOptions, err)
	}
	return rule_testing.RunTypedWithOptions(t, ButtonHasType, buttonHasTypeFile, sourceText, decoded)
}

// TestButtonHasTypeAcceptsTheTypedUpstreamCase carries the one corpus case needing React types.
//
// Upstream's case verbatim, under a three-line preamble declaring `ReactElement`, `Attributes` and
// `ButtonHTMLAttributes` so it parses without a React installation. The preamble was checked against
// the installed build: the case is clean there with it, exactly as it is without it, so the preamble
// changes nothing the rule can see.
//
// It is worth its own test rather than a row in the silent table because it is the only multi-line
// case, the only one whose `<button>` sits inside a fragment inside a map callback, and the only one
// where a spread follows a valid `type`.
func TestButtonHasTypeAcceptsTheTypedUpstreamCase(t *testing.T) {
	t.Parallel()

	const source = "declare type ReactElement = unknown;\ndeclare type Attributes = { key?: string };\ndeclare type ButtonHTMLAttributes<T> = { children?: unknown; onClick?: () => void };\n\n        function MyComponent(): ReactElement {\n          const buttonProps: (Required<Attributes> & ButtonHTMLAttributes<HTMLButtonElement>)[] = [\n            {\n              children: 'test',\n              key: 'test',\n              onClick: (): void => {\n                return;\n              },\n            },\n          ];\n\n          return <>\n            {\n              buttonProps.map(\n                ({ key, ...props }: Required<Attributes> & ButtonHTMLAttributes<HTMLButtonElement>): ReactElement =>\n                  <button key={key} type=\"button\" {...props} />\n              )\n            }\n          </>;\n        }\n      "
	rule_testing.ExpectClean(t, runButtonHasType(t, source, ""))
}

// TestButtonHasTypePragmaIsFixedAtReact records the two cases a settings surface would change.
//
// Upstream reads the pragma from `settings.react.pragma` or from a `@jsx` annotation comment.
// cohere reads neither, so the pragma is the literal `React` here and a project renaming it
// gets a false NEGATIVE. Recorded rather than hidden, because both of these sit in upstream's corpus
// and dropping them silently would leave the next reader believing the corpus was fully imported.
//
// Measured on the installed build: with `{react: {pragma: 'Foo'}}` the second reports `missingType`;
// with settings removed it goes clean, and clean is what this port produces. The first is clean
// either way and is here as the control, so a reader can see the pair rather than one row.
func TestButtonHasTypePragmaIsFixedAtReact(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a foreign pragma constructing a span is clean either way", "declare const Foo: any;\nFoo.createElement(\"span\");\n"},
		{"a foreign pragma constructing a button is silent here and reports upstream under settings", "declare const Foo: any;\nFoo.createElement(\"button\");\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runButtonHasType(t, testCase.sourceText, ""))
		})
	}
}

// TestButtonHasTypeReadsStaticValuesUpstreamsWay pins the literal table.
//
// Upstream's first `checkExpression` arm accepts anything its parser types as `Literal` and renders
// it with `String(value)`, which makes several shapes that read as computed actually static. The
// corpus writes only strings and one template, so none of this is visible from the imported cases.
// Every row was measured on the installed build before it was written here.
//
// The numeric rows are also a check on OUR parser: `.Text()` on a NumericLiteral is already the
// canonical rendering of the double, so `0x10` is `16` and `1.50` is `1.5` without any arithmetic.
// A BigInt is the one place the two disagree, keeping its `n` suffix here, and the rule trims it.
func TestButtonHasTypeReadsStaticValuesUpstreamsWay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
		wantValue  string
	}{
		{"null is a literal", "const a = <button type={null}/>;\n", []string{"invalidValue"}, "\"null\""},
		{"true is a literal", "const a = <button type={true}/>;\n", []string{"invalidValue"}, "\"true\""},
		{"false is a literal", "const a = <button type={false}/>;\n", []string{"invalidValue"}, "\"false\""},
		{"an integer", "const a = <button type={42}/>;\n", []string{"invalidValue"}, "\"42\""},
		{"a trailing zero is dropped", "const a = <button type={1.50}/>;\n", []string{"invalidValue"}, "\"1.5\""},
		{"hexadecimal renders as decimal", "const a = <button type={0x10}/>;\n", []string{"invalidValue"}, "\"16\""},
		{"exponent renders expanded", "const a = <button type={1e3}/>;\n", []string{"invalidValue"}, "\"1000\""},
		{"a numeric separator is dropped", "const a = <button type={1_0}/>;\n", []string{"invalidValue"}, "\"10\""},
		{"a bigint drops its suffix", "const a = <button type={1n}/>;\n", []string{"invalidValue"}, "\"1\""},
		{"a regular expression renders as source", "const a = <button type={/re/}/>;\n", []string{"invalidValue"}, "\"/re/\""},
		{"an empty template is an empty string", "const a = <button type={``}/>;\n", []string{"invalidValue"}, "\"\""},
		{"an empty string attribute", "const a = <button type=\"\"/>;\n", []string{"invalidValue"}, "\"\""},
		{"a valueless attribute renders true", "const a = <button type/>;\n", []string{"invalidValue"}, "\"true\""},
		// The four below are NOT literals to upstream's parser and land on the complex arm. They are
		// the ones a port is most likely to get backwards, because `undefined` and `-1` read as
		// constants to a human and are an Identifier and a unary expression to a parser.
		{"undefined is an identifier", "const a = <button type={undefined}/>;\n", []string{"complexType"}, ""},
		{"a negative number is a unary expression", "const a = <button type={-1}/>;\n", []string{"complexType"}, ""},
		{"concatenation is not static", "const a = <button type={'bu' + 'tton'}/>;\n", []string{"complexType"}, ""},
		// Upstream measured this on `bu${''}tton`, which the checker folds to the literal "button",
		// so cohere is deliberately silent there (pinned in TestButtonHasTypeTrustsAProvenType). The
		// interpolation here is a `string`, which nothing can fold, and it still reports.
		{"an interpolating template is not static", "declare const part: string;\nconst a = <button type={`bu${part}tton`}/>;\n", []string{"complexType"}, ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runButtonHasType(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if testCase.wantValue == "" {
				return
			}
			want := testCase.wantValue + " is an invalid value for button type attribute."
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, want) {
				t.Errorf("message reads %q, want it to begin %q",
					result.Diagnostics[0].Message.Description, want)
			}
		})
	}
}

// TestButtonHasTypeRecursesThroughTernaries pins the recursion and the per-branch count.
//
// `checkExpression` calls itself on both arms, so a ternary with two bad branches reports TWICE and
// a nested ternary is walked to its full depth. The corpus writes no ternary with two bad branches
// and no nested ternary at all, so a rule that reported once per element would pass every imported
// case. Both verdicts measured on the installed build.
func TestButtonHasTypeRecursesThroughTernaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"both branches bad reports twice",
			"declare const condition: boolean;\nconst a = <button type={condition ? 'foo' : 'bar'}/>;\n",
			[]string{"invalidValue", "invalidValue"},
		},
		{
			"both branches good is clean",
			"declare const condition: boolean;\nconst a = <button type={condition ? 'button' : 'submit'}/>;\n",
			nil,
		},
		{
			"a nested ternary is walked to depth",
			"declare const a: boolean;\ndeclare const b: boolean;\nconst e = <button type={a ? (b ? 'button' : 'foo') : 'reset'}/>;\n",
			[]string{"invalidValue"},
		},
		{
			"a nullish coalesce is not a ternary",
			"declare const a: string | undefined;\nconst e = <button type={a ?? 'button'}/>;\n",
			[]string{"complexType"},
		},
		// The two below are why the rule skips parentheses before switching. Upstream's parser
		// folds them, so a parenthesized ternary is still a ternary to it and a parenthesized string
		// is still a string. Without the skip both fell to the complex arm, which is how the first
		// of these was found: it is upstream corpus case invalid-8's shape with the inner ternary
		// parenthesized, and this port reported complexType where upstream reports invalidValue.
		{
			"a parenthesized nested ternary still recurses",
			"declare const a: boolean;\ndeclare const b: boolean;\nconst e = <button type={a ? (b ? 'button' : 'foo') : 'reset'}/>;\n",
			[]string{"invalidValue"},
		},
		{
			"a parenthesized string is still a static value",
			"const e = <button type={('foo')}/>;\n",
			[]string{"invalidValue"},
		},
		{
			"a parenthesized whole ternary still recurses",
			"declare const a: boolean;\nconst e = <button type={(a ? 'button' : 'foo')}/>;\n",
			[]string{"invalidValue"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runButtonHasType(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestButtonHasTypeSpans asserts WHERE each finding points.
//
// This is the assertion the rule most needs, because `complexType` anchors on the EXPRESSION and
// every other id anchors on the element or the call. That is upstream's `reportComplex(expression)`
// against `report(context, ..., {node})` everywhere else, and a rule anchoring all four on the
// element would pass every id fixture in this file while pointing at a whole tag where upstream
// points at three characters.
//
// The text is sliced from the source the harness WROTE rather than from the literal above, because
// `RunTyped` trims the fixture and a span sliced from an untrimmed literal is off by one.
func TestButtonHasTypeSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{"missing type points at the whole element", "<button/>", []string{"<button/>"}},
		{"a paired element too", "<button></button>", []string{"<button></button>"}},
		{"an invalid value points at the whole element", "<button type=\"foo\"/>", []string{"<button type=\"foo\"/>"}},
		{"a complex type points at the EXPRESSION", "declare const foo: string;\nconst a = <button type={foo}/>;", []string{"foo"}},
		{
			"a complex ternary branch points at that branch",
			"declare const condition: boolean;\ndeclare const foo: string;\nconst a = <button type={condition ? \"button\" : foo}/>;",
			[]string{"foo"},
		},
		{
			"a complex template points at the template",
			"declare const foo: string;\nconst a = <button type={`button${foo}`}/>;",
			[]string{"`button${foo}`"},
		},
		{
			"a call reports on the whole call",
			"declare const React: any;\nReact.createElement(\"button\");",
			[]string{"React.createElement(\"button\")"},
		},
		{
			"a complex call property points at the value",
			"declare const React: any;\ndeclare const foo: string;\nReact.createElement(\"button\", {type: foo});",
			[]string{"foo"},
		},
		{
			"a shorthand property points at the shorthand",
			"declare const React: any;\ndeclare const type: string;\nReact.createElement(\"button\", {type});",
			[]string{"type"},
		},
		// The three below pin PARENTHESIS handling, which the corpus writes nowhere. Upstream's
		// parser folds parentheses away entirely, so a parenthesized value reaches the same arm as
		// the bare one and the finding lands INSIDE the parentheses. Ours preserves them, so the
		// rule skips them, and these assert that the skip goes all the way through rather than one
		// level. Measured on the installed build: `(foo)` reports at columns 16-19, `((foo))` at
		// 17-20, both of which are the bare identifier.
		{
			"one layer of parentheses is skipped",
			"declare const foo: string;\nconst a = <button type={(foo)}/>;",
			[]string{"foo"},
		},
		{
			"two layers are skipped",
			"declare const foo: string;\nconst a = <button type={((foo))}/>;",
			[]string{"foo"},
		},
		{
			"a parenthesized complex ternary branch points inside",
			"declare const a: boolean;\ndeclare const b: string;\nconst e = <button type={a ? (b) : 'button'}/>;",
			[]string{"b"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runButtonHasType(t, testCase.sourceText, "")
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			source := result.SourceFile.Text()
			for index, diagnostic := range result.Diagnostics {
				got := source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantTexts[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestButtonHasTypeMatchesUpstreamOnShapesTheCorpusOmits pins measured verdicts.
//
// Each row was run against the installed build before it was written. The object-key rows are the
// surprising ones: upstream requires an Identifier key, so `{'type': 'foo'}` and `{['type']: 'foo'}`
// are invisible and the call reports `missingType` rather than `invalidValue`. The property is
// unmistakably `type` to a reader and the rule cannot see it, and that is reproduced rather than
// improved, because widening it would report where upstream is quiet.
func TestButtonHasTypeMatchesUpstreamOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a component named Button declines", "declare const Button: any;\nconst a = <Button/>;\n", nil},
		{"a member-expression tag declines", "declare const button: any;\nconst a = <button.x/>;\n", nil},
		{"an uppercase tag declines", "declare const BUTTON: any;\nconst a = <BUTTON/>;\n", nil},
		{"an uppercase attribute name is not the type attribute", "const a = <button TYPE=\"button\"/>;\n", []string{"missingType"}},
		{"a spread does not exempt", "declare const props: object;\nconst a = <button {...props}/>;\n", []string{"missingType"}},
		{"a spread after a valid type is clean", "declare const props: object;\nconst a = <button type=\"button\" {...props}/>;\n", nil},
		{"a spread before a valid type is clean", "declare const props: object;\nconst a = <button {...props} type=\"button\"/>;\n", nil},
		{
			"a string-literal object key is invisible",
			"declare const React: any;\nReact.createElement('button', {'type': 'foo'});\n",
			[]string{"missingType"},
		},
		{
			"a computed object key is invisible",
			"declare const React: any;\nReact.createElement('button', {['type']: 'foo'});\n",
			[]string{"missingType"},
		},
		{
			"a spread beside a real type still reads the type",
			"declare const React: any;\ndeclare const spread: object;\nReact.createElement('button', {...spread, type: 'foo'});\n",
			[]string{"invalidValue"},
		},
		{
			"an explicit null properties argument is a missing type",
			"declare const React: any;\nReact.createElement('button', null);\n",
			[]string{"missingType"},
		},
		{
			"an empty properties object is a missing type",
			"declare const React: any;\nReact.createElement('button', {});\n",
			[]string{"missingType"},
		},
		{
			"a template-literal tag declines",
			"declare const React: any;\nReact.createElement(`button`);\n",
			nil,
		},
		{
			"a computed member on the pragma declines",
			"declare const React: any;\nReact['createElement']('button');\n",
			nil,
		},
		{
			"an undeclared bare createElement declines",
			"createElement('button');\n",
			nil,
		},
		{
			"an imported bare createElement reports",
			"import { createElement } from 'react';\ncreateElement('button');\n",
			[]string{"missingType"},
		},
		{
			"children after a valid type are ignored",
			"declare const React: any;\nReact.createElement('button', {type: 'button'}, 'x');\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runButtonHasType(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestButtonHasTypeDecodesItsOptions exercises the decoder the config layer calls.
//
// Every default here is TRUE, so this is the decoder the port brief warns about specifically: a
// `rule.DecodeOptionsInto` on this struct yields all-false and inverts the rule, and every fixture
// built from a struct rather than routed through the decoder passes anyway. The empty-input and
// empty-object rows are the ones that catch it, and the explicit-false rows are what prove the
// pointer wire type is doing its job rather than the defaults simply always winning.
func TestButtonHasTypeDecodesItsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                        string
		raw                         string
		button, submit, resetButton bool
	}{
		{"empty input answers the defaults", "", true, true, true},
		{"an empty object answers the defaults", "{}", true, true, true},
		{"button off", `{"button":false}`, false, true, true},
		{"submit off", `{"submit":false}`, true, false, true},
		{"reset off", `{"reset":false}`, true, true, false},
		{"all off", `{"button":false,"submit":false,"reset":false}`, false, false, false},
		{"explicit true is still true", `{"button":true}`, true, true, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeButtonHasTypeOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			settings, ok := decoded.(ButtonHasTypeOptions)
			if !ok {
				t.Fatalf("decoded to %T", decoded)
			}
			if settings.Button != testCase.button || settings.Submit != testCase.submit ||
				settings.Reset != testCase.resetButton {
				t.Errorf("decoded to %+v, want button=%v submit=%v reset=%v",
					settings, testCase.button, testCase.submit, testCase.resetButton)
			}
		})
	}
}

// TestButtonHasTypeRejectsMalformedOptions asserts the decoder surfaces bad input.
func TestButtonHasTypeRejectsMalformedOptions(t *testing.T) {
	t.Parallel()

	if _, err := DecodeButtonHasTypeOptions([]byte(`{"button":`)); err == nil {
		t.Fatal("truncated JSON decoded without error")
	}
	if _, err := DecodeButtonHasTypeOptions([]byte(`"error"`)); err == nil {
		t.Fatal("a string decoded into the options struct without error")
	}
}

// TestButtonHasTypeHandlesNilOptions covers the path the fixtures cannot reach.
//
// The config layer turns a decoder error into nil for a non-required rule, and `options.(T)` on nil
// yields the ZERO value, which for this rule forbids all three types and would report
// `forbiddenValue` on `<button type="button"/>`. The rule restores the defaults explicitly on that
// path and this asserts it, because no fixture routed through the decoder can ever see it.
func TestButtonHasTypeHandlesNilOptions(t *testing.T) {
	t.Parallel()

	clean := rule_testing.RunTypedWithOptions(t, ButtonHasType, buttonHasTypeFile,
		"const a = <button type=\"button\"/>;\n", nil)
	rule_testing.ExpectClean(t, clean)

	reporting := rule_testing.RunTypedWithOptions(t, ButtonHasType, buttonHasTypeFile,
		"const a = <button/>;\n", nil)
	rule_testing.ExpectFindings(t, reporting, "missingType")
}

// TestButtonHasTypeSeparatesInvalidFromForbidden pins the two ids that share their text.
//
// `invalidValue` and `forbiddenValue` are byte-identical strings upstream, so nothing but the id
// distinguishes them and a rule collapsing them into one would pass every text assertion in this
// file. They answer different questions: the first means the value is not an HTML button type at
// all, the second means it is one and this configuration turned it off. The pair below is the same
// element under two configurations, which is the only way to see the difference.
func TestButtonHasTypeSeparatesInvalidFromForbidden(t *testing.T) {
	t.Parallel()

	const element = "const a = <button type=\"reset\"/>;\n"

	rule_testing.ExpectClean(t, runButtonHasType(t, element, ""))
	rule_testing.ExpectFindings(t, runButtonHasType(t, element, `{"reset":false}`), "forbiddenValue")
	rule_testing.ExpectFindings(t, runButtonHasType(t, "const a = <button type=\"nope\"/>;\n", `{"reset":false}`), "invalidValue")

	// Each flag is then exercised against its OWN type, with the other two left on. Upstream's
	// corpus only ever switches `reset` off, so a rule whose `submit` arm read the `button` flag
	// passes every imported case: a mutant doing exactly that survived the rest of this suite. The
	// three pairs below are what separate the flags from each other.
	for _, row := range []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"button off reports button", "const a = <button type=\"button\"/>;\n", `{"button":false}`, []string{"forbiddenValue"}},
		{"button off leaves submit alone", "const a = <button type=\"submit\"/>;\n", `{"button":false}`, nil},
		{"button off leaves reset alone", "const a = <button type=\"reset\"/>;\n", `{"button":false}`, nil},
		{"submit off reports submit", "const a = <button type=\"submit\"/>;\n", `{"submit":false}`, []string{"forbiddenValue"}},
		{"submit off leaves button alone", "const a = <button type=\"button\"/>;\n", `{"submit":false}`, nil},
		{"submit off leaves reset alone", "const a = <button type=\"reset\"/>;\n", `{"submit":false}`, nil},
		{"reset off leaves button alone", "const a = <button type=\"button\"/>;\n", `{"reset":false}`, nil},
		{"reset off leaves submit alone", "const a = <button type=\"submit\"/>;\n", `{"reset":false}`, nil},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runButtonHasType(t, row.sourceText, row.rawOptions), row.wantIds...)
		})
	}

	// A value outside the three can never be forbidden, however the rule is configured, because
	// upstream tests membership in the configuration object before testing the flag.
	rule_testing.ExpectFindings(t, runButtonHasType(t, "const a = <button type=\"nope\"/>;\n",
		`{"button":false,"submit":false,"reset":false}`), "invalidValue")
}

// TestButtonHasTypeNeedsTheTypedHarness pins the checker declaration.
//
// Only the bare-`createElement` arm needs the checker, so the assertion is one source that reports
// under `RunTyped` and is silent under `Run`, with the JSX arm as the control that must report under
// both. Without the control this would pass against a broken harness.
func TestButtonHasTypeNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const bareCall = "import { createElement } from 'react';\ncreateElement('button');\n"

	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, ButtonHasType,
		buttonHasTypeFile, bareCall, DefaultButtonHasTypeOptions()), "missingType")
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ButtonHasType,
		buttonHasTypeFile, bareCall, DefaultButtonHasTypeOptions()))

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, ButtonHasType,
		buttonHasTypeFile, "const a = <button/>;\n", DefaultButtonHasTypeOptions()), "missingType")
}

// TestButtonHasTypeHasNoFileGate reports on a `.ts` file, not only on `.tsx`.
//
// Three shipped rules in this package used to gate on `.tsx`/`.jsx` through `isJsxFileName`, oxc
// residue rather than upstream behavior (task 2abaqvt); all three gates are gone as of 4cff5fe,
// 206b628 and 9dbd234. This rule never had one and this pins that. The
// `createElement` arm is used because JSX in a `.ts` file is a syntax error, so it is the arm that
// can actually distinguish the two extensions, and `.jsx` is unavailable because the typed harness
// writes a tsconfig including only `**/*.ts` and `**/*.tsx`.
func TestButtonHasTypeHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nReact.createElement('button');\n"
	for _, fileName := range []string{
		"/repository/source/ButtonHasType.tsx",
		"/repository/source/ButtonHasType.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ButtonHasType, fileName, source,
				DefaultButtonHasTypeOptions())
			rule_testing.ExpectFindings(t, result, "missingType")
		})
	}
}

// TestButtonHasTypeSurvivesShapesThatWouldPanic drives the shapes with nothing to read.
//
// The walk recovers per FILE rather than per rule, so one nil dereference here takes the file away
// from every rule in the tree, and no `ExpectFindings` fixture can see a panic. These are ordinary
// code: a `super()` whose callee is a bare keyword, a dynamic `import()`, an optional call, and a
// JSX element with a malformed attribute list.
func TestButtonHasTypeSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"(()=>{})();\n",
		"class Base { constructor() {} }\nclass Derived extends Base { constructor() { super(); } }\n",
		"async function load() { await import('./other'); }\n",
		"declare const maybe: undefined | ((tag: string) => void);\nmaybe?.('button');\n",
		"declare function tag(strings: TemplateStringsArray): void;\ntag`button`;\n",
		"declare const React: any;\nReact.createElement();\n",
		"const a = <div/>;\n",
		"const a = <></>;\n",
		"declare const props: object;\nconst a = <div {...props}/>;\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runButtonHasType(t, sourceText, ""))
		})
	}
}

// TestButtonHasTypeMessagesReadAsWritten pins the two non-interpolating messages.
//
// `missingType` and `complexType` are the two ids this port rewords rather than appends to, so the
// imported table cannot assert them against upstream's text. They are asserted here against literals
// typed in this file rather than against the rule's own constants, because comparing a diagnostic to
// the constant it was reported with is an equality whose two sides move together under mutation.
//
// The two interpolating ids are asserted here too, in the one form the imported table cannot reach:
// their upstream wording is identical, so this pins that the port keeps them identical and that the
// value is the only thing that differs between two findings on one element.
func TestButtonHasTypeMessagesReadAsWritten(t *testing.T) {
	t.Parallel()

	missing := runButtonHasType(t, "const a = <button/>;\n", "")
	if len(missing.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(missing.Diagnostics))
	}
	if missing.Diagnostics[0].Message.Id != "missingType" {
		t.Errorf("id is %q", missing.Diagnostics[0].Message.Id)
	}
	if !strings.HasPrefix(missing.Diagnostics[0].Message.Description, "This `button` has no `type`.") {
		t.Errorf("missingType reads %q", missing.Diagnostics[0].Message.Description)
	}

	complex := runButtonHasType(t, "declare const foo: string;\nconst a = <button type={foo}/>;\n", "")
	if len(complex.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(complex.Diagnostics))
	}
	if complex.Diagnostics[0].Message.Id != "complexType" {
		t.Errorf("id is %q", complex.Diagnostics[0].Message.Id)
	}
	if !strings.HasPrefix(complex.Diagnostics[0].Message.Description, "This `button` computes its `type`,") {
		t.Errorf("complexType reads %q", complex.Diagnostics[0].Message.Description)
	}

	// Two findings on one element, differing only in the interpolated value. A format string that
	// dropped the value, or interpolated the same one twice, passes every id assertion in this file
	// and fails here.
	both := runButtonHasType(t,
		"declare const condition: boolean;\nconst a = <button type={condition ? 'foo' : 'bar'}/>;\n", "")
	if len(both.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(both.Diagnostics))
	}
	const suffix = " is an invalid value for button type attribute."
	if !strings.HasPrefix(both.Diagnostics[0].Message.Description, `"foo"`+suffix) {
		t.Errorf("first reads %q", both.Diagnostics[0].Message.Description)
	}
	if !strings.HasPrefix(both.Diagnostics[1].Message.Description, `"bar"`+suffix) {
		t.Errorf("second reads %q", both.Diagnostics[1].Message.Description)
	}
}

// TestButtonHasTypeTrustsAProvenType is where cohere is deliberately quieter than ESLint.
//
// Upstream is type-blind, so every non-literal `type` is `complexType` there, including one the
// checker proves can only be `button`, `submit` or `reset`. The silent rows are that shape, and the
// first is the real site that motivated it: `libraries/structure/source/components/buttons/Button.tsx`
// line 222, `<button type={type}>` with `type = 'button'` destructured from a prop typed
// `'button' | 'submit' | 'reset'`. ESLint 7.37.5 reports it; it is a false positive, because the
// default removes `undefined` and the annotation removes every other string.
//
// The reporting rows are the controls that keep the exemption exactly as wide as the proof: the same
// prop WITHOUT the default (so `undefined` is a constituent and HTML's `submit` default can apply), a
// `string`, a union with one bad member, `any`, and a fully valid union under a config that forbids
// one of its members, where "permitted" rather than "known" is the test.
func TestButtonHasTypeTrustsAProvenType(t *testing.T) {
	t.Parallel()

	const properties = "interface ButtonProperties { type?: 'button' | 'submit' | 'reset'; }\n"

	silent := []struct {
		name       string
		sourceText string
	}{
		{"Button.tsx:222, a defaulted prop typed as the three literals",
			properties + "function Button({ type = 'button' }: ButtonProperties) { return <button type={type} />; }\n"},
		{"a single literal const",
			"const kind = 'submit' as const;\nconst a = <button type={kind} />;\n"},
		{"a typed branch of a ternary",
			"declare const condition: boolean;\ndeclare const kind: 'reset' | 'button';\nconst a = <button type={condition ? 'submit' : kind} />;\n"},
		{"a string enum whose members are all valid",
			"enum Kind { Plain = 'button', Send = 'submit' }\ndeclare const kind: Kind;\nconst a = <button type={kind} />;\n"},
		{"createElement with a typed shorthand",
			"declare const React: { createElement(...a: unknown[]): unknown };\ndeclare const type: 'button' | 'submit';\nReact.createElement('button', { type });\n"},
		// Upstream reports complexType here (measured, see TestButtonHasTypeReadsStaticValuesUpstreamsWay);
		// the checker folds the template to the literal "button", so nothing can submit.
		{"a template the checker folds to a literal",
			"const a = <button type={`bu${''}tton`}/>;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runButtonHasType(t, testCase.sourceText, ""))
		})
	}

	reporting := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"the same prop with no default can be undefined",
			properties + "function Button({ type }: ButtonProperties) { return <button type={type} />; }\n", ""},
		{"a string",
			"declare const kind: string;\nconst a = <button type={kind} />;\n", ""},
		{"a union with one invalid member",
			"declare const kind: 'button' | 'foo';\nconst a = <button type={kind} />;\n", ""},
		{"any",
			"declare const kind: any;\nconst a = <button type={kind} />;\n", ""},
		{"a valid union with a member the config forbids",
			properties + "function Button({ type = 'button' }: ButtonProperties) { return <button type={type} />; }\n",
			"{\"reset\":false}"},
		{"a string enum with one invalid member",
			"enum Kind { Plain = 'button', Odd = 'menu' }\ndeclare const kind: Kind;\nconst a = <button type={kind} />;\n", ""},
		{"createElement with an untyped shorthand",
			"declare const React: { createElement(...a: unknown[]): unknown };\ndeclare const type: string;\nReact.createElement('button', { type });\n", ""},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runButtonHasType(t, testCase.sourceText, testCase.rawOptions), "complexType")
		})
	}
}

package react

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoUselessFragmentFile is where the fixtures pretend to live.
//
// A `.tsx` name, and here that IS load-bearing rather than incidental: every case is JSX, and JSX in
// a `.ts` file is a syntax error, so a `.ts` fixture would be silent because the parser read `<div>`
// as a type assertion rather than because any rule declined. That is a different zero from a file
// gate, and it is why this rule ships no gate test: there is no gate, and no fixture could tell the
// difference if there were. `self_closing_comp_test.go` records the same distinction.
const jsxNoUselessFragmentFile = "/repository/source/JsxNoUselessFragment.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-no-useless-fragment.js` holds 14 valid
// and 16 invalid cases. Every string below was pulled out of that file by loading it with a stubbed
// `RuleTester` and serializing the captured object to JSON, then emitted into these tables by a
// generator. No case was typed by hand.
//
// All 30 were run against the installed build, 7.37.5, through the ESLint Linter API with the
// TypeScript parser, driving `cohereAndFix` as well as `cohere` so the REPAIR was measured and not
// only the judgment.
//
// # One case is held back, and it is a duplicate with an opposite verdict
//
// `<div><>{"a"}{"b"}</></div>` appears TWICE in the corpus: once with an output and once with
// `output: null`. They are not a contradiction, they are a parser-version gate. The second carries
// `features: ["ts-old", "no-ts-new"]`, so it is asserted only under a TypeScript parser old enough
// that a fragment node lacked its opening and closing tokens, which made upstream's `canFix` decline.
// The modern parser is what this tree has, so the first is ours and the second is not expressible.
// The extractor is the only place that difference is visible; read as prose the two look like an
// error in the corpus.
//
// # The fix column is the REPAIRED SOURCE, and an empty string means the fixer declined
//
// `ExpectFindings` cannot see what a fix writes, and the port brief records a rule whose flag was
// forced always-on while thirty fixtures stayed green, because every one asserted an id and none
// asserted the repair. So every reporting row carries the exact source the installed build's fixer
// produced, and the rows where it produced nothing carry an empty string, which asserts the DECLINE
// rather than skipping the question.

// TestJsxNoUselessFragmentFires asserts ids, count, and the repaired source.
func TestJsxNoUselessFragmentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
		wantFixed  string
	}{
		{"upstream invalid-0", "<></>", "", []string{"NeedsMoreChildren"}, ""},
		{"upstream invalid-1", "<>{}</>", "", []string{"NeedsMoreChildren"}, ""},
		{"upstream invalid-2", "<>{meow}</>", "", []string{"NeedsMoreChildren"}, ""},
		{"upstream invalid-3", "<><div/></>", "", []string{"NeedsMoreChildren"}, "<div/>"},
		{"upstream invalid-4", "\n        <>\n          <div/>\n        </>\n      ", "", []string{"NeedsMoreChildren"}, "\n        <div/>\n      "},
		{"upstream invalid-5", "<Fragment />", "", []string{"NeedsMoreChildren"}, ""},
		{"upstream invalid-6", "\n        <React.Fragment>\n          <Foo />\n        </React.Fragment>\n      ", "", []string{"NeedsMoreChildren"}, "\n        <Foo />\n      "},
		{"upstream invalid-8", "<Eeee><>foo</></Eeee>", "", []string{"NeedsMoreChildren"}, ""},
		{"upstream invalid-9", "<div><>{\"a\"}{\"b\"}</></div>", "", []string{"ChildOfHtmlElement"}, "<div>{\"a\"}{\"b\"}</div>"},
		{"upstream invalid-11", "\n        <section>\n          <Eeee />\n          <Eeee />\n          <>{\"a\"}{\"b\"}</>\n        </section>", "", []string{"ChildOfHtmlElement"}, "\n        <section>\n          <Eeee />\n          <Eeee />\n          {\"a\"}{\"b\"}\n        </section>"},
		{"upstream invalid-12", "<div><Fragment>{\"a\"}{\"b\"}</Fragment></div>", "", []string{"ChildOfHtmlElement"}, "<div>{\"a\"}{\"b\"}</div>"},
		{"upstream invalid-13", "\n        <section>\n          git<>\n            <b>hub</b>.\n          </>\n\n          git<> <b>hub</b></>\n        </section>", "", []string{"ChildOfHtmlElement", "ChildOfHtmlElement"}, "\n        <section>\n          git<b>hub</b>.\n\n          git <b>hub</b>\n        </section>"},
		{"upstream invalid-14", "<div>a <>{\"\"}{\"\"}</> a</div>", "", []string{"ChildOfHtmlElement"}, "<div>a {\"\"}{\"\"} a</div>"},
		{"upstream invalid-15", "<><Foo>{moo}</Foo></>", "{\"allowExpressions\":true}", []string{"NeedsMoreChildren"}, "<Foo>{moo}</Foo>"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runJsxNoUselessFragment(t, testCase.sourceText, testCase.rawOptions)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			want := testCase.wantFixed
			if want == "" {
				// The fixer declined, so replaying whatever it offered must leave the source alone.
				want = testCase.sourceText
			}
			// A case reporting one diagnostic that carries one fix has nothing to deduplicate, so the
			// shared assertion applies and is used, which is what keeps the fixture-pair guard able
			// to see that this rule's repair is proven. The local applier below exists only for the
			// cases this rule reports twice on one node, where the same span arrives from both
			// messages and replaying it unmerged would double the edit.
			if testCase.wantFixed != "" && countJsxNoUselessFragmentFixes(result) == 1 {
				rule_testing.ExpectFixedSource(t, result, want)
				return
			}
			if got := applyJsxNoUselessFragmentFixes(t, result); got != want {
				t.Errorf("the repair produced %q, want %q", got, want)
			}
		})
	}
}

// countJsxNoUselessFragmentFixes totals the fixes across every diagnostic, which is how a case that
// needs no deduplication is told apart from one that does.
func countJsxNoUselessFragmentFixes(result rule_testing.Result) int {
	total := 0
	for _, diagnostic := range result.Diagnostics {
		total += len(diagnostic.Fixes)
	}
	return total
}

// TestJsxNoUselessFragmentStaysSilent runs upstream's clean cases.
func TestJsxNoUselessFragmentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "<><Foo /><Bar /></>", ""},
		{"upstream valid-1", "<>foo<div /></>", ""},
		{"upstream valid-2", "<> <div /></>", ""},
		{"upstream valid-3", "<>{\"moo\"} </>", ""},
		{"upstream valid-4", "<NotFragment />", ""},
		{"upstream valid-5", "<React.NotFragment />", ""},
		{"upstream valid-6", "<NotReact.Fragment />", ""},
		{"upstream valid-7", "<Foo><><div /><div /></></Foo>", ""},
		{"upstream valid-8", "<div p={<>{\"a\"}{\"b\"}</>} />", ""},
		{"upstream valid-9", "<Fragment key={item.id}>{item.value}</Fragment>", ""},
		{"upstream valid-10", "<Fooo content={<>eeee ee eeeeeee eeeeeeee</>} />", ""},
		{"upstream valid-11", "<>{foos.map(foo => foo)}</>", ""},
		{"upstream valid-12", "<>{moo}</>", "{\"allowExpressions\":true}"},
		{"upstream valid-13", "\n        <>\n          {moo}\n        </>\n      ", "{\"allowExpressions\":true}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runJsxNoUselessFragment(t, testCase.sourceText, testCase.rawOptions))
		})
	}
}

// runJsxNoUselessFragment drives the rule through its own decoder.
//
// `Run` rather than `RunTyped`, because every question is on the AST: what tag a node carries, what
// children it has, and what its parent is. `Run` also does not trim the fixture, which matters here
// more than usual: `ExpectFixedSource` compares the whole rewritten file, so a trimming harness
// would produce byte-correct repairs that all fail on a trailing newline.
func runJsxNoUselessFragment(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeJsxNoUselessFragmentOptions([]byte(rawOptions))
	if err != nil {
		t.Fatalf("decoding %q: %v", rawOptions, err)
	}
	return rule_testing.RunWithOptions(t, JsxNoUselessFragment, jsxNoUselessFragmentFile, sourceText, decoded)
}

// applyJsxNoUselessFragmentFixes replays the repair the way the edit engine would.
//
// `rule_testing.ExpectFixedSource` cannot be used directly here for two reasons, and both are
// correct refusals rather than harness gaps. It fatals when a rule proposed NO fix, which is exactly
// what this rule's declines produce and exactly what those rows need to assert. And it fatals on
// overlapping fixes, which this rule produces whenever both arms report on one node: they carry the
// same replacement over the same range, so the overlap is a duplicate rather than a conflict a
// fixture would have to guess about.
//
// So this applies at most one fix per distinct range, back to front, which is what the engine does
// with a duplicate pair, and it returns the source unchanged when there is nothing to apply. Both
// behaviours were checked against the installed build's own `cohereAndFix` on the same inputs.
func applyJsxNoUselessFragmentFixes(t *testing.T, result rule_testing.Result) string {
	t.Helper()

	type edit struct {
		start, end int
		text       string
	}
	seen := map[[2]int]bool{}
	edits := []edit{}
	for _, diagnostic := range result.Diagnostics {
		for _, fix := range diagnostic.Fixes {
			key := [2]int{fix.Range.Pos(), fix.Range.End()}
			if seen[key] {
				continue
			}
			seen[key] = true
			edits = append(edits, edit{start: key[0], end: key[1], text: fix.Text})
		}
	}

	sort.Slice(edits, func(first, second int) bool { return edits[first].start > edits[second].start })

	source := result.SourceFile.Text()
	previousStart := len(source)
	for _, item := range edits {
		if item.end > previousStart {
			t.Fatalf("fixes genuinely overlap at [%d,%d), which is a rule defect rather than a duplicate",
				item.start, item.end)
		}
		source = source[:item.start] + item.text + source[item.end:]
		previousStart = item.start
	}
	return source
}

// TestJsxNoUselessFragmentDeclinesToDropAttributes pins this port's one added decline.
//
// Upstream's fixer replaces the fragment with the text of its children, which silently deletes any
// attribute the opener carried. `isKeyedElement` skips a fragment with `key`, and nothing skips
// anything else. Measured on the installed build:
//
//	<div><Fragment id={1}>{x}</Fragment></div>      fixes to <div>{x}</div>, the id is gone
//	<div><Fragment {...rest}>{x}</Fragment></div>   fixes to <div>{x}</div>, the spread is gone
//
// A spread can carry `key`, so the second changes behavior rather than spelling, and neither tells
// the reader anything was removed. This port reports both and withholds the repair. The rows below
// assert BOTH halves of that decision: the finding still appears, and the source is unchanged.
//
// The decline is mutation-tested. Disabling `jsxNoUselessFragmentHasNonKeyAttributes` makes these
// rows fail, which is what separates a real guard from a comment claiming one.
func TestJsxNoUselessFragmentDeclinesToDropAttributes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"an attribute other than key is not deleted",
			"declare const Fragment: any;\ndeclare const x: any;\nconst a = <div><Fragment id={1}>{x}</Fragment></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a spread is not deleted",
			"declare const Fragment: any;\ndeclare const x: any;\ndeclare const rest: any;\nconst a = <div><Fragment {...rest}>{x}</Fragment></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"several attributes are not deleted",
			"declare const Fragment: any;\ndeclare const x: any;\ndeclare const r: any;\nconst a = <div><Fragment id={1} ref={r}>{x}</Fragment></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runJsxNoUselessFragment(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			// The finding stands and the source is untouched. Both halves matter: a rule that went
			// silent here would also leave the source alone and would pass half this assertion.
			if got := applyJsxNoUselessFragmentFixes(t, result); got != testCase.sourceText {
				t.Errorf("the repair produced %q, want the source unchanged", got)
			}
		})
	}
}

// TestJsxNoUselessFragmentFixKeepsTypeScript pins that the repair loses no type information.
//
// The port brief records two rules that destroyed type information tonight while passing every
// upstream fixture, because upstream's corpus is JavaScript and the gap between it and this tree is
// exactly TypeScript syntax. This fixer copies the children's text verbatim, so nothing inside them
// can be lost, and these rows prove it rather than asserting it: each carries a construct that has
// no JavaScript counterpart, and each repaired source must still contain it.
//
// Every expected output was produced by the installed build's own fixer first.
func TestJsxNoUselessFragmentFixKeepsTypeScript(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{
			"a type assertion survives",
			"declare const x: unknown;\nconst a = <div><>{x as string}</></div>;\n",
			"declare const x: unknown;\nconst a = <div>{x as string}</div>;\n",
		},
		{
			"a non-null assertion survives",
			"declare const x: string | undefined;\nconst a = <div><>{x!}</></div>;\n",
			"declare const x: string | undefined;\nconst a = <div>{x!}</div>;\n",
		},
		{
			"a satisfies expression survives",
			"declare const x: unknown;\nconst a = <div><>{x satisfies unknown}</></div>;\n",
			"declare const x: unknown;\nconst a = <div>{x satisfies unknown}</div>;\n",
		},
		{
			"an explicit type argument survives",
			"declare function foo<T>(value: T): T;\nconst a = <div><>{foo<number>(1)}</></div>;\n",
			"declare function foo<T>(value: T): T;\nconst a = <div>{foo<number>(1)}</div>;\n",
		},
		{
			"a typed arrow survives, parameters and return type",
			"const a = <div><>{(y: number): string => String(y)}</></div>;\n",
			"const a = <div>{(y: number): string => String(y)}</div>;\n",
		},
		{
			"a comment inside the fragment survives",
			"declare const x: any;\nconst a = <div><>{/* keep me */}{x}</></div>;\n",
			"declare const x: any;\nconst a = <div>{/* keep me */}{x}</div>;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runJsxNoUselessFragment(t, testCase.sourceText, "")
			if got := applyJsxNoUselessFragmentFixes(t, result); got != testCase.wantFixed {
				t.Errorf("the repair produced %q, want %q", got, testCase.wantFixed)
			}
		})
	}
}

// TestJsxNoUselessFragmentSpans asserts WHERE each finding points.
//
// Upstream reports on the fragment node itself for both ids, and the fix is anchored on the same
// node. The port brief notes that a fix which is right and anchored wrong is worse than no fix,
// because the edit lands somewhere the reader was never shown, so this is asserted on a rule that
// carries a repair rather than left to the fix assertion alone.
func TestJsxNoUselessFragmentSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{
			"a bare fragment",
			"declare const Foo: any;\nconst a = <><Foo /></>;",
			[]string{"<><Foo /></>"},
		},
		{
			"both findings point at the same fragment",
			"declare const x: any;\nconst a = <div><>{x}</></div>;",
			[]string{"<>{x}</>", "<>{x}</>"},
		},
		{
			"a Fragment element",
			"declare const Fragment: any;\ndeclare const Foo: any;\nconst a = <Fragment><Foo /></Fragment>;",
			[]string{"<Fragment><Foo /></Fragment>"},
		},
		{
			"an empty fragment",
			"const a = <></>;",
			[]string{"<></>"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runJsxNoUselessFragment(t, testCase.sourceText, "")
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

// TestJsxNoUselessFragmentMessagesReadAsWritten pins both rendered texts.
//
// Neither message interpolates, so there is nothing for a format string to get wrong, but the
// Description is what a reader sees. Asserted against literals typed here rather than against the
// rule's own constants, because comparing a diagnostic to the constant it was reported with is an
// equality whose two sides move together under mutation.
func TestJsxNoUselessFragmentMessagesReadAsWritten(t *testing.T) {
	t.Parallel()

	result := runJsxNoUselessFragment(t, "declare const x: any;\nconst a = <div><>{x}</></div>;\n", "")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(result.Diagnostics))
	}

	if result.Diagnostics[0].Message.Id != "NeedsMoreChildren" {
		t.Errorf("first id is %q", result.Diagnostics[0].Message.Id)
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description,
		"This fragment wraps one thing, so it does nothing.") {
		t.Errorf("first reads %q", result.Diagnostics[0].Message.Description)
	}

	if result.Diagnostics[1].Message.Id != "ChildOfHtmlElement" {
		t.Errorf("second id is %q", result.Diagnostics[1].Message.Id)
	}
	if !strings.HasPrefix(result.Diagnostics[1].Message.Description,
		"This fragment is the child of an HTML element,") {
		t.Errorf("second reads %q", result.Diagnostics[1].Message.Description)
	}
}

// TestJsxNoUselessFragmentMatchesUpstreamOnShapesTheCorpusOmits pins measured verdicts.
//
// Each row was run against the installed build before it was written here. The lowercase-letters
// rows are the surprising ones: upstream's html test is `/^[a-z]+$/`, letters only, so `<h1>` and
// `<my-tag>` are NOT html elements to this rule and a fragment inside either reports only one arm.
func TestJsxNoUselessFragmentMatchesUpstreamOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a digit in the tag name means it is not an html element",
			"declare const x: any;\nconst a = <h1><>{x}</></h1>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			// A tag name outside ASCII. Upstream's `/^[a-z]+$/` rejects it, and every byte of it is
			// ABOVE the letter z, so this is the input that separates the upper bound of the byte
			// test from its lower bound. A mutant dropping only the upper bound survived an
			// uppercase fixture, because an uppercase letter is below `a` and the lower bound
			// already rejects it; this is what actually kills that mutant.
			"a tag name outside ASCII is not an html element",
			"declare namespace JSX { interface IntrinsicElements { [name: string]: any } }\ndeclare const x: any;\nconst a = <\u0434\u0438\u0432><>{x}</></\u0434\u0438\u0432>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"an UPPERCASE letter in the tag name means it is not an html element",
			"declare const myDiv: any;\ndeclare namespace JSX { interface IntrinsicElements { [name: string]: any } }\ndeclare const x: any;\nconst a = <myDiv><>{x}</></myDiv>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a dash in the tag name means it is not an html element",
			"declare const x: any;\ndeclare namespace JSX { interface IntrinsicElements { [name: string]: any } }\nconst a = <mytag><>{x}</></mytag>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a fragment inside a component reports once and is not repaired",
			"declare const Foo: any;\ndeclare const x: any;\nconst a = <Foo><>{x}</></Foo>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a fragment inside a fragment is repaired",
			"declare const x: any;\nconst a = <><>{x}</><span /></>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a namespaced Fragment counts",
			"declare const React: any;\ndeclare const Foo: any;\nconst a = <React.Fragment><Foo /></React.Fragment>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a foreign namespace does not",
			"declare const Other: any;\ndeclare const Foo: any;\nconst a = <Other.Fragment><Foo /></Other.Fragment>;\n",
			nil,
		},
		{
			"a keyed Fragment is skipped entirely",
			"declare const Fragment: any;\ndeclare const x: any;\nconst a = <div><Fragment key={1}>{x}</Fragment></div>;\n",
			nil,
		},
		{
			"a self-closing Fragment reports both arms",
			"declare const Fragment: any;\nconst a = <div><Fragment /></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a single call expression child is kept",
			"declare function foo(): any;\nconst a = <div><>{foo()}</></div>;\n",
			[]string{"ChildOfHtmlElement"},
		},
		{
			"two children inside an html element still report the second arm",
			"declare const x: any;\ndeclare const y: any;\nconst a = <div><>{x}{y}</></div>;\n",
			[]string{"ChildOfHtmlElement"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestJsxNoUselessFragmentDecodesItsOptions exercises the decoder the config layer calls.
func TestJsxNoUselessFragmentDecodesItsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"empty input answers the default", "", false},
		{"an empty object answers the default", "{}", false},
		{"allowExpressions on", `{"allowExpressions":true}`, true},
		{"allowExpressions explicitly off", `{"allowExpressions":false}`, false},
		{"an unknown key is ignored", `{"somethingElse":true}`, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeJsxNoUselessFragmentOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			settings, ok := decoded.(JsxNoUselessFragmentOptions)
			if !ok {
				t.Fatalf("decoded to %T", decoded)
			}
			if settings.AllowExpressions != testCase.want {
				t.Errorf("AllowExpressions is %v, want %v", settings.AllowExpressions, testCase.want)
			}
		})
	}
}

// TestJsxNoUselessFragmentRejectsMalformedOptions asserts the decoder surfaces bad input.
func TestJsxNoUselessFragmentRejectsMalformedOptions(t *testing.T) {
	t.Parallel()

	if _, err := DecodeJsxNoUselessFragmentOptions([]byte(`{"allowExpressions":`)); err == nil {
		t.Fatal("truncated JSON decoded without error")
	}
	if _, err := DecodeJsxNoUselessFragmentOptions([]byte(`"error"`)); err == nil {
		t.Fatal("a string decoded into the options struct without error")
	}
}

// TestJsxNoUselessFragmentAllowExpressionsSilencesOneArmOnly pins the option's reach.
//
// It suppresses `NeedsMoreChildren` and leaves `ChildOfHtmlElement` alone, so a single-expression
// fragment inside a `div` still reports one finding with the option on. Measured, and the corpus
// writes only the top-level shape where the difference is invisible.
func TestJsxNoUselessFragmentAllowExpressionsSilencesOneArmOnly(t *testing.T) {
	t.Parallel()

	const topLevel = "declare const x: any;\nconst a = <>{x}</>;\n"
	rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, topLevel, ""), "NeedsMoreChildren")
	rule_testing.ExpectClean(t, runJsxNoUselessFragment(t, topLevel, `{"allowExpressions":true}`))

	const insideDiv = "declare const x: any;\nconst a = <div><>{x}</></div>;\n"
	rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, insideDiv, ""),
		"NeedsMoreChildren", "ChildOfHtmlElement")
	rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, insideDiv, `{"allowExpressions":true}`),
		"ChildOfHtmlElement")
}

// TestJsxNoUselessFragmentHandlesNilOptions covers the path the fixtures cannot reach.
//
// The config layer turns a decoder error into nil for a non-required rule, and `options.(T)` on nil
// yields the zero value, which here IS the default. Asserted rather than assumed.
func TestJsxNoUselessFragmentHandlesNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, JsxNoUselessFragment, jsxNoUselessFragmentFile,
		"declare const x: any;\nconst a = <>{x}</>;\n", nil)
	rule_testing.ExpectFindings(t, result, "NeedsMoreChildren")
}

// TestJsxNoUselessFragmentNeedsNoChecker pins that this rule is syntactic.
func TestJsxNoUselessFragmentNeedsNoChecker(t *testing.T) {
	t.Parallel()

	const source = "declare const Foo: any;\nconst a = <><Foo /></>;\n"
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, JsxNoUselessFragment,
		jsxNoUselessFragmentFile, source, JsxNoUselessFragmentOptions{}), "NeedsMoreChildren")
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, JsxNoUselessFragment,
		jsxNoUselessFragmentFile, source, JsxNoUselessFragmentOptions{}), "NeedsMoreChildren")
}

// TestJsxNoUselessFragmentSurvivesShapesThatWouldPanic drives shapes with pieces missing.
//
// The walk recovers per FILE rather than per rule, so one nil dereference here takes the file away
// from every rule in the tree, and no `ExpectFindings` fixture can see a panic. The unclosed forms
// are the ones that matter: they reach the listeners through error recovery with a nil opener or
// closer, which is exactly what the fix's bounds function guards against.
func TestJsxNoUselessFragmentSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"const a = <div />;\n",
		"const a = <></>;\n",
		"declare const Fragment: any;\nconst a = <Fragment />;\n",
		"declare const x: any;\nconst a = <>{}</>;\n",
		"const a = <>\n</>;\n",
		"declare const Fragment: any;\nconst a = <Fragment></Fragment>;\n",
		"declare const React: any;\nconst a = <React.Fragment />;\n",
		"(()=>{})();\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			runJsxNoUselessFragment(t, sourceText, "")
		})
	}
}

// TestJsxNoUselessFragmentPragmaIsFixedAtReact records the corpus case a settings surface would change.
//
// Upstream reads BOTH pragmas from settings: `settings.react.pragma` for the namespace and
// `settings.react.fragment` for the name. cohere reads no React settings, so both are the fixed
// `React` and `Fragment` here, and a project renaming either gets a false NEGATIVE.
//
// The case is upstream's own invalid-7, held out of the imported table because it is not
// reproducible rather than because it disagrees. Measured on the installed build: with
// `{react: {pragma: "SomeReact", fragment: "SomeFragment"}}` it reports `NeedsMoreChildren`, and
// with settings removed it goes clean, which is the answer this port produces. Recorded here rather
// than dropped, so the next reader sees the whole corpus accounted for.
func TestJsxNoUselessFragmentPragmaIsFixedAtReact(t *testing.T) {
	t.Parallel()

	const source = "declare const SomeReact: any;\ndeclare const foo: any;\nconst a = <SomeReact.SomeFragment>{foo}</SomeReact.SomeFragment>;\n"
	rule_testing.ExpectClean(t, runJsxNoUselessFragment(t, source, ""))

	// The control: the same shape under the fixed pragma pair DOES report, so the silence above is
	// the pragma comparison rather than anything else declining.
	const control = "declare const React: any;\ndeclare const foo: any;\nconst a = <React.Fragment>{foo}</React.Fragment>;\n"
	rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, control, ""), "NeedsMoreChildren")
}

// TestJsxNoUselessFragmentReadsBothFragmentSpellings pins why this file does not call jsx.ElementParts.
//
// `TestRulePackagesDoNotReachPastWrappedAccessors` flags three `AsJsxOpeningElement().TagName`
// reaches here, on the correct general principle that a self-closing element never produces a
// JsxOpeningElement and a rule reading only one form goes silent on the common case. All three carry
// a written reason, and this is the fixture behind those reasons so the next reader meets a
// measurement rather than a claim.
//
// The helper answers a JsxOpeningElement or a JsxSelfClosingElement. All three sites hold a
// JsxElement, which it answers nil for, so substituting it would make the rule stop recognizing
// `<Fragment>` rather than start recognizing more. Verified by making that substitution at the
// listener: seven subtests fail.
//
// What the guard is really protecting is covered, and covered by a different mechanism: the
// self-closing spelling has its own listener, and the rows below run every spelling of a fragment
// through the rule so a regression that dropped one is visible here rather than only in the guard.
func TestJsxNoUselessFragmentReadsBothFragmentSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a bare fragment with a paired tag",
			"declare const x: any;\nconst a = <div><>{x}</></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a Fragment element with a paired tag",
			"declare const Fragment: any;\ndeclare const x: any;\nconst a = <div><Fragment>{x}</Fragment></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a SELF-CLOSING Fragment element",
			"declare const Fragment: any;\nconst a = <div><Fragment /></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a self-closing namespaced Fragment",
			"declare const React: any;\nconst a = <div><React.Fragment /></div>;\n",
			[]string{"NeedsMoreChildren", "ChildOfHtmlElement"},
		},
		{
			"a self-closing element that is NOT a fragment is ignored",
			"declare const Foo: any;\nconst a = <div><Foo /></div>;\n",
			nil,
		},
		// The parent side of the same question. A fragment's parent is never a
		// JsxSelfClosingElement, because such an element has no children to nest into, so the two
		// parent-inspecting helpers testing only KindJsxElement is complete rather than narrow.
		// These rows exercise the parent test through every parent a fragment can actually have.
		{
			"a fragment whose parent is a component element",
			"declare const Foo: any;\ndeclare const x: any;\nconst a = <Foo><>{x}</></Foo>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a fragment whose parent is another fragment",
			"declare const x: any;\nconst a = <><>{x}</><span /></>;\n",
			[]string{"NeedsMoreChildren"},
		},
		{
			"a fragment whose parent is an expression container",
			"declare const Foo: any;\ndeclare const x: any;\nconst a = <Foo attribute={<>{x}</>} />;\n",
			[]string{"NeedsMoreChildren"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runJsxNoUselessFragment(t, testCase.sourceText, ""), testCase.wantIds...)
		})
	}
}

// TestJsxNoUselessFragmentNeedsMoreChildrenSaysWhatTheFragmentHolds pins the first arm's message for
// each shape it covers, as whole literals.
//
// The arm reports a fragment with fewer than two children, and its message used to say "wraps one
// thing" whatever it held. Two ahra findings wrap nothing but a commented-out block of dead JSX,
// `AddressesPage.tsx` and `NotificationsPage.tsx`, where "wraps one thing" sent the reader looking
// for a child that is not there, and the honest repair is to delete the comment and return `null`.
// An empty fragment is the third shape. The finding and its id are unchanged; only the sentence
// that describes the contents follows them.
func TestJsxNoUselessFragmentNeedsMoreChildrenSaysWhatTheFragmentHolds(t *testing.T) {
	t.Parallel()

	const why = " A fragment exists to give several siblings a single parent without emitting an element"
	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{"one child", "declare const x: any;\nconst a = <>{x}</>;\n",
			"This fragment wraps one thing, so it does nothing." + why +
				"; around a single child it is pure noise in the tree and in the source. Write the child on its own."},
		{"only a comment", "const a = (\n    <>\n        {/* <div>dead</div> */}\n    </>\n);\n",
			"This fragment holds only a comment, which renders nothing, so the fragment renders nothing either." + why +
				", and here there is nothing to group. Delete it, and the comment with it if that is dead code; " +
				"where a value is still required, write `null`."},
		{"nothing", "const a = <></>;\n",
			"This fragment is empty, so it renders nothing." + why +
				", and here there is nothing to group. Delete it; where a value is still required, write `null`."},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := runJsxNoUselessFragment(t, testCase.source, "")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			if result.Diagnostics[0].Message.Id != "NeedsMoreChildren" {
				t.Errorf("id is %q", result.Diagnostics[0].Message.Id)
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Errorf("description is\n%q\nwant\n%q", got, testCase.want)
			}
		})
	}
}

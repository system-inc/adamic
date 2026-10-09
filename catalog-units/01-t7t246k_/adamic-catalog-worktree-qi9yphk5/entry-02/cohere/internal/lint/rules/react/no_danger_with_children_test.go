package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// dangerWithChildrenFile is where the fixtures pretend to live.
//
// A `.tsx` name because most of the corpus is JSX and would not parse otherwise. Unlike
// `no-string-refs`, this rule has no JSX-file gate: the `createElement` half of it decides about
// ordinary call expressions and would be just as correct in a `.ts` file.
const dangerWithChildrenFile = "/repository/source/DangerWithChildren.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_danger_with_children.rs`: 17 pass and 14 fail from a
// single tester block. The snapshot carries 14 diagnostics against those 14 failing inputs, so it
// is one finding per input and there was no per-input count to recover.
//
// The strings were emitted by the extractor's own dumper rather than transcribed, and then checked
// back against the Rust source byte for byte by a script that searched for each one as both a raw
// string literal and an escaped one. All 31 matched. That check is not decoration: three porters
// before this one shipped a fixture the tool writing it had silently cooked, and reading them does
// not catch it.

// TestNoDangerWithChildrenFires covers every failing input upstream ships.
//
// The pattern in the list is worth naming because it looks like duplication and is not: nearly
// every case appears twice, once against `div` and once against `Hello`. The rule is deliberately
// indifferent to whether the tag is a host element or a component, and that indifference is
// upstream's, so dropping either half of each pair would leave it untested.
func TestNoDangerWithChildrenFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the danger prop and children between the tags", "\n        <div dangerouslySetInnerHTML={{ __html: \"HTML\" }}>\n            Children\n        </div>\n        "},
		{"the danger prop and a children prop", "<div dangerouslySetInnerHTML={{ __html: \"HTML\" }} children=\"Children\" />"},
		{"a spread carrying danger, with children between the tags", "\n        const props = { dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        <div {...props}>Children</div>\n        "},
		{"a spread carrying both danger and children", "\n        const props = { children: \"Children\", dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        <div {...props} />\n        "},
		{"a component with the danger prop and children between the tags", "\n        <Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }}>\n            Children\n        </Hello>\n        "},
		{"a component with the danger prop and a children prop", "<Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }} children=\"Children\" />"},
		{"a component whose only child is a single space", "<Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }}> </Hello>"},
		{"createElement with danger in the props and a third argument", "\n        React.createElement(\n            \"div\",\n            { dangerouslySetInnerHTML: { __html: \"HTML\" } },\n            \"Children\"\n        );\n        "},
		{"createElement with danger and children both in the props", "\n        React.createElement(\n            \"div\",\n            {\n                dangerouslySetInnerHTML: { __html: \"HTML\" },\n                children: \"Children\",\n            }\n        );\n        "},
		{"a component through createElement with danger and a third argument", "\n        React.createElement(\n            \"Hello\",\n            { dangerouslySetInnerHTML: { __html: \"HTML\" } },\n            \"Children\"\n        );\n        "},
		{"a component through createElement with danger and children in the props", "\n        React.createElement(\n            \"Hello\",\n            {\n                dangerouslySetInnerHTML: { __html: \"HTML\" },\n                children: \"Children\",\n            }\n        );\n        "},
		{"a spread variable carrying danger, with a third argument", "\n        const props = { dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        React.createElement(\"div\", props, \"Children\");\n        "},
		{"a spread variable carrying both danger and children", "\n        const props = { children: \"Children\", dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        React.createElement(\"div\", props);\n        "},
		{"a chain of spreads reaching children two levels down", "\n        const moreProps = { children: \"Children\" };\n        const otherProps = { ...moreProps };\n        const props = { ...otherProps, dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        React.createElement(\"div\", props);\n        "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
		})
	}
}

// TestNoDangerWithChildrenStaysSilent covers every passing input upstream ships.
//
// These are the cases that catch a port, and each declines for its own reason rather than by
// sharing one. Four turn on the checker answering something other than a variable declaration: a
// spread of a name this file never declares, a spread of `undefined`, a rest binding element, and
// `undefined` in the props position. Two turn on the props object carrying only one of the two
// props. Two turn on `createElement` being given exactly two arguments with no `children` key. One
// turns on a formatting newline not counting as a child. And one is a self-referential
// `const props = {...props}`, which is a cycle a naive recursion hangs on.
func TestNoDangerWithChildrenStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"children between the tags with no danger prop", "<div>Children</div>"},
		{"a spread of a variable this file does not declare", "<div {...props} />"},
		{"the danger prop with no children at all", "<div dangerouslySetInnerHTML={{ __html: \"HTML\" }} />"},
		{"a children prop with no danger prop", "<div children=\"Children\" />"},
		{"a spread whose object carries danger and no children", "\n        const props = { dangerouslySetInnerHTML: { __html: \"HTML\" } };\n        <div {...props} />\n        "},
		{"a spread whose object carries children and no danger", "\n        const moreProps = { className: \"eslint\" };\n        const props = { children: \"Children\", ...moreProps };\n        <div {...props} />\n        "},
		{"a rest binding element, which is not a variable declaration", "\n        const otherProps = { children: \"Children\" };\n        const { a, b, ...props } = otherProps;\n        <div {...props} />\n        "},
		{"a component with children and no danger prop", "<Hello>Children</Hello>"},
		{"a component with the danger prop and no children", "<Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }} />"},
		{"a component whose only child is a formatting newline", "\n        <Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }}>\n        </Hello>\n        "},
		{"createElement with danger in the props and no third argument", "React.createElement(\"div\", { dangerouslySetInnerHTML: { __html: \"HTML\" } });"},
		{"createElement with children but empty props", "React.createElement(\"div\", {}, \"Children\");"},
		{"a component through createElement with danger and no children", "React.createElement(\"Hello\", { dangerouslySetInnerHTML: { __html: \"HTML\" } });"},
		{"a component through createElement with children but empty props", "React.createElement(\"Hello\", {}, \"Children\");"},
		{"a spread of undefined, which resolves to no declaration", "<Hello {...undefined}>Children</Hello>"},
		{"undefined in the props position of createElement", "React.createElement(\"Hello\", undefined, \"Children\")"},
		{"a self-referential spread, which upstream guards against as a cycle", "\n        const props = {...props, scratch: {mode: 'edit'}};\n        const component = shallow(<TaskEditableTitle {...props} />);\n        "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoDangerWithChildrenPointsAtTheWholeElement pins where the finding lands.
//
// `ExpectFindings` asserts ids and a count and nothing else, so a rule pointing at the wrong node
// passes the whole corpus above. Upstream reports `jsx.span` and `call_expr.span`, which the
// snapshot confirms by underlining the entire element and the entire call, the call's span stopping
// before the trailing semicolon. The sibling rule in this package reports the attribute *name*
// instead, so this is a real choice between two shapes rather than the only thing a react rule can
// do, and getting it wrong here would underline one word where upstream underlines a block.
func TestNoDangerWithChildrenPointsAtTheWholeElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantReported string
	}{
		{
			"a self-closing element",
			"<div dangerouslySetInnerHTML={{ __html: \"HTML\" }} children=\"Children\" />",
			"<div dangerouslySetInnerHTML={{ __html: \"HTML\" }} children=\"Children\" />",
		},
		{
			"an element with a closing tag",
			"<Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }}> </Hello>",
			"<Hello dangerouslySetInnerHTML={{ __html: \"HTML\" }}> </Hello>",
		},
		{
			// The semicolon is outside the call's span, which is the detail a rule reporting the
			// statement rather than the call would get wrong while still underlining most of it.
			"a createElement call, whose span stops before the semicolon",
			"React.createElement(\"div\", { dangerouslySetInnerHTML: { __html: \"HTML\" } }, \"Children\");",
			"React.createElement(\"div\", { dangerouslySetInnerHTML: { __html: \"HTML\" } }, \"Children\")",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
			finding := result.Diagnostics[0]
			reported := testCase.sourceText[finding.Range.Pos():finding.Range.End()]
			if reported != testCase.wantReported {
				t.Fatalf("want the finding to cover %q, got %q", testCase.wantReported, reported)
			}
		})
	}
}

// TestNoDangerWithChildrenRendersItsMessage asserts the rendered text by equality.
//
// Equality rather than `strings.Contains`, because a predicate weaker than the property it guards
// is not a guard: a doubled word or a stray prefix is a substring match away from passing while the
// message is wrong. This message interpolates nothing, so the assertion is exact.
func TestNoDangerWithChildrenRendersItsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile,
		"<div dangerouslySetInnerHTML={{ __html: \"HTML\" }}>Children</div>")
	rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
	got := result.Diagnostics[0].Message.Description
	if got != messageNoDangerWithChildren.Description {
		t.Fatalf("want the message %q, got %q", messageNoDangerWithChildren.Description, got)
	}
	if !strings.HasPrefix(got, "This element sets `dangerouslySetInnerHTML`") {
		t.Fatalf("want the message to open by naming the prop, got %q", got)
	}
}

// TestNoDangerWithChildrenMatchesUpstreamsNarrowerCalleeTest pins the three ways this rule's callee
// test differs from the shared shelf helper.
//
// None of these appears in the imported corpus, which writes every call as `React.createElement`,
// so nothing upstream ships can tell the two apart. That is exactly why they are here: the shelf's
// `IsCreateElementCall` is a name-shaped invitation to simplify this rule into reporting where
// upstream is silent, and `internal/utilities/react`'s own doc comment names this rule when warning
// about it. Each case below would flip if someone swapped the inline test for the helper.
//
// oxc destructures `Expression::StaticMemberExpression` at `no_danger_with_children.rs:76` and then
// compares `callee.property.name`, so a bare call is not a member expression, an element access is
// not a static member, and `document` is never special-cased.
func TestNoDangerWithChildrenMatchesUpstreamsNarrowerCalleeTest(t *testing.T) {
	t.Parallel()

	danger := "{ dangerouslySetInnerHTML: { __html: \"HTML\" } }"

	t.Run("a bare createElement call is silent, where the shelf helper would accept it", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile,
			"createElement(\"div\", "+danger+", \"Children\");")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("a computed member is silent, where the shelf helper would accept it", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile,
			"React[\"createElement\"](\"div\", "+danger+", \"Children\");")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("document.createElement reports, where the shelf helper would reject it", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile,
			"document.createElement(\"div\", "+danger+", \"Children\");")
		rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
	})

	t.Run("any other object is accepted, since the receiver is never checked", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile,
			"Preact.createElement(\"div\", "+danger+", \"Children\");")
		rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
	})
}

// TestNoDangerWithChildrenReadsKeysThatNeedNoEvaluation covers the property-key shapes the corpus
// never writes.
//
// Upstream reads keys through `key.static_name()`, which answers for a string literal and for a
// computed key holding one, and declines anything it would have to evaluate. The corpus writes only
// bare identifier keys, so every one of these distinctions is invisible to it. The computed case is
// the one that has already cost this codebase: a porter documented it as a deliberate divergence
// when upstream in fact resolves and reports.
//
// The first case is also the one that would crash rather than misjudge. `Node.Text()` panics
// outright on a `ComputedPropertyName`, so a version reading key text before checking the kind
// takes the whole run down on any `{[k]: v}` anywhere in the tree.
func TestNoDangerWithChildrenReadsKeysThatNeedNoEvaluation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantFinding bool
	}{
		{
			"a computed key holding a string literal resolves and reports",
			"React.createElement(\"div\", { [\"dangerouslySetInnerHTML\"]: { __html: \"H\" } }, \"Children\");",
			true,
		},
		{
			"a computed key holding a variable names something else and is silent",
			"const dangerouslySetInnerHTML = \"x\";\nReact.createElement(\"div\", { [dangerouslySetInnerHTML]: { __html: \"H\" } }, \"Children\");",
			false,
		},
		{
			"a plain string literal key resolves and reports",
			"React.createElement(\"div\", { \"dangerouslySetInnerHTML\": { __html: \"H\" } }, \"Children\");",
			true,
		},
		{
			// A shorthand and a method are one `ObjectProperty` upstream and two distinct node
			// kinds here, so both are asked through the same accessor rather than listed.
			"a shorthand property carries a readable key",
			"const dangerouslySetInnerHTML = { __html: \"H\" };\nReact.createElement(\"div\", { dangerouslySetInnerHTML }, \"Children\");",
			true,
		},
		{
			"a method declaration carries a readable key",
			"React.createElement(\"div\", { dangerouslySetInnerHTML() {} }, \"Children\");",
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile, testCase.sourceText)
			if testCase.wantFinding {
				rule_testing.ExpectFindings(t, result, "noDangerWithChildren")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoDangerWithChildrenNeedsTheTypedHarness fails loudly if the checker declaration is reverted.
//
// The plain harness hands the rule a nil checker, so every spread-resolving case would go silent
// and every clean fixture above would keep passing vacuously while half the rule stopped working.
// Asserting the declaration directly is what makes that revert visible, since no fixture can see
// the difference between "correctly silent" and "silent because the checker was nil".
func TestNoDangerWithChildrenNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoDangerWithChildren.NeedsTypeChecker {
		t.Fatal("this rule resolves spread identifiers through the checker and must declare it")
	}
}

// TestNoDangerWithChildrenIsRegisteredWithoutItsNamespace guards the name the config has to match.
//
// The inventory writes `react/no-danger-with-children` and the parity guard strips the namespace on
// a `/` boundary, so a namespaced name here would lint zero files while every fixture above stayed
// green. That failure has already shipped once in `internal/rules/next/`.
func TestNoDangerWithChildrenIsRegisteredWithoutItsNamespace(t *testing.T) {
	t.Parallel()

	if NoDangerWithChildren.Name != "react/no-danger-with-children" {
		t.Fatalf("want the bare rule name, got %q", NoDangerWithChildren.Name)
	}
}

// TestNoDangerWithChildrenDeclinesCallsTooShortToCarryProps covers the argument-count gate.
//
// Added for a surviving mutant rather than from the corpus, which writes no call shorter than two
// arguments. The gate is not an optimization: the code below it reads `arguments.Nodes[1]`
// directly, so a call with one argument reaching that line panics with an index out of range and
// takes the whole run down. Loosening the gate to `<= 0` was scored and does exactly that, which is
// how this case earned its place.
//
// Upstream states the same gate as its own first move, with the reason that a call this short is
// not a proper createElement call.
func TestNoDangerWithChildrenDeclinesCallsTooShortToCarryProps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a createElement call with one argument", "React.createElement(\"div\");"},
		{"a createElement call with no arguments at all", "React.createElement();"},
		{"an unrelated one-argument call", "shallow(\"div\");"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDangerWithChildren, dangerWithChildrenFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

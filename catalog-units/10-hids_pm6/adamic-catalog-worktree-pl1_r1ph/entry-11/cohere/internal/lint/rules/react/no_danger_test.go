package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noDangerFile is where the fixtures pretend to live.
//
// A .tsx extension because the rule reads JSX, and NO file-suffix gate exists in this rule. Three
// shipped react rules here gate on .tsx or .jsx, which is oxc-era residue rather than upstream
// behavior, and TestNoDangerHasNoFileSuffixGate below pins the absence so nobody reintroduces it by
// pattern-matching a neighbour.
const noDangerFile = "/repository/source/Danger.tsx"

// runNoDanger routes every option fixture through the rule's OWN exported decoder rather than
// building the options struct directly.
//
// That is what puts the decoder under test, and it is the one line with no upstream counterpart.
// A fixture handing RunWithOptions a struct would leave both the default fallback and the JSON tag
// untested, and either could be wrong while every fixture stayed green.
//
// The raw JSON here is the BARE object, not upstream's [{...}] array. cohere's config layer unwraps
// the [severity, options] tuple before dispatch, so copying ESLint's spelling into a fixture fails
// on every row.
func runNoDanger(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	var decoded any
	if rawOptions == "" {
		var err error
		decoded, err = DecodeNoDangerOptions(nil)
		if err != nil {
			t.Fatalf("decoding absent options: %v", err)
		}
	} else {
		var err error
		decoded, err = DecodeNoDangerOptions(json.RawMessage(rawOptions))
		if err != nil {
			t.Fatalf("decoding %s: %v", rawOptions, err)
		}
	}
	return rule_testing.RunWithOptions(t, NoDanger, noDangerFile, sourceText, decoded)
}

// The corpus is eslint-plugin-react's own, imported verbatim from tests/lib/rules/no-danger.js by
// evaluating the upstream tester with a stub RuleTester and serializing what it was handed, so no
// case was retyped and no escape could be cooked on the way in. 6 valid and 7 invalid.
//
// The options column matters more here than in most corpora: four of the 6 valid cases and six of
// the 7 invalid ones carry customComponentNames, and one input appears in BOTH lists separated
// only by its pattern. <TextMUI> is clean under ["MUI*"] and reports under ["*MUI"], which is the
// single best case in this corpus because it pins the glob dialect rather than the rule's shape.
//
// The last invalid case names THREE entries in its errors array, so its finding count is upstream's
// statement rather than my inference.
func TestNoDangerFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		findings   []string
	}{
		{"<div dangerouslySetInnerHTML={{ __html: \"\" }}></div>;", "", []string{"dangerousProp"}},
		{"<App dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;", "{\"customComponentNames\":[\"*\"]}", []string{"dangerousProp"}},
		{"\n        function App() {\n          return <Title dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"Title\"]}", []string{"dangerousProp"}},
		{"\n        function App() {\n          return <TextFoo dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"*Foo\"]}", []string{"dangerousProp"}},
		{"\n        function App() {\n          return <FooText dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"Foo*\"]}", []string{"dangerousProp"}},
		{"\n        function App() {\n          return <TextMUI dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"*MUI\"]}", []string{"dangerousProp"}},
		{"\n        import type { ComponentProps } from \"react\";\n\n        const Comp = \"div\";\n        const Component = () => <></>;\n\n        const NestedComponent = (_props: ComponentProps<\"div\">) => <></>;\n\n        Component.NestedComponent = NestedComponent;\n\n        function App() {\n          return (\n            <>\n              <div dangerouslySetInnerHTML={{ __html: \"<div>aaa</div>\" }} />\n              <Comp dangerouslySetInnerHTML={{ __html: \"<div>aaa</div>\" }} />\n\n              <Component.NestedComponent\n                dangerouslySetInnerHTML={{ __html: '<div>aaa</div>' }}\n              />\n            </>\n          );\n        }\n      ", "{\"customComponentNames\":[\"*\"]}", []string{"dangerousProp", "dangerousProp", "dangerousProp"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoDanger(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

func TestNoDangerStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{"<App />;", ""},
		{"<App dangerouslySetInnerHTML={{ __html: \"\" }} />;", ""},
		{"<div className=\"bar\"></div>;", ""},
		{"<div className=\"bar\"></div>;", "{\"customComponentNames\":[\"*\"]}"},
		{"\n        function App() {\n          return <Title dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"Home\"]}"},
		{"\n        function App() {\n          return <TextMUI dangerouslySetInnerHTML={{ __html: \"<span>hello</span>\" }} />;\n        }\n      ", "{\"customComponentNames\":[\"MUI*\"]}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoDanger(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Beyond the imported corpus. Every verdict below was measured against the installed
// eslint-plugin-react through the Linter API before being written here, except where noted as a
// deliberate divergence.

// TestNoDangerElementNameBoundary pins WHICH tags count as DOM elements.
//
// Upstream renders the tag to a string and tests /^[a-z]/ against it, so this is a first-character
// test on a rendered name rather than a check for an intrinsic element. The dotted and namespaced
// lines are the ones that look wrong and are not: a port testing "is the tag a lowercase Identifier"
// would be silent on both and pass every imported fixture, because the corpus writes neither.
func TestNoDangerElementNameBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"a plain DOM element", "<div dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp"}},
		{"a custom element, still lowercase", "<x-custom dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp"}},
		{"a dotted tag whose first character is lowercase", "<a.b dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp"}},
		{"a namespaced tag whose first character is lowercase", "<div:ns dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp"}},
		{"a component is silent by the capital", "<App dangerouslySetInnerHTML={{__html:''}} />;", nil},
		{"and so is a dotted one", "<A.B dangerouslySetInnerHTML={{__html:''}} />;", nil},
		{"a paired tag reports on the attribute of the opening element", "<div dangerouslySetInnerHTML={{__html:''}}></div>;", []string{"dangerousProp"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDanger(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoDangerAttributeNameBoundary pins WHICH attribute spellings count.
//
// The comparison is exact and case-sensitive, and the attribute's VALUE is never examined. The last
// two lines matter for a reason no fixture makes obvious: reaching for the attribute's expression
// would be the natural way to write this rule and would panic on the string-valued form, which the
// port brief names as a crash class that costs a whole file.
func TestNoDangerAttributeNameBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"the exact spelling", "<div dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp"}},
		{"lowercased is a different attribute", "<div dangerouslysetinnerhtml={{__html:''}} />;", nil},
		{"uppercased likewise", "<div DANGEROUSLYSETINNERHTML={{__html:''}} />;", nil},
		{"a namespaced attribute has no plain name to read", "<div ns:dangerouslySetInnerHTML={{__html:''}} />;", nil},
		{"an unrelated attribute", "<div className=\"bar\" />;", nil},
		{"a spread carries no name and is skipped, even though it may supply the property", "<div {...props} />;", nil},
		{"a bare attribute with no value still reports", "<div dangerouslySetInnerHTML />;", []string{"dangerousProp"}},
		{"and a string-valued one does too, which is where a value-reading port would panic", "<div dangerouslySetInnerHTML=\"x\" />;", []string{"dangerousProp"}},
		{"written twice reports twice, because the loop does not stop at the first", "<div dangerouslySetInnerHTML={{__html:''}} dangerouslySetInnerHTML={{__html:''}} />;", []string{"dangerousProp", "dangerousProp"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDanger(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoDangerCustomComponentNames pins the glob dialect and the two declined upstream defects.
//
// The first four rows are the dialect, and MUI* against TextMUI is the one that looks wrong: both
// minimatch and filepath.Match say false, and upstream's corpus has exactly that pair in its VALID
// list, so a port that "fixed" it would break an imported clean case.
//
// The last three rows are DELIBERATE DIVERGENCES, not measurements of agreement, and each is stated
// at the line. Upstream compares against a name it builds with a shallow expression that is not the
// one it renders elsewhere, and the two shapes below are where that expression breaks.
func TestNoDangerCustomComponentNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		findings   []string
	}{
		{"a star matches any component", "<App dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["*"]}`, []string{"dangerousProp"}},
		{"an exact name matches itself", "<Title dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["Title"]}`, []string{"dangerousProp"}},
		{"a different exact name does not", "<Title dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["Home"]}`, nil},
		{"a trailing star anchors the prefix, so MUI* does not match TextMUI", "<TextMUI dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["MUI*"]}`, nil},
		{"a leading star does match it", "<TextMUI dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["*MUI"]}`, []string{"dangerousProp"}},
		{"an empty list checks nothing extra", "<App dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":[]}`, nil},
		{"an absent key behaves as an empty list", "<App dangerouslySetInnerHTML={{__html:''}} />;", `{}`, nil},
		{"a DOM element still reports whatever the option says", "<div dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["Nothing"]}`, []string{"dangerousProp"}},

		// DIVERGENCE. Upstream builds the compared name with an expression that does not recurse
		// through a nested member, so <A.B.C /> renders as "undefined.C" there. Measured
		// behaviorally: with ["A.B.C"] upstream is SILENT and with ["undefined.C"] it REPORTS.
		// Declined, because carrying it means matching a pattern nobody would write.
		{"a nested member renders fully here, which upstream does not", "<A.B.C dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["A.B.C"]}`, []string{"dangerousProp"}},
		{"a two-part member matches under both", "<A.B dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["A.B"]}`, []string{"dangerousProp"}},

		// DIVERGENCE, and this one is a CRASH upstream rather than a wrong answer. The same
		// expression yields the name NODE for a namespaced tag, minimatch calls split on an object
		// and throws a TypeError that takes down the lint run for that file. Reproduced directly
		// against the installed build. Declined, because reproducing it would take a file away
		// from every rule, the walk recovering per file rather than per rule.
		//
		// The tag has to carry a CAPITAL namespace to reach the option path at all, which is a
		// distinction I got wrong first and the fixture caught. A lowercase one renders as
		// "ns:App", whose first character is lowercase, so it is a DOM element by the regex above
		// and reports with no options set. Both lines measured on the installed build: <ns:App>
		// with no options REPORTS, and <Ns:App> with any customComponentNames CRASHES.
		{"a lowercase namespace is a DOM element and needs no option", "<ns:App dangerouslySetInnerHTML={{__html:''}} />;", "", []string{"dangerousProp"}},
		{"a capital namespace is not, and is silent unconfigured", "<Ns:App dangerouslySetInnerHTML={{__html:''}} />;", "", nil},
		{"and matches its rendered name rather than crashing, which is where upstream throws", "<Ns:App dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["Ns:App"]}`, []string{"dangerousProp"}},
		{"declining cleanly on a pattern that does not match, where upstream also throws", "<Ns:App dangerouslySetInnerHTML={{__html:''}} />;", `{"customComponentNames":["Other"]}`, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoDanger(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoDangerDecodesOptions exercises the decoder directly, including the unconfigured path.
//
// A rule configured as a bare severity is handed nil rather than a struct, and the port brief
// records a rule that registered on 3,407 files and was completely inert because nothing in its
// suite reached that path. Every fixture above goes through the decoder, so this is what covers the
// nil input and the JSON tag, neither of which has an upstream counterpart.
func TestNoDangerDecodesOptions(t *testing.T) {
	t.Parallel()

	t.Run("absent options give the documented default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoDangerOptions(nil)
		if err != nil {
			t.Fatalf("decoding nil: %v", err)
		}
		settings, ok := decoded.(NoDangerOptions)
		if !ok {
			t.Fatalf("decoded to %T", decoded)
		}
		if len(settings.CustomComponentNames) != 0 {
			t.Fatalf("default custom names were %v, wanted none", settings.CustomComponentNames)
		}
	})

	t.Run("the wire name is customComponentNames", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoDangerOptions(json.RawMessage(`{"customComponentNames":["A","B*"]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		settings := decoded.(NoDangerOptions)
		if len(settings.CustomComponentNames) != 2 || settings.CustomComponentNames[0] != "A" || settings.CustomComponentNames[1] != "B*" {
			t.Fatalf("decoded to %v", settings.CustomComponentNames)
		}
	})

	t.Run("an unknown key is accepted and ignored, matching a schema without additionalProperties false", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoDangerOptions(json.RawMessage(`{"somethingElse":1}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		if len(decoded.(NoDangerOptions).CustomComponentNames) != 0 {
			t.Fatal("unknown key changed the options")
		}
	})

	t.Run("the rule falls back to the default when handed nil rather than the struct", func(t *testing.T) {
		t.Parallel()
		// This bypasses the decoder entirely, which is the one path every other fixture misses.
		//
		// The COMPONENT is the case that measures anything, and a first version of this used a
		// <div>. That was a fixture over a distinction it could not see: a div reports under the
		// real default AND under a wrong fallback of ["*"], so both paths reach the same verdict
		// and a mutation of the fallback SURVIVED. A component separates them, because it reports
		// only if the fallback wrongly matches everything.
		//
		// This is the port brief's point that a survivor needs an input on which the two versions
		// produce different OUTPUT rather than merely take different internal paths.
		component := rule_testing.RunWithOptions(t, NoDanger, noDangerFile, "<App dangerouslySetInnerHTML={{__html:''}} />;", nil)
		rule_testing.ExpectClean(t, component)

		element := rule_testing.RunWithOptions(t, NoDanger, noDangerFile, "<div dangerouslySetInnerHTML={{__html:''}} />;", nil)
		rule_testing.ExpectFindings(t, element, "dangerousProp")
	})
}

// TestNoDangerSpanAndMessage asserts WHERE the finding points and WHAT it says.
//
// Upstream reports on the JSXAttribute rather than on the element, measured: the span for
// <div dangerouslySetInnerHTML={{__html:'}} /> runs columns 6 to 43, which is the attribute alone.
// ExpectFindings cannot see any of this, so a rule anchoring on the element would pass every fixture
// above while pointing somewhere the reader was never shown.
func TestNoDangerSpanAndMessage(t *testing.T) {
	t.Parallel()

	t.Run("the finding points at the attribute, not the element", func(t *testing.T) {
		t.Parallel()
		const sourceText = "const a = <div dangerouslySetInnerHTML={{__html:''}} />;"
		const want = "dangerouslySetInnerHTML={{__html:''}}"
		result := runNoDanger(t, sourceText, "")
		rule_testing.ExpectFindings(t, result, "dangerousProp")
		reported := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != want {
			t.Fatalf("reported on %q, wanted %q", reported, want)
		}
	})

	t.Run("a bare attribute spans just the name", func(t *testing.T) {
		t.Parallel()
		const sourceText = "const a = <div dangerouslySetInnerHTML />;"
		result := runNoDanger(t, sourceText, "")
		rule_testing.ExpectFindings(t, result, "dangerousProp")
		reported := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "dangerouslySetInnerHTML" {
			t.Fatalf("reported on %q", reported)
		}
	})

	t.Run("the message names the property and explains the defect", func(t *testing.T) {
		t.Parallel()
		result := runNoDanger(t, "<div dangerouslySetInnerHTML={{__html:''}} />;", "")
		rule_testing.ExpectFindings(t, result, "dangerousProp")
		got := result.Diagnostics[0].Message.Description
		// Compared against a literal typed here rather than against the rule's own format string,
		// because a comparison to the rule's own constant moves with it under mutation. Asserted as
		// a prefix rather than with Contains, since a weaker predicate than the property it guards
		// is not a guard.
		const wantPrefix = "This element sets dangerouslySetInnerHTML, which hands React a raw string"
		if !strings.HasPrefix(got, wantPrefix) {
			t.Fatalf("message was %q, wanted it to start %q", got, wantPrefix)
		}
		if result.Diagnostics[0].Message.Id != "dangerousProp" {
			t.Fatalf("id was %q", result.Diagnostics[0].Message.Id)
		}
	})
}

// TestNoDangerDoesNotPanicOnUnusualTagNames is crash protection, not a behavioral fixture.
//
// ast.Node.Text panics on a PropertyAccessExpression, measured with a probe, so a rule rendering an
// arbitrary tag name through it would take the whole FILE down for all of this package's rules. No
// ExpectFindings fixture can see a panic, so the sweep reports SURVIVED identically whether the
// part-by-part rendering matters or not. This names what it prevents.
func TestNoDangerDoesNotPanicOnUnusualTagNames(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"<a.b.c.d dangerouslySetInnerHTML={{__html:''}} />;",
		"<this.Foo dangerouslySetInnerHTML={{__html:''}} />;",
		"<A.B.C.D className=\"x\" />;",
		"<ns:tag className=\"x\" />;",
		"<a.b></a.b>;",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			// The verdict is not the point; surviving the render is. Both verdicts are pinned in
			// the boundary tables above for the shapes that have one.
			runNoDanger(t, sourceText, `{"customComponentNames":["*"]}`)
		})
	}
}

// TestNoDangerHasNoFileSuffixGate pins the ABSENCE of a suffix gate.
//
// Three shipped react rules in this package gate on .tsx or .jsx, which is oxc-era residue rather
// than upstream behavior, and it blinds them to most of this tree. Measured against
// eslint-plugin-react with the same source under four suffixes: all four report.
//
// A .ts file cannot hold JSX, so only the two JSX-capable suffixes can carry this rule's subject.
// That is a fact about the language rather than about the rule, and it is stated here rather than
// left as an unexplained narrowing.
func TestNoDangerHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const sourceText = "<div dangerouslySetInnerHTML={{__html:''}} />;"
	for _, fileName := range []string{
		"/repository/source/Danger.tsx",
		"/repository/source/Danger.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoDangerOptions(nil)
			if err != nil {
				t.Fatalf("decoding: %v", err)
			}
			result := rule_testing.RunWithOptions(t, NoDanger, fileName, sourceText, decoded)
			rule_testing.ExpectFindings(t, result, "dangerousProp")
		})
	}
}

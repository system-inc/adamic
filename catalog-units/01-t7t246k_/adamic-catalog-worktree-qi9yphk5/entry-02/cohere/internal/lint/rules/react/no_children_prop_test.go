package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// childrenPropFile is where the fixtures pretend to live.
//
// A `.tsx` extension rather than `.ts`, and it is load-bearing rather than cosmetic. The harness
// picks its script kind from the suffix, and `.ts` parses `<div children />` as a type assertion
// instead of JSX, producing a tree with no JsxAttribute node in it at all. Every JSX fixture below
// would then pass by finding nothing, which reads exactly like a rule that works.
const childrenPropFile = "/repository/source/Children.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/react/no_children_prop.rs`,
// pulled from oxc's inline Tester block and written into this file by a script rather than
// by hand, so no transcription step existed that could cook an escape. One Tester block, 38 pass
// and 15 fail, and the snapshot records 15 diagnostics from those 15 inputs, so one finding per
// input is measured here rather than assumed.
func TestNoChildrenPropFires(t *testing.T) {
	t.Parallel()

	cases := []string{
		"<div children />;",
		"<div children=\"Children\" />;",
		"<div children={<div />} />;",
		"<div children={[<div />, <div />]} />;",
		"<div children=\"Children\">Children</div>;",
		"React.createElement(\"div\", {children: \"Children\"});",
		"React.createElement(\"div\", {children: \"Children\"}, \"Children\");",
		"React.createElement(\"div\", {children: React.createElement(\"div\")});",
		"React.createElement(\"div\", {children: [React.createElement(\"div\"), React.createElement(\"div\")]});",
		"<MyComponent children=\"Children\" />",
		"React.createElement(MyComponent, {children: \"Children\"});",
		"<MyComponent className=\"class-name\" children=\"Children\" />;",
		"React.createElement(MyComponent, {children: \"Children\", className: \"class-name\"});",
		"<MyComponent {...props} children=\"Children\" />;",
		"React.createElement(MyComponent, {...props, children: \"Children\"})",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText), "noChildrenProp")
		})
	}
}

// The clean cases are the whole discrimination and each declines for a different reason.
//
// Three groups are worth naming because a port can drop any of them and still look finished. The
// `createElement("div", undefined)` family checks that a second argument which is not an object
// literal is passed over rather than searched. The `React.createElement("div", "Children")` family
// checks that a child in the props position is not read as props. And `foo(MyComponent, {...props,
// children: "Children"})` is the sharpest: it writes the offending property exactly, and is clean
// only because the callee is not a createElement call, so a port that forgot the callee gate would
// report it while every other clean case still passed.
func TestNoChildrenPropStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"<div />;",
		"<div></div>;",
		"React.createElement(\"div\", {});",
		"React.createElement(\"div\", undefined);",
		"<div className=\"class-name\"></div>;",
		"React.createElement(\"div\", {className: \"class-name\"});",
		"<div>Children</div>;",
		"React.createElement(\"div\", \"Children\");",
		"React.createElement(\"div\", {}, \"Children\");",
		"React.createElement(\"div\", undefined, \"Children\");",
		"<div className=\"class-name\">Children</div>;",
		"React.createElement(\"div\", {className: \"class-name\"}, \"Children\");",
		"<div><div /></div>;",
		"React.createElement(\"div\", React.createElement(\"div\"));",
		"React.createElement(\"div\", {}, React.createElement(\"div\"));",
		"React.createElement(\"div\", undefined, React.createElement(\"div\"));",
		"<div><div /><div /></div>;",
		"React.createElement(\"div\", React.createElement(\"div\"), React.createElement(\"div\"));",
		"React.createElement(\"div\", {}, React.createElement(\"div\"), React.createElement(\"div\"));",
		"React.createElement(\"div\", undefined, React.createElement(\"div\"), React.createElement(\"div\"));",
		"React.createElement(\"div\", [React.createElement(\"div\"), React.createElement(\"div\")]);",
		"React.createElement(\"div\", {}, [React.createElement(\"div\"), React.createElement(\"div\")]);",
		"React.createElement(\"div\", undefined, [React.createElement(\"div\"), React.createElement(\"div\")]);",
		"<MyComponent />",
		"React.createElement(MyComponent);",
		"React.createElement(MyComponent, {});",
		"React.createElement(MyComponent, undefined);",
		"<MyComponent>Children</MyComponent>;",
		"React.createElement(MyComponent, \"Children\");",
		"React.createElement(MyComponent, {}, \"Children\");",
		"React.createElement(MyComponent, undefined, \"Children\");",
		"<MyComponent className=\"class-name\"></MyComponent>;",
		"React.createElement(MyComponent, {className: \"class-name\"});",
		"<MyComponent className=\"class-name\">Children</MyComponent>;",
		"React.createElement(MyComponent, {className: \"class-name\"}, \"Children\");",
		"<MyComponent className=\"class-name\" {...props} />;",
		"foo(MyComponent, {...props, children: \"Children\"})",
		"React.createElement(MyComponent, {className: \"class-name\", ...props});",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText))
		})
	}
}

// Where every finding points, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule reporting the whole attribute
// or the whole call expression passes all 53 cases above while underlining most of the line. oxc
// reports the *name* alone, and its snapshot pins that: `<div children />` is labelled at column 6
// spanning eight characters, and `React.createElement("div", {children: "Children"})` at column 29
// spanning the same eight. Both spellings of upstream disagree here, ESLint reporting the enclosing
// node and oxc reporting the key, so this is the assertion that says which one was ported.
//
// The wanted text is `children` in every row, which is the point: it should never widen to
// `children="Children"` or to the call.
func TestNoChildrenPropPointsAtTheName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantOffset int
	}{
		{`<div children />;`, 5},
		{`<div children="Children" />;`, 5},
		{`<div children="Children">Children</div>;`, 5},
		{`<MyComponent className="class-name" children="Children" />;`, 36},
		{`React.createElement("div", {children: "Children"});`, 28},
		{`React.createElement(MyComponent, {children: "Children", className: "class-name"});`, 34},
		{`React.createElement(MyComponent, {...props, children: "Children"})`, 44},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoChildrenProp, childrenPropFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0].Range
			if reported.Pos() != testCase.wantOffset {
				t.Errorf("finding starts at %d, want %d", reported.Pos(), testCase.wantOffset)
			}
			if got := testCase.sourceText[reported.Pos():reported.End()]; got != "children" {
				t.Errorf("finding underlines %q, want %q", got, "children")
			}
		})
	}
}

// The rendered message, asserted exactly rather than by substring.
//
// A fixture whose predicate is weaker than the property it guards is not a guard: a `Contains`
// check on a message that had grown a duplicated fragment would still pass. This message carries no
// interpolation at all, which is itself the thing being pinned, since a later edit adding a `%s`
// without a value would render a stray verb and every other test here would stay green.
func TestNoChildrenPropRendersItsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoChildrenProp, childrenPropFile, `<div children />;`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoChildrenProp.Description {
		t.Errorf("rendered %q, want %q", got, messageNoChildrenProp.Description)
	}
	if strings.Contains(messageNoChildrenProp.Description, "%") {
		t.Errorf("message carries a format verb but is reported without arguments: %q",
			messageNoChildrenProp.Description)
	}
}

// The four object-literal member kinds oxc reports and our AST splits apart.
//
// oxc's tree has one `ObjectProperty` covering a plain property, a shorthand, a method and an
// accessor, so its corpus needed only the plain form to exercise all four. Ours gives each its own
// node kind, so a port matching only `KindPropertyAssignment` passes every one of the 53 imported
// cases while being silent on three shapes upstream reports.
//
// Not inferred from reading oxc's types. These five inputs were added to oxc's own pass vector and
// its test run, which reports a pass case that produced a diagnostic, named exactly these as
// reporting. `{[children]: 1}` and `{2: 1}` stayed silent in the same run and are the clean side
// below.
func TestNoChildrenPropCoversEveryPropertyKind(t *testing.T) {
	t.Parallel()

	reports := []string{
		`React.createElement("div", {children});`,
		`React.createElement("div", {children() {}});`,
		`React.createElement("div", {get children() { return 1; }});`,
		`React.createElement("div", {set children(value) {}});`,
		`React.createElement("div", {"children": 1});`,
		`React.createElement("div", {["children"]: 1});`,
		// A template with no substitutions names the property as literally as a string does, and
		// oxc answers for it through `static_name`'s `single_quasi` arm. Only the bracketed
		// position exists: a bare `` {`children`: 1} `` is not valid JavaScript and our parser
		// produces no object literal for it at all, which is why `staticKeyName` reads a template
		// only through the computed branch. A sweep dropping that branch survived all 53 imported
		// fixtures, so this case is the only thing standing between it and a silent regression.
		"React.createElement(\"div\", {[`children`]: 1});",
	}
	for _, sourceText := range reports {
		t.Run("reports "+sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText), "noChildrenProp")
		})
	}

	// A computed key holding an identifier names a variable rather than the property, so its real
	// property name is not known here. oxc declines it through `static_name`, whose own
	// documentation gives `[a]: 1` as the returning-None example, and its test run confirmed the
	// silence rather than leaving it read from the type definitions.
	//
	// `{2: 1}` is a numeric key. `staticKeyName` declines it by kind rather than reading it, which
	// is a deliberate narrowing of oxc's `static_name`: no number renders as `children`, so keeping
	// the branch could not change a verdict, and a sweep dropping it survived every fixture for
	// exactly that reason. This case pins the silence whichever way that decision is later revised,
	// since a numeric key must stay clean under both spellings.
	silent := []string{
		`React.createElement("div", {[children]: 1});`,
		`React.createElement("div", {2: 1});`,
		`React.createElement("div", {[someKey]: 1});`,
	}
	for _, sourceText := range silent {
		t.Run("silent on "+sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText))
		})
	}
}

// A computed key must not crash the linter, which is not a hypothetical.
//
// `Node.Text()` panics outright on a ComputedPropertyName rather than returning empty, and a probe
// written before the rule existed hit exactly that panic. So `staticKeyName` checks the kind before
// reading any text. This pins the order: a refactor that reads the text first and filters by kind
// afterwards takes down every file in the tree containing a computed key, and no imported fixture
// covers one.
func TestNoChildrenPropSurvivesAComputedKey(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile,
		"declare const key: string;\nReact.createElement(\"div\", {[key]: 1});\n"))
}

// A template key with substitutions, which no imported case writes.
//
// oxc's `static_name` answers for a template only through `single_quasi`, so a substituting
// template returns None there. `staticKeyName` declines it by kind for the same reason: its text is
// not the property name, and reading the cooked text of the first chunk would report on a property
// that is actually called something else at runtime.
func TestNoChildrenPropDeclinesASubstitutingTemplateKey(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile,
		"declare const part: string;\nReact.createElement(\"div\", {[`children${part}`]: 1});\n"))
}

// The callee gate, exercised past what upstream's single `foo(...)` case reaches.
//
// `IsCreateElementCall` is shared shelf code with no other production caller, so this rule is its
// first real consumer and its behavior is pinned here rather than trusted. Three directions matter
// and the imported corpus tests none of them: a bare `createElement` is the modern spelling and
// must report, `document.createElement` shares the property name and constructs a DOM node so must
// not, and a computed member is the same call written differently and must report.
func TestNoChildrenPropReadsTheCalleeTheWayTheShelfDoes(t *testing.T) {
	t.Parallel()

	reports := []string{
		`createElement("div", {children: 1});`,
		`React["createElement"]("div", {children: 1});`,
		`Preact.createElement("div", {children: 1});`,
		`(React).createElement("div", {children: 1});`,
	}
	for _, sourceText := range reports {
		t.Run("reports "+sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText), "noChildrenProp")
		})
	}

	silent := []string{
		`document.createElement("div", {children: 1});`,
		`React.createFragment("div", {children: 1});`,
	}
	for _, sourceText := range silent {
		t.Run("silent on "+sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText))
		})
	}
}

// A namespaced attribute name, which our shelf declines and no imported case writes.
//
// `<svg xlink:children="x" />` parses with a JsxNamespacedName rather than an Identifier, and
// `jsx.AttributeName` returns not-named for it. oxc reaches the same answer by destructuring
// `JSXAttributeName::Identifier` and returning early on anything else, so the silence is upstream's
// judgment rather than a gap in ours.
func TestNoChildrenPropDeclinesANamespacedAttributeName(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile,
		`<svg xlink:children="x" />;`))
}

// An object writing the property twice reports once, matching oxc's `find_map`.
//
// Upstream stops at the first match and emits one diagnostic per call expression. Reporting both
// would be a divergence nothing asked for, and the imported corpus never writes a duplicate key so
// it cannot say which behavior was ported.
func TestNoChildrenPropReportsOncePerCall(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile,
		`React.createElement("div", {children: 1, "children": 2});`), "noChildrenProp")
}

// Two JSX attributes on one element report twice, which is the opposite of the case above.
//
// The JSX side listens per attribute rather than per element, so an element writing the prop twice
// produces two findings. Upstream does the same, listening on `JSXAttribute`. The pair of tests
// exists because "reports once" and "reports per occurrence" are both defensible and only one of
// them is upstream's, per surface.
func TestNoChildrenPropReportsPerJsxAttribute(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile,
		`<div children="a" children="b" />;`), "noChildrenProp", "noChildrenProp")
}

// Malformed and degenerate input must not take the linter down.
//
// The tree is not a fixture: it holds files mid-edit and files the parser recovers from, and a rule
// that indexes an argument list or reads a key's text without checking a kind first crashes the
// whole run rather than reporting a finding. `Node.Text()` panicking on a ComputedPropertyName is
// the concrete instance already hit here, and these are the neighbours of it.
//
// `<div {...{children: 1}} />` is the one worth naming: it writes the property inside a spread's
// own object literal, which is genuinely invisible to both tools. Neither reports it, because the
// JSX side listens per attribute and a spread is not one, and the call side never runs on JSX at
// all. It is here so the silence is a recorded decision rather than an unexamined gap.
func TestNoChildrenPropSurvivesDegenerateInput(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		`React.createElement("div", {});`,
		`React.createElement();`,
		`React.createElement("div",);`,
		`React.createElement(...args);`,
		`<div {...{children: 1}} />;`,
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("panicked rather than reporting: %v", recovered)
				}
			}()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText))
		})
	}
}

// TestNoChildrenPropAllowFunctions runs upstream's twelve rows that pass `allowFunctions`, through
// the registered decoder, plus the controls that pin what the option changes.
//
// Eight clean: a function as the `children` prop, in JSX and in a props object, in each of the four
// function spellings. Eight reporting: the same functions passed the other way, nested as the only
// child or as createElement's third argument. Upstream's messages for those are `nestFunction` and
// `passFunctionAsArgs`, and they point at the whole element and the whole call.
func TestNoChildrenPropAllowFunctions(t *testing.T) {
	t.Parallel()

	allowed, err := rule.DecodeOptionsInto[NoChildrenPropOptions]()([]byte(`{"allowFunctions": true}`))
	if err != nil {
		t.Fatalf("decoding allowFunctions: %v", err)
	}
	functions := []string{"() => {}", "function() {}", "async function() {}", "function* () {}"}

	for _, function := range functions {
		clean := []string{
			"<MyComponent children={" + function + "} />;",
			"React.createElement(MyComponent, {children: " + function + "});",
		}
		for _, sourceText := range clean {
			t.Run(sourceText, func(t *testing.T) {
				t.Parallel()
				rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoChildrenProp, childrenPropFile, sourceText, allowed))
				// The control: without the option the same source reports, so the option is what
				// made it clean.
				rule_testing.ExpectFindings(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, sourceText), "noChildrenProp")
			})
		}

		reporting := []struct {
			sourceText string
			wantId     string
			wantSpan   string
		}{
			{"<MyComponent>{" + function + "}</MyComponent>;", "nestFunction", "<MyComponent>{" + function + "}</MyComponent>"},
			{"React.createElement(MyComponent, {}, " + function + ");", "passFunctionAsArgs", "React.createElement(MyComponent, {}, " + function + ")"},
		}
		for _, testCase := range reporting {
			t.Run(testCase.sourceText, func(t *testing.T) {
				t.Parallel()
				result := rule_testing.RunWithOptions(t, NoChildrenProp, childrenPropFile, testCase.sourceText, allowed)
				rule_testing.ExpectFindings(t, result, testCase.wantId)
				diagnostic := result.Diagnostics[0]
				if got := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.wantSpan {
					t.Errorf("the finding points at %q, want %q", got, testCase.wantSpan)
				}
				// Without the option, a function passed this way is ordinary children.
				rule_testing.ExpectClean(t, rule_testing.Run(t, NoChildrenProp, childrenPropFile, testCase.sourceText))
			})
		}
	}

	// Under the option a non-function children prop still reports, and a function child with
	// company is not the one-child shape.
	stillReporting := []string{
		`<MyComponent children="Children" />;`,
		`React.createElement(MyComponent, {children: "Children"});`,
	}
	for _, sourceText := range stillReporting {
		t.Run(sourceText+" under allowFunctions", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoChildrenProp, childrenPropFile, sourceText, allowed), "noChildrenProp")
		})
	}
	// Measured against the installed build under the option, beyond upstream's rows: parentheses,
	// which ESTree does not materialize, and a method or getter, whose ESTree value is a function.
	t.Run("a parenthesized function nested as the only child under allowFunctions", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoChildrenProp, childrenPropFile,
			"<MyComponent>{(() => {})}</MyComponent>;", allowed), "nestFunction")
	})
	for _, sourceText := range []string{
		"<MyComponent>text {() => {}}</MyComponent>;",
		"React.createElement(MyComponent, {}, 'a', () => {});",
		"React.createElement(MyComponent, null, () => {});",
		"<MyComponent children={(() => {})} />;",
		"React.createElement(MyComponent, {children() {}});",
		"React.createElement(MyComponent, {get children() { return 1 }});",
	} {
		t.Run(sourceText+" under allowFunctions", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoChildrenProp, childrenPropFile, sourceText, allowed))
		})
	}
}

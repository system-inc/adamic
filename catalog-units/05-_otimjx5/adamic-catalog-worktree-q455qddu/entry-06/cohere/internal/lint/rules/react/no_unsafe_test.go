package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// unsafeFile is where the fixtures pretend to live.
//
// A `.tsx` name because upstream's corpus is written for a JSX source type and upstream gates the
// whole rule on it. Unlike oxc this rule declares no file gate; see the note on the rule about
// `should_run`.
const unsafeFile = "/repository/source/Unsafe.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every upstream case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_unsafe.rs`: 10 pass and 4 fail, from a single tester
// block. The extractor warned of a discrepancy, 12 diagnostics against 4 fail inputs, and the
// warning is real here rather than the multi-snapshot artefact the brief describes: there is one
// snapshot file, it carries 12 location headers, and each of the 4 fail inputs writes 3 lifecycle
// members, so the partition is 3 findings per input and needed no recovery by line alignment.
//
// The strings were decoded from the extractor's `-dump` output as JSON and then compared byte
// against byte into the Rust source by a script before any Go was written. All 14 matched, and a
// second script confirmed none of them contains a backslash or a quote, so the escape-cooking
// hazard that has bitten four previous porters could not arise here.
//
// # Four of the fourteen cases are byte-identical pairs across an option axis
//
// Upstream's corpus is 10 distinct sources, not 14. Cases 7 and 11 are the same bytes, as are 9
// and 13, differing only in a configured React version; cases 7 and 11 likewise for `checkAliases`.
// The pairs split by `checkAliases` are reproduced here on both sides of the flag. The pairs split
// by React version cannot be, and the two pass halves are recorded as reporting; the reasoning is
// at the line.
func TestNoUnsafeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		checkAliases bool
		wantIds      []string
	}{
		// Upstream, verbatim.
		{"unprefixed lifecycle methods on a class component under checkAliases", "\n\t\t\t        class Foo extends React.Component {\n\t\t\t          componentWillMount() {}\n\t\t\t          componentWillReceiveProps() {}\n\t\t\t          componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", true, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},
		{"unprefixed lifecycle properties on a createReactClass component under checkAliases", "\n\t\t\t        const Foo = createReactClass({\n\t\t\t          componentWillMount: function() {},\n\t\t\t          componentWillReceiveProps: function() {},\n\t\t\t          componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", true, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},
		{"prefixed lifecycle methods on a class component", "\n\t\t\t        class Foo extends React.Component {\n\t\t\t          UNSAFE_componentWillMount() {}\n\t\t\t          UNSAFE_componentWillReceiveProps() {}\n\t\t\t          UNSAFE_componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", false, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},
		{"prefixed lifecycle properties on a createReactClass component", "\n\t\t\t        const Foo = createReactClass({\n\t\t\t          UNSAFE_componentWillMount: function() {},\n\t\t\t          UNSAFE_componentWillReceiveProps: function() {},\n\t\t\t          UNSAFE_componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", false, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},
		// The two cases below are upstream PASS cases recorded here as reporting, deliberately.
		//
		// Each is clean upstream only because its tester block configures
		// `settings.react.version = "16.2.0"`, which predates the `UNSAFE_` prefix, and each is
		// byte-identical to a FAIL case above that differs only in carrying `"16.3.0"`. cohere
		// reads no React settings, so no React version can reach a rule here
		// and oxc's own default branch, `is_none_or(supports_unsafe_lifecycle_prefix)`, is the only
		// reachable one. Measured rather than reasoned about: with no version configured that exact
		// source produces 3 findings on the release binary, and with `"version": "16.2.0"` configured
		// it produces 0, in the same session against the same file. The cases are pinned as reporting,
		// which is upstream's answer under the configuration we can actually produce, and the reasoning
		// is here so the next reader neither deletes them nor breaks the rule to green them.
		{"upstream's eighth pass case, clean only under a React version we cannot configure", "\n\t\t\t        class Foo extends React.Component {\n\t\t\t          UNSAFE_componentWillMount() {}\n\t\t\t          UNSAFE_componentWillReceiveProps() {}\n\t\t\t          UNSAFE_componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", false, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},
		{"upstream's tenth pass case, clean only under a React version we cannot configure", "\n\t\t\t        const Foo = createReactClass({\n\t\t\t          UNSAFE_componentWillMount: function() {},\n\t\t\t          UNSAFE_componentWillReceiveProps: function() {},\n\t\t\t          UNSAFE_componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", false, []string{messageUnsafeComponentWillMount.Id, messageUnsafeComponentWillReceiveProps.Id, messageUnsafeComponentWillUpdate.Id}},

		// Everything below is ours rather than upstream's, and every one was measured on the release
		// oxlint binary rather than reasoned about. They cover the discriminations our abstract syntax
		// tree spells differently than oxc's does, the shapes upstream's corpus is silent on, and the
		// two places our own shelf would have given a different answer.

		// The component bases, all four spellings. Upstream's corpus writes only `React.Component`, so
		// nothing imported covers the bare identifiers or the pure variant.
		{"a class extending the bare Component identifier", "class Foo extends Component {\n  UNSAFE_componentWillMount() {}\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		{"a class extending the bare PureComponent identifier", "class Foo extends PureComponent {\n  UNSAFE_componentWillMount() {}\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		{"a class extending React.PureComponent", "class Foo extends React.PureComponent {\n  UNSAFE_componentWillMount() {}\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		{"a class expression extending React.Component", "const Foo = class extends React.Component {\n  UNSAFE_componentWillMount() {}\n};\n", false, []string{messageUnsafeComponentWillMount.Id}},
		// The namespaced factory spelling of the one accepted name. oxc's `is_es5_component` keys both
		// arms on `createReactClass`, so this counts while `React.createClass` does not; both measured.
		{"the namespaced createReactClass factory", "const Foo = React.createReactClass({\n  UNSAFE_componentWillUpdate: function() {},\n});\n", false, []string{messageUnsafeComponentWillUpdate.Id}},
		// Parentheses are skipped on the factory callee and nowhere else. The two silent heritage
		// spellings are in the clean list below; all three were measured in one invocation.
		{"a parenthesized createReactClass callee", "const Foo = (createReactClass)({\n  UNSAFE_componentWillMount: function() {},\n});\n", false, []string{messageUnsafeComponentWillMount.Id}},
		// The key is resolved statically rather than matched as an identifier. oxc calls
		// `static_name()`, which answers for a string-literal key and for a computed key holding one.
		// A port matching only the identifier spelling loses both findings, and reading `Text()` on the
		// computed one crashes the linter outright.
		{"a computed lifecycle key holding a string literal", "class Foo extends React.Component {\n  [\"UNSAFE_componentWillMount\"]() {}\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		{"a string-literal lifecycle key", "class Foo extends React.Component {\n  \"UNSAFE_componentWillReceiveProps\"() {}\n}\n", false, []string{messageUnsafeComponentWillReceiveProps.Id}},
		// Nothing asks what kind of member carries the name. A static method is not a lifecycle method
		// at all and reports, and both accessors report, because oxc reaches all three through one
		// `MethodDefinition` arm and never asks. Reproduced rather than corrected.
		{"a static method named for the lifecycle", "class Foo extends React.Component {\n  static UNSAFE_componentWillUpdate() {}\n}\n", false, []string{messageUnsafeComponentWillUpdate.Id}},
		{"a getter named for the lifecycle", "class Foo extends React.Component {\n  get UNSAFE_componentWillMount() { return 1; }\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		{"a setter named for the lifecycle", "class Foo extends React.Component {\n  set UNSAFE_componentWillMount(v) {}\n}\n", false, []string{messageUnsafeComponentWillMount.Id}},
		// Upstream's ownership hole on the object arm. The walk stops at the first factory call above
		// the property, however far away and whatever lies between, so an unrelated object literal
		// nested inside a `createReactClass` call is judged as that component's own lifecycle property.
		// Its class-component twin is in the clean list, and the pair is the whole asymmetry.
		{"an unrelated object literal nested inside a createReactClass component", "const Foo = createReactClass({\n  render: function() {\n    const notMine = { UNSAFE_componentWillMount: function() {} };\n    return null;\n  },\n});\n", false, []string{messageUnsafeComponentWillMount.Id}},
		// checkAliases widens to the three unprefixed spellings and to nothing else, so a name that is
		// neither family stays silent under it. The prefixed spellings are unaffected by the flag.
		{"an unprefixed lifecycle method under checkAliases", "class Foo extends React.Component {\n  componentWillReceiveProps() {}\n}\n", true, []string{messageUnsafeComponentWillReceiveProps.Id}},
		{"a prefixed lifecycle method under checkAliases", "class Foo extends React.Component {\n  UNSAFE_componentWillUpdate() {}\n}\n", true, []string{messageUnsafeComponentWillUpdate.Id}},
		{"an unprefixed lifecycle property on a createReactClass component under checkAliases", "const Foo = createReactClass({\n  componentWillUpdate: function() {},\n});\n", true, []string{messageUnsafeComponentWillUpdate.Id}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnsafe, unsafeFile, testCase.sourceText,
				NoUnsafeOptions{CheckAliases: testCase.checkAliases})
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoUnsafeStaysSilent holds every clean case, upstream's and ours.
//
// Six of upstream's ten pass cases are clean for the option default rather than for anything
// structural: they write the unprefixed spellings in real components, and `checkAliases` is false.
// Two more are clean for the React version, and those two are in the fires list instead. The
// remaining upstream cases here are the structural ones.
func TestNoUnsafeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		checkAliases bool
	}{
		// Upstream, verbatim.
		{"a class component using only safe lifecycle methods", "\n\t\t\t        class Foo extends React.Component {\n\t\t\t          componentDidUpdate() {}\n\t\t\t          render() {}\n\t\t\t        }\n\t\t\t      ", false},
		{"a createReactClass component using only safe lifecycle methods", "\n\t\t\t        const Foo = createReactClass({\n\t\t\t          componentDidUpdate: function() {},\n\t\t\t          render: function() {}\n\t\t\t        });\n\t\t\t      ", false},
		// The ownership gate. `Bar` is not a React component base, so neither spelling reports however
		// the members are named. These two are the structural half of upstream's pass list.
		{"unprefixed lifecycle methods on a class extending something that is not a component", "\n\t\t\t        class Foo extends Bar {\n\t\t\t          componentWillMount() {}\n\t\t\t          componentWillReceiveProps() {}\n\t\t\t          componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", false},
		{"prefixed lifecycle methods on a class extending something that is not a component", "\n\t\t\t        class Foo extends Bar {\n\t\t\t          UNSAFE_componentWillMount() {}\n\t\t\t          UNSAFE_componentWillReceiveProps() {}\n\t\t\t          UNSAFE_componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", false},
		// The same gate on the object arm. `bar(...)` is not the factory, so the walk finds nothing.
		{"unprefixed lifecycle properties on an object passed to an unrelated factory", "\n\t\t\t        const Foo = bar({\n\t\t\t          componentWillMount: function() {},\n\t\t\t          componentWillReceiveProps: function() {},\n\t\t\t          componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", false},
		{"prefixed lifecycle properties on an object passed to an unrelated factory", "\n\t\t\t        const Foo = bar({\n\t\t\t          UNSAFE_componentWillMount: function() {},\n\t\t\t          UNSAFE_componentWillReceiveProps: function() {},\n\t\t\t          UNSAFE_componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", false},
		// These two are real components carrying the real deprecated names, and they are clean purely
		// because `checkAliases` defaults to false. Each is byte-identical to a fail case above that
		// differs only in the flag, which is the pair that pins the option surface exists at all. Our
		// inventory records this rule as taking no options; these two cases are why that is wrong.
		{"unprefixed lifecycle methods on a real class component with checkAliases left off", "\n\t\t\t        class Foo extends React.Component {\n\t\t\t          componentWillMount() {}\n\t\t\t          componentWillReceiveProps() {}\n\t\t\t          componentWillUpdate() {}\n\t\t\t        }\n\t\t\t      ", false},
		{"unprefixed lifecycle properties on a real createReactClass component with checkAliases left off", "\n\t\t\t        const Foo = createReactClass({\n\t\t\t          componentWillMount: function() {},\n\t\t\t          componentWillReceiveProps: function() {},\n\t\t\t          componentWillUpdate: function() {},\n\t\t\t        });\n\t\t\t      ", false},

		// Everything below is ours rather than upstream's, and every one was measured on the release
		// oxlint binary rather than reasoned about.

		// A class field is not a method definition. oxc matches `AstKind::MethodDefinition` and never
		// looks at `PropertyDefinition`, so all three field spellings are silent. This is a real
		// difference from the sibling rule `no-will-update-set-state`, which does reach class fields,
		// and both ports are faithful to their own upstream.
		{"an arrow-valued lifecycle class field", "class Foo extends React.Component {\n  UNSAFE_componentWillMount = () => {};\n}\n", false},
		{"a function-expression-valued lifecycle class field", "class Foo extends React.Component {\n  UNSAFE_componentWillReceiveProps = function() {};\n}\n", false},
		{"a bare lifecycle class field with no initializer", "class Foo extends React.Component {\n  UNSAFE_componentWillUpdate;\n}\n", false},
		// The heritage clause is read without skipping parentheses, matching `heritage_expression()`
		// followed by `as_member_expression()`, neither of which sees through one. Our tree keeps the
		// parenthesis as a real node exactly where oxc's declines it, so both spellings are silent here
		// for the same reason they are silent upstream. Measured alongside the reporting
		// `(createReactClass)(...)` case in the fires list, in one invocation.
		{"a parenthesized React.Component heritage expression", "class Foo extends (React.Component) {\n  UNSAFE_componentWillMount() {}\n}\n", false},
		{"a parenthesized receiver inside the heritage expression", "class Foo extends (React).Component {\n  UNSAFE_componentWillMount() {}\n}\n", false},
		// The shelf's `IsEs5ComponentCall` accepts both of these and oxc accepts neither, because
		// `is_es5_component` keys both arms on the single constant `createReactClass`. Reaching for the
		// shelf would report here and no imported fixture writes either spelling.
		{"a bare createClass factory", "const Foo = createClass({\n  UNSAFE_componentWillMount: function() {},\n});\n", false},
		{"the React.createClass factory", "const Foo = React.createClass({\n  UNSAFE_componentWillMount: function() {},\n});\n", false},
		// Another namespace's Component is somebody else's base class.
		{"a class extending another namespace's Component", "class Foo extends Foo2.Component {\n  UNSAFE_componentWillMount() {}\n}\n", false},

		// Both of these were found by a mutation sweep rather than by reading, and both are measured
		// silent on the release binary. Widening either receiver check survived the whole imported
		// corpus and every case above, because upstream writes neither spelling.
		{"a namespaced createReactClass on a receiver that is not React", "const Foo = Bar.createReactClass({\n  UNSAFE_componentWillMount: function() {},\n});\n", false},
		{"a class extending a React member that is not a component base", "class Foo extends React.Anything {\n  UNSAFE_componentWillMount() {}\n}\n", false},

		// A TypeScript `implements` clause is a heritage clause too, and oxc never sees it: it reads
		// `heritage_expression()`, which is the superclass alone. Dropping the token check survived
		// every other case in this file, and both spellings are measured silent on the release
		// binary. The second is the one that matters, because it carries a real extends clause as
		// well and so cannot be declined by the absence of heritage.
		{"a lifecycle method on a class implementing a Component interface", "interface Component {}\nclass Foo implements Component {\n  UNSAFE_componentWillMount() {}\n}\n", false},
		{"a class extending a non-component while implementing Component", "interface Component {}\nclass Foo extends Bar implements Component {\n  UNSAFE_componentWillMount() {}\n}\n", false},
		// The object arm never asks about a class. A lifecycle-shaped property on a plain object
		// nested inside a real class component is silent, while the identical property nested inside a
		// `createReactClass` call reports. That pair is the whole ownership asymmetry and its reporting
		// half is in the fires list.
		{"an unrelated object literal nested inside a class component", "class Foo extends React.Component {\n  render() {\n    const notMine = { UNSAFE_componentWillMount: function() {} };\n    return null;\n  }\n}\n", false},
		{"a lifecycle-shaped property on a bare top-level object literal", "const Foo = { UNSAFE_componentWillMount: function() {} };\n", false},
		// A class with no heritage clause at all is the shape the ownership gate mostly exists to stay
		// silent on.
		{"a lifecycle method on a class extending nothing", "class Foo {\n  UNSAFE_componentWillMount() {}\n}\n", false},
		// checkAliases widens to exactly three names. A neighbouring lifecycle method that was never
		// deprecated stays silent under it, which pins that the flag is a name list rather than a
		// prefix-stripping rule.
		{"a safe lifecycle method under checkAliases", "class Foo extends React.Component {\n  componentDidMount() {}\n  componentWillUnmount() {}\n}\n", true},
		// A computed key that names nothing statically resolves to no name, so the rule must decline it
		// rather than panic. `ast.TryGetTextOfPropertyName` answers false here where `Text()` would
		// crash the whole run.
		{"a computed lifecycle key holding a non-literal expression", "const key = whatever;\nclass Foo extends React.Component {\n  [key]() {}\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnsafe, unsafeFile, testCase.sourceText,
				NoUnsafeOptions{CheckAliases: testCase.checkAliases})
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnsafeReportsAtTheKey pins where the finding points and what it renders.
//
// `ExpectFindings` asserts message identifiers and count and nothing else, so a rule pointing at
// the whole member rather than at its key passes a complete fixture pair while being wrong.
// Upstream's span is `key.span()` on both arms: all twelve snapshot diagnostics underline the name
// alone, stopping before the parenthesis on a method and before the colon on a property. Both
// spellings are checked, and so are the two key shapes that are not identifiers, because a port
// reporting the whole member would pass a span assertion written only against the dotted method.
//
// The rendered description is compared with equality rather than `strings.Contains`, because a
// predicate weaker than the property it guards is not a guard: a truncated or doubled rendering
// contains the needle and stays green.
func TestNoUnsafeReportsAtTheKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantText    string
		wantMessage rule.Message
	}{
		{
			"a class method underlines the name and not the parameter list",
			"class Foo extends React.Component {\n  UNSAFE_componentWillMount() {}\n}\n",
			"UNSAFE_componentWillMount", messageUnsafeComponentWillMount,
		},
		{
			"an object property underlines the key and not the value",
			"const Foo = createReactClass({\n  UNSAFE_componentWillReceiveProps: function() {},\n});\n",
			"UNSAFE_componentWillReceiveProps", messageUnsafeComponentWillReceiveProps,
		},
		{
			"a string-literal key underlines the quotes as well",
			"class Foo extends React.Component {\n  \"UNSAFE_componentWillUpdate\"() {}\n}\n",
			"\"UNSAFE_componentWillUpdate\"", messageUnsafeComponentWillUpdate,
		},
		{
			// The brackets are outside upstream's span and the quotes are inside it, measured in
			// one invocation: the computed key underlines from column 4 and the string-literal key
			// above it from column 3. oxc's property key for a computed member is the inner
			// expression with no wrapper node, so `key.span()` never covers a bracket; ours is a
			// `ComputedPropertyName` that does, and the rule strips it. Every identifier assertion
			// in this file was green with the brackets included, which is what this case exists to
			// stop.
			"a computed key underlines the inner literal and not the brackets",
			"class Foo extends React.Component {\n  [\"UNSAFE_componentWillMount\"]() {}\n}\n",
			"\"UNSAFE_componentWillMount\"", messageUnsafeComponentWillMount,
		},
		{
			// The object arm strips brackets identically, measured in the same way. Both arms share
			// one reporting closure, so this case is what stops a later split from diverging.
			"a computed object key underlines the inner literal and not the brackets",
			"const Foo = createReactClass({\n  [\"UNSAFE_componentWillUpdate\"]: function() {},\n});\n",
			"\"UNSAFE_componentWillUpdate\"", messageUnsafeComponentWillUpdate,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnsafe, unsafeFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			if reported := testCase.sourceText[finding.Range.Pos():finding.Range.End()]; reported != testCase.wantText {
				t.Errorf("reported span is %q, want %q", reported, testCase.wantText)
			}
			if got := finding.Message.Id; got != testCase.wantMessage.Id {
				t.Errorf("reported identifier is %q, want %q", got, testCase.wantMessage.Id)
			}
			if got := finding.Message.Description; got != testCase.wantMessage.Description {
				t.Errorf("rendered description is %q, want %q", got, testCase.wantMessage.Description)
			}
		})
	}
}

// TestNoUnsafeNamesTheRightReplacement pins that the three findings are not interchangeable.
//
// Upstream renders one message shape whose help text names a different replacement per lifecycle
// family, and this port splits that into three identifiers so a fixture can tell them apart. That
// only buys a guard if something asserts the mapping, and a single class writing all three members
// is the input that does it: a rule reporting the same identifier three times, or reporting them in
// the wrong order, passes every other test in this file.
func TestNoUnsafeNamesTheRightReplacement(t *testing.T) {
	t.Parallel()

	sourceText := "class Foo extends React.Component {\n" +
		"  UNSAFE_componentWillMount() {}\n" +
		"  UNSAFE_componentWillReceiveProps() {}\n" +
		"  UNSAFE_componentWillUpdate() {}\n" +
		"}\n"

	result := rule_testing.Run(t, NoUnsafe, unsafeFile, sourceText)
	rule_testing.ExpectFindings(t, result,
		messageUnsafeComponentWillMount.Id,
		messageUnsafeComponentWillReceiveProps.Id,
		messageUnsafeComponentWillUpdate.Id)

	// The identifier and the span have to agree, which is the mapping this test exists for. Reading
	// the reported text back out of the source is what makes a transposed pair visible.
	wantByIdentifier := map[string]string{
		messageUnsafeComponentWillMount.Id:        "UNSAFE_componentWillMount",
		messageUnsafeComponentWillReceiveProps.Id: "UNSAFE_componentWillReceiveProps",
		messageUnsafeComponentWillUpdate.Id:       "UNSAFE_componentWillUpdate",
	}
	for _, finding := range result.Diagnostics {
		reported := sourceText[finding.Range.Pos():finding.Range.End()]
		if want := wantByIdentifier[finding.Message.Id]; reported != want {
			t.Errorf("finding %q points at %q, want %q", finding.Message.Id, reported, want)
		}
	}
}

// TestNoUnsafeDeclinesTheUntypedHarnessNothing records that this rule needs no type checker.
//
// It resolves everything from the syntax tree: a property key by `ast.TryGetTextOfPropertyName`,
// and component membership by what a class extends and what a call is named. Nothing asks which
// declaration an identifier binds to, so `class Foo extends Component` reports whether or not
// `Component` is React's, exactly as upstream does, and the rule declares no checker. Determined by
// reading what upstream's helpers ask rather than by whether they are reached through a context:
// `is_es5_component` and `is_es6_component` both destructure the node and compare names, and
// neither calls `ctx.scoping()` or resolves a reference.
//
// Pinned as a test rather than left in prose so that adding a resolution step later fails here.
func TestNoUnsafeDeclinesTheUntypedHarnessNothing(t *testing.T) {
	t.Parallel()

	if NoUnsafe.NeedsTypeChecker {
		t.Fatal("the rule declares a type checker it does not use")
	}
	result := rule_testing.Run(t, NoUnsafe, unsafeFile,
		"class Foo extends Component {\n  UNSAFE_componentWillMount() {}\n}\n")
	rule_testing.ExpectFindings(t, result, messageUnsafeComponentWillMount.Id)
}

package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoUndefFile is where the fixtures pretend to live.
//
// A `.tsx` name, matching upstream's own snapshot file name (`jsx_no_undef.tsx`) and required for
// the sources below to parse as JSX at all.
const jsxNoUndefFile = "/repository/source/JsxNoUndef.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Block 1 of `oxc/crates/oxc_linter/src/rules/react/jsx_no_undef.rs` carries 12 pass and 8 fail
// cases, and the snapshot reports 8 diagnostics against those 8 fail inputs, so exactly one finding
// per input and no per-input recovery was needed. Block 2 is a second tester that is not
// snapshotted and carries one pair over `let x = <A.B />;`, distinguished only by a `globals`
// option. Its fail half is below. Its pass half is expressed in TestJsxNoUndefAllowGlobals, with a
// global declared by a script file standing in for ESLint's `globals`.
//
// Every string was decoded from the extractor's `-dump` output and then checked byte against byte
// into the Rust source by script before any Go was written. All 22 matched, which is worth stating
// because two of them are multi-line raw strings whose leading newline and eight-space indentation
// are part of what upstream tests.
func TestJsxNoUndefFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"an undeclared uppercase component", "var React; React.render(<App />);", []string{"jsxIdentifierNotDefined"}},
		{"an undeclared uppercase member object", "var React; React.render(<Appp.Foo />);", []string{"jsxIdentifierNotDefined"}},
		{"an undeclared lowercase member object", "var React; React.render(<appp.Foo />);", []string{"jsxIdentifierNotDefined"}},
		{"an undeclared lowercase member object, two levels deep", "var React; React.render(<appp.foo.Bar />);", []string{"jsxIdentifierNotDefined"}},
		{"another undeclared uppercase component", "var React; React.render(<Foo />);", []string{"jsxIdentifierNotDefined"}},
		{"an undeclared name mentioned as a bare expression too", "var React; Unknown; React.render(<Unknown />)", []string{"jsxIdentifierNotDefined"}},
		{"a binding declared only inside a nested block", "var React; { const App = null; }; React.render(<App />);", []string{"jsxIdentifierNotDefined"}},
		{"an enum member that is not a value binding", "var React; enum A { App }; React.render(<App />);", []string{"jsxIdentifierNotDefined"}},
		// Upstream block 2, the fail half. The pass half of this pair sets `globals: {A: "readonly"}`
		// and has no expressible form here.
		{"an undeclared member object with no surrounding code", "let x = <A.B />;", []string{"jsxIdentifierNotDefined"}},

		// Cases below are ours, not upstream's, and each covers a distinction our tree makes that
		// oxc's corpus never exercises.

		// The intrinsic split is a parser variant upstream and a text predicate here, so the three
		// shapes that a reading of "lowercase means intrinsic" gets wrong need their own cases.
		// Measured against the release oxlint binary before being written: all three report there.
		{"an underscore-led name, which upstream treats as a reference", "var React; React.render(<_foo />);", []string{"jsxIdentifierNotDefined"}},
		{"a dollar-led name, which upstream treats as a reference", "var React; React.render(<$foo />);", []string{"jsxIdentifierNotDefined"}},
		{"a non-ASCII name, always a reference whatever its case", "var React; React.render(<אב />);", []string{"jsxIdentifierNotDefined"}},

		// The member walk has to reach the leftmost object rather than the immediate one. A single
		// step would read `A.B` as the object of `A.B.C.D` and resolve nothing either way, so this
		// case only distinguishes the two once the object is declared, which the pass twin covers.
		{"a four-deep member expression reports on its leftmost object", "var React; React.render(<A.B.C.D />);", []string{"jsxIdentifierNotDefined"}},

		// A lowercase member object is a reference, so the case predicate must not be applied to
		// it. Applying it there would silence this and two upstream fail cases at once.
		{"a lowercase leftmost object with an undeclared name", "var React; React.render(<zzz.Bar />);", []string{"jsxIdentifierNotDefined"}},

		// A type-only binding is not a value binding, so both of these report. Written the other way
		// round first, on the belief that anything the checker can name is enough, and the fixture
		// failed against the rule. The release oxlint binary settled it: both report there too, so
		// the rule was right and the guess was wrong. Kept as two cases because a later narrowing of
		// the resolution call to value symbols would pass the interface case and fail the alias one.
		{"a component bound only by an interface declaration", "interface App {}; var React; React.render(<App />);", []string{"jsxIdentifierNotDefined"}},
		{"a component bound only by a type alias", "type App = any; var React; React.render(<App />);", []string{"jsxIdentifierNotDefined"}},

		// A JSX element written in its non-self-closing form parses as KindJsxOpeningElement rather
		// than KindJsxSelfClosingElement, and a rule listening to only one of the two is silent on
		// the other. Upstream's corpus is entirely self-closing, so nothing in it covers this.
		{"an undeclared component written with a closing tag", "var React; React.render(<App></App>);", []string{"jsxIdentifierNotDefined"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoUndef, jsxNoUndefFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

func TestJsxNoUndefStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an uppercase component declared in the same var statement", "var React, App; React.render(<App />);"},
		{"a lowercase intrinsic element", "var React; React.render(<img />);"},
		{"a dashed custom element", "var React; React.render(<x-gif />);"},
		{"a lowercase member object that is declared", "var React, app; React.render(<app.Foo />);"},
		{"a lowercase member object declared, two levels deep", "var React, app; React.render(<app.foo.Bar />);"},
		{"a namespaced name", "var React; React.render(<Apppp:Foo />);"},
		{"a this-rooted member expression inside a class", "\n        var React;\n        class Hello extends React.Component {\n          render() {\n            return <this.props.tag />\n          }\n        }\n        "},
		{"a component imported by default import", "\n        import Text from \"cool-module\";\n        const TextWrapper = function (props) {\n          return (\n            <Text />\n          );\n        };\n        "},
		{"a var declared before an enum of the same member name", "var App; var React; enum A { App };  React.render(<App />);"},
		{"an enum before a var of the same name", "var React; enum A { App }; var App; React.render(<App />);"},
		{"an import-equals require binding", "var React; import App = require('./app'); React.render(<App />);"},
		{"an import-equals alias of an imported namespace member", "\n        var React;\n        import { Foo } from './foo';\n        import App = Foo.App;\n        React.render(<App />);\n        "},
		// Cases below are ours. Each is a shape our tree produces that upstream's corpus does not
		// reach, and every one was checked against the release oxlint binary first.

		// The intrinsic predicate's dash arm: a dash makes a custom element whatever the case of
		// the first letter, so this stays silent while `<_foo />` reports. Silent on oxlint too.
		{"a dashed name whose first letter is uppercase", "var React; React.render(<Foo-bar />);"},

		// A lowercase intrinsic with a hyphen and an uppercase segment, the web-component spelling.
		{"a dashed custom element with an uppercase tail", "var React; React.render(<my-Element />);"},

		// A bare `this` tag. typescript-go gives it KindThisKeyword rather than KindIdentifier, so
		// it never reaches the name predicate, and it must not report.
		{"a bare this tag", "var React; React.render(<this />);"},

		// A this-rooted member expression outside any class. Upstream declines every this-rooted
		// name structurally, so this is silent there too, and it is the case that would have
		// reported had the rule leaned on resolution instead: probed, `this` inside a class
		// resolves to the class and outside one resolves to nothing.
		{"a this-rooted member expression at module top level", "var React; React.render(<this.props.tag />);"},

		// A declared leftmost object under a four-deep member expression, the pass twin of the
		// four-deep fail case above. Together they pin that the walk reaches the leftmost node.
		{"a declared object under a four-deep member expression", "var React, A; React.render(<A.B.C.D />);"},

		// An ambient declaration is the shape our tree has where upstream has its `globals` option,
		// and it is what makes the block-2 pass case inexpressible rather than simply missing.
		{"a component declared in an ambient declare statement", "declare var A: any; var React; React.render(<A.B />);"},

		// A function declaration is the ordinary component spelling and appears nowhere in
		// upstream's corpus, which declares everything with `var`.
		{"a component declared as a function", "var React; function App() { return null; }; React.render(<App />);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoUndef, jsxNoUndefFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoUndefNeedsTheTypedHarness pins that this rule cannot be run through `rule_testing.Run`.
//
// The plain harness hands the rule a nil checker, and the rule's own guard then returns before
// reporting anything, which makes every silent case above pass vacuously and looks from the outside
// exactly like a working rule. A revert of `NeedsTypeChecker`, or a fixture switched to the untyped
// harness by copying a neighbouring react rule that does not need types, fails here loudly instead.
func TestJsxNoUndefNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "var React; React.render(<App />);"

	untyped := rule_testing.Run(t, JsxNoUndef, jsxNoUndefFile, source)
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTyped(t, JsxNoUndef, jsxNoUndefFile, source)
	rule_testing.ExpectFindings(t, typed, "jsxIdentifierNotDefined")
}

// TestJsxNoUndefPointsAtTheReferencedIdentifier asserts where each finding lands and what it says.
//
// `ExpectFindings` sees message ids and counts and nothing else, so a rule reporting on the whole
// element, on the member expression rather than its object, or on the wrong end of a member chain
// passes the two tables above completely. Upstream snapshots the span for all eight of its fail
// cases and underlines the object identifier alone in every one, including the columns for the
// member cases, so this is upstream's assertion rather than an invented one.
func TestJsxNoUndefPointsAtTheReferencedIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   string
	}{
		{"a bare component name", "var React; React.render(<App />);", "App"},
		{"the object of a member expression, not the whole name", "var React; React.render(<Appp.Foo />);", "Appp"},
		{"a lowercase object of a member expression", "var React; React.render(<appp.Foo />);", "appp"},
		{"the leftmost object of a two-level member expression", "var React; React.render(<appp.foo.Bar />);", "appp"},
		{"the leftmost object of a three-level member expression", "var React; React.render(<A.B.C.D />);", "A"},
		{"the tag name and not the closing tag", "var React; React.render(<App></App>);", "App"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoUndef, jsxNoUndefFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "jsxIdentifierNotDefined")

			finding := result.Diagnostics[0]
			reported := testCase.sourceText[finding.Range.Pos():finding.Range.End()]
			if reported != testCase.reported {
				t.Fatalf("finding covers %q, want %q", reported, testCase.reported)
			}
		})
	}
}

// TestJsxNoUndefAllowGlobals pins upstream's `allowGlobals` and the default it implies.
//
// ESLint's `globals` is a declaration in the global scope, so a script `.d.ts` declaring `Text`
// stands in for it. Every row was measured against the installed build (eslint-plugin-react 7.37.5,
// ESLint 10.8.1, typescript-eslint's parser) with `globals: {Text: true}`: in a module the default
// reports a tag naming only a global, a lib global included, and `allowGlobals: true` lets it
// resolve. An import, and a `declare global` member declared in the same file, both count as in
// scope under the default.
func TestJsxNoUndefAllowGlobals(t *testing.T) {
	t.Parallel()

	const globals = "/repository/source/globals.d.ts"
	allowed, err := rule.DecodeOptionsInto[JsxNoUndefOptions]()([]byte(`{"allowGlobals": true}`))
	if err != nil {
		t.Fatalf("decoding allowGlobals: %v", err)
	}
	disallowed, err := rule.DecodeOptionsInto[JsxNoUndefOptions]()([]byte(`{"allowGlobals": false}`))
	if err != nil {
		t.Fatalf("decoding allowGlobals false: %v", err)
	}

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantCount  int
	}{
		{"a global under the default", "export {}; var React: any; React.render(<Text />);", nil, 1},
		{"a global under allowGlobals false", "export {}; var React: any; React.render(<Text />);", disallowed, 1},
		{"a global under allowGlobals", "export {}; var React: any; React.render(<Text />);", allowed, 0},
		{"a lib global under the default", "export {}; <Map />;", nil, 1},
		{"a lib global under allowGlobals", "export {}; <Map />;", allowed, 0},
		{"an import under allowGlobals false", "import Text from './globals'; <Text />;", disallowed, 0},
		{"a declare global member in the same file", "declare global { var G: any } export {}; <G />;", nil, 0},
		{"an undeclared name under allowGlobals", "export {}; <Missing />;", allowed, 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithOptions(t, JsxNoUndef, map[string]string{
				globals:        "declare var Text: any;\n",
				jsxNoUndefFile: testCase.sourceText,
			}, jsxNoUndefFile, testCase.options)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Errorf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

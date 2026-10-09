package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noRedundantShouldComponentUpdateFile is where the fixtures pretend to live.
//
// A .tsx extension because the neighbours use one, and NO file-suffix gate exists in this rule.
// Three shipped react rules here gate on .tsx or .jsx, which is oxc-era residue rather than
// upstream behavior, and TestNoRedundantShouldComponentUpdateHasNoFileSuffixGate below pins the
// absence so nobody reintroduces it by pattern-matching a neighbour. This rule never touches JSX,
// so every suffix is exercised there.
const noRedundantShouldComponentUpdateFile = "/repository/source/Pure.tsx"

// The corpus is eslint-plugin-react's own, imported verbatim from
// tests/lib/rules/no-redundant-should-component-update.js by evaluating the upstream tester with a
// stub RuleTester and serializing what it was handed, so no case was retyped and no escape could be
// cooked on the way in. 3 valid and 6 invalid, and every invalid case names exactly one
// entry in its errors array, so one finding per input is stated by upstream rather than recovered.
//
// Every invalid case also carries the class name it expects in the message data, so the expected
// component name is upstream's own rather than mine. That third column is what pins the name
// derivation, which no message-id assertion can see.
//
// The three valid cases are all the same shape with Component in place of PureComponent, which is
// exactly the distinction the shelf's IsEs6ComponentClass would erase.
func TestNoRedundantShouldComponentUpdateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText    string
		findings      []string
		componentName string
	}{
		{"\n        class Foo extends React.PureComponent {\n          shouldComponentUpdate() {\n            return true;\n          }\n        }\n      ", []string{"noShouldCompUpdate"}, "Foo"},
		{"\n        class Foo extends PureComponent {\n          shouldComponentUpdate() {\n            return true;\n          }\n        }\n      ", []string{"noShouldCompUpdate"}, "Foo"},
		{"\n        class Foo extends React.PureComponent {\n          shouldComponentUpdate = () => {\n            return true;\n          }\n        }\n      ", []string{"noShouldCompUpdate"}, "Foo"},
		{"\n        function Foo() {\n          return class Bar extends React.PureComponent {\n            shouldComponentUpdate() {\n              return true;\n            }\n          };\n        }\n      ", []string{"noShouldCompUpdate"}, "Bar"},
		{"\n        function Foo() {\n          return class Bar extends PureComponent {\n            shouldComponentUpdate() {\n              return true;\n            }\n          };\n        }\n      ", []string{"noShouldCompUpdate"}, "Bar"},
		{"\n        var Foo = class extends PureComponent {\n          shouldComponentUpdate() {\n            return true;\n          }\n        }\n      ", []string{"noShouldCompUpdate"}, "Foo"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
			// The name is interpolated into the message, so an id assertion cannot see it and a
			// wrong derivation would ship green. Asserted as a prefix of the rendered text against
			// a literal built here rather than against the rule's own format string, since a
			// comparison to the rule's own constant moves with it under mutation.
			wantPrefix := testCase.componentName + " extends PureComponent and also writes shouldComponentUpdate."
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, wantPrefix) {
				t.Fatalf("message was %q, wanted it to start %q", result.Diagnostics[0].Message.Description, wantPrefix)
			}
		})
	}
}

func TestNoRedundantShouldComponentUpdateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\n        class Foo extends React.Component {\n          shouldComponentUpdate() {\n            return true;\n          }\n        }\n      ",
		"\n        class Foo extends React.Component {\n          shouldComponentUpdate = () => {\n            return true;\n          }\n        }\n      ",
		"\n        function Foo() {\n          return class Bar extends React.Component {\n            shouldComponentUpdate() {\n              return true;\n            }\n          };\n        }\n      ",
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Beyond the imported corpus. The corpus is nine cases over three shapes, so almost every
// discrimination this rule makes is invisible to it. Each verdict below was measured against the
// installed eslint-plugin-react through the Linter API before being written here.

// TestNoRedundantShouldComponentUpdateBaseClassBoundary pins WHICH heritage counts.
//
// This is the test that separates this rule from react.IsEs6ComponentClass, which accepts Component
// and PureComponent alike. Routing through the shelf would report upstream's first valid case, and
// the three imported valid cases are all one shape, so they cannot show the boundary on their own.
func TestNoRedundantShouldComponentUpdateBaseClassBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"the namespaced spelling", "class Foo extends React.PureComponent { shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"the bare spelling a direct import writes", "class Foo extends PureComponent { shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"a parenthesized heritage, which upstream reports because its tree has no parenthesis node", "class Foo extends (React.PureComponent) { shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"Component is the whole point of the boundary and is silent", "class Foo extends React.Component { shouldComponentUpdate() { return true; } }", nil},
		// The BARE Component spelling, which neither upstream's corpus nor the namespaced case
		// above covers. Added for a surviving mutant that broke the identifier comparison into an
		// always-true test: every fixture stayed green, because the only bare identifier any of
		// them writes in a heritage IS PureComponent, so a rule accepting every bare identifier
		// agreed with all of them. Both lines measured silent on the installed build.
		{"the bare Component spelling a direct import writes is silent", "class Foo extends Component { shouldComponentUpdate() { return true; } }", nil},
		{"and so is an unrelated base class", "class Foo extends Whatever { shouldComponentUpdate() { return true; } }", nil},
		{"another package's PureComponent is silent", "class Foo extends Preact.PureComponent { shouldComponentUpdate() { return true; } }", nil},
		{"and so is any other object", "class Foo extends Foo.PureComponent { shouldComponentUpdate() { return true; } }", nil},
		{"PureComponent without the method is silent", "class Foo extends PureComponent {}", nil},
		{"the method without PureComponent is silent", "class Foo { shouldComponentUpdate() { return true; } }", nil},
		{"an implements clause is not an extends clause", "interface P {} class Foo extends React.Component implements P { shouldComponentUpdate() { return true; } }", nil},
		// TypeScript-only shapes, which upstream cannot express at all: its default parser rejects
		// `implements` outright, and its rule reads `node.superClass`, which an implements clause
		// never populates. So silence is upstream's answer structurally rather than by measurement,
		// and that is stated rather than claimed as a probe result.
		//
		// These exist for a surviving mutant, and the first one is a REAL false positive the token
		// guard prevents. With that guard neutralized to a constant, `implements (PureComponent)`
		// REPORTS: the parser hands a malformed implements clause a KindExpressionWithTypeArguments
		// whose expression is a parenthesized identifier, and the parenthesis skip in
		// isPureComponentBase then unwraps it to a bare `PureComponent`.
		//
		// The plain `implements PureComponent` spelling is clean either way, because a well-formed
		// implements clause carries a KindTypeReference that the type-kind guard declines. So the
		// two guards cover different halves and neither subsumes the other, which is why the first
		// mutation of this line read as a survivor: rewriting `!= Extends` into `== Implements`
		// skips the same clauses and is inert. Only neutralizing it to a constant separates them.
		{"a well-formed implements clause carries a different type kind", "class Foo implements PureComponent { shouldComponentUpdate() { return true; } }", nil},
		{"and a malformed parenthesized one is what the extends-token guard actually prevents", "class Foo implements (PureComponent) { shouldComponentUpdate() { return true; } }", nil},
		{"nor does a doubly parenthesized one", "class Foo implements ((PureComponent)) { shouldComponentUpdate() { return true; } }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoRedundantShouldComponentUpdateMemberShapes pins WHICH member spellings count.
//
// Upstream applies no filter to the class body at all, so static, accessor, field and private
// spellings all report, while a computed key and a string key do not. The corpus writes exactly two
// of these ten shapes, so most of the distinctions here are invisible to it.
//
// The two silent lines are not a special case and would be easy to "fix" into a defect: upstream's
// getPropertyName reads nameNode.name, a field only an Identifier and a PrivateIdentifier carry, so
// a string key and a computed key both yield the empty string mechanically.
func TestNoRedundantShouldComponentUpdateMemberShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"a method", "class Foo extends PureComponent { shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"a class field holding an arrow", "class Foo extends PureComponent { shouldComponentUpdate = () => true; }", []string{"noShouldCompUpdate"}},
		{"a bare field with no initializer", "class Foo extends PureComponent { shouldComponentUpdate; }", []string{"noShouldCompUpdate"}},
		{"static is not excluded", "class Foo extends PureComponent { static shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"nor is a getter", "class Foo extends PureComponent { get shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"nor a private name, which React never calls and upstream reports anyway", "class Foo extends PureComponent { #shouldComponentUpdate() { return true; } }", []string{"noShouldCompUpdate"}},
		{"a computed key is silent, because it has no name to read", "class Foo extends PureComponent { ['shouldComponentUpdate']() { return true; } }", nil},
		{"and a string key is silent for the same reason", "class Foo extends PureComponent { 'shouldComponentUpdate'() { return true; } }", nil},
		{"a differently named method is silent", "class Foo extends PureComponent { componentDidUpdate() { return true; } }", nil},
		{"the method written twice reports once, because upstream stops at the first match", "class Foo extends PureComponent { shouldComponentUpdate() {} shouldComponentUpdate() {} }", []string{"noShouldCompUpdate"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

// TestNoRedundantShouldComponentUpdateNameDerivation pins the interpolated class name.
//
// ExpectFindings cannot see a single character of this, so a rule deriving the name from the wrong
// node would pass every fixture above. Upstream reads node.id, then node.parent.id, then the empty
// string, and the empty string is real output rather than a failure: three of these six shapes
// render a message beginning with a space.
//
// The message is compared against a literal built here rather than against the rule's own format
// string, because a comparison to the rule's own constant moves with it under mutation.
func TestNoRedundantShouldComponentUpdateNameDerivation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantName   string
	}{
		{"a declaration uses its own name", "class Foo extends PureComponent { shouldComponentUpdate() {} }", "Foo"},
		{"a named class expression uses its own name over the variable's", "const Foo = class Bar extends PureComponent { shouldComponentUpdate() {} };", "Bar"},
		{"an anonymous class expression borrows the declarator's name", "var Foo = class extends PureComponent { shouldComponentUpdate() {} };", "Foo"},
		{"an anonymous class in a call argument has no name at all", "declare function foo(x: unknown): void; foo(class extends PureComponent { shouldComponentUpdate() {} });", ""},
		{"nor does one in an object property, which is not a declarator", "const o = { Foo: class extends PureComponent { shouldComponentUpdate() {} } };", ""},
		{"nor does one on the right of an assignment", "let a: unknown; a = class extends PureComponent { shouldComponentUpdate() {} };", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noShouldCompUpdate")
			wantPrefix := testCase.wantName + " extends PureComponent and also writes shouldComponentUpdate."
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, wantPrefix) {
				t.Fatalf("message was %q, wanted it to start %q", result.Diagnostics[0].Message.Description, wantPrefix)
			}
		})
	}
}

// TestNoRedundantShouldComponentUpdateSpan asserts WHERE the finding points.
//
// Upstream reports on the class node itself rather than on the offending member, measured: a class
// expression assigned to a variable spans from `class` to the closing brace and excludes the
// declaration prefix. A rule anchoring on the method would pass every fixture above while pointing
// somewhere the reader was never shown.
func TestNoRedundantShouldComponentUpdateSpan(t *testing.T) {
	t.Parallel()

	t.Run("a declaration reports on the whole class", func(t *testing.T) {
		t.Parallel()
		const sourceText = "class Foo extends PureComponent { shouldComponentUpdate() {} }"
		result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, sourceText)
		rule_testing.ExpectFindings(t, result, "noShouldCompUpdate")
		reported := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != sourceText {
			t.Fatalf("reported on %q, wanted the whole class", reported)
		}
	})

	t.Run("a class expression reports on the expression, not the declaration", func(t *testing.T) {
		t.Parallel()
		const sourceText = "var Foo = class extends PureComponent { shouldComponentUpdate() {} };"
		const want = "class extends PureComponent { shouldComponentUpdate() {} }"
		result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, sourceText)
		rule_testing.ExpectFindings(t, result, "noShouldCompUpdate")
		reported := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != want {
			t.Fatalf("reported on %q, wanted %q", reported, want)
		}
	})
}

// TestNoRedundantShouldComponentUpdateHasNoFileSuffixGate pins the ABSENCE of a suffix gate.
//
// Three shipped react rules in this package gate on .tsx or .jsx, which is residue from an oxc-era
// port rather than upstream behavior, and it blinds them to most of this tree. Measured against
// eslint-plugin-react with one source under four suffixes: all four report.
//
// This rule never touches JSX, so all four are reachable here, unlike no_namespace.go where the
// typed harness cannot build a JavaScript program.
func TestNoRedundantShouldComponentUpdateHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const sourceText = "class Foo extends PureComponent { shouldComponentUpdate() { return true; } }"
	for _, fileName := range []string{
		"/repository/source/Pure.ts",
		"/repository/source/Pure.tsx",
		"/repository/source/Pure.js",
		"/repository/source/Pure.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, fileName, sourceText)
			rule_testing.ExpectFindings(t, result, "noShouldCompUpdate")
		})
	}
}

// TestNoRedundantShouldComponentUpdateDoesNotPanicOnComputedMembers is crash protection, not a
// behavioral fixture.
//
// ast.Node.Text has no case for a ComputedPropertyName and reaches its unhandled-case panic,
// measured directly with a probe. The walk recovers per FILE rather than per rule, so a single
// computed member inside a PureComponent would cost every rule in this package every finding in that
// file, and the run would still print a plausible summary line.
//
// No ExpectFindings fixture can see a panic, so the sweep reports SURVIVED identically whether the
// kind guard in identifierMemberName matters or not. This test exists to name what that guard
// prevents rather than what it decides.
//
// The computed keys here are deliberately mixed with a real reporting member in the second case, so
// the test also shows the walk reaching past the computed member rather than merely surviving it.
func TestNoRedundantShouldComponentUpdateDoesNotPanicOnComputedMembers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   []string
	}{
		{"a computed member alone", "class Foo extends PureComponent { ['anything']() {} }", nil},
		{"a computed member beside a reporting one", "class Foo extends PureComponent { ['anything']() {} shouldComponentUpdate() {} }", []string{"noShouldCompUpdate"}},
		{"a computed field", "const k = 'a'; class Foo extends PureComponent { [k] = 1; }", nil},
		{"a string key beside a reporting one", "class Foo extends PureComponent { 'anything'() {} shouldComponentUpdate() {} }", []string{"noShouldCompUpdate"}},
		{"a numeric key", "class Foo extends PureComponent { 1() {} }", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRedundantShouldComponentUpdate, noRedundantShouldComponentUpdateFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		})
	}
}

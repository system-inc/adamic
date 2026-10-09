package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// forbidForeignPropTypesFile is where the fixtures pretend to live.
//
// A .tsx extension because several cases hold JSX. The rule has no suffix gate, which a case below
// pins by writing reporting source to a plain `.ts` file: most of what this rule catches needs no
// JSX at all.
const forbidForeignPropTypesFile = "/repository/source/ForbidForeignPropTypes.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/forbid-foreign-prop-types.js, read by
// evaluating the two arrays in the tester with a stubbed RuleTester and emitting Go raw strings from
// the decoded values, so no escape sequence was typed on the way here. Upstream carries 12 valid and
// 8 invalid cases, each invalid one naming exactly one finding.
//
// All 20 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and the corpus and the running rule agreed on every one.
// The replay needed `ecmaVersion: 2022` rather than the 2018 the tester declares, because three
// cases use class fields; at 2018 they come back as parse errors rather than as findings, which
// reads exactly like a rule disagreement and is not one.
//
// Two invalid cases carry upstream's `no-ts` marker, which says its own TypeScript parsers fail
// them. Ours parses both, and both are imported as reporting, which is upstream's verdict under the
// parser that can read them. Their marker is a TODO in upstream's file rather than a statement
// about the rule.

// forbidForeignPropTypesOptions decodes a raw option body the way the config layer does.
//
// Fixtures route through the registered decoder rather than building the struct, so the decode path
// is under test rather than assumed. An empty body is the nil-options case: the decoder errors, the
// config turns that into nil for a non-required rule, and the rule's type assertion yields the zero
// struct.
func forbidForeignPropTypesOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	decoded, err := rule.DecodeOptionsInto[ForbidForeignPropTypesOptions]()([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestForbidForeignPropTypesFires runs the eight failing cases from upstream.
func TestForbidForeignPropTypesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"invalid 0 a plain member read inside a factory", `
        var Foo = createReactClass({
          propTypes: Bar.propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `, ""},
		{"invalid 1 a computed string member read inside a factory", `
        var Foo = createReactClass({
          propTypes: Bar["propTypes"],
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `, ""},
		{"invalid 2 a shorthand destructuring", `
        var { propTypes } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `, ""},
		{"invalid 3 a renamed destructuring beside a rest", `
        var { propTypes: things, ...foo } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `, ""},
		// Upstream marks this `no-ts`, meaning its TypeScript parsers fail it. Ours parses it and
		// reports, which is upstream's own verdict under a parser that can read the class field.
		{"invalid 4 a class field whose key is not propTypes", `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `, ""},
		{"invalid 5 a renamed destructuring used later", `
        var { propTypes: typesOfProps } = SomeComponent
        var Foo = createReactClass({
          propTypes: typesOfProps,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `, ""},
		{"invalid 6 an assignment into propTypes with the option off", `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `, `{"allowInPropTypes":false}`},
		{"invalid 7 a propTypes class field with the option off", `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `, `{"allowInPropTypes":false}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(
				t,
				ForbidForeignPropTypes,
				forbidForeignPropTypesFile,
				testCase.sourceText,
				forbidForeignPropTypesOptions(t, testCase.rawOptions),
			)
			rule_testing.ExpectFindings(t, result, "forbiddenPropType")
		})
	}
}

// TestForbidForeignPropTypesStaysSilent runs the twelve passing cases from upstream.
func TestForbidForeignPropTypesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		// The import cases matter: the rule never asks what a name means, only what shape reads it,
		// and an import specifier is neither a member access nor an object pattern.
		{"valid 0 an import of a name that happens to be propTypes", `import { propTypes } from "SomeComponent";`, ""},
		{"valid 1 an aliased import", `import { propTypes as someComponentPropTypes } from "SomeComponent";`, ""},
		{"valid 2 a bare identifier", `const foo = propTypes`, ""},
		{"valid 3 a bare identifier as an argument", `foo(propTypes)`, ""},
		{"valid 4 a bare identifier in an operand", `foo + propTypes`, ""},
		{"valid 5 a bare identifier in an array", `const foo = [propTypes]`, ""},
		// An object LITERAL is not an object PATTERN, and only the pattern arm exists.
		{"valid 6 a shorthand in an object literal", `const foo = { propTypes }`, ""},
		{"valid 7 a plain member as an assignment target", `Foo.propTypes = propTypes`, ""},
		{"valid 8 a computed member as an assignment target", `Foo["propTypes"] = propTypes`, ""},
		// A computed identifier has no literal text to compare, whatever it holds at runtime.
		{"valid 9 a computed member through a variable", `const propTypes = "bar"; Foo[propTypes];`, ""},
		{"valid 10 an assignment into propTypes with the option on", `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `, `{"allowInPropTypes":true}`},
		{"valid 11 a propTypes class field with the option on", `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `, `{"allowInPropTypes":true}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(
				t,
				ForbidForeignPropTypes,
				forbidForeignPropTypesFile,
				testCase.sourceText,
				forbidForeignPropTypesOptions(t, testCase.rawOptions),
			)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestForbidForeignPropTypesHasNoFileSuffixGate pins that the rule reads a plain `.ts` file.
//
// Three siblings in this package gate on `.tsx`/`.jsx`, inherited from oxc, which costs them every
// finding in a `.ts` file. This rule is ported from the authority, which has no such gate, and the
// gate would matter more here than on most: `const { propTypes } = SomeComponent` involves no JSX,
// so a suffix gate would blind the rule on the shape it most needs to catch.
//
// `.jsx` and `.js` are not covered, and that is a fact about the harness rather than the rule: the
// typed program's tsconfig includes only TypeScript extensions. This rule needs no checker, so the
// untyped harness runs it under any name, but the two suffixes that decide the question are the two
// here.
func TestForbidForeignPropTypesHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{".tsx", ".ts", ".jsx", ".js"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(
				t,
				ForbidForeignPropTypes,
				"/repository/source/SuffixProbe"+suffix,
				`const x = Foo.propTypes;`,
			)
			rule_testing.ExpectFindings(t, result, "forbiddenPropType")
		})
	}
}

// TestForbidForeignPropTypesAnchorsOnTheProperty asserts where each finding points.
//
// The corpus asserts message ids only, so nothing in it can see a rule reporting the right judgment
// in the wrong place. Upstream anchors on the PROPERTY node for a member access and on the whole
// matched property for a destructuring, which are two different choices in one rule, and both were
// read from upstream's reported columns rather than from its source.
func TestForbidForeignPropTypesAnchorsOnTheProperty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a plain member points at the property alone", `const x = Foo.propTypes;`, `propTypes`},
		// The quotes are inside the span, which is what separates anchoring on the literal from
		// anchoring on its cooked text.
		{"a computed member points at the literal with its quotes", `const x = Foo["propTypes"];`, `"propTypes"`},
		{"a shorthand pattern points at the binding", `const { propTypes } = Foo;`, `propTypes`},
		// The renamed form is the case that separates the two anchors: the finding covers both
		// halves rather than just the key.
		{"a renamed pattern points at both halves", `const { propTypes: p } = Foo;`, `propTypes: p`},
		{"a chained member points at the last property", `const x = Foo.Bar.propTypes;`, `propTypes`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// The untyped harness does not trim, so the literal and the bytes on disk agree. The
			// transform is applied anyway so the slice cannot drift if that ever changes.
			onDisk := testCase.sourceText
			result := rule_testing.Run(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "forbiddenPropType")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			got := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if got != testCase.wantText {
				t.Errorf("finding spans %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestForbidForeignPropTypesExemptsOnlyAssignmentTargets covers upstream's `isAssignmentLHS`.
//
// The corpus writes two clean assignment targets and nothing else in this family, so the boundary
// of the exemption is entirely unasserted by it. Every case here was measured against the installed
// build, and two of them are upstream reporting on code that plainly writes rather than reads.
func TestForbidForeignPropTypesExemptsOnlyAssignmentTargets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain assignment target is exempt", `Foo.propTypes = {};`, nil},
		// A compound assignment is still one AssignmentExpression to upstream's parser, so it is
		// exempt for the same reason. `ast.IsAssignmentExpression(node, false)` is what reproduces
		// that; passing true would report here and diverge.
		{"a compound assignment target is exempt", `Foo.propTypes += 1;`, nil},
		// An update expression is NOT an AssignmentExpression, so upstream reports on a write.
		// Reproduced rather than corrected.
		{"an update expression is not exempt", `Foo.propTypes++;`, []string{"forbiddenPropType"}},
		{"a delete is not exempt", `delete Foo.propTypes;`, []string{"forbiddenPropType"}},
		{"the right side of an assignment reports", `Foo.propTypes = Bar.propTypes;`, []string{"forbiddenPropType"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidForeignPropTypesUnwrapsParentheses covers the parser difference, in the direction that
// costs a false positive.
//
// `(Foo.propTypes) = {}` is CLEAN upstream, measured. Its parser folds the parenthesis away so the
// member access IS the assignment's left. Ours keeps a `KindParenthesizedExpression` between them,
// so a naive parent test sees a parenthesis instead of an assignment, concludes the access is not a
// target, and reports where upstream is silent. No imported fixture can express this: upstream's
// corpus cannot write the shape, because its parser deleted the node before any test existed.
func TestForbidForeignPropTypesUnwrapsParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a parenthesized assignment target is exempt", `(Foo.propTypes) = {};`, nil},
		// `((x))` nests, which is why the unwrap is a loop rather than a single step.
		{"a doubly parenthesized target is exempt", `((Foo.propTypes)) = {};`, nil},
		{"a parenthesized receiver still reports", `const x = (Foo).propTypes;`, []string{"forbiddenPropType"}},
		{"a parenthesized read still reports", `const x = (Foo.propTypes);`, []string{"forbiddenPropType"}},
		{"a doubly parenthesized receiver still reports", `const x = ((Foo)).propTypes;`, []string{"forbiddenPropType"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidForeignPropTypesReadsOnlyLiteralKeys covers what counts as the property name.
//
// The corpus carries the reporting computed-string case and the declining computed-identifier one,
// and nothing else in this family. The rest were measured against the installed build.
func TestForbidForeignPropTypesReadsOnlyLiteralKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a template literal key declines", "const x = Foo[`propTypes`];", nil},
		{"optional chaining still reports", `const x = Foo?.propTypes;`, []string{"forbiddenPropType"}},
		// The receiver is never examined, which is what makes this rule as broad as it is.
		{"the receiver is not examined", `const x = anything.propTypes;`, []string{"forbiddenPropType"}},
		{"reading through propTypes reports once", `const x = Foo.propTypes.bar;`, []string{"forbiddenPropType"}},
		{"an array pattern declines", `const [ propTypes ] = Foo;`, nil},
		{"a computed pattern key declines", `const { ["propTypes"]: p } = Foo;`, nil},
		{"a string literal pattern key declines", `const { "propTypes": p } = Foo;`, nil},
		{"a pattern without the key declines", `const { other } = Foo;`, nil},
		{"a rest only pattern declines", `const { ...rest } = Foo;`, nil},
		// Upstream uses `properties.find`, so a pattern naming the key twice reports ONCE. A loop
		// that did not stop would report twice, and nothing in the corpus writes the shape.
		{"a pattern naming it twice reports once", `const { propTypes, propTypes: q } = Foo;`, []string{"forbiddenPropType"}},
		{"any object pattern counts, including a parameter", `function f({ propTypes }) {}`, []string{"forbiddenPropType"}},
		{"a nested pattern counts", `const { a: { propTypes } } = Foo;`, []string{"forbiddenPropType"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidForeignPropTypesAllowInPropTypesReachesUpward covers the option's two walks.
//
// Both are unbounded up to the file, which is what lets the exemption reach through nesting and
// through a closure. The corpus writes one case per walk at one level of nesting, so the depth, the
// negative case, and the two key-shape declines are all measured rather than imported.
func TestForbidForeignPropTypesAllowInPropTypesReachesUpward(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"an assignment exempts through two levels of nesting", `Foo.propTypes = { a: { b: Bar.propTypes.c } };`, nil},
		{"a class field exempts through a closure", `class C { static propTypes = { a: () => Q.propTypes.c }; }`, nil},
		// The assignment walk tests the LEFT's property name, so assigning into anything else does
		// not exempt. This is the case that makes the walk a discrimination rather than a blanket.
		{"assigning into another property does not exempt", `Foo.other = Bar.propTypes;`, []string{"forbiddenPropType"}},
		// Upstream does not test `static`, measured, so an instance field exempts too.
		{"a non static class field exempts", `class C { propTypes = { a: Q.propTypes.b }; }`, nil},
		// Upstream reads `classProperty.key.name`, which a computed key does not have.
		{"a computed class key does not exempt", `class C { static ["propTypes"] = { a: Q.propTypes.b }; }`, []string{"forbiddenPropType"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(
				t,
				ForbidForeignPropTypes,
				forbidForeignPropTypesFile,
				testCase.sourceText,
				forbidForeignPropTypesOptions(t, `{"allowInPropTypes":true}`),
			)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidForeignPropTypesDoesNotApplyTheOptionToPatterns pins that the destructuring arm ignores
// the option.
//
// Upstream's `ObjectPattern` handler calls neither `isAllowedAssignment` nor anything like it, which
// is visible in the source rather than measured. So a destructuring inside a `propTypes` assignment
// still reports even with the option on, and a port that threaded the option through both arms
// would silence a case upstream reports.
func TestForbidForeignPropTypesDoesNotApplyTheOptionToPatterns(t *testing.T) {
	t.Parallel()

	source := `Foo.propTypes = (function () { const { propTypes } = Bar; return propTypes; })();`
	result := rule_testing.RunWithOptions(
		t,
		ForbidForeignPropTypes,
		forbidForeignPropTypesFile,
		source,
		forbidForeignPropTypesOptions(t, `{"allowInPropTypes":true}`),
	)
	rule_testing.ExpectFindings(t, result, "forbiddenPropType")
}

// TestForbidForeignPropTypesDefaultsWithoutADecoder pins the nil-options path.
//
// A rule configured as a bare severity is handed nil rather than a decoded struct, so the rule's
// type assertion yields the zero value. For this option surface the zero value is upstream's
// default of false, and this asserts that directly rather than through the decoder, which is the
// one path a fixture routed through the decoder cannot reach.
func TestForbidForeignPropTypesDefaultsWithoutADecoder(t *testing.T) {
	t.Parallel()

	source := `Foo.propTypes = { a: Bar.propTypes.b };`

	withoutOptions := rule_testing.RunWithOptions(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, source, nil)
	rule_testing.ExpectFindings(t, withoutOptions, "forbiddenPropType")

	withExplicitFalse := rule_testing.RunWithOptions(
		t,
		ForbidForeignPropTypes,
		forbidForeignPropTypesFile,
		source,
		forbidForeignPropTypesOptions(t, `{"allowInPropTypes":false}`),
	)
	rule_testing.ExpectFindings(t, withExplicitFalse, "forbiddenPropType")

	// The option is what moves the verdict, which is what makes the two above a measurement rather
	// than a coincidence.
	withTrue := rule_testing.RunWithOptions(
		t,
		ForbidForeignPropTypes,
		forbidForeignPropTypesFile,
		source,
		forbidForeignPropTypesOptions(t, `{"allowInPropTypes":true}`),
	)
	rule_testing.ExpectClean(t, withTrue)
}

// TestForbidForeignPropTypesMessageNamesTheProductionFailure asserts the rendered text.
//
// The rule builds no message with a format verb, so there is nothing to interpolate and nothing to
// render wrong. What is still worth pinning is that the description explains the production-build
// mechanism rather than restating the rule name, because that is the sentence a reader acts on.
// Asserted against a literal typed here rather than against the rule's own constant, so both sides
// cannot move together under mutation.
func TestForbidForeignPropTypesMessageNamesTheProductionFailure(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ForbidForeignPropTypes, forbidForeignPropTypesFile, `const x = Foo.propTypes;`)
	rule_testing.ExpectFindings(t, result, "forbiddenPropType")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	message := result.Diagnostics[0].Message
	if message.Id != "forbiddenPropType" {
		t.Errorf("message id is %q", message.Id)
	}
	if !strings.HasPrefix(message.Description, "Reading another component's `propTypes` breaks in production.") {
		t.Errorf("description does not open on the production failure: %q", message.Description)
	}
}

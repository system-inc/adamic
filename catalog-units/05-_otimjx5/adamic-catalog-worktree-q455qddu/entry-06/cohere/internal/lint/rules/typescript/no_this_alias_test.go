package typescript

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const thisAliasFile = "/repository/source/Thing.ts"

// decodeThisAliasOptionsForTest runs a fixture's configuration through the rule's own decoder.
//
// Fixtures configure the rule the way the config layer does, in upstream's JSON spelling, rather
// than by building the internal struct directly. That is deliberate: the decoder is where the
// `allowDestructuring` inversion and the `allowNames` alias live, so a fixture that skipped it
// would test the rule while leaving the two most breakable lines in the port unasserted.
func decodeThisAliasOptionsForTest(t *testing.T, configuration string) any {
	t.Helper()
	options, err := DecodeNoThisAliasOptions(json.RawMessage(configuration))
	if err != nil {
		t.Fatalf("decoding %s: %v", configuration, err)
	}
	return options
}

// TestNoThisAliasFires is the corpus's ten failing inputs, copied byte for byte.
//
// The snapshot carries sixteen diagnostics against these ten inputs, and the extractor flags that
// as a discrepancy. It is not one: nine inputs report once and the class case reports seven, which
// the snapshot's own line headers confirm (lines 3, 4, 12, 13, 14, 15 and 16 of that input).
func TestNoThisAliasFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		sourceText    string
		configuration string
		expected      []string
	}{
		{"an identifier alias with no configuration", "const self = this;", "{}", []string{"thisAssignment"}},
		{"an object pattern with destructuring disallowed", "const { props, state } = this;", `{"allowDestructuring": false}`, []string{"thisDestructure"}},
		{"an array pattern with destructuring disallowed", "const [ props, state ] = this;", `{"allowDestructuring": false}`, []string{"thisDestructure"}},
		{"an assignment to an existing binding", "let foo; \nconst other =3;\n\n\n\nfoo = this", "{}", []string{"thisAssignment"}},
		{"an assignment through a type assertion", "let foo; (foo as any) = this", "{}", []string{"thisAssignment"}},
		{"an alias inside a function", "function testFunction() {\n            let inFunction = this;\n          }", "{}", []string{"thisAssignment"}},
		{"an alias inside an arrow function", "const testLambda = () => {\n            const inLambda = this;\n          };", "{}", []string{"thisAssignment"}},
		// The corpus's one multi-finding input, and the only place the two messages appear
		// together. Two identifier aliases in the constructor, one in the method, then four
		// destructuring patterns, in source order.
		{
			"a class with every shape at once",
			"class TestClass {\n            constructor() {\n              const inConstructor = this;\n              const asThis: this = this;\n\n              const asString = 'this';\n              const asArray = [this];\n              const asArrayString = ['this'];\n            }\n\n            public act(scope: this = this) {\n              const inMemberFunction = this;\n              const { act1 } = this;\n              const { act2, constructor } = this;\n              const [foo1] = this;\n              const [foo, bar] = this;\n            }\n          }",
			`{"allowDestructuring": false}`,
			[]string{
				"thisAssignment", "thisAssignment", "thisAssignment",
				"thisDestructure", "thisDestructure", "thisDestructure", "thisDestructure",
			},
		},
		{"an allow list naming a different name", "const self = this;", `{"allowedNames": ["bar"]}`, []string{"thisAssignment"}},
		{"the typo'd allow list naming a different name", "const self = this;", `{"allowNames": ["bar"]}`, []string{"thisAssignment"}},

		// Cases upstream does not write, each measured on the release binary before being recorded.
		//
		// The assignment arm never reads the operator, so a compound assignment reports even when
		// it is not aliasing at all. `+=` on a receiver produces a string, so this finding is
		// arguably wrong, and it is reproduced rather than improved because the gate runs upstream.
		{"a compound assignment", "let foo; foo += this", "{}", []string{"thisAssignment"}},
		{"a nullish assignment", "let foo; foo ??= this", "{}", []string{"thisAssignment"}},
		// Parentheses on the assignment target unwrap recursively, and the finding points at the
		// bare identifier rather than at the parenthesized expression. The span assertions below
		// are what pin the second half of that.
		{"a parenthesized assignment target", "let foo; (foo) = this", "{}", []string{"thisAssignment"}},
		{"a doubly parenthesized assignment target", "let foo; ((foo)) = this", "{}", []string{"thisAssignment"}},
		{"an assignment through satisfies", "let foo; (foo satisfies any) = this", "{}", []string{"thisAssignment"}},
		{"an assignment through a non-null assertion", "let foo; (foo!) = this", "{}", []string{"thisAssignment"}},
		// The destructuring assignment targets, which parse as literals rather than as patterns and
		// are therefore the arm most likely to be missed. Upstream reports both under the flag.
		{"an array destructuring assignment", "let foo; [foo] = this", `{"allowDestructuring": false}`, []string{"thisDestructure"}},
		{"an object destructuring assignment", "let foo; ({foo} = this)", `{"allowDestructuring": false}`, []string{"thisDestructure"}},
		// A declaration with two declarators produces two findings, which is the fixture that would
		// catch a listener anchored on the declaration list rather than on each declarator.
		{"two declarators in one declaration", "var b = this, c = this;", "{}", []string{"thisAssignment", "thisAssignment"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoThisAlias, thisAliasFile, testCase.sourceText,
				decodeThisAliasOptionsForTest(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.expected...)
		})
	}
}

// TestNoThisAliasStaysSilent is the corpus's seven passing inputs plus the cases our own tree makes
// reachable.
func TestNoThisAliasStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		sourceText    string
		configuration string
	}{
		{"an object pattern with destructuring allowed", "const { props, state } = this;", `{"allowDestructuring": true}`},
		{"a single-member object pattern", "const { length } = this;", `{"allowDestructuring": true}`},
		{"a two-member object pattern", "const { length, toString } = this;", `{"allowDestructuring": true}`},
		{"a single-element array pattern", "const [foo] = this;", `{"allowDestructuring": true}`},
		{"a two-element array pattern", "const [foo, bar] = this;", `{"allowDestructuring": true}`},
		{"an allow list naming the alias", "const self = this;", `{"allowedNames": ["self"]}`},
		{"the typo'd allow list naming the alias", "const self = this;", `{"allowNames": ["self"]}`},

		// Cases upstream does not write.
		//
		// The five below are the initializer's narrowness, which is the half of the parenthesis
		// question that has to be got right by NOT skipping. Each is silent on the release binary
		// and each would report under a `SkipParentheses` on the right-hand side.
		{"a parenthesized initializer", "const self = (this);", "{}"},
		{"an initializer through a type assertion", "const self = this as any;", "{}"},
		{"an initializer through a non-null assertion", "const self = this!;", "{}"},
		{"an initializer through satisfies", "const self = this satisfies any;", "{}"},
		{"an initializer inside a comma expression", "const self = (0, this);", "{}"},
		// A member target yields no identifier reference, so upstream declines it, including the
		// most common hand-written spelling of this exact defect.
		{"an assignment to a property of this", "this.self = this;", "{}"},
		{"an assignment to a nested property", "declare let m: any;\nm.n.o = this;", "{}"},
		{"an assignment to an element access", "declare let m: any;\nm['n'] = this;", "{}"},
		// Destructuring is allowed by default, so an unconfigured rule is silent on every pattern.
		// This is the fixture that fails if the option's inversion is dropped and the Go zero value
		// is read as upstream's `false`.
		{"an object pattern with no configuration at all", "const { props, state } = this;", "{}"},
		{"an array pattern with no configuration at all", "const [foo] = this;", "{}"},
		{"a destructuring assignment with no configuration at all", "let foo; [foo] = this", "{}"},
		// A class property initialized to `this` is not a variable declaration and not an
		// assignment, so neither arm sees it. Silent on the release binary.
		{"a class property holding this", "class Z { p = this; }", "{}"},
		// The right-hand side has to be `this` itself rather than something derived from it.
		{"an initializer reading a member of this", "const self = this.foo;", "{}"},
		{"a declaration with no initializer", "let self;", "{}"},
		// A non-assignment binary operator carrying `this` on the right.
		{"a comparison against this", "declare let foo: any;\nconst r = foo === this;", "{}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoThisAlias, thisAliasFile,
				testCase.sourceText, decodeThisAliasOptionsForTest(t, testCase.configuration)))
		})
	}
}

// TestNoThisAliasDeclinesJavaScript pins the file gate, which upstream writes as `should_run`.
//
// It is not an optimization. Measured on the release binary, the identical source reports in a
// `.ts` file and is silent in a `.js` one, so dropping the gate would report every `const self =
// this` in every JavaScript file in a tree upstream leaves alone.
func TestNoThisAliasDeclinesJavaScript(t *testing.T) {
	t.Parallel()

	const source = "const self = this;"

	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, NoThisAlias, "/repository/source/Thing.ts", source,
			NoThisAliasOptions{}),
		"thisAssignment")

	for _, fileName := range []string{"/repository/source/Thing.js", "/repository/source/Thing.jsx"} {
		rule_testing.ExpectClean(t,
			rule_testing.RunWithOptions(t, NoThisAlias, fileName, source, NoThisAliasOptions{}))
	}
}

// TestNoThisAliasPointsAtTheAliasItself asserts where each finding lands.
//
// `ExpectFindings` sees only ids and counts, so a rule reporting the whole declaration, or the
// parenthesized target rather than the identifier inside it, passes every fixture above while
// pointing somewhere upstream does not. The expected text is a literal typed here rather than a
// value read back off the rule, so a mutation moving the report site cannot move the assertion with
// it.
func TestNoThisAliasPointsAtTheAliasItself(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		sourceText    string
		configuration string
		wantReported  string
	}{
		// Upstream's snapshot puts this at 1:7, which is the `self` identifier and not the
		// declaration.
		{"an identifier alias", "const self = this;", "{}", "self"},
		// Upstream's snapshot puts this at 1:7 spanning sixteen columns, which is the pattern
		// without the `const` and without the initializer.
		{"an object pattern", "const { props, state } = this;", `{"allowDestructuring": false}`, "{ props, state }"},
		{"an array pattern", "const [ props, state ] = this;", `{"allowDestructuring": false}`, "[ props, state ]"},
		// The span excludes the parentheses and the assertion, which is the visible consequence of
		// unwrapping the target rather than reporting it whole. Upstream's snapshot for the second
		// of these is at 1:11, the `foo` inside `(foo as any)`.
		{"a parenthesized target", "let foo; (foo) = this", "{}", "foo"},
		{"a target through a type assertion", "let foo; (foo as any) = this", "{}", "foo"},
		{"a doubly parenthesized target", "let foo; ((foo)) = this", "{}", "foo"},
		// A destructuring assignment target reports whole, unlike an identifier target.
		{"an object destructuring assignment", "let foo; ({foo} = this)", `{"allowDestructuring": false}`, "{foo}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoThisAlias, thisAliasFile, testCase.sourceText,
				decodeThisAliasOptionsForTest(t, testCase.configuration))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantReported {
				t.Errorf("finding covered %q, wanted %q", reported, testCase.wantReported)
			}
		})
	}
}

// TestNoThisAliasMessagesSayWhichJudgmentFired asserts the two messages are distinct and that each
// carries the text this port wrote.
//
// Assertions are against literals typed here rather than against the rule's own message constants,
// because comparing a finding to the constant it was built from is an equality that moves with any
// mutation of that constant and therefore guards nothing.
func TestNoThisAliasMessagesSayWhichJudgmentFired(t *testing.T) {
	t.Parallel()

	identifier := rule_testing.RunWithOptions(t, NoThisAlias, thisAliasFile, "const self = this;",
		NoThisAliasOptions{})
	if len(identifier.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(identifier.Diagnostics))
	}

	destructure := rule_testing.RunWithOptions(t, NoThisAlias, thisAliasFile,
		"const { props } = this;", NoThisAliasOptions{ReportDestructuring: true})
	if len(destructure.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(destructure.Diagnostics))
	}

	if messageNoThisAlias.Id != "thisAssignment" {
		t.Errorf("identifier message id is %q, wanted %q", messageNoThisAlias.Id, "thisAssignment")
	}
	if messageNoThisDestructure.Id != "thisDestructure" {
		t.Errorf("destructure message id is %q, wanted %q", messageNoThisDestructure.Id, "thisDestructure")
	}
	if messageNoThisAlias.Description == messageNoThisDestructure.Description {
		t.Error("the two messages carry the same description, so a reader cannot tell which judgment fired")
	}
}

// TestNoThisAliasDecodeMapsUpstreamSpellings covers the decoder directly.
//
// The inversion and the alias are the two lines in this port with no counterpart upstream, so they
// get assertions of their own rather than only being exercised through fixtures.
func TestNoThisAliasDecodeMapsUpstreamSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                    string
		configuration           string
		wantReportDestructuring bool
		wantAllowedNames        []string
	}{
		{"an empty object leaves upstream's permissive default", "{}", false, nil},
		{"allowDestructuring true is the default spelled out", `{"allowDestructuring": true}`, false, nil},
		{"allowDestructuring false turns reporting on", `{"allowDestructuring": false}`, true, nil},
		{"allowedNames binds", `{"allowedNames": ["self"]}`, false, []string{"self"}},
		{"the typo'd allowNames binds the same field", `{"allowNames": ["self"]}`, false, []string{"self"}},
		// Upstream refuses both keys at once as a configuration error. There is no error channel
		// for that here, so the correctly-spelled key wins, which is recorded rather than left to
		// be discovered.
		{"both spellings at once prefers the correct one", `{"allowedNames": ["a"], "allowNames": ["b"]}`, false, []string{"a"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoThisAliasOptions(json.RawMessage(testCase.configuration))
			if err != nil {
				t.Fatalf("decoding: %v", err)
			}
			options, ok := decoded.(NoThisAliasOptions)
			if !ok {
				t.Fatalf("decoder produced %T", decoded)
			}
			if options.ReportDestructuring != testCase.wantReportDestructuring {
				t.Errorf("ReportDestructuring is %v, wanted %v",
					options.ReportDestructuring, testCase.wantReportDestructuring)
			}
			if len(options.AllowedNames) != len(testCase.wantAllowedNames) {
				t.Fatalf("AllowedNames is %v, wanted %v", options.AllowedNames, testCase.wantAllowedNames)
			}
			for index, name := range testCase.wantAllowedNames {
				if options.AllowedNames[index] != name {
					t.Errorf("AllowedNames[%d] is %q, wanted %q", index, options.AllowedNames[index], name)
				}
			}
		})
	}
}

package base

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// consistencyRequirePaginationArgumentNameFile is where the fixtures pretend to live.
const consistencyRequirePaginationArgumentNameFile = "/repository/source/ConsistencyRequirePaginationArgumentName.ts"

// These fixtures are WRITTEN rather than imported, which inverts the usual advice.
//
// This is one of our own base rules, so there is no upstream corpus, and a fixture written beside
// the port encodes the same beliefs the port does: it passes for exactly the reason the code would
// be wrong. Nothing in this file can catch that on its own.
//
// The mitigation is that the source repository is a real oracle. Every expectation below was
// measured by running the ORIGINAL rule, through the ESLint API against
// api-phi-health's own LintConfiguration.ts, over a seeded file placed inside that project's
// tsconfig include, and comparing its findings to this port's on the same source. Seven reporting
// shapes and nine clean ones agreed on every line and column.
//
// Two instrument failures were hit and fixed on the way, and both are worth naming because each
// produced a plausible wrong answer rather than an error. A seed placed in a dot-directory was
// outside the project's include, so the run came back with one "finding" that was a parsing error
// rather than a verdict. And the whole-tree run reports zero for this rule, which is a real property
// of that tree rather than a broken instrument: the control said 2310 files linted and 44 findings
// from other rules, so the run happened and this rule genuinely has nothing to say there.
func TestPaginationDecoratorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a literal name that is not a pagination name",
			"class C { m(@GraphQlArgument('bad') p: PaginationInput) {} }"},
		{"an explicit thunk with a wrongly named parameter",
			"class C { m(@GraphQlArgument(() => UserPaginationInput) items: UserPaginationInput) {} }"},
		{"no literal, so the parameter's own name is judged",
			"class C { m(@GraphQlArgument() wrong: PaginationInput) {} }"},
		{"an options object supplies no name, so the parameter's is judged",
			"class C { m(@GraphQlArgument({ description: 'x' }) wrong: PaginationInput) {} }"},
		{"a defaulted parameter",
			"class C { m(@GraphQlArgument('bad') p: PaginationInput = d) {} }"},
		{"a thunk followed by a literal takes the parameter name, not the literal",
			"class C { m(@GraphQlArgument(() => UserPaginationInput, 'ignored') wrong: string) {} }"},
		{"a constructor parameter property",
			"class C { constructor(@GraphQlArgument('bad') private p: PaginationInput) {} }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ConsistencyRequirePaginationArgumentName,
				consistencyRequirePaginationArgumentNameFile, testCase.sourceText), "invalidName")
		})
	}
}

// The clean cases, and the last four are where this rule's real discrimination lives.
//
// The two name forms and the non-pagination type are the obvious half. The interesting half is what
// the rule declines to look at: a qualified-name thunk and a block-bodied thunk are not the shape
// the original matches, a decorator that is not GraphQlArgument is not its business, and a decorator
// on a property or a method is not on a parameter at all.
//
// Those last two rows are also what keeps this rule from taking the whole package down. Our walk
// recovers per FILE rather than per rule, so one Text() call on a property access costs every rule
// in this package every finding in that file, and both a qualified type name and a member-expression
// arrow body are property accesses. Each is measured clean against the original rather than assumed.
func TestPaginationDecoratorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"named pagination exactly",
			"class C { m(@GraphQlArgument('pagination') pagination: PaginationInput) {} }"},
		{"a disambiguating name ending in Pagination",
			"class C { m(@GraphQlArgument('postsPagination') p: PaginationInput) {} }"},
		{"another disambiguating name",
			"class C { m(@GraphQlArgument('alsoPagination') p: PaginationInput) {} }"},
		{"no literal and the parameter is named pagination",
			"class C { m(@GraphQlArgument() pagination: PaginationInput) {} }"},
		{"not a pagination type at all",
			"class C { m(@GraphQlArgument('items') items: string) {} }"},
		{"a type whose name does not end in PaginationInput",
			"class C { m(@GraphQlArgument('bad') p: SomethingElse) {} }"},
		{"a qualified name in the thunk is not a bare identifier",
			"class C { m(@GraphQlArgument(() => Namespace.PaginationInput) wrong: string) {} }"},
		{"a block bodied thunk is not a bare identifier",
			"class C { m(@GraphQlArgument(() => { return PaginationInput; }) wrong: string) {} }"},
		{"a different decorator entirely",
			"class C { m(@Other('bad') p: PaginationInput) {} }"},
		{"a decorator on a property rather than a parameter",
			"class C { @GraphQlArgument('bad') p = 1; }"},
		{"a decorator on a method rather than a parameter",
			"class C { @GraphQlArgument('bad') m() {} }"},
		{"a decorator on the class itself",
			"@GraphQlArgument('bad') class C {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ConsistencyRequirePaginationArgumentName,
				consistencyRequirePaginationArgumentNameFile, testCase.sourceText))
		})
	}
}

// The finding points at whichever token supplied the name.
//
// The literal when the decorator gives one, the parameter's own name when it does not. That is the
// token the author has to edit, and every ExpectFindings assertion above is satisfied by a rule
// reporting the whole decorator instead. Measured against the original: it reports column 24 for the
// literal form and column 26 for the parameter form on the seeded file, which are those two tokens.
func TestPaginationDecoratorPointsAtTheName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"the literal when the decorator names the argument",
			"class C { m(@GraphQlArgument('bad') p: PaginationInput) {} }", "'bad'"},
		{"the parameter name when the decorator does not",
			"class C { m(@GraphQlArgument() wrong: PaginationInput) {} }", "wrong"},
		{"the parameter name when the first argument is a thunk",
			"class C { m(@GraphQlArgument(() => UserPaginationInput, 'x') wrong: string) {} }", "wrong"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequirePaginationArgumentName, consistencyRequirePaginationArgumentNameFile,
				testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "invalidName")
			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("expected the finding on %q, pointed at %q", testCase.want, reported)
			}
		})
	}
}

// The suffix match is on the type name, and it is anchored at the end.
//
// `PaginationInput$` is the original's own test and it accepts every `@PaginationInputFor` subclass
// without a list of them existing anywhere. Anchoring matters in both directions: a type merely
// containing the word is not a paginator, and the base type itself is.
func TestPaginationDecoratorMatchesTheTypeSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"the base type", "class C { m(@GraphQlArgument('bad') p: PaginationInput) {} }",
			[]string{"invalidName"}},
		{"a generated subclass", "class C { m(@GraphQlArgument('bad') p: PostPaginationInput) {} }",
			[]string{"invalidName"}},
		{"a name merely containing the word",
			"class C { m(@GraphQlArgument('bad') p: PaginationInputBuilder) {} }", nil},
		{"a name that is a prefix only",
			"class C { m(@GraphQlArgument('bad') p: Pagination) {} }", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ConsistencyRequirePaginationArgumentName,
				consistencyRequirePaginationArgumentNameFile, testCase.sourceText), testCase.messages...)
		})
	}
}

// The argument-name test is anchored at the END, and only an anchored form is correct.
//
// Written for a surviving mutant: unanchoring `Pagination$` changed no fixture, because every case
// above either matches at the end or does not contain the word at all. The distinguishing shapes are
// a name carrying the word in the middle and one carrying it at the start, and both of them REPORT.
//
// Measured against the original rule rather than reasoned about, on a seeded file inside
// api-phi-health's tsconfig: it reports all three of these at the same columns this port does.
// `pagination` exactly and a name ending in `Pagination` are the only two accepted spellings.
func TestPaginationDecoratorAnchorsTheNameSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"the word in the middle of the name",
			"class C { m(@GraphQlArgument('myPaginationThing') p: PaginationInput) {} }",
			[]string{"invalidName"}},
		{"the word at the start of the name",
			"class C { m(@GraphQlArgument('Paginationx') p: PaginationInput) {} }",
			[]string{"invalidName"}},
		{"the exact name with a suffix appended",
			"class C { m(@GraphQlArgument('paginationExtra') p: PaginationInput) {} }",
			[]string{"invalidName"}},
		{"the word at the end, which is the disambiguating form",
			"class C { m(@GraphQlArgument('postsPagination') p: PaginationInput) {} }", nil},
		{"the exact name", "class C { m(@GraphQlArgument('pagination') p: PaginationInput) {} }", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ConsistencyRequirePaginationArgumentName,
				consistencyRequirePaginationArgumentNameFile, testCase.sourceText), testCase.messages...)
		})
	}
}

// A qualified type annotation is declined, and the guard that declines it prevents a panic.
//
// `p: Namespace.PaginationInput` gives the parameter a type name of kind KindQualifiedName rather
// than an identifier, and `Text()` on one PANICS: measured, `Unhandled case in Node.Text:
// *ast.QualifiedName`. The walk recovers per FILE rather than per rule, so that would cost every
// rule in this package every finding in the file while the run still printed a plausible summary.
//
// Written for a surviving mutant, and the mutant is the reason this is here at all: no fixture above
// reached the guard, because none of them wrote a qualified annotation. Two things had to be
// established separately. That the shape is reachable, which a probe showed. And that declining is
// CORRECT rather than merely safe, which the original settles: measured on a seeded file inside
// api-phi-health's tsconfig, it reports the bare annotation and stays silent on the qualified one.
//
// So the guard is load bearing twice over, and a message-id fixture is the wrong instrument for half
// of it: no ExpectFindings assertion can see a panic. The clean assertion here is what would fail,
// by taking the whole run down rather than by disagreeing.
func TestPaginationDecoratorDeclinesAQualifiedTypeName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"a namespaced annotation",
			"class C { m(@GraphQlArgument('bad') p: Namespace.PaginationInput) {} }", nil},
		{"a deeply namespaced annotation",
			"class C { m(@GraphQlArgument('bad') p: A.B.PaginationInput) {} }", nil},
		{"the control, a bare annotation, which still reports",
			"class C { m(@GraphQlArgument('bad') p: PaginationInput) {} }",
			[]string{"invalidName"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ConsistencyRequirePaginationArgumentName,
				consistencyRequirePaginationArgumentNameFile, testCase.sourceText), testCase.messages...)
		})
	}
}

package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferRestParamsFile is where the fixtures pretend to live.
const preferRestParamsFile = "/repository/source/PreferRestParams.ts"

// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/prefer_rest_params.rs`:
// 7 pass, 6 fail, and the snapshot records 6 diagnostics from those 6 inputs, so one finding per
// input is measured rather than assumed. Copied rather than rewritten, because a fixture a porter
// invents encodes the same belief as the port and passes for exactly the reason the code is wrong.
//
// The rule declares NeedsTypeChecker, so every case here runs through rule_testing.RunTyped. With the
// plain harness the rule receives a nil checker, goes completely silent, and every StaysSilent case
// below would pass while proving nothing. TestPreferRestParamsRequiresTheTypedHarness pins that.
func TestPreferRestParamsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare reference in a function", "function foo() { arguments; }"},
		{"a computed access with a literal index", "function foo() { arguments[0]; }"},
		{"a computed access with another index", "function foo() { arguments[1]; }"},
		{"a computed access with a symbol key", "function foo() { arguments[Symbol.iterator]; }"},
		{"a reference in a nested function", "function foo() { function bar() { arguments; } }"},
		{"a reference in an arrow inside a function", "function foo() { var bar = () => arguments; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, testCase.sourceText),
				"preferRestParams")
		})
	}
}

// The clean cases are the whole discrimination and each fails a different way.
//
// The first has no enclosing function, so there is no implicit binding to prefer a rest parameter
// over. The next two declare `arguments` themselves, as a parameter and as a local, so the name
// resolves to source rather than to the implicit binding. The fourth is an arrow at the top level,
// which has no `arguments` of its own and nothing to inherit from. The fifth already uses a rest
// parameter. The last two are normal member access, which upstream deliberately permits: reading
// `arguments.length` or `arguments.callee` is not the array-like abuse this rule targets.
func TestPreferRestParamsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a reference outside any function", "arguments;"},
		{"a parameter shadowing the implicit binding", "function foo(arguments) { arguments; }"},
		{"a local shadowing the implicit binding", "function foo() { var arguments; arguments; }"},
		{"an arrow at the top level", "var foo = () => arguments;"},
		{"a function already using a rest parameter", "function foo(...args) { args; }"},
		{"a dotted length read", "function foo() { arguments.length; }"},
		{"a dotted callee read", "function foo() { arguments.callee; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, testCase.sourceText))
		})
	}
}

// ExpectFindings asserts message ids and count and nothing else, so a rule whose defect is *where*
// it points passes a complete fixture pair while being wrong. This rule carries no fix, which is
// exactly the case where a span defect has no other guard at all.
//
// Every finding must point at the `arguments` identifier itself, not at the enclosing statement,
// the member expression, or the function.
func TestPreferRestParamsPointsAtTheIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a bare reference", "function foo() { arguments; }", "arguments"},
		// The span must be the object, not the whole `arguments[0]` element access.
		{"a computed access", "function foo() { arguments[0]; }", "arguments"},
		// The nested case is where an enclosing-function span would be caught: reporting the outer
		// `foo` would still be one finding with the right id.
		{"a nested function", "function foo() { function bar() { arguments; } }", "arguments"},
		// Likewise the arrow, which must report at its own identifier rather than at the function
		// whose binding it inherits.
		{"an arrow inside a function", "function foo() { var bar = () => arguments; }", "arguments"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding to check the span of, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Errorf("the finding points at %q, wanted %q", reported, testCase.wantText)
			}
		})
	}
}

// The arrow case deserves its own assertion rather than only a count, because "reports once" and
// "reports at the right one of two candidate identifiers" are different claims. An arrow inherits
// the enclosing function's implicit binding, so the finding belongs to the arrow's own reference.
func TestPreferRestParamsReportsTheArrowsOwnReference(t *testing.T) {
	t.Parallel()

	const sourceText = "function outer() { arguments; var bar = () => arguments; }"

	result := rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, sourceText)
	rule_testing.ExpectFindings(t, result, "preferRestParams", "preferRestParams")

	// Both references are to the same inherited binding, and each is reported where it is written.
	wantPositions := []int{19, 46}
	for index, want := range wantPositions {
		if got := result.Diagnostics[index].Range.Pos(); got != want {
			t.Errorf("finding %d starts at %d, wanted %d", index, got, want)
		}
	}
}

// This rule proposes no repair, and that is a decision rather than an omission: rewriting to a rest
// parameter changes the enclosing function's signature and every use of the object, which is not a
// single-span edit and has to invent a parameter name. Under this tree's rule that is not even a
// suggestion. A later commit adding a fix should have to delete this test on purpose.
func TestPreferRestParamsProposesNoRepair(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, "function foo() { arguments; }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
		t.Errorf("wanted no fix, got %d: %+v", len(fixes), fixes)
	}
}

// The rule declares NeedsTypeChecker, so the plain harness hands it a nil checker and it goes
// completely silent. Every StaysSilent case above would then pass vacuously and the Fires cases
// would fail in a way that reads like a rule bug. This pins the distinction so a revert to
// rule_testing.Run fails here, loudly, saying what it actually is.
func TestPreferRestParamsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "function foo() { arguments; }"

	if !PreferRestParams.NeedsTypeChecker {
		t.Fatal("the rule no longer declares NeedsTypeChecker, so this guard is measuring nothing")
	}
	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, sourceText), "preferRestParams")
	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferRestParams, preferRestParamsFile, sourceText))
}

// Cases upstream does not cover, from reading our own code and from probing the checker.
//
// Upstream's corpus has exactly one shadow shape per kind and none of the reach cases, so the
// distinctions that a structural walk would get wrong are untested by it entirely. Each of these
// was measured against the checker before it was written down.
func TestPreferRestParamsResolvesShadowsWherverTheyAreDeclared(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		// A `var` hoists to the function, so a declaration nested two blocks deep still shadows a
		// reference written outside those blocks. No bounded walk up from the reference finds it.
		{"a var nested in blocks", "function foo() { { { var arguments; } } arguments; }"},
		// A catch clause binding shadows within its own block, which a walk stopping at the
		// function boundary would have to know about specifically.
		{"a catch clause binding", "function foo() { try {} catch (arguments) { arguments; } }"},
		// A block-scoped `let` is a different declaration kind reaching a different distance.
		{"a block scoped let", "function foo() { { let arguments = 1; arguments; } }"},
		// The property name is not the object, so it is not the implicit binding at all.
		{"a property spelled the same", "function foo() { foo.arguments; }"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, testCase.sourceText))
		})
	}

	// And the direction that catches an over-eager shadow test. A function boundary reintroduces
	// the implicit binding, so an outer declaration does NOT shadow an inner function's own
	// `arguments`. A rule that walked outward looking for any declaration of the name would go
	// silent on all three of these.
	fires := []struct {
		name       string
		sourceText string
	}{
		{"an outer parameter", "function outer(arguments) { function inner() { arguments; } }"},
		{"a module scope var", "var arguments = 1; function foo() { arguments; }"},
		{"a method of a class", "class C { m() { arguments; } }"},
		{"a method of an object literal", "var o = { m() { arguments; } };"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, PreferRestParams, preferRestParamsFile, testCase.sourceText),
				"preferRestParams")
		})
	}
}

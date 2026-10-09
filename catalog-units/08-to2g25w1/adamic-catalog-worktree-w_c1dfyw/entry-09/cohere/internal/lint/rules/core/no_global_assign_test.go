package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// globalAssignFile is where the fixtures pretend to live.
//
// A real path, because this rule reads the checker: the harness builds an actual program and the
// file has to sit somewhere a tsconfig can reach. It also has to be a file whose standard library
// declares `Object` and `String`, since the whole discrimination is which file declares the name.
const globalAssignFile = "/repository/source/GlobalAssign.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_global_assign.rs`:
// one Tester block, 8 pass and 8 fail. The extractor reported 9 diagnostics from those 8 fail
// inputs, so one finding per input is wrong here. The extra one belongs to
// `({Object = 0, String = 0} = {});`, which writes two different names and reports twice.
//
// # Which upstream cases this port can carry, and which it cannot
//
// Upstream answers "is this name a read-only global" from a vendored `javascript_globals` table
// keyed by an `env` config (`browser`, `node`) plus a user `globals` map. Neither the table nor
// either config surface exists in this tree, and five of upstream's sixteen cases turn on nothing
// else. Those five are listed in `TestNoGlobalAssignBoundary` below with what each would need,
// rather than deleted, so the gap is a stated divergence and not a silent one.
//
// This port answers the same question from the TypeScript standard library instead: a name declared
// in a declaration file is a global, a name declared in source is a shadow. That covers every
// builtin upstream's `GLOBALS_BUILTIN` covers and nothing that depends on an `env`.
func TestNoGlobalAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// Upstream fail cases, verbatim.
		{"a plain assignment to a builtin", "String = 'hello world';",
			[]string{"noGlobalAssign"}},
		{"an increment of a builtin", "String++;",
			[]string{"noGlobalAssign"}},
		// The discrepancy case. Two names, two findings, and the reason a fixture pinned at one
		// finding per input would be wrong.
		{"two shorthand destructuring targets with defaults", "({Object = 0, String = 0} = {});",
			[]string{"noGlobalAssign", "noGlobalAssign"}},
		{"an assignment from inside a function", "function f() { Object = 1; }",
			[]string{"noGlobalAssign"}},
		{"an assignment to Array", "Array = 1;",
			[]string{"noGlobalAssign"}},

		// Cases upstream does not cover, added from reading our own code and from the shapes
		// `reference.WritesToBinding` distinguishes. Upstream gets its write classification from
		// `is_write()` on a semantic-layer reference, so its corpus never exercises most of these
		// and cannot tell a port which arm it got short.
		{"a compound assignment", "Object += 1;", []string{"noGlobalAssign"}},
		{"a logical assignment", "Object ||= 1;", []string{"noGlobalAssign"}},
		{"a prefix decrement", "--Object;", []string{"noGlobalAssign"}},
		{"an array destructuring target", "[Object] = [];", []string{"noGlobalAssign"}},
		{"a for-of head", "for (Object of []) {}", []string{"noGlobalAssign"}},
		{"a for-in head", "for (Object in {}) {}", []string{"noGlobalAssign"}},
		{"a bare shorthand destructuring target", "({Object} = {});",
			[]string{"noGlobalAssign"}},
		{"a destructuring target on the value side of a property",
			"({b: Object} = {});", []string{"noGlobalAssign"}},
		{"an assignment through parentheses", "(Object) = 1;", []string{"noGlobalAssign"}},
		{"a rest target", "({...Object} = {});", []string{"noGlobalAssign"}},
		// Other builtins, to show the predicate is the declaring file rather than a hardcoded list.
		{"an assignment to Math", "Math = 1;", []string{"noGlobalAssign"}},
		{"an assignment to JSON", "JSON = 1;", []string{"noGlobalAssign"}},
		{"an assignment to Promise", "Promise = 1;", []string{"noGlobalAssign"}},
		{"an assignment to NaN", "NaN = 1;", []string{"noGlobalAssign"}},
		{"an assignment to Infinity", "Infinity = 1;", []string{"noGlobalAssign"}},
		{"two writes to the same global report twice", "Object = 1; Object = 2;",
			[]string{"noGlobalAssign", "noGlobalAssign"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases, and the one that upstream's corpus structurally cannot contain.
//
// Upstream iterates `root_unresolved_references()`, and a shadowed name resolves, so it never enters
// the iteration at all. The shadow filter is the choice of collection rather than a test, which is
// why upstream's 8 pass cases contain **zero** shadowing cases. A port matching on the name passes
// every visible clean case and still reports `function f(Object) { Object = 1; }`, which is correct
// code. The absent test is the hazard here, so it is written below rather than imported.
func TestNoGlobalAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// Upstream pass cases, verbatim, minus the five that turn on config this tree does not
		// model. Those are in TestNoGlobalAssignBoundary.
		{"an assignment to a lowercase name that is not a builtin", "string='1';"},
		{"a declaration of a lowercase name", "var string;"},
		{"an assignment to a member of a global", "window[parseInt('42', 10)] = 99;"},

		// The shadowing cases upstream cannot have. Each of these is textually a write to a builtin
		// name and each binds to a local declaration instead, so a spelling-only port reports all of
		// them.
		{"a write to a parameter shadowing a builtin", "function f(Object) { Object = 1; }"},
		{"a write to a let shadowing a builtin", "let Object; Object = 1;"},
		{"a write to a var shadowing a builtin", "var Object; Object = 1;"},
		{"a write to a const-declared shadow", "{ const Object = 1; foo(Object); }"},
		{"a write to a class declaration shadowing a builtin", "class Object {} Object = 1;"},
		{"a write to a function declaration shadowing a builtin",
			"function Object() {} Object = 1;"},
		{"a write to an imported binding shadowing a builtin",
			"import Object from 'x'; Object = 1;"},
		{"a write to a catch parameter shadowing a builtin",
			"try {} catch (Object) { Object = 1; }"},
		{"a shorthand destructuring target that binds to a parameter",
			"function f(Object) { ({Object} = {}); }"},
		{"a write to a block-scoped shadow", "{ let Object; Object = 1; }"},

		// Reads of a global, which resolve to the global and do not write it. This is the half
		// symbol identity alone cannot see, and it is why the rule needs the structural test too.
		{"a property write on a global", "Object.x = 1;"},
		{"a global passed as an argument", "foo(Object);"},
		{"a global read in an expression", "let x = Object;"},
		{"a global in a shorthand property that is not a destructuring target",
			"const o = {Object};"},
		{"a global named as a property key rather than a target", "({Object: b} = {});"},
		{"a call on a global", "Object.keys({});"},
		{"a negation of a global", "!Object;"},
		{"a global on the right of an assignment", "let x; x = Object;"},
		{"a global compared", "if (x === Object) {}"},
		{"a fresh binding in a for-of head", "for (const Object of []) {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, testCase.sourceText))
		})
	}
}

// The exceptions option, which the inventory column said this rule does not have.
//
// It does. `NoGlobalAssignConfig { exceptions: Vec<CompactStr> }` in oxc, and the matching
// `schema: [{ properties: { exceptions: { type: "array", items: { type: "string" } } } }]` in
// ESLint. Upstream exercises it with one pass case, which is the first below.
func TestNoGlobalAssignExceptions(t *testing.T) {
	t.Parallel()

	t.Run("a listed name is not reported", func(t *testing.T) {
		t.Parallel()
		// Upstream pass case 3, verbatim source and options.
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoGlobalAssign, globalAssignFile,
			"Object = 0;", NoGlobalAssignOptions{Exceptions: []string{"Object"}}))
	})

	t.Run("an unlisted name is still reported", func(t *testing.T) {
		t.Parallel()
		// The other half of the option, which upstream does not test: an exceptions list must
		// exempt the names in it and nothing else. Without this, an option handler that exempts
		// everything once any name is listed passes upstream's only case.
		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoGlobalAssign, globalAssignFile,
			"String = 0;", NoGlobalAssignOptions{Exceptions: []string{"Object"}}), "noGlobalAssign")
	})

	t.Run("an empty exceptions list exempts nothing", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoGlobalAssign, globalAssignFile,
			"Object = 0;", NoGlobalAssignOptions{Exceptions: []string{}}), "noGlobalAssign")
	})

	t.Run("several listed names are all exempt", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoGlobalAssign, globalAssignFile,
			"({Object = 0, String = 0} = {});",
			NoGlobalAssignOptions{Exceptions: []string{"Object", "String"}}))
	})
}

// The span, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the whole statement
// or at the declaring interface in the standard library passes every case above. Upstream's snapshot
// labels the identifier being written and nothing wider, and the offsets in it are the assertion
// here: `({Object = 0, String = 0} = {});` labels column 3 and column 15, which are the two names
// rather than the object literal around them.
//
// Sliced out of the source with the finding's own range rather than compared against an offset the
// test computes, since a computed offset is wrong in the same direction as the code that produced it.
func TestNoGlobalAssignPointsAtTheWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a plain assignment", "String = 'hello world';", []string{"String"}},
		{"an increment", "String++;", []string{"String"}},
		{"two shorthand targets point at their own names", "({Object = 0, String = 0} = {});",
			[]string{"Object", "String"}},
		{"an assignment inside a function", "function f() { Object = 1; }", []string{"Object"}},
		{"a parenthesized target points at the name rather than the parentheses",
			"(Object) = 1;", []string{"Object"}},
		{"two writes point at their own sites", "Object = 1; Object = 2;",
			[]string{"Object", "Object"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			// The harness trims and appends a newline, so read positions against that same text.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			for i, want := range testCase.want {
				reported := source[result.Diagnostics[i].Range.Pos():result.Diagnostics[i].Range.End()]
				if reported != want {
					t.Errorf("finding %d pointed at %q, wanted %q", i, reported, want)
				}
			}
		})
	}
}

// The two shorthand findings must not share a span.
//
// A rule reporting the object literal twice satisfies both the count and the id assertions above and
// points both findings at the same place. `no-class-assign` carries the same guard for the same
// reason.
func TestNoGlobalAssignShorthandFindingsHaveDistinctSpans(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, "({Object = 0, String = 0} = {});")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted 2 findings, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Errorf("both findings pointed at offset %d; each should point at its own name",
			result.Diagnostics[0].Range.Pos())
	}
}

// The typed harness is required, and a revert to the untyped one must fail loudly.
//
// The engine hands a rule a nil checker unless it declares NeedsTypeChecker. This rule answers
// nothing without one, so under `rule_testing.Run` it goes completely silent: every clean case above
// would pass vacuously and the whole suite would look green over a rule that reports nothing.
func TestNoGlobalAssignRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoGlobalAssign.NeedsTypeChecker {
		t.Fatal("NoGlobalAssign must declare NeedsTypeChecker; without it the checker is nil and " +
			"the rule reports nothing while every StaysSilent fixture passes vacuously")
	}
	untyped := rule_testing.Run(t, NoGlobalAssign, globalAssignFile, "String = 'hello world';")
	if len(untyped.Diagnostics) != 0 {
		t.Fatalf("expected the untyped harness to produce nothing, got %d findings; if this "+
			"changed, the fixtures above may no longer be measuring what they claim",
			len(untyped.Diagnostics))
	}
	typed := rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, "String = 'hello world';")
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("expected the typed harness to report once, got %d", len(typed.Diagnostics))
	}
}

// The five upstream cases this port does not carry, and exactly what each would need.
//
// Not deleted, because a divergence nobody wrote down is indistinguishable from a bug. Each of these
// is a real upstream case whose verdict turns on a configuration surface this tree does not have:
//
//	("top = 0;", None, None)                                     pass    no env, so `top` is nothing
//	("top = 0;", None, {"env": {"browser": true}})               fail    needs the browser table
//	("onload = 0;", None, {"env": {"browser": true}})            pass    browser, and writable
//	("require = 0;", None, None)                                 pass    no env
//	("require = 0;", None, {"env": {"node": true}})              fail    needs the node table
//	("a = 1", None, {"globals": {"a": true}})                    pass    user globals, writable
//	("a = 1", None, {"globals": {"a": false}})                   fail    user globals, readonly
//
// Three of those source texts appear on both sides, separated only by the third tuple slot, which is
// the lint configuration rather than the rule's own options. Without an `env` table and a `globals`
// map there is no information in the file that separates them, so this port reports on none of them.
// That is the conservative direction: a name this tree cannot prove is a read-only global is left
// alone. The test below pins that behavior so it is a decision rather than a drift.
func TestNoGlobalAssignBoundary(t *testing.T) {
	t.Parallel()

	// Each of these is a name with no declaration anywhere the program can see. Upstream would
	// report the ones its env tables cover; this port does not, and asserting that keeps the
	// boundary visible if a globals table ever lands.
	cases := []struct {
		name       string
		sourceText string
	}{
		{"an env-dependent browser global", "top = 0;"},
		{"a writable browser global", "onload = 0;"},
		{"an env-dependent node global", "require = 0;"},
		{"a user-declared global", "a = 1"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoGlobalAssign, globalAssignFile, testCase.sourceText))
		})
	}
}

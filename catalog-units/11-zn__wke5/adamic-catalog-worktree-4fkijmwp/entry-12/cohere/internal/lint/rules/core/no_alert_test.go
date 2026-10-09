package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// alertFile is where the fixtures pretend to live.
const alertFile = "/repository/source/Alert.ts"

// The corpus is upstream's, extracted from its tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/no-alert.js` carries 20 valid cases and 22 invalid ones,
// every invalid case naming exactly one `unexpected`. Twenty are reproduced below; the remaining two
// are decided above this rule by source type and have their own test.
//
// The clean list is almost entirely shadowing, in five different binding positions: a function
// declaration, a var, a nested var, a parameter, and a var declared in an enclosing function. Read
// against the reporting list they pair up, and the pairs are the whole point: `var alert = function()
// {}; window.alert(foo)` REPORTS while `var alert = function() {}; alert();` is clean, because
// shadowing the name `alert` says nothing about the property `window.alert`.
func TestNoAlertStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"a[o.k](1)"},
		{"foo.alert(foo)"},
		{"foo.confirm(foo)"},
		{"foo.prompt(foo)"},
		{"function alert() {} alert();"},
		{"var alert = function() {}; alert();"},
		{"function foo() { var alert = bar; alert(); }"},
		{"function foo(alert) { alert(); }"},
		{"var alert = function() {}; function test() { alert(); }"},
		{"function foo() { var alert = function() {}; function test() { alert(); } }"},
		{"function confirm() {} confirm();"},
		{"function prompt() {} prompt();"},
		{"window[alert]();"},
		{"function foo() { this.alert(); }"},
		{"function foo() { var window = bar; window.alert(); }"},
		{"var globalThis = foo; globalThis.alert();"},
		{"function foo() { var globalThis = foo; globalThis.alert(); }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText))
		})
	}
}

// The twenty reporting cases this port reproduces exactly.
func TestNoAlertFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"alert(foo)"},
		{"window.alert(foo)"},
		{"window['alert'](foo)"},
		{"confirm(foo)"},
		{"window.confirm(foo)"},
		{"window['confirm'](foo)"},
		{"prompt(foo)"},
		{"window.prompt(foo)"},
		{"window['prompt'](foo)"},
		{"function alert() {} window.alert(foo)"},
		{"var alert = function() {};\nwindow.alert(foo)"},
		{"function foo(alert) { window.alert(); }"},
		{"function foo() { alert(); }"},
		{"function foo() { var alert = function() {}; }\nalert();"},
		{"function foo() { var window = bar; window.alert(); }\nwindow.alert();"},
		{"globalThis['alert'](foo)"},
		{"globalThis.alert();"},
		{"function foo() { var globalThis = bar; globalThis.alert(); }\nglobalThis.alert();"},
		{"window?.alert(foo)"},
		{"(window?.alert)(foo)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText), "unexpected")
		})
	}
}

// The two upstream reporting cases that are clean here, and the reason is not this rule.
//
// Upstream reports `this.alert(foo)` at the top level of a SCRIPT, where `this` is the global
// object. Our harness pins `moduleDetection: "force"` in `internal/rule_testing/program.go` and our tree
// is modules throughout, and in a module top-level `this` is `undefined` rather than the global.
//
// Measured on the installed rule rather than reasoned about: with `sourceType: "module"` both of
// these come back CLEAN from upstream itself, and with `sourceType: "script"` both report. So this
// is not a divergence in the port at all, it is the same rule reaching the same verdict about a
// different kind of file, and no fixture in this tree can express the reporting half because the
// tree cannot produce a script.
//
// Recorded as passing cases with the measurement rather than deleted, so the next reader does not
// read the missing `this` arm as an oversight and helpfully add one. The control is `window.alert`,
// which reports in both source types.
func TestNoAlertLeavesGlobalThisToScriptFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"this.alert(foo)"},
		{"this['alert'](foo)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText))
		})
	}

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAlert, alertFile,
		"window.alert(foo)"), "unexpected")
}

// Two upstream CLEAN cases that report here, and the discriminator is the language version.
//
// `globalThis.alert();` appears in upstream's corpus twice, once in each list, with identical source
// text. The only difference is `ecmaVersion`: the clean copies are pinned to the default and to
// 2017, and the reporting copy to 2020. Upstream's `isGlobalThisReferenceOrGlobalWindow` asks the
// scope whether a variable named `globalThis` exists, which is how it detects an environment old
// enough not to have one, and below ES2020 the answer is no.
//
// Driven on the installed rule across versions to fix this rather than infer it:
//
//	ecmaVersion=2017  globalThis.alert();       CLEAN
//	ecmaVersion=2020  globalThis.alert();       unexpected
//	ecmaVersion=2022  globalThis.alert();       unexpected
//	ecmaVersion=6     globalThis["alert"]();    CLEAN
//	ecmaVersion=2022  globalThis["alert"]();    unexpected
//
// `internal/rule_testing/program.go` pins `target: "ES2022"` and our tree compiles to the same, so
// `globalThis` is always in the standard library the checker reads and the pre-2020 verdict is not
// reachable here. Reporting these is therefore agreement with upstream AT OUR TARGET rather than a
// divergence from it, and a port reproducing them as clean would be silent on the modern spelling of
// the exact thing this rule exists to catch.
//
// The control is the shadowed form, which upstream reports clean at every version and so does this.
func TestNoAlertReportsGlobalThisAtOurTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"globalThis.alert();"},
		{"globalThis['alert']();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText), "unexpected")
		})
	}

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAlert, alertFile,
		"var globalThis = foo; globalThis.alert();"))
}

// The rendered message, which the id assertions above cannot see.
//
// The Description interpolates the function's name, and a `rule.Message` carries no rendering layer.
// A `confirm` finding whose sentence says `alert` has the right id and the right span while naming
// something that is not on the line.
func TestNoAlertNamesTheFunction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantName   string
	}{
		{"alert(foo)", "alert"},
		{"confirm(foo)", "confirm"},
		{"prompt(foo)", "prompt"},
		{"window['confirm'](foo)", "confirm"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.HasPrefix(description, "This calls `"+testCase.wantName+"`") {
				t.Fatalf("the message reads %q, which does not name %q", description, testCase.wantName)
			}
		})
	}
}

// The reported span is the whole call, which no id fixture above can see.
//
// Upstream passes `node`, the CallExpression, for both shapes. A port anchoring on the callee
// satisfies every assertion above while pointing past the arguments.
func TestNoAlertReportsTheWholeCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{"alert(foo)", "alert(foo)"},
		{"window['prompt'](foo)", "window['prompt'](foo)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAlert, alertFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			written := result.SourceFile.Text()
			reported := written[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantSpan {
				t.Fatalf("reported %q, wanted %q", reported, testCase.wantSpan)
			}
		})
	}
}

// Shapes the corpus does not write, from reading our own code.
//
// The first pair is the optional-chain form of the member call, which upstream ships for `window`
// and not for `globalThis`. The second is a computed access through a variable named for a
// prohibited function, which must stay clean for the same reason `window[alert]()` does: the
// property read is whatever the variable holds.
//
// The last is the one that separates this rule from a text match. `foo.alert(foo)` is upstream's own
// clean case because the receiver is not the global object; the control beside it changes only the
// receiver and reports.
func TestNoAlertHandlesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	t.Run("globalThis with an optional chain reports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAlert, alertFile,
			"globalThis?.alert(foo)"), "unexpected")
	})

	t.Run("a computed access through a variable stays clean", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAlert, alertFile,
			"declare const confirm: string;\nglobalThis[confirm]();\n"))
	})

	t.Run("a receiver that is not the global object stays clean", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAlert, alertFile,
			"declare const foo: any;\nfoo.alert(foo);\n"))
		// The control changes only the receiver.
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAlert, alertFile,
			"declare const foo: any;\nwindow.alert(foo);\n"), "unexpected")
	})
}

// The typed harness is required, so a later revert to `rule_testing.Run` fails loudly.
func TestNoAlertNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoAlert.NeedsTypeChecker {
		t.Fatal("this rule resolves shadowing through the checker and must declare it")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoAlert, alertFile, "alert(foo)"))
}

// A callee that is neither an identifier nor a member access must not crash the file.
//
// `memberAccessObject` answers nil for such a callee and the member branch fed that nil straight into
// `ast.SkipParentheses`, which dereferences it. The walk recovers per FILE rather than per rule, so
// the panic took the whole file away from every rule: 167 files, about five percent of the tree we
// lint, silently unchecked while the run still printed green and the node count sat a quarter of a
// million nodes low.
//
// The two shapes that occur are `super(...)` and `import(...)` — call expressions whose callee is a
// bare keyword. Every subclass constructor carries a `super()`, which is why this was not a rare edge.
//
// # What this test does and does not prove
//
// It pins the shapes so a future edit to the member branch has them written down, and it is NOT the
// regression proof: it passes against the unfixed rule too. The panic needs the real program, and the
// fixtures below never reach the nil path in this harness. The proof is an A/B on the tree itself,
// same tree and same flags, differing only in this file: 167 crashed files before, 0 after. Anyone
// changing this branch should re-run that rather than trust the green below.
func TestNoAlertSurvivesACalleeThatIsNotAMemberAccess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "super call", source: "class Base { constructor(a: number) {} }\nclass Derived extends Base { constructor() { super(1); } }"},
		{name: "super method call", source: "class Base { go() {} }\nclass Derived extends Base { go() { super.go(); } }"},
		{name: "dynamic import", source: "async function load() { await import('./Other'); }"},
		{name: "immediately invoked function", source: "(function() { return 1; })();"},
		{name: "call returning a function", source: "function outer() { return function() {}; }\nouter()();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoAlert, alertFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

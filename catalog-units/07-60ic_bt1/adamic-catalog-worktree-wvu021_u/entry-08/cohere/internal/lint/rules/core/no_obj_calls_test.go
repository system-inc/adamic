package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The ids and spans ESLint gives are in no_obj_calls_corpus_test.go, read off the installed ESLint.
// The tests here pin what that table cannot: the fixture pair at its smallest, the departures, the
// messages, and the checker the rule cannot run without.

// TestNoObjCallsFiresAndStaysQuiet is the fixture pair in its smallest form.
func TestNoObjCallsFiresAndStaysQuiet(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "Math();"), "unexpectedCall")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "let j = JSON; j();"), "unexpectedRefCall")
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "var Math; Math();"))
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "JSON.parse('{}');"))
}

// TestNoObjCallsReportsOnlyTheNamespaceObjects pins the list as the whole discrimination.
//
// ESLint's corpus never calls a different global, so a tracker asked about every global would pass it
// while reporting `parseInt()`. These resolve exactly as `Math` does and are callable.
func TestNoObjCallsReportsOnlyTheNamespaceObjects(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"parseInt('1');",
		"isNaN(1);",
		"let p = parseInt; p('1');",
		"globalThis.parseInt('1');",
		"new globalThis.Date();",
		"new Date();",
	} {
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", source))
	}
}

// TestNoObjCallsTakesAnUndeclaredGlobalObjectAsTheGlobal pins the edge left out of the ESLint table.
//
// The harness's lib has no DOM types, so `window` and `self` resolve to nothing. An undeclared name can
// only be the runtime's global, so `window.JSON()` reports here, where ESLint with no browser globals
// configured does not know `window` and stays silent. In a project with DOM types `window` is declared
// in its lib, and the two agree. A local named `window` is not the global in either.
func TestNoObjCallsTakesAnUndeclaredGlobalObjectAsTheGlobal(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "window.JSON();"), "unexpectedCall")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "var foo = self.Math; foo();"), "unexpectedRefCall")
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "function f(window: any) { window.JSON(); }"))
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "function f(globalThis: any) { globalThis.Math(); }"))
}

// TestNoObjCallsDoesNotFollowAnArithmeticAssignment pins the tracker's one departure from
// eslint-utils, as this rule meets it: after `x += JSON`, x is a string, and the installed ESLint
// reports `x()` as a call of JSON.
func TestNoObjCallsDoesNotFollowAnArithmeticAssignment(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "let x; x += JSON; x();"))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoObjCalls, "input.ts", "let x; x ??= JSON; x();"), "unexpectedRefCall")
}

// TestNoObjCallsNamesTheCalleeAndTheGlobal pins what each message says, which the table's ids cannot.
func TestNoObjCallsNamesTheCalleeAndTheGlobal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		want   []string
	}{
		{"Math();", []string{"`Math`"}},
		{"let m = globalThis.Math; new m();", []string{"`m`", "`Math`"}},
		{"(a ? JSON : b)();", []string{"a value", "`JSON`"}},
	}
	for _, testCase := range cases {
		result := rule_testing.RunTyped(t, NoObjCalls, "input.ts", testCase.source)
		if len(result.Diagnostics) != 1 {
			t.Errorf("%s: %d findings, want 1", testCase.source, len(result.Diagnostics))
			continue
		}
		for _, want := range testCase.want {
			if !strings.Contains(result.Diagnostics[0].Message.Description, want) {
				t.Errorf("%s: the message %q does not name %s", testCase.source, result.Diagnostics[0].Message.Description, want)
			}
		}
	}
}

// TestNoObjCallsRequiresTheTypedHarness fails loudly if the checker declaration is reverted.
//
// Without a checker the rule cannot tell `var Math; Math();` from the global, so it reports nothing
// at all, member calls through `globalThis` included. Every clean case would then pass vacuously.
func TestNoObjCallsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoObjCalls.NeedsTypeChecker {
		t.Fatal("NoObjCalls must declare NeedsTypeChecker; it tells a local from a global by asking the checker")
	}
	for _, source := range []string{"Math();", "globalThis.Math();"} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoObjCalls, "input.ts", source))
		if result := rule_testing.RunTyped(t, NoObjCalls, "input.ts", source); len(result.Diagnostics) != 1 {
			t.Errorf("%s: the typed harness must reach one finding, got %d", source, len(result.Diagnostics))
		}
	}
}

// TestNoObjCallsIsRegistered fails if a rule that passes every fixture lints no file at all.
func TestNoObjCallsIsRegistered(t *testing.T) {
	t.Parallel()

	for _, registration := range rule.Registered() {
		if registration.Rule.Name == "no-obj-calls" {
			if !registration.Rule.NeedsTypeChecker {
				t.Error("the registered copy does not declare NeedsTypeChecker")
			}
			return
		}
	}
	t.Fatal("no-obj-calls is absent from the registry")
}

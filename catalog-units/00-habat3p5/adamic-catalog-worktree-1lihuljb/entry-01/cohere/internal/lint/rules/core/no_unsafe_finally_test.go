package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoUnsafeFinallyReportsEscapingControlFlow is the fixture that must fire.
//
// All four statement kinds, each in the position where its jump actually leaves the finally block.
func TestNoUnsafeFinallyReportsEscapingControlFlow(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		wantId string
	}{
		{"var foo = function() { try { return 1; } catch (err) { return 2; } finally { return 3; } };", "unsafeReturn"},
		{"var foo = function() { try { return 1; } finally { return 3; } };", "unsafeReturn"},
		{"var foo = function() { try { return 1; } catch (err) { return 2; } finally { throw new Error(); } };", "unsafeThrow"},
		{"var foo = function() { try {} finally { throw new Error(); } };", "unsafeThrow"},
		{"while (true) try {} finally { break; }", "unsafeBreak"},
		{"while (true) try {} finally { continue; }", "unsafeContinue"},
		{"for (;;) try {} finally { break; }", "unsafeBreak"},
		{"for (var x in obj) try {} finally { continue; }", "unsafeContinue"},
		{"for (var x of arr) try {} finally { break; }", "unsafeBreak"},
		{"do { try {} finally { break; } } while (true);", "unsafeBreak"},
	} {
		result := rule_testing.Run(t, NoUnsafeFinally, "finally.ts", testCase.source)
		rule_testing.ExpectFindings(t, result, testCase.wantId)
	}
}

// TestNoUnsafeFinallyReportsThroughNestedStatements pins that a statement which does not absorb the
// jump does not hide it either.
//
// A `return` inside a loop inside finally still escapes, because a loop absorbs a break and not a
// return. A rule that stopped at any enclosing statement rather than at the ones that actually
// absorb this kind of jump would go silent on every one of these.
func TestNoUnsafeFinallyReportsThroughNestedStatements(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		wantId string
	}{
		{"var foo = function() { try {} finally { while (true) { return 1; } } };", "unsafeReturn"},
		{"var foo = function() { try {} finally { if (x) { return 1; } } };", "unsafeReturn"},
		{"var foo = function() { try {} finally { switch (x) { case 1: return 1; } } };", "unsafeReturn"},
		{"var foo = function() { try {} finally { { throw new Error(); } } };", "unsafeThrow"},
		{"var foo = function() { try {} finally { try {} catch (e) { return 1; } } };", "unsafeReturn"},
	} {
		result := rule_testing.Run(t, NoUnsafeFinally, "nested.ts", testCase.source)
		rule_testing.ExpectFindings(t, result, testCase.wantId)
	}
}

// TestNoUnsafeFinallyReportsLabeledJumpsLeavingFinally pins the labeled cases.
//
// A labeled break is absorbed by the labeled statement it names, so whether it escapes depends
// entirely on where that label sits. Here every label is outside the finally block.
func TestNoUnsafeFinallyReportsLabeledJumpsLeavingFinally(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		wantId string
	}{
		{"label: try { return 0; } finally { break label; }", "unsafeBreak"},
		{"label: try {} finally { break label; }", "unsafeBreak"},
		{"outer: { try {} finally { inner: { break outer; } } }", "unsafeBreak"},
		{"label: while (true) try {} finally { continue label; }", "unsafeContinue"},
		{"label: for (;;) try {} finally { continue label; }", "unsafeContinue"},
		{"label: while (true) { try {} finally { while (true) { break label; } } }", "unsafeBreak"},
	} {
		result := rule_testing.Run(t, NoUnsafeFinally, "labeled.ts", testCase.source)
		rule_testing.ExpectFindings(t, result, testCase.wantId)
	}
}

// TestNoUnsafeFinallyReportsContinueThroughSwitch pins the asymmetry between break and continue.
//
// This is the case that separates two stopping sets that otherwise look interchangeable. A switch is
// a legal target for `break` and not for `continue`, so a switch written inside the finally block
// absorbs one and not the other. A rule that used one stopping set for both would go silent here
// while every other fixture stayed green.
func TestNoUnsafeFinallyReportsContinueThroughSwitch(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnsafeFinally, "switch.ts",
		"while (true) try {} finally { switch (true) { case true: continue; } }")
	rule_testing.ExpectFindings(t, result, "unsafeContinue")
}

// TestNoUnsafeFinallyStaysSilentWhenTheJumpIsAbsorbed is the half that catches a rule firing on
// correct code.
//
// The function-boundary cases are the bulk of it, and they are the reason the rule cannot simply ask
// whether a statement sits inside a finally block: a callback defined in finally is its own control
// flow, and its `return` belongs to it.
func TestNoUnsafeFinallyStaysSilentWhenTheJumpIsAbsorbed(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		`var foo = function() { try { return 1; } catch (err) { return 2; } finally { console.log("done"); } };`,
		"var foo = function() { try {} finally { function a(x) { return x; } } };",
		"var foo = function() { try {} finally { var a = function(x) { return x; }; } };",
		"var foo = function() { try {} finally { var a = (x) => { return x; }; } };",
		"var foo = function() { try {} finally { var obj = { method() { return 1; } }; } };",
		"var foo = function() { try {} finally { var obj = { get x() { return 1; } }; } };",
		"var foo = function() { try {} finally { class C { constructor() { return; } } } };",
		"var foo = function() { try {} finally { class C { method() { return 1; } } } };",
		"var foo = function() { try {} finally { var C = class { method() { return 1; } }; } };",
		"var foo = function() { try {} finally { async function a() { return 1; } } };",
		"var foo = function() { try {} finally { function* gen() { return 1; } } };",
		"var foo = function() { try {} finally { async function* gen() { return 1; } } };",
		"var foo = function() { try {} finally { var a = async () => { return 1; }; } };",
		"var foo = function() { try {} finally { function a() { throw new Error(); } } };",
		"var foo = function() { try {} finally { var a = function(x) { while (true) { if (x) { break; } else { continue; } } }; } };",

		// A loop or switch written inside finally absorbs its own unlabeled break.
		"var foo = function() { try {} finally { while (true) break; } };",
		"var foo = function() { try {} finally { for (var i = 0; i < 10; i++) break; } };",
		"var foo = function() { try {} finally { for (var x in obj) break; } };",
		"var foo = function() { try {} finally { for (var x of arr) break; } };",
		"var foo = function() { try {} finally { do { break; } while (true); } };",
		"var foo = function() { try {} finally { switch (true) { case true: break; } } };",

		// A loop written inside finally absorbs its own continue.
		"var foo = function() { try {} finally { while (true) continue; } };",
		"var foo = function() { try {} finally { for (var i = 0; i < 10; i++) continue; } };",
		"var foo = function() { try {} finally { for (var x of arr) continue; } };",
		"var foo = function() { try {} finally { do { continue; } while (true); } };",

		// A label declared inside finally is a target inside finally, so the jump never leaves.
		"var foo = function() { try {} finally { label: while (true) { break label; } } };",
		"var foo = function() { try {} finally { label: while (true) { continue label; } } };",
		"var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { continue label; } } };",
		"var foo = function() { try {} finally { label: do { break label; } while (true); } };",
		"var foo = function() { try {} finally { label: { break label; } } };",

		// Control flow in the try or catch block is the point of a try statement.
		"var foo = function() { try { return 1; } catch (err) { return 2; } };",
		`var foo = function() { try { throw new Error(); } catch (err) { return 2; } finally { console.log("done"); } };`,
		`var foo = function() { try { return 1; } catch (err) { throw new Error(); } finally { console.log("done"); } };`,

		// A nested try whose own finally is harmless.
		"var foo = function() { try {} finally { try { console.log(1); } finally { console.log(2); } } };",
		"var foo = function() { try {} finally { var fn = function() { try { return 1; } finally { console.log(2); } }; } };",
		"var foo = function() { try {} finally { try { while (true) { break; } } catch (e) {} } };",

		// A function boundary reached through several layers still stops the walk.
		"var foo = function() { try {} finally { if (true) { var fn = () => { return 1; }; } } };",
		"var foo = function() { try {} finally { while (true) { (function() { return 1; })(); break; } } };",
		"var foo = function() { try {} finally { (function() { class C { method() { return (() => { return 1; })(); } } })(); } };",
	} {
		result := rule_testing.Run(t, NoUnsafeFinally, "clean.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}

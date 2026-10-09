package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const debuggerFile = "/repository/source/Thing.ts"

func TestNoDebuggerFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare statement", "debugger;\n"},
		{"inside a function", "export function run() {\n    debugger;\n}\n"},
		// The single-statement form of an if has no block, so the statement is the branch itself.
		// This is the shape a listener keyed to blocks rather than to statements would miss.
		{"as an unbraced if branch", "export function run(flag: boolean) {\n    if(flag) debugger;\n}\n"},
		{"inside a loop", "export function run() {\n    for(let index = 0; index < 1; index++) {\n        debugger;\n    }\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDebugger, debuggerFile, testCase.sourceText),
				"unexpectedDebugger")
		})
	}
}

func TestNoDebuggerStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The boundary this rule decides on is "is this the debugger statement", and the way to
		// get that wrong is to match the word rather than the syntax. Each of these contains the
		// text and none of them is the statement.
		{"an identifier named debugger-ish", "const debuggerEnabled = true;\nexport const Flag = debuggerEnabled;\n"},
		{"a property named debugger", "export const Settings = { debugger: true };\n"},
		{"the word inside a string", "export const Message = 'debugger';\n"},
		{"the word inside a comment", "// debugger\nexport const Value = 1;\n"},
		{"no statement at all", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDebugger, debuggerFile, testCase.sourceText))
		})
	}
}

// The fix removes the statement rather than commenting it out, because there is no form of it that
// belongs in committed code.
func TestNoDebuggerRemovesTheStatement(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDebugger, debuggerFile, "export function run() {\n    debugger;\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 1 {
		t.Fatalf("want one fix, got %d", len(result.Diagnostics[0].Fixes))
	}
	if text := result.Diagnostics[0].Fixes[0].Text; text != "" {
		t.Fatalf("want a deletion (empty text), got %q", text)
	}

	// The text alone does not pin the rewrite. An empty replacement over the wrong range is still an
	// empty replacement, and it deletes whatever that range covers: pointing this fix at the
	// enclosing function instead of the statement removes the whole function, keeps the text empty,
	// and passes every assertion above. Asserting the resulting source is what closes that.
	// The leading indentation survives, because RemoveNode deletes the node and not the trivia
	// before it. Asserted as it actually is rather than as it reads best: a fixture that states a
	// tidier result than the engine produces is a fixture that will be "fixed" by making the engine
	// wrong.
	rule_testing.ExpectFixedSource(t, result, "export function run() {\n    \n}\n")
}

// TestNoDebuggerRepairsOnlyInAStatementList pins where the removal is offered. In a list of
// statements it is; as the whole body of an unbraced branch, a loop or a label it is not, since
// `if (flag)` with nothing after it does not parse. ESLint's corpus writes that shape and offers no fix
// for any shape (#kq9vtva).
func TestNoDebuggerRepairsOnlyInAStatementList(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDebugger, debuggerFile, "export function run() {\n    debugger;\n}\n")
	rule_testing.ExpectFixedSource(t, result, "export function run() {\n    \n}\n")

	for _, sourceText := range []string{
		"export function run(flag: boolean) {\n    if(flag) debugger;\n}\n",
		"export function run(flag: boolean) {\n    while(flag) debugger;\n}\n",
		"export function run() {\n    here: debugger;\n}\n",
	} {
		result := rule_testing.Run(t, NoDebugger, debuggerFile, sourceText)
		rule_testing.ExpectFindings(t, result, "unexpectedDebugger")
		if len(result.Diagnostics[0].Fixes) != 0 {
			t.Errorf("a fix was offered for the lone body in %q", sourceText)
		}
	}
}

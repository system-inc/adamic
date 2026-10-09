package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// implicitGlobalsFile is where the fixtures pretend to live.
const implicitGlobalsFile = "/repository/source/ImplicitGlobals.ts"

// implicitGlobalsScript is the sloppy JavaScript script the leak half can still fire in.
const implicitGlobalsScript = "/repository/source/ImplicitGlobals.js"

// implicitGlobalsScriptTsConfig is the typed harness's project with TypeScript's own module detection
// rather than `force`, so a file with no import or export is a script, and with JavaScript allowed.
const implicitGlobalsScriptTsConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "auto",
		"allowJs": true,
		"types": []
	},
	"include": ["**/*.ts", "**/*.js"]
}`

// runImplicitGlobalsAsAScript builds a typed program in which the fixture is a script: sloppy for a
// .js file, and strict under alwaysStrict for a .ts one.
func runImplicitGlobalsAsAScript(t *testing.T, fileName string, sourceText string, options any) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFilesWithSetupAndOptions(t, NoImplicitGlobals,
		map[string]string{fileName: sourceText}, fileName, options, func(directory string) {
			path := filepath.Join(directory, "tsconfig.json")
			if err := os.WriteFile(path, []byte(implicitGlobalsScriptTsConfig), 0o644); err != nil {
				t.Fatalf("writing the script tsconfig: %v", err)
			}
		})
}

// implicitGlobalsOptions routes the option through the rule's own exported decoder.
func implicitGlobalsOptions(t *testing.T, lexicalBindings bool) any {
	t.Helper()
	raw, err := json.Marshal(map[string]bool{"lexicalBindings": lexicalBindings})
	if err != nil {
		t.Fatalf("could not marshal the option: %v", err)
	}
	decoded, err := DecodeNoImplicitGlobalsOptions(raw)
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// Each half is decided by what the file is, so the fixtures build each kind of file on purpose.
//
// The declaration halves fire only in a SCRIPT, since a module has no global scope. The leak half
// fires only in SLOPPY code, since in strict code an assignment to an undeclared name throws, and it
// needs the checker to ask whether a name was ever declared. Three harnesses give three kinds of file:
//
//	the untyped harness               a script, with no checker, so only the declaration halves
//	the typed harness                 a module (it forces module detection), strict throughout
//	runImplicitGlobalsAsAScript       a typed script: a .js one sloppy, so both halves fire, and a
//	                                  .ts one strict under alwaysStrict, so only declarations do
//
// Measured with controls in TestNoImplicitGlobalsReadsWhatTheFileIs below.

// The corpus is ESLint's own, imported from eslint/tests/lib/rules/no-implicit-globals.js.
//
// Upstream ships 129 pass and 116 fail. 51 pass and 33 fail are expressible here; the rest
// need eslint configuration surfaces this tool does not have, and each exclusion is counted
// rather than silently dropped:
//
//	configured globals      22 pass, 2 fail
//	/*global*/ directive    21 pass, 49 fail
//	/*exported*/ directive  20 pass, 27 fail
//	module source type      15 pass, 2 fail
//	readonly globals only   3 fail, pinned in TestNoImplicitGlobalsDeclinesReadonlyGlobals
//
// The cases were extracted by loading upstream's tester with a stubbed RuleTester and rendering
// these literals from that JSON, so nothing was retyped. The generator refuses a case holding a
// byte outside printable ASCII.
func TestNoImplicitGlobalsFiresOnGlobalDeclarations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		lexicalBindings bool
		messages        []string
	}{
		{"var foo = 1;", false, []string{"globalNonLexicalBinding"}},
		{"function foo() {}", false, []string{"globalNonLexicalBinding"}},
		{"function *foo() {}", false, []string{"globalNonLexicalBinding"}},
		{"async function foo() {}", false, []string{"globalNonLexicalBinding"}},
		{"async function *foo() {}", false, []string{"globalNonLexicalBinding"}},
		{"var foo = function() {};", false, []string{"globalNonLexicalBinding"}},
		{"var foo = function foo() {};", false, []string{"globalNonLexicalBinding"}},
		{"var foo = function*() {};", false, []string{"globalNonLexicalBinding"}},
		{"var foo = function *foo() {};", false, []string{"globalNonLexicalBinding"}},
		{"var foo = 1, bar = 2;", false, []string{"globalNonLexicalBinding", "globalNonLexicalBinding"}},
		{"const a = 1;", true, []string{"globalLexicalBinding"}},
		{"let a;", true, []string{"globalLexicalBinding"}},
		{"let a = 1;", true, []string{"globalLexicalBinding"}},
		{"class A {}", true, []string{"globalLexicalBinding"}},
		{"const a = 1; const b = 2;", true, []string{"globalLexicalBinding", "globalLexicalBinding"}},
		{"const a = 1, b = 2;", true, []string{"globalLexicalBinding", "globalLexicalBinding"}},
		{"let a, b = 1;", true, []string{"globalLexicalBinding", "globalLexicalBinding"}},
		{"const a = 1; let b; class C {}", true, []string{"globalLexicalBinding", "globalLexicalBinding", "globalLexicalBinding"}},
		{"const [a, b, ...c] = [];", true, []string{"globalLexicalBinding", "globalLexicalBinding", "globalLexicalBinding"}},
		{"let { a, foo: b, bar: { c } } = {};", true, []string{"globalLexicalBinding", "globalLexicalBinding", "globalLexicalBinding"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			// The untyped harness, because a script is the only place these can fire.
			result := rule_testing.RunWithOptions(t, NoImplicitGlobals, implicitGlobalsFile,
				testCase.sourceText, implicitGlobalsOptions(t, testCase.lexicalBindings))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// The leak half, which asks resolution whether a name was ever declared, in the sloppy script it
// can still fire in.
//
// It is not gated on the assignment being at the top level, because a leak creates a global from a
// function too. Upstream's window.foo = function() { bar = 1; } is the case that says so: the
// assignment is inside a function and it still reports.
func TestNoImplicitGlobalsFiresOnLeaks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		messages   []string
	}{
		{"foo = 1", []string{"globalVariableLeak"}},
		{"foo = function() {};", []string{"globalVariableLeak"}},
		{"foo = function*() {};", []string{"globalVariableLeak"}},
		{"window.foo = function() { bar = 1; }", []string{"globalVariableLeak"}},
		{"(function() {}(foo = 1));", []string{"globalVariableLeak"}},
		{"for (foo in {});", []string{"globalVariableLeak"}},
		{"for (foo of []);", []string{"globalVariableLeak"}},
		{"window.foo = { bar() { foo = 1 } }", []string{"globalVariableLeak"}},
		{"foo = 1, bar = 2;", []string{"globalVariableLeak", "globalVariableLeak"}},
		{"foo = bar = 1", []string{"globalVariableLeak", "globalVariableLeak"}},
		{"[foo, bar] = [];", []string{"globalVariableLeak", "globalVariableLeak"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runImplicitGlobalsAsAScript(t, implicitGlobalsScript, testCase.sourceText,
				implicitGlobalsOptions(t, false))
			rule_testing.ExpectFindings(t, result, testCase.messages...)
		})
	}
}

// Two upstream cases report from both halves at once, which a sloppy script shows in one run. The
// declarations are reported first, since the declaration half runs first.
func TestNoImplicitGlobalsFiresFromBothHalves(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		messages   []string
	}{
		{"foo = 1; var bar;", []string{"globalNonLexicalBinding", "globalVariableLeak"}},
		{"var foo = bar = 1;", []string{"globalNonLexicalBinding", "globalVariableLeak"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runImplicitGlobalsAsAScript(t, implicitGlobalsScript,
				testCase.sourceText, implicitGlobalsOptions(t, true)), testCase.messages...)
		})
	}
}

// The clean cases, which are where the rule's real discrimination lives.
//
// Most wrap a declaration in a function or a block, which is exactly the fix the messages
// recommend, so a port reporting a declaration without checking whether it is at the top level
// reports nearly all of them. The lexicalBindings rows are the other half: a global let or const
// is clean by default and reports only when the option asks.
//
// Run through the untyped harness and a sloppy typed script, because a clean case must stay clean in
// the one place both halves can fire. A case clean only where the rule cannot fire is not evidence.
func TestNoImplicitGlobalsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		lexicalBindings bool
	}{
		{"this.foo = 1;", false},
		{"this.foo = function() {};", false},
		{"this.foo = function bar() {};", false},
		{"typeof function() {}", false},
		{"typeof function foo() {}", false},
		{"(function() {}) + (function foo() {})", false},
		{"typeof function *foo() {}", false},
		{"typeof async function foo() {}", false},
		{"typeof async function *foo() {}", false},
		{"(function() { var foo = 1; })();", false},
		{"(function() { function foo() {} })();", false},
		{"(function() { function *foo() {} })();", false},
		{"(function() { async function foo() {} })();", false},
		{"(function() { async function *foo() {} })();", false},
		{"const foo = 1; let bar; class Baz {}", false},
		{"const foo = 1; let bar; class Baz {}", false},
		{"const Array = 1; let Object; class Math {}", false},
		{"typeof class {}", true},
		{"typeof class foo {}", true},
		{"{ const foo = 1; let bar; class Baz {} }", true},
		{"(function() { const foo = 1; let bar; class Baz {} })();", true},
		{"window.foo = (function() { const bar = 1; let baz; class Quux {} return function () {} })();", true},
		{"const foo = 1;", false},
		{"let foo = 1;", false},
		{"let foo = function() {};", false},
		{"const foo = function() {};", false},
		{"class Foo {}", false},
		{"(function() { let foo = 1; })();", false},
		{"(function() { const foo = 1; })();", false},
		{"foo", false},
		{"foo + bar", false},
		{"foo(bar)", false},
		{"foo++", false},
		{"--foo", false},
		{"foo += 1", false},
		{"foo ||= 1", false},
		{"'use strict';foo = 1;", false},
		{"(function() {'use strict'; foo = 1; })();", false},
		{"{ class Foo { constructor() { bar = 1; } baz() { bar = 1; } } }", false},
		{"Foo.bar = 1;", false},
		{"Utils.foo = 1;", false},
		{"Utils.foo = function() {};", false},
		{"window.foo = 1;", false},
		{"window.foo = function() {};", false},
		{"window.foo = function foo() {};", false},
		{"self.foo = 1;", false},
		{"self.foo = function() {};", false},
		{"++foo", false},
		{"foo--", false},
		{"Array.from = 1;", false},
		{"Object['assign'] = 1;", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoImplicitGlobals,
				implicitGlobalsFile, testCase.sourceText, implicitGlobalsOptions(t, testCase.lexicalBindings)))
			rule_testing.ExpectClean(t, runImplicitGlobalsAsAScript(t, implicitGlobalsScript,
				testCase.sourceText, implicitGlobalsOptions(t, testCase.lexicalBindings)))
		})
	}
}

// The two messages this port declines, pinned rather than left as silence.
//
// Upstream reports assignmentToReadonlyGlobal and redeclarationOfReadonlyGlobal for a name it knows
// to be a read-only global. Both sources are eslint configuration surfaces rather than anything in
// the tree: the globals config key, and the /*global foo:readonly*/ directive. (Under
// typescript-eslint's parser the TypeScript lib's types are read-only globals too, which is the
// surface @typescript-eslint/no-redeclare's builtinGlobals reads since #e1zk9s0. This rule's two
// messages were not revisited there.)
//
// The shape of the decline is an UPGRADE that does not happen, rather than a finding that goes
// missing, and that distinction is worth stating because it is much the smaller gap. Measured
// against the installed rule:
//
//	Array = 1                  upstream assignmentToReadonlyGlobal, here SILENT
//	var Array = 1              upstream redeclarationOfReadonlyGlobal, here globalNonLexicalBinding
//	var Array = 1; Array = 2;  upstream redeclarationOfReadonlyGlobal, here globalNonLexicalBinding
//	var notABuiltin = 1        upstream globalNonLexicalBinding, here the same
//
// The two declaration rows still REPORT here, at the same place, under the message that would be
// correct if the name were not a builtin, so the gap there is only that the finding does not say the
// stronger thing. The assignment row is the one genuinely lost finding: `Array` resolves to the
// standard library's own declaration, so the leak check correctly answers that the name WAS declared
// and stays silent. That is resolution working, and it is exactly why the readonly judgment needs a
// configured-globals surface rather than the checker.
//
// Asserting what this port actually produces is what keeps the gap recorded: a later globals surface
// should turn these rows into the upstream ids, and this test is what will notice. The count
// assertion below is deliberately about the two declaration rows and is skipped where nothing is
// reported at all.
func TestNoImplicitGlobalsDeclinesReadonlyGlobals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		// What upstream reports, recorded so the gap is legible rather than inferred.
		upstreamMessages []string
		// What this port reports, which is the weaker message rather than nothing.
		messages []string
		typed    bool
	}{
		{"Array = 1", []string{"assignmentToReadonlyGlobal"}, nil, true},
		{"var Array = 1", []string{"redeclarationOfReadonlyGlobal"}, []string{"globalNonLexicalBinding"}, false},
		{"var Array = 1; Array = 2;", []string{"redeclarationOfReadonlyGlobal"}, []string{"globalNonLexicalBinding"}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.typed {
				result = runImplicitGlobalsAsAScript(t, implicitGlobalsScript, testCase.sourceText,
					implicitGlobalsOptions(t, true))
			} else {
				result = rule_testing.RunWithOptions(t, NoImplicitGlobals,
					implicitGlobalsFile, testCase.sourceText, implicitGlobalsOptions(t, true))
			}
			rule_testing.ExpectFindings(t, result, testCase.messages...)
			if len(testCase.messages) != 0 &&
				len(testCase.messages) != len(testCase.upstreamMessages) {
				t.Errorf("where this port still reports, the decline should change which message "+
					"is reported and not how many, got %d against upstream's %d",
					len(testCase.messages), len(testCase.upstreamMessages))
			}
		})
	}
}

// What the file is decides each half, measured with controls in each kind of file.
//
// A module has no global scope and is strict, so neither half fires. A TypeScript script is strict
// under alwaysStrict, which every TypeScript file is, so its top-level var is still a global and
// reports while its assignment to an undeclared name throws and leaks nothing (#hks3djf, #jjfa7qb).
// A sloppy JavaScript script reports both. The untyped harness has no checker, so it shows only the
// declaration half.
func TestNoImplicitGlobalsReadsWhatTheFileIs(t *testing.T) {
	t.Parallel()

	run := map[string]func(t *testing.T, sourceText string) rule_testing.Result{
		"the untyped harness": func(t *testing.T, sourceText string) rule_testing.Result {
			return rule_testing.RunWithOptions(t, NoImplicitGlobals, implicitGlobalsFile, sourceText, implicitGlobalsOptions(t, true))
		},
		"a module": func(t *testing.T, sourceText string) rule_testing.Result {
			return rule_testing.RunTypedWithOptions(t, NoImplicitGlobals, implicitGlobalsFile, sourceText, implicitGlobalsOptions(t, true))
		},
		"a TypeScript script under alwaysStrict": func(t *testing.T, sourceText string) rule_testing.Result {
			return runImplicitGlobalsAsAScript(t, implicitGlobalsFile, sourceText, implicitGlobalsOptions(t, true))
		},
		"a sloppy JavaScript script": func(t *testing.T, sourceText string) rule_testing.Result {
			return runImplicitGlobalsAsAScript(t, implicitGlobalsScript, sourceText, implicitGlobalsOptions(t, true))
		},
	}
	cases := []struct {
		file       string
		sourceText string
		messages   []string
	}{
		{"the untyped harness", "var foo = 1;", []string{"globalNonLexicalBinding"}},
		{"the untyped harness", "foo = 1", nil},
		{"a module", "var foo = 1;", nil},
		{"a module", "foo = 1", nil},
		{"a TypeScript script under alwaysStrict", "var foo = 1;", []string{"globalNonLexicalBinding"}},
		{"a TypeScript script under alwaysStrict", "foo = 1", nil},
		{"a TypeScript script under alwaysStrict", "foo = 1; var bar;", []string{"globalNonLexicalBinding"}},
		{"a sloppy JavaScript script", "var foo = 1;", []string{"globalNonLexicalBinding"}},
		{"a sloppy JavaScript script", "foo = 1", []string{"globalVariableLeak"}},
		{"a sloppy JavaScript script", "foo = 1; var bar;", []string{"globalNonLexicalBinding", "globalVariableLeak"}},
		// A JavaScript module is strict by being a module, with no compiler option involved.
		{"a sloppy JavaScript script", "export {}; foo = 1", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.file+": "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, run[testCase.file](t, testCase.sourceText), testCase.messages...)
		})
	}
}

// Strict mode suppresses the leak, and a class body is strict with no directive written.
//
// In strict mode an assignment to an undeclared name throws rather than creating a global, so there
// is nothing to leak. This does not appear in upstream's rule body at all: it falls out of
// eslint-scope declining to record an implicit global in a strict scope, which is why a port reading
// only the rule file misses it and reports ordinary code.
//
// The class row is the one most likely to be missed, since nothing in the source says "strict". It
// was a live false positive in this port before these cases existed. Every expectation measured
// against the installed rule.
func TestNoImplicitGlobalsIsSilentInStrictMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messages   []string
	}{
		{"the control, which does leak", "foo = 1;", []string{"globalVariableLeak"}},
		{"a file level use strict", "'use strict';foo = 1;", nil},
		{"a function level use strict", "(function() {'use strict'; foo = 1; })();", nil},
		{"a class body, strict with no directive",
			"{ class Foo { constructor() { bar = 1; } baz() { bar = 1; } } }", nil},
		{"a directive that is not use strict still leaks", "'use asm';foo = 1;", []string{"globalVariableLeak"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runImplicitGlobalsAsAScript(t, implicitGlobalsScript,
				testCase.sourceText, implicitGlobalsOptions(t, false)), testCase.messages...)
		})
	}
}

// The decoder's own shapes, which no source fixture can reach.
//
// Upstream's default for lexicalBindings is FALSE, so the zero value happens to be right here and
// the usual default-inversion hazard does not apply. It is decoded through a pointer anyway so an
// explicit false stays distinguishable from an absent key, and this test is what would notice if the
// default ever moved and the pointer were dropped as redundant.
func TestDecodeNoImplicitGlobalsOptions(t *testing.T) {
	t.Parallel()

	t.Run("an absent option leaves lexical bindings off", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoImplicitGlobalsOptions(nil)
		if err != nil {
			t.Fatalf("the decoder refused an absent option: %v", err)
		}
		if decoded.(NoImplicitGlobalsSettings).LexicalBindings {
			t.Error("expected lexicalBindings to default to false")
		}
	})

	t.Run("an empty object leaves it off", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoImplicitGlobalsOptions(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("the decoder refused an empty object: %v", err)
		}
		if decoded.(NoImplicitGlobalsSettings).LexicalBindings {
			t.Error("expected lexicalBindings to stay false")
		}
	})

	t.Run("an explicit true turns it on", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoImplicitGlobalsOptions(json.RawMessage(`{"lexicalBindings":true}`))
		if err != nil {
			t.Fatalf("the decoder refused an explicit true: %v", err)
		}
		if !decoded.(NoImplicitGlobalsSettings).LexicalBindings {
			t.Error("expected lexicalBindings to be true")
		}
	})

	t.Run("a bare severity leaves the rule on its default", func(t *testing.T) {
		t.Parallel()
		// A rule configured as "error" is handed nil options, which arrives at Run as an untyped
		// nil rather than as settings. Asserting through the rule covers the fallback inside Run,
		// which no decoder test can reach: a global const must stay clean.
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoImplicitGlobals,
			implicitGlobalsFile, "const a = 1;", nil))
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoImplicitGlobals,
			implicitGlobalsFile, "var a = 1;", nil), "globalNonLexicalBinding")
	})
}

// The finding points at the declared name, not at the whole statement.
//
// Upstream reports the VariableDeclarator and the FunctionDeclaration, so a var reports on
// `foo = 1` rather than on `var foo = 1;`, and every ExpectFindings assertion above is satisfied by
// either reading.
func TestNoImplicitGlobalsPointsAtTheDeclaration(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoImplicitGlobals, implicitGlobalsFile,
		"var foo = 1;", implicitGlobalsOptions(t, false))
	rule_testing.ExpectFindings(t, result, "globalNonLexicalBinding")
	source := result.SourceFile.Text()
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "foo = 1" {
		t.Errorf("expected the finding on the whole declarator, pointed at %q", reported)
	}

	leak := runImplicitGlobalsAsAScript(t, implicitGlobalsScript, "foo = 1", implicitGlobalsOptions(t, false))
	rule_testing.ExpectFindings(t, leak, "globalVariableLeak")
	leakSource := leak.SourceFile.Text()
	leakReported := leakSource[leak.Diagnostics[0].Range.Pos():leak.Diagnostics[0].Range.End()]
	if leakReported != "foo" {
		t.Errorf("expected the leak finding on the assigned name, pointed at %q", leakReported)
	}
}

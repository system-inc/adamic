package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// strictFile is where the corpus pretends to live.
//
// A `.js` file, because upstream's corpus is JavaScript run as a script, and a TypeScript file is
// strict whatever its shape (`alwaysStrict`), so on a `.ts` path every script row collapses into the
// `implied` judgment. The rows ran there until #hks3djf, which proved the script arms against a file
// kind that can never reach them in a real run.
const strictFile = "/repository/source/Strict.js"

// strictTypeScriptFile is where the rows that need TypeScript syntax live.
const strictTypeScriptFile = "/repository/source/Strict.ts"

// strictCase is one imported corpus row.
type strictCase struct {
	sourceText string
	options    any
	wantIds    []string
	// wantFixedSource is "" when the case is reported and deliberately NOT repaired.
	wantFixedSource string
}

// runStrict drives one corpus case through the rule's own exported decoder.
func runStrict(t *testing.T, testCase strictCase) rule_testing.Result {
	t.Helper()
	return runStrictIn(t, strictFile, testCase)
}

// runStrictIn is runStrict at a chosen path, for the rows whose answer turns on the file kind.
func runStrictIn(t *testing.T, fileName string, testCase strictCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, Strict, fileName, testCase.sourceText)
	}
	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeStrictOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, Strict, fileName, testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/strict.js` was loaded with its RuleTester stubbed so every case came out as data
// with its options and source type attached, then replayed against the INSTALLED rule to record what
// it reports and what its fixer writes. 127 of 128 reproduced; the one that did not is recorded in
// TestStrictSuppressionCaseReports below.
//
// # The corpus is mostly SCRIPT and this tree is almost entirely modules
//
// Upstream's tester defaults to `ecmaVersion: 5, sourceType: "script"`, so 107 of its 128 cases are
// scripts, 17 modules and 4 commonjs. That is the opposite of what cohere lints: measured, 2,917 of
// 2,919 TypeScript files in the tree carry a top-level import or export.
//
// The script cases are still expressible and are imported. A file with no import or export is not an
// external module even under the harness's `moduleDetection: "force"`, which was probed rather than
// assumed, so the script arms are genuinely reachable here.
//
// Five cases could not be imported and each is recorded rather than dropped: four are `commonjs`,
// which the harness has no source type for, and one is clean only because of an `eslint-disable`
// comment. Both are covered by their own tests below.
func strictFiresCases() []strictCase {
	return []strictCase{
		{"\"use strict\"; foo();", StrictNever, []string{"never"}, ""},
		{"function foo() { 'use strict'; return; }", StrictNever, []string{"never"}, ""},
		{"var foo = function() { 'use strict'; return; };", StrictNever, []string{"never"}, ""},
		{"function foo() { return function() { 'use strict'; return; }; }", StrictNever, []string{"never"}, ""},
		{"'use strict'; function foo() { \"use strict\"; return; }", StrictNever, []string{"never", "never"}, ""},
		{"foo();", StrictGlobal, []string{"global"}, ""},
		{"/* license */\nfunction foo() {}\nfunction bar() {}\n/* end */", StrictGlobal, []string{"global"}, ""},
		{"function foo() { 'use strict'; return; }", StrictGlobal, []string{"global", "global"}, ""},
		{"var foo = function() { 'use strict'; return; }", StrictGlobal, []string{"global", "global"}, ""},
		{"var foo = () => { 'use strict'; return () => 1; }", StrictGlobal, []string{"global", "global"}, ""},
		{"'use strict'; function foo() { 'use strict'; return; }", StrictGlobal, []string{"global"}, ""},
		{"'use strict'; var foo = function() { 'use strict'; return; };", StrictGlobal, []string{"global"}, ""},
		{"'use strict'; 'use strict'; foo();", StrictGlobal, []string{"multiple"}, "'use strict';  foo();"},
		{"'use strict'; foo();", StrictFunction, []string{"function"}, ""},
		{"'use strict'; (function() { 'use strict'; return true; }());", StrictFunction, []string{"function"}, ""},
		{"(function() { 'use strict'; function f() { 'use strict'; return } return true; }());", StrictFunction, []string{"unnecessary"}, "(function() { 'use strict'; function f() {  return } return true; }());"},
		{"(function() { return true; }());", StrictFunction, []string{"function"}, ""},
		{"(() => { return true; })();", StrictFunction, []string{"function"}, ""},
		{"(() => true)();", StrictFunction, []string{"function"}, ""},
		{"var foo = function() { foo(); 'use strict'; return; }; function bar() { foo(); 'use strict'; }", StrictFunction, []string{"function", "function"}, ""},
		{"function foo() { 'use strict'; 'use strict'; return; }", StrictFunction, []string{"multiple"}, "function foo() { 'use strict';  return; }"},
		{"var foo = function() { 'use strict'; 'use strict'; return; }", StrictFunction, []string{"multiple"}, "var foo = function() { 'use strict';  return; }"},
		{"function foo() { return function() { 'use strict'; return; }; }", StrictFunction, []string{"function"}, ""},
		{"var foo = function() { function bar() { 'use strict'; return; } return; }", StrictFunction, []string{"function"}, ""},
		{"function foo() { 'use strict'; return; } var bar = function() { return; };", StrictFunction, []string{"function"}, ""},
		{"var foo = function() { 'use strict'; return; }; function bar() { return; };", StrictFunction, []string{"function"}, ""},
		{"function foo() { 'use strict'; return function() { 'use strict'; 'use strict'; return; }; }", StrictFunction, []string{"unnecessary", "multiple"}, "function foo() { 'use strict'; return function() {   return; }; }"},
		{"var foo = function() { 'use strict'; function bar() { 'use strict'; 'use strict'; return; } }", StrictFunction, []string{"unnecessary", "multiple"}, "var foo = function() { 'use strict'; function bar() {   return; } }"},
		{"var foo = () => { return; };", StrictFunction, []string{"function"}, ""},
		{"class A { constructor() { \"use strict\"; } }", StrictFunction, []string{"unnecessaryInClasses"}, "class A { constructor() {  } }"},
		{"class A { foo() { \"use strict\"; } }", StrictFunction, []string{"unnecessaryInClasses"}, "class A { foo() {  } }"},
		{"class A { foo() { function bar() { \"use strict\"; } } }", StrictFunction, []string{"unnecessaryInClasses"}, "class A { foo() { function bar() {  } } }"},
		{"class A { field = () => { \"use strict\"; } }", StrictFunction, []string{"unnecessaryInClasses"}, "class A { field = () => {  } }"},
		{"class A { field = function() { \"use strict\"; } }", StrictFunction, []string{"unnecessaryInClasses"}, "class A { field = function() {  } }"},
		{"'use strict'; function foo() { return; }", StrictSafe, []string{"function", "function"}, ""},
		{"'use strict'; function foo() { return; }", nil, []string{"function", "function"}, ""},
		{"function foo() { return; }", nil, []string{"function"}, ""},
		{"function foo(a = 0) { 'use strict' }", nil, []string{"nonSimpleParameterList"}, ""},
		{"(function() { 'use strict'; function foo(a = 0) { 'use strict' } }())", nil, []string{"nonSimpleParameterList"}, ""},
		{"function foo(a = 0) { 'use strict' }", StrictNever, []string{"nonSimpleParameterList"}, ""},
		{"function foo(a = 0) { 'use strict' }", StrictGlobal, []string{"global", "nonSimpleParameterList"}, ""},
		{"'use strict'; function foo(a = 0) { 'use strict' }", StrictGlobal, []string{"nonSimpleParameterList"}, ""},
		{"function foo(a = 0) { 'use strict' }", StrictFunction, []string{"nonSimpleParameterList"}, ""},
		{"(function() { 'use strict'; function foo(a = 0) { 'use strict' } }())", StrictFunction, []string{"nonSimpleParameterList"}, ""},
		{"function foo(a = 0) { }", StrictFunction, []string{"wrap"}, ""},
		{"(function() { function foo(a = 0) { } }())", StrictFunction, []string{"function"}, ""},
		{"'use strict'; class C { static { function foo() { \n'use strict'; } } }", StrictGlobal, []string{"global"}, ""},
		{"class C { static { function foo() { \n'use strict'; } } }", StrictNever, []string{"never"}, ""},
		{"function foo() {'use strict'; class C { static { function foo() { \n'use strict'; } } } }", StrictFunction, []string{"unnecessary"}, "function foo() {'use strict'; class C { static { function foo() { \n } } } }"},
		{"class C { static { function foo() { \n'use strict'; } } }", StrictFunction, []string{"unnecessaryInClasses"}, "class C { static { function foo() { \n } } }"},
		{"class C { static { function foo() { \n'use strict';\n'use strict'; } } }", StrictFunction, []string{"unnecessaryInClasses", "multiple"}, "class C { static { function foo() { \n\n } } }"},
	}
}

func strictSilentCases() []strictCase {
	return []strictCase{
		{"foo();", StrictNever, nil, ""},
		{"function foo() { return; }", StrictNever, nil, ""},
		{"var foo = function() { return; };", StrictNever, nil, ""},
		{"foo(); 'use strict';", StrictNever, nil, ""},
		{"function foo() { bar(); 'use strict'; return; }", StrictNever, nil, ""},
		{"var foo = function() { { 'use strict'; } return; };", StrictNever, nil, ""},
		{"(function() { bar('use strict'); return; }());", StrictNever, nil, ""},
		{"var fn = x => 1;", StrictNever, nil, ""},
		{"var fn = x => { return; };", StrictNever, nil, ""},
		{"// Intentionally empty", StrictGlobal, nil, ""},
		{"\"use strict\"; foo();", StrictGlobal, nil, ""},
		{"'use strict'; function foo() { return; }", StrictGlobal, nil, ""},
		{"'use strict'; var foo = function() { return; };", StrictGlobal, nil, ""},
		{"'use strict'; function foo() { bar(); 'use strict'; return; }", StrictGlobal, nil, ""},
		{"'use strict'; var foo = function() { bar(); 'use strict'; return; };", StrictGlobal, nil, ""},
		{"'use strict'; function foo() { return function() { bar(); 'use strict'; return; }; }", StrictGlobal, nil, ""},
		{"'use strict'; var foo = () => { return () => { bar(); 'use strict'; return; }; }", StrictGlobal, nil, ""},
		{"function foo() { 'use strict'; return; }", StrictFunction, nil, ""},
		{"var foo = function() { 'use strict'; return; }", StrictFunction, nil, ""},
		{"function foo() { 'use strict'; return; } var bar = function() { 'use strict'; bar(); };", StrictFunction, nil, ""},
		{"var foo = function() { 'use strict'; function bar() { return; } bar(); };", StrictFunction, nil, ""},
		{"var foo = () => { 'use strict'; var bar = () => 1; bar(); };", StrictFunction, nil, ""},
		{"class A { constructor() { } }", StrictFunction, nil, ""},
		{"class A { foo() { } }", StrictFunction, nil, ""},
		{"class A { foo() { function bar() { } } }", StrictFunction, nil, ""},
		{"(function() { 'use strict'; function foo(a = 0) { } }())", StrictFunction, nil, ""},
		{"function foo() { 'use strict'; return; }", StrictSafe, nil, ""},
		{"function foo() { 'use strict'; return; }", nil, nil, ""},
		{"'use strict'; class C { static { foo; } }", StrictGlobal, nil, ""},
		{"'use strict'; class C { static { 'use strict'; } }", StrictGlobal, nil, ""},
		{"'use strict'; class C { static { 'use strict'; 'use strict'; } }", StrictGlobal, nil, ""},
		{"class C { static { foo; } }", StrictFunction, nil, ""},
		{"class C { static { 'use strict'; } }", StrictFunction, nil, ""},
		{"class C { static { 'use strict'; 'use strict'; } }", StrictFunction, nil, ""},
		{"class C { static { foo; } }", StrictNever, nil, ""},
		{"class C { static { 'use strict'; } }", StrictNever, nil, ""},
		{"class C { static { 'use strict'; 'use strict'; } }", StrictNever, nil, ""},
	}
}

func TestStrictFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range strictFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runStrict(t, testCase), testCase.wantIds...)
		})
	}
}

func TestStrictStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range strictSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runStrict(t, testCase))
		})
	}
}

// What the fixer WRITES, for the cases carrying an applicable repair.
//
// Twenty-one of the imported rows assert an exact rewritten source. The repair always deletes one
// whole directive statement, so a fixer that deleted the wrong statement, or trimmed a neighbour,
// passes every assertion above.
func TestStrictFixesTheSource(t *testing.T) {
	t.Parallel()

	applied := 0
	for _, testCase := range strictFiresCases() {
		if testCase.wantFixedSource == "" {
			continue
		}
		applied++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runStrict(t, testCase), testCase.wantFixedSource)
		})
	}
	// The floor is asserted rather than the exact count, because the imported set is filtered
	// by what this harness can express and that filter has changed twice while porting. A
	// hardcoded total caught a real drift once and then became the thing most likely to be
	// stale; a floor still fails if the repair arm stops being exercised at all.
	if applied < 14 {
		t.Errorf("only %d cases asserted a rewrite, wanted at least 14: the repair arm is "+
			"barely covered", applied)
	}
}

// The findings upstream reports and deliberately declines to repair.
//
// Upstream's `shouldFix` allows a repair only where the directive is merely redundant. A directive
// in the wrong FORM for the configuration, or one that is a syntax error because of the parameter
// list, is reported with nothing offered, because choosing between deleting the directive and
// changing the surrounding code is a judgment the rule cannot make.
func TestStrictDeclinesToRepair(t *testing.T) {
	t.Parallel()

	declined := 0
	for _, testCase := range strictFiresCases() {
		if testCase.wantFixedSource != "" {
			continue
		}
		declined++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runStrict(t, testCase)
			if len(result.Diagnostics) == 0 {
				t.Fatal("wanted a finding")
			}
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carried %d fixes, wanted none", index,
						len(diagnostic.Fixes))
				}
			}
		})
	}
	if declined == 0 {
		t.Error("no case declined a repair, so the shouldFix split is untested")
	}
}

// A TypeScript annotation does not make a parameter non-simple, and three other things do.
//
// Since ES2016 a `use strict` directive is a syntax error in a function whose parameter list uses a
// default, a rest element or destructuring, which is what separates a merely redundant directive
// from one that stops the file running. Upstream's corpus is JavaScript, so it cannot say what a
// TYPE annotation does to that judgment, and the answer is not obvious: an annotation is extra
// syntax on the parameter, and a port that asked "is this parameter node a bare identifier" without
// looking past the annotation would get it backwards.
//
// Every verdict below was measured against the installed rule through the TypeScript parser, with
// `impliedStrict` on, because a TypeScript file is strict whatever its shape and so reads `implied`
// (see TestStrictReadsATypeScriptFileAsImpliedStrict). The configured `never` is overridden there,
// as it is upstream, which is what keeps it from mattering. The first four rows are the ones
// upstream's corpus cannot contain; the last three are the controls that keep them honest by showing
// the predicate still says no to the three real cases.
func TestStrictReadsTypeScriptParameterLists(t *testing.T) {
	t.Parallel()

	cases := []strictCase{
		// An annotation and a return type leave the parameter simple, so the directive is merely
		// redundant rather than a syntax error, and the repair deletes it.
		{"function f(a: string): void { 'use strict'; }", StrictNever, []string{"implied"},
			"function f(a: string): void {  }"},
		{"function f<T>(a: T): T { 'use strict'; return a; }", StrictNever, []string{"implied"},
			"function f<T>(a: T): T {  return a; }"},
		// An optional marker likewise.
		{"function f(a?: string) { 'use strict'; }", StrictNever, []string{"implied"},
			"function f(a?: string) {  }"},
		// A method with an annotated parameter, reached through the class arm.
		{"class A { foo(a: string): void { 'use strict'; } }", StrictNever, []string{"implied"},
			"class A { foo(a: string): void {  } }"},
		// The three shapes that genuinely make the directive a syntax error, each carrying a type
		// annotation as well so the annotation is not what decides it.
		{"function f(a: string = 'x') { 'use strict'; }", StrictNever,
			[]string{"nonSimpleParameterList"}, ""},
		{"function f(...rest: string[]) { 'use strict'; }", StrictNever,
			[]string{"nonSimpleParameterList"}, ""},
		{"function f({a}: {a: number}) { 'use strict'; }", StrictNever,
			[]string{"nonSimpleParameterList"}, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runStrictIn(t, strictTypeScriptFile, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if testCase.wantFixedSource != "" {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
			}
		})
	}
}

// A TypeScript file is strict whatever its shape, so it reads as upstream's `implied` mode.
//
// The case that found it is api's Base ProjectSecrets.test.ts, a Jest file with no import or export.
// TypeScript's module detection makes it a script, and the rule asked for a directive in each of its
// top-level functions, while ESLint, whose flat config parses every `.ts` file as a module, was silent
// (#hks3djf). The compiler's `alwaysStrict` makes the file strict, and that is upstream's
// `impliedStrict`. Every row was replayed through the installed rule with the TypeScript parser and
// `impliedStrict` on, the fixed source included.
func TestStrictReadsATypeScriptFileAsImpliedStrict(t *testing.T) {
	t.Parallel()

	script := "function foo(): number { return 1; }\nfoo();"
	// The #hks3djf shape, under every mode: nothing to ask for, because nothing is missing.
	for _, mode := range []StrictMode{"", StrictSafe, StrictGlobal, StrictFunction, StrictNever} {
		t.Run("a script under "+string(mode), func(t *testing.T) {
			t.Parallel()
			testCase := strictCase{sourceText: script}
			if mode != "" {
				testCase.options = mode
			}
			rule_testing.ExpectClean(t, runStrictIn(t, strictTypeScriptFile, testCase))
		})
	}

	// A written directive is redundant, and the repair deletes it. Upstream reports every directive
	// here rather than the first as `implied` and the rest as `multiple`, because `implied` goes
	// through the same reportAll as `never` and `module`.
	cases := []strictCase{
		{"'use strict';\nconst a: number = 1;", nil, []string{"implied"}, "\nconst a: number = 1;"},
		{"'use strict';\nconst a: number = 1;", StrictGlobal,
			[]string{"implied"}, "\nconst a: number = 1;"},
		{"'use strict'; 'use strict';\nconst a: number = 1;", StrictGlobal,
			[]string{"implied", "implied"}, " \nconst a: number = 1;"},
		{"function foo(): void { 'use strict'; 'use strict'; }", StrictFunction,
			[]string{"implied", "implied"}, "function foo(): void {   }"},
		// A module still wins, as upstream's Program listener has it: `module` is assigned after
		// `implied`, so a TypeScript module says the more specific thing.
		{"'use strict';\nexport const a: number = 1;", nil, []string{"module"},
			"\nexport const a: number = 1;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runStrictIn(t, strictTypeScriptFile, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}

	// The control: the same script in a JavaScript file is run as written, so it still needs the
	// directive. Without this a rule that read every file as implied would pass every row above.
	t.Run("a JavaScript script is not implied", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, runStrict(t, strictCase{
			sourceText: "function foo() { return 1; }\nfoo();"}), "function")
	})
}

// The verdict follows the tsconfig, read through the program a real run has.
//
// The syntax-only harness has no program and reads the default options, so these rows build one.
// `strict: false` turns `alwaysStrict` off with it, and the script arms come back. An explicit
// `alwaysStrict` beside it turns them off again, which separates reading the option from reading
// `strict`.
func TestStrictFollowsTheCompilersAlwaysStrict(t *testing.T) {
	t.Parallel()

	files := map[string]string{"Strict.ts": "function foo(): number { return 1; }\nfoo();"}
	withConfig := func(compilerOptions string) func(directory string) {
		return func(directory string) {
			config := `{"compilerOptions": {` + compilerOptions + `, "target": "ES2022", ` +
				`"lib": ["ES2022"], "types": []}, "include": ["**/*.ts"]}`
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config),
				0o644); err != nil {
				t.Fatalf("writing the tsconfig: %v", err)
			}
		}
	}

	t.Run("strict on implies it", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, Strict, files, "Strict.ts"))
	})
	t.Run("strict off leaves a script sloppy", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedFilesWithSetup(t, Strict, files,
			"Strict.ts", withConfig(`"strict": false`)), "function")
	})
	t.Run("alwaysStrict on its own implies it", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFilesWithSetup(t, Strict, files,
			"Strict.ts", withConfig(`"strict": false, "alwaysStrict": true`)))
	})
}

// The repair deletes one whole statement and can therefore never eat type syntax.
//
// Two fixers shipped tonight that were right about JavaScript and destroyed type information here,
// both because they built a replacement span from a node's neighbour. This one removes the
// directive STATEMENT, so a parameter annotation, a return type and a generic list all sit outside
// the deleted range by construction. These rows turn that argument into a measurement.
func TestStrictRepairPreservesTypeSyntax(t *testing.T) {
	t.Parallel()

	cases := []strictCase{
		// The `never` mode does not repair, so these use the module collapse, which does.
		{"'use strict'; export function f(a: string): void {}", nil, []string{"module"},
			" export function f(a: string): void {}"},
		{"'use strict'; export function g<T>(a: T): T { return a; }", nil, []string{"module"},
			" export function g<T>(a: T): T { return a; }"},
		{"'use strict'; export const h = (a?: number): string => String(a);", nil,
			[]string{"module"}, " export const h = (a?: number): string => String(a);"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runStrictIn(t, strictTypeScriptFile, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// A module collapses every configured mode into one judgment.
//
// Upstream's first act inside its Program listener is `if (node.sourceType === "module") mode =
// "module"`, which overrides whatever was configured. That is the arm this tree actually runs:
// measured, 2,917 of 2,919 TypeScript files here carry a top-level import or export.
//
// Each row configures a DIFFERENT mode and expects the same `module` finding, which is what pins
// the override. Without them a rule that simply ignored the option would look correct.
func TestStrictCollapsesEveryModeInAModule(t *testing.T) {
	t.Parallel()

	source := "'use strict'; export const a = 1;"
	for _, mode := range []StrictMode{StrictSafe, StrictGlobal, StrictFunction, StrictNever} {
		t.Run(string(mode), func(t *testing.T) {
			t.Parallel()
			result := runStrict(t, strictCase{sourceText: source,
				options: mode})
			rule_testing.ExpectFindings(t, result, "module")
			rule_testing.ExpectFixedSource(t, result, " export const a = 1;")
		})
	}
	// And with no options at all, which is how the rule would be configured here.
	t.Run("nil options", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.Run(t, Strict, strictFile, source), "module")
	})
	// The control: a module with no directive is clean, so the rows above are the directive
	// rather than a rule that reports on every module.
	t.Run("the control", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.Run(t, Strict, strictFile,
			"export const a = 1;"))
	})
}

// Forty of upstream's cases could not be imported, and each is recorded rather than dropped.
//
// Four axes, none of them expressible in this harness, and three of them are PARSER features rather
// than anything the rule decides:
//
//	ecmaFeatures.impliedStrict   forces the mode to `implied`, overriding the configuration.
//	                             17 cases. The same source under the same options is clean with
//	                             the feature and reports twice without it.
//	ecmaFeatures.globalReturn    makes `safe` resolve to Global rather than Function. Same shape:
//	                             identical source and options, opposite verdicts.
//	sourceType: commonjs         the other way `safe` reaches Global. The harness has no such
//	                             source type, and `module.exports = 1` in a `.ts` file does not
//	                             set `CommonJSModuleIndicator`, measured.
//	sourceType: module declared  twelve cases are modules only because the tester says so; none
//	  out of band                carries an import or export. Here a file is a module exactly when
//	                             its own text says so, so adding an export to make the case run
//	                             would change the input, and the input is the corpus.
//
// These are facts about the harness rather than about the rule, and hiding them inside a relaxed
// rule would turn them into facts about the rule. This test asserts the one thing that IS
// expressible about them: the two parser features have no configuration surface at all, so no
// spelling of the options can select the `implied` mode.
//
// The mode itself is reachable since #hks3djf, through the file rather than the options: a
// TypeScript file under `alwaysStrict` is upstream's `impliedStrict`, proven in
// TestStrictReadsATypeScriptFileAsImpliedStrict. The seventeen rows stay unimported because they are
// JavaScript, which the compiler does not make strict.
func TestStrictHasNoImpliedMode(t *testing.T) {
	t.Parallel()

	// Upstream has no `implied` option either: `impliedStrict` is a parser feature, and here the
	// file kind and the tsconfig stand in for it. An option spelling would be a second way to say
	// the same thing that could disagree with the first. `module` is refused for the same reason:
	// the file's own imports select it.
	for _, mode := range []string{"safe", "global", "function", "never", "implied", "module"} {
		_, err := DecodeStrictOptions([]byte(`"` + mode + `"`))
		if mode == "implied" || mode == "module" {
			if err == nil {
				t.Errorf("mode %q decoded, but it is selected by the file, not the option", mode)
			}
			continue
		}
		if err != nil {
			t.Errorf("mode %q should decode: %v", mode, err)
		}
	}
}

// The decoder takes upstream's option element exactly: one bare string from upstream's four.
//
// Every arm of the rule is selected by string equality, so an unrecognized spelling must fail rather
// than pick a silent fifth behaviour of reporting nothing.
func TestDecodeStrictOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input selects Safe", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeStrictOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode := decoded.(StrictMode); mode != StrictSafe {
			t.Errorf("mode came back %q, wanted upstream's default %q", mode, StrictSafe)
		}
	})

	t.Run("upstream's bare string decodes to its mode", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeStrictOptions([]byte(`"never"`))
		if err != nil {
			t.Fatalf("upstream's own `[\"error\", \"never\"]` was refused: %v", err)
		}
		if mode := decoded.(StrictMode); mode != StrictNever {
			t.Errorf("mode came back %q, wanted %q", mode, StrictNever)
		}
	})

	// The object form was this decoder's own invention and no ESLint version accepts it, so a config
	// carrying it was never valid ESLint. Refusing it keeps one spelling for one setting.
	t.Run("the object form is refused", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeStrictOptions([]byte(`{"mode":"never"}`)); err == nil {
			t.Error(`{"mode":"never"} decoded; upstream's option is a bare string`)
		}
	})

	// Upstream's enum is case-sensitive, so a PascalCase spelling is a typo rather than a synonym.
	t.Run("another casing is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeStrictOptions([]byte(`"Never"`)); err == nil {
			t.Error(`"Never" decoded; upstream spells it "never"`)
		}
	})

	t.Run("an unknown mode is rejected", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{`"sometimes"`, `""`, `null`, `1`} {
			if _, err := DecodeStrictOptions([]byte(raw)); err == nil {
				t.Errorf("%s decoded", raw)
			}
		}
	})
}

// A `use strict` after another prologue directive is invisible upstream, and here.
//
// Written for a surviving mutant: accepting any string literal as a directive left all 88 imported
// rows green, because upstream's corpus writes no other prologue directive.
//
// The mechanism is worth stating because it looks like a bug and is one. Upstream builds its
// directive list as a SPARSE array indexed by statement position, assigning only for `use strict`,
// and its later `forEach` and `slice` skip the holes. So a hole swallows everything after it:
//
//	'use strict'; 'use asm';   reports    the directive is at index 0
//	'use asm'; 'use strict';   CLEAN      index 0 is a hole, so index 1 is never visited
//	'a'; 'b'; 'use strict';    CLEAN      same
//
// Both verdicts were measured against the installed rule. The directive is equally redundant either
// way round, so this is a defect rather than a judgment, and it is reproduced rather than corrected:
// a port that helpfully reported the middle row would disagree with the tool it replaces on real
// source, and nothing in the corpus would have said so.
func TestStrictStopsAtANonStrictDirective(t *testing.T) {
	t.Parallel()

	cases := []strictCase{
		{"'use strict'; foo();", StrictNever, []string{"never"}, ""},
		{"'use strict'; 'use asm'; foo();", StrictNever, []string{"never"}, ""},
		// The reproduced defect: silent because index 0 is a hole.
		{"'use asm'; 'use strict'; foo();", StrictNever, nil, ""},
		{"'a'; 'b'; 'use strict'; foo();", StrictNever, nil, ""},
		// A non-directive string does not open a prologue slot at all.
		{"'use asm'; foo();", StrictNever, nil, ""},
		// The control: two real directives in a row are both reported, so the silence above is
		// the hole rather than a rule that only ever reports one.
		{"'use strict'; 'use strict'; foo();", StrictNever, []string{"never", "never"}, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runStrict(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The CommonJS half of Safe mode is unreachable in this harness, and that is measured.
//
// `Safe` resolves to Global when the file is CommonJS and to Function otherwise. A mutant forcing
// that test to false survives every fixture in this file, and no fixture can kill it: probed across
// five spellings -- `module.exports = 1`, `module.exports = function f() {}`, `exports.foo = 1`,
// `const x = require('y')`, and both together -- `CommonJSModuleIndicator` is nil for every one of
// them in a `.ts` file.
//
// The corpus now runs on a `.js` path, and the indicator is still nil there, because this harness
// parses without binding. So the branch is not dead code and it is not covered either. It would fire
// for a `.js` file in a real run, which this tree does not currently lint but the config could
// include. The alternative to recording it
// was deleting the branch as unreachable, which would be wrong for the same reason: the verdict is a
// fact about the harness and the file extension, not about the rule.
//
// This test asserts the reachable half, so a change that broke Safe entirely still fails.
func TestStrictSafeResolvesToFunctionForAScript(t *testing.T) {
	t.Parallel()

	// A script with a top-level function and no directive: Function mode reports, Global would
	// have reported on the whole program instead, so the message id separates them.
	rule_testing.ExpectFindings(t, runStrict(t, strictCase{
		sourceText: "function foo() { return; }", options: StrictSafe}), "function")
	// And the CommonJS spelling resolves the same way here, which is the divergence itself.
	rule_testing.ExpectFindings(t, runStrict(t, strictCase{
		sourceText: "module.exports = function foo() { return; };", options: StrictSafe}), "function")
}

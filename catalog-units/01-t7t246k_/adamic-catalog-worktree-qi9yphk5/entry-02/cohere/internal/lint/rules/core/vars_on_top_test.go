package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, 34 valid and 27 invalid cases, extracted by loading its test file
// with a stubbed rule tester so nothing was retyped, then verified byte against byte.

func TestVarsOnTopStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "valid0", source: "var first = 0;\nfunction foo() {\n    first = 2;\n}"},
		{name: "valid1", source: "function foo() {\n}"},
		{name: "valid2", source: "function foo() {\n   var first;\n   if (true) {\n       first = true;\n   } else {\n       first = 1;\n   }\n}"},
		{name: "valid3", source: "function foo() {\n   var first;\n   var second = 1;\n   var third;\n   var fourth = 1, fifth, sixth = third;\n   var seventh;\n   if (true) {\n       third = true;\n   }\n   first = second;\n}"},
		{name: "valid4", source: "function foo() {\n   var i;\n   for (i = 0; i < 10; i++) {\n       alert(i);\n   }\n}"},
		{name: "valid5", source: "function foo() {\n   var outer;\n   function inner() {\n       var inner = 1;\n       var outer = inner;\n   }\n   outer = 1;\n}"},
		{name: "valid6", source: "function foo() {\n   var first;\n   //Hello\n   var second = 1;\n   first = second;\n}"},
		{name: "valid7", source: "function foo() {\n   var first;\n   /*\n       Hello Clarice\n   */\n   var second = 1;\n   first = second;\n}"},
		{name: "valid8", source: "function foo() {\n   var first;\n   var second = 1;\n   function bar(){\n       var first;\n       first = 5;\n   }\n   first = second;\n}"},
		{name: "valid9", source: "function foo() {\n   var first;\n   var second = 1;\n   function bar(){\n       var third;\n       third = 5;\n   }\n   first = second;\n}"},
		{name: "valid10", source: "function foo() {\n   var first;\n   var bar = function(){\n       var third;\n       third = 5;\n   }\n   first = 5;\n}"},
		{name: "valid11", source: "function foo() {\n   var first;\n   first.onclick(function(){\n       var third;\n       third = 5;\n   });\n   first = 5;\n}"},
		{name: "valid12", source: "function foo() {\n   var i = 0;\n   for (let j = 0; j < 10; j++) {\n       alert(j);\n   }\n   i = i + 1;\n}"},
		{name: "valid13", source: "'use strict'; var x; f();"},
		{name: "valid14", source: "'use strict'; 'directive'; var x; var y; f();"},
		{name: "valid15", source: "function f() { 'use strict'; var x; f(); }"},
		{name: "valid16", source: "function f() { 'use strict'; 'directive'; var x; var y; f(); }"},
		{name: "valid17", source: "import React from 'react'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid18", source: "'use strict'; import React from 'react'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid19", source: "import React from 'react'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid20", source: "import * as foo from 'mod.js'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid21", source: "import { square, diag } from 'lib'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid22", source: "import { default as foo } from 'lib'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid23", source: "import 'src/mylib'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid24", source: "import theDefault, { named1, named2 } from 'src/mylib'; 'use strict'; var y; function f() { 'use strict'; var x; var y; f(); }"},
		{name: "valid25", source: "export var x;\nvar y;\nvar z;"},
		{name: "valid26", source: "var x;\nexport var y;\nvar z;"},
		{name: "valid27", source: "var x;\nvar y;\nexport var z;"},
		{name: "valid28", source: "class C {\n    static {\n        var x;\n    }\n}"},
		{name: "valid29", source: "class C {\n    static {\n        var x;\n        foo();\n    }\n}"},
		{name: "valid30", source: "class C {\n    static {\n        var x;\n        var y;\n    }\n}"},
		{name: "valid31", source: "class C {\n    static {\n        var x;\n        var y;\n        foo();\n    }\n}"},
		{name: "valid32", source: "class C {\n    static {\n        let x;\n        var y;\n    }\n}"},
		{name: "valid33", source: "class C {\n    static {\n        foo();\n        let x;\n    }\n}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts", testCase.source))
		})
	}
}

func TestVarsOnTopFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantCount int
	}{
		{name: "invalid0", source: "var first = 0;\nfunction foo() {\n    first = 2;\n    second = 2;\n}\nvar second = 0;", wantCount: 1},
		{name: "invalid1", source: "function foo() {\n   var first;\n   first = 1;\n   first = 2;\n   first = 3;\n   first = 4;\n   var second = 1;\n   second = 2;\n   first = second;\n}", wantCount: 1},
		{name: "invalid2", source: "function foo() {\n   var first;\n   if (true) {\n       var second = true;\n   }\n   first = second;\n}", wantCount: 1},
		{name: "invalid3", source: "function foo() {\n   for (var i = 0; i < 10; i++) {\n       alert(i);\n   }\n}", wantCount: 1},
		{name: "invalid4", source: "function foo() {\n   var first = 10;\n   var i;\n   for (i = 0; i < first; i ++) {\n       var second = i;\n   }\n}", wantCount: 1},
		{name: "invalid5", source: "function foo() {\n   var first = 10;\n   var i;\n   switch (first) {\n       case 10:\n           var hello = 1;\n           break;\n   }\n}", wantCount: 1},
		{name: "invalid6", source: "function foo() {\n   var first = 10;\n   var i;\n   try {\n       var hello = 1;\n   } catch (e) {\n       alert('error');\n   }\n}", wantCount: 1},
		{name: "invalid7", source: "function foo() {\n   var first = 10;\n   var i;\n   try {\n       asdf;\n   } catch (e) {\n       var hello = 1;\n   }\n}", wantCount: 1},
		{name: "invalid8", source: "function foo() {\n   var first = 10;\n   while (first) {\n       var hello = 1;\n   }\n}", wantCount: 1},
		{name: "invalid9", source: "function foo() {\n   var first = 10;\n   do {\n       var hello = 1;\n   } while (first == 10);\n}", wantCount: 1},
		{name: "invalid10", source: "function foo() {\n   var first = [1,2,3];\n   for (var item in first) {\n       item++;\n   }\n}", wantCount: 1},
		{name: "invalid11", source: "function foo() {\n   var first = [1,2,3];\n   var item;\n   for (item in first) {\n       var hello = item;\n   }\n}", wantCount: 1},
		{name: "invalid12", source: "var foo = () => {\n   var first = [1,2,3];\n   var item;\n   for (item in first) {\n       var hello = item;\n   }\n}", wantCount: 1},
		{name: "invalid13", source: "'use strict'; 0; var x; f();", wantCount: 1},
		{name: "invalid14", source: "'use strict'; var x; 'directive'; var y; f();", wantCount: 1},
		{name: "invalid15", source: "function f() { 'use strict'; 0; var x; f(); }", wantCount: 1},
		{name: "invalid16", source: "function f() { 'use strict'; var x; 'directive';  var y; f(); }", wantCount: 1},
		{name: "invalid17", source: "export function f() {}\nvar x;", wantCount: 1},
		{name: "invalid18", source: "var x;\nexport function f() {}\nvar y;", wantCount: 1},
		{name: "invalid19", source: "import {foo} from 'foo';\nexport {foo};\nvar test = 1;", wantCount: 1},
		{name: "invalid20", source: "export {foo} from 'foo';\nvar test = 1;", wantCount: 1},
		{name: "invalid21", source: "export * from 'foo';\nvar test = 1;", wantCount: 1},
		{name: "invalid22", source: "class C {\n    static {\n        foo();\n        var x;\n    }\n}", wantCount: 1},
		{name: "invalid23", source: "class C {\n    static {\n        'use strict';\n        var x;\n    }\n}", wantCount: 1},
		{name: "invalid24", source: "class C {\n    static {\n        var x;\n        foo();\n        var y;\n    }\n}", wantCount: 1},
		{name: "invalid25", source: "class C {\n    static {\n        if (foo) {\n            var x;\n        }\n    }\n}", wantCount: 1},
		{name: "invalid26", source: "class C {\n    static {\n        if (foo)\n            var x;\n    }\n}", wantCount: 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VarsOnTop, "file.ts", testCase.source)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "top"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// The span and the message, neither of which ExpectFindings can see. Upstream reports the whole
// variable declaration statement rather than the declarator or the keyword.
func TestVarsOnTopSpanAndMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantText string
	}{
		{"insideAnIf", "function foo() { var a; if (x) { var second = true; } }", "var second = true;"},
		{"afterAStatement", "function foo() { var a; a = 1; var b = 1; }", "var b = 1;"},
		{"inAForInitializer", "function foo() { for (var i = 0; i < 10; i++) { } }", "var i = 0"},
		{"exportedAfterAFunction", "export function f() {}\nvar x;", "var x;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VarsOnTop, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			if got := testCase.source[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
			if finding.Message.Id != "top" {
				t.Errorf("message id: got %q", finding.Message.Id)
			}
			wantPrefix := "All 'var' declarations must be at the top of the function scope. "
			if got := finding.Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message:\n got %q\nwant prefix %q", got, wantPrefix)
			}
		})
	}
}

// The directive prologue is a PREFIX scan rather than a filter, and the two answers differ on an
// input upstream writes in both lists.
//
// Upstream's own pair is `'use strict'; 'directive'; var x; var y; f();` clean against
// `'use strict'; var x; 'directive'; var y; f();` reporting. A filter that removed every
// directive-looking statement wherever it sat would call the second one clean too.
func TestVarsOnTopDirectiveScanIsAPrefix(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"'use strict'; 'directive'; var x; var y; f();"))
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"'use strict'; var x; 'directive'; var y; f();"), "top")

	// And a non-string statement closes the prologue immediately.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"'use strict'; 0; var x; f();"), "top")
}

// A class static block has no directive prologue and no imports, so upstream turns the skip off
// there and a string statement is an ordinary expression. This is the only place the two container
// kinds behave differently, and unifying them by accident would make the reporting case clean.
func TestVarsOnTopStaticBlockHasNoPrologue(t *testing.T) {
	t.Parallel()

	// The same shape that is CLEAN at the top of a program.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"class C { static { 'use strict'; var x; } }"), "top")
	rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts", "'use strict'; var x;"))

	// And a var really is allowed at the top of a static block, so the verdict above is about the
	// prologue rather than about static blocks being rejected outright.
	rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"class C { static { var x; foo(); } }"))
}

// `export var` needs no wrapper handling here because our parser makes the export a modifier
// rather than a node. These three are upstream's exported cases and they pin that the run scan
// still sees the declaration.
func TestVarsOnTopExportedDeclarations(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"export var x;\nvar y;\nvar z;",
		"var x;\nexport var y;\nvar z;",
		"var x;\nvar y;\nexport var z;",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts", source))
	}

	// An exported FUNCTION is not a declaration and closes the run, which is the discrimination
	// the three clean cases above cannot make on their own.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"export function f() {}\nvar x;"), "top")
}

// Only `var`. Everything else is scoped to its block already, so the rule has nothing to say about
// it wherever it is written.
func TestVarsOnTopIgnoresBlockScopedDeclarations(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"function foo() { foo(); let x = 1; }",
		"function foo() { foo(); const x = 1; }",
		"function foo() { if (a) { let x = 1; } }",
		"class C { static { foo(); let x; } }",
	} {
		rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts", source))
	}

	// A control, so four clean verdicts are not vacuous.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"function foo() { foo(); var x = 1; }"), "top")
}

// A template literal is not a directive, which upstream gets from its grammar and this port has to
// state. Found by a surviving mutant and then measured against the installed build at 10.8.1.
//
// The tempting reading is that a template with no substitutions is a string and should count. It
// does not, in the rule or in any engine, so accepting it would silently make a reporting input
// clean. The pair below differs only in the quote character.
func TestVarsOnTopTemplateIsNotADirective(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, VarsOnTop, "file.ts", `"use strict"; var x; f();`))
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"`use strict`; var x; f();"), "top")
	rule_testing.ExpectFindings(t, rule_testing.Run(t, VarsOnTop, "file.ts",
		"`hello`; var x;"), "top")
}

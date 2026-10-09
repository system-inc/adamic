package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// shadowRestrictedNamesFile is where the fixtures pretend to live.
//
// A real path matters here because this rule reads the checker for the `undefined` carve-out, so
// the harness builds an actual program and the file has to sit somewhere a tsconfig can reach.
const shadowRestrictedNamesFile = "/repository/source/ShadowRestrictedNames.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two tests below is verbatim from
// `oxc/crates/oxc_linter/src/rules/eslint/no_shadow_restricted_names.rs`: one Tester block, 19 pass
// and 24 fail. The snapshot records 69 diagnostics from those 24 inputs, so one finding per input
// is badly wrong here and the extractor said so before any code was written. The counts below were
// recovered per input from the source line each snapshot diagnostic prints, rather than by walking
// the snapshot in order: upstream lists the twelve-diagnostic `globalThis` case twice
// byte-identically, and an in-order walk would hand all twelve to the first copy.
//
// The dense cases are dense for a reason worth stating. `function NaN(NaN) { var NaN; !function
// NaN(NaN) { try {} catch(NaN) {} }; }` is six separate bindings named `NaN` in five scopes, and it
// reports six times. That single input is most of this rule's declaration-site coverage, which is
// why a port that finds only variable declarations still passes a suite asserting one finding per
// input and fails this one loudly.
func TestNoShadowRestrictedNamesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		// Six bindings: the function name, its parameter, the `var`, the inner function
		// expression's name, its parameter, and the catch parameter.
		{"every declaration form for NaN in one input",
			"function NaN(NaN) { var NaN; !function NaN(NaN) { try {} catch(NaN) {} }; }", 6},
		// Five, not six: this input has no `var undefined`, because a bare `var undefined` would
		// be the carved-out clean form and upstream kept the case honest by omitting it.
		{"every declaration form for undefined in one input",
			"function undefined(undefined) { !function undefined(undefined) { try {} catch(undefined) {} }; }", 5},
		{"every declaration form for Infinity in one input",
			"function Infinity(Infinity) { var Infinity; !function Infinity(Infinity) { try {} catch(Infinity) {} }; }", 6},
		{"every declaration form for arguments in one input",
			"function arguments(arguments) { var arguments; !function arguments(arguments) { try {} catch(arguments) {} }; }", 6},
		{"every declaration form for eval in one input",
			"function eval(eval) { var eval; !function eval(eval) { try {} catch(eval) {} }; }", 6},
		// An arrow parameter rather than a function parameter, and a `var eval` initialized with
		// the arrow, so the variable declaration is not the carved-out shape either.
		{"an arrow parameter and an initialized var named eval",
			"var eval = (eval) => { var eval; !function eval(eval) { try {} catch(eval) {} }; }", 6},
		// An array binding pattern. `var undefined` inside a pattern is not a bare
		// `VariableDeclarator` with no init, so the carve-out does not reach it.
		{"an array destructuring binding named undefined", "var [undefined] = [1]", 1},
		// Four object-pattern shapes in one input: shorthand, renamed value, nested, and rest.
		{"four object destructuring bindings named undefined",
			"var {undefined} = obj; var {a: undefined} = obj; var {a: {b: {undefined}}} = obj; var {a, ...undefined} = obj;", 4},
		// The carve-out's own boundary. The declaration is bare, so only the later write makes it
		// report, and it reports once at the declaration rather than at the write.
		{"a bare var undefined that is later assigned", "var undefined; undefined = 5;", 1},
		{"a default import named undefined", "import undefined from 'foo';", 1},
		{"a named import named undefined", "import { undefined } from 'foo';", 1},
		{"an import renamed to undefined", "import { baz as undefined } from 'foo';", 1},
		{"a namespace import named undefined", "import * as undefined from 'foo';", 1},
		// globalThis, which reports by default. Upstream lists this input twice byte-identically,
		// once for ecmaVersion 2015 and once for 2020; the two are the same input to us.
		{"every declaration form for globalThis in one input",
			"function globalThis(globalThis) { var globalThis; !function globalThis(globalThis) { try {} catch(globalThis) {} }; }", 6},
		{"an array destructuring binding named globalThis", "const [globalThis] = [1]", 1},
		{"four object destructuring bindings named globalThis",
			"var {globalThis} = obj; var {a: globalThis} = obj; var {a: {b: {globalThis}}} = obj; var {a, ...globalThis} = obj;", 4},
		// No carve-out for globalThis: a bare `let globalThis` reports on its own, and this input
		// reports once rather than twice even though it also writes.
		{"a bare let globalThis that is later assigned", "let globalThis; globalThis = 5;", 1},
		{"a class declaration named globalThis", "class globalThis {}", 1},
		{"a class expression named globalThis", "(class globalThis {})", 1},
		{"a default import named globalThis", "import globalThis from 'foo';", 1},
		{"a named import named globalThis", "import { globalThis } from 'foo';", 1},
		{"an import renamed to globalThis", "import { baz as globalThis } from 'foo';", 1},
		{"a namespace import named globalThis", "import * as globalThis from 'foo';", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "shadowingRestrictedName"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, testCase.sourceText),
				wantIds...)
		})
	}
}

// The clean cases, verbatim from the same block.
//
// Most of them are the `undefined` carve-out, which is the only judgment in this rule that is not a
// name test. `var undefined;` is legal and harmless because the binding keeps the global's value,
// and it stops being harmless the moment anything writes to it. Distinguishing the two needs to
// know which declaration a later `undefined = 5` binds to, which is why this rule reads the checker.
func TestNoShadowRestrictedNamesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"ordinary names in every declaration position", "function foo(bar){ var baz; }"},
		{"ordinary names in a function expression", "!function foo(bar){ var baz; }"},
		{"ordinary names in an anonymous function expression", "!function(bar){ var baz; }"},
		{"an ordinary catch parameter", "try {} catch(e) {}"},
		{"a default export with no name", "export default function() {}"},
		{"a catch clause with no parameter", "try {} catch {}"},
		// The carve-out itself, in its four upstream shapes.
		{"a bare var undefined", "var undefined;"},
		{"a bare var undefined that is only read", "var undefined; doSomething(undefined);"},
		{"a bare var undefined declared twice", "var undefined; var undefined;"},
		{"a bare let undefined", "let undefined"},
		{"an import renamed away from undefined", "import { undefined as undef } from 'foo';"},
		// globalThis with reporting turned off lives in the options test below rather than here,
		// because these cases run with no options at all.
		{"a read of globalThis as a property base", "globalThis.foo"},
		{"a read of globalThis into a variable", "const foo = globalThis"},
		{"a read of globalThis from inside a function", "function foo() { return globalThis; }"},
		{"an import renamed away from globalThis", "import { globalThis as foo } from 'bar'"},
		// Enum members are properties of the enum object rather than bindings in any scope, so
		// none of these shadows anything. Upstream carves them out explicitly.
		{"restricted names as enum members",
			"export enum Globals { undefined = 'undefined', NaN = 'nan', Infinity = 'infinity', eval = 'eval', arguments = 'arguments' }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, testCase.sourceText))
		})
	}
}

// The three upstream cases that configure `reportGlobalThis: false`.
//
// Our option is spelled the other way round, `AllowGlobalThis`, so that the Go zero value is
// upstream's default of reporting it. A struct field named `ReportGlobalThis` would default to
// false and silently turn off a third of the rule for every caller who passed no options.
func TestNoShadowRestrictedNamesRespectsAllowGlobalThis(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a let named globalThis", "let globalThis;"},
		{"a class named globalThis", "class globalThis {}"},
		{"an import renamed to globalThis", "import { baz as globalThis } from 'foo';"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoShadowRestrictedNames,
				shadowRestrictedNamesFile, testCase.sourceText,
				NoShadowRestrictedNamesOptions{AllowGlobalThis: true}))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a port pointing at the whole
// declaration rather than at the name passes every fixture above while being wrong about the only
// thing a reader looks at. Upstream underlines exactly the binding identifier in all 69 snapshot
// diagnostics and this asserts the same.
//
// Sliced out of the source with the finding's own range rather than compared against an offset the
// test computed, since an offset computed by the test is wrong in the same direction as the code
// that produced it. That distinction is what makes the `class globalThis {}` case worth listing:
// reporting the class declaration and reporting its name both produce one finding with the id
// `shadowingRestrictedName`, and only the slice tells them apart.
func TestNoShadowRestrictedNamesPointsAtTheName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a function declaration and its parameter", "function NaN(NaN) {}",
			[]string{"NaN", "NaN"}},
		{"a class declaration", "class globalThis {}", []string{"globalThis"}},
		{"a class expression", "(class globalThis {})", []string{"globalThis"}},
		{"a catch parameter", "try {} catch(eval) {}", []string{"eval"}},
		{"a namespace import", "import * as undefined from 'foo';", []string{"undefined"}},
		{"a renamed import binding", "import { baz as undefined } from 'foo';",
			[]string{"undefined"}},
		// The write is at column 16 and the declaration at column 5. Upstream points at the
		// declaration, which is the counterintuitive half: the write is what makes this report but
		// the declaration is what has to be renamed.
		{"a bare var undefined made unsafe by a later write", "var undefined; undefined = 5;",
			[]string{"undefined"}},
		{"a rest element in an object pattern", "var {a, ...arguments} = obj;",
			[]string{"arguments"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoShadowRestrictedNames,
				shadowRestrictedNamesFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.want))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.want[index] {
					t.Errorf("finding %d points at %q, want %q", index, reported, testCase.want[index])
				}
			}
		})
	}
}

// Cases upstream does not cover, each here for a reason stated at the line.
//
// Upstream's corpus was written against a semantic layer that already answered "does any reference
// write to this symbol", so it never had to exercise the shapes a write can take. Ours computes
// that answer, so every shape is a place the carve-out could be silently too generous, and a
// too-generous carve-out is the quiet direction: the rule simply stops reporting.
func TestNoShadowRestrictedNamesCarveOutBoundary(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		// The write shapes. Each of these makes an otherwise-bare `var undefined` unsafe, and each
		// reaches `reference.WritesToBinding` through a different arm.
		{"a compound assignment to the binding", "var undefined; undefined += 1;", 1},
		{"a logical assignment to the binding", "var undefined; undefined ??= 1;", 1},
		{"a postfix increment of the binding", "var undefined; undefined++;", 1},
		{"a prefix decrement of the binding", "var undefined; --undefined;", 1},
		{"a for-of head assigning to the binding", "var undefined; for (undefined of []) {}", 1},
		{"an array destructuring assignment to the binding",
			"var undefined; [undefined] = [1];", 1},
		// The shorthand accessor, which brief step 8b names and which our checker confirmed live:
		// `GetSymbolAtLocation` here returns the shorthand property's own symbol, so without
		// `GetShorthandAssignmentValueSymbol` the carve-out holds and this input goes silent.
		{"a shorthand destructuring assignment to the binding",
			"var undefined; ({undefined} = obj);", 1},
		// A write from inside a nested function still binds to the outer declaration, so the
		// carve-out has to search the whole file rather than the declaration's own statement.
		{"a write from inside a nested function",
			"var undefined; function f() { undefined = 1; }", 1},
		// A `var` hoists, so a write textually before the declaration binds to it. A carve-out
		// searching only forward from the declaration would miss this.
		{"a write textually before the declaration", "undefined = 1; var undefined;", 1},
		// An initializer alone makes it unsafe with no write anywhere.
		{"a var undefined with an initializer", "var undefined = 5;", 1},
		// The identity trap from brief step 8b, written for this rule's declaration type. Two
		// `KindVariableDeclaration` anchors with the same kind and the same name in different
		// scopes, and only the outer one is written to. Comparing declaration KIND instead of node
		// identity calls the inner declaration a match for the outer one's write, carves nothing
		// out, and reports twice on code where upstream reports once. Confirmed against our
		// checker: the two declarations have distinct pointers.
		{"a nested bare undefined beside an outer one that is written",
			"function f(){ var undefined; } var undefined; undefined = 1;", 1},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "shadowingRestrictedName"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, testCase.sourceText),
				wantIds...)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		// The reads. Each of these is the near-miss of a write shape above, and each is the case
		// that would break if the carve-out treated any mention as a write.
		{"a read of the binding as a property base", "var undefined; undefined.x;"},
		{"a write to a property of the binding", "var undefined; undefined.x = 1;"},
		{"a read of the binding on the right of an assignment",
			"var undefined; var other; other = undefined;"},
		{"a read of the binding as a key rather than a value",
			"var undefined; var other; ({undefined: other} = obj);"},
		{"a negation of the binding", "var undefined; !undefined;"},
		// Two bare declarations of the same `var` merge to one symbol, which is why upstream's
		// `var undefined; var undefined;` is clean. This adds a read to confirm the merge is not
		// what carries it.
		{"two bare var undefined declarations with a read",
			"var undefined; var undefined; doSomething(undefined);"},
		// A `let undefined` inside a function is a different binding from a module-scope one and is
		// judged on its own references, not the outer one's. Confirmed against our checker: the two
		// resolve to distinct declarations.
		{"a bare let undefined inside a function", "function f() { let undefined; }"},
		// The other five names have no carve-out at all, so this is only clean because nothing is
		// named after them. It is the control for the carve-out being `undefined`-only.
		{"an ordinary bare declaration", "var notRestricted;"},
		// A property named after a restricted global is not a binding.
		{"restricted names as object literal keys", "var o = { undefined: 1, NaN: 2, eval: 3 };"},
		// A class member named after a restricted global is a property, not a binding.
		{"restricted names as class members", "class C { undefined = 1; NaN() {} }"},
		// A type alias and an interface named after a restricted global live in the type space and
		// shadow no value. Upstream never sees these because it snapshots JavaScript.
		{"an interface named after a restricted global", "interface NaN { x: number }"},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, testCase.sourceText))
		})
	}
}

// The typed harness is required, and this fails loudly if someone later swaps it for the plain one.
//
// `rule_testing.Run` hands the rule a nil checker. This rule's carve-out declines on a nil checker and
// reports, so a revert to the plain harness would not go silent the way most type-aware rules do;
// it would go *loud*, turning upstream's clean `var undefined;` into a finding. Either direction is
// a defect, and asserting the difference here names which one it is.
func TestNoShadowRestrictedNamesNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "var undefined; doSomething(undefined);"

	rule_testing.ExpectClean(t,
		rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, sourceText))

	withoutChecker := rule_testing.Run(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, sourceText)
	if len(withoutChecker.Diagnostics) != 1 {
		t.Fatalf("the untyped harness produced %d findings, want 1; the carve-out is supposed to "+
			"decline without a checker rather than guess", len(withoutChecker.Diagnostics))
	}
}

// The identity trap, asserted on offset because the text cannot see it.
//
// Two bare `var undefined` declarations, one at module scope and one inside a function, and a write
// that binds to the inner one. Exactly one binding reports and it has to be the inner one, which no
// count assertion and no text-slice assertion can check: both declarations spell `undefined`, so
// reporting the wrong one produces a finding identical in id, in count, and in reported text. Only
// the offset differs, by 30 bytes.
//
// This is the case brief step 8b names, written for a variable declaration rather than a class. A
// rule comparing declaration KIND rather than node identity finds both declarations to be
// `KindVariableDeclaration`, treats either as the anchor of that write, carves out neither, and
// reports twice. Confirmed against our checker before the rule was written: the two declarations
// have distinct pointers and the write resolves to the inner one.
func TestNoShadowRestrictedNamesReportsTheWrittenBindingNotItsNamesake(t *testing.T) {
	t.Parallel()

	const sourceText = "var undefined; function f() { var undefined; undefined = 1; }"

	// Byte 34 is the inner declaration's name; byte 4 is the outer one's, 30 bytes earlier.
	const innerNamePosition = 34

	result := rule_testing.RunTyped(t, NoShadowRestrictedNames, shadowRestrictedNamesFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1; two would mean the carve-out matched on declaration "+
			"kind rather than on node identity", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Range.Pos(); got != innerNamePosition {
		t.Errorf("the finding starts at byte %d, want %d; the outer declaration is bare and never "+
			"written to, so reporting it rather than the inner one is the defect this asserts", got,
			innerNamePosition)
	}
}

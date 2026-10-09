package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, 59 valid and 33 invalid cases, extracted by loading its test file
// with a stubbed rule tester so nothing was retyped. Every invalid case carries the exact text its
// fixer must write, which is the highest-value artifact a fixable rule's corpus has, and all 33 are
// asserted with ExpectFixedSource rather than by eye.
//
// One valid case is not here and its absence is deliberate. Upstream writes
// `/* global Object: off */` above an otherwise-reporting input, which turns the global off in its
// own configuration surface and makes the scope guard fail. Our configuration has no such
// directive and the checker has nothing to answer it with, so that input REPORTS here. It is
// carried below as a recorded divergence rather than dropped or quietly greened, because deleting
// it would leave no trace that upstream calls it clean.
//
// RunTyped writes strings.TrimSpace(source)+"\n" to disk, so every expected output is transformed
// the same way the harness transforms the input. Padding the rule to make the comparison line up
// would be the wrong repair.

func TestPreferObjectHasOwnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "valid0", source: "Object"},
		{name: "valid1", source: "Object(obj, prop)"},
		{name: "valid2", source: "Object.hasOwnProperty"},
		{name: "valid3", source: "Object.hasOwnProperty(prop)"},
		{name: "valid4", source: "hasOwnProperty(obj, prop)"},
		{name: "valid5", source: "foo.hasOwnProperty(prop)"},
		{name: "valid6", source: "foo.hasOwnProperty(obj, prop)"},
		{name: "valid7", source: "Object.hasOwnProperty.call"},
		{name: "valid8", source: "foo.Object.hasOwnProperty.call(obj, prop)"},
		{name: "valid9", source: "foo.hasOwnProperty.call(obj, prop)"},
		{name: "valid10", source: "foo.call(Object.prototype.hasOwnProperty, Object.prototype.hasOwnProperty.call)"},
		{name: "valid11", source: "Object.foo.call(obj, prop)"},
		{name: "valid12", source: "Object.hasOwnProperty.foo(obj, prop)"},
		{name: "valid13", source: "Object.hasOwnProperty.call.foo(obj, prop)"},
		{name: "valid14", source: "Object[hasOwnProperty].call(obj, prop)"},
		{name: "valid15", source: "Object.hasOwnProperty[call](obj, prop)"},
		{name: "valid16", source: "class C { #hasOwnProperty; foo() { Object.#hasOwnProperty.call(obj, prop) } }"},
		{name: "valid17", source: "class C { #call; foo() { Object.hasOwnProperty.#call(obj, prop) } }"},
		{name: "valid18", source: "(Object) => Object.hasOwnProperty.call(obj, prop)"},
		{name: "valid19", source: "Object.prototype"},
		{name: "valid20", source: "Object.prototype(obj, prop)"},
		{name: "valid21", source: "Object.prototype.hasOwnProperty"},
		{name: "valid22", source: "Object.prototype.hasOwnProperty(obj, prop)"},
		{name: "valid23", source: "Object.prototype.hasOwnProperty.call"},
		{name: "valid24", source: "foo.Object.prototype.hasOwnProperty.call(obj, prop)"},
		{name: "valid25", source: "foo.prototype.hasOwnProperty.call(obj, prop)"},
		{name: "valid26", source: "Object.foo.hasOwnProperty.call(obj, prop)"},
		{name: "valid27", source: "Object.prototype.foo.call(obj, prop)"},
		{name: "valid28", source: "Object.prototype.hasOwnProperty.foo(obj, prop)"},
		{name: "valid29", source: "Object.prototype.hasOwnProperty.call.foo(obj, prop)"},
		{name: "valid30", source: "Object.prototype.prototype.hasOwnProperty.call(a, b);"},
		{name: "valid31", source: "Object.hasOwnProperty.prototype.hasOwnProperty.call(a, b);"},
		{name: "valid32", source: "Object.prototype[hasOwnProperty].call(obj, prop)"},
		{name: "valid33", source: "Object.prototype.hasOwnProperty[call](obj, prop)"},
		{name: "valid34", source: "class C { #hasOwnProperty; foo() { Object.prototype.#hasOwnProperty.call(obj, prop) } }"},
		{name: "valid35", source: "class C { #call; foo() { Object.prototype.hasOwnProperty.#call(obj, prop) } }"},
		{name: "valid36", source: "Object[prototype].hasOwnProperty.call(obj, prop)"},
		{name: "valid37", source: "class C { #prototype; foo() { Object.#prototype.hasOwnProperty.call(obj, prop) } }"},
		{name: "valid38", source: "(Object) => Object.prototype.hasOwnProperty.call(obj, prop)"},
		{name: "valid39", source: "({})"},
		{name: "valid40", source: "({}(obj, prop))"},
		{name: "valid41", source: "({}.hasOwnProperty)"},
		{name: "valid42", source: "({}.hasOwnProperty(prop))"},
		{name: "valid43", source: "({}.hasOwnProperty(obj, prop))"},
		{name: "valid44", source: "({}.hasOwnProperty.call)"},
		{name: "valid45", source: "({}).prototype.hasOwnProperty.call(a, b);"},
		{name: "valid46", source: "({}.foo.call(obj, prop))"},
		{name: "valid47", source: "({}.hasOwnProperty.foo(obj, prop))"},
		{name: "valid48", source: "({}[hasOwnProperty].call(obj, prop))"},
		{name: "valid49", source: "({}.hasOwnProperty[call](obj, prop))"},
		{name: "valid50", source: "({}).hasOwnProperty[call](object, property)"},
		{name: "valid51", source: "({})[hasOwnProperty].call(object, property)"},
		{name: "valid52", source: "class C { #hasOwnProperty; foo() { ({}.#hasOwnProperty.call(obj, prop)) } }"},
		{name: "valid53", source: "class C { #call; foo() { ({}.hasOwnProperty.#call(obj, prop)) } }"},
		{name: "valid54", source: "({ foo }.hasOwnProperty.call(obj, prop))"},
		{name: "valid55", source: "(Object) => ({}).hasOwnProperty.call(obj, prop)"},
		{name: "valid56", source: "\n        let obj = {};\n        Object.hasOwn(obj,\"\");\n        "},
		{name: "valid57", source: "const hasProperty = Object.hasOwn(object, property);"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", testCase.source))
		})
	}
}

// preferObjectHasOwnDeclinesToFix marks a case upstream reports and deliberately does not repair,
// which its corpus spells `output: null`. A fixer that repaired one of these would be a defect no
// message-id fixture could see.
const preferObjectHasOwnDeclinesToFix = "\x00declines"

func TestPreferObjectHasOwnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantFixed string
	}{
		{name: "invalid0", source: "Object.hasOwnProperty.call(obj, 'foo')", wantFixed: "Object.hasOwn(obj, 'foo')\n"},
		{name: "invalid1", source: "Object.hasOwnProperty.call(obj, property)", wantFixed: "Object.hasOwn(obj, property)\n"},
		{name: "invalid2", source: "Object.prototype.hasOwnProperty.call(obj, 'foo')", wantFixed: "Object.hasOwn(obj, 'foo')\n"},
		{name: "invalid3", source: "({}).hasOwnProperty.call(obj, 'foo')", wantFixed: "Object.hasOwn(obj, 'foo')\n"},
		{name: "invalid4", source: "Object/* comment */.prototype.hasOwnProperty.call(a, b);", wantFixed: preferObjectHasOwnDeclinesToFix},
		{name: "invalid5", source: "const hasProperty = Object.prototype.hasOwnProperty.call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid6", source: "const hasProperty = (( Object.prototype.hasOwnProperty.call(object, property) ));", wantFixed: "const hasProperty = (( Object.hasOwn(object, property) ));\n"},
		{name: "invalid7", source: "const hasProperty = (( Object.prototype.hasOwnProperty.call ))(object, property);", wantFixed: "const hasProperty = (( Object.hasOwn ))(object, property);\n"},
		{name: "invalid8", source: "const hasProperty = (( Object.prototype.hasOwnProperty )).call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid9", source: "const hasProperty = (( Object.prototype )).hasOwnProperty.call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid10", source: "const hasProperty = (( Object )).prototype.hasOwnProperty.call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid11", source: "const hasProperty = {}.hasOwnProperty.call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid12", source: "const hasProperty={}.hasOwnProperty.call(object, property);", wantFixed: "const hasProperty=Object.hasOwn(object, property);\n"},
		{name: "invalid13", source: "const hasProperty = (( {}.hasOwnProperty.call(object, property) ));", wantFixed: "const hasProperty = (( Object.hasOwn(object, property) ));\n"},
		{name: "invalid14", source: "const hasProperty = (( {}.hasOwnProperty.call ))(object, property);", wantFixed: "const hasProperty = (( Object.hasOwn ))(object, property);\n"},
		{name: "invalid15", source: "const hasProperty = (( {}.hasOwnProperty )).call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid16", source: "const hasProperty = (( {} )).hasOwnProperty.call(object, property);", wantFixed: "const hasProperty = Object.hasOwn(object, property);\n"},
		{name: "invalid17", source: "function foo(){return {}.hasOwnProperty.call(object, property)}", wantFixed: "function foo(){return Object.hasOwn(object, property)}\n"},
		{name: "invalid18", source: "function foo(){return{}.hasOwnProperty.call(object, property)}", wantFixed: "function foo(){return Object.hasOwn(object, property)}\n"},
		{name: "invalid19", source: "function foo(){return/*comment*/{}.hasOwnProperty.call(object, property)}", wantFixed: "function foo(){return/*comment*/Object.hasOwn(object, property)}\n"},
		{name: "invalid20", source: "async function foo(){return await{}.hasOwnProperty.call(object, property)}", wantFixed: "async function foo(){return await Object.hasOwn(object, property)}\n"},
		{name: "invalid21", source: "async function foo(){return await/*comment*/{}.hasOwnProperty.call(object, property)}", wantFixed: "async function foo(){return await/*comment*/Object.hasOwn(object, property)}\n"},
		{name: "invalid22", source: "for (const x of{}.hasOwnProperty.call(object, property).toString());", wantFixed: "for (const x of Object.hasOwn(object, property).toString());\n"},
		{name: "invalid23", source: "for (const x of/*comment*/{}.hasOwnProperty.call(object, property).toString());", wantFixed: "for (const x of/*comment*/Object.hasOwn(object, property).toString());\n"},
		{name: "invalid24", source: "for (const x in{}.hasOwnProperty.call(object, property).toString());", wantFixed: "for (const x in Object.hasOwn(object, property).toString());\n"},
		{name: "invalid25", source: "for (const x in/*comment*/{}.hasOwnProperty.call(object, property).toString());", wantFixed: "for (const x in/*comment*/Object.hasOwn(object, property).toString());\n"},
		{name: "invalid26", source: "function foo(){return({}.hasOwnProperty.call)(object, property)}", wantFixed: "function foo(){return(Object.hasOwn)(object, property)}\n"},
		{name: "invalid27", source: "Object['prototype']['hasOwnProperty']['call'](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
		{name: "invalid28", source: "Object[`prototype`][`hasOwnProperty`][`call`](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
		{name: "invalid29", source: "Object['hasOwnProperty']['call'](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
		{name: "invalid30", source: "Object[`hasOwnProperty`][`call`](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
		{name: "invalid31", source: "({})['hasOwnProperty']['call'](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
		{name: "invalid32", source: "({})[`hasOwnProperty`][`call`](object, property);", wantFixed: "Object.hasOwn(object, property);\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "useHasOwn")
			if testCase.wantFixed == preferObjectHasOwnDeclinesToFix {
				if len(result.Diagnostics) > 0 && len(result.Diagnostics[0].Fixes) > 0 {
					t.Errorf("upstream declines to fix this case, and this offered %d fixes",
						len(result.Diagnostics[0].Fixes))
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// The span, which ExpectFindings cannot see. Upstream reports the whole call expression rather than
// the callee it replaces, and the two are different nodes: a finding on the callee would carry a
// correct repair while pointing somewhere the reader was never shown.
func TestPreferObjectHasOwnSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantText string
	}{
		{"plainCall", "Object.prototype.hasOwnProperty.call(obj, prop)",
			"Object.prototype.hasOwnProperty.call(obj, prop)"},
		{"insideADeclaration", "const has = Object.hasOwnProperty.call(object, property);",
			"Object.hasOwnProperty.call(object, property)"},
		{"emptyLiteral", "({}).hasOwnProperty.call(obj, prop)", "({}).hasOwnProperty.call(obj, prop)"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			// The harness trimmed the source, so slice what it actually wrote.
			written := strings.TrimSpace(testCase.source) + "\n"
			if got := written[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
			if finding.Message.Id != "useHasOwn" {
				t.Errorf("message id: got %q", finding.Message.Id)
			}
			wantPrefix := "Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'. "
			if got := finding.Message.Description; !strings.HasPrefix(got, wantPrefix) {
				t.Errorf("message:\n got %q\nwant prefix %q", got, wantPrefix)
			}
		})
	}
}

// The scope guard, which is the half of this rule that does not fall out of the syntax. Upstream
// asks a scope table whether `Object` resolves to the global binding; this asks the checker whether
// every declaration of it is ambient. Both answers were probed on all four shapes before the rule
// was built on either.
func TestPreferObjectHasOwnScopeGuard(t *testing.T) {
	t.Parallel()

	shadowed := []struct {
		name   string
		source string
	}{
		{"arrowParameter", "((Object: any) => Object.prototype.hasOwnProperty.call(obj, prop))(null);"},
		{"functionParameter", "function f(Object: any) { Object.hasOwnProperty.call(obj, prop); }"},
		{"blockScopedConst", "{ const Object: any = {}; Object.prototype.hasOwnProperty.call(obj, prop); }"},
		{"emptyLiteralUnderAShadow", "((Object: any) => ({}).hasOwnProperty.call(obj, prop))(null);"},
	}
	for _, testCase := range shadowed {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", testCase.source))
		})
	}

	// A control, so the four clean verdicts above mean something. The same call with no shadow
	// reports, which is the only thing that separates a working guard from a dead rule.
	t.Run("controlWithNoShadowReports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts",
			"Object.prototype.hasOwnProperty.call(obj, prop);"), "useHasOwn")
	})
}

// Upstream's one valid case this port cannot express, recorded as reporting rather than deleted.
//
// `/* global Object: off */` is an ESLint configuration comment that removes a global from its
// scope analysis, so the rule's `variable && variable.scope.type === "global"` test fails and the
// input is clean. Nothing in cohere reads that comment and the type checker still resolves `Object`
// to lib.es5.d.ts, so the guard passes and the finding stands.
//
// Pinned at the layer that decides it: this is a configuration-surface difference rather than a
// rule difference, and a port bent to make it clean would be wrong about every other input.
func TestPreferObjectHasOwnDivergesOnTheGlobalOffDirective(t *testing.T) {
	t.Parallel()

	source := "/* global Object: off */\n({}).hasOwnProperty.call(a, b);"
	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", source), "useHasOwn")
}

// Private names, where the shelf helper and upstream disagree and the disagreement is invisible in
// the verdict.
//
// property.AccessedName answers "#call" for `x.#call`, while upstream's getStaticPropertyName
// answers null for a private name. Neither string equals "call", so both reach the same silence on
// upstream's four private-name cases, and no fixture over those cases could tell the two apart.
// Recorded because building on the helper without knowing this would be building on a coincidence.
func TestPreferObjectHasOwnPrivateNames(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class C { #hasOwnProperty: any; foo() { (Object as any).#hasOwnProperty.call(obj, prop) } }",
		"class C { #call: any; foo() { (Object as any).hasOwnProperty.#call(obj, prop) } }",
	}
	for _, source := range cases {
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", source))
	}
}

// The identifier must be `Object` by name, and that test is not redundant with the ambient check.
//
// Found by a surviving mutant. Every other global built into the language resolves to ambient
// declarations exactly as `Object` does, so dropping the name test leaves the rule reporting
// `Math.hasOwnProperty.call(a, b)` and offering to rewrite it as `Object.hasOwn(a, b)`, which asks
// a different question of a different object. Measured clean upstream with a control that fired.
//
// Upstream gets this for free because it looks the name up rather than resolving a node, so its
// corpus has no case for it: `foo.hasOwnProperty.call(obj, prop)` is the nearest, and `foo` is
// undeclared, which our ambient check already declines for a different reason.
func TestPreferObjectHasOwnRequiresTheNameObject(t *testing.T) {
	t.Parallel()

	otherGlobals := []struct {
		name   string
		source string
	}{
		{"math", "Math.hasOwnProperty.call(obj, prop);"},
		{"json", "JSON.hasOwnProperty.call(obj, prop);"},
		{"arrayThroughPrototype", "Array.prototype.hasOwnProperty.call(obj, prop);"},
	}
	for _, testCase := range otherGlobals {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts", testCase.source))
		})
	}

	// A control, because three clean verdicts prove nothing on their own.
	t.Run("controlObjectStillReports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts",
			"Object.hasOwnProperty.call(obj, prop);"), "useHasOwn")
	})

	// `globalThis` resolves to a symbol with ZERO declarations, which is the shape that crashed a
	// probe written for this measurement. The rule's length guard is what keeps that a decline
	// rather than a panic, and no upstream case reaches it.
	t.Run("globalThisHasNoDeclarationsAndDoesNotPanic", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferObjectHasOwn, "file.ts",
			"globalThis.hasOwnProperty.call(obj, prop);"))
	})
}

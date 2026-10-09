package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// extendNativeFile is where the fixtures pretend to live.
const extendNativeFile = "/repository/source/ExtendNative.ts"

// The corpus is ESLint's own, extracted rather than retyped.
//
// Every case is verbatim from `eslint/tests/lib/rules/no-extend-native.js` at eslint 10.8.1: 19
// valid and 21 invalid. Extracted by executing that file against a stub RuleTester, with each case's
// own `options` and each finding's own `data.builtin` carried across, so the expectations are
// upstream's rather than recovered. All 40 were additionally driven through the installed build.
//
// The message interpolates the builtin's name, so the invalid table names the expected builtin per
// finding rather than only the message id: `ExpectFindings` cannot see anything the format string
// does, and one input reports twice with a different name each time.
func TestNoExtendNativeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		builtins   []string
	}{
		{"Object.prototype.p = 0", nil, []string{"Object"}},
		{"BigInt.prototype.p = 0", nil, []string{"BigInt"}},
		{"WeakRef.prototype.p = 0", nil, []string{"WeakRef"}},
		{"FinalizationRegistry.prototype.p = 0", nil, []string{"FinalizationRegistry"}},
		{"AggregateError.prototype.p = 0", nil, []string{"AggregateError"}},
		{"Function.prototype['p'] = 0", nil, []string{"Function"}},
		{"String['prototype'].p = 0", nil, []string{"String"}},
		{"Number['prototype']['p'] = 0", nil, []string{"Number"}},
		{"Object.defineProperty(Array.prototype, 'p', {value: 0})", nil, []string{"Array"}},
		{"Object.defineProperties(Array.prototype, {p: {value: 0}})", nil, []string{"Array"}},
		{"Object.defineProperties(Array.prototype, {p: {value: 0}, q: {value: 0}})", nil, []string{"Array"}},
		{"Number['prototype']['p'] = 0", NoExtendNativeOptions{Exceptions: []string{"Object"}}, []string{"Number"}},
		{"Object.prototype.p = 0; Object.prototype.q = 0", nil, []string{"Object", "Object"}},
		{"function foo() { Object.prototype.p = 0 }", nil, []string{"Object"}},
		{"(Object?.prototype).p = 0", nil, []string{"Object"}},
		{"Object.defineProperty(Object?.prototype, 'p', { value: 0 })", nil, []string{"Object"}},
		{"Object?.defineProperty(Object.prototype, 'p', { value: 0 })", nil, []string{"Object"}},
		{"(Object?.defineProperty)(Object.prototype, 'p', { value: 0 })", nil, []string{"Object"}},
		{"Array.prototype.p &&= 0", nil, []string{"Array"}},
		{"Array.prototype.p ||= 0", nil, []string{"Array"}},
		{"Array.prototype.p ??= 0", nil, []string{"Array"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
				testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.builtins) {
				t.Fatalf("wanted %d diagnostics, got %d", len(testCase.builtins), len(result.Diagnostics))
			}
			for index, wantBuiltin := range testCase.builtins {
				diagnostic := result.Diagnostics[index]
				if diagnostic.Message.Id != "unexpected" {
					t.Fatalf("finding %d had id %q", index, diagnostic.Message.Id)
				}
				// The interpolated name, which the id assertion cannot see. A rule reusing the
				// wrong variable in the format slot passes every id fixture in this file.
				if want := "`" + wantBuiltin + ".prototype`"; !strings.Contains(
					diagnostic.Message.Description, want) {
					t.Fatalf("finding %d did not name %s: %q", index, want, diagnostic.Message.Description)
				}
			}
		})
	}
}

// The clean cases, with their own options.
//
// Four groups: a receiver that is not a builtin, a builtin reached through something other than its
// own identifier, a shadowed builtin, and a builtin named in the `exceptions` option. The
// `parseFloat.prototype.x = 1` row is the one that pins the uppercase-first filter on the builtin
// set: `parseFloat` is a global and its prototype is being extended, and it is clean purely because
// its name does not start with a capital.
func TestNoExtendNativeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"x.prototype.p = 0", nil},
		{"x.prototype['p'] = 0", nil},
		{"Object.p = 0", nil},
		{"Object.toString.bind = 0", nil},
		{"Object['toString'].bind = 0", nil},
		{"Object.defineProperty(x, 'p', {value: 0})", nil},
		{"Object.defineProperties(x, {p: {value: 0}})", nil},
		{"global.Object.prototype.toString = 0", nil},
		{"this.Object.prototype.toString = 0", nil},
		{"with(Object) { prototype.p = 0; }", nil},
		{"o = Object; o.prototype.toString = 0", nil},
		{"eval('Object.prototype.toString = 0')", nil},
		{"parseFloat.prototype.x = 1", nil},
		{"Object.prototype.g = 0", NoExtendNativeOptions{Exceptions: []string{"Object"}}},
		{"obj[Object.prototype] = 0", nil},
		{"Object.defineProperty()", nil},
		{"Object.defineProperties()", nil},
		{"function foo() { var Object = function() {}; Object.prototype.p = 0 }", nil},
		{"{ let Object = function() {}; Object.prototype.p = 0 }", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtendNative,
				extendNativeFile, testCase.sourceText, testCase.options))
		})
	}
}

// Where the finding points, which the message-id fixtures above cannot see.
//
// The two shapes point at different nodes and both are measured on eslint 10.8.1: an assignment
// reports the whole assignment, and a define call reports the whole call. A port reporting the
// `.prototype` member access instead passes every id fixture in this file.
func TestNoExtendNativeReportsTheWholeExtension(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPos    int
		wantEnd    int
		builtin    string
	}{
		{"Object.prototype.p = 0", 0, 22, "Object"},
		{"Object.defineProperty(Array.prototype, 'p', {value: 0})", 0, 55, "Array"},
		{"Array.prototype.p &&= 0", 0, 23, "Array"},
		{"(Object?.prototype).p = 0", 0, 25, "Object"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
				testCase.sourceText, nil)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			got := result.Diagnostics[0].Range
			if got.Pos() != testCase.wantPos || got.End() != testCase.wantEnd {
				t.Fatalf("reported [%d:%d) which is %q, wanted [%d:%d) which is %q",
					got.Pos(), got.End(), onDisk[got.Pos():got.End()],
					testCase.wantPos, testCase.wantEnd, onDisk[testCase.wantPos:testCase.wantEnd])
			}
		})
	}
}

// Silent shapes the corpus does not write, each measured upstream first.
//
// The first two are the ones that read like defects: `Object.prototype.p++` and
// `delete Object.prototype.p` both extend or change a native prototype and both are clean on eslint
// 10.8.1, because upstream matches an assignment expression and nothing else. Reproduced rather than
// improved on, and recorded here so the next reader does not helpfully "fix" it.
func TestNoExtendNativeDeclinesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an update expression rather than an assignment", "Object.prototype.p++"},
		{"a delete rather than an assignment", "delete Object.prototype.p"},
		{"an assignment to prototype itself", "Object.prototype = 0"},
		{"an assignment two levels below the prototype", "Object.prototype.p.q = 0"},
		{"a call that is not one of the two define methods", "Object.freeze(Array.prototype)"},
		{"a prototype in a later argument", "Object.defineProperty(x, Array.prototype)"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtendNative,
				extendNativeFile, testCase.sourceText, nil))
		})
	}
}

// The option surface, routed through the rule's own registered decoder rather than by building the
// struct here.
//
// Handing the harness a struct directly would leave the serde tag untested, and the tag is the line
// most likely to have no upstream counterpart. This decodes the same JSON the config would carry.
func TestNoExtendNativeDecodesItsOptions(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoExtendNativeOptions]()

	decoded, err := decode(json.RawMessage(`{"exceptions": ["Object"]}`))
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	settings, ok := decoded.(NoExtendNativeOptions)
	if !ok {
		t.Fatalf("decoded to %T", decoded)
	}
	if len(settings.Exceptions) != 1 || settings.Exceptions[0] != "Object" {
		t.Fatalf("exceptions decoded as %#v", settings.Exceptions)
	}

	// And the decoded value reaches the rule: the same input reports without the exception.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
		"Object.prototype.p = 0", decoded))
	if got := len(rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
		"Object.prototype.p = 0", nil).Diagnostics); got != 1 {
		t.Fatalf("without the exception the same input gave %d diagnostics", got)
	}
}

// A rule configured as a bare "error" is handed nil, which must land on the strict default.
//
// This bypasses the decoder entirely, which is the path a bare `"error"` actually takes: the config
// turns a decode error on empty input into nil, and a type assertion on nil yields the zero value.
// The zero value being correct here is a coincidence worth pinning rather than relying on.
func TestNoExtendNativeDefaultsWithoutTheDecoder(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
		"Object.prototype.p = 0", nil), "unexpected")
}

// The typed harness is required, and a revert to the plain one must fail loudly.
//
// The rule guards on a nil checker, so under `rule_testing.Run` it goes silent rather than panicking,
// and every StaysSilent case above would pass vacuously.
func TestNoExtendNativeNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "Object.prototype.p = 0"

	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
		source, nil), "unexpected")

	if got := len(rule_testing.RunWithOptions(t, NoExtendNative, extendNativeFile,
		source, nil).Diagnostics); got != 0 {
		t.Fatalf("the untyped harness produced %d diagnostics, so the nil-checker guard has moved "+
			"and this test no longer measures what it claims", got)
	}
}

// READING a native prototype property is not extending it, which the imported corpus cannot see.
//
// The corpus writes 21 reporting cases and every one of them is an assignment or a define call, so a
// port that anchors on `KindBinaryExpression` and forgets to ask which operator it found passes all
// 40 imported cases. Removing the operator guard survived them, which is how this test came to
// exist.
//
// Reading is legitimate and common: comparing, adding, or testing a native prototype property does
// not change the shared object. All five rows are clean on eslint 10.8.1, measured before they were
// written here.
func TestNoExtendNativeDeclinesReadingAPrototypeProperty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a strict comparison", "Object.prototype.p === 0"},
		{"an addition", "Object.prototype.p + 0"},
		{"a comma expression", "Object.prototype.p, 0"},
		{"a comparison in a condition", "if (Object.prototype.p == 0) {}"},
		{"an instanceof", "Object.prototype.p instanceof Foo"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtendNative,
				extendNativeFile, testCase.sourceText, nil))
		})
	}
}

// The define call's RECEIVER must be `Object`, which the imported corpus cannot see.
//
// Every define case in the corpus writes `Object.defineProperty` or `Object.defineProperties`, so a
// port that checks the method name and forgets the receiver passes all 40 imported cases. A mutation
// dropping exactly that check survived them.
//
// `Reflect.defineProperty` is the sharp row: it is a real API that extends the prototype in exactly
// the same way, and eslint 10.8.1 is silent on it because upstream matches the receiver name
// `Object` and nothing else. That is a genuine gap in upstream's coverage, reproduced here rather
// than improved on, and recorded so the next reader does not helpfully widen it.
//
// All four measured clean on eslint 10.8.1 before being written here.
func TestNoExtendNativeDeclinesADefineCallOnAnotherReceiver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"Reflect.defineProperty, which really does the same thing", "Reflect.defineProperty(Array.prototype, \"p\", {value:0})"},
		{"an unrelated object's defineProperty", "Foo.defineProperty(Array.prototype, \"p\", {value:0})"},
		{"another builtin's defineProperty", "Math.defineProperty(Array.prototype, \"p\", {value:0})"},
		{"another builtin's defineProperties", "JSON.defineProperties(Array.prototype, {})"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoExtendNative,
				extendNativeFile, testCase.sourceText, nil))
		})
	}
}

// A no-substitution template subscript names the property too, which the corpus cannot see.
//
// The corpus writes `String['prototype']` and `Number['prototype']['p']`, so it exercises the string
// subscript but never the template one, and the narrower `property.Textual` accept set covers
// strings. A mutation narrowing the set therefore survived all 40 imported cases.
//
// Upstream asks for the static property name, which resolves a no-substitution template as well, and
// eslint 10.8.1 reports all three rows below. They cover the three positions the rule reads a
// property name in: the prototype of an assignment target, the prototype inside a define call, and
// the define method's own name.
func TestNoExtendNativeReadsATemplateSubscript(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		builtin    string
	}{
		{"a template subscript for prototype", "Object[`prototype`].p = 0", "Object"},
		{"a template prototype inside a define call", "Object.defineProperty(Array[`prototype`], \"p\", {value:0})", "Array"},
		{"a template subscript for the define method", "Object[`defineProperty`](Array.prototype, \"p\", {value:0})", "Array"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoExtendNative, extendNativeFile,
				testCase.sourceText, nil)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			if want := "`" + testCase.builtin + ".prototype`"; !strings.Contains(
				result.Diagnostics[0].Message.Description, want) {
				t.Fatalf("did not name %s: %q", want, result.Diagnostics[0].Message.Description)
			}
		})
	}
}

package core

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// classMethodsUseThisFile is where the fixtures pretend to live.
//
// A `.ts` name: this rule has no file gate, and nothing in its corpus is JSX.
const classMethodsUseThisFile = "/repository/source/ClassMethodsUseThis.ts"

// classMethodsUseThisCase is one row of upstream's corpus.
type classMethodsUseThisCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that
	// can be found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// options is the RAW JSON of upstream's options object, routed through the rule's own
	// exported decoder rather than built as a struct, so the decoder's defaults and its
	// empty-input path are under test.
	options string

	// ids are the message ids upstream produced for this input, in order.
	ids []string
}

// classMethodsUseThisFiresCases are the rows upstream reports on.
var classMethodsUseThisFiresCases = []classMethodsUseThisCase{
	{
		name:    "invalid-0",
		source:  "class A { foo() {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-1",
		source:  "class A { foo() {/**this**/} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-2",
		source:  "class A { foo() {var a = function () {this};} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-3",
		source:  "class A { foo() {var a = function () {var b = function(){this}};} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-4",
		source:  "class A { foo() {window.this} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-5",
		source:  "class A { foo() {that.this = 'this';} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-6",
		source:  "class A { foo() { () => undefined; } }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-7",
		source:  "class A { foo() {} bar() {} }",
		options: "{\"exceptMethods\": [\"bar\"]}",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-8",
		source:  "class A { foo() {} hasOwnProperty() {} }",
		options: "{\"exceptMethods\": [\"foo\"]}",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-9",
		source:  "class A { [foo]() {} }",
		options: "{\"exceptMethods\": [\"foo\"]}",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-10",
		source:  "class A { #foo() { } foo() {} #bar() {} }",
		options: "{\"exceptMethods\": [\"#foo\"]}",
		ids:     []string{"missingThis", "missingThis"},
	},
	{
		name:    "invalid-11",
		source:  "class A { foo(){} 'bar'(){} 123(){} [`baz`](){} [a](){} [f(a)](){} get quux(){} set[a](b){} *quuux(){} }",
		options: "",
		ids:     []string{"missingThis", "missingThis", "missingThis", "missingThis", "missingThis", "missingThis", "missingThis", "missingThis", "missingThis"},
	},
	{
		name:    "invalid-12",
		source:  "class A { foo = function() {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-13",
		source:  "class A { foo = () => {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-14",
		source:  "class A { #foo = function() {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-15",
		source:  "class A { #foo = () => {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-16",
		source:  "class A { #foo() {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-17",
		source:  "class A { get #foo() {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-18",
		source:  "class A { set #foo(x) {} }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-19",
		source:  "class A { foo () { return class { foo = this }; } }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-20",
		source:  "class A { foo () { return function () { foo = this }; } }",
		options: "",
		ids:     []string{"missingThis"},
	},
	{
		name:    "invalid-21",
		source:  "class A { foo () { return class { static { this; } } } }",
		options: "",
		ids:     []string{"missingThis"},
	},
}

// classMethodsUseThisSilentCases are the rows upstream is clean on.
var classMethodsUseThisSilentCases = []classMethodsUseThisCase{
	{
		name:    "valid-0",
		source:  "class A { constructor() {} }",
		options: "",
	},
	{
		name:    "valid-1",
		source:  "class A { foo() {this} }",
		options: "",
	},
	{
		name:    "valid-2",
		source:  "class A { foo() {this.bar = 'bar';} }",
		options: "",
	},
	{
		name:    "valid-3",
		source:  "class A { foo() {bar(this);} }",
		options: "",
	},
	{
		name:    "valid-4",
		source:  "class A extends B { foo() {super.foo();} }",
		options: "",
	},
	{
		name:    "valid-5",
		source:  "class A { foo() { if(true) { return this; } } }",
		options: "",
	},
	{
		name:    "valid-6",
		source:  "class A { static foo() {} }",
		options: "",
	},
	{
		name:    "valid-7",
		source:  "({ a(){} });",
		options: "",
	},
	{
		name:    "valid-8",
		source:  "class A { foo() { () => this; } }",
		options: "",
	},
	{
		name:    "valid-9",
		source:  "({ a: function () {} });",
		options: "",
	},
	{
		name:    "valid-10",
		source:  "class A { foo() {this} bar() {} }",
		options: "{\"exceptMethods\": [\"bar\"]}",
	},
	{
		name:    "valid-11",
		source:  "class A { \"foo\"() { } }",
		options: "{\"exceptMethods\": [\"foo\"]}",
	},
	{
		name:    "valid-12",
		source:  "class A { 42() { } }",
		options: "{\"exceptMethods\": [\"42\"]}",
	},
	{
		name:    "valid-13",
		source:  "class A { foo = function() {this} }",
		options: "",
	},
	{
		name:    "valid-14",
		source:  "class A { foo = () => {this} }",
		options: "",
	},
	{
		name:    "valid-15",
		source:  "class A { foo = () => {super.toString} }",
		options: "",
	},
	{
		name:    "valid-16",
		source:  "class A { static foo = function() {} }",
		options: "",
	},
	{
		name:    "valid-17",
		source:  "class A { static foo = () => {} }",
		options: "",
	},
	{
		name:    "valid-18",
		source:  "class A { #bar() {} }",
		options: "{\"exceptMethods\": [\"#bar\"]}",
	},
	{
		name:    "valid-19",
		source:  "class A { foo = function () {} }",
		options: "{\"enforceForClassFields\": false}",
	},
	{
		name:    "valid-20",
		source:  "class A { foo = () => {} }",
		options: "{\"enforceForClassFields\": false}",
	},
	{
		name:    "valid-21",
		source:  "class A { foo() { return class { [this.foo] = 1 }; } }",
		options: "",
	},
	{
		name:    "valid-22",
		source:  "class A { static {} }",
		options: "",
	},
}

// decodedClassMethodsUseThis routes a row's raw JSON through the rule's own exported decoder.
func decodedClassMethodsUseThis(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeClassMethodsUseThisOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return decoded
}

func TestClassMethodsUseThisFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range classMethodsUseThisFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassMethodsUseThis, classMethodsUseThisFile, testCase.source,
				decodedClassMethodsUseThis(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestClassMethodsUseThisStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range classMethodsUseThisSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassMethodsUseThis, classMethodsUseThisFile, testCase.source,
				decodedClassMethodsUseThis(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestClassMethodsUseThisSkipsBodilessMembers covers two TypeScript shapes upstream cannot express.
//
// A member with no body cannot use `this`, so judging it is a false positive rather than a finding.
// Upstream's corpus is JavaScript and contains neither shape, so nothing imported can see this:
//
//	abstract parse(value: unknown): TOutput;   an abstract method
//	is(a: T): this;                            an overload signature, whose implementation IS judged
//
// Found by differential rather than by fixtures. Driving ESLint over the 1,163 files this rule
// touches on the real tree gave 217 against an earlier draft's 226, and every one of the nine extras
// was bodiless. The overload case is the sharper half, because the same NAME then reports twice and
// reads as a duplicated finding rather than as a wrong verdict.
//
// After the fix the two agree on all 1,163 files with zero disagreements.
func TestClassMethodsUseThisSkipsBodilessMembers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports bool
	}{
		{
			name:    "abstract-method-is-not-judged",
			source:  "abstract class A {\n  abstract parse(value: unknown): string;\n}\n",
			reports: false,
		},
		{
			name: "overload-signatures-are-not-judged",
			// One finding, not three: the two signatures are declarations and only the
			// implementation carries a body.
			source:  "class A {\n  is(a: string): this;\n  is(a: number): this;\n  is(a: unknown): unknown { return 1; }\n}\n",
			reports: true,
		},
		{
			// The control. Without it, a rule that had stopped judging methods entirely would
			// satisfy both rows above.
			name:    "control-an-ordinary-method-still-reports",
			source:  "class A {\n  parse(value: unknown): string { return String(value); }\n}\n",
			reports: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ClassMethodsUseThis, classMethodsUseThisFile,
				testCase.source, decodedClassMethodsUseThis(t, ""))
			if !testCase.reports {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "missingThis")
		})
	}
}

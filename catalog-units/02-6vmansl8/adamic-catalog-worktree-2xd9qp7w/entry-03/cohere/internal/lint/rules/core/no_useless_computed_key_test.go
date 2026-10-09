package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// uselessComputedKeyFile is where the fixtures pretend to live.
const uselessComputedKeyFile = "/repository/source/UselessComputedKey.ts"

// noUselessComputedKeyCase is one upstream case.
//
// `optionsJson` is the raw config text rather than a built struct, so every case is routed through
// the rule's own decoder. That is the only thing that puts the default inversion under test: the
// option defaults to TRUE, so a struct built by hand in a fixture would silently agree with a
// decoder that had the default backwards.
type noUselessComputedKeyCase struct {
	source      string
	optionsJson string
	fixedSource *string
	properties  []string
}

func stringPointer(value string) *string { return &value }

// decodeUselessComputedKeyOptionsForTest routes a case's options through the shipped decoder.
//
// A nil `optionsJson` is upstream writing no `options` at all, which reaches a rule configured as a
// bare severity as nil, so it is passed through as nil rather than as an empty object. The two are
// different inputs to the decoder and both have to work.
func decodeUselessComputedKeyOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeNoUselessComputedKeyOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return decoded
}

// The valid cases, verbatim from upstream. Extracted by evaluating the corpus against a stub
// RuleTester and serialising each case, so no source string was retyped and no escape could be
// cooked on the way in. All 96 were then located byte for byte in the corpus file.
var noUselessComputedKeyCleanCases = []noUselessComputedKeyCase{
	{source: "({ 'a': 0, b(){} })", optionsJson: ""},
	{source: "({ [x]: 0 });", optionsJson: ""},
	{source: "({ a: 0, [b](){} })", optionsJson: ""},
	{source: "({ ['__proto__']: [] })", optionsJson: ""},
	{source: "var { 'a': foo } = obj", optionsJson: ""},
	{source: "var { [a]: b } = obj;", optionsJson: ""},
	{source: "var { a } = obj;", optionsJson: ""},
	{source: "var { a: a } = obj;", optionsJson: ""},
	{source: "var { a: b } = obj;", optionsJson: ""},
	{source: "class Foo { a() {} }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { 'a'() {} }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { [x]() {} }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { ['constructor']() {} }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { static ['prototype']() {} }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "(class { 'a'() {} })", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "(class { [x]() {} })", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "(class { ['constructor']() {} })", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "(class { static ['prototype']() {} })", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { 'x'() {} }", optionsJson: ""},
	{source: "(class { [x]() {} })", optionsJson: ""},
	{source: "class Foo { static constructor() {} }", optionsJson: ""},
	{source: "class Foo { prototype() {} }", optionsJson: ""},
	{source: "class Foo { ['x']() {} }", optionsJson: "{\"enforceForClassMembers\":false}"},
	{source: "(class { ['x']() {} })", optionsJson: "{\"enforceForClassMembers\":false}"},
	{source: "class Foo { static ['constructor']() {} }", optionsJson: "{\"enforceForClassMembers\":false}"},
	{source: "class Foo { ['prototype']() {} }", optionsJson: "{\"enforceForClassMembers\":false}"},
	{source: "class Foo { a }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { ['constructor'] }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { static ['constructor'] }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "class Foo { static ['prototype'] }", optionsJson: "{\"enforceForClassMembers\":true}"},
	{source: "({ [99999999999999999n]: 0 })", optionsJson: ""},
}

// The invalid cases, verbatim from upstream, each carrying its own expected repair.
// A nil fixedSource is upstream's `output: null`: reported, deliberately not fixed.
var noUselessComputedKeyFiringCases = []noUselessComputedKeyCase{
	{source: "({ ['0']: 0 })", optionsJson: "", fixedSource: stringPointer("({ '0': 0 })"), properties: []string{"'0'"}},
	{source: "var { ['0']: a } = obj", optionsJson: "", fixedSource: stringPointer("var { '0': a } = obj"), properties: []string{"'0'"}},
	{source: "({ ['0+1,234']: 0 })", optionsJson: "", fixedSource: stringPointer("({ '0+1,234': 0 })"), properties: []string{"'0+1,234'"}},
	{source: "({ [0]: 0 })", optionsJson: "", fixedSource: stringPointer("({ 0: 0 })"), properties: []string{"0"}},
	{source: "var { [0]: a } = obj", optionsJson: "", fixedSource: stringPointer("var { 0: a } = obj"), properties: []string{"0"}},
	{source: "({ ['x']: 0 })", optionsJson: "", fixedSource: stringPointer("({ 'x': 0 })"), properties: []string{"'x'"}},
	{source: "var { ['x']: a } = obj", optionsJson: "", fixedSource: stringPointer("var { 'x': a } = obj"), properties: []string{"'x'"}},
	{source: "var { ['__proto__']: a } = obj", optionsJson: "", fixedSource: stringPointer("var { '__proto__': a } = obj"), properties: []string{"'__proto__'"}},
	{source: "({ ['x']() {} })", optionsJson: "", fixedSource: stringPointer("({ 'x'() {} })"), properties: []string{"'x'"}},
	{source: "({ [/* this comment prevents a fix */ 'x']: 0 })", optionsJson: "", fixedSource: nil, properties: []string{"'x'"}},
	{source: "({ ['x' /* this comment also prevents a fix */]: 0 })", optionsJson: "", fixedSource: nil, properties: []string{"'x'"}},
	{source: "({ [('x')]: 0 })", optionsJson: "", fixedSource: stringPointer("({ 'x': 0 })"), properties: []string{"'x'"}},
	{source: "var { [('x')]: a } = obj", optionsJson: "", fixedSource: stringPointer("var { 'x': a } = obj"), properties: []string{"'x'"}},
	{source: "({ *['x']() {} })", optionsJson: "", fixedSource: stringPointer("({ *'x'() {} })"), properties: []string{"'x'"}},
	{source: "({ async ['x']() {} })", optionsJson: "", fixedSource: stringPointer("({ async 'x'() {} })"), properties: []string{"'x'"}},
	{source: "({ get[.2]() {} })", optionsJson: "", fixedSource: stringPointer("({ get.2() {} })"), properties: []string{".2"}},
	{source: "({ set[.2](value) {} })", optionsJson: "", fixedSource: stringPointer("({ set.2(value) {} })"), properties: []string{".2"}},
	{source: "({ async[.2]() {} })", optionsJson: "", fixedSource: stringPointer("({ async.2() {} })"), properties: []string{".2"}},
	{source: "({ [2]() {} })", optionsJson: "", fixedSource: stringPointer("({ 2() {} })"), properties: []string{"2"}},
	{source: "({ get [2]() {} })", optionsJson: "", fixedSource: stringPointer("({ get 2() {} })"), properties: []string{"2"}},
	{source: "({ set [2](value) {} })", optionsJson: "", fixedSource: stringPointer("({ set 2(value) {} })"), properties: []string{"2"}},
	{source: "({ async [2]() {} })", optionsJson: "", fixedSource: stringPointer("({ async 2() {} })"), properties: []string{"2"}},
	{source: "({ get[2]() {} })", optionsJson: "", fixedSource: stringPointer("({ get 2() {} })"), properties: []string{"2"}},
	{source: "({ set[2](value) {} })", optionsJson: "", fixedSource: stringPointer("({ set 2(value) {} })"), properties: []string{"2"}},
	{source: "({ async[2]() {} })", optionsJson: "", fixedSource: stringPointer("({ async 2() {} })"), properties: []string{"2"}},
	{source: "({ get['foo']() {} })", optionsJson: "", fixedSource: stringPointer("({ get'foo'() {} })"), properties: []string{"'foo'"}},
	{source: "({ *[2]() {} })", optionsJson: "", fixedSource: stringPointer("({ *2() {} })"), properties: []string{"2"}},
	{source: "({ async*[2]() {} })", optionsJson: "", fixedSource: stringPointer("({ async*2() {} })"), properties: []string{"2"}},
	{source: "({ ['constructor']: 1 })", optionsJson: "", fixedSource: stringPointer("({ 'constructor': 1 })"), properties: []string{"'constructor'"}},
	{source: "({ ['prototype']: 1 })", optionsJson: "", fixedSource: stringPointer("({ 'prototype': 1 })"), properties: []string{"'prototype'"}},
	{source: "class Foo { ['0']() {} }", optionsJson: "{\"enforceForClassMembers\":true}", fixedSource: stringPointer("class Foo { '0'() {} }"), properties: []string{"'0'"}},
	{source: "class Foo { ['0+1,234']() {} }", optionsJson: "{}", fixedSource: stringPointer("class Foo { '0+1,234'() {} }"), properties: []string{"'0+1,234'"}},
	{source: "class Foo { ['x']() {} }", optionsJson: "{}", fixedSource: stringPointer("class Foo { 'x'() {} }"), properties: []string{"'x'"}},
	{source: "class Foo { [/* this comment prevents a fix */ 'x']() {} }", optionsJson: "", fixedSource: nil, properties: []string{"'x'"}},
	{source: "class Foo { ['x' /* this comment also prevents a fix */]() {} }", optionsJson: "", fixedSource: nil, properties: []string{"'x'"}},
	{source: "class Foo { [('x')]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { 'x'() {} }"), properties: []string{"'x'"}},
	{source: "class Foo { *['x']() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { *'x'() {} }"), properties: []string{"'x'"}},
	{source: "class Foo { async ['x']() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { async 'x'() {} }"), properties: []string{"'x'"}},
	{source: "class Foo { get[.2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { get.2() {} }"), properties: []string{".2"}},
	{source: "class Foo { set[.2](value) {} }", optionsJson: "", fixedSource: stringPointer("class Foo { set.2(value) {} }"), properties: []string{".2"}},
	{source: "class Foo { async[.2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { async.2() {} }"), properties: []string{".2"}},
	{source: "class Foo { [2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { 2() {} }"), properties: []string{"2"}},
	{source: "class Foo { get [2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { get 2() {} }"), properties: []string{"2"}},
	{source: "class Foo { set [2](value) {} }", optionsJson: "", fixedSource: stringPointer("class Foo { set 2(value) {} }"), properties: []string{"2"}},
	{source: "class Foo { async [2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { async 2() {} }"), properties: []string{"2"}},
	{source: "class Foo { get[2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { get 2() {} }"), properties: []string{"2"}},
	{source: "class Foo { set[2](value) {} }", optionsJson: "", fixedSource: stringPointer("class Foo { set 2(value) {} }"), properties: []string{"2"}},
	{source: "class Foo { async[2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { async 2() {} }"), properties: []string{"2"}},
	{source: "class Foo { get['foo']() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { get'foo'() {} }"), properties: []string{"'foo'"}},
	{source: "class Foo { *[2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { *2() {} }"), properties: []string{"2"}},
	{source: "class Foo { async*[2]() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { async*2() {} }"), properties: []string{"2"}},
	{source: "class Foo { static ['constructor']() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { static 'constructor'() {} }"), properties: []string{"'constructor'"}},
	{source: "class Foo { ['prototype']() {} }", optionsJson: "", fixedSource: stringPointer("class Foo { 'prototype'() {} }"), properties: []string{"'prototype'"}},
	{source: "(class { ['x']() {} })", optionsJson: "", fixedSource: stringPointer("(class { 'x'() {} })"), properties: []string{"'x'"}},
	{source: "(class { ['__proto__']() {} })", optionsJson: "", fixedSource: stringPointer("(class { '__proto__'() {} })"), properties: []string{"'__proto__'"}},
	{source: "(class { static ['__proto__']() {} })", optionsJson: "", fixedSource: stringPointer("(class { static '__proto__'() {} })"), properties: []string{"'__proto__'"}},
	{source: "(class { static ['constructor']() {} })", optionsJson: "", fixedSource: stringPointer("(class { static 'constructor'() {} })"), properties: []string{"'constructor'"}},
	{source: "(class { ['prototype']() {} })", optionsJson: "", fixedSource: stringPointer("(class { 'prototype'() {} })"), properties: []string{"'prototype'"}},
	{source: "class Foo { ['0'] }", optionsJson: "", fixedSource: stringPointer("class Foo { '0' }"), properties: []string{"'0'"}},
	{source: "class Foo { ['0'] = 0 }", optionsJson: "", fixedSource: stringPointer("class Foo { '0' = 0 }"), properties: []string{"'0'"}},
	{source: "class Foo { static[0] }", optionsJson: "", fixedSource: stringPointer("class Foo { static 0 }"), properties: []string{"0"}},
	{source: "class Foo { ['#foo'] }", optionsJson: "", fixedSource: stringPointer("class Foo { '#foo' }"), properties: []string{"'#foo'"}},
	{source: "(class { ['__proto__'] })", optionsJson: "", fixedSource: stringPointer("(class { '__proto__' })"), properties: []string{"'__proto__'"}},
	{source: "(class { static ['__proto__'] })", optionsJson: "", fixedSource: stringPointer("(class { static '__proto__' })"), properties: []string{"'__proto__'"}},
	{source: "(class { ['prototype'] })", optionsJson: "", fixedSource: stringPointer("(class { 'prototype' })"), properties: []string{"'prototype'"}},
}

// TestNoUselessComputedKeyFires runs upstream's 65 invalid cases.
//
// Each asserts three things rather than one: that the finding appears, that its rendered message
// carries the key's SOURCE text, and that the repair writes exactly what upstream's `output` says.
// The last two are what a message-id assertion cannot see, and this rule is precisely the shape
// where that matters -- a fixer replacing a constructed span can be anchored correctly, report the
// right id, and still write the wrong bytes.
func TestNoUselessComputedKeyFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUselessComputedKeyFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUselessComputedKey, uselessComputedKeyFile,
				testCase.source, decodeUselessComputedKeyOptionsForTest(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.properties))
			for i := range wantIds {
				wantIds[i] = "unnecessarilyComputedProperty"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// Upstream asserts `data: { property }`, which renders into its message as
			// `Unnecessarily computed property [{{property}}] found.`. Assert the rendered prefix
			// exactly rather than with a substring test: a `strings.Contains` on an interpolated
			// value stays green when the interpolation doubles a character, which is a defect this
			// tree has actually shipped.
			for i, property := range testCase.properties {
				want := "Unnecessarily computed property [" + property + "] found."
				got := result.Diagnostics[i].Message.Description
				if !strings.HasPrefix(got, want) {
					t.Errorf("finding %d: message should begin %q, got %q", i, want, got)
				}
			}

			if testCase.fixedSource == nil {
				// Upstream's `output: null`: reported, deliberately not repaired. Reproducing the
				// decline is the assertion.
				for i, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d proposes %d fixes; upstream declines to fix this case",
							i, len(diagnostic.Fixes))
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, *testCase.fixedSource)
		})
	}
}

// TestNoUselessComputedKeyStaysSilent runs upstream's 31 valid cases.
//
// These are the false positives upstream already thought about, and several encode a distinction
// nothing else would suggest: the four reserved names, the position dependence between a static and
// an instance member, the bigint decline, and the plain `[x]` that computes something real.
func TestNoUselessComputedKeyStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUselessComputedKeyCleanCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUselessComputedKey,
				uselessComputedKeyFile, testCase.source,
				decodeUselessComputedKeyOptionsForTest(t, testCase.optionsJson)))
		})
	}
}

// noUselessComputedKeyAddedCases are shapes upstream's corpus cannot contain or does not write.
//
// Two families. The parenthesis and whitespace shapes exist because our parser keeps a node
// upstream's folds away, so no imported fixture can reach the code that handles it. The TypeScript
// shapes exist because upstream's corpus is JavaScript, and the gap between it and our tree is
// exactly TypeScript syntax.
//
// Every verdict below was measured by driving the installed rule with the typescript-eslint parser
// rather than reasoned about, and the ones that report were checked with a control in the same file
// so a zero could be told from a parse failure.
var noUselessComputedKeyAddedCases = []struct {
	name        string
	source      string
	optionsJson string
	// findings is how many times the rule reports.
	findings int
	// fixedSource is the whole rewritten file, or "" when no repair is proposed.
	fixedSource string
	// reason says why this case is here, since none of them came from the corpus.
	reason string
}{
	{
		name:        "a parenthesized key repairs through the parens",
		source:      "({ [(('x'))]: 0 });",
		findings:    1,
		fixedSource: "({ 'x': 0 });",
		reason: "Our parser keeps KindParenthesizedExpression and upstream's folds it away, so " +
			"the unwrap has to loop. Measured: upstream repairs the doubly-wrapped form to 'x' too.",
	},
	{
		name:     "a comment inside the parens still declines the repair",
		source:   "({ [(/* c */ 'x')]: 0 });",
		findings: 1,
		reason: "Upstream's commentsExistBetween spans bracket to bracket, so a comment inside " +
			"the parens declines as well. Measured against the installed rule, which reports and " +
			"leaves the source unchanged. This port reaches the same answer by scanning against " +
			"the OUTERMOST wrapper rather than the inner literal.",
	},
	{
		name:        "whitespace inside the brackets is absorbed by the repair",
		source:      "({ [ 'x' ]: 0 });",
		findings:    1,
		fixedSource: "({ 'x': 0 });",
		reason:      "The replaced span is bracket to bracket, so padding disappears with them.",
	},
	{
		name:        "a spaced accessor keeps exactly one space",
		source:      "({ get [ 2 ]() {} });",
		findings:    1,
		fixedSource: "({ get 2() {} });",
		reason: "The tokens are not flush, so no space is inserted, and the one already written " +
			"survives because it sits outside the bracket span. Measured upstream.",
	},
	{
		name:        "a parenthesized numeric key after an accessor still needs its space",
		source:      "({ get[(2)]() {} });",
		findings:    1,
		fixedSource: "({ get 2() {} });",
		reason: "The space decision reads the UNWRAPPED key's first character, so the parens must " +
			"not hide the digit. Measured upstream.",
	},
	{
		name:        "a parenthesized string key after an accessor needs no space",
		source:      "({ get[('x')]() {} });",
		findings:    1,
		fixedSource: "({ get'x'() {} });",
		reason:      "A quote cannot continue an identifier. Measured upstream.",
	},
	{
		name:   "an interface member is silent",
		source: "interface I { ['x']: number; }",
		reason: "typescript-eslint gives it TSPropertySignature, which upstream's create never " +
			"listens on. Measured with a reporting control in the same file.",
	},
	{
		name:   "a type literal member is silent",
		source: "type T = { ['x']: number };",
		reason: "Also TSPropertySignature.",
	},
	{
		name:   "an abstract method is silent",
		source: "abstract class Foo { abstract ['x'](): void; }",
		reason: "TSAbstractMethodDefinition, again a node type with no listener. Measured with a " +
			"reporting control in the same file.",
	},
	{
		name:   "an accessor property is silent",
		source: "class Foo { accessor ['x'] = 1; }",
		reason: "AccessorProperty, again a node type with no listener. Measured with a reporting " +
			"control in the same file.",
	},
	{
		name:        "an optional marker survives the repair",
		source:      "class Foo { ['x']?: number; }",
		findings:    1,
		fixedSource: "class Foo { 'x'?: number; }",
		reason: "This is the family that destroyed eight declarations elsewhere in this tree. The " +
			"marker sits OUTSIDE the bracket span, so it cannot be swallowed.",
	},
	{
		name:        "a definite marker and a type annotation survive the repair",
		source:      "class Foo { ['x']!: number; }",
		findings:    1,
		fixedSource: "class Foo { 'x'!: number; }",
		reason:      "Same span argument as the optional marker.",
	},
	{
		name:        "modifiers and a type annotation survive the repair",
		source:      "class Foo { protected readonly ['x']: number = 1; }",
		findings:    1,
		fixedSource: "class Foo { protected readonly 'x': number = 1; }",
		reason:      "Modifiers precede the brackets and the annotation follows them.",
	},
	{
		name:        "type parameters and a return type survive the repair",
		source:      "class Foo { ['x']<T>(a: T): T { return a; } }",
		findings:    1,
		fixedSource: "class Foo { 'x'<T>(a: T): T { return a; } }",
		reason: "A return annotation is what the no-arrow-function-lifecycle fixer dropped, because " +
			"it sits outside the parameter list. Here it sits outside the replaced span entirely.",
	},
	{
		name:        "a decorator survives the repair",
		source:      "class Foo { @dec ['x']() {} }",
		findings:    1,
		fixedSource: "class Foo { @dec 'x'() {} }",
		reason: "The token before the bracket is the decorator's `]`-less identifier end, which " +
			"cannot fuse with a quote. Measured upstream.",
	},
	{
		name:   "a bigint key is declined",
		source: "({ [99999999999999999n]: 0 });",
		reason: "Upstream deliberately leaves these alone; its own comment says browsers throw on " +
			"a bigint property name. Reproducing the gap rather than improving on it.",
	},
	{
		name:   "a template key is declined",
		source: "({ [`x`]: 0 });",
		reason: "A no-substitution template is a TemplateLiteral upstream rather than a Literal, " +
			"so it fails the `key.type !== 'Literal'` test. Our parser gives it its own kind and " +
			"the shelf's property.Name would accept it, so listing the two accepted kinds rather " +
			"than excluding kinds is what keeps this silent.",
	},
	{
		name:   "an as-expression key is declined and does not panic",
		source: "({ ['x' as string]: 0 });",
		reason: "Node.Text() PANICS on this shape, and a panic costs every rule its verdict on the " +
			"whole file rather than just this one. Declined by kind before any text is read. " +
			"Measured silent upstream too.",
	},
	{
		name:        "the option can be switched off for class members",
		source:      "class Foo { ['x']() {} }",
		optionsJson: `{"enforceForClassMembers":false}`,
		reason:      "The narrowing arm, reached through the shipped decoder.",
	},
	{
		name:        "switching off class members leaves object literals enforced",
		source:      "({ ['x']: 0 });",
		optionsJson: `{"enforceForClassMembers":false}`,
		findings:    1,
		fixedSource: "({ 'x': 0 });",
		reason: "The other half of the option, and the one that separates 'narrowed' from " +
			"'disabled'. A decoder defaulting the wrong way passes the case above and fails here.",
	},
	{
		name:        "an object method is judged as a property rather than as a method",
		source:      "({ ['constructor']() {} });",
		findings:    1,
		fixedSource: "({ 'constructor'() {} });",
		reason: "Upstream types an object literal's method as Property, not MethodDefinition, so " +
			"the constructor exemption does not apply to it. Measured: this reports upstream " +
			"while the class form does not. Routing it to the method arm would silence it.",
	},
	{
		name:   "an object method named __proto__ is exempt",
		source: "({ ['__proto__']() {} });",
		reason: "The other side of the same routing. Measured silent upstream.",
	},
}

// TestNoUselessComputedKeyAddedCases covers what upstream's corpus structurally cannot.
func TestNoUselessComputedKeyAddedCases(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUselessComputedKeyAddedCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUselessComputedKey, uselessComputedKeyFile,
				testCase.source, decodeUselessComputedKeyOptionsForTest(t, testCase.optionsJson))

			if len(result.Diagnostics) != testCase.findings {
				t.Fatalf("expected %d findings, got %d %v (%s)",
					testCase.findings, len(result.Diagnostics), result.MessageIds(), testCase.reason)
			}
			if testCase.findings == 0 {
				return
			}
			if testCase.fixedSource == "" {
				for i, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d proposes a fix; upstream declines here (%s)",
							i, testCase.reason)
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.fixedSource)
		})
	}
}

// TestNoUselessComputedKeySpan asserts where the finding points.
//
// Upstream reports the whole member rather than the key, so `({ ['x']: 0 })` points at
// `['x']: 0` and not at `'x'`. A message-id fixture cannot see this, and the choice between the two
// is exactly the kind that stays green over the wrong answer.
func TestNoUselessComputedKeySpan(t *testing.T) {
	t.Parallel()

	const source = "({ ['x']: 0 });"
	result := rule_testing.Run(t, NoUselessComputedKey, uselessComputedKeyFile, source)
	rule_testing.ExpectFindings(t, result, "unnecessarilyComputedProperty")

	// `Run` does not trim its input, so the literal above is what is on disk and can be sliced
	// directly. `RunTyped` would trim, and this rule needs no checker.
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "['x']: 0" {
		t.Errorf("finding should point at the whole member %q, got %q", "['x']: 0", reported)
	}
}

// TestNoUselessComputedKeyReportsRawTextNotCookedText pins the raw-versus-cooked distinction.
//
// The message carries `sourceCode.getText(key)` and the fixer writes `key.raw`, so an escape stays
// written as an escape in both. Reading `key.Text()` instead would cook `A` into `A`, which
// changes what the repair writes into the file and is invisible to every other assertion here:
// the id is the same, the count is the same, and the span is the same.
func TestNoUselessComputedKeyReportsRawTextNotCookedText(t *testing.T) {
	t.Parallel()

	const source = "({ ['\\u0041']: 0 });"
	result := rule_testing.Run(t, NoUselessComputedKey, uselessComputedKeyFile, source)
	rule_testing.ExpectFindings(t, result, "unnecessarilyComputedProperty")

	want := "Unnecessarily computed property ['\\u0041'] found."
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, want) {
		t.Errorf("message should begin %q, got %q", want, got)
	}
	rule_testing.ExpectFixedSource(t, result, "({ '\\u0041': 0 });")
}

// TestNoUselessComputedKeyDecoderDefaultsToEnforcingClassMembers pins the default that a generic
// decoder would invert.
//
// `enforceForClassMembers` defaults to TRUE, so nil options, an empty object, and an explicit true
// must all enforce, while only an explicit false narrows. A `DecodeOptionsInto` would answer the
// zero value for the first two and silently disagree with upstream on both.
func TestNoUselessComputedKeyDecoderDefaultsToEnforcingClassMembers(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		optionsJson string
		want        bool
	}{
		{"no options at all", "", true},
		{"an empty object", "{}", true},
		{"an explicit true", `{"enforceForClassMembers":true}`, true},
		{"an explicit false", `{"enforceForClassMembers":false}`, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var settings NoUselessComputedKeySettings
			if testCase.optionsJson == "" {
				settings = DefaultNoUselessComputedKeySettings()
			} else {
				decoded, err := DecodeNoUselessComputedKeyOptions(json.RawMessage(testCase.optionsJson))
				if err != nil {
					t.Fatalf("decoding %q: %v", testCase.optionsJson, err)
				}
				settings = decoded.(NoUselessComputedKeySettings)
			}
			if settings.EnforceForClassMembers != testCase.want {
				t.Errorf("enforceForClassMembers should be %v, got %v",
					testCase.want, settings.EnforceForClassMembers)
			}
		})
	}

	// A rule configured as a bare severity is handed nil options rather than a struct, and the
	// zero value narrows the rule instead of disabling it. This is the nil-options fallback the
	// brief calls for, asserted by bypassing the decoder entirely.
	result := rule_testing.RunWithOptions(t, NoUselessComputedKey, uselessComputedKeyFile,
		"class Foo { ['x']() {} }", nil)
	rule_testing.ExpectFindings(t, result, "unnecessarilyComputedProperty")
}

// TestNoUselessComputedKeyDeclinesToRepairARecoveredParse pins the bracket sanity check.
//
// The check reads as defensive coding and a mutation neutralising it survived every other fixture
// here, so its reachability was probed rather than argued. Error recovery synthesizes an
// unterminated computed name from source no well-formed file contains: across eleven malformed
// inputs, seven produced a `KindComputedPropertyName` whose token range does NOT end in `]`, with
// spans like `[`, `['x'`, and `['x' ()`.
//
// That matters because this rule's repair replaces bracket to bracket. Handed a span whose last
// character is not `]`, it would write the key's text over a stretch of source that is not the
// bracketed key, corrupting the file. The guard turns that into a report with no repair, which is
// the shape the brief calls for: report the finding, withhold the repair.
//
// No fixture asserting a message id can see this, because the id is identical either way. The
// assertion has to be on what the finding OFFERS.
func TestNoUselessComputedKeyDeclinesToRepairARecoveredParse(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"({ ['x' });",
		"({ ['x': 0 });",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessComputedKey, uselessComputedKeyFile, source)
			rule_testing.ExpectFindings(t, result, "unnecessarilyComputedProperty")
			for i, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d proposes %d fixes on a recovered parse; the bracket span "+
						"is not delimited by brackets, so any repair would corrupt the file",
						i, len(diagnostic.Fixes))
				}
			}
		})
	}
}

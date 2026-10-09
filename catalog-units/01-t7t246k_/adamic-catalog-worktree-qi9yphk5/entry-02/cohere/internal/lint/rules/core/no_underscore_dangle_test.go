package core

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noUnderscoreDangleFile is where the fixtures pretend to live.
const noUnderscoreDangleFile = "/repository/source/NoUnderscoreDangle.ts"

// noUnderscoreDangleExpectation is one finding, with its span and the identifier the message names.
//
// The identifier is asserted because all forty-six findings share one message id, so it is the only
// varying part, and the span is asserted because several destructuring cases report TWICE at the
// same span. A fixture checking ids and count alone cannot see either.
type noUnderscoreDangleExpectation struct {
	line       int
	column     int
	endLine    int
	endColumn  int
	identifier string
}

type noUnderscoreDangleCase struct {
	name        string
	source      string
	optionsJson string
	findings    []noUnderscoreDangleExpectation
	reason      string
}

// decodeNoUnderscoreDangleOptionsForTest routes a case's options through the SHIPPED decoder.
func decodeNoUnderscoreDangleOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeNoUnderscoreDangleOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", optionsJson, err)
	}
	return decoded
}

// noUnderscoreDangleOffsetOf converts a 1-based line and column into a byte offset.
func noUnderscoreDangleOffsetOf(t *testing.T, source string, line int, column int) int {
	t.Helper()
	offset := 0
	for current := 1; current < line; current++ {
		next := strings.IndexByte(source[offset:], '\n')
		if next < 0 {
			t.Fatalf("the fixture has no line %d", line)
		}
		offset += next + 1
	}
	return offset + column - 1
}

func runNoUnderscoreDangle(t *testing.T, testCase noUnderscoreDangleCase) rule_testing.Result {
	t.Helper()
	return rule_testing.RunWithOptions(t, NoUnderscoreDangle, noUnderscoreDangleFile,
		testCase.source, decodeNoUnderscoreDangleOptionsForTest(t, testCase.optionsJson))
}

// TestNoUnderscoreDangleStaysSilent runs upstream's whole `valid` list.
func TestNoUnderscoreDangleStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noUnderscoreDangleCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoUnderscoreDangle(t, testCase))
		})
	}
}

// TestNoUnderscoreDangleFires runs upstream's whole `invalid` list, asserting spans and identifiers.
func TestNoUnderscoreDangleFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noUnderscoreDangleReportingCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoUnderscoreDangle(t, testCase)

			wantIds := make([]string, 0, len(testCase.findings))
			for range testCase.findings {
				wantIds = append(wantIds, messageNoUnderscoreDangle.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := noUnderscoreDangleOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := noUnderscoreDangleOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d), got [%d,%d) %q",
						index, wantPos, wantEnd, diagnostic.Range.Pos(), diagnostic.Range.End(),
						testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				want := "Unexpected dangling '_' in '" + expected.identifier + "'."
				if !strings.HasPrefix(diagnostic.Message.Description, want) {
					t.Errorf("finding %d: expected the message to name %q, got %q",
						index, expected.identifier, diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestNoUnderscoreDangleDecoderAcceptsTheShapesTheConfigLayerDelivers crosses the config boundary.
//
// Every fixture above reaches the decoder with bytes this test file built; the config layer builds
// different bytes. `meta.schema` declares ONE element here, so the cohere spelling and upstream's
// coincide, and that is pinned rather than assumed.
//
// The unknown-key rows matter more for this rule than for most. There are nine keys, four beginning
// `allow` and two beginning `enforce`, and `allowInMethodNames` is a plausible thing to type for
// `enforceInMethodNames`. A decoder ignoring it leaves a project believing it turned an arm off
// when that arm was never on.
func TestNoUnderscoreDangleDecoderAcceptsTheShapesTheConfigLayerDelivers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		check   func(*testing.T, NoUnderscoreDangleOptions)
		wantErr bool
	}{
		{
			name: "absent options, which is what a bare \"error\" delivers",
			check: func(t *testing.T, o NoUnderscoreDangleOptions) {
				if o.EnforceInMethodNames || o.EnforceInClassFields {
					t.Error("both enforce flags default OFF, so the two extra arms are silent")
				}
				if o.AllowAfterThis || o.AllowAfterSuper || len(o.Allow) != 0 {
					t.Error("these allow flags default OFF, so nothing is exempt")
				}
				// The three that default TRUE arrive as nil and are resolved by the accessor, which
				// is the distinction a plain bool cannot carry.
				if o.AllowFunctionParams != nil || o.AllowInArrayDestructuring != nil ||
					o.AllowInObjectDestructuring != nil {
					t.Error("an absent tri-state option must stay nil, so `explicitly false` and " +
						"`absent, use upstream's true` remain distinguishable")
				}
				if !noUnderscoreDangleFlag(o.AllowFunctionParams, true) ||
					!noUnderscoreDangleFlag(o.AllowInArrayDestructuring, true) ||
					!noUnderscoreDangleFlag(o.AllowInObjectDestructuring, true) {
					t.Error("these three default TRUE upstream, so parameters and destructured " +
						"names are exempt out of the box")
				}
			},
		},
		{
			name: "the cohere spelling: one option object after the severity",
			raw:  `{"allow":["__proto__"],"allowAfterThis":true,"enforceInMethodNames":true}`,
			check: func(t *testing.T, o NoUnderscoreDangleOptions) {
				if strings.Join(o.Allow, ",") != "__proto__" || !o.AllowAfterThis ||
					!o.EnforceInMethodNames {
					t.Errorf("the three keys did not all arrive: %#v", o)
				}
			},
		},
		{
			name: "every one of the nine keys at once",
			raw: `{"allow":["x"],"allowAfterThis":true,"allowAfterSuper":true,` +
				`"allowAfterThisConstructor":true,"enforceInMethodNames":true,` +
				`"allowFunctionParams":true,"enforceInClassFields":true,` +
				`"allowInArrayDestructuring":true,"allowInObjectDestructuring":true}`,
			check: func(t *testing.T, o NoUnderscoreDangleOptions) {
				if !o.AllowAfterSuper || !o.AllowAfterThisConstructor || !o.EnforceInClassFields {
					t.Errorf("a boolean key was dropped: %#v", o)
				}
				if !noUnderscoreDangleFlag(o.AllowFunctionParams, false) ||
					!noUnderscoreDangleFlag(o.AllowInArrayDestructuring, false) ||
					!noUnderscoreDangleFlag(o.AllowInObjectDestructuring, false) {
					t.Errorf("a tri-state key was dropped: %#v", o)
				}
			},
		},
		{name: "an empty object exempts nothing", raw: `{}`},
		{
			name:    "a shape that is neither must error rather than decode to a default",
			raw:     `"allowAfterThis"`,
			wantErr: true,
		},
		{
			name:    "upstream's variadic spelling is not one this layer delivers",
			raw:     `[{"allowAfterThis":true}]`,
			wantErr: true,
		},
		{
			name:    "an unknown key is refused, matching upstream's additionalProperties:false",
			raw:     `{"allowInMethodNames":true}`,
			wantErr: true,
		},
		{
			name:    "a misspelled known key is refused rather than silently ignored",
			raw:     `{"allowAfterThat":true}`,
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoUnderscoreDangleOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected the decoder to refuse %s, got %#v", testCase.raw, decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			options, ok := decoded.(NoUnderscoreDangleOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than NoUnderscoreDangleOptions", decoded)
			}
			if testCase.check != nil {
				testCase.check(t, options)
			}
		})
	}
}

// TestNoUnderscoreDangleEveryOptionReachesTheRule proves each decoded flag changes a verdict.
//
// A decoder that round-trips and a rule that ignores what it decoded look identical from the test
// above. Each row here is one source under two configurations with opposite verdicts, which is the
// only shape that separates them. Nine options, nine pairs.
func TestNoUnderscoreDangleEveryOptionReachesTheRule(t *testing.T) {
	t.Parallel()
	cases := []struct {
		option        string
		source        string
		on            string
		reports       bool
		defaultsAllow bool
	}{
		{option: "allow", source: `var _foo = 1`, on: `{"allow":["_foo"]}`},
		{option: "allowAfterThis", source: `this._prop;`, on: `{"allowAfterThis":true}`},
		{option: "allowAfterSuper", source: `class A { constructor() { super._p; } }`,
			on: `{"allowAfterSuper":true}`},
		{option: "allowAfterThisConstructor", source: `this.constructor._bar`,
			on: `{"allowAfterThisConstructor":true}`},
		// These three default TRUE upstream, so the pair runs the other way: silent by default,
		// reporting when explicitly turned off. That direction is what proves an EXPLICIT false is
		// distinguishable from an absent key, which a plain bool field would collapse.
		{option: "allowFunctionParams", source: `function foo(_bar) {}`,
			on: `{"allowFunctionParams":false}`, defaultsAllow: true},
		{option: "allowInArrayDestructuring", source: `const [_bar] = xs`,
			on: `{"allowInArrayDestructuring":false}`, defaultsAllow: true},
		{option: "allowInObjectDestructuring", source: `const { _bar } = o`,
			on: `{"allowInObjectDestructuring":false}`, defaultsAllow: true},
		// These two run the other way: the flag turns an arm ON rather than exempting it.
		{option: "enforceInMethodNames", source: `class A { _m() {} }`,
			on: `{"enforceInMethodNames":true}`, reports: true},
		{option: "enforceInClassFields", source: `class A { _f = 1; }`,
			on: `{"enforceInClassFields":true}`, reports: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.option, func(t *testing.T) {
			t.Parallel()
			withOption := runNoUnderscoreDangle(t,
				noUnderscoreDangleCase{source: testCase.source, optionsJson: testCase.on})
			without := runNoUnderscoreDangle(t, noUnderscoreDangleCase{source: testCase.source})

			if testCase.defaultsAllow {
				// Silent by default because upstream's default exempts it; reporting once the
				// option is explicitly false.
				if len(without.Diagnostics) != 0 {
					t.Errorf("%s defaults TRUE upstream, so this must be silent without it, got %v",
						testCase.option, without.MessageIds())
				}
				if len(withOption.Diagnostics) == 0 {
					t.Errorf("an explicit false for %s did not reach the rule: still silent",
						testCase.option)
				}
				return
			}
			if testCase.reports {
				// An `enforce` flag: silent by default, reporting when set.
				if len(without.Diagnostics) != 0 {
					t.Errorf("%s defaults OFF, so this must be silent without it, got %v",
						testCase.option, without.MessageIds())
				}
				if len(withOption.Diagnostics) == 0 {
					t.Errorf("%s did not reach the rule: still silent with the flag set",
						testCase.option)
				}
				return
			}
			// An `allow` flag: reporting by default, silent when set.
			if len(without.Diagnostics) == 0 {
				t.Errorf("this source must report without %s, or the pair proves nothing",
					testCase.option)
			}
			if len(withOption.Diagnostics) != 0 {
				t.Errorf("%s did not reach the rule: still reporting %v with the flag set",
					testCase.option, withOption.MessageIds())
			}
		})
	}
}

// TestNoUnderscoreDangleNestedDestructuringSeparatesTheTwoFlags is the case that pins the walk.
//
// `allowInArrayDestructuring` and `allowInObjectDestructuring` are separate questions asked about
// the pattern that COINED each name, not about the declarator. A port testing "is this declarator
// destructured at all" collapses them into one flag and still passes most of the corpus.
//
// The distinguishing source coins two names under different pattern kinds in one declaration, and
// upstream's own corpus carries it. Verdicts measured against the installed 10.8.1 build.
func TestNoUnderscoreDangleNestedDestructuringSeparatesTheTwoFlags(t *testing.T) {
	t.Parallel()
	const source = `const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }`

	both := runNoUnderscoreDangle(t, noUnderscoreDangleCase{
		source:      source,
		optionsJson: `{"allowInArrayDestructuring":false,"allowInObjectDestructuring":false}`,
	})
	rule_testing.ExpectFindings(t, both,
		messageNoUnderscoreDangle.Id, messageNoUnderscoreDangle.Id)

	// `_bar` is coined by the ARRAY pattern and `_a` by the OBJECT one. Allowing arrays must
	// silence exactly one of the two, and it must be `_bar`.
	arraysAllowed := runNoUnderscoreDangle(t, noUnderscoreDangleCase{
		source:      source,
		optionsJson: `{"allowInArrayDestructuring":true,"allowInObjectDestructuring":false}`,
	})
	rule_testing.ExpectFindings(t, arraysAllowed, messageNoUnderscoreDangle.Id)
	if !strings.Contains(arraysAllowed.Diagnostics[0].Message.Description, "'_a'") {
		t.Errorf("allowing array patterns must leave the OBJECT-coined `_a`, got %q",
			arraysAllowed.Diagnostics[0].Message.Description)
	}

	// And the mirror, which is what shows the two flags are not one flag read twice.
	objectsAllowed := runNoUnderscoreDangle(t, noUnderscoreDangleCase{
		source:      source,
		optionsJson: `{"allowInArrayDestructuring":false,"allowInObjectDestructuring":true}`,
	})
	rule_testing.ExpectFindings(t, objectsAllowed, messageNoUnderscoreDangle.Id)
	if !strings.Contains(objectsAllowed.Diagnostics[0].Message.Description, "'_bar'") {
		t.Errorf("allowing object patterns must leave the ARRAY-coined `_bar`, got %q",
			objectsAllowed.Diagnostics[0].Message.Description)
	}
}

// TestNoUnderscoreDangleSpecialCasesAreScopedWhereUpstreamScopesThem pins the three exemptions.
//
// Each is exempt in a different set of places, and the corpus asserts each only where it applies, so
// a port widening or narrowing one passes. Verdicts measured against the installed build.
func TestNoUnderscoreDangleSpecialCasesAreScopedWhereUpstreamScopesThem(t *testing.T) {
	t.Parallel()
	t.Run("a lone underscore is clean in every arm", func(t *testing.T) {
		t.Parallel()
		for _, source := range []string{
			`var _ = require('underscore');`,
			`function _() {}`,
			`foo._;`,
			`const [_] = xs`,
		} {
			result := runNoUnderscoreDangle(t, noUnderscoreDangleCase{source: source})
			if len(result.Diagnostics) != 0 {
				t.Errorf("%q: `_` alone is exempt everywhere, because the test sits inside "+
					"hasDanglingUnderscore rather than beside one arm. Got %v",
					source, result.MessageIds())
			}
		}
	})

	t.Run("__proto__ is exempt in a member access only", func(t *testing.T) {
		t.Parallel()
		member := runNoUnderscoreDangle(t, noUnderscoreDangleCase{source: `foo.bar.__proto__;`})
		if len(member.Diagnostics) != 0 {
			t.Errorf("a member access to __proto__ is exempt, got %v", member.MessageIds())
		}

		// The other direction, which is the half a scoped exemption gets wrong.
		variable := runNoUnderscoreDangle(t, noUnderscoreDangleCase{source: `var __proto__ = 1;`})
		rule_testing.ExpectFindings(t, variable, messageNoUnderscoreDangle.Id)
	})

	t.Run("a private identifier renders with its hash", func(t *testing.T) {
		t.Parallel()
		result := runNoUnderscoreDangle(t, noUnderscoreDangleCase{
			source:      `class A { #_bar() {} }`,
			optionsJson: `{"enforceInMethodNames":true}`,
		})
		rule_testing.ExpectFindings(t, result, messageNoUnderscoreDangle.Id)
		if !strings.Contains(result.Diagnostics[0].Message.Description, "'#_bar'") {
			t.Errorf("expected the message to name '#_bar', got %q",
				result.Diagnostics[0].Message.Description)
		}
	})
}

// noUnderscoreDangleCleanCases are upstream's `valid` list, verbatim.
var noUnderscoreDangleCleanCases = []noUnderscoreDangleCase{
	{name: "upstream valid[0]", source: `var foo_bar = 1;`, optionsJson: ""},
	{name: "upstream valid[1]", source: `function foo_bar() {}`, optionsJson: ""},
	{name: "upstream valid[2]", source: `foo.bar.__proto__;`, optionsJson: ""},
	{name: "upstream valid[3]", source: `console.log(__filename); console.log(__dirname);`, optionsJson: ""},
	{name: "upstream valid[4]", source: `var _ = require('underscore');`, optionsJson: ""},
	{name: "upstream valid[5]", source: `var a = b._;`, optionsJson: ""},
	{name: "upstream valid[6]", source: `function foo(_bar) {}`, optionsJson: ""},
	{name: "upstream valid[7]", source: `function foo(bar_) {}`, optionsJson: ""},
	{name: "upstream valid[8]", source: `(function _foo() {})`, optionsJson: ""},
	{name: "upstream valid[9]", source: `function foo(_bar) {}`, optionsJson: `{}`},
	{name: "upstream valid[10]", source: `function foo( _bar = 0) {}`, optionsJson: ""},
	{name: "upstream valid[11]", source: `const foo = { onClick(_bar) { } }`, optionsJson: ""},
	{name: "upstream valid[12]", source: `const foo = { onClick(_bar = 0) { } }`, optionsJson: ""},
	{name: "upstream valid[13]", source: `const foo = (_bar) => {}`, optionsJson: ""},
	{name: "upstream valid[14]", source: `const foo = (_bar = 0) => {}`, optionsJson: ""},
	{name: "upstream valid[15]", source: `function foo( ..._bar) {}`, optionsJson: ""},
	{name: "upstream valid[16]", source: `const foo = (..._bar) => {}`, optionsJson: ""},
	{name: "upstream valid[17]", source: `const foo = { onClick(..._bar) { } }`, optionsJson: ""},
	{name: "upstream valid[18]", source: `export default function() {}`, optionsJson: ""},
	{name: "upstream valid[19]", source: `var _foo = 1`, optionsJson: `{"allow": ["_foo"]}`},
	{name: "upstream valid[20]", source: `var __proto__ = 1;`, optionsJson: `{"allow": ["__proto__"]}`},
	{name: "upstream valid[21]", source: `foo._bar;`, optionsJson: `{"allow": ["_bar"]}`},
	{name: "upstream valid[22]", source: `function _foo() {}`, optionsJson: `{"allow": ["_foo"]}`},
	{name: "upstream valid[23]", source: `this._bar;`, optionsJson: `{"allowAfterThis": true}`},
	{name: "upstream valid[24]", source: `class foo { constructor() { super._bar; } }`, optionsJson: `{"allowAfterSuper": true}`},
	{name: "upstream valid[25]", source: `class foo { _onClick() { } }`, optionsJson: ""},
	{name: "upstream valid[26]", source: `class foo { onClick_() { } }`, optionsJson: ""},
	{name: "upstream valid[27]", source: `const o = { _onClick() { } }`, optionsJson: ""},
	{name: "upstream valid[28]", source: `const o = { onClick_() { } }`, optionsJson: ""},
	{name: "upstream valid[29]", source: `const o = { _onClick() { } }`, optionsJson: `{"allow": ["_onClick"], "enforceInMethodNames": true}`},
	{name: "upstream valid[30]", source: `const o = { _foo: 'bar' }`, optionsJson: ""},
	{name: "upstream valid[31]", source: `const o = { foo_: 'bar' }`, optionsJson: ""},
	{name: "upstream valid[32]", source: `this.constructor._bar`, optionsJson: `{"allowAfterThisConstructor": true}`},
	{name: "upstream valid[33]", source: `const foo = { onClick(bar) { } }`, optionsJson: ""},
	{name: "upstream valid[34]", source: `const foo = (bar) => {}`, optionsJson: ""},
	{name: "upstream valid[35]", source: `function foo(_bar) {}`, optionsJson: `{"allowFunctionParams": true}`},
	{name: "upstream valid[36]", source: `function foo( _bar = 0) {}`, optionsJson: `{"allowFunctionParams": true}`},
	{name: "upstream valid[37]", source: `const foo = { onClick(_bar) { } }`, optionsJson: `{"allowFunctionParams": true}`},
	{name: "upstream valid[38]", source: `const foo = (_bar) => {}`, optionsJson: `{"allowFunctionParams": true}`},
	{name: "upstream valid[39]", source: `function foo(bar) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[40]", source: `const foo = { onClick(bar) { } }`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[41]", source: `const foo = (bar) => {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[42]", source: `function foo(_bar) {}`, optionsJson: `{"allowFunctionParams": false, "allow": ["_bar"]}`},
	{name: "upstream valid[43]", source: `const foo = { onClick(_bar) { } }`, optionsJson: `{"allowFunctionParams": false, "allow": ["_bar"]}`},
	{name: "upstream valid[44]", source: `const foo = (_bar) => {}`, optionsJson: `{"allowFunctionParams": false, "allow": ["_bar"]}`},
	{name: "upstream valid[45]", source: `function foo([_bar]) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[46]", source: `function foo([_bar] = []) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[47]", source: `function foo( { _bar }) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[48]", source: `function foo( { _bar = 0 } = {}) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[49]", source: `function foo(...[_bar]) {}`, optionsJson: `{"allowFunctionParams": false}`},
	{name: "upstream valid[50]", source: `const [_foo] = arr`, optionsJson: ""},
	{name: "upstream valid[51]", source: `const [_foo] = arr`, optionsJson: `{}`},
	{name: "upstream valid[52]", source: `const [_foo] = arr`, optionsJson: `{"allowInArrayDestructuring": true}`},
	{name: "upstream valid[53]", source: `const [foo, ...rest] = [1, 2, 3]`, optionsJson: `{"allowInArrayDestructuring": false}`},
	{name: "upstream valid[54]", source: `const [foo, _bar] = [1, 2, 3]`, optionsJson: `{"allowInArrayDestructuring": false, "allow": ["_bar"]}`},
	{name: "upstream valid[55]", source: `const { _foo } = obj`, optionsJson: ""},
	{name: "upstream valid[56]", source: `const { _foo } = obj`, optionsJson: `{}`},
	{name: "upstream valid[57]", source: `const { _foo } = obj`, optionsJson: `{"allowInObjectDestructuring": true}`},
	{name: "upstream valid[58]", source: `const { foo, bar: _bar } = { foo: 1, bar: 2 }`, optionsJson: `{"allowInObjectDestructuring": false, "allow": ["_bar"]}`},
	{name: "upstream valid[59]", source: `const { foo, _bar } = { foo: 1, _bar: 2 }`, optionsJson: `{"allowInObjectDestructuring": false, "allow": ["_bar"]}`},
	{name: "upstream valid[60]", source: `const { foo, _bar: bar } = { foo: 1, _bar: 2 }`, optionsJson: `{"allowInObjectDestructuring": false}`},
	{name: "upstream valid[61]", source: `class foo { _field; }`, optionsJson: ""},
	{name: "upstream valid[62]", source: `class foo { _field; }`, optionsJson: `{"enforceInClassFields": false}`},
	{name: "upstream valid[63]", source: `class foo { #_field; }`, optionsJson: ""},
	{name: "upstream valid[64]", source: `class foo { #_field; }`, optionsJson: `{"enforceInClassFields": false}`},
	{name: "upstream valid[65]", source: `class foo { _field; }`, optionsJson: `{}`},
	{name: "upstream valid[66]", source: `import foo from 'foo.json' with { _type: 'json' }`, optionsJson: ""},
	{name: "upstream valid[67]", source: `export * from 'foo.json' with { _type: 'json' }`, optionsJson: ""},
	{name: "upstream valid[68]", source: `export { default } from 'foo.json' with { _type: 'json' }`, optionsJson: ""},
	{name: "upstream valid[69]", source: `import('foo.json', { _with: { _type: 'json' } })`, optionsJson: ""},
	{name: "upstream valid[70]", source: `import('foo.json', { 'with': { _type: 'json' } })`, optionsJson: ""},
	{name: "upstream valid[71]", source: `import('foo.json', { _with: { _type } })`, optionsJson: ""},
}

// noUnderscoreDangleReportingCases are upstream's `invalid` list, verbatim.
//
// Upstream asserts a count and an identifier per case; the SPANS were measured by driving the
// installed 10.8.1 build. Several destructuring cases report twice at one span, which is the
// shape a per-identifier anchor would get wrong while passing every message-id assertion.
var noUnderscoreDangleReportingCases = []noUnderscoreDangleCase{
	{
		name:        "upstream invalid[0]",
		source:      `var _foo = 1`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 13, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[1]",
		source:      `var foo_ = 1`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 13, identifier: "foo_"},
		},
	},
	{
		name:        "upstream invalid[2]",
		source:      `function _foo() {}`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 19, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[3]",
		source:      `function foo_() {}`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 19, identifier: "foo_"},
		},
	},
	{
		name:        "upstream invalid[4]",
		source:      `var __proto__ = 1;`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 18, identifier: "__proto__"},
		},
	},
	{
		name:        "upstream invalid[5]",
		source:      `foo._bar;`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 9, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[6]",
		source:      `this._prop;`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 11, identifier: "_prop"},
		},
	},
	{
		name:        "upstream invalid[7]",
		source:      `class foo { constructor() { super._prop; } }`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 29, endLine: 1, endColumn: 40, identifier: "_prop"},
		},
	},
	{
		name:        "upstream invalid[8]",
		source:      `class foo { constructor() { this._prop; } }`,
		optionsJson: `{"allowAfterSuper": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 29, endLine: 1, endColumn: 39, identifier: "_prop"},
		},
	},
	{
		name:        "upstream invalid[9]",
		source:      `class foo { _onClick() { } }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 27, identifier: "_onClick"},
		},
	},
	{
		name:        "upstream invalid[10]",
		source:      `class foo { onClick_() { } }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 27, identifier: "onClick_"},
		},
	},
	{
		name:        "upstream invalid[11]",
		source:      `const o = { _onClick() { } }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 27, identifier: "_onClick"},
		},
	},
	{
		name:        "upstream invalid[12]",
		source:      `const o = { onClick_() { } }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 27, identifier: "onClick_"},
		},
	},
	{
		name:        "upstream invalid[13]",
		source:      `this.constructor._bar`,
		optionsJson: "",
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 22, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[14]",
		source:      `function foo(_bar) {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 18, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[15]",
		source:      `(function foo(_bar) {})`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 15, endLine: 1, endColumn: 19, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[16]",
		source:      `function foo(bar, _foo) {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 19, endLine: 1, endColumn: 23, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[17]",
		source:      `const foo = { onClick(_bar) { } }`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 23, endLine: 1, endColumn: 27, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[18]",
		source:      `const foo = (_bar) => {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 18, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[19]",
		source:      `function foo(_bar = 0) {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 22, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[20]",
		source:      `const foo = { onClick(_bar = 0) { } }`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 23, endLine: 1, endColumn: 31, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[21]",
		source:      `const foo = (_bar = 0) => {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 22, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[22]",
		source:      `function foo(..._bar) {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 21, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[23]",
		source:      `const foo = { onClick(..._bar) { } }`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 23, endLine: 1, endColumn: 30, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[24]",
		source:      `const foo = (..._bar) => {}`,
		optionsJson: `{"allowFunctionParams": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 14, endLine: 1, endColumn: 21, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[25]",
		source:      `const [foo, _bar] = [1, 2]`,
		optionsJson: `{"allowInArrayDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 27, identifier: "_bar"},
		},
	},
	{
		name:        "upstream invalid[26]",
		source:      `const [_foo = 1] = arr`,
		optionsJson: `{"allowInArrayDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 23, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[27]",
		source:      `const [foo, ..._rest] = [1, 2, 3]`,
		optionsJson: `{"allowInArrayDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 34, identifier: "_rest"},
		},
	},
	{
		name:        "upstream invalid[28]",
		source:      `const [foo, [bar_, baz]] = [1, [2, 3]]`,
		optionsJson: `{"allowInArrayDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 39, identifier: "bar_"},
		},
	},
	{
		name:        "upstream invalid[29]",
		source:      `const { _foo, bar } = { _foo: 1, bar: 2 }`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 42, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[30]",
		source:      `const { _foo = 1 } = obj`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 25, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[31]",
		source:      `const { bar: _foo = 1 } = obj`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 30, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[32]",
		source:      `const { foo: _foo, bar } = { foo: 1, bar: 2 }`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 46, identifier: "_foo"},
		},
	},
	{
		name:        "upstream invalid[33]",
		source:      `const { foo, ..._rest} = { foo: 1, bar: 2, baz: 3 }`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 52, identifier: "_rest"},
		},
	},
	{
		name:        "upstream invalid[34]",
		source:      `const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }`,
		optionsJson: `{"allowInArrayDestructuring": false, "allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 72, identifier: "_bar"},
			{line: 1, column: 7, endLine: 1, endColumn: 72, identifier: "_a"},
		},
	},
	{
		name:        "upstream invalid[35]",
		source:      `const { foo: [_bar, { a: _a, b } ] } = { foo: [1, { a: 'a', b: 'b' }] }`,
		optionsJson: `{"allowInArrayDestructuring": true, "allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 72, identifier: "_a"},
		},
	},
	{
		name:        "upstream invalid[36]",
		source:      `const [{ foo: [_bar, _, { bar: _baz }] }] = [{ foo: [1, 2, { bar: 'a' }] }]`,
		optionsJson: `{"allowInArrayDestructuring": false, "allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 76, identifier: "_bar"},
			{line: 1, column: 7, endLine: 1, endColumn: 76, identifier: "_baz"},
		},
	},
	{
		name:        "upstream invalid[37]",
		source:      `const { foo, bar: { baz, _qux } } = { foo: 1, bar: { baz: 3, _qux: 4 } }`,
		optionsJson: `{"allowInObjectDestructuring": false}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 7, endLine: 1, endColumn: 73, identifier: "_qux"},
		},
	},
	{
		name:        "upstream invalid[38]",
		source:      `class foo { #_bar() {} }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 23, identifier: "#_bar"},
		},
	},
	{
		name:        "upstream invalid[39]",
		source:      `class foo { #bar_() {} }`,
		optionsJson: `{"enforceInMethodNames": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 23, identifier: "#bar_"},
		},
	},
	{
		name:        "upstream invalid[40]",
		source:      `class foo { _field; }`,
		optionsJson: `{"enforceInClassFields": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 20, identifier: "_field"},
		},
	},
	{
		name:        "upstream invalid[41]",
		source:      `class foo { #_field; }`,
		optionsJson: `{"enforceInClassFields": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 21, identifier: "#_field"},
		},
	},
	{
		name:        "upstream invalid[42]",
		source:      `class foo { field_; }`,
		optionsJson: `{"enforceInClassFields": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 20, identifier: "field_"},
		},
	},
	{
		name:        "upstream invalid[43]",
		source:      `class foo { #field_; }`,
		optionsJson: `{"enforceInClassFields": true}`,
		findings: []noUnderscoreDangleExpectation{
			{line: 1, column: 13, endLine: 1, endColumn: 21, identifier: "#field_"},
		},
	},
}

// TestNoUnderscoreDangleThisConstructorIsThreeConditions pins each half of the receiver test.
//
// `allowAfterThisConstructor` exempts `this.constructor._bar` and nothing else, but upstream's
// corpus contains only the POSITIVE case, so a port that checked merely "the receiver is a member
// access on `this`" passes every imported fixture. A mutation dropping the `constructor` name check
// survived the whole 116-case corpus for exactly that reason.
//
// Three conditions, and each row below removes one. Every verdict measured against the installed
// 10.8.1 build.
func TestNoUnderscoreDangleThisConstructorIsThreeConditions(t *testing.T) {
	t.Parallel()
	const on = `{"allowAfterThisConstructor":true}`

	t.Run("this.constructor._bar is exempt", func(t *testing.T) {
		t.Parallel()
		result := runNoUnderscoreDangle(t,
			noUnderscoreDangleCase{source: `this.constructor._bar`, optionsJson: on})
		if len(result.Diagnostics) != 0 {
			t.Errorf("expected silence, got %v", result.MessageIds())
		}
	})

	reporting := []struct {
		name   string
		source string
		reason string
	}{
		{
			name:   "this.prototype._bar reports, because the property must be `constructor`",
			source: `this.prototype._bar`,
			reason: "this is the row a port checking only `a member access on this` gets wrong, " +
				"and it is the mutation that survived the imported corpus",
		},
		{
			name:   "this.foo._bar reports, for the same reason with an ordinary name",
			source: `this.foo._bar`,
			reason: "the name check is on the inner property, not on its shape",
		},
		{
			name:   "that.constructor._bar reports, because the receiver must be `this`",
			source: `that.constructor._bar`,
			reason: "the innermost object has to be a ThisExpression",
		},
		{
			name:   "this._bar reports, because the exemption is two levels deep",
			source: `this._bar`,
			reason: "a bare this-member is allowAfterThis's job, not this one's",
		},
		{
			name:   "this.a.b._bar reports, so the walk is not `anything under this`",
			source: `this.a.b._bar`,
			reason: "the receiver must be exactly this.constructor, not any chain rooted at this",
		},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoUnderscoreDangle(t,
				noUnderscoreDangleCase{source: testCase.source, optionsJson: on})
			if len(result.Diagnostics) == 0 {
				t.Errorf("expected a finding and got none. %s", testCase.reason)
			}
		})
	}
}

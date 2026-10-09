package core

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// preferDestructuringFile is where the fixtures pretend to live.
const preferDestructuringFile = "/repository/source/PreferDestructuring.ts"

// preferDestructuringExpectation is one finding, with its span and the destructuring kind named in
// the message.
//
// The kind is asserted because both arms share one message id and differ only in that word, so a
// port reporting "object" where upstream reports "array" passes every message-id assertion.
type preferDestructuringExpectation struct {
	line      int
	column    int
	endLine   int
	endColumn int
	kind      string
}

type preferDestructuringCase struct {
	name        string
	source      string
	optionsJson string
	// output is upstream's fixed source, or "" when upstream asserts `output: null` -- a case it
	// reports and deliberately declines to repair.
	output   string
	findings []preferDestructuringExpectation
	reason   string
}

// decodePreferDestructuringOptionsForTest routes a case's options through the SHIPPED decoder.
func decodePreferDestructuringOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodePreferDestructuringOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", optionsJson, err)
	}
	return decoded
}

// preferDestructuringOffsetOf converts a 1-based line and column into a byte offset.
func preferDestructuringOffsetOf(t *testing.T, source string, line int, column int) int {
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

func runPreferDestructuring(t *testing.T, testCase preferDestructuringCase) rule_testing.Result {
	t.Helper()
	return rule_testing.RunWithOptions(t, PreferDestructuring, preferDestructuringFile,
		testCase.source, decodePreferDestructuringOptionsForTest(t, testCase.optionsJson))
}

// TestPreferDestructuringStaysSilent runs upstream's whole `valid` list.
func TestPreferDestructuringStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range preferDestructuringCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runPreferDestructuring(t, testCase))
		})
	}
}

// TestPreferDestructuringFires runs upstream's whole `invalid` list, asserting spans and kinds.
func TestPreferDestructuringFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range preferDestructuringReportingCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, testCase)

			wantIds := make([]string, 0, len(testCase.findings))
			for range testCase.findings {
				wantIds = append(wantIds, messagePreferDestructuring.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := preferDestructuringOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := preferDestructuringOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d), got [%d,%d) %q",
						index, wantPos, wantEnd, diagnostic.Range.Pos(), diagnostic.Range.End(),
						testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				want := "Use " + expected.kind + " destructuring."
				if !strings.HasPrefix(diagnostic.Message.Description, want) {
					t.Errorf("finding %d: expected %q, got %q",
						index, want, diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestPreferDestructuringRewrites asserts the eighteen `output` vectors upstream ships.
//
// These are before-and-after pairs on the whole file rather than message ids, so they catch a fixer
// that repairs the right span with the wrong text. Fourteen of the eighteen are about comments.
func TestPreferDestructuringRewrites(t *testing.T) {
	t.Parallel()
	asserted := 0
	for _, testCase := range preferDestructuringReportingCases {
		if testCase.output == "" {
			continue
		}
		asserted++
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runPreferDestructuring(t, testCase), testCase.output)
		})
	}
	// A table-driven fix assertion that silently matched nothing would pass while checking no
	// rewrite at all, which is the shape this repository keeps finding.
	if asserted != 18 {
		t.Fatalf("expected upstream's eighteen fix vectors, ran %d", asserted)
	}
}

// TestPreferDestructuringDeclinesEveryRepairUpstreamDeclines is the other half of the fixer.
//
// Thirty of upstream's forty-eight reporting cases carry `output: null`: reported, deliberately not
// repaired. A fixer that repaired any of them would pass every message-id assertion and every one of
// the eighteen rewrite vectors, because none of those cases is in either set.
//
// The count is asserted for the same reason as above: a loop that matched nothing would pass.
func TestPreferDestructuringDeclinesEveryRepairUpstreamDeclines(t *testing.T) {
	t.Parallel()
	declined := 0
	for _, testCase := range preferDestructuringReportingCases {
		if testCase.output != "" {
			continue
		}
		declined++
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, testCase)
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d proposed a repair upstream declines (output: null): %q",
						index, diagnostic.Fixes[0].Text)
				}
			}
		})
	}
	if declined != 30 {
		t.Fatalf("expected upstream's thirty declines, checked %d", declined)
	}
}

// TestPreferDestructuringDecoderReadsBothSchemaElements is the config-boundary contract.
//
// Upstream's `meta.schema` has TWO elements and reads `context.options[1].enforceForRenamedProperties`.
// The config layer used to keep `tuple[1]` and drop the rest without an error, so a config written
// the upstream way ran with that option permanently off while the file said otherwise. The rule now
// registers with `DecodeOptionList`, and these rows are the list shapes the config layer delivers.
//
// No fixture can prove the config layer delivers them, because every fixture hands the decoder bytes
// this file built. `TestPreferDestructuringSecondElementCrossesTheConfigBoundary` is the one that
// goes through the real layer.
func TestPreferDestructuringDecoderReadsBothSchemaElements(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		check   func(*testing.T, PreferDestructuringOptions)
		wantErr bool
	}{
		{
			name: "absent options turn every arm on, which is upstream's default",
			check: func(t *testing.T, o PreferDestructuringOptions) {
				if !o.VariableDeclarator.enabled(true) || !o.VariableDeclarator.enabled(false) ||
					!o.AssignmentExpression.enabled(true) || !o.AssignmentExpression.enabled(false) {
					t.Errorf("all four switches default on: %#v", o)
				}
				if o.EnforceForRenamedProperties {
					t.Error("enforceForRenamedProperties defaults off")
				}
			},
		},
		{
			name: "enforceForRenamedProperties arrives as the SECOND element, upstream's spelling",
			raw:  `[{"VariableDeclarator":{"object":true}},{"enforceForRenamedProperties":true}]`,
			check: func(t *testing.T, o PreferDestructuringOptions) {
				if !o.EnforceForRenamedProperties {
					t.Error("the second schema element was handed over and not read")
				}
				if !o.VariableDeclarator.enabled(false) {
					t.Error("the enabling keys must survive alongside it")
				}
				if o.AssignmentExpression.enabled(false) {
					t.Error("naming only VariableDeclarator disables the other kind, which is " +
						"upstream replacing normalizedOptions wholesale")
				}
			},
		},
		{
			name: "the flat spelling applies to both node kinds",
			raw:  `[{"object":true}]`,
			check: func(t *testing.T, o PreferDestructuringOptions) {
				if !o.VariableDeclarator.enabled(false) || !o.AssignmentExpression.enabled(false) {
					t.Error("a top-level object/array key is upstream's flat form and covers both")
				}
				if o.VariableDeclarator.enabled(true) || o.AssignmentExpression.enabled(true) {
					t.Error("the flat form replaces both pairs, so an unnamed switch is OFF")
				}
				if o.EnforceForRenamedProperties {
					t.Error("enforceForRenamedProperties defaults off when the second element is absent")
				}
			},
		},
		{
			name: "an explicit false in the second element stays false",
			raw:  `[{"object":true},{"enforceForRenamedProperties":false}]`,
			check: func(t *testing.T, o PreferDestructuringOptions) {
				if o.EnforceForRenamedProperties {
					t.Error("an explicit false was read as true")
				}
			},
		},
		{
			// The workaround spelling from when the config layer could deliver one element. It was
			// never upstream's, and upstream's `additionalProperties: false` refuses the key there.
			name:    "enforceForRenamedProperties inside the FIRST element is refused",
			raw:     `[{"object":true,"enforceForRenamedProperties":true}]`,
			wantErr: true,
		},
		{
			name:    "a third element is refused rather than dropped",
			raw:     `[{"object":true},{"enforceForRenamedProperties":true},{"object":false}]`,
			wantErr: true,
		},
		{
			name:    "a bare object, which the config layer never delivers to a list rule, is refused",
			raw:     `{"object":true}`,
			wantErr: true,
		},
		{
			name:    "a first element that is not an object must error rather than decode to a default",
			raw:     `["object"]`,
			wantErr: true,
		},
		{
			name:    "an unknown key in the second element is refused, matching additionalProperties:false",
			raw:     `[{"object":true},{"enforceForRenamedProperty":true}]`,
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodePreferDestructuringOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected the decoder to refuse %s, got %#v", testCase.raw, decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			options, ok := decoded.(PreferDestructuringOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than PreferDestructuringOptions", decoded)
			}
			if testCase.check != nil {
				testCase.check(t, options)
			}
		})
	}
}

// TestPreferDestructuringEnforceForRenamedPropertiesReachesTheRule is the moving control.
//
// A decoder that round-trips and a rule that ignores what it decoded look identical from the test
// above. One source, two configurations, opposite verdicts.
func TestPreferDestructuringEnforceForRenamedPropertiesReachesTheRule(t *testing.T) {
	t.Parallel()
	const source = `var foobar = object.bar;`

	off := runPreferDestructuring(t, preferDestructuringCase{source: source})
	if len(off.Diagnostics) != 0 {
		t.Errorf("the names differ, so nothing is repeated and this is clean by default, got %v",
			off.MessageIds())
	}

	on := runPreferDestructuring(t, preferDestructuringCase{
		source:      source,
		optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`,
	})
	rule_testing.ExpectFindings(t, on, messagePreferDestructuring.Id)
}

// preferDestructuringCleanCases are upstream's `valid` list, verbatim.
//
// The `optionsJson` is upstream's option list as written, both schema elements where upstream passes
// both: thirty of upstream's cases pass a second element, which the config layer used to discard.
var preferDestructuringCleanCases = []preferDestructuringCase{
	{name: "upstream valid[0]", source: `var [foo] = array;`, optionsJson: ""},
	{name: "upstream valid[1]", source: `var { foo } = object;`, optionsJson: ""},
	{name: "upstream valid[2]", source: `var foo;`, optionsJson: ""},
	{name: "upstream valid[3]", source: `var foo = object.bar;`, optionsJson: `[{"VariableDeclarator": {"object": true}}]`},
	{name: "upstream valid[4]", source: `var foo = object.bar;`, optionsJson: `[{"object": true}]`},
	{name: "upstream valid[5]", source: `var foo = object.bar;`, optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[6]", source: `var foo = object.bar;`, optionsJson: `[{"object": true}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[7]", source: `var foo = object['bar'];`, optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[8]", source: `var foo = object[bar];`, optionsJson: `[{"object": true}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[9]", source: `var { bar: foo } = object;`, optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[10]", source: `var { bar: foo } = object;`, optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[11]", source: `var { [bar]: foo } = object;`, optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[12]", source: `var { [bar]: foo } = object;`, optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[13]", source: `var foo = array[0];`, optionsJson: `[{"VariableDeclarator": {"array": false}}]`},
	{name: "upstream valid[14]", source: `var foo = array[0];`, optionsJson: `[{"array": false}]`},
	{name: "upstream valid[15]", source: `var foo = object.foo;`, optionsJson: `[{"VariableDeclarator": {"object": false}}]`},
	{name: "upstream valid[16]", source: `var foo = object['foo'];`, optionsJson: `[{"VariableDeclarator": {"object": false}}]`},
	{name: "upstream valid[17]", source: `({ foo } = object);`, optionsJson: ""},
	{name: "upstream valid[18]", source: `var foo = array[0];`, optionsJson: `[{"VariableDeclarator": {"array": false}}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[19]", source: `var foo = array[0];`, optionsJson: `[{"array": false}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[20]", source: `[foo] = array;`, optionsJson: ""},
	{name: "upstream valid[21]", source: `foo += array[0]`, optionsJson: ""},
	{name: "upstream valid[22]", source: `foo &&= array[0]`, optionsJson: ""},
	{name: "upstream valid[23]", source: `foo += bar.foo`, optionsJson: ""},
	{name: "upstream valid[24]", source: `foo ||= bar.foo`, optionsJson: ""},
	{name: "upstream valid[25]", source: `foo ??= bar['foo']`, optionsJson: ""},
	{name: "upstream valid[26]", source: `foo = object.foo;`, optionsJson: `[{"AssignmentExpression": {"object": false}}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[27]", source: `foo = object.foo;`, optionsJson: `[{"AssignmentExpression": {"object": false}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[28]", source: `foo = array[0];`, optionsJson: `[{"AssignmentExpression": {"array": false}}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[29]", source: `foo = array[0];`, optionsJson: `[{"AssignmentExpression": {"array": false}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[30]", source: `foo = array[0];`, optionsJson: `[{"VariableDeclarator": {"array": true}, "AssignmentExpression": {"array": false}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[31]", source: `var foo = array[0];`, optionsJson: `[{"VariableDeclarator": {"array": false}, "AssignmentExpression": {"array": true}}, {"enforceForRenamedProperties": false}]`},
	{name: "upstream valid[32]", source: `foo = object.foo;`, optionsJson: `[{"VariableDeclarator": {"object": true}, "AssignmentExpression": {"object": false}}]`},
	{name: "upstream valid[33]", source: `var foo = object.foo;`, optionsJson: `[{"VariableDeclarator": {"object": false}, "AssignmentExpression": {"object": true}}]`},
	{name: "upstream valid[34]", source: `class Foo extends Bar { static foo() {var foo = super.foo} }`, optionsJson: ""},
	{name: "upstream valid[35]", source: `foo = bar[foo];`, optionsJson: ""},
	{name: "upstream valid[36]", source: `var foo = bar[foo];`, optionsJson: ""},
	{name: "upstream valid[37]", source: `var {foo: {bar}} = object;`, optionsJson: `[{"object": true}]`},
	{name: "upstream valid[38]", source: `var {bar} = object.foo;`, optionsJson: `[{"object": true}]`},
	{name: "upstream valid[39]", source: `var foo = array?.[0];`, optionsJson: ""},
	{name: "upstream valid[40]", source: `var foo = object?.foo;`, optionsJson: ""},
	{name: "upstream valid[41]", source: `class C { #x; foo() { const x = this.#x; } }`, optionsJson: ""},
	{name: "upstream valid[42]", source: `class C { #x; foo() { x = this.#x; } }`, optionsJson: ""},
	{name: "upstream valid[43]", source: `class C { #x; foo(a) { x = a.#x; } }`, optionsJson: ""},
	{name: "upstream valid[44]", source: `class C { #x; foo() { const x = this.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[45]", source: `class C { #x; foo() { const y = this.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[46]", source: `class C { #x; foo() { x = this.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[47]", source: `class C { #x; foo() { y = this.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[48]", source: `class C { #x; foo(a) { x = a.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[49]", source: `class C { #x; foo(a) { y = a.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[50]", source: `class C { #x; foo() { x = this.a.#x; } }`, optionsJson: `[{"array": true, "object": true}, {"enforceForRenamedProperties": true}]`},
	{name: "upstream valid[51]", source: `using foo = array[0];`, optionsJson: ""},
	{name: "upstream valid[52]", source: `using foo = object.foo;`, optionsJson: ""},
	{name: "upstream valid[53]", source: `await using foo = array[0];`, optionsJson: ""},
	{name: "upstream valid[54]", source: `await using foo = object.foo;`, optionsJson: ""},
}

// preferDestructuringReportingCases are upstream's `invalid` list, verbatim.
//
// `output` is upstream's fixed source where it asserts one, and "" where it asserts `null`,
// meaning the case is reported and deliberately NOT repaired. Thirty of the forty-eight are
// declines, and they are the fixer's real specification.
var preferDestructuringReportingCases = []preferDestructuringCase{
	{
		name:        "upstream invalid[0]",
		source:      `var foo = array[0];`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 19, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[1]",
		source:      `foo = array[0];`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 15, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[2]",
		source:      `var foo = object.foo;`,
		optionsJson: "",
		output:      `var {foo} = object;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[3]",
		source:      `var foo = (a, b).foo;`,
		optionsJson: "",
		output:      `var {foo} = (a, b);`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[4]",
		source:      `var length = (() => {}).length;`,
		optionsJson: "",
		output:      `var {length} = () => {};`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 31, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[5]",
		source:      `var foo = (a = b).foo;`,
		optionsJson: "",
		output:      `var {foo} = a = b;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 22, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[6]",
		source:      `var foo = (a || b).foo;`,
		optionsJson: "",
		output:      `var {foo} = a || b;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 23, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[7]",
		source:      `var foo = (f()).foo;`,
		optionsJson: "",
		output:      `var {foo} = f();`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 20, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[8]",
		source:      `var foo = object.bar.foo;`,
		optionsJson: "",
		output:      `var {foo} = object.bar;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 25, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[9]",
		source:      `var foobar = object.bar;`,
		optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 24, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[10]",
		source:      `var foobar = object.bar;`,
		optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 24, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[11]",
		source:      `var foo = object[bar];`,
		optionsJson: `[{"VariableDeclarator": {"object": true}}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 22, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[12]",
		source:      `var foo = object[bar];`,
		optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 22, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[13]",
		source:      `var foo = object[foo];`,
		optionsJson: `[{"object": true}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 22, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[14]",
		source:      `var foo = object['foo'];`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 24, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[15]",
		source:      `foo = object.foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 17, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[16]",
		source:      `foo = object['foo'];`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 20, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[17]",
		source:      `var foo = array[0];`,
		optionsJson: `[{"VariableDeclarator": {"array": true}}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 19, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[18]",
		source:      `foo = array[0];`,
		optionsJson: `[{"AssignmentExpression": {"array": true}}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 15, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[19]",
		source:      `var foo = array[0];`,
		optionsJson: `[{"VariableDeclarator": {"array": true}, "AssignmentExpression": {"array": false}}, {"enforceForRenamedProperties": true}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 19, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[20]",
		source:      `var foo = array[0];`,
		optionsJson: `[{"VariableDeclarator": {"array": true}, "AssignmentExpression": {"array": false}}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 19, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[21]",
		source:      `foo = array[0];`,
		optionsJson: `[{"VariableDeclarator": {"array": false}, "AssignmentExpression": {"array": true}}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 15, kind: "array"},
		},
	},
	{
		name:        "upstream invalid[22]",
		source:      `foo = object.foo;`,
		optionsJson: `[{"VariableDeclarator": {"array": true, "object": false}, "AssignmentExpression": {"object": true}}]`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 17, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[23]",
		source:      `class Foo extends Bar { static foo() {var bar = super.foo.bar} }`,
		optionsJson: "",
		output:      `class Foo extends Bar { static foo() {var {bar} = super.foo} }`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 43, endLine: 1, endColumn: 62, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[24]",
		source:      `var /* comment */ foo = object.foo;`,
		optionsJson: "",
		output:      `var /* comment */ {foo} = object;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 19, endLine: 1, endColumn: 35, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[25]",
		source:      `var a, /* comment */foo = object.foo;`,
		optionsJson: "",
		output:      `var a, /* comment */{foo} = object;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 21, endLine: 1, endColumn: 37, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[26]",
		source:      `var foo /* comment */ = object.foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 35, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[27]",
		source:      `var a, foo /* comment */ = object.foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 8, endLine: 1, endColumn: 38, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[28]",
		source:      `var foo /* comment */ = object.foo, a;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 35, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[29]",
		source:      "var foo // comment\n = object.foo;",
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 2, endColumn: 14, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[30]",
		source:      `var foo = /* comment */ object.foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 35, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[31]",
		source:      "var foo = // comment\n object.foo;",
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 2, endColumn: 12, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[32]",
		source:      `var foo = (/* comment */ object).foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 37, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[33]",
		source:      `var foo = (object /* comment */).foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 37, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[34]",
		source:      `var foo = bar(/* comment */).foo;`,
		optionsJson: "",
		output:      `var {foo} = bar(/* comment */);`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 33, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[35]",
		source:      `var foo = bar/* comment */.baz.foo;`,
		optionsJson: "",
		output:      `var {foo} = bar/* comment */.baz;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 35, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[36]",
		source:      "var foo = bar[// comment\nbaz].foo;",
		optionsJson: "",
		output:      "var {foo} = bar[// comment\nbaz];",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 2, endColumn: 9, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[37]",
		source:      "var foo // comment\n = bar(/* comment */).foo;",
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 2, endColumn: 26, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[38]",
		source:      `var foo = bar/* comment */.baz/* comment */.foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 48, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[39]",
		source:      "var foo = object// comment\n.foo;",
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 2, endColumn: 5, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[40]",
		source:      `var foo = object./* comment */foo;`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 34, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[41]",
		source:      `var foo = (/* comment */ object.foo);`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 37, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[42]",
		source:      `var foo = (object.foo /* comment */);`,
		optionsJson: "",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 37, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[43]",
		source:      `var foo = object.foo/* comment */;`,
		optionsJson: "",
		output:      `var {foo} = object/* comment */;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[44]",
		source:      `var foo = object.foo// comment`,
		optionsJson: "",
		output:      `var {foo} = object// comment`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[45]",
		source:      `var foo = object.foo/* comment */, a;`,
		optionsJson: "",
		output:      `var {foo} = object/* comment */, a;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[46]",
		source:      "var foo = object.foo// comment\n, a;",
		optionsJson: "",
		output:      "var {foo} = object// comment\n, a;",
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
	{
		name:        "upstream invalid[47]",
		source:      `var foo = object.foo, /* comment */ a;`,
		optionsJson: "",
		output:      `var {foo} = object, /* comment */ a;`,
		findings: []preferDestructuringExpectation{
			{line: 1, column: 5, endLine: 1, endColumn: 21, kind: "object"},
		},
	},
}

// TestPreferDestructuringArrayIndexIsIntegerByVALUE pins the array arm's discriminator.
//
// Upstream tests `Number.isInteger(node.property.value)`, which is a question about the VALUE rather
// than the spelling. Its corpus writes only `array[0]`, so a mutation loosening the digit test
// survives every imported case.
//
// Our parser has already canonicalised the numeric text, which is what makes a text test correct
// here rather than merely convenient: `0x1` arrives as `1`, `1e2` as `100`, and `1.0` as `1`, so the
// canonical forms are exactly the integers and the non-integers keep a `.` or a `-`. Every verdict
// below was measured against the installed 10.8.1 build.
func TestPreferDestructuringArrayIndexIsIntegerByVALUE(t *testing.T) {
	t.Parallel()
	reporting := []struct {
		name   string
		source string
		reason string
	}{
		{"a plain integer", `var foo = array[0];`, "the corpus case"},
		{"a hex integer", `var foo = array[0x1];`,
			"0x1 is the integer 1, and our parser canonicalises the text to \"1\""},
		{"an exponent that is an integer", `var foo = array[1e2];`,
			"1e2 is 100, canonicalised to \"100\""},
		{"a float that is an integer", `var foo = array[1.0];`,
			"1.0 is the integer 1, canonicalised to \"1\". This is the row that shows the test is " +
				"on the value rather than on whether the source contains a dot."},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name+" reports", func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, preferDestructuringCase{source: testCase.source})
			rule_testing.ExpectFindings(t, result, messagePreferDestructuring.Id)
			if !strings.HasPrefix(result.Diagnostics[0].Message.Description, "Use array") {
				t.Errorf("expected the ARRAY arm, got %q. %s",
					result.Diagnostics[0].Message.Description, testCase.reason)
			}
		})
	}

	clean := []struct {
		name   string
		source string
		reason string
	}{
		{"a non-integer float", `var foo = array[1.5];`, "1.5 is not an integer"},
		{"a negative exponent", `var foo = array[1e-2];`, "1e-2 is 0.01"},
		{"a BigInt", `var foo = array[0n];`,
			"a BigInt literal has no numeric `.value`, so Number.isInteger is false"},
		{"a string index", `var foo = array['0'];`,
			"a string is not a number, and the names differ so the object arm declines too"},
		{"a negative index", `var foo = array[-1];`,
			"the minus is a unary operator rather than part of the literal, so the property is not " +
				"a Literal at all"},
	}
	for _, testCase := range clean {
		t.Run(testCase.name+" is clean", func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, preferDestructuringCase{source: testCase.source})
			if len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v. %s", result.MessageIds(), testCase.reason)
			}
		})
	}
}

// TestPreferDestructuringOptionalChainingIsExcluded pins a divergence our parser creates.
//
// In ESTree `object?.foo` is a ChainExpression wrapping the member access, so upstream's
// `rightNode.type !== "MemberExpression"` guard declines it and both of its optional-chaining cases
// are VALID. Our parser keeps one node kind and marks the token, so without an explicit exclusion
// this rule reports two of upstream's own passing cases.
//
// Both directions are asserted, because a guard that declined ALL property access would also pass
// the two clean rows.
func TestPreferDestructuringOptionalChainingIsExcluded(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`var foo = object?.foo;`, `var foo = array?.[0];`} {
		t.Run(source+" is clean", func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, preferDestructuringCase{source: source})
			if len(result.Diagnostics) != 0 {
				t.Errorf("optional chaining is excluded upstream, got %v", result.MessageIds())
			}
		})
	}

	// The control: the same shapes without the `?.` must still report, or the two rows above pass
	// for the wrong reason.
	for _, source := range []string{`var foo = object.foo;`, `var foo = array[0];`} {
		t.Run(source+" still reports", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				runPreferDestructuring(t, preferDestructuringCase{source: source}),
				messagePreferDestructuring.Id)
		})
	}
}

// TestPreferDestructuringWorksOnConstAndLet is the regression no imported fixture could carry.
//
// Upstream's whole corpus is written in `var`, so every one of its 103 cases passed while this rule
// was skipping every `const` declaration in the tree. The cause was a composite node flag:
// `NodeFlagsAwaitUsing = NodeFlagsConst | NodeFlagsUsing`, so a `using`-declaration guard written as
// a mask against it matched all consts.
//
// Found by a dry run through the real config layer against a seeded tree, not by a test. This makes
// it a test.
func TestPreferDestructuringWorksOnConstAndLet(t *testing.T) {
	t.Parallel()
	for _, keyword := range []string{"var", "let", "const"} {
		t.Run(keyword+" reports", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runPreferDestructuring(t,
				preferDestructuringCase{source: keyword + ` foo = object.foo;`}),
				messagePreferDestructuring.Id)
		})
		t.Run(keyword+" is repaired", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runPreferDestructuring(t,
				preferDestructuringCase{source: keyword + ` foo = object.foo;`}),
				keyword+` {foo} = object;`)
		})
	}

	// The exemption this guard exists for must still hold, in both spellings, or the fix above
	// would just be "delete the guard". A destructured `using` is a parse error, so a finding there
	// would be unactionable.
	for _, source := range []string{
		`using foo = object.foo;`,
		`async function f() { await using foo = object.foo; }`,
	} {
		t.Run("using stays exempt: "+source, func(t *testing.T) {
			t.Parallel()
			result := runPreferDestructuring(t, preferDestructuringCase{source: source})
			if len(result.Diagnostics) != 0 {
				t.Errorf("a `using` declaration cannot be destructured, so it must be exempt, got %v",
					result.MessageIds())
			}
		})
	}
}

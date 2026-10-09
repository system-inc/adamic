package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// TestPreferNullishCoalescingDefaultsMatchUpstream pins every default, each proved by a moving
// verdict against the installed 8.67.0 build rather than read off upstream's defaultOptions block.
//
// `ignoreConditionalTests` is the one that matters and the reason this table exists: it defaults
// TRUE, so a port assuming Go's zero value reports on every `if (a || b)` in the tree. That is the
// second non-zero default found in two batches, and both were invisible from the shape of the code.
func TestPreferNullishCoalescingDefaultsMatchUpstream(t *testing.T) {
	t.Parallel()
	decoded, err := DecodePreferNullishCoalescingOptions(nil)
	if err != nil {
		t.Fatalf("the decoder refused absent options: %v", err)
	}
	options, ok := decoded.(PreferNullishCoalescingOptions)
	if !ok {
		t.Fatalf("the decoder returned %T", decoded)
	}

	if !options.IgnoreConditionalTests {
		t.Error("ignoreConditionalTests defaults TRUE upstream, measured absent-vs-explicit-false")
	}
	if options.IgnoreTernaryTests || options.IgnoreMixedLogicalExpressions ||
		options.IgnoreBooleanCoercion || options.IgnoreIfStatements ||
		options.AllowWithoutStrictNullChecks {
		t.Errorf("every other flag defaults false: %#v", options)
	}
	if options.IgnorePrimitives != (PreferNullishCoalescingPrimitives{}) {
		t.Errorf("ignorePrimitives defaults to exempting nothing: %#v", options.IgnorePrimitives)
	}
}

// TestPreferNullishCoalescingIgnorePrimitivesBothSpellings pins the oneOf.
//
// Upstream's schema is an object of four booleans OR the bare literal `true`. Every row here was
// measured by driving the installed rule BEFORE this decoder was written, which is the order that
// matters: the corpus splits 110 object-form against 25 bare-true, so a decoder handling only the
// object form passes 110 cases and is wrong on the rest.
//
// Three rows are invisible to the corpus entirely. `{}` and the all-true object bracket the object
// form from both ends, and `false` is REFUSED BY THE SCHEMA rather than meaning "ignore nothing" —
// the row a decoder would most naturally get wrong by accepting it and doing something reasonable.
func TestPreferNullishCoalescingIgnorePrimitivesBothSpellings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		want    PreferNullishCoalescingPrimitives
		wantErr bool
	}{
		{
			name: "absent exempts nothing",
			raw:  `{}`,
		},
		{
			name: "an EMPTY object exempts nothing, which is not the same as absent at the wire",
			raw:  `{"ignorePrimitives":{}}`,
		},
		{
			name: "the bare literal true exempts everything",
			raw:  `{"ignorePrimitives":true}`,
			want: PreferNullishCoalescingPrimitives{IgnoreAll: true},
		},
		{
			// The row a decoder gets wrong by being helpful. Upstream's `enum: [true]` makes this
			// illegal, and accepting it means a config that loads here and fails in ESLint.
			name:    "the literal false is REFUSED, because enum:[true] admits only one boolean",
			raw:     `{"ignorePrimitives":false}`,
			wantErr: true,
		},
		{
			name: "one primitive, leaving the other three reporting",
			raw:  `{"ignorePrimitives":{"string":true}}`,
			want: PreferNullishCoalescingPrimitives{String: true},
		},
		{
			name: "an explicit false for one primitive exempts nothing",
			raw:  `{"ignorePrimitives":{"string":false}}`,
		},
		{
			name: "all four spelled out, which behaves like the bare true but is a different shape",
			raw:  `{"ignorePrimitives":{"bigint":true,"boolean":true,"number":true,"string":true}}`,
			want: PreferNullishCoalescingPrimitives{Bigint: true, Boolean: true, Number: true, String: true},
		},
		{
			name:    "an unknown primitive is refused, matching additionalProperties:false",
			raw:     `{"ignorePrimitives":{"strings":true}}`,
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodePreferNullishCoalescingOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected the decoder to refuse %s, got %#v", testCase.raw, decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			options := decoded.(PreferNullishCoalescingOptions)
			if options.IgnorePrimitives != testCase.want {
				t.Errorf("expected %#v, got %#v", testCase.want, options.IgnorePrimitives)
			}
		})
	}
}

// TestPreferNullishCoalescingDecoderAcceptsTheShapesTheConfigLayerDelivers crosses the boundary.
//
// `meta.schema` declares ONE element here, so the cohere spelling and upstream's coincide. Stated
// rather than assumed, because a fixture hands the decoder bytes the test built and only bytes the
// config layer sliced reveal a mismatch.
func TestPreferNullishCoalescingDecoderAcceptsTheShapesTheConfigLayerDelivers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		check   func(*testing.T, PreferNullishCoalescingOptions)
		wantErr bool
	}{
		{
			name: "a bare \"error\" delivers no options at all",
			check: func(t *testing.T, o PreferNullishCoalescingOptions) {
				if !o.IgnoreConditionalTests {
					t.Error("the non-zero default must survive an absent options value")
				}
			},
		},
		{
			name: "every flag at once",
			raw: `{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing":true,` +
				`"ignoreBooleanCoercion":true,"ignoreConditionalTests":false,` +
				`"ignoreIfStatements":true,"ignoreMixedLogicalExpressions":true,` +
				`"ignoreTernaryTests":true,"ignorePrimitives":true}`,
			check: func(t *testing.T, o PreferNullishCoalescingOptions) {
				if !o.AllowWithoutStrictNullChecks || !o.IgnoreBooleanCoercion ||
					!o.IgnoreIfStatements || !o.IgnoreMixedLogicalExpressions ||
					!o.IgnoreTernaryTests || !o.IgnorePrimitives.IgnoreAll {
					t.Errorf("a key was dropped: %#v", o)
				}
				if o.IgnoreConditionalTests {
					t.Error("an explicit false must override the true default, which is the whole " +
						"reason these are pointers on the wire")
				}
			},
		},
		{
			name:    "a shape that is neither must error rather than decode to a default",
			raw:     `"ignoreTernaryTests"`,
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodePreferNullishCoalescingOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected a refusal, got %#v", decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			if testCase.check != nil {
				testCase.check(t, decoded.(PreferNullishCoalescingOptions))
			}
		})
	}
}

// preferNullishCoalescingFile is where the fixtures pretend to live.
const preferNullishCoalescingFile = "/repository/source/PreferNullish.ts"

type preferNullishCoalescingCase struct {
	name        string
	source      string
	optionsJson string
	findings    int
	reason      string
}

func runPreferNullishCoalescing(t *testing.T, testCase preferNullishCoalescingCase) rule_testing.Result {
	t.Helper()
	var options any
	if testCase.optionsJson != "" {
		decoded, err := DecodePreferNullishCoalescingOptions(json.RawMessage(testCase.optionsJson))
		if err != nil {
			t.Fatalf("the decoder refused %s: %v", testCase.optionsJson, err)
		}
		options = decoded
	} else {
		// A rule configured as a bare "error" is handed nil, and this rule's ignoreConditionalTests
		// default is TRUE, so routing the empty case through the decoder is what proves the rule
		// gets that default rather than Go's zero value.
		decoded, err := DecodePreferNullishCoalescingOptions(nil)
		if err != nil {
			t.Fatalf("the decoder refused absent options: %v", err)
		}
		options = decoded
	}
	return rule_testing.RunTypedWithOptions(t, PreferNullishCoalescing,
		preferNullishCoalescingFile, testCase.source, options)
}

// TestPreferNullishCoalescingStaysSilent runs upstream's whole `valid` list.
//
// All 275, across every reporting path: a valid case is a false positive this rule must not
// produce, whichever path would have produced it.
func TestPreferNullishCoalescingStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range preferNullishCoalescingCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runPreferNullishCoalescing(t, testCase))
		})
	}
}

// TestPreferNullishCoalescingFires runs every invalid case whose findings are all preferNullishOverOr.
func TestPreferNullishCoalescingFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range preferNullishCoalescingOrCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferNullishCoalescing(t, testCase)
			wantIds := make([]string, 0, testCase.findings)
			for range testCase.findings {
				wantIds = append(wantIds, messagePreferNullishOverOr.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// preferNullishCoalescingCleanCases is upstream's whole `valid` list, verbatim.
var preferNullishCoalescingCleanCases = []preferNullishCoalescingCase{
	{name: "upstream valid[0]", source: `
declare let x: string;
x || 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[1]", source: `
declare let x: number;
x || 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[2]", source: `
declare let x: boolean;
x || 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[3]", source: `
declare let x: object;
x || 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[4]", source: `
declare let x: string;
x ||= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[5]", source: `
declare let x: number;
x ||= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[6]", source: `
declare let x: boolean;
x ||= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[7]", source: `
declare let x: object;
x ||= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[8]", source: `
declare let x: string | null | undefined;
x ?? 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[9]", source: `
declare let x: string | null | undefined;
x ??= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[10]", source: `
declare let x: number | null | undefined;
x ?? 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[11]", source: `
declare let x: number | null | undefined;
x ??= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[12]", source: `
declare let x: boolean | null | undefined;
x ?? 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[13]", source: `
declare let x: boolean | null | undefined;
x ??= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[14]", source: `
declare let x: object | null | undefined;
x ?? 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[15]", source: `
declare let x: object | null | undefined;
x ??= 'foo';
    `, optionsJson: ""},
	{name: "upstream valid[16]", source: `x !== undefined && x !== null ? x : y;`, optionsJson: `{"ignoreTernaryTests": true}`},
	{name: "upstream valid[17]", source: `x !== undefined && x !== null ? 'foo' : 'bar';`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[18]", source: `x !== null && x !== undefined && x !== 5 ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[19]", source: `x === null || x === undefined || x === 5 ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[20]", source: `x === undefined && x !== null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[21]", source: `x === undefined && x === null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[22]", source: `x !== undefined && x === null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[23]", source: `x === undefined || x !== null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[24]", source: `x === undefined || x === null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[25]", source: `x !== undefined || x === null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[26]", source: `x !== undefined || x === null ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[27]", source: `x === null || x === null ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[28]", source: `x === undefined || x === undefined ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[29]", source: `x == null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[30]", source: `undefined == null ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[31]", source: `undefined != z ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[32]", source: `x == undefined ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[33]", source: `x != null ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[34]", source: `x != undefined ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[35]", source: `null == x ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[36]", source: `undefined == x ? x : y;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[37]", source: `null != x ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[38]", source: `undefined != x ? y : x;`, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[39]", source: `
declare let x: number | undefined;
x !== 15 && x !== undefined ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[40]", source: `
declare let x: number | undefined;
x !== undefined && x !== 15 ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[41]", source: `
declare let x: number | undefined;
15 !== x && undefined !== x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[42]", source: `
declare let x: number | undefined;
undefined !== x && 15 !== x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[43]", source: `
declare let x: number | undefined;
15 !== x && x !== undefined ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[44]", source: `
declare let x: number | undefined;
undefined !== x && x !== 15 ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[45]", source: `
declare let x: string | undefined;
x !== 'foo' && x !== undefined ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[46]", source: `
function test(value: number | undefined): number {
  return value !== foo() && value !== undefined ? value : 1;
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[47]", source: `
const test = (value: boolean | undefined): boolean =>
  value !== undefined && value !== false ? value : false;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[48]", source: `
declare let x: string;
x === null ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[49]", source: `
declare let x: string | undefined | null;
x === null ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[50]", source: `
declare let x: string | undefined | null;
x === undefined ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[51]", source: `
declare let x: string | undefined | null;
x !== null ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[52]", source: `
declare let x: string | undefined | null;
x !== undefined ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[53]", source: `
declare let x: any;
x === null ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[54]", source: `
declare let x: unknown;
x === null ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[55]", source: `
declare let x: string;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[56]", source: `
declare let x: string;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[57]", source: `
declare let x: string | object;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[58]", source: `
declare let x: string | object;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[59]", source: `
declare let x: number;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[60]", source: `
declare let x: number;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[61]", source: `
declare let x: bigint;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[62]", source: `
declare let x: bigint;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[63]", source: `
declare let x: boolean;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[64]", source: `
declare let x: boolean;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[65]", source: `
declare let x: object;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[66]", source: `
declare let x: object;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[67]", source: `
declare let x: string[];
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[68]", source: `
declare let x: string[];
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[69]", source: `
declare let x: Function;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[70]", source: `
declare let x: Function;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[71]", source: `
declare let x: () => string;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[72]", source: `
declare let x: () => string;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[73]", source: `
declare let x: () => string | null | undefined;
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[74]", source: `
declare let x: () => string | null | undefined;
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[75]", source: `
declare let x: () => string | null;
x() ? x() : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[76]", source: `
declare let x: () => string | null;
!x() ? y : x();
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[77]", source: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
x(a) ? x(a) : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[78]", source: `
const a = 'foo';
declare let x: (a: string | null) => string | null;
!x(a) ? y : x(a);
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[79]", source: `
declare let x: { n: string };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[80]", source: `
declare let x: { n: string };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[81]", source: `
declare let x: { n: string | object };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[82]", source: `
declare let x: { n: string | object };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[83]", source: `
declare let x: { n: number };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[84]", source: `
declare let x: { n: number };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[85]", source: `
declare let x: { n: bigint };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[86]", source: `
declare let x: { n: bigint };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[87]", source: `
declare let x: { n: boolean };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[88]", source: `
declare let x: { n: boolean };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[89]", source: `
declare let x: { n: object };
x ? x : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[90]", source: `
declare let x: { n: object };
!x ? y : x;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[91]", source: `
declare let x: { n: string[] };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[92]", source: `
declare let x: { n: string[] };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[93]", source: `
declare let x: { n: Function };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[94]", source: `
declare let x: { n: Function };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[95]", source: `
declare let x: { n: () => string };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[96]", source: `
declare let x: { n: () => string };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[97]", source: `
declare let x: { n: () => string | null | undefined };
x.n ? x.n : y;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[98]", source: `
declare let x: { n: () => string | null | undefined };
!x.n ? y : x.n;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[99]", source: `
declare let foo: string;
declare function makeFoo(): string;

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[100]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo) {
    foo = makeFoo();
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[101]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo != null) {
    foo = makeFoo();
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[102]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    return foo;
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[103]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else {
    return 'bar';
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[104]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  } else if (foo.a) {
    return 'bar';
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[105]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
function shadowed() {
  if (foo == null) {
    const foo = makeFoo();
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[106]", source: `
declare let foo: { foo: string } | null;
declare function makeFoo(): { foo: { foo: string } };
function weirdDestructuringAssignment() {
  if (foo == null) {
    ({ foo } = makeFoo());
  }
}
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[107]", source: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null !== null ? nullOrObject : 42;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[108]", source: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject !== undefined && null != null ? nullOrObject : 42;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[109]", source: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject !== undefined && null != undefined ? nullOrObject : 42;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[110]", source: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null === null ? 42 : nullOrObject;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[111]", source: `
declare const nullOrObject: null | { a: string };

const test = nullOrObject === undefined || null == null ? 42 : nullOrObject;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[112]", source: `
declare const nullOrObject: null | { a: string };

const test =
  nullOrObject === undefined || null == undefined ? 42 : nullOrObject;
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[113]", source: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.a != null ? x[a] : 'foo';
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[114]", source: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x.a : 'foo';
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[115]", source: `
declare let x: { a: string } | null;
declare let y: { a: string } | null;

x?.a ? y?.a : 'foo';
      `, optionsJson: `{"ignoreTernaryTests": false}`},
	{name: "upstream valid[116]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `, optionsJson: `{"ignoreIfStatements": true}`},
	{name: "upstream valid[117]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) foo = makeFoo();
}
      `, optionsJson: `{"ignoreIfStatements": true}`},
	{name: "upstream valid[118]", source: `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[119]", source: `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[120]", source: `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[121]", source: `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[122]", source: `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[123]", source: `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[124]", source: `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[125]", source: `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
    `, optionsJson: ""},
	{name: "upstream valid[126]", source: `
declare let x: string | null | undefined;
if (x || 'foo') {
}
    `, optionsJson: ""},
	{name: "upstream valid[127]", source: `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
    `, optionsJson: ""},
	{name: "upstream valid[128]", source: `
declare let x: number | null | undefined;
if (x || 'foo') {
}
    `, optionsJson: ""},
	{name: "upstream valid[129]", source: `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
    `, optionsJson: ""},
	{name: "upstream valid[130]", source: `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
    `, optionsJson: ""},
	{name: "upstream valid[131]", source: `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
    `, optionsJson: ""},
	{name: "upstream valid[132]", source: `
declare let x: object | null | undefined;
if (x || 'foo') {
}
    `, optionsJson: ""},
	{name: "upstream valid[133]", source: `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
    `, optionsJson: ""},
	{name: "upstream valid[134]", source: `
declare let x: string | null | undefined;
do {} while (x || 'foo');
    `, optionsJson: ""},
	{name: "upstream valid[135]", source: `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
    `, optionsJson: ""},
	{name: "upstream valid[136]", source: `
declare let x: number | null | undefined;
do {} while (x || 'foo');
    `, optionsJson: ""},
	{name: "upstream valid[137]", source: `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
    `, optionsJson: ""},
	{name: "upstream valid[138]", source: `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
    `, optionsJson: ""},
	{name: "upstream valid[139]", source: `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
    `, optionsJson: ""},
	{name: "upstream valid[140]", source: `
declare let x: object | null | undefined;
do {} while (x || 'foo');
    `, optionsJson: ""},
	{name: "upstream valid[141]", source: `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
    `, optionsJson: ""},
	{name: "upstream valid[142]", source: `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
    `, optionsJson: ""},
	{name: "upstream valid[143]", source: `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
    `, optionsJson: ""},
	{name: "upstream valid[144]", source: `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
    `, optionsJson: ""},
	{name: "upstream valid[145]", source: `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
    `, optionsJson: ""},
	{name: "upstream valid[146]", source: `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
    `, optionsJson: ""},
	{name: "upstream valid[147]", source: `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
    `, optionsJson: ""},
	{name: "upstream valid[148]", source: `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
    `, optionsJson: ""},
	{name: "upstream valid[149]", source: `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
    `, optionsJson: ""},
	{name: "upstream valid[150]", source: `
declare let x: string | null | undefined;
while (x || 'foo') {}
    `, optionsJson: ""},
	{name: "upstream valid[151]", source: `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
    `, optionsJson: ""},
	{name: "upstream valid[152]", source: `
declare let x: number | null | undefined;
while (x || 'foo') {}
    `, optionsJson: ""},
	{name: "upstream valid[153]", source: `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
    `, optionsJson: ""},
	{name: "upstream valid[154]", source: `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
    `, optionsJson: ""},
	{name: "upstream valid[155]", source: `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
    `, optionsJson: ""},
	{name: "upstream valid[156]", source: `
declare let x: object | null | undefined;
while (x || 'foo') {}
    `, optionsJson: ""},
	{name: "upstream valid[157]", source: `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
    `, optionsJson: ""},
	{name: "upstream valid[158]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[159]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[160]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[161]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[162]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[163]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[164]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[165]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[166]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[167]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[168]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[169]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": true}`},
	{name: "upstream valid[170]", source: `
declare let x: string | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[171]", source: `
declare let x: number | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[172]", source: `
declare let x: boolean | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[173]", source: `
declare let x: bigint | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[174]", source: `
declare let x: string | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[175]", source: `
declare let x: number | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[176]", source: `
declare let x: boolean | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[177]", source: `
declare let x: bigint | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[178]", source: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[179]", source: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[180]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[181]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[182]", source: `
declare let x: (string & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[183]", source: `
declare let x: (number & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[184]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[185]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[186]", source: `
declare let x: string | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[187]", source: `
declare let x: number | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[188]", source: `
declare let x: boolean | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[189]", source: `
declare let x: bigint | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[190]", source: `
declare let x: string | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[191]", source: `
declare let x: number | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[192]", source: `
declare let x: boolean | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[193]", source: `
declare let x: bigint | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[194]", source: `
declare let x: string | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[195]", source: `
declare let x: number | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[196]", source: `
declare let x: boolean | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[197]", source: `
declare let x: bigint | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[198]", source: `
declare let x: string | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[199]", source: `
declare let x: number | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[200]", source: `
declare let x: boolean | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[201]", source: `
declare let x: bigint | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[202]", source: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[203]", source: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[204]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[205]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[206]", source: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[207]", source: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[208]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true}}`},
	{name: "upstream valid[209]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true}}`},
	{name: "upstream valid[210]", source: `
declare let x: (string & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[211]", source: `
declare let x: (number & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[212]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[213]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[214]", source: `
declare let x: (string & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[215]", source: `
declare let x: (number & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[216]", source: `
declare let x: (boolean & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[217]", source: `
declare let x: (bigint & { __brand?: any }) | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[218]", source: `
declare let x: never;
declare let y: number;
x || y;
    `, optionsJson: ""},
	{name: "upstream valid[219]", source: `
declare let x: never;
declare let y: number;
x ? x : y;
    `, optionsJson: ""},
	{name: "upstream valid[220]", source: `
declare let x: never;
declare let y: number;
!x ? y : x;
    `, optionsJson: ""},
	{name: "upstream valid[221]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `, optionsJson: ""},
	{name: "upstream valid[222]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `, optionsJson: ""},
	{name: "upstream valid[223]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
    `, optionsJson: ""},
	{name: "upstream valid[224]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
    `, optionsJson: ""},
	{name: "upstream valid[225]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`},
	{name: "upstream valid[226]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`},
	{name: "upstream valid[227]", source: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": true}}`},
	{name: "upstream valid[228]", source: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": false}}`},
	{name: "upstream valid[229]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[230]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[231]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[232]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[233]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`},
	{name: "upstream valid[234]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`},
	{name: "upstream valid[235]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`},
	{name: "upstream valid[236]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`},
	{name: "upstream valid[237]", source: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": true}}`},
	{name: "upstream valid[238]", source: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": true}}`},
	{name: "upstream valid[239]", source: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": false}}`},
	{name: "upstream valid[240]", source: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true, "string": false}}`},
	{name: "upstream valid[241]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[242]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[243]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[244]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[245]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[246]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[247]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[248]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, optionsJson: `{"ignorePrimitives": {"string": true}}`},
	{name: "upstream valid[249]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a || b);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[250]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || b || c);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[251]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || (b && c));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[252]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a || b) ?? c);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[253]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ?? (b || c));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[254]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? b || c : 'fail');
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[255]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? 'success' : b || c);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[256]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(((a = b), b || c));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[257]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a ? a : b) || c);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[258]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(c || (!a ? b : a));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`},
	{name: "upstream valid[259]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || b || c) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[260]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || (b && c)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[261]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a || b) ?? c) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[262]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ?? (b || c)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[263]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? b || c : 'fail') {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[264]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? 'success' : b || c) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[265]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (((a = b), b || c)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[266]", source: `
let a: string | undefined;
let b: string | undefined;

if (!(a || b)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[267]", source: `
let a: string | undefined;
let b: string | undefined;

if (!!(a || b)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[268]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

if (a ? a : b) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[269]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

if (!a ? b : a) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[270]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a ? a : b) || c) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[271]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (c || (!a ? b : a)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`},
	{name: "upstream valid[272]", source: `
declare const a: any;
declare const b: any;
a ? a : b;
      `, optionsJson: `{"ignorePrimitives": true}`},
	{name: "upstream valid[273]", source: `
declare const a: any;
declare const b: any;
a ? a : b;
      `, optionsJson: `{"ignorePrimitives": {"number": true}}`},
	{name: "upstream valid[274]", source: `
declare const a: unknown;
const b = a || 'bar';
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": false, "number": false, "string": false}}`},
}

// preferNullishCoalescingOrCases are the invalid cases whose findings are ALL preferNullishOverOr.
//
// The ternary and if-statement paths' 221 cases live in prefer_nullish_coalescing_ternary_test.go,
// with their spans and suggestions as the installed build replayed them. The one left over is
// `noStrictNullCheck`, which needs a non-strict program.
var preferNullishCoalescingOrCases = []preferNullishCoalescingCase{
	{name: "upstream invalid[0]", source: `
declare let x: string | null | undefined;
x || 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[1]", source: `
declare let x: string | null | undefined;
x ||= 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[2]", source: `
declare let x: number | null | undefined;
x || 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[3]", source: `
declare let x: number | null | undefined;
x ||= 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[4]", source: `
declare let x: boolean | null | undefined;
x || 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[5]", source: `
declare let x: boolean | null | undefined;
x ||= 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[6]", source: `
declare let x: object | null | undefined;
x || 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[7]", source: `
declare let x: object | null | undefined;
x ||= 'foo';
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[156]", source: `
declare let x: string | null | undefined;
x || 'foo' ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[157]", source: `
declare let x: string | null | undefined;
(x ||= 'foo') ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[158]", source: `
declare let x: number | null | undefined;
x || 'foo' ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[159]", source: `
declare let x: number | null | undefined;
(x ||= 'foo') ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[160]", source: `
declare let x: boolean | null | undefined;
x || 'foo' ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[161]", source: `
declare let x: boolean | null | undefined;
(x ||= 'foo') ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[162]", source: `
declare let x: object | null | undefined;
x || 'foo' ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[163]", source: `
declare let x: object | null | undefined;
(x ||= 'foo') ? null : null;
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[164]", source: `
declare let x: string | null | undefined;
if (x || 'foo') {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[165]", source: `
declare let x: string | null | undefined;
if ((x ||= 'foo')) {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[166]", source: `
declare let x: number | null | undefined;
if (x || 'foo') {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[167]", source: `
declare let x: number | null | undefined;
if ((x ||= 'foo')) {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[168]", source: `
declare let x: boolean | null | undefined;
if (x || 'foo') {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[169]", source: `
declare let x: boolean | null | undefined;
if ((x ||= 'foo')) {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[170]", source: `
declare let x: object | null | undefined;
if (x || 'foo') {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[171]", source: `
declare let x: object | null | undefined;
if ((x ||= 'foo')) {
}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[172]", source: `
declare let x: string | null | undefined;
do {} while (x || 'foo');
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[173]", source: `
declare let x: string | null | undefined;
do {} while ((x ||= 'foo'));
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[174]", source: `
declare let x: number | null | undefined;
do {} while (x || 'foo');
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[175]", source: `
declare let x: number | null | undefined;
do {} while ((x ||= 'foo'));
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[176]", source: `
declare let x: boolean | null | undefined;
do {} while (x || 'foo');
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[177]", source: `
declare let x: boolean | null | undefined;
do {} while ((x ||= 'foo'));
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[178]", source: `
declare let x: object | null | undefined;
do {} while (x || 'foo');
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[179]", source: `
declare let x: object | null | undefined;
do {} while ((x ||= 'foo'));
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[180]", source: `
declare let x: string | null | undefined;
for (; x || 'foo';) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[181]", source: `
declare let x: string | null | undefined;
for (; (x ||= 'foo');) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[182]", source: `
declare let x: number | null | undefined;
for (; x || 'foo';) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[183]", source: `
declare let x: number | null | undefined;
for (; (x ||= 'foo');) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[184]", source: `
declare let x: boolean | null | undefined;
for (; x || 'foo';) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[185]", source: `
declare let x: boolean | null | undefined;
for (; (x ||= 'foo');) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[186]", source: `
declare let x: object | null | undefined;
for (; x || 'foo';) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[187]", source: `
declare let x: object | null | undefined;
for (; (x ||= 'foo');) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[188]", source: `
declare let x: string | null | undefined;
while (x || 'foo') {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[189]", source: `
declare let x: string | null | undefined;
while ((x ||= 'foo')) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[190]", source: `
declare let x: number | null | undefined;
while (x || 'foo') {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[191]", source: `
declare let x: number | null | undefined;
while ((x ||= 'foo')) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[192]", source: `
declare let x: boolean | null | undefined;
while (x || 'foo') {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[193]", source: `
declare let x: boolean | null | undefined;
while ((x ||= 'foo')) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[194]", source: `
declare let x: object | null | undefined;
while (x || 'foo') {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[195]", source: `
declare let x: object | null | undefined;
while ((x ||= 'foo')) {}
      `, optionsJson: `{"ignoreConditionalTests": false}`, findings: 1},
	{name: "upstream invalid[196]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 1},
	{name: "upstream invalid[197]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 1},
	{name: "upstream invalid[198]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 1},
	{name: "upstream invalid[199]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a || (b && c);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 1},
	{name: "upstream invalid[200]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[201]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[202]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[203]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
a || b || (c && d);
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[204]", source: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[205]", source: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
declare let d: number | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[206]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
declare let d: boolean | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[207]", source: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
declare let d: object | null | undefined;
(a && b) || c || d;
      `, optionsJson: `{"ignoreMixedLogicalExpressions": false}`, findings: 2},
	{name: "upstream invalid[208]", source: `
declare let x: string | null | undefined;
if (() => x || 'foo') {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[209]", source: `
declare let x: string | null | undefined;
if (() => (x ||= 'foo')) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[210]", source: `
declare let x: number | null | undefined;
if (() => x || 'foo') {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[211]", source: `
declare let x: number | null | undefined;
if (() => (x ||= 'foo')) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[212]", source: `
declare let x: boolean | null | undefined;
if (() => x || 'foo') {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[213]", source: `
declare let x: boolean | null | undefined;
if (() => (x ||= 'foo')) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[214]", source: `
declare let x: object | null | undefined;
if (() => x || 'foo') {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[215]", source: `
declare let x: object | null | undefined;
if (() => (x ||= 'foo')) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[216]", source: `
declare let x: string | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[217]", source: `
declare let x: string | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[218]", source: `
declare let x: number | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[219]", source: `
declare let x: number | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[220]", source: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[221]", source: `
declare let x: boolean | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[222]", source: `
declare let x: object | null | undefined;
if (
  function weird() {
    return x || 'foo';
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[223]", source: `
declare let x: object | null | undefined;
if (
  function weird() {
    return (x ||= 'foo');
  }
) {
}
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[224]", source: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;
a || b || c;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[225]", source: `
declare let a: number | null | undefined;
declare let b: number;
declare let c: number;
a || b || c;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[226]", source: `
declare let a: boolean | null | undefined;
declare let b: boolean;
declare let c: boolean;
a || b || c;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[227]", source: `
declare let a: object | null | undefined;
declare let b: object;
declare let c: object;
a || b || c;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[228]", source: `
declare let x: string | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true}}`, findings: 1},
	{name: "upstream invalid[229]", source: `
declare let x: number | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[230]", source: `
declare let x: boolean | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[231]", source: `
declare let x: bigint | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"boolean": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[236]", source: `
declare let x: '' | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[237]", source: "\ndeclare let x: `` | undefined;\nx || y;\n      ", optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[238]", source: `
declare let x: 0 | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`, findings: 1},
	{name: "upstream invalid[239]", source: `
declare let x: 0n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[240]", source: `
declare let x: false | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[246]", source: `
declare let x: 'a' | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[247]", source: "\ndeclare let x: `hello${'string'}` | undefined;\nx || y;\n      ", optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[248]", source: `
declare let x: 1 | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`, findings: 1},
	{name: "upstream invalid[249]", source: `
declare let x: 1n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[250]", source: `
declare let x: true | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[261]", source: `
declare let x: 'a' | 'b' | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[262]", source: "\ndeclare let x: 'a' | `b` | undefined;\nx || y;\n      ", optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`, findings: 1},
	{name: "upstream invalid[263]", source: `
declare let x: 0 | 1 | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`, findings: 1},
	{name: "upstream invalid[264]", source: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`, findings: 1},
	{name: "upstream invalid[265]", source: `
declare let x: 0n | 1n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[266]", source: `
declare let x: 1n | 2n | 3n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[267]", source: `
declare let x: true | false | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[282]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": false, "boolean": true, "number": false, "string": true}}`, findings: 1},
	{name: "upstream invalid[283]", source: `
declare let x: true | false | null | undefined;
x || y;
      `, optionsJson: `{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`, findings: 1},
	{name: "upstream invalid[288]", source: `
declare let x: null;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[289]", source: `
const x = undefined;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[290]", source: `
null || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[291]", source: `
undefined || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[292]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[293]", source: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[294]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[295]", source: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[296]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a || b);
      `, optionsJson: `{"ignoreBooleanCoercion": false}`, findings: 1},
	{name: "upstream invalid[297]", source: `
function outer() {
  const Boolean = (x: unknown) => x;

  return (a: string | null, b: string) => Boolean(a || b);
}
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[298]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a || b);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[299]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a || b);
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[300]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a || b;
});
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[301]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a || b));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[302]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a || b));
      `, optionsJson: `{"ignoreBooleanCoercion": true}`, findings: 1},
	{name: "upstream invalid[305]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a || b)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`, findings: 1},
	{name: "upstream invalid[306]", source: `
declare const a: string | undefined;
declare const b: string;

if (+(a || b)) {
}
      `, optionsJson: `{"ignoreConditionalTests": true}`, findings: 1},
	{name: "upstream invalid[307]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox || getFallbackBox();
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[310]", source: `
declare const x: any;
declare const y: any;
x || y;
      `, optionsJson: "", findings: 1},
	{name: "upstream invalid[311]", source: `
declare const x: unknown;
declare const y: any;
x || y;
      `, optionsJson: "", findings: 1},
}

// preferNullishCoalescingRenderingCase is one shape with everything a message id cannot see.
type preferNullishCoalescingRenderingCase struct {
	name    string
	source  string
	options string
	// findings, in order, each with the exact span, whole message, and suggested replacement text.
	findings []preferNullishCoalescingRendering
	reason   string
}

type preferNullishCoalescingRendering struct {
	line      int
	column    int
	endLine   int
	endColumn int
	message   string
	suggested string
}

// TestPreferNullishCoalescingRenderings is the section 3c and section 8 table.
//
// This rule has ONE message id and reports on two arms that render differently, propose different
// repair text, and parenthesise differently when chained. A fixture asserting the id and the count
// cannot see any of that, and the imported corpus asserts nothing more: all 398 cases pass with a
// rule that renders the wrong arm's message, writes the wrong replacement, and proposes a syntax
// error on every chained expression.
//
// Every expectation below was measured by driving the installed 8.67.0 build and reading the
// message and suggestion text, not the id.
func TestPreferNullishCoalescingRenderings(t *testing.T) {
	t.Parallel()
	const orMessage = "Prefer using nullish coalescing operator (`??`) instead of a logical or " +
		"(`||`), as it is a safer operator."
	const assignMessage = "Prefer using nullish coalescing operator (`??=`) instead of a logical " +
		"assignment (`||=`), as it is a safer operator."

	cases := []preferNullishCoalescingRenderingCase{
		{
			name:   "the || arm renders `or` and suggests ??",
			source: "declare let a: string | null;\ndeclare const b: string;\nconst x = a || b;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 3, column: 13, endLine: 3, endColumn: 15, message: orMessage, suggested: "??"},
			},
			reason: "the span is the OPERATOR token, not the expression",
		},
		{
			name:   "the ||= arm renders `assignment` and suggests ??=",
			source: "declare let a: string | null;\ndeclare const b: string;\na ||= b;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 3, column: 3, endLine: 3, endColumn: 6, message: assignMessage, suggested: "??="},
			},
			reason: "one message id, a different rendering AND a different repair; nothing in the " +
				"corpus separates this arm from the one above",
		},
		{
			name: "a chained || parenthesises the inner finding, because `a ?? b || c` is a syntax error",
			source: "declare let a: string | null;\ndeclare let b: string | null;\n" +
				"declare const c: string;\nconst x = a || b || c;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 4, column: 13, endLine: 4, endColumn: 15, message: orMessage, suggested: "(a ?? b)"},
				{line: 4, column: 18, endLine: 4, endColumn: 20, message: orMessage, suggested: "??"},
			},
			reason: "the inner || has a || parent so its whole expression is replaced; the outer one " +
				"gets the plain swap. A port writing ?? everywhere passes all 398 corpus cases.",
		},
		{
			name: "the parenthesised chain puts the replacement on the other side",
			source: "declare let a: string | null;\ndeclare let b: string | null;\n" +
				"declare const c: string;\nconst x = a || (b || c);\n",
			findings: []preferNullishCoalescingRendering{
				{line: 4, column: 13, endLine: 4, endColumn: 15, message: orMessage, suggested: "??"},
				{line: 4, column: 19, endLine: 4, endColumn: 21, message: orMessage, suggested: "(b ?? c)"},
			},
			reason: "which finding gets parenthesised follows the PARENT, so writing the source the " +
				"other way round swaps them",
		},
		{
			name: "three deep, so two findings parenthesise and only the outermost does not",
			source: "declare let a: string | null;\ndeclare let b: string | null;\n" +
				"declare let c: string | null;\ndeclare const d: string;\nconst x = a || b || c || d;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 5, column: 13, endLine: 5, endColumn: 15, message: orMessage, suggested: "(a ?? b)"},
				{line: 5, column: 18, endLine: 5, endColumn: 20, message: orMessage, suggested: "(b ?? c)"},
				{line: 5, column: 23, endLine: 5, endColumn: 25, message: orMessage, suggested: "??"},
			},
			reason: "the rule is per-finding rather than per-expression",
		},
		{
			name: "a && parent is NOT a || parent, so the plain swap stands",
			source: "declare let a: string | null;\ndeclare const b: string;\ndeclare const d: string;\n" +
				"const x = (a || b) && d;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 4, column: 14, endLine: 4, endColumn: 16, message: orMessage, suggested: "??"},
			},
			reason: "the row that shows the test is on a logical OR parent specifically, not on any " +
				"logical parent",
		},
		{
			name:   "optional chaining reports exactly like a plain member access",
			source: "declare const o: { a?: string } | null;\ndeclare const c: string;\nconst x = o?.a || c;\n",
			findings: []preferNullishCoalescingRendering{
				{line: 3, column: 16, endLine: 3, endColumn: 18, message: orMessage, suggested: "??"},
			},
			reason: "predicted as a divergence and measured as a non-divergence: the rule anchors on " +
				"the operator, so ESTree's ChainExpression wrapper never gates it",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferNullishCoalescing(t, preferNullishCoalescingCase{
				source: testCase.source, optionsJson: testCase.options,
			})
			if len(result.Diagnostics) != len(testCase.findings) {
				t.Fatalf("expected %d findings, got %d: %v. %s",
					len(testCase.findings), len(result.Diagnostics), result.MessageIds(), testCase.reason)
			}
			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]

				wantPos := preferNullishCoalescingOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := preferNullishCoalescingOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d) %q, got [%d,%d) %q",
						index, wantPos, wantEnd,
						asTheTypedPreferNullishHarnessWroteIt(testCase.source)[wantPos:wantEnd],
						diagnostic.Range.Pos(), diagnostic.Range.End(),
						asTheTypedPreferNullishHarnessWroteIt(testCase.source)[diagnostic.Range.Pos():diagnostic.Range.End()])
				}

				// The WHOLE upstream sentence, not a prefix. A prefix assertion cannot separate the
				// two arms, which is the defect this table exists to catch.
				if !strings.HasPrefix(diagnostic.Message.Description, expected.message) {
					t.Errorf("finding %d: expected the message to open with\n  %q\ngot\n  %q",
						index, expected.message, diagnostic.Message.Description)
				}

				if len(diagnostic.Suggestions) != 1 {
					t.Fatalf("finding %d: expected one suggestion, got %d",
						index, len(diagnostic.Suggestions))
				}
				fixes := diagnostic.Suggestions[0].Fixes
				if len(fixes) != 1 {
					t.Fatalf("finding %d: expected one fix, got %d", index, len(fixes))
				}
				if fixes[0].Text != expected.suggested {
					t.Errorf("finding %d: expected the repair to write %q, got %q. %s",
						index, expected.suggested, fixes[0].Text, testCase.reason)
				}
			}
		})
	}
}

// preferNullishCoalescingOffsetOf converts a 1-based line and column into a byte offset in the file
// as RunTyped wrote it.
//
// `RunTyped` writes `strings.TrimSpace(contents)+"\n"`, so a fixture with leading whitespace is
// offset from the Go literal. These sources have none, and the helper is written to handle it anyway
// so a later case with a leading newline is not silently off by one.
func preferNullishCoalescingOffsetOf(t *testing.T, source string, line int, column int) int {
	t.Helper()
	text := asTheTypedPreferNullishHarnessWroteIt(source)
	offset := 0
	for current := 1; current < line; current++ {
		next := strings.IndexByte(text[offset:], '\n')
		if next < 0 {
			t.Fatalf("the fixture has no line %d", line)
		}
		offset += next + 1
	}
	return offset + column - 1
}

func asTheTypedPreferNullishHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestPreferNullishCoalescingArmCoverage counts upstream's corpus by the path that reports it.
//
// Upstream reports four message ids from three paths. A count that silently drifts is how a partial
// port comes to look complete, and this rule did look complete for a month: its decoder parsed the
// ternary and if-statement options while neither check existed (#10kqgs6). So every path's cases are
// counted here, and dropping one fails by name.
//
//	preferNullishOverOr           123 invalid cases    prefer_nullish_coalescing_test.go
//	preferNullishOverTernary      205 invalid cases    prefer_nullish_coalescing_ternary_test.go
//	preferNullishOverAssignment    14 invalid cases    prefer_nullish_coalescing_ternary_test.go
//	  and also ...OverOr            2 invalid cases    prefer_nullish_coalescing_ternary_test.go
//	noStrictNullCheck               1 invalid case     not implemented, needs a non-strict program
//
// All 275 of upstream's valid cases run against every path.
func TestPreferNullishCoalescingArmCoverage(t *testing.T) {
	t.Parallel()
	if len(preferNullishCoalescingOrCases) != 123 {
		t.Errorf("expected upstream's 123 preferNullishOverOr cases, have %d",
			len(preferNullishCoalescingOrCases))
	}
	if len(preferNullishCoalescingCleanCases) != 275 {
		t.Errorf("expected all 275 of upstream's valid cases, have %d", len(preferNullishCoalescingCleanCases))
	}

	byPath := map[string]int{}
	for _, testCase := range preferNullishCoalescingTernaryAndIfCases {
		ids := map[string]bool{}
		for _, finding := range testCase.findings {
			ids[finding.id] = true
		}
		switch {
		case ids[messagePreferNullishOverTernary.Id] && len(ids) == 1:
			byPath["ternary"]++
		case ids[messagePreferNullishOverAssignment.Id] && len(ids) == 1:
			byPath["assignment"]++
		case ids[messagePreferNullishOverAssignment.Id] && ids[messagePreferNullishOverOr.Id] && len(ids) == 2:
			byPath["assignment and or"]++
		default:
			t.Errorf("%s reports an unexpected mix of ids: %v", testCase.name, ids)
		}
	}
	for path, want := range map[string]int{"ternary": 205, "assignment": 14, "assignment and or": 2} {
		if byPath[path] != want {
			t.Errorf("expected %d %s cases from upstream's corpus, have %d", want, path, byPath[path])
		}
	}
}

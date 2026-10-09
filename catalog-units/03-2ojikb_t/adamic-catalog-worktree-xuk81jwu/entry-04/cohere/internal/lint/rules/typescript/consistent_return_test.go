package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// consistentReturnFile is where the fixtures pretend to live.
//
// A `.ts` rather than a `.tsx`: nothing in this rule is JSX and upstream gates on no file extension,
// so a `.tsx` fixture would only hide a gate if one were ever added.
const consistentReturnFile = "/repository/source/ConsistentReturn.ts"

// consistentReturnExpectation is one finding upstream asserts, with the span it points at.
//
// The span is here because `ExpectFindings` asserts message ids and count and nothing else, and this
// rule computes a different report position per node kind. Upstream states line, column, endLine and
// endColumn on every one of its eleven reporting cases, so there is no reason to assert less.
type consistentReturnExpectation struct {
	id        string
	line      int
	column    int
	endLine   int
	endColumn int
}

type consistentReturnCase struct {
	name        string
	source      string
	optionsJson string
	findings    []consistentReturnExpectation
}

// decodeConsistentReturnOptionsForTest routes a case's options through the SHIPPED decoder.
//
// Not a hand-built settings struct. A fixture that constructs the options directly never exercises
// the decoder, so a decoder wrong about the shape the config layer delivers passes every fixture.
func decodeConsistentReturnOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeConsistentReturnOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", optionsJson, err)
	}
	return decoded
}

// asTheTypedConsistentReturnHarnessWroteIt transforms a fixture the way RunTyped transforms it.
//
// `internal/lint/testing/program.go` writes each fixture as `strings.TrimSpace(contents)+"\n"`, so
// the file on disk is offset from the string in the Go literal above it. Every case here is copied
// from an upstream tester and carries a leading newline, so a span computed against the literal
// would be off by one at both ends.
func asTheTypedConsistentReturnHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// consistentReturnOffsetOf converts a 1-based line and column into a byte offset in the file as the
// harness wrote it, so upstream's asserted positions can be compared against reported ranges.
func consistentReturnOffsetOf(t *testing.T, source string, line int, column int) int {
	t.Helper()
	text := asTheTypedConsistentReturnHarnessWroteIt(source)
	// Upstream's asserted line numbers count the leading newline every case in its tester carries,
	// and `RunTyped` trims it. So upstream's line N is the harness's line N-1, and a helper that did
	// not shift would report an off-by-one-LINE that reads exactly like a defect in the rule's span
	// computation. Measured: every reported span here already covered the right TEXT while the
	// arithmetic disagreed, which is the tell.
	line -= consistentReturnLeadingBlankLines(source)
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

// consistentReturnLeadingBlankLines counts what TrimSpace removes from the front of a fixture, in
// lines, so an upstream position can be rebased onto the file the harness actually wrote.
func consistentReturnLeadingBlankLines(source string) int {
	trimmed := strings.TrimLeft(source, " \t\r\n")
	return strings.Count(source[:len(source)-len(trimmed)], "\n")
}

func runConsistentReturn(t *testing.T, testCase consistentReturnCase) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedWithOptions(t, ConsistentReturn, consistentReturnFile,
		testCase.source, decodeConsistentReturnOptionsForTest(t, testCase.optionsJson))
}

// TestConsistentReturnStaysSilent runs upstream's whole `valid` list.
//
// Thirteen of these nineteen report under the bare core rule and are silent here, which is the whole
// difference between the two rules. See the rule's doc comment for the measurement.
func TestConsistentReturnStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range consistentReturnCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runConsistentReturn(t, testCase))
		})
	}
}

// TestConsistentReturnFires runs upstream's whole `invalid` list and asserts every span it states.
func TestConsistentReturnFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range consistentReturnReportingCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConsistentReturn(t, testCase)

			// The message ids and the count through the harness, so the mechanical fixture-pair
			// guard in `internal/lint/registry` can see that this rule is proven to fire. The spans
			// below are the part that guard does not check and that this rule most needs: it
			// computes a different report position per node kind, and upstream states all four
			// coordinates on every one of these cases.
			wantIds := make([]string, 0, len(testCase.findings))
			for _, expected := range testCase.findings {
				wantIds = append(wantIds, expected.id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := consistentReturnOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := consistentReturnOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d) (%d:%d-%d:%d), got [%d,%d) %q",
						index, wantPos, wantEnd,
						expected.line, expected.column, expected.endLine, expected.endColumn,
						diagnostic.Range.Pos(), diagnostic.Range.End(),
						asTheTypedConsistentReturnHarnessWroteIt(testCase.source)[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
			}
		})
	}
}

// consistentReturnCleanCases are upstream's `valid` list, verbatim.
var consistentReturnCleanCases = []consistentReturnCase{
	{
		name: "upstream valid[0]",
		source: `
function foo() {
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[1]",
		source: `
const foo = (flag: boolean) => {
  if (flag) return true;
  return false;
};
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[2]",
		source: `
class A {
  foo() {
    if (a) return true;
    return false;
  }
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[3]",
		source: `
const foo = (flag: boolean) => {
  if (flag) return;
  else return undefined;
};
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
	},
	{
		name: "upstream valid[4]",
		source: `
declare function bar(): void;
function foo(flag: boolean): void {
  if (flag) {
    return bar();
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[5]",
		source: `
declare function bar(): void;
const foo = (flag: boolean): void => {
  if (flag) {
    return;
  }
  return bar();
};
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[6]",
		source: `
function foo(flag?: boolean): number | void {
  if (flag) {
    return 42;
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[7]",
		source: `
function foo(): boolean;
function foo(flag: boolean): void;
function foo(flag?: boolean): boolean | void {
  if (flag) {
    return;
  }
  return true;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[8]",
		source: `
class Foo {
  baz(): void {}
  bar(flag: boolean): void {
    if (flag) return baz();
    return;
  }
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[9]",
		source: `
declare function bar(): void;
function foo(flag: boolean): void {
  function fn(): string {
    return '1';
  }
  if (flag) {
    return bar();
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[10]",
		source: `
class Foo {
  foo(flag: boolean): void {
    const bar = (): void => {
      if (flag) return;
      return this.foo();
    };
    if (flag) {
      return this.bar();
    }
    return;
  }
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[11]",
		source: `
declare function bar(): void;
async function foo(flag?: boolean): Promise<void> {
  if (flag) {
    return bar();
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[12]",
		source: `
declare function bar(): Promise<void>;
async function foo(flag?: boolean): Promise<ReturnType<typeof bar>> {
  if (flag) {
    return bar();
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[13]",
		source: `
async function foo(flag?: boolean): Promise<Promise<void | undefined>> {
  if (flag) {
    return undefined;
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[14]",
		source: `
type PromiseVoidNumber = Promise<void | number>;
async function foo(flag?: boolean): PromiseVoidNumber {
  if (flag) {
    return 42;
  }
  return;
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[15]",
		source: `
class Foo {
  baz(): void {}
  async bar(flag: boolean): Promise<void> {
    if (flag) return baz();
    return;
  }
}
    `,
		optionsJson: "",
	},
	{
		name: "upstream valid[16]",
		source: `
declare const undef: undefined;
function foo(flag: boolean) {
  if (flag) {
    return undef;
  }
  return 'foo';
}
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": false}`,
	},
	{
		name: "upstream valid[17]",
		source: `
function foo(flag: boolean): undefined {
  if (flag) {
    return undefined;
  }
  return;
}
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
	},
	{
		name: "upstream valid[18]",
		source: `
declare const undef: undefined;
function foo(flag: boolean): undefined {
  if (flag) {
    return undef;
  }
  return;
}
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
	},
}

// consistentReturnReportingCases are upstream's `invalid` list, verbatim, with the spans it asserts.
var consistentReturnReportingCases = []consistentReturnCase{
	{
		name: "upstream invalid[0]",
		source: `
function foo(flag: boolean): any {
  if (flag) return true;
  else return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "missingReturnValue", line: 4, column: 8, endLine: 4, endColumn: 15},
		},
	},
	{
		name: "upstream invalid[1]",
		source: `
function bar(): undefined {}
function foo(flag: boolean): undefined {
  if (flag) return bar();
  return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "missingReturnValue", line: 5, column: 3, endLine: 5, endColumn: 10},
		},
	},
	{
		name: "upstream invalid[2]",
		source: `
declare function foo(): void;
function bar(flag: boolean): undefined {
  function baz(): undefined {
    if (flag) return;
    return undefined;
  }
  if (flag) return baz();
  return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 6, column: 5, endLine: 6, endColumn: 22},
			{id: "missingReturnValue", line: 9, column: 3, endLine: 9, endColumn: 10},
		},
	},
	{
		name: "upstream invalid[3]",
		source: `
function foo(flag: boolean): Promise<void> {
  if (flag) return Promise.resolve(void 0);
  else return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "missingReturnValue", line: 4, column: 8, endLine: 4, endColumn: 15},
		},
	},
	{
		name: "upstream invalid[4]",
		source: `
async function foo(flag: boolean): Promise<string> {
  if (flag) return;
  else return 'value';
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 4, column: 8, endLine: 4, endColumn: 23},
		},
	},
	{
		name: "upstream invalid[5]",
		source: `
async function foo(flag: boolean): Promise<string | undefined> {
  if (flag) return 'value';
  else return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "missingReturnValue", line: 4, column: 8, endLine: 4, endColumn: 15},
		},
	},
	{
		name: "upstream invalid[6]",
		source: `
async function foo(flag: boolean) {
  if (flag) return;
  return 1;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 4, column: 3, endLine: 4, endColumn: 12},
		},
	},
	{
		name: "upstream invalid[7]",
		source: `
function foo(flag: boolean): Promise<string | undefined> {
  if (flag) return;
  else return 'value';
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 4, column: 8, endLine: 4, endColumn: 23},
		},
	},
	{
		name: "upstream invalid[8]",
		source: `
declare function bar(): Promise<void>;
function foo(flag?: boolean): Promise<void> {
  if (flag) {
    return bar();
  }
  return;
}
      `,
		optionsJson: "",
		findings: []consistentReturnExpectation{
			{id: "missingReturnValue", line: 7, column: 3, endLine: 7, endColumn: 10},
		},
	},
	{
		name: "upstream invalid[9]",
		source: `
function foo(flag: boolean): undefined | boolean {
  if (flag) {
    return undefined;
  }
  return true;
}
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 6, column: 3, endLine: 6, endColumn: 15},
		},
	},
	{
		name: "upstream invalid[10]",
		source: `
declare const undefOrNum: undefined | number;
function foo(flag: boolean) {
  if (flag) {
    return;
  }
  return undefOrNum;
}
      `,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
		findings: []consistentReturnExpectation{
			{id: "unexpectedReturnValue", line: 7, column: 3, endLine: 7, endColumn: 21},
		},
	},
}

// TestConsistentReturnDecoderAcceptsTheShapesTheConfigLayerDelivers crosses the config boundary.
//
// Every fixture above reaches the decoder with bytes this test file built. The config layer builds
// different bytes: `internal/lint/configuration/configuration.go` parses a rule setting as
// `["severity", <options>]` and keeps `tuple[1]`, a single JSON value, while ESLint's
// `context.options` is every element after the severity. A decoder wrong for the cohere shape is
// right on every fixture and wrong on every real configuration, and nothing else in this file can
// see that.
//
// `meta.schema` for this rule declares ONE element, so the two spellings coincide and the object
// arrives whole. That is the fact being pinned, rather than assumed: if it were variadic there would
// be a second shape to accept.
func TestConsistentReturnDecoderAcceptsTheShapesTheConfigLayerDelivers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		raw      string
		wantFlag bool
		wantErr  bool
	}{
		{
			name: "absent options, which is what a bare \"error\" delivers",
			raw:  "",
		},
		{
			name:     "the cohere spelling: the option object as the single value after the severity",
			raw:      `{"treatUndefinedAsUnspecified": true}`,
			wantFlag: true,
		},
		{
			name: "the same object with the flag explicitly off",
			raw:  `{"treatUndefinedAsUnspecified": false}`,
		},
		{
			name: "an empty object, which must not turn the flag on",
			raw:  `{}`,
		},
		{
			// A decoder answering "no options" to a shape it did not understand produces a rule
			// that registers, reports, and enforces something other than what the config says.
			// Refusing is the only answer that cannot ship silently.
			name:    "a shape that is neither, which must error rather than decode to a default",
			raw:     `"treatUndefinedAsUnspecified"`,
			wantErr: true,
		},
		{
			name:    "an array, which is upstream's variadic spelling and not one this layer delivers",
			raw:     `[{"treatUndefinedAsUnspecified": true}]`,
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeConsistentReturnOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected the decoder to refuse %s, got %#v", testCase.raw, decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			settings, ok := decoded.(ConsistentReturnSettings)
			if !ok {
				t.Fatalf("the decoder returned %T rather than ConsistentReturnSettings", decoded)
			}
			if settings.TreatUndefinedAsUnspecified != testCase.wantFlag {
				t.Errorf("treatUndefinedAsUnspecified: expected %v, got %v",
					testCase.wantFlag, settings.TreatUndefinedAsUnspecified)
			}
		})
	}
}

// TestConsistentReturnOptionActuallyReachesTheRule proves the decoded flag changes what is reported.
//
// A decoder that round-trips correctly and a rule that ignores what it decoded look identical from
// the test above. One source, two configurations, opposite verdicts.
func TestConsistentReturnOptionActuallyReachesTheRule(t *testing.T) {
	t.Parallel()
	const source = `
function foo(flag: boolean) {
  if (flag) return undefined;
  return;
}
`
	withOption := runConsistentReturn(t, consistentReturnCase{
		source:      source,
		optionsJson: `{"treatUndefinedAsUnspecified": true}`,
	})
	if len(withOption.Diagnostics) != 0 {
		t.Errorf("with treatUndefinedAsUnspecified the two returns agree, got %v",
			withOption.MessageIds())
	}

	withoutOption := runConsistentReturn(t, consistentReturnCase{source: source})
	if len(withoutOption.Diagnostics) == 0 {
		t.Error("without the option `return undefined` carries a value and must contradict `return`")
	}
}

// TestConsistentReturnVoidSuppressionShapesTheCorpusDoesNotWrite is the adversarial pass.
//
// Upstream's corpus proves this rule agrees with upstream where upstream looked. These are shapes it
// never writes, aimed one per decision the void filter makes. Each expectation was taken by driving
// the installed 8.67.0 extension rather than reasoned about, because every one of them is a place
// where a locally sound argument gets the wrong answer.
func TestConsistentReturnVoidSuppressionShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		source  string
		options string
		reports bool
		reason  string
	}{
		{
			name: "a union of void with a value type suppresses",
			source: `
function foo(flag: boolean): string | void {
  if (flag) return 'a';
  return;
}
`,
			reports: false,
			reason: "The shelf's IsTypeFlagSet reads the type's own flags and answers false for a " +
				"union, while upstream ORs the constituents' flags first. This is the shape that " +
				"separates the two, and reaching for the shelf helper by name reports it.",
		},
		{
			name: "a non-async function returning Promise<void> does NOT suppress",
			source: `
declare function bar(): Promise<void>;
function foo(flag: boolean): Promise<void> {
  if (flag) return bar();
  return;
}
`,
			reports: true,
			reason: "Upstream chooses the promise branch on the function's own `async` modifier, " +
				"not on whether the return type is thenable. Keying on thenability instead reads " +
				"as an improvement and silences this.",
		},
		{
			name: "an async function whose promise resolves to a value does NOT suppress",
			source: `
async function foo(flag: boolean): Promise<string> {
  if (flag) return 'a';
  return;
}
`,
			reports: true,
			reason:  "The awaited type is string, so the recursion must bottom out at false.",
		},
		{
			name: "a triply nested promise of void suppresses",
			source: `
async function foo(flag: boolean): Promise<Promise<Promise<void>>> {
  if (flag) return undefined as never;
  return;
}
`,
			reports: false,
			reason: "Upstream's isPromiseVoid recurses on the awaited type. A single unwrap answers " +
				"false here and the corpus only ever nests twice.",
		},
		{
			name: "an inferred void return type suppresses, with no annotation anywhere",
			source: `
declare function bar(): void;
function foo(flag: boolean) {
  if (flag) return bar();
  return;
}
`,
			reports: false,
			reason: "The filter reads the checker's answer rather than an annotation node, so an " +
				"inferred void is the same question. Nothing in the corpus omits the annotation.",
		},
		{
			name: "a method with a void return type suppresses",
			source: `
declare function bar(): void;
class A {
  foo(flag: boolean): void {
    if (flag) return bar();
    return;
  }
}
`,
			reports: false,
			reason: "Upstream's wrapper listens on FunctionExpression, which is what a method's " +
				"value is there; here a method is its own node kind and must reach the same filter.",
		},
		{
			name: "an arrow with a void return type suppresses",
			source: `
declare function bar(): void;
const foo = (flag: boolean): void => {
  if (flag) return bar();
  return;
};
`,
			reports: false,
			reason:  "The third of upstream's three listener kinds.",
		},
		{
			name: "a function returning never does NOT suppress",
			source: `
declare function bar(): never;
function foo(flag: boolean): string | never {
  if (flag) return 'a';
  return;
}
`,
			reports: true,
			reason: "`never` is absorbed by the union rather than carrying the Void flag, so a " +
				"filter matching VoidLike instead of Void would wrongly silence this.",
		},
		{
			name: "a function returning undefined does NOT suppress",
			source: `
function foo(flag: boolean): undefined {
  if (flag) return undefined;
  return;
}
`,
			reports: true,
			reason: "`undefined` and `void` are different flags. Upstream's corpus asserts this " +
				"indirectly through invalid[1]; stated directly here because a filter matching " +
				"VoidLike would cover both and pass that case anyway.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConsistentReturn(t, consistentReturnCase{
				source: testCase.source, optionsJson: testCase.options,
			})
			if testCase.reports && len(result.Diagnostics) == 0 {
				t.Errorf("expected a finding and got none. %s", testCase.reason)
			}
			if !testCase.reports && len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v. %s", result.MessageIds(), testCase.reason)
			}
		})
	}
}

// TestConsistentReturnTypedUndefinedReadsTheArgumentsType pins the second filter's two sides.
//
// Both rows come from upstream's corpus and both are about the TYPE rather than the spelling: the
// syntactic test in the shared judgment matches a bare `undefined` identifier and a `void`
// expression, and neither row here is either of those.
//
// What this deliberately does NOT assert is that the filter's `==` differs from an `&`. A first
// draft claimed it did, with a reasoned argument that a union would carry extra bits. Measured, a
// union carries the Union flag alone and no input separates the two spellings. The rule's doc
// comment carries the measurement; there is no distinction here to pin.
func TestConsistentReturnTypedUndefinedReadsTheArgumentsType(t *testing.T) {
	t.Parallel()
	const options = `{"treatUndefinedAsUnspecified": true}`

	exactly := runConsistentReturn(t, consistentReturnCase{
		source: `
declare const undef: undefined;
function foo(flag: boolean): undefined {
  if (flag) {
    return undef;
  }
  return;
}
`,
		optionsJson: options,
	})
	if len(exactly.Diagnostics) != 0 {
		t.Errorf("a value typed exactly `undefined` counts as unspecified, got %v",
			exactly.MessageIds())
	}

	union := runConsistentReturn(t, consistentReturnCase{
		source: `
declare const undefOrNum: undefined | number;
function foo(flag: boolean) {
  if (flag) {
    return;
  }
  return undefOrNum;
}
`,
		optionsJson: options,
	})
	if len(union.Diagnostics) == 0 {
		t.Error("a union containing undefined is not exactly undefined and must still report")
	}
}

// TestConsistentReturnDoesNotPanicOnShapesWithNoUsefulType is the crash guard.
//
// `Checker_getTypeArguments` PANICS on a type that is not a reference, reproduced directly while
// this rule was being written. The walk recovers per FILE rather than per rule, so one such node
// costs every rule in the package its verdict on that file. Each row here reaches the filter with
// something other than an ordinary annotated function.
func TestConsistentReturnDoesNotPanicOnShapesWithNoUsefulType(t *testing.T) {
	t.Parallel()
	sources := []string{
		"declare function foo(flag: boolean): void;\n",
		"function foo(flag: boolean): any { if (flag) return true; return; }\n",
		"abstract class A { abstract foo(flag: boolean): void; }\n",
		"function foo({ a, b }: { a: number; b: number }) { if (a) return b; return; }\n",
		"const foo = function* (flag: boolean) { if (flag) return 1; return; };\n",
		"class A { get foo(): number | void { if (1) return 1; return; } }\n",
		"class A { constructor(flag: boolean) { if (flag) return; return; } }\n",
		"declare module 'x' { export function foo(): void; }\n",
		"function foo(this: void, flag: boolean): void { if (flag) return; return; }\n",
	}
	for _, source := range sources {
		t.Run(strings.TrimSpace(source), func(t *testing.T) {
			t.Parallel()
			// The assertion is that this returns at all. A panic in the rule surfaces as a failure
			// here rather than as a silently lost file.
			runConsistentReturn(t, consistentReturnCase{source: source})
		})
	}
}

// TestConsistentReturnPromiseVoidGuardsAreBothLoadBearing pins the two tests inside isPromiseVoid.
//
// `isPromiseVoid` asks two questions before it unwraps: is the type thenable, and is it a type
// reference. Neither is implied by the other, and a mutation removing either SURVIVED upstream's
// whole 30-case corpus, because every promise upstream writes answers both the same way.
//
// Measured in `internal/consistent_return_ext_probe` across ten return types, five diverge:
//
//	Promise<void>, PromiseLike<void>, Box<void>   thenable T  reference T   agree
//	void[], Map<string, void>, Pair<void>         thenable F  reference T   DIVERGE
//	a bare `interface T { then(...) }`            thenable T  reference F   DIVERGE
//
// The two directions fail differently and that is why both rows below exist. Dropping the THENABLE
// test lets an ordinary generic whose first type argument happens to be `void` suppress a real
// finding -- `Promise<Map<string, void>>` unwraps to `Map<string, void>`, then to `string`, and a
// version that unwrapped without checking thenability would walk into it. Dropping the REFERENCE
// test is worse than wrong: `Checker_getTypeArguments` panics on a non-reference, and the walk
// recovers per FILE, so one such return type costs every rule in the package its verdict on that
// file.
//
// Every expectation here was taken by driving the installed 8.67.0 extension. All six report.
func TestConsistentReturnPromiseVoidGuardsAreBothLoadBearing(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		reason string
	}{
		{
			name: "an async function returning a promise of a void array still reports",
			source: `
async function foo(flag: boolean): Promise<void[]> {
  if (flag) return [];
  return;
}
`,
			reason: "`void[]` is a type reference and is not thenable. Without the thenable test the " +
				"unwrap continues into it.",
		},
		{
			name: "an async function returning a promise of Map<string, void> still reports",
			source: `
async function foo(flag: boolean): Promise<Map<string, void>> {
  if (flag) return new Map();
  return;
}
`,
			reason: "The first type argument is `string`, so this survives a single unwrap and needs " +
				"the thenable test to stop the recursion cleanly.",
		},
		{
			name: "an async function returning a promise of a generic whose argument is void reports",
			source: `
interface Pair<A> {
  a: A;
}
async function foo(flag: boolean): Promise<Pair<void>> {
  if (flag) return null as any;
  return;
}
`,
			reason: "`Pair<void>` is a reference whose first type argument IS void. Only the thenable " +
				"test separates it from a promise of void, and dropping that test silences this.",
		},
		{
			name: "an async function returning a bare thenable interface reports without panicking",
			source: `
interface T {
  then(cb: (v: void) => void): void;
}
async function foo(flag: boolean): T {
  if (flag) return null as any;
  return;
}
`,
			reason: "Thenable but NOT a type reference, so `Checker_getTypeArguments` panics on it " +
				"without the reference guard, costing every rule in the package this file.",
		},
		{
			name: "an aliased bare thenable reports without panicking",
			source: `
interface Weird {
  then(cb: (v: void) => void): void;
}
async function foo(flag: boolean): Weird {
  if (flag) return null as any;
  return;
}
`,
			reason: "The same shape reached through an interface rather than inline, which is how it " +
				"tends to appear in real source.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runConsistentReturn(t, consistentReturnCase{source: testCase.source})
			if len(result.Diagnostics) == 0 {
				t.Errorf("expected a finding and got none. %s", testCase.reason)
			}
		})
	}
}

// TestConsistentReturnSuppressingTheFirstReturnMovesTheExpectation is the shape found on real source.
//
// This is the case that proves the void filter has to be an INPUT to the judgment rather than a
// filter over its findings, and it was found by running the rule over the ahra tree rather than
// reasoned about. Two of the tree's findings appear under the extension that the bare core rule does
// not report at all, which looks backwards for a rule that only ever suppresses.
//
// The mechanism: in an `async` method returning `Promise<void>`, the first `return;` is suppressed,
// so the first return the judgment actually sees is `return this.deviceIdPromise;`. That carries a
// value, so the expectation flips, and the method's end becomes reachable-without-a-value.
//
//	core rule       unexpectedReturnValue, at the value-carrying return
//	extension rule  missingReturn, at the method head
//
// Both verdicts confirmed by driving the installed 8.67.0 extension and the installed ESLint 10.8.1
// core over this exact source. A post-filter cannot produce this: dropping the finding a suppressed
// return generates still leaves it setting the expectation.
func TestConsistentReturnSuppressingTheFirstReturnMovesTheExpectation(t *testing.T) {
	t.Parallel()
	result := runConsistentReturn(t, consistentReturnCase{
		source: `
class DeviceId {
  private deviceIdPromise: Promise<void> | undefined;
  private deviceIdChecked = false;
  private readonly isServerSide: boolean = false;

  async ensure(refresh: boolean = false): Promise<void> {
    if (this.isServerSide) {
      return;
    }
    if (this.deviceIdPromise) {
      return this.deviceIdPromise;
    }
    if (this.deviceIdChecked && !refresh) {
      return;
    }
    this.deviceIdPromise = this.check(refresh);
  }
  async check(r: boolean): Promise<void> {}
}
`,
	})
	rule_testing.ExpectFindings(t, result, "missingReturn")
}

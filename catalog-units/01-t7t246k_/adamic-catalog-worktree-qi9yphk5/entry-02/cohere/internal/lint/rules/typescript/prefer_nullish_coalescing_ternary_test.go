package typescript

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// preferNullishCoalescingReplayed is one finding as the installed typescript-eslint 8.67.0 reported
// it: the id, where it starts, the exact text it covers, and the whole file after its one suggestion.
//
// Line and column are 1-based in the file as the typed harness writes it, which is the source
// trimmed of surrounding whitespace, so they read directly against the fixture.
type preferNullishCoalescingReplayed struct {
	id        string
	line      int
	column    int
	reported  string
	suggested string
}

type preferNullishCoalescingReplayedCase struct {
	name        string
	source      string
	optionsJson string
	findings    []preferNullishCoalescingReplayed
}

// TestPreferNullishCoalescingTernaryAndIfFire runs every upstream invalid case that reports from the
// ternary or the if-statement path, 221 of them, two of which also report a `||`.
//
// # Every expectation was replayed through the installed build, not copied from upstream's file
//
// Upstream's corpus was extracted by running its test file against a stub RuleTester, and each case
// was then linted by ESLint 10 with typescript-eslint 8.67.0 on a strict program. The replay agreed
// with upstream's own expected ids, spans and suggestion outputs on all 344 strict cases, so the rows
// below are both: what upstream asserts and what the installed rule does.
//
// They assert more than the imported `||` corpus does. A ternary's suggestion rewrites the whole
// expression, choosing the subject's spelling (the first mention, with its optional links), keeping
// one pair of its parentheses, and parenthesising the fallback by precedence; an if's suggestion
// carries comments out of the block. An id and a count see none of that, so each row pins the
// spanned text and the whole file after the suggestion.
func TestPreferNullishCoalescingTernaryAndIfFire(t *testing.T) {
	t.Parallel()
	cases := append(append([]preferNullishCoalescingReplayedCase(nil),
		preferNullishCoalescingTernaryAndIfCases...), preferNullishCoalescingBeyondTheCorpusCases...)
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runPreferNullishCoalescing(t, preferNullishCoalescingCase{
				source: testCase.source, optionsJson: testCase.optionsJson,
			})
			wantIds := make([]string, 0, len(testCase.findings))
			for _, expected := range testCase.findings {
				wantIds = append(wantIds, expected.id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			text := asTheTypedPreferNullishHarnessWroteIt(testCase.source)
			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]

				start := preferNullishCoalescingOffsetOf(t, testCase.source, expected.line, expected.column)
				if diagnostic.Range.Pos() != start || diagnostic.Range.End() != start+len(expected.reported) {
					t.Errorf("finding %d: expected the span %q at line %d column %d, got %q",
						index, expected.reported, expected.line, expected.column,
						text[diagnostic.Range.Pos():diagnostic.Range.End()])
				}

				if len(diagnostic.Suggestions) != 1 {
					t.Fatalf("finding %d: expected one suggestion, got %d", index, len(diagnostic.Suggestions))
				}
				// Applied from the back, so an earlier edit never shifts a later one's offsets.
				fixes := append([]rule.Fix(nil), diagnostic.Suggestions[0].Fixes...)
				sort.Slice(fixes, func(left, right int) bool {
					return fixes[left].Range.Pos() > fixes[right].Range.Pos()
				})
				applied := text
				for _, fix := range fixes {
					applied = applied[:fix.Range.Pos()] + fix.Text + applied[fix.Range.End():]
				}
				if applied != expected.suggested {
					t.Errorf("finding %d: the suggestion writes\n%s\nupstream's writes\n%s",
						index, applied, expected.suggested)
				}
			}
		})
	}
}

// preferNullishCoalescingBeyondTheCorpusCases report where upstream's corpus never looks, each
// replayed through the installed build the same way.
//
// A mutation sweep over the new paths left two checks no corpus case could fail. A join mixing `===`
// with `==` reads as loose, and a loose comparison rules out null and undefined both, so a test that
// names only null that way is fixable whatever the type; read as strict, `string | null | undefined`
// would make it unfixable. These rows, both arms of the join and the if path, are what kill it.
var preferNullishCoalescingBeyondTheCorpusCases = []preferNullishCoalescingReplayedCase{
	{name: "replayed: === and == joined by || read loose, so null alone is enough", source: "declare let x: string | null | undefined;\ndeclare const y: string;\nconst r = x === null || x == null ? y : x;\n", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 11, reported: `x === null || x == null ? y : x`, suggested: "declare let x: string | null | undefined;\ndeclare const y: string;\nconst r = x ?? y;\n"},
	}},
	{name: "replayed: !== and != joined by && read loose", source: "declare let x: string | null | undefined;\ndeclare const y: string;\nconst r = x !== null && x != null ? x : y;\n", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 11, reported: `x !== null && x != null ? x : y`, suggested: "declare let x: string | null | undefined;\ndeclare const y: string;\nconst r = x ?? y;\n"},
	}},
	{name: "replayed: the if path reads the same join loose", source: "declare let x: string | null | undefined;\ndeclare const y: string;\nif (x === null || x == null) { x = y; }\n", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 3, column: 1, reported: `if (x === null || x == null) { x = y; }`, suggested: "declare let x: string | null | undefined;\ndeclare const y: string;\nx ??= y;\n"},
	}},
}

// TestPreferNullishCoalescingTernaryAndIfStaySilentBeyondTheCorpus is the other survivor's control.
//
// A test that compares its subject only to itself rules out neither null nor undefined, so upstream
// declines it; a port answering "fixable" there is silent on every corpus case and wrong here. The
// strict `=== null || === null` row is the loose join's own control: the same shape read strict, on
// a type that also holds undefined, is not fixable. All four replayed silent on the installed build.
func TestPreferNullishCoalescingTernaryAndIfStaySilentBeyondTheCorpus(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"declare let x: string | null | undefined;\ndeclare const y: string;\nconst r = x === null || x === null ? y : x;\n",
		"declare let x: string | null;\ndeclare const y: string;\nconst r = x !== x ? x : y;\n",
		"declare let x: string | null;\ndeclare const y: string;\nconst r = x === x || x === x ? y : x;\n",
		"declare let x: string | null;\ndeclare const y: string;\nif (x === x) x = y;\n",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runPreferNullishCoalescing(t, preferNullishCoalescingCase{source: source}))
		})
	}
}

// TestPreferNullishCoalescingTernaryAndIfMessages pins the two new renderings whole, since the
// replay table asserts ids and the message text is ours past upstream's first sentence.
func TestPreferNullishCoalescingTernaryAndIfMessages(t *testing.T) {
	t.Parallel()
	ternary := runPreferNullishCoalescing(t, preferNullishCoalescingCase{
		source: "declare const a: string | null;\nconst x = a !== null ? a : 'b';\n",
	})
	assignment := runPreferNullishCoalescing(t, preferNullishCoalescingCase{
		source: "declare let a: string | null;\nif (!a) a = 'b';\n",
	})
	for _, check := range []struct {
		name    string
		result  rule_testing.Result
		opening string
	}{
		{"ternary", ternary,
			"Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read."},
		{"assignment", assignment,
			"Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read."},
	} {
		if len(check.result.Diagnostics) != 1 ||
			!strings.HasPrefix(check.result.Diagnostics[0].Message.Description, check.opening) {
			t.Errorf("%s: expected one finding opening with upstream's sentence %q, got %v",
				check.name, check.opening, check.result.Diagnostics)
		}
	}
}

// preferNullishCoalescingTernaryAndIfCases are upstream's invalid cases that report from the ternary
// or if-statement path, generated from the replay described on the test above.
var preferNullishCoalescingTernaryAndIfCases = []preferNullishCoalescingReplayedCase{
	{name: "upstream invalid[8]", source: `x !== undefined && x !== null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && x !== null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[9]", source: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[10]", source: `x !== undefined && x !== null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && x !== null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[11]", source: `x !== null && x !== undefined ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== null && x !== undefined ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[12]", source: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] !== null &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[13]", source: `x !== null && x !== undefined ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== null && x !== undefined ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[14]", source: `x === undefined || x === null ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === undefined || x === null ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[15]", source: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[16]", source: `x === undefined || x === null ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === undefined || x === null ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[17]", source: `x === null || x === undefined ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === null || x === undefined ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[18]", source: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] === null ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[19]", source: `x === null || x === undefined ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === null || x === undefined ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[20]", source: `undefined !== x && x !== null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x && x !== null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[21]", source: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[22]", source: `undefined !== x && x !== null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x && x !== null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[23]", source: `null !== x && x !== undefined ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x && x !== undefined ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[24]", source: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x.z[1][this[this.o]]['3'][a.b.c] &&
x.z[1][this[this.o]]['3'][a.b.c] !== undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[25]", source: `null !== x && x !== undefined ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x && x !== undefined ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[26]", source: `undefined === x || x === null ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x || x === null ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[27]", source: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[28]", source: `undefined === x || x === null ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x || x === null ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[29]", source: `null === x || x === undefined ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x || x === undefined ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[30]", source: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x.z[1][this[this.o]]['3'][a.b.c] ||
x.z[1][this[this.o]]['3'][a.b.c] === undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[31]", source: `null === x || x === undefined ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x || x === undefined ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[32]", source: `x !== undefined && null !== x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && null !== x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[33]", source: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[34]", source: `x !== undefined && null !== x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && null !== x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[35]", source: `x !== null && undefined !== x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== null && undefined !== x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[36]", source: `
x.z[1][this[this.o]]['3'][a.b.c] !== null &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] !== null &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[37]", source: `x !== null && undefined !== x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== null && undefined !== x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[38]", source: `x === undefined || null === x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === undefined || null === x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[39]", source: `
x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] === undefined ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[40]", source: `x === undefined || null === x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === undefined || null === x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[41]", source: `x === null || undefined === x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === null || undefined === x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[42]", source: `
x.z[1][this[this.o]]['3'][a.b.c] === null ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] === null ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[43]", source: `x === null || undefined === x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x === null || undefined === x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[44]", source: `undefined !== x && null !== x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x && null !== x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[45]", source: `
undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x.z[1][this[this.o]]['3'][a.b.c] &&
null !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[46]", source: `undefined !== x && null !== x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined !== x && null !== x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[47]", source: `null !== x && undefined !== x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x && undefined !== x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[48]", source: `
null !== x.z[1][this[this.o]]['3'][a.b.c] &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x.z[1][this[this.o]]['3'][a.b.c] &&
undefined !== x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[49]", source: `null !== x && undefined !== x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null !== x && undefined !== x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[50]", source: `undefined === x || null === x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x || null === x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[51]", source: `
undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x.z[1][this[this.o]]['3'][a.b.c] ||
null === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[52]", source: `undefined === x || null === x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined === x || null === x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[53]", source: `null === x || undefined === x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x || undefined === x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[54]", source: `
null === x.z[1][this[this.o]]['3'][a.b.c] ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x.z[1][this[this.o]]['3'][a.b.c] ||
undefined === x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[55]", source: `null === x || undefined === x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null === x || undefined === x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[56]", source: `x != undefined && x != null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined && x != null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[57]", source: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[58]", source: `x != undefined && x != null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined && x != null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[59]", source: `x == undefined || x == null ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined || x == null ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[60]", source: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] == null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] == null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[61]", source: `x == undefined || x == null ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined || x == null ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[62]", source: `x != undefined && x !== null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined && x !== null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[63]", source: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] != undefined &&
x.z[1][this[this.o]]['3'][a.b.c] !== null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[64]", source: `x != undefined && x !== null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined && x !== null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[65]", source: `x == undefined || x === null ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined || x === null ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[66]", source: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] == undefined ||
x.z[1][this[this.o]]['3'][a.b.c] === null
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[67]", source: `x == undefined || x === null ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined || x === null ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[68]", source: `x !== undefined && x != null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && x != null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[69]", source: `
x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] !== undefined &&
x.z[1][this[this.o]]['3'][a.b.c] != null
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[70]", source: `x !== undefined && x != null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x !== undefined && x != null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[71]", source: `undefined != x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined != x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[72]", source: `
undefined != x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined != x.z[1][this[this.o]]['3'][a.b.c]
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[73]", source: `undefined != x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined != x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[74]", source: `null != x ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null != x ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[75]", source: `null != x.z[1][this[this.o]]['3'][a.b.c] ? x.z[1][this[this.o]]['3'][a.b.c] : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null != x.z[1][this[this.o]]['3'][a.b.c] ? x.z[1][this[this.o]]['3'][a.b.c] : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[76]", source: `null != x ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null != x ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[77]", source: `undefined == x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined == x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[78]", source: `
undefined == x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined == x.z[1][this[this.o]]['3'][a.b.c]
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[79]", source: `undefined == x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `undefined == x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[80]", source: `null == x ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null == x ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[81]", source: `null == x.z[1][this[this.o]]['3'][a.b.c] ? y : x.z[1][this[this.o]]['3'][a.b.c];`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null == x.z[1][this[this.o]]['3'][a.b.c] ? y : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[82]", source: `null == x ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `null == x ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[83]", source: `x != undefined ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[84]", source: `
x.z[1][this[this.o]]['3'][a.b.c] != undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] != undefined
  ? x.z[1][this[this.o]]['3'][a.b.c]
  : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[85]", source: `x != undefined ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != undefined ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[86]", source: `x != null ? x : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != null ? x : y`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[87]", source: `x.z[1][this[this.o]]['3'][a.b.c] != null ? x.z[1][this[this.o]]['3'][a.b.c] : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] != null ? x.z[1][this[this.o]]['3'][a.b.c] : y`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[88]", source: `x != null ? x : (z = y);`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x != null ? x : (z = y)`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[89]", source: `x == undefined ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[90]", source: `
x.z[1][this[this.o]]['3'][a.b.c] == undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c];
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] == undefined
  ? y
  : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[91]", source: `x == undefined ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == undefined ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[92]", source: `x == null ? y : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == null ? y : x`, suggested: `x ?? y;
`},
	}},
	{name: "upstream invalid[93]", source: `x.z[1][this[this.o]]['3'][a.b.c] == null ? y : x.z[1][this[this.o]]['3'][a.b.c];`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x.z[1][this[this.o]]['3'][a.b.c] == null ? y : x.z[1][this[this.o]]['3'][a.b.c]`, suggested: `x.z[1][this[this.o]]['3'][a.b.c] ?? y;
`},
	}},
	{name: "upstream invalid[94]", source: `x == null ? (z = y) : x;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `x == null ? (z = y) : x`, suggested: `x ?? (z = y);
`},
	}},
	{name: "upstream invalid[95]", source: `this != undefined ? this : y;`, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 1, column: 1, reported: `this != undefined ? this : y`, suggested: `this ?? y;
`},
	}},
	{name: "upstream invalid[96]", source: `
declare let x: string | null | undefined;
x ? x : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: string | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[97]", source: `
declare let x: string | null | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: string | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[98]", source: `
declare let x: number | null | undefined;
x ? x : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: number | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[99]", source: `
declare let x: number | null | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: number | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[100]", source: `
declare let x: boolean | null | undefined;
x ? x : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: boolean | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[101]", source: `
declare let x: boolean | null | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: boolean | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[102]", source: `
declare let x: object | null | undefined;
x ? x : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: object | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[103]", source: `
declare let x: object | null | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: object | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[104]", source: `
declare let x: { n: string | null | undefined };
x.n ? x.n : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n ? x.n : y`, suggested: `declare let x: { n: string | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[105]", source: `
declare let x: { n: string | null | undefined };
!x.n ? y : x.n;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x.n ? y : x.n`, suggested: `declare let x: { n: string | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[106]", source: `
declare let x: { n: number | null | undefined };
x.n ? x.n : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n ? x.n : y`, suggested: `declare let x: { n: number | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[107]", source: `
declare let x: { n: number | null | undefined };
!x.n ? y : x.n;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x.n ? y : x.n`, suggested: `declare let x: { n: number | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[108]", source: `
declare let x: { n: boolean | null | undefined };
x.n ? x.n : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n ? x.n : y`, suggested: `declare let x: { n: boolean | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[109]", source: `
declare let x: { n: boolean | null | undefined };
!x.n ? y : x.n;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x.n ? y : x.n`, suggested: `declare let x: { n: boolean | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[110]", source: `
declare let x: { n: object | null | undefined };
x.n ? x.n : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n ? x.n : y`, suggested: `declare let x: { n: object | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[111]", source: `
declare let x: { n: object | null | undefined };
!x.n ? y : x.n;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x.n ? y : x.n`, suggested: `declare let x: { n: object | null | undefined };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[112]", source: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[113]", source: `
declare let x: { n?: { a?: string } };
x.n?.a ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[114]", source: `
declare let x: { n?: { a?: string } };
x.n?.a ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[115]", source: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a !== undefined ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[116]", source: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a !== undefined ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[117]", source: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a !== undefined ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[118]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != undefined ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[119]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != undefined ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[120]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != undefined ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != undefined ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[121]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != null ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[122]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != null ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[123]", source: `
declare let x: { n?: { a?: string } };
x.n?.a != null ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a != null ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[124]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[125]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a !== undefined && x.n.a !== null ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[126]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[127]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[128]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[129]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[130]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[131]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[132]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[133]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a !== undefined ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[134]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != undefined ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[135]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != undefined ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[136]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != undefined ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[137]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != undefined ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != undefined ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[138]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != null ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[139]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != null ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[140]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != null ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[141]", source: `
declare let x: { n?: { a?: string } };
x?.n?.a != null ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a != null ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[142]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[143]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined && x.n.a !== null ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[144]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined && x.n.a !== null ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[145]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a !== undefined && x.n.a !== null ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a !== undefined && x.n.a !== null ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[146]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[147]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[148]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[149]", source: `
declare let x: { n?: { a?: string | null } };
x?.n?.a ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x?.n?.a ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x?.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[150]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x?.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[151]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n?.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x.n?.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[152]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x?.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x?.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[153]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? x.n.a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? x.n.a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[154]", source: `
declare let x: { n?: { a?: string | null } };
x.n?.a ? (x?.n).a : y;
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x.n?.a ? (x?.n).a : y`, suggested: `declare let x: { n?: { a?: string | null } };
x.n?.a ?? y;
`},
	}},
	{name: "upstream invalid[232]", source: `
declare let x: string | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: string | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[233]", source: `
declare let x: number | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: number | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[234]", source: `
declare let x: boolean | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: boolean | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[235]", source: `
declare let x: bigint | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: bigint | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[241]", source: `
declare let x: '' | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: '' | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[242]", source: "\ndeclare let x: `` | undefined;\nx ? x : y;\n      ", optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: "declare let x: `` | undefined;\nx ?? y;\n"},
	}},
	{name: "upstream invalid[243]", source: `
declare let x: 0 | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 0 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[244]", source: `
declare let x: 0n | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 0n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[245]", source: `
declare let x: false | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: false | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[251]", source: `
declare let x: 'a' | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 'a' | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[252]", source: `
declare let x: 'a' | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 'a' | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[253]", source: "\ndeclare let x: `hello${'string'}` | undefined;\nx ? x : y;\n      ", optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: "declare let x: `hello${'string'}` | undefined;\nx ?? y;\n"},
	}},
	{name: "upstream invalid[254]", source: "\ndeclare let x: `hello${'string'}` | undefined;\n!x ? y : x;\n      ", optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: "declare let x: `hello${'string'}` | undefined;\nx ?? y;\n"},
	}},
	{name: "upstream invalid[255]", source: `
declare let x: 1 | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 1 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[256]", source: `
declare let x: 1 | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 1 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[257]", source: `
declare let x: 1n | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[258]", source: `
declare let x: 1n | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[259]", source: `
declare let x: true | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: true | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[260]", source: `
declare let x: true | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: true | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[268]", source: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 'a' | 'b' | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[269]", source: `
declare let x: 'a' | 'b' | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 'a' | 'b' | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[270]", source: "\ndeclare let x: 'a' | `b` | undefined;\nx ? x : y;\n      ", optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: "declare let x: 'a' | `b` | undefined;\nx ?? y;\n"},
	}},
	{name: "upstream invalid[271]", source: "\ndeclare let x: 'a' | `b` | undefined;\n!x ? y : x;\n      ", optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":true,\"string\":false}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: "declare let x: 'a' | `b` | undefined;\nx ?? y;\n"},
	}},
	{name: "upstream invalid[272]", source: `
declare let x: 0 | 1 | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 0 | 1 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[273]", source: `
declare let x: 0 | 1 | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 0 | 1 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[274]", source: `
declare let x: 1 | 2 | 3 | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 1 | 2 | 3 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[275]", source: `
declare let x: 1 | 2 | 3 | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 1 | 2 | 3 | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[276]", source: `
declare let x: 0n | 1n | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 0n | 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[277]", source: `
declare let x: 0n | 1n | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 0n | 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[278]", source: `
declare let x: 1n | 2n | 3n | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 1n | 2n | 3n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[279]", source: `
declare let x: 1n | 2n | 3n | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 1n | 2n | 3n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[280]", source: `
declare let x: true | false | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: true | false | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[281]", source: `
declare let x: true | false | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: true | false | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[284]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[285]", source: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":false,\"boolean\":true,\"number\":false,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[286]", source: `
declare let x: true | false | null | undefined;
x ? x : y;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `x ? x : y`, suggested: `declare let x: true | false | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[287]", source: `
declare let x: true | false | null | undefined;
!x ? y : x;
      `, optionsJson: "{\"ignorePrimitives\":{\"bigint\":true,\"boolean\":false,\"number\":true,\"string\":true}}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x ? y : x`, suggested: `declare let x: true | false | null | undefined;
x ?? y;
`},
	}},
	{name: "upstream invalid[303]", source: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ? a : b);
      `, optionsJson: "{\"ignoreBooleanCoercion\":true}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 4, column: 19, reported: `a ? a : b`, suggested: `let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ?? b);
`},
	}},
	{name: "upstream invalid[304]", source: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(!a ? b : a);
      `, optionsJson: "{\"ignoreBooleanCoercion\":true}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 4, column: 22, reported: `!a ? b : a`, suggested: `let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(a ?? b);
`},
	}},
	{name: "upstream invalid[308]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ? defaultBox : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBox ? defaultBox : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[309]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a?.b : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b != null ? defaultBoxOptional.a?.b : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[312]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a.b : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b != null ? defaultBoxOptional.a.b : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[313]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a?.b : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b ? defaultBoxOptional.a?.b : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[314]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a.b : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b ? defaultBoxOptional.a.b : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[315]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[316]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[317]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a?.b
  : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[318]", source: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `, optionsJson: "{\"ignoreTernaryTests\":false}", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 7, column: 1, reported: `defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a.b
  : getFallbackBox()`, suggested: `interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
`},
	}},
	{name: "upstream invalid[319]", source: `
declare let x: unknown;
declare let y: number;
!x ? y : x;
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 1, reported: `!x ? y : x`, suggested: `declare let x: unknown;
declare let y: number;
x ?? y;
`},
	}},
	{name: "upstream invalid[320]", source: `
declare let x: unknown;
declare let y: number;
x ? x : y;
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 1, reported: `x ? x : y`, suggested: `declare let x: unknown;
declare let y: number;
x ?? y;
`},
	}},
	{name: "upstream invalid[321]", source: `
declare let x: { n: unknown };
!x.n ? y : x.n;
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `!x.n ? y : x.n`, suggested: `declare let x: { n: unknown };
x.n ?? y;
`},
	}},
	{name: "upstream invalid[322]", source: `
declare let x: { a: string } | null;

x?.['a'] != null ? x['a'] : 'foo';
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 1, reported: `x?.['a'] != null ? x['a'] : 'foo'`, suggested: `declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
`},
	}},
	{name: "upstream invalid[323]", source: `
declare let x: { a: string } | null;

x?.['a'] != null ? x.a : 'foo';
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 1, reported: `x?.['a'] != null ? x.a : 'foo'`, suggested: `declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
`},
	}},
	{name: "upstream invalid[324]", source: `
declare let x: { a: string } | null;

x?.a != null ? x['a'] : 'foo';
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 3, column: 1, reported: `x?.a != null ? x['a'] : 'foo'`, suggested: `declare let x: { a: string } | null;

x?.a ?? 'foo';
`},
	}},
	{name: "upstream invalid[325]", source: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x[a] : 'foo';
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 4, column: 1, reported: `x?.[a] != null ? x[a] : 'foo'`, suggested: `const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] ?? 'foo';
`},
	}},
	{name: "upstream invalid[326]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (!foo) {
    foo = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[327]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) {
    foo = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[328]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) {
    foo ??= makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[329]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ||= makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) {
    foo ||= makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
		{id: "preferNullishOverOr", line: 6, column: 9, reported: `||=`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
`},
	}},
	{name: "upstream invalid[330]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === null) {
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo === null) {
    foo = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[331]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
  const bar = 42;
  return bar;
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) foo = makeFoo();`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
	}},
	{name: "upstream invalid[332]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) foo ??= makeFoo();`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
	}},
	{name: "upstream invalid[333]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ||= makeFoo();
  const bar = 42;
  return bar;
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) foo ||= makeFoo();`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
		{id: "preferNullishOverOr", line: 5, column: 24, reported: `||=`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
`},
	}},
	{name: "upstream invalid[334]", source: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined) {
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo === undefined) {
    foo = makeFoo();
  }`, suggested: `declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[335]", source: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined || foo === null) {
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo === undefined || foo === null) {
    foo = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[336]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo.a == null) {
    foo.a = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo.a == null) {
    foo.a = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[337]", source: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo?.a == null) {
    foo.a = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo?.a == null) {
    foo.a = makeFoo();
  }`, suggested: `declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[338]", source: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // comment
    foo = makeFoo();
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (foo == null) {
    // comment
    foo = makeFoo();
  }`, suggested: `declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
foo ??= makeFoo();
}
`},
	}},
	{name: "upstream invalid[339]", source: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 4, column: 1, reported: `if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}`, suggested: `declare let foo: string | null;
declare function makeFoo(): string;

// comment before 1
/* comment before 2 */
/* comment before 3
    which is multiline
  */
/**
   * comment before 4
   * which is also multiline
   */
foo ??= makeFoo(); // comment inline
// comment after 1
/* comment after 2 */
/* comment after 3
    which is multiline
  */
/**
   * comment after 4
   * which is also multiline
   */
`},
	}},
	{name: "upstream invalid[340]", source: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo(); // comment inline
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 4, column: 1, reported: `if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo();`, suggested: `declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
`},
	}},
	{name: "upstream invalid[341]", source: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }
}
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverAssignment", line: 5, column: 3, reported: `if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }`, suggested: `declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
`},
	}},
	{name: "upstream invalid[342]", source: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : b ? 1 : 2;
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 4, column: 13, reported: `a ? a : b ? 1 : 2`, suggested: `let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
`},
	}},
	{name: "upstream invalid[343]", source: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : (b ? 1 : 2);
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 4, column: 13, reported: `a ? a : (b ? 1 : 2)`, suggested: `let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
`},
	}},
	{name: "upstream invalid[344]", source: `
declare const c: string | null;
c !== null ? c : c ? 1 : 2;
      `, optionsJson: "", findings: []preferNullishCoalescingReplayed{
		{id: "preferNullishOverTernary", line: 2, column: 1, reported: `c !== null ? c : c ? 1 : 2`, suggested: `declare const c: string | null;
c ?? (c ? 1 : 2);
`},
	}},
}

// TestPreferNullishCoalescingIfIsSilentInsideReactCompiledFunctions covers the `reactCompiler` option,
// which is ours rather than upstream's, on the if-statement check it gates.
//
// The rows are shared verbatim with Nexus's NexusTypeScriptEsLintPlugin.test.ts, whose wrapper draws the
// same line in ESLint, so one table proves the ruling in both engines (#cn8sthd). Three are the sites
// the waves rewrote to `??=`, each of which lost its compilation to the rewrite, in their long form:
// SecretRow, WebSocketViaSharedWorkerProviderInternal and ChatReasoningAndTools' component body. The
// rest pin where the compiled region ends, from the measurements recorded in the react shelf's
// compiled.go: a bare block, an outer function whose name claims nothing, a call wrapper, a non-ASCII
// capital, primitive props, a second parameter that is not a ref, and JSX only inside a closure.
//
// Every row runs with the compiler on and off, so a silent row is shown to be the gate rather than a
// shape the check never reported. The harness has no React types, so each ref is annotated: read as
// `any`, its `current` is not nullable and the check would never fire.
func TestPreferNullishCoalescingIfIsSilentInsideReactCompiledFunctions(t *testing.T) {
	t.Parallel()

	decode := func(raw string) PreferNullishCoalescingOptions {
		decoded, err := DecodePreferNullishCoalescingOptions(json.RawMessage(raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		return decoded.(PreferNullishCoalescingOptions)
	}
	compilerOn := decode(`{}`)
	compilerOff := decode(`{"reactCompiler": false}`)

	for _, testCase := range preferNullishCoalescingCompiledRows {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			off := rule_testing.RunTypedWithOptions(t, PreferNullishCoalescing, "Compiled.tsx", testCase.sourceText, compilerOff)
			rule_testing.ExpectFindings(t, off, "preferNullishOverAssignment")
			on := rule_testing.RunTypedWithOptions(t, PreferNullishCoalescing, "Compiled.tsx", testCase.sourceText, compilerOn)
			if testCase.compiled {
				rule_testing.ExpectClean(t, on)
				return
			}
			rule_testing.ExpectFindings(t, on, "preferNullishOverAssignment")
		})
	}

	// The ternary's repair is `??`, an expression the compiler lowers, so it still reports in a
	// component with the compiler on.
	ternary := rule_testing.RunTypedWithOptions(t, PreferNullishCoalescing, "Compiled.tsx", `
export function Badge(properties: { label: string | null }) {
    const label = properties.label !== null ? properties.label : 'none';
    return <span>{label}</span>;
}`, compilerOn)
	rule_testing.ExpectFindings(t, ternary, "preferNullishOverTernary")
}

// preferNullishCoalescingCompiledRows is the table both engines run: each source holds one `if` the
// check reports with the compiler off, and `compiled` says whether React Compiler compiles the function
// holding it.
var preferNullishCoalescingCompiledRows = []struct {
	name       string
	sourceText string
	compiled   bool
}{
	{"a component", `export function Badge(properties: { label: string | null }) {
    let label = properties.label;
    if(label === null) label = 'none';
    return <span>{label}</span>;
}
`, true},
	{"an async callback inside a component, as SecretRow", `import React from 'react';
declare function reveal(): Promise<string | null>;
export function SecretRow(properties: { revealed: string | null }) {
    const onCopy = React.useCallback(async function() {
        let value: string | null = properties.revealed;
        if(value === null) {
            value = await reveal();
        }
        return value;
    }, [properties.revealed]);
    return <button onClick={onCopy}>copy</button>;
}
`, true},
	{"an effect callback inside a provider, as WebSocketViaSharedWorkerProviderInternal", `import React from 'react';
declare function createMonitor(): object;
export function WebSocketProviderInternal(properties: { children: React.ReactNode }) {
    const monitorReference: { current: object | null } = React.useRef(null);
    React.useEffect(function() {
        if(monitorReference.current === null) {
            monitorReference.current = createMonitor();
        }
    }, []);
    return <>{properties.children}</>;
}
`, true},
	{"a hook body", `import React from 'react';
export function useStart(initial: number | null) {
    const [count] = React.useState(0);
    let start = initial;
    if(start === null) start = count;
    return start;
}
`, true},
	{"a hook named with a digit after use", `import { useState } from 'react';
export function use2Things(initial: number | null) {
    const [count] = useState(0);
    let start = initial;
    if(start === null) start = count;
    return start;
}
`, true},
	{"a component inside a component-named function with no evidence of its own", `export function Outer() {
    const Inner = (properties: { label: string | null }) => {
        let label = properties.label;
        if(label === null) label = 'none';
        return <span>{label}</span>;
    };
    return Inner;
}
`, true},
	{"a component that takes a ref second", `export function Field(properties: { label: string | null }, forwardedRef: unknown) {
    let label = properties.label;
    if(label === null) label = 'none';
    return <span ref={forwardedRef}>{label}</span>;
}
`, true},
	{"a helper outside any component", `export function normalize(stored: string | null) {
    let label = stored;
    if(label === null) label = 'none';
    return label;
}
`, false},
	{"a hook-named function that calls no hook and writes no JSX", `export function useLabel(stored: string | null) {
    let label = stored;
    if(label === null) label = 'none';
    return label;
}
`, false},
	{"a component-named function with no JSX and no hooks", `export function Defaults(properties: { label: string | null }) {
    let label = properties.label;
    if(label === null) label = 'none';
    return label;
}
`, false},
	{"a component inside a bare block", `{
    function Badge(properties: { label: string | null }) {
        let label = properties.label;
        if(label === null) label = 'none';
        return <span>{label}</span>;
    }
    console.info(Badge);
}
`, false},
	{"a component inside a function whose name claims nothing", `export function plainOuter() {
    const Inner = (properties: { label: string | null }) => {
        let label = properties.label;
        if(label === null) label = 'none';
        return <span>{label}</span>;
    };
    return Inner;
}
`, false},
	{"a component wrapped in a call, which names nothing", `declare function memo<Type>(component: Type): Type;
export const Badge = memo(function(properties: { label: string | null }) {
    let label = properties.label;
    if(label === null) label = 'none';
    return <span>{label}</span>;
});
`, false},
	{"a component name with a non-ASCII capital", `export function Émile(properties: { label: string | null }) {
    let label = properties.label;
    if(label === null) label = 'none';
    return <span>{label}</span>;
}
`, false},
	{"a component whose props are annotated as a primitive", `export function Badge(text: string) {
    let label: string | null = text.length > 0 ? text : null;
    if(label === null) label = 'none';
    return <span>{label}</span>;
}
`, false},
	{"a component whose second parameter is not a ref", `export function Badge(properties: { label: string | null }, other: unknown) {
    let label = properties.label;
    if(label === null) label = 'none';
    return <span>{label}{String(other)}</span>;
}
`, false},
	{"a component whose only JSX is inside a closure", `export function Badge(properties: { label: string | null }) {
    let label = properties.label;
    if(label === null) label = 'none';
    return () => <span>{label}</span>;
}
`, false},
}

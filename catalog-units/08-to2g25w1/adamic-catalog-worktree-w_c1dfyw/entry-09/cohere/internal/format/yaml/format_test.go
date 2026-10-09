package yaml

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/differential"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/oracletest"
	"github.com/system-inc/cohere/internal/format/printing"
)

// formatFixtures reach the printer's branches: every node type, comments in every slot they attach to,
// flow collections that break, each block scalar header, the quote choices, prettier-ignore, document
// markers and directives. Each is parsed by the real parser and printed by the port, and formatted by
// the embedded Prettier, under every option set below.
var formatFixtures = []string{
	// Empty and trivial files.
	"", "\n", "\n\n\n", "a", "a\n", "  a  \n", "# c", "# c\n", "#", "---", "---\n", "...", "...\n",

	// Plain scalars, single and multi-line, and their whitespace.
	"a b  c", "a:    b", "a: b   c  ", "a: b\tc", "a:\n  b\n  c", "a: b\n  c\n\n  d", "a: b\n\n\n  c",
	"- a\n  b", "a\n b\n  c", "a: 'x'\nb: c\n", "a: é\nb: 🍐 x\nc: 中文 中文", "a: \u00a0b", "key: value # c",
	"a: b\n   c   \n   d", "a: 1\nb: 2.5\nc: true\nd: null\ne: ~", "a: -1\nb: .inf",

	// Quotes: which quote, escapes, embedded quotes, and multi-line quoted scalars.
	`"a"`, `'a'`, `a: "b"`, `a: 'b'`, `"a'b"`, `'a"b'`, `"a\"b"`, `'a''b'`, `"a\nb"`, `'a\b'`, `"a\"b'c"`,
	`'a''b"c'`, `"a\\"`, `"\""`, `"\\\""`, `''''`, `""`, `''`, "\"a\n  b\"", "'a\n\n  b'", "\"a \\\n  b\"",
	"a: \"multi\n  line  \n  text\"", "- 'x''y'\n- \"x'y\"", `"a\tb"`, `"\u263A"`,

	// Block scalars: style, chomping, explicit indentation, indicator comments, trailing lines.
	"a: |\n  x\n  y\n", "a: >\n  x\n  y\n", "a: |-\n  x\n", "a: |+\n  x\n", "a: >-\n  x\n", "a: >+\n  x\n\n",
	"a: |2\n   x\n  y\n", "a: |1\n  x\n", "a: >2-\n   x\n", "a: |+\n  x\n\nb: c", "a: |\n  x\n\n\nb: c",
	"a: |\n  x\n\n\n", "a: | # c\n  x\n", "a: >- # c\n  x\n  y", "|\n  x\n", ">\n  x\n\n  y\n", "- |\n  x\n- y",
	"a:\n  b: |2\n     x\n    y", "a: >\n  a\n    b\n  c\n", "a: |\nb: c", "a: |\n", "a: >\n\n  x\n",
	"a: |+\n  x\n\n\n", "- |+\n  x\n\n", "a:\n  - |-\n    x\n  - >+\n    y\n\nb: c", "a: |\n  x  \n  y\t\n",
	"a: >\n  x\n\n\n  y\n", "a: |\n    x\n  y\n", "a: >\n  é 🍐\n  中文\n", "--- |\n  x\n", "--- >-\n  x\n...\n",
	"a: |\n  x\n# c\n", "a: |\n  x\n\n# c\n", "a: |+\n  x\n# c\n", "- >\n  long folded text that goes on and on\n  and on\n",

	// Anchors, aliases, tags.
	"a: &x 1\nb: *x", "*x : a", "&a a: b", "!!str a", "!t &a b", "&a !t b", "!!map\na: b", "&a\n- b",
	"a: !!seq\n- b", "a: &x\n  b: c", "a: !t\n  - b", "--- !t\na: b", "!!set\n? a\n? b", "? a\n? b",
	"a: !!str", "a: &x", "- !t\n- &a\n- *a", "a: !<tag:x> b", "a: !local b", "[*a, &b c, !t d]",
	"{*a : b, &c d: e}", "a: !t # c\n  b", "a: &x # c\n  # d\n  b: c", "!t\n# c\nb",

	// Comments in every slot: leading, trailing, middle, end, head, indicator, document.
	"# c\na: b", "a: b # c", "a: # c\n  b", "a:\n  # c\n  b", "a:\n  - b\n  # c\n", "a:\n  b: c\n  # d",
	"a: b\n# c", "a: b\n\n# c", "a:\n  b: c\n\n# d", "# c\n---\na: b", "--- # c\na", "a\n... # c",
	"a: b\n# c\nd: e", "a: b\n\n\nc: d", "- a # c\n- b", "- # c\n  a", "- # c", "a: # c",
	"? a # c\n: b", "? a\n# c\n: b", "a: # c\n  - b", "a:\n  - b # c\n  - c", "- a\n  # c\n- b",
	"- - a\n  # c\n- b", "a:\n  b:\n    c: d\n    # e\n  # f\n# g", "# a\n\n# b\n\nc: d", "a: b # c\n# d",
	"a: [b] # c", "a: {b: c} # d", "a:\n  # b\n\n  # c\n  d: e", "%YAML 1.2 # c\n---\na",
	"# c\n%YAML 1.2\n---\na", "a:\n- b\n# c\n\n# d", "a: b\n  # c", "? # c\n  a\n: b",
	"key: # c\n  value", "a: !t # c\n  - b",

	// prettier-ignore.
	"# prettier-ignore\na:   [1,   2]\nb:   [1,  2]", "a:\n  # prettier-ignore\n  b:   {c:   d}\n  e:   {f:   g}",
	"# prettier-ignore\n---\na:   [b,   c]", "- # prettier-ignore\n  a:   [b,   c]", "#  prettier-ignore  \nkey  :  value",
	"# prettier-ignore\na: |\n    x   \n", "a: 1\n# prettier-ignore\nb:    [1,   2]   \n\nc:   3",

	// Flow collections: empty, nested, long enough to break, comments and empty lines inside.
	"[a, b, c]", "{a: 1, b: 2}", "[]", "{}", "[ ]", "{ }", "a: []", "a: {}", "[a, b, ]", "{a, b}", "{a: }",
	"[a: b]", "{? a}", "{: b}", "[? a : b]", "[a: b, c: d]", "{a: [1, 2], b: {c: d}}", "[[a], [b, [c]]]",
	"[{a: b}, {c: d}]", "[\n  a,\n\n  b\n]", "[\n  a\n  # c\n]", "{\n  a: b,\n  # c\n}", "[a, # c\n b]",
	"{a: 1, # c\n b: 2}", "a: [] # c", "[a\n b]", "{a: b\n c}", "[\"a\", 'b']",
	"key: [aaaaaaaa, bbbbbbbbb, cccccccccc, dddddddddd, eeeeeeeeee, ffffffffff, gggggggggg, hhhhhhhhhh, iiiiiiiiii, jjjjjjjjjj]",
	"key: {aaaaaaaa: 1, bbbbbbbbb: 2, cccccccccc: 3, dddddddddd: 4, eeeeeeeeee: 5, ffffffffff: 6, gggggggggg: 7, hh: 8}",
	"- [aaaaaaaa, bbbbbbbbb, cccccccccc, dddddddddd, eeeeeeeeee]\n- {a: [bbbbbbbbbbbbbbbbbb, cccccccccccccccccc]}",
	"[a, b]: c", "{a: b}: c", "? [a, b]\n: c", "a: [b,\n\n  c, d]", "{a: b, c: }", "{a, b: c, }", "[a, {b: c}, ]",
	"a: {\n}", "a: [\n  # c\n]", "a: {\n  # c\n}", "a: { } # c\n", "{a: [b]}: c", "[,]",

	// Mapping items: explicit and implicit keys, empty keys and values, long keys.
	"? a\n: b", "? - a\n: b", "? |\n  a\n: b", "? a\n", ": b", "? \n: ", ":", "? a\n  b\n: c", "? !t a\n: b",
	"a:\n  # c\n  b", "a: !!map\n  b: c", "a: &x\n  - b", "a:\n  b:\n    c:\n      d: e",
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa: b",
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:\n  - b",
	"\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\": b",
	"a: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	"a: bbbbbbbbbbbbbbbbbbbbbbbbbbb cccccccccccccccccccccccccc dddddddddddddddddddddddddddddd eeeeeeeeeeeeeeeeeeee fffffffffffff",
	"a: \"bbbbbbbbbbbbbbbbbbbbbbbbbbb cccccccccccccccccccccccccc dddddddddddddddddddddddd eeeeeeeeeeeeeeeeeeee fffffffffffff\"",
	"a: b\n  c", "a b: c", "a:b: c", "a:\n  b", "a:\n\n  b", "a: \n  - b\n  - c", "a:\n- b\n- c", "a:\n  -\n  - b",
	"? a\n: - b\n  - c", "? a\n: |\n  b", "*a :", "*a : b", "? *a\n: b", "&a a:", "{*a : b}", "a: [b]\nc: {d: e}",
	"? a\n: b\n? c\n: d", "a: &x b\nc: *x\n", "!!set\n? aaaaaaaa\n? b", "? \"a\n  b\"\n: c", "'a\n  b': c",

	// Sequences.
	"- a\n- b", "- - a\n  - b", "- a: b\n  c: d", "-\n  a", "- ", "-", "- - - a", "- a\n\n- b", "- a\n\n\n- b",
	"-   a\n-   b", "- [a]\n- {b: c}", "- ? a\n  : b", "- &a\n  - b", "- !t\n  a: b", "a:\n  - b\n  -\n    - c",

	// Documents, markers and directives.
	"---\na\n---\nb", "a\n...\n---\nb", "%YAML 1.2\n---\na", "%TAG ! tag:x,2000:\n---\n!a b", "---\n---",
	"--- a", "a\n--- # c\nb", "a\n...\n# c\n---\nb", "%YAML 1.2\n---\na\n...\n%YAML 1.2\n---\nb",
	"a\n...\n", "---\na\n...\n", "--- # c\n", "---\n# c\n", "a\n---\n", "---\n...\n---\n...\n", "# c\n---\n# d\n...",
	"%FOO bar baz\n---\na", "---\n- a\n---\n- b\n...\n", "a\n---\n\n\nb", "--- !!map\na: b\n--- !!seq\n- c",

	// Edges: an empty last flow item, end comments on a value, end comments after an empty line, a
	// document's own trailing comment, blank and whitespace-only block scalars, tabs inside folded text.
	"{a: 1, : }", "[a, : ]", "{ : }", "[ : ]", "{a, :}", "a: []\n  # c", "a: {}\n  # c\n", "a:\n  # c",
	"a: b\n  # c\nd: e", "a:\n  - b\n\n  # c", "a:\n  b: c\n\n  # d\ne: f", "- a\n\n  # c\n- b",
	"a: b\n... # c\n---\nc", "a\n... # c\n", "--- a\n... # c\n--- b", "a: |\n  \n\nb: c", "a: |\n   \n",
	"a: >\n\n\n", "- |\n\n- b", "a: >\n  x\t y\n", "a: >\n  x\t y\n  z\n", "a: |\n  x\n\n\n\nb: c",
	"- |\n  x\n\n\n- y", "a:\n  b: |\n    x\n\n\n  c: d", "a: [] # c\n  # d", "a: !t []\n  # c",
	"? a\n  b\n: # c", "\"a\n  b\": # c",

	// Line endings, byte order marks, tabs, non-ASCII and astral characters.
	"a: b\r\nc: d\r\n", "a: |\r\n  x\r\n  y\r\n", "\ufeffa: b\n", "a: b\rc: d", "a:\tb", "- é\n- 中文\n- 🍐🍐",
	"\"é 🍐\": '中'", "a: \"\u2028\"", "a: |\n  🍐 x\n  中\n", "a: >\n  aaaa bbbb 中文 cccc 🍐 dddd eeee ffff gggg hhhh\n",
}

// formatVariants are the option sets every fixture runs under: our defaults (tabWidth 4, single quotes,
// printWidth 120), narrow (printWidth 40, tabWidth 2, double quotes), tabs, and the less common
// settings the printer reads (no bracket spacing, no trailing comma, CRLF).
var formatVariants = func() []formatoptions.Options {
	defaults := formatoptions.Default()
	narrow := formatoptions.Default()
	narrow.PrintWidth, narrow.TabWidth, narrow.SingleQuote = 40, 2, false
	tabs := formatoptions.Default()
	tabs.UseTabs, tabs.SingleQuote, tabs.PrintWidth, tabs.EndOfLine = true, false, 40, "cr"
	other := formatoptions.Default()
	other.PrintWidth, other.BracketSpacing, other.TrailingComma, other.EndOfLine = 40, false, "none", "crlf"
	return []formatoptions.Options{defaults, narrow, tabs, other}
}()

// formatParsed prints a parsed input the way Prettier's core wraps the printer: the byte order mark comes
// back on the output.
func formatParsed(fileName string, tree parsed, text string, hasByteOrderMark bool, options formatoptions.Options, proseWrap string, textToDoc printing.TextToDoc) (string, error) {
	if tree.err != "" {
		return "", fmt.Errorf("the parser failed: %s", tree.err)
	}
	formatted, err := printFile(fileName, tree.root, text, options, proseWrap, textToDoc, 0)
	if err != nil {
		return "", err
	}
	if hasByteOrderMark {
		formatted = "\ufeff" + formatted
	}
	return formatted, nil
}

// formatFullStack is Format as native.Formatter calls it: the byte order mark removed and line endings
// normalized before, the mark restored after.
func formatFullStack(fileName string, input string, options formatoptions.Options, proseWrap string, textToDoc printing.TextToDoc) (string, error) {
	text, hasByteOrderMark := normalizeInput(input)
	formatted, err := formatWithProseWrap(fileName, text, options, proseWrap, textToDoc)
	if err != nil {
		return "", err
	}
	if hasByteOrderMark {
		formatted = "\ufeff" + formatted
	}
	return formatted, nil
}

// parseInputs normalizes and parses every input.
func parseInputs(t testing.TB, inputs []string) ([]string, []bool, []parsed) {
	t.Helper()
	texts := make([]string, len(inputs))
	byteOrderMarks := make([]bool, len(inputs))
	for index, input := range inputs {
		texts[index], byteOrderMarks[index] = normalizeInput(input)
	}
	return texts, byteOrderMarks, parseTrees(t, texts)
}

// compareFormat formats every input with Prettier and with the port under every variant, the port two
// ways: the full stack (Format: the Go parser and the printer, from the text) and the printer alone on
// the tree the real parser built (the tree loader). Each difference is reported. Where Prettier fails,
// both must fail too, and Format with an error printing.IsSyntax recognizes.
func compareFormat(t *testing.T, inputs []string) int {
	t.Helper()
	texts, byteOrderMarks, trees := parseInputs(t, inputs)
	failures, fullStackFailures, compared, oracleFailures := 0, 0, 0, 0
	// The oracle's answers, recorded (see oracletest), so no JavaScript runs here.
	golden := oracletest.Open(t, t.Name())
	for variantIndex, options := range formatVariants {
		engine := golden.Engine(options)
		for index, input := range inputs {
			expected, oracleErr := engine.Format("fixture.yaml", input)
			fullStack, fullStackErr := formatFullStack("fixture.yaml", input, options, "preserve", nil)
			if oracleErr != nil {
				// Prettier cannot format it; the parser must refuse it too.
				oracleFailures++
				if trees[index].err == "" {
					failures++
					t.Errorf("variant %d, %q: the oracle failed (%v) but the parser did not", variantIndex, input, oracleErr)
				}
				if fullStackErr == nil || !printing.IsSyntax(fullStackErr) {
					fullStackFailures++
					t.Errorf("variant %d, %q: the oracle failed (%v) but Format gave %q, %v", variantIndex, input, oracleErr, fullStack, fullStackErr)
				}
				continue
			}
			compared++
			if fullStackErr != nil {
				fullStackFailures++
				t.Errorf("variant %d, %q: Format: %v", variantIndex, input, fullStackErr)
			} else if fullStack != expected {
				fullStackFailures++
				t.Errorf("variant %d, %q: Format: %s", variantIndex, input, differential.FirstDifference(expected, fullStack))
			}
			actual, err := formatParsed("fixture.yaml", trees[index], texts[index], byteOrderMarks[index], options, "preserve", nil)
			if err != nil {
				failures++
				t.Errorf("variant %d, %q: %v", variantIndex, input, err)
				continue
			}
			if actual != expected {
				failures++
				t.Errorf("variant %d, %q: %s", variantIndex, input, differential.FirstDifference(expected, actual))
			}
		}
	}
	t.Logf("%d comparisons over %d fixtures and %d option sets: full stack %d differ, tree loader %d differ; %d oracle failures (syntax errors)",
		compared, len(inputs), len(formatVariants), fullStackFailures, failures, oracleFailures)
	return failures + fullStackFailures
}

func TestFormatFixturesMatchOracle(t *testing.T) {
	t.Parallel()
	compareFormat(t, formatFixtures)
}

// TestFormatOracleCanFail proves the comparison sees a difference: the port's output for one input
// against the oracle's for another.
func TestFormatOracleCanFail(t *testing.T) {
	t.Parallel()
	engine := oracletest.Open(t, t.Name()).Engine(formatoptions.Default())
	expected, err := engine.Format("fixture.yaml", "a: [b]")
	if err != nil {
		t.Fatal(err)
	}
	texts, byteOrderMarks, trees := parseInputs(t, []string{"a: [b, c]"})
	if byteOrderMarks[0] {
		t.Fatal("the fixture has no byte order mark")
	}
	actual, err := Print(trees[0].root, texts[0], formatoptions.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if actual == expected {
		t.Fatalf("the comparison found no difference: both printed %q", actual)
	}
}

// TestPrintDocLeavesTheTrailingHardline checks the embedding contract: PrintDoc's doc ends in the
// hardline that upstream's textToDoc strips, and the caller strips it.
func TestPrintDocLeavesTheTrailingHardline(t *testing.T) {
	t.Parallel()
	texts, _, trees := parseInputs(t, []string{"a: b\n"})
	document, err := PrintDoc(trees[0].root, texts[0], formatoptions.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	printed := docString(document)
	if printed != "a: b\n" {
		t.Fatalf("PrintDoc laid out is %q", printed)
	}
	if stripped := docString(stripTrailingHardline(document)); stripped != "a: b" {
		t.Fatalf("stripped, it is %q", stripped)
	}
	if strings.Contains(printed, "\r") {
		t.Fatal("a doc has no line endings of its own")
	}
}

const suiteScript = `
import suite from "FORK/node_modules/yaml-test-suite/index.js";
const inputs = [];
for (const test of suite) {
  for (const testCase of test.cases) {
    inputs.push(testCase.yaml);
    if (typeof testCase.dump === "string") inputs.push(testCase.dump);
  }
}
process.stdout.write(JSON.stringify(inputs));
`

// suiteInputs is every input of the yaml-test-suite in the fork's node_modules, and each case's canonical
// dump.
func suiteInputs(t *testing.T) []string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	script := filepath.Join(t.TempDir(), "suite.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(suiteScript, "FORK", forkRoot(t))), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", script).Output()
	if err != nil {
		t.Skipf("the yaml-test-suite is not loadable: %v", err)
	}
	var inputs []string
	if err := json.Unmarshal(output, &inputs); err != nil {
		t.Fatal(err)
	}
	if len(inputs) < 400 {
		t.Fatalf("the suite gave %d inputs", len(inputs))
	}
	return inputs
}

// TestYamlTestSuiteMatchesOracle runs the suite through both sides under every option set. Inputs the
// oracle cannot format (the suite's error cases) must fail in the parser too.
func TestYamlTestSuiteMatchesOracle(t *testing.T) {
	t.Parallel()
	compareFormat(t, suiteInputs(t))
}

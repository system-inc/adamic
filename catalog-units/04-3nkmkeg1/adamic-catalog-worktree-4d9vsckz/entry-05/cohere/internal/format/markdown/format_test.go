package markdown

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/differential"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/oracletest"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// formatFixtures reach the printer's branches the corpora rarely or never do: our markdown is almost all
// already formatted, and some constructs (escapes inside emphasis, CJK spacing, ignore ranges) barely
// occur in it. Each is formatted by the embedded Prettier and by the port, under every option set below.
var formatFixtures = []string{
	// Emphasis and strong: the style choice and the escapes of delimiter runs inside them.
	"*a*", "_a_", "1*2*3", "a*b*", "*a*b", "**a**", "__a__", "***a***", "1***2***3", "1**_2_**3",
	"*a*b*c*", "_a*b_", "*a_b*", "_a_b_", "**a*b**", "*a**b*", "*a\\*b*", "*\\*a*", "*\\_a*",
	"_*a*_", "*_a_*", "**_a_**", "_**a**_", "*a **b** c*", "*a _b_ c*", "*foo*bar", "foo*bar*",
	"*a* *b*", "*(a)*", "*a.*b", "_a._b", "*\\\\*a*", "**a__b**", "*a__b*", "__a*b__", "*a—b*",
	"*a b*c", "a *b c*d", "**a* b**", "**a *b**", "**a*. b**", "**a .*b**", "*a_ b*", "*a _b*",
	"**a_. b**", "**a ._b**", "**aaa*_**", "_aaa*._", "*aa**_*", "_aa**._", "**aa*_**", "**(a_)b**", "**a(_b)**", "_a*_", "_*a_", "*a**b**c*", "**a\\*b**", "*a*\u00a0b", "*a*\u3000b", "*é*b", "*🍐*b", "a_b_c", "*<https://a.b>*",
	// Inline code, with backticks and padding.
	"`a`", "`` a`b ``", "` `` `", "`  a  `", "` a`", "`a\nb`", "| `a|b` |\n| - |\n| `c` |",
	// Links, images, references and their titles.
	"[a](b)", "[a](b \"c\")", "[a](b 'c')", "[a](b (c))", "[a](b \"c'd\")", "[a](b \"c\\\"d\")",
	"[a](<b c>)", "[a](b)c", "[a](<>)", "![a](b)", "![a](b \"c\")", "![a *b*](c)", "<https://a.b>",
	"<a@b.c>", "<mailto:a@b.c>", "https://a.b", "www.a.b", "[a]\n\n[a]: b", "[a][]\n\n[a]: b",
	"[x][a]\n\n[a]: b", "![a]\n\n[a]: b", "![x][a]\n\n[a]: b", "[A  b]\n\n[a b]: c", "[a]: b \"c\"",
	"[a]: <b c> 'd'", "[a]:\n  b\n  \"c\"", "[a\\]b]: c\n\n[a\\]b]", "[*a* b]\n\n[*a* b]: c",
	// Headings, setext and ATX.
	"# a", "# a #", "a\n=", "a\n---", "a\nb\n===", "# a *b*", "#   a   ", "a\n===\n\nb\n---",
	// Lists: markers, numbering, spacing, nesting, alignment.
	"- a\n- b", "* a\n* b", "+ a\n+ b", "- a\n\n- b", "1. a\n2. b", "1. a\n1. b", "0. a\n0. b",
	"2. a\n3. b", "1) a\n2) b", "- a\n  - b\n    - c", "- a\n\n  b", "-   a\n-   b", "1.  a\n2.  b",
	"- [ ] a\n- [x] b", "- a\n\n\n- b", "- a\n<!-- -->\n- b", "* a\n\n- b", "- a\n- b\n\n* c\n* d",
	"10. a\n11. b", "1. a\n\n   b", "- ```\n  a\n  ```", "- > a", "> - a\n> - b", "- a\n---\n- b",
	"* * *\n- a", "- a\n\n  ***",
	// Block quotes, code, html, thematic breaks.
	"> a\n> b", "> a\nb", "> # a\n> b", ">     a", "    a\n    b", "```\na\n```", "~~~\na\n~~~",
	"````\n```\n````", "```js meta\na\n```", "<div>\na\n</div>", "<!-- a -->", "a <b>c</b> d",
	"***", "___", "- - -", "a\n\n***\n\nb",
	// Breaks, escapes, entities.
	"a  \nb", "a\\\nb", "a\\*b", "&amp; &copy;", "a\\\\", "\\# a", "a\n\\===",
	// Prettier-ignore.
	"<!-- prettier-ignore -->\n| a |  b |\n|-|-|\n|  c | d |",
	"<!-- prettier-ignore-start -->\n*  a\n*  b\n<!-- prettier-ignore-end -->\n\n*  c",
	"> <!-- prettier-ignore -->\n> - a\n>   - b\n>\n> c",
	// Tables: alignment, widths, escapes, CJK width.
	"|a|b|\n|-|-|\n|c|d|", "|a|b|c|\n|:-|:-:|-:|\n|d|e|f|", "| 中文 | b |\n| - | - |\n| c | d |",
	"|a|\n|-|\n|b\\|c|", "a|b\n-|-\nc|d", "|a|b|\n|-|-|\n|c|", "| é | 🍐 |\n|--|--|\n| a | b |",
	// Footnotes, math, liquid, wiki links, strikethrough.
	"a[^1]\n\n[^1]: b", "[^a]: b\n    c", "[^a]:\n    b\n\n    c", "$a$", "$$\na\n$$", "$$ b\na\n$$",
	"{{ a }}", "{% a %}\nb", "[[a]]", "[[a b|c]]", "~a~", "~~a~~", "~~a *b*~~",
	// CJK and Korean spacing, and the line breaks between them.
	"中文\n中文", "中文\nabc", "abc\n中文", "한국어\n한국어", "中文 abc 中文\nabc", "中文。\n中文",
	"a\n한국어", "中文\n：中文", "中文\n(abc)", "中文 abc\n中文", "汉字abc汉字", "中文，abc",
	// Pseudo setext lines and leading characters that would change syntax.
	"a\n    ===", "a\n    ---", "a\n-", "a\n+ b", "a\n# b", "a\n1. b", "a\n> b", "a\n10) b",
	// Front matter.
	"---\na: b\n---\n\nc", "---\n---\n\na", "+++\na = 1\n+++\n\nb",
	// Whitespace at the edges. Byte order marks and carriage returns are native.Formatter's, and
	// tested there for every language.
	"", "\n", "a", "a\n", "\n\na\n\n",
}

var formatVariants = func() []formatoptions.Options {
	defaults := formatoptions.Default()
	narrow := formatoptions.Default()
	narrow.PrintWidth, narrow.TabWidth = 40, 2
	tabs := formatoptions.Default()
	tabs.UseTabs, tabs.SingleQuote = true, false
	return []formatoptions.Options{defaults, narrow, tabs}
}()

// compareFormat formats every input both ways and reports each difference; it returns how many differ.
func compareFormat(t *testing.T, inputs []string) int {
	t.Helper()
	failures := 0
	// The oracle's answers, recorded (see oracletest), so no JavaScript runs here.
	golden := oracletest.Open(t, t.Name())
	for variantIndex, options := range formatVariants {
		engine := golden.Engine(options)
		for _, input := range inputs {
			expected, err := engine.Format("fixture.md", input)
			if err != nil {
				t.Fatalf("the oracle failed on %q: %v", input, err)
			}
			if dependsOnEmbed(input) {
				continue
			}
			actual, err := Format(input, options, nil)
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
	return failures
}

func TestFormatFixturesMatchOracle(t *testing.T) {
	t.Parallel()
	compareFormat(t, formatFixtures)
}

// TestFormatOracleCanFail proves the comparison sees a difference: the port's output for one input
// compared against the oracle's for another must differ.
func TestFormatOracleCanFail(t *testing.T) {
	t.Parallel()
	// The oracle's answers, recorded (see oracletest), so no JavaScript runs here.
	engine := oracletest.Open(t, t.Name()).Engine(formatoptions.Default())
	expected, err := engine.Format("fixture.md", "* a")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := Format("1. a", formatoptions.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(expected, actual) {
		t.Fatal("the comparison found no difference between a bullet list and an ordered one")
	}
}

// TestExhaustiveDelimitersMatchOracle formats every string of up to five of `a`, `*`, `_`, space and `.`,
// bare and inside emphasis and strong: the escapes word.js adds to delimiter runs depend on which side
// of a run is flanking, and only an exhaustive sweep reaches every combination. A measuring run.
func TestExhaustiveDelimitersMatchOracle(t *testing.T) {
	t.Parallel()
	if os.Getenv("COHERE_MARKDOWN_EXHAUSTIVE") == "" {
		t.Skip("set COHERE_MARKDOWN_EXHAUSTIVE to run the sweep")
	}
	alphabet := []string{"a", "*", "_", " ", "."}
	var inputs []string
	var extend func(prefix string, length int)
	extend = func(prefix string, length int) {
		if prefix != "" {
			inputs = append(inputs, prefix, "*"+prefix+"*", "**"+prefix+"**", "_"+prefix+"_", "x "+prefix+" y")
		}
		if length == 5 {
			return
		}
		for _, letter := range alphabet {
			extend(prefix+letter, length+1)
		}
	}
	extend("", 0)

	// The live oracle, not a golden (see oracletest): an opt-in measuring sweep of some 19,500 inputs that
	// no default run reaches, so recording it would check in its answers for nothing a default run reads.
	engine, err := prettier.New(formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}
	failures := 0
	for _, input := range inputs {
		expected, err := engine.Format("fixture.md", input)
		if err != nil {
			continue
		}
		actual, err := Format(input, formatoptions.Default(), nil)
		if err != nil || actual != expected {
			failures++
			if failures <= 20 {
				t.Errorf("%q: %v %s", input, err, differential.FirstDifference(expected, actual))
			}
		}
	}
	t.Logf("%d of %d inputs differ", failures, len(inputs))
}

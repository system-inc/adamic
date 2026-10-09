package micromark

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gfmTableFixtures hold no definition, html text, strikethrough, autolink literal, footnote, task list
// item, math, liquid or wiki-link syntax: those constructs are still pending, so a fixture leaning on them
// would differ for reasons that are not the table's. That rules out `[`, `<`, `~`, `$`, `{`, `@` and
// `www.`/`http` in cells.
var gfmTableFixtures = []string{
	// GFM spec examples (section 4.10, Tables).
	"| foo | bar |\n| --- | --- |\n| baz | bim |\n",
	"| abc | defghi |\n:-: | -----------:\nbar | baz\n",
	"| f\\|oo  |\n| ------ |\n| b `\\|` az |\n| b **\\|** im |\n",
	"| abc | def |\n| --- | --- |\n| bar | baz |\n> bar\n",
	"| abc | def |\n| --- | --- |\n| bar | baz |\nbar\n\nbar\n",
	"| abc | def |\n| --- |\n| bar |\n",
	"| abc | def |\n| --- | --- |\n| bar |\n| bar | baz | boo |\n",
	"| abc | def |\n| --- | --- |\n",

	// Head row: pipes optional at either end, a single cell, a lone pipe, whitespace.
	"a|b\n-|-",
	"a | b\n--|--\nc | d",
	"|a|\n|-|",
	"a|\n-|",
	"|a\n|-",
	"a\n-",
	"|\n-",
	"|\n|-",
	"|\n|",
	"| \n|\n|",
	"||\n||",
	"| |\n|-|",
	"||\n|-|",
	"|a||\n|-|-|",
	"a|b\n-|-|-",
	"a|b|c\n-|-",
	"  a|b\n  -|-",
	"   |a|\n   |-|",
	"    |a|\n|-|",
	"| a \\| b |\n| - |",
	"| a \\\\| b |\n| - | - |",
	"| a\\b |\n| - |",
	"| a\\",
	"| a\\\n| - |",
	"| `a|b` |\n| - |",
	"| `a\\|b` | c |\n| - | - |\n| `d|e` |",
	"| *a* | **b** |\n| - | - |\n| _c_ | d |",
	"| é | 𝔸 |\n| - | - |\n| \u00a0 | 😀 |",
	"|a|\t|b|\n|-|-|-|",

	// Delimiter row: alignments, whitespace, markers out of place.
	"|a|b|c|d|\n|:-|-:|:-:|-|",
	"|a|\n|:|",
	"|a|\n|::|",
	"|a|\n|:-:-|",
	"|a|\n|-:-|",
	"|a|\n| - |",
	"|a|\n|\t-\t|",
	"|a|\n  |-|",
	"|a|\n   |-|",
	"|a|\n    |-|",
	"|a|\n\t|-|",
	"|a|\n|-| ",
	"|a|\n|-|x",
	"|a|\n|x|",
	"|a|\n-",
	"|a|\n---",
	"a|b\n- -",
	"a|b\n-|-\n",
	"|a|\n|-|\n\n|b|",
	"|a|\n|",
	"|a|\n||",
	"|a|b|\n|-||",
	"|a|\n|- |",
	"|a|\n|-:|\n|b|",
	"|a|\n| :--- |",
	"|a|\n",
	"|a|",
	"a|b",
	"a|b\nc",

	// Body rows: missing and extra cells, empty rows, escapes, data runs.
	"|a|b|\n|-|-|\n|c|",
	"|a|b|\n|-|-|\n|c|d|e|f|",
	"|a|\n|-|\n|",
	"|a|\n|-|\n||",
	"|a|\n|-|\n|  |",
	"|a|\n|-|\nb",
	"|a|\n|-|\nb c d|e",
	"|a|\n|-|\n|b\\|c|",
	"|a|\n|-|\n|b\\\\|c|",
	"|a|\n|-|\n|b\\",
	"|a|\n|-|\n\\",
	"|a|\n|-|\n|b|\n|c|\n|d|",
	"|a|\n|-|\n|b|\n\n|c|",
	"|a|\n|-|\n|b|\n# c",
	"|a|\n|-|\n|b|\n***",
	"|a|\n|-|\n|b|\n    c",
	"|a|\n|-|\n|b|\n```\nc\n```",

	// Around it: paragraphs, interrupting, other tables, headings.
	"p\n|a|\n|-|",
	"p\na|b\n-|-",
	"p\n|a|\n|-|\n|b|",
	"p\n\n|a|\n|-|",
	"|a|\n|-|\n|b|\n|c|\n|-|",
	"|a|\n|-|\n\n|b|\n|-|\n|c|",
	"# h\n|a|\n|-|",
	"|a|\n===",
	"|a|\n|-|\n===",
	"a\n|b|\n===",

	// Inside containers, and lazy lines.
	"> |a|\n> |-|\n> |b|",
	"> |a|\n> |-|\n|b|",
	"> |a|\n|-|",
	"> p\n|a|\n|-|",
	"- |a|\n  |-|\n  |b|",
	"- |a|\n  |-|\n|b|",
	"- |a|\n|-|",
	"- a\n- |b|\n  |-|",
	"> - |a|\n>   |-|\n>   |b|",
	"1. |a|b|\n   |:-|-:|\n   |c|d|",

	// Line endings.
	"|a|\r\n|-|\r\n|b|\r\n",
	"|a|\r|-|\r|b|\r",
	"|a|\r\n|-|\n|b|\r",
	"|a|b|\r\n|:-|-:|\r\n\r\n|c|",
}

func TestGfmTableEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, gfmTableFixtures)
}

// gfmTableAlignOracleScript is the event oracle's script, describing only each table's `_align`, which the
// event descriptions leave out and mdast-util-gfm-table reads.
var gfmTableAlignOracleScript = strings.Replace(oracleScript,
	`return events.map(([kind, token]) => describe(kind, token));`,
	`return events.filter(([kind, token]) => kind === "enter" && token.type === "table").map(([, token]) => token._align.join(","));`, 1)

// TestGfmTableAlignMatchesUpstream compares every fixture's table alignments with upstream's `_align`.
func TestGfmTableAlignMatchesUpstream(t *testing.T) {
	t.Parallel()
	if !strings.Contains(gfmTableAlignOracleScript, "_align") {
		t.Fatal("the align oracle script no longer matches the event oracle's")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the event oracle needs it")
	}
	directory := t.TempDir()
	script := filepath.Join(directory, "align.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(gfmTableAlignOracleScript, "FORK", forkRoot(t))), 0o644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(gfmTableFixtures)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", script)
	command.Stdin = strings.NewReader(string(encoded))
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	var expected [][]string
	if err := json.Unmarshal(output, &expected); err != nil {
		t.Fatal(err)
	}
	tables := 0
	for index, input := range gfmTableFixtures {
		var actual []string
		for _, event := range Parse(SourceUnits(input), MarkdownConstructs(), nil) {
			if event.Enter && event.Token.Type == typeTable {
				actual = append(actual, strings.Join(event.Token.Align, ","))
			}
		}
		tables += len(actual)
		if fmt.Sprint(expected[index]) != fmt.Sprint(actual) {
			t.Errorf("input %q: want aligns %q, got %q", truncateInput(input), expected[index], actual)
		}
	}
	if tables == 0 {
		t.Fatal("no fixture produced a table")
	}
	t.Logf("%d tables across %d fixtures", tables, len(gfmTableFixtures))
}

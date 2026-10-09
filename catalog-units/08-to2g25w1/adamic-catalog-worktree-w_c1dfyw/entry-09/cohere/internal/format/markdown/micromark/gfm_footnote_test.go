package micromark

import (
	"strings"
	"testing"
)

// gfmFootnoteFixtures hold no link reference definition, raw html, table, strikethrough, task list item,
// autolink literal, math, liquid or wiki-link syntax (`[[`): those constructs are still pending stubs, so a
// fixture leaning on them would differ for reasons that are not the footnote extension's. Links with
// resources, images, lists, block quotes, code and emphasis are ported and appear beside footnotes. A
// refused definition at the start of a paragraph would fall to the pending link reference definition, so
// refusals sit on a paragraph's second line, where the document still tries containers.
var gfmFootnoteFixtures = []string{
	// GFM's footnote example, and the shapes of micromark-extension-gfm-footnote's own tests that stay
	// inside ported constructs.
	"Here is a simple footnote[^1].\n\nA footnote can also have multiple lines[^2].\n\n[^1]: My reference.\n[^2]: To add line breaks within a footnote, prefix new lines with 2 spaces.\n  This is a second line.",
	"[^a]: b",
	"[^a]: b\n\nc [^a]",
	"c [^a]\n\n[^a]: b",
	"[^a]",
	"a [^b] c",

	// Calls: defined, undefined, before and after their definition, normalized.
	"[^x]: y\n\n[^x] [^y] [^X] [^ x]",
	"[^x]\n\n[^x]: y\n\n[^x]",
	"[^Ü]: y\n\n[^ü] [^Ü]",
	"x\n[^a b]: c\n\n[^a b]",
	"[^a]: x\n[^b]: y\n\n[^a][^b]",
	"[^a]: x\n\n[^a]b[^a]",
	"[^a]: x\n\n*[^a]*",
	"[^a]: x\n\n`[^a]`",

	// `[^]`, `[^`, a bare caret, and labels at the end of input in every state.
	"[^]",
	"x\n[^]: a",
	"[^",
	"[^a",
	"[^a\\",
	"[",
	"a [^",
	"[^a]: x\n\n[^a",
	"[^a]:",
	"[^a]",
	"[^a]: x\n\n[^",
	"[^a]: x\n\n[^]",

	// Spaces, tabs, line endings and brackets inside the label refuse it.
	"[^a]: x\n\n[^a\tb]",
	"[^a]: x\n\n[^a\nb]",
	"[^a]: x\r\n\r\n[^a\r\nb]",
	"[^a]: x\r\rb [^a]\r",
	"[^a]: x\n\n[^a[b]",
	"x\n[^a\tb]: c",
	"x\n[^a\nb]: c",
	"[^a[b]: c",

	// Escapes in labels: of brackets and backslashes, of other characters, and at the end.
	"[^a\\]b]: c\n\n[^a\\]b]",
	"[^a\\[b]: c\n\n[^a\\[b]",
	"[^a\\\\]: c\n\n[^a\\\\]",
	"[^a\\*b]: c\n\n[^a\\*b] [^a*b]",
	"[^\\]]: c\n\n[^\\]]",
	"[^\\]: c",
	"[^a]: x\n\n[^a\\]",
	"[^&amp;]: c\n\n[^&amp;] [^&]",

	// The length limit: 999 is fine, one more is too long, for definitions and calls.
	"[^" + strings.Repeat("a", 999) + "]: b\n\n[^" + strings.Repeat("a", 999) + "]",
	"[^" + strings.Repeat("a", 1000) + "]: b\n\n[^" + strings.Repeat("a", 1000) + "]",
	"[^" + strings.Repeat("a", 998) + "\\]]: b",
	"[^" + strings.Repeat("a", 999) + "\\]]: b",

	// Definition prefix: no colon, no space after, tabs, indentation before.
	"[^a] b",
	"[^a]:b",
	"[^a]:\tb",
	"[^a]:     b",
	"   [^a]: b",
	"    [^a]: b",
	"\t[^a]: b",
	"[^a]: \t b",

	// Continuation: indented, under-indented, lazy, blank lines, code inside.
	"[^a]: b\n    c",
	"[^a]: b\n  c",
	"[^a]: b\n\n    c",
	"[^a]: b\n\n  c",
	"[^a]: b\n\n\tc",
	"[^a]: b\n\n        code",
	"[^a]: b\n\n     c",
	"[^a]: b\n\n\n    c\n\nd",
	"[^a]:\n    b",
	"[^a]:\n\n    b",
	"[^a]: b\r\n    c\r\n\r\n    d",
	"[^a]: b\r    c",
	"[^a]: ```\n    x\n    ```",
	"[^a]: > b\n    > c",
	"[^a]: - b\n    - c",
	"[^a]: b\n[^c]: d",
	"[^a]: b\n    [^c]: d",
	"[^a]: # b\nc",
	"a\n[^b]: c",
	"> [^a]: b\n> c",
	"- [^a]: b\n\n  [^a]",
	"[^a]: [^a]",

	// Adjacent to links and images: the potential call through labelEnd.
	"[^a]: x\n\n![^a]",
	"![^a]",
	"[^a]: x\n\n![^a](b)",
	"[^a]: x\n\n[^a](b)",
	"[^a]: x\n\n[b][^a]",
	"[^a]: x\n\n[b](c)[^a]",
	"[^a]: x\n\n[^a][b](c)",
	"[^a]: x\n\n[b [^a]](c)",
	"[^a]: x\n\n[b [^a] c",
	"[^a]: x\n\n![b [^a]](c)",
	"[^a]: x\n\n![^a b]",
	"[^a]: x\n\n![^ab]",
	"[^a]: x\n\n![^A]",
	"[^a]: x\n\n![^]",
	"[^a]: x\n\n![^a\\]]",
	"[^a]]: x\n\n![^a\\]]",
	"[^a]: x\n\n![x](y) ![^a]",
	"[^a]: x\n\n[x](y) ![^a]",
	"[^a]: x\n\n[x ![^a]",
	"[^a]: x\n\n[^a]]",
	"[^a]: x\n\n[ [^a]]",
	"[^a]: x\n\n[^[^a]]",

	// Non-ASCII and astral characters beside and inside.
	"[^😀]: x\n\né[^😀]𝔸",
	"[^a]: 😀\n\n😀![^a]😀",
}

func TestGfmFootnoteEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, gfmFootnoteFixtures)
}

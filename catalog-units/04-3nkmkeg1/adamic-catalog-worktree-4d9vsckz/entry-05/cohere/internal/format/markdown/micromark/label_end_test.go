package micromark

import (
	"strings"
	"testing"
)

// labelEndFixtures hold no attention, autolink, autolink literal (URLs, www, emails), code, definition,
// html, list, block quote, setext, strikethrough, table, footnote (`[^`), math, liquid or wiki-link
// (`[[`) syntax: those constructs are not ported yet, so a fixture leaning on them would differ for
// reasons that are not labelEnd's. With definitions still a stub nothing is ever defined, so every
// reference here is an undefined one: the shortcut, full and collapsed forms all end in labelEndNok, and
// referenceCollapsed (reached only through a defined label) cannot be exercised yet. CommonMark examples
// with definitions, emphasis, code, html or autolinks are left out for the same reason, and so are 491
// and 494, whose failed resources fall back to html text (`<foo\nbar>`, `<b>`).
var labelEndFixtures = []string{
	// CommonMark examples (Links, Images), the ones that stay inside ported constructs.
	"[link](/uri \"title\")\n",
	"[link](/uri)\n",
	"[](./target.md)\n",
	"[link]()\n",
	"[link](<>)\n",
	"[]()\n",
	"[link](/my uri)\n",
	"[link](</my uri>)\n",
	"[link](foo\nbar)\n",
	"[a](<b)c>)\n",
	"[link](<foo\\>)\n",
	"[link](\\(foo\\))\n",
	"[link](foo(and(bar)))\n",
	"[link](foo(and(bar))\n",
	"[link](foo\\(and\\(bar\\))\n",
	"[link](<foo(and(bar)>)\n",
	"[link](foo\\)\\:)\n",
	"[link](foo\\bar)\n",
	"[link](foo%20b&auml;)\n",
	"[link](\"title\")\n",
	"[link](/url \"title\")\n[link](/url 'title')\n[link](/url (title))\n",
	"[link](/url \"title \\\"&quot;\")\n",
	"[link](/url\u00a0\"title\")\n",
	"[link](/url \"title \"and\" title\")\n",
	"[link](/url 'title \"and\" title')\n",
	"[link](   /uri\n  \"title\"  )\n",
	"[link] (/uri)\n",
	"[link [foo [bar]]](/uri)\n",
	"[link] bar](/uri)\n",
	"[link [bar](/uri)\n",
	"[link \\[bar](/uri)\n",
	"[![moon](moon.jpg)](/uri)\n",
	"[foo [bar](/uri)](/uri)\n",
	"![foo](/url \"title\")\n",
	"![foo ![bar](/url)](/url2)\n",
	"![foo [bar](/url)](/url2)\n",
	"![foo](train.jpg)\n",
	"My ![foo bar](/path/to/train.jpg  \"title\"   )\n",
	"![foo](<url>)\n",
	"![](/url)\n",

	// start: no opening, an inactive opening, a balanced one.
	"]",
	"a]",
	"a]]",
	"[a]]",
	"[a](b)]",
	"[a](b)](c)",
	"[a [b](c) d](e)",
	"[a [b [c](d) e](f) g](h)",
	"[ [a](b)](c)",
	"![a [b](c)](d)",
	"[a ![b](c)](d)",
	"![a ![b](c)](d)",
	"[a]b](c)",
	"[a] [b](c)",
	"[a](b) [c](d)",
	"![a](b)![c](d)",

	// after: shortcut, full and collapsed references, all undefined.
	"[a]",
	"![a]",
	"[a][b]",
	"[a][]",
	"![a][]",
	"[a][ ]",
	"[a][\n]",
	"[a][b",
	"[a][",
	"[a][b]c",
	"[a][b][c]",
	"[a][b](c)",
	"[a][b\nc]",
	"[a][b\\]c]",
	"[a][" + strings.Repeat("x", 998) + "]",
	"[a][" + strings.Repeat("x", 999) + "]",
	"[a][" + strings.Repeat("x", 1000) + "]",

	// The resource: each state, and the end of input in each.
	"[a](",
	"[a](b",
	"[a]( ",
	"[a](b ",
	"[a](b \"c",
	"[a](b \"c\"",
	"[a](b \"c\" ",
	"[a](b \"c\" d)",
	"[a](b\"c\")",
	"[a](b )",
	"[a]( b )",
	"[a](\tb\t'c'\t)",
	"[a](\nb\n)",
	"[a](b\r\n\"c\"\r\n)",
	"[a](b\r'c'\r)",
	"[a](\r\n)",
	"[a](b (c))",
	"[a](b (c)",
	"[a](b (c) )",
	"[a](())",
	"[a](b \"c\nd\")",
	"[a](b \"c\" \"d\")",
	"[a](b \"c\"\n\n)",
	"[a](\n\n)",
	"[a](<b\n)",
	"[a](<<) b",
	"[a](b) c",
	"[a](" + strings.Repeat("(", 32) + strings.Repeat(")", 32) + ")",
	"[a](" + strings.Repeat("(", 33) + strings.Repeat(")", 33) + ")",

	// Labels: line endings, escapes, references, breaks, non-ASCII and astral characters.
	"[a\nb](c)",
	"[a\r\nb](c)",
	"[a\rb](c)",
	"[a\\]](b)",
	"\\[a](b)",
	"[a](b\\))",
	"[a&amp;](b&amp;)",
	"[a\\\nb](c)",
	"[a  \nb](c)",
	"[a\t\nb](c)",
	"[é](ü)",
	"[𝔸](𝔹 \"𝔸\")",
	"𝔸[a](b)𝔸",
	"![𝔸](b)",
	"[a](b)\u00a0c",

	// Around it: other flow.
	"# [a](b)",
	"# [a](b) #",
	"    [a](b)",
	"a\n[b](c)\nd",
	"[a](b)\n\n[c](d)",
	"***\n[a](b)",
	"[a\n\nb](c)",
}

func TestLabelEndEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, labelEndFixtures)
}

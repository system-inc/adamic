package micromark

import "testing"

// codeTextFixtures hold no attention, label end, autolink, html, character reference, strikethrough,
// math, liquid, wiki-link, fenced code (a bare "```" line), GFM autolink literal (`www.`, `http`, `@`)
// or container syntax: those constructs are still pending, so a fixture leaning on them would differ for
// reasons that are not codeText's. That leaves out CommonMark examples 341, 342, 344 and 346.
var codeTextFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"`foo`\n",
	"`` foo ` bar ``\n",
	"` `` `\n",
	"`  ``  `\n",
	"` a`\n",
	"` b `\n",
	"` `\n`  `\n",
	"``\nfoo\nbar  \nbaz\n``\n",
	"``\nfoo \n``\n",
	"`foo   bar \nbaz`\n",
	"`foo\\`bar`\n",
	"``foo`bar``\n",
	"` foo `` bar `\n",
	"`<a href=\"`\">`\n",
	"`<https://foo.bar.`baz>`\n",
	"```foo``\n",
	"`foo\n",
	"`foo``bar``\n",

	// start and sequenceOpen: one, two, many, and end of input inside the opening.
	"`",
	"``",
	"a```",
	"`a`",
	"``a``",
	"``````````a``````````",
	"``````````a`````````",
	"`a``",
	"``a`",
	"`` `a` ``",

	// between: end of input after spaces and line endings, spaces, line endings of each kind.
	"` ",
	"`\n",
	"` a",
	"`  `",
	"`   a   `",
	"`a b`",
	"`a  b`",
	"`a\nb`",
	"`a\r\nb`",
	"`a\rb`",
	"`\na\n`",
	"`\r\na\r\n`",
	"`\ra\r`",
	"` \na\n `",
	"`\n`",
	"a`\n`b",
	"`a\n`",
	"`\na`",
	"` a\n`",

	// data: tabs, non-breaking spaces, escapes, non-ASCII and astral characters inside.
	"`\ta\t`",
	"`a\tb`",
	"` \t `",
	"`\u00a0a\u00a0`",
	"`é`",
	"`𝔸`",
	"`𝔸 𝔸`",
	"`a\\`",
	"`a\\\nb`",
	"`\\\\`",

	// sequenceClose: shorter and longer runs become data, the matching one closes.
	"``a`b``",
	"``a```b``",
	"`a``b`",
	"``a``b``",
	"` `` ` ``",

	// previous: a backtick after a backtick, after an escape, after text.
	"\\`a`",
	"\\``a`",
	"\\``a``",
	"a\\``b`",
	"\\```a``",
	"a`b`c",
	"a``b`",
	"`a``b",
	"`a` `b`",
	"`a``b``c`",
	"é`a`é",
	"𝔸`a`𝔸",

	// Around it: headings, paragraphs, trailing whitespace, indented code.
	"# `a` #",
	"# `a #`",
	"## ` a `",
	"`a`  \nb",
	"`a`\\\nb",
	"a\n`b`",
	"a\n  `b\n  c`",
	"`a\n\nb`",
	"    `a`",
	"***\n`a`",
	"`a`\n***",
}

func TestCodeTextEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, codeTextFixtures)
}

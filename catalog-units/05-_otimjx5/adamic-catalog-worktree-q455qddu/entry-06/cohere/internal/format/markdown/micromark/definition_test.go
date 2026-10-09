package micromark

import (
	"strings"
	"testing"
)

// definitionFixtures are definitions on their own, with no later reference, `[[` or `[^`: written while
// labelEnd, wiki-link and gfm-footnote were stubs, so nothing here leans on them. References to defined
// labels are definitionReferenceFixtures below.
var definitionFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs, with their trailing references
	// dropped.
	"[foo]: /url \"title\"\n",
	"   [foo]: \n      /url  \n           'the title'  \n",
	"[Foo*bar\\]]:my_(url) 'title (with parens)'\n",
	"[Foo bar]:\n<my url>\n'title'\n",
	"[foo]: /url '\ntitle\nline1\nline2\n'\n",
	"[foo]:\n/url\n",
	"[foo]: <>\n",
	"[foo]: /url\\bar\\*baz \"foo\\\"bar\\baz\"\n",
	"[foo]: first\n[foo]: second\n",
	"[FOO]: /url\n",
	"[ΑΓΩ]: /φου\n",
	"[foo]: /url\n",
	"[\nfoo\n]: /url\nbar\n",
	"[foo]: /url\n\"title\" ok\n",
	"    [foo]: /url \"title\"\n",
	"```\n[foo]: /url\n```\n",
	"# Foo\n[foo]: /url\n> bar\n",
	"[foo]: /url\nbar\n===\n",
	"[foo]: /url\n===\n",
	"[foo]: /foo-url \"foo\"\n[bar]: /bar-url\n  \"bar\"\n[baz]: /baz-url\n",
	"> [foo]: /url\n",
	// The label: escapes, line endings inside, tabs, references, non-ASCII and astral characters.
	"[a\\]]: b",
	"[a\\\\]: b",
	"[a\\b]: c",
	"[a\nb]: c",
	"[a\r\nb]: c",
	"[a\rb]: c",
	"[\ta]: b",
	"[a\t]: b",
	"[ a ]: b",
	"[a&amp;]: b",
	"[é]: b",
	"[𝔸]: b",
	"[a𝔸]: 𝔸",

	// The label length limit: 999 characters is the most a label holds.
	"[" + strings.Repeat("a", 998) + "]: b",
	"[" + strings.Repeat("a", 999) + "]: b",

	// The marker and the whitespace after it.
	"[a]:b",
	"[a]:\tb",
	"[a]: \t b",
	"[a]:\n b",
	"[a]:\r\nb",
	"[a]:\rb",
	"[a]:\n\tb",

	// The destination: enclosed, raw, escapes, balance, references.
	"[a]: <b c>",
	"[a]: <b\\>c>",
	"[a]: <b\\<c>",
	"[a]: <b\\\\>",
	"[a]: <b\\c>",
	"[a]: b(c)",
	"[a]: b(c(d))e",
	"[a]: b\\(c",
	"[a]: b\\)c",
	"[a]: b\\\\",
	"[a]: b&amp;c",
	"[a]: " + strings.Repeat("(", 40) + "b" + strings.Repeat(")", 40),
	"[a]: /é𝔸",
	"[a]: b\u00a0c",

	// After the destination: whitespace, line endings, end of input.
	"[a]: b",
	"[a]: b ",
	"[a]: b\t",
	"[a]: b  \n",
	"[a]: b\r\n",
	"[a]: b\r",
	"[a]: <b>",
	"[a]: <b> \n",

	// The title: every marker, empty, escapes, on further lines, trailing whitespace.
	"[a]: b \"c\"",
	"[a]: b 'c'",
	"[a]: b (c)",
	"[a]: b \"\"",
	"[a]: b ''",
	"[a]: b ()",
	"[a]: b \"c\\\"d\"",
	"[a]: b 'c\\'d'",
	"[a]: b (c\\)d)",
	"[a]: b (c\\\\)",
	"[a]: b \"c\\d\"",
	"[a]: b \"c&quot;\"",
	"[a]: b \"c\" ",
	"[a]: b \"c\"\t\n",
	"[a]: b\t\"c\"",
	"[a]: b\n\"c\"",
	"[a]: b\r\n'c'",
	"[a]: b\r(c)",
	"[a]: b\n  \"c\"  \nd",
	"[a]: b \"c\nd\"",
	"[a]: b \"c\r\n  d\"",
	"[a]: b \"c\rd\"",
	"[a]: <b>\n\"c\"",
	"[a]: b \"é𝔸\"",

	// A title that fails, leaving the definition without one.
	"[a]: b\n\"c\" d",
	"[a]: b\n'c",
	"[a]: b\n(c",
	"[a]: b\n\"c\n\nd\"",

	// Around it: indentation, more definitions, paragraphs, other flow and containers.
	"   [a]: b",
	" \t[a]: b",
	"[a]: b\n[c]: d",
	"[a]: b\n  [c]: d",
	"[a]: b\nc",
	"[a]: b \"c\"\nd",
	"[a]: b\n\n[c]: d",
	"[a]: b\n# c",
	"[a]: b\n***",
	"[a]: b\n```\nc\n```",
	"[a]: b\n    c",
	"[a]: b\n[c]: d\n===",
	"[a]: b\nc\n---",
	"# a\n[b]: c",
	"***\n[a]: b",
	"> [a]: b\n> c",
	"> [a]:\nb",
	"> [a]: b\n> \"c\"",
	"> [a]: b\n\"c\"",
	"> > [a]: b",
}

// definitionFailureFixtures are the paths where definition fails: each leaves a `[` in a paragraph, which
// labelEnd's resolveAll turns into data.
var definitionFailureFixtures = []string{
	"[foo]: /url 'title\n\nwith blank line'\n",
	"[foo]:\n",
	"[foo]: /url \"title\" ok\n",
	"Foo\n[bar]: /baz\n",
	"[]: a",
	"[ ]: a",
	"[a] : b",
	"[a[]: b",
	"[a]: a(b",
	"[a]: a)b",
	"[a]: b \"c\" d",
	"[a]: b 'c",
	"[a]: b\x01c",
	"[a]: b\x7fc",
	"[a]:\n\nb",
	"[a]:\r\n\r\nb",
	"[" + strings.Repeat("a", 1000) + "]: b",
	"[a",
	"[a]",
	"[a]:",
	"[a]: ",
	"[a]: <",
	"[a]: <b",
	"[a]: b(",
	"[a]: b \"",
	"[a]: b \"c",
	"[a]: b \"c\" x",
	"[a]: b \"c\"d",
	"[a]: b c",
	"a\n[b]: c",
	"[a]: b\n[c]: d e",
	"> [a]:\n\nb",
}

// definitionReferenceFixtures are references to defined labels, which only the identifiers definition
// records let labelEnd resolve: shortcut, full and collapsed, before and after the definition, normalized.
var definitionReferenceFixtures = []string{
	"[foo]: /url\n\n[foo]",
	"[foo]\n\n[foo]: /url",
	"[a]: b\n\n[a][]",
	"[a]: b\n\n[x][a]",
	"[a]: b\n\n![a]",
	"[a]: b\n\n![a][]",
	"[a]: b\n\n![x][a]",
	"[Foo  Bar]: b\n\n[foo bar]",
	"[ÄÖ]: b\n\n[äö]",
	"[a]: b\n\n[a] [a]",
	"[a]: b\n\n[c][]",
	"> [a]: b\n\n[a]",
}

func TestDefinitionEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, definitionFixtures)
}

func TestDefinitionFailureEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, definitionFailureFixtures)
}

func TestDefinitionReferenceEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, definitionReferenceFixtures)
}

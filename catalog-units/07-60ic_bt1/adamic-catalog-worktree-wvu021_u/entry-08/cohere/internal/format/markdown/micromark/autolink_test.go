package micromark

import (
	"strings"
	"testing"
)

// autolinkFixtures avoid the constructs still pending, outside a matched autolink (inside one, the autolink
// consumes the characters and nothing else is tried): attention (`*`, `_`), code text, character
// references, label ends, math, liquid, wiki links, strikethrough, html (text and flow: no `<` that opens
// a valid tag, comment, instruction, declaration or closing tag, and no `<` line that htmlFlow would take),
// and the GFM autolink literals: a `<` that fails to be an autolink never leaves behind `http://`,
// `https://`, `www.` or an `a@b.c` email that the literal extension would turn into a link upstream.
// That last rule is why a second domain label one past the 63 limit is not here (upstream links it as a
// literal email); the size reset at a dot is held by the 63.63 case succeeding instead. CommonMark examples
// 602 and 608 and the bare 611 and 612 are left out for the same reason.
var autolinkFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"<http://foo.bar.baz>\n",
	"<https://foo.bar.baz/test?q=hello&id=22&boolean>\n",
	"<irc://foo.bar:2233/baz>\n",
	"<MAILTO:FOO@BAR.BAZ>\n",
	"<a+b+c:d>\n",
	"<made-up-scheme://foo,bar>\n",
	"<https://../>\n",
	"<localhost:5001/foo>\n",
	"<https://example.com/\\[\\>\n",
	"<foo@bar.example.com>\n",
	"<foo+special@Bar.baz-bar0.com>\n",
	"<foo\\+@bar.example.com>\n",
	"<>\n",
	"<m:abc>\n",
	"<foo.bar.baz>\n",

	// start and open: what may follow `<`.
	"a <ab:c> b",
	"a<ab:c>b",
	"a <@b.c> d",
	"a <1@b> c",
	"a <+@b> c",
	"a <.@b> c",
	"a <#$%&'*+-/=?^_`{|}~@b> c",
	"a <,b> c",
	"a < b> c",
	"a <\nb> c",
	"a <",
	"a <\n",

	// schemeOrEmailAtext and schemeInsideOrEmailAtext: the scheme, its limit, and the fall back to atext.
	"a <ab:> c",
	"a <a+:b> c",
	"a <a-:b> c",
	"a <a.:b> c",
	"a <a9:b> c",
	"a <a:b> c",
	"a <a@b> c",
	"a <ab@c> c",
	"a <a.b@c> d",
	"a <ab+c@d> e",
	"a <a,b@c> d",
	"a <ab,c:d> e",
	"a <a" + strings.Repeat("b", 30) + ":c> d",
	"a <a" + strings.Repeat("b", 31) + ":c> d",
	"a <a" + strings.Repeat("b", 32) + ":c> d",
	"a <a" + strings.Repeat("b", 31) + "@c> d",
	"a <a" + strings.Repeat("b", 32) + "@c> d",
	"a <a" + strings.Repeat("b", 40) + "@c> d",
	"a <a",
	"a <ab",
	"a <ab:",
	"a <ab:c",
	"a <ab\n",

	// urlInside: what ends it, what refuses it.
	"a <ab:c<1> e",
	"a <ab:c d> e",
	"a <ab:c\td> e",
	"a <ab:c\nd> e",
	"a <ab:c\r\nd> e",
	"a <ab:c\rd> e",
	"a <ab:c\u007fd> e",
	"a <ab:c\u0001d> e",
	"a <ab:>> c",
	"a <ab:\"'()[]> c",
	"a <ab:*_`$~{}&!> c",
	"a <ab:é> c",
	"a <ab:𝔸> c",
	"a <ab:\u00a0> c",
	"a <ab:\uFEFF> c",
	"a <ab::> c",
	"a <ab:@> c",

	// emailAtext, emailAtSignOrDot, emailLabel, emailValue.
	"a <b@c.d> e",
	"a <b@c.d.e> f",
	"a <b@c-d> e",
	"a <b@c--d> e",
	"a <b@c-> d",
	"a <b@-c> d",
	"a <b@.,> d",
	"a <b@c.> d",
	"a <b@@c> d",
	"a <b@c@d> e",
	"a <b@c,d> e",
	"a <b@> c",
	"a <b@",
	"a <b@c",
	"a <b@c-",
	"a <b@c.",
	"a <é@c> d",
	"a <b@é> c",
	"a <b@c\nd> e",
	"a <b c@d> e",
	"a <b@" + strings.Repeat("c", 62) + "> d",
	"a <b@" + strings.Repeat("c", 63) + "> d",
	"a <b@" + strings.Repeat("c", 64) + "> d",
	"a <b@" + strings.Repeat("c-", 31) + "c> d",
	"a <b@" + strings.Repeat("c-", 32) + "> d",
	"a <b@" + strings.Repeat("c", 63) + "." + strings.Repeat("d", 63) + "> e",

	// Around it: line endings, non-ASCII and astral neighbours, escapes, headings, several in a row.
	"a\n<ab:c>\nd",
	"a\r\n<ab:c>\r\nd",
	"a\r<b@c>\rd",
	"<ab:c>",
	"<b@c>",
	"<ab:c>\n",
	"é<ab:c>é",
	"𝔸<b@c>𝔸",
	"a\t<ab:c>\t",
	"\\<ab:c>",
	"a\\<b@c>",
	"# <ab:c>",
	"# a <b@c> #",
	"<ab:c><d@e>",
	"<ab:c<d@e>",
	"a <ab:c> <d@e.f> <g:h>",
	"a  \n<ab:c>",
	"a\\\n<ab:c>",
	"    <ab:c>",
	"***\n<ab:c>",
}

func TestAutolinkEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, autolinkFixtures)
}

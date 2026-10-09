package micromark

import (
	"strings"
	"testing"
)

// characterReferenceFixtures hold no attention, autolink, code (fenced or text), definition, html,
// label end, list, setext, strikethrough, math, liquid or wiki-link syntax, and no URLs or emails:
// those constructs are not ported yet, so a fixture leaning on them would differ for reasons that are
// not characterReference's. CommonMark examples 31 to 35, 37, 38 and 41 are left out for that reason.
var characterReferenceFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"&nbsp; &amp; &copy; &AElig; &Dcaron;\n&frac34; &HilbertSpace; &DifferentialD;\n&ClockwiseContourIntegral; &ngE;\n",
	"&#35; &#1234; &#992; &#0;\n",
	"&#X22; &#XD06; &#xcab;\n",
	"&nbsp &x; &#; &#x;\n&#87654321;\n&#abcdef0;\n&ThisIsNotDefined; &hi?;\n",
	"&copy\n",
	"&MadeUpEntity;\n",
	"    f&ouml;f&ouml;\n",
	"foo&#10;&#10;bar\n",
	"&#9;foo\n",

	// start and open: end of input and what may follow `&`.
	"&",
	"&\n",
	"&\r\n",
	"&\r",
	"& amp;",
	"&\tamp;",
	"&;",
	"&&amp;",
	"&é;",
	"&𝔸;",

	// Named values: known, unknown, case, digits, the 31 limit and one past.
	"&amp;",
	"&AMP;",
	"&Amp;",
	"&amp",
	"&am",
	"&ampx;",
	"&amp;;",
	"&amp;&amp;",
	"&frac34;",
	"&sup2;",
	"&1;",
	"&a1b;",
	"&CounterClockwiseContourIntegral;",
	"&CounterClockwiseContourIntegralX;",
	"&" + strings.Repeat("a", 31) + ";",
	"&" + strings.Repeat("a", 32) + ";",
	"&" + strings.Repeat("a", 33),
	"&amp\n;",
	"&amp\r\n;",
	"&amp\r;",
	"&amp ;",
	"&amp\t;",
	"&amp-;",

	// numeric: `#`, then `x`, `X`, digit, or anything else.
	"&#",
	"&#\n",
	"&#;",
	"&#a;",
	"&# 1;",
	"&#1",
	"&#12",
	"&#12\n;",
	"&#1;",
	"&#0000000;",
	"&#9999999;",
	"&#99999999;",
	"&#1234567",
	"&#12345678",
	"&#1a;",

	// Hexadecimals: the 6 limit and one past, case, non-hex digits.
	"&#x",
	"&#X",
	"&#x;",
	"&#X;",
	"&#x\n",
	"&#x1",
	"&#x1;",
	"&#X1;",
	"&#xAbCdEf;",
	"&#x10FFFF;",
	"&#x110000;",
	"&#x0000041;",
	"&#xg;",
	"&#x1g;",
	"&#xx1;",

	// Around it: other text, escapes, breaks, non-ASCII, astral, flow.
	"a&amp;b",
	"é&amp;é",
	"𝔸&amp;𝔸",
	"&amp;é",
	"\\&amp;",
	"&\\amp;",
	"&amp;\\",
	"&amp;\\\nb",
	"&amp;  \nb",
	"a\n&amp;",
	"a\r\n&#35;\r\nb",
	"  &amp;",
	"\t&amp;",
	"# &amp; #",
	"# &#35;",
	"&#35; a",
	"    &amp;",
	"***\n&amp;",
	"&amp;\n\n&copy;",
}

func TestCharacterReferenceEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, characterReferenceFixtures)
}

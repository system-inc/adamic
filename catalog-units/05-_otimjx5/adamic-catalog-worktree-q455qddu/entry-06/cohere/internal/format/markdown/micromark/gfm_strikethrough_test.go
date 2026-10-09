package micromark

import "testing"

// gfmStrikethroughFixtures hold no definition, footnote (no `[^`), html text (no `<`), table (no `|`),
// task list item, math (no `$`), liquid (no `{`) or wiki-link (no `[[`) syntax, and no URLs, `www.` or
// emails: those constructs are still pending stubs, so a fixture leaning on them would differ for reasons
// that are not strikethrough's. A `~~~` at the start of a line is a code fence, which is ported, so those
// stay in. Unicode neighbours are long-assigned code points, so their classification is the same in the
// bundle's table and in Node's.
var gfmStrikethroughFixtures = []string{
	// GFM spec examples 491 to 493, and the readme's.
	"~~Hi~~ Hello, ~there~ world!",
	"This ~~has a\n\nnew paragraph~~.",
	"This will ~~~not~~~ strike.",
	"Some ~strikethrough~.",
	"~~Hi~~ Hello, world!",

	// One, two, three markers, and one past.
	"~a~",
	"~~a~~",
	"~~~a~~~",
	"~~~~a~~~~",
	"~",
	"~~",
	"~~~",
	"a ~",
	"a ~~",
	"a ~~~",
	"a ~~~~",
	"a~",
	"~a",
	"~~a",
	"a~~",
	"~ ~",
	"~~ ~~",

	// Mismatched runs: sizes must be equal.
	"~a~~",
	"~~a~",
	"~~a~ b~~",
	"~a~~ b~",
	"~~a~~~",
	"~~~a~~",
	"~a ~~b~~ c~",
	"~~a ~b~ c~~",
	"~a~b~c~",
	"~~a~~b~~c~~",
	"~a ~b~",
	"~a~ b~",

	// Flanking: whitespace, punctuation and letters on either side.
	"~ a~",
	"~a ~",
	"a~b~c",
	"a ~b~ c",
	"a~ b ~c",
	"(~a~)",
	"~(a)~",
	"a~(b)~c",
	"(~~a~~)b",
	"a~~.b~~",
	".~~a~~.",
	"~.~",
	"~~.~~",
	"a~.~b",
	"\"~a~\"",
	"~a~\t",
	"\t~a~",
	"~a~\u00a0b",
	"\u00a0~a~",
	"é~a~é",
	"𝔸~a~𝔸",
	"~𝔸~",
	"~~é~~",
	"「~a~」",
	"\u3000~a\u3000~",

	// Escapes: an escaped tilde before a sequence does not stop it.
	"\\~a~",
	"~a\\~",
	"\\~~a~",
	"\\~~~a~~",
	"~~a\\~~",
	"a\\~~b~",

	// Line endings around and inside.
	"~a\nb~",
	"~a\r\nb~",
	"~a\rb~",
	"~~a\n~~",
	"~~\na~~",
	"a\n~b~\nc",
	"~a~\n",
	"~a~\r\n",
	"~a~\r",
	"a ~~b  \nc~~",
	"a ~~b\\\nc~~",

	// Nesting with emphasis and with itself.
	"*~a~*",
	"~*a*~",
	"**~~a~~**",
	"~~**a**~~",
	"*~a*~",
	"~*a~*",
	"_~a~_",
	"~_a_~",
	"~~a *b~~ c*",
	"*a ~~b* c~~",
	"~~a ~b~ c~~",
	"~a ~~b~~ c~",
	"~~*a* ~b~ **c**~~",
	"~a*~*",
	"*~*a",
	"~~a~~*b*",

	// Inside links, and links inside.
	"[~a~](b)",
	"[~~a~~](b)",
	"~[a](b)~",
	"~~[a](b)~~",
	"[~a](b)~",
	"~[a~](b)",
	"![~a~](b)",
	"[a ~b c](d) e~",
	"[~a~]",

	// Code and other text constructs beside it.
	"~`a`~",
	"`~a~`",
	"~a`~`",
	"~&amp;~",
	"&#126;a~",

	// In flow containers and headings.
	"# ~a~",
	"> ~a~",
	"- ~a~",
	"> ~a\n> b~",
	"- ~a\n  b~",
	"~a\n===",
	"~~~\na\n~~~",
	"a\n~~~b~~~",
}

func TestGfmStrikethroughEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, gfmStrikethroughFixtures)
}

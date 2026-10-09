package micromark

import "testing"

// wikiLinkFixtures are inputs where every `[` that reaches the text tokenizer opens a wiki link that
// closes, so labelStartLink and labelEnd never keep a token. They hold no list, emphasis, strikethrough,
// math, liquid, html, table, footnote or autolink-literal syntax, and no stray `[`: where those
// constructs are still pending a fixture leaning on them would differ for reasons that are not
// wikiLink's. The refusals, which leave their `[` to labelStartLink, are wikiLinkRefusalFixtures below.
var wikiLinkFixtures = []string{
	// The plain link, and the alias divider the fork disables: `|` is an ordinary target character.
	"[[a]]",
	"[[Some Page]]",
	"[[a|b]]",
	"[[|]]",
	"[[a|]]",
	"[[|b]]",
	"[[ | ]]",
	"[[a||b]]",

	// The target: what counts as data, spaces and tabs around it, and brackets inside it.
	"[[ a]]",
	"[[a ]]",
	"[[ a ]]",
	"[[\ta]]",
	"[[a\tb]]",
	"[[a\t]]",
	"[[ \t a \t ]]",
	"[[[]]",
	"[[[a]]",
	"[[a[b]]",
	"[[[[a]]]]",
	"[[a [[b]] c]]",
	"[[a]]]",
	"[[a]]]]",
	"[[^a]]",
	"[[a\\]]",
	"[[\\]]",
	"[[&amp;]]",
	"[[`a`]]",
	"[[*a*]]",
	"[[<a>]]",
	"[[\x00]]",

	// Non-ASCII and astral characters inside and beside it.
	"[[é]]",
	"[[日本語]]",
	"[[𝔸]]",
	"[[a𝔸b]]",
	"é[[a]]é",
	"𝔸[[a]]𝔸",
	"[[a]]\u00a0",
	"\ufeff[[a]]",

	// Around it in text.
	"a[[b]]c",
	"a [[b]] c",
	"[[a]][[b]]",
	"[[a]] [[b]]",
	"\\\\[[a]]",
	"&amp;[[a]]",
	"`[[a]]`",
	"`a`[[b]]",
	"a]][[b]]",
	"]][[a]]",
	"[[a]]  ",
	"[[a]]\t",
	"a\t[[b]]",
	"[[a]]  \nb",
	"[[a]]\\\nb",
	"a  \n[[b]]",

	// Line endings after it, and links on following lines.
	"[[a]]\n",
	"[[a]]\r\n",
	"[[a]]\r",
	"[[a]]\nb",
	"[[a]]\r\nb",
	"[[a]]\rb",
	"[[a]]\n[[b]]",
	"[[a]]\r\n[[b]]",
	"[[a]]\r[[b]]",
	"[[a]]\n\n[[b]]",
	"a\n[[b]]\nc",

	// Flow around it: headings, block quotes, indented and fenced code (no text, so no link).
	"# [[a]]",
	"# [[a]] #",
	"## [[a|b]] ##",
	"[[a]]\n===",
	"[[a]]\n---",
	"> [[a]]",
	"> a\n> [[b]]",
	"> [[a]]\nb",
	">\t[[a]]",
	"   [[a]]",
	"    [[a]]",
	"\t[[a]]",
	"```\n[[a]]\n```",
	"```[[a]]\n```",
	"***\n[[a]]",

	// No length limit.
	"[[" + longTarget(1000) + "]]",
}

// wikiLinkRefusalFixtures are the inputs where the attempt fails and the `[` falls to labelStartLink,
// whose leftover tokens labelEnd's resolveAll turns into data.
var wikiLinkRefusalFixtures = []string{
	// End of input in each state.
	"[",
	"[[",
	"[[a",
	"[[ ",
	"[[a]",
	"[[a] ",

	// No data: empty, spaces and tabs only.
	"[[]]",
	"[[ ]]",
	"[[\t]]",
	"[[ \t ]]",
	"[[]]]",

	// The start marker: one bracket only.
	"[a]",
	"[a]]",
	"[ [a]]",
	"a[b]]c",

	// The end marker: one bracket, then something else.
	"[[a]b]]",
	"[[a](b)]]",
	"[[a] ]]",
	"[[a]\n]]",

	// Line endings in consumeData and consumeTarget.
	"[[\na]]",
	"[[\r\na]]",
	"[[\ra]]",
	"[[a\nb]]",
	"[[a\r\nb]]",
	"[[a\rb]]",
	"> [[a\n> b]]",

	// A refusal beside a link, and the links inside other labels.
	"[[]] [[a]]",
	"[[a\n[[b]]",
	"\\[[a]]",
	"![[a]]",
	"[x [[a]]](b)",
	"[[a]](b)",
}

func longTarget(size int) string {
	target := make([]byte, size)
	for index := range target {
		target[index] = 'a'
	}
	return string(target)
}

func TestWikiLinkEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, wikiLinkFixtures)
}

func TestWikiLinkRefusalEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, wikiLinkRefusalFixtures)
}

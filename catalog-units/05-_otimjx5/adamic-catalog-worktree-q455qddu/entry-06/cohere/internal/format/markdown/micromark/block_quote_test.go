package micromark

import "testing"

// blockQuoteFixtures hold no list, fenced code, html, setext, definition, table, emphasis, code text,
// link, autolink, math, liquid or wiki-link syntax: those constructs are not ported yet, so a fixture
// leaning on them would differ for reasons that are not blockQuote's. CommonMark examples 235 (a list) and
// 237 (fenced code) are left out for that reason. Thematic breaks use `***` and `___`, never `---`, so a
// setext underline is never in play.
var blockQuoteFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	">\t\tfoo\n",
	"> # Foo\n> bar\n> baz\n",
	"># Foo\n>bar\n> baz\n",
	"   > # Foo\n   > bar\n > baz\n",
	"    > # Foo\n    > bar\n    > baz\n",
	"> # Foo\n> bar\nbaz\n",
	"> bar\nbaz\n> foo\n",
	">     foo\n    bar\n",
	"> foo\n    - bar\n",
	">\n",
	">\n>  \n> \n",
	">\n> foo\n>  \n",
	"> foo\n\n> bar\n",
	"> foo\n> bar\n",
	"> foo\n>\n> bar\n",
	"foo\n> bar\n",
	"> aaa\n***\n> bbb\n",
	"> bar\nbaz\n",
	"> bar\n\nbaz\n",
	"> bar\n>\nbaz\n",
	"> > > foo\nbar\n",
	">>> foo\n> bar\n>>baz\n",
	">     code\n\n>    not code\n",

	// Start: what follows the marker, end of input right after it.
	">",
	"> ",
	">  ",
	">\t",
	"> \t",
	">a",
	"> a",
	">  a",
	">\ta",
	"> \ta",
	">\t a",
	">é",
	"> 𝔸",
	">\u00a0a",
	"> a ",
	"> a\\",

	// Line endings right after the marker and after the whitespace.
	">\n",
	">\r\n",
	">\r",
	"> \n",
	"> \r\n",
	">\t\r",
	"> a\r\n> b",
	"> a\rb",
	"> a\r\n\r\nb",
	"> a\r> b\r",

	// Continuation: indentation before the marker, up to the limit and past it.
	"> a\n > b",
	"> a\n  > b",
	"> a\n   > b",
	"> a\n    > b",
	"> a\n\t> b",
	"> a\n \t> b",
	"> a\n   \t> b",
	">\n    > b",
	"> ***\n    > b",
	"> a\n>",
	"> a\n> ",
	"> a\n ",
	"> a\n",
	"> a\n\n",

	// Indentation before the first marker.
	" > a",
	"  > a",
	"   > a",
	"\t> a",
	"   >\ta",
	" \t> a",

	// Nesting.
	"> > a\n> b",
	">>a\n>>b",
	"> > a\n>\n> > b",
	">>>>>>>>>>a",
	"> > a\n> > b\n> c",
	"> > a\nb",
	"> > a\n> b\nc",
	"> > a\n\n> b",
	"> > # a\n> b",
	">\n>>\n>>>",
	"> >\ta\n>  > b",

	// Lazy continuation and what ends it.
	"> a\nb\nc",
	"> a\n    b",
	"> a\n# b",
	"> a\n___",
	"> a\n\tb",
	">     code\nb",
	"> # a\nb",
	"> ***\nb",
	"> a\n\\> b",
	"> foo\n    \\- bar",
	"> a\n&gt; b",
	"> > a\nb\n> c",
	"> a\nb\n\n> c",

	// Quotes around ported flow.
	"> # a\n> ## b",
	"> ***\n> ___",
	"> a\n> ***\n> # b",
	">\tcode\n>\tmore",
	">     a\n>     b\n> c",
	">     a\n>\n>     b",
	"# a\n> b\n# c",
	"***\n> a\n***",
	"    code\n> a",
	"a\n    > b",
	"> a\n\n    code",
	"> é\n> 𝔸\n𝔸",
}

func TestBlockQuoteEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, blockQuoteFixtures)
}

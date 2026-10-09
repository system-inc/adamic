package micromark

import "testing"

// codeIndentedFixtures hold no container, setext, fenced code or html syntax: those constructs are not
// ported yet, so a fixture leaning on them would differ for reasons that are not codeIndented's. The
// lazy-line branch of furtherStart is only reachable inside a block quote or list item.
var codeIndentedFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"    a simple\n      indented code block",
	"    chunk1\n\n    chunk2\n  \n \n \n    chunk3",
	"    chunk1\n      \n      chunk2",
	"Foo\n    bar",
	"    foo\nbar",
	"# Heading\n    foo\n# Heading\n    foo\n***",
	"        foo\n    bar",
	"\n    \n    foo\n    ",
	"    foo  ",

	// The prefix: three, four and five columns, and tabs that expand to four.
	"   a",
	"    a",
	"     a",
	"\ta",
	" \ta",
	"  \ta",
	"   \ta",
	"\t\ta",
	"  \t\ta",
	" \t a",
	"\t a",

	// End of input in every state.
	"    ",
	"    \n",
	"    a\n",
	"    a\n\n",
	"    a\n    ",
	"    a\n     ",
	"    a\n\t",
	"    a\n\n\n",

	// furtherStart: blank lines, short indents, tabs, and what ends the block.
	"    a\n    b",
	"    a\n\n\n    b",
	"    a\n   b",
	"    a\n  \n    b",
	"    a\n   \n   b",
	"    a\n\tb",
	"    a\n \tb",
	"    a\n\n\tb",
	"    a\n\n   b",
	"    a\n\n\n\n",
	"    a\n# b",
	"    a\n***",
	"    a\nb\n    c",
	"a\n\n    b",
	"a\n    b\n\n    c",
	"    # a",
	"    ***",
	"    a\\\n    b",
}

func TestCodeIndentedEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, codeIndentedFixtures)
}

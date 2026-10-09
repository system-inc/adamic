package micromark

import (
	"strings"
	"testing"
)

// setextUnderlineFixtures hold no container (block quote, list), emphasis, code, html, character
// reference, definition, link, table, autolink literal, strikethrough, math, liquid or wiki-link syntax:
// those constructs are not ported yet, so a fixture leaning on them would differ for reasons that are not
// setextUnderline's. That leaves two parts of the port unexercised here: the resolver's definition branch
// (definitions are pending) and the lazy-line refusal (lazy lines need a container).
var setextUnderlineFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs. 80 to 82 use emphasis, so they
	// appear with it removed.
	"Foo bar\n=========\n\nFoo bar\n---------\n",
	"Foo bar\nbaz\n====\n",
	"  Foo bar\nbaz\t\n====\n",
	"Foo\n-------------------------\n\nFoo\n=\n",
	"   Foo\n---\n\n  Foo\n-----\n\n  Foo\n  ===\n",
	"    Foo\n    ---\n\n    Foo\n---\n",
	"Foo\n   ----      \n",
	"Foo\n    ---\n",
	"Foo\n= =\n\nFoo\n--- -\n",
	"Foo  \n-----\n",
	"Foo\\\n----\n",
	"Foo\nBar\n---\n",
	"---\nFoo\n---\nBar\n---\nBaz\n",
	"\n====\n",
	"---\n---\n",
	"    foo\n---\n",
	"\\> foo\n------\n",
	"Foo\n\nbar\n---\nbaz\n",
	"Foo\nbar\n\n---\n\nbaz\n",
	"Foo\nbar\n* * *\nbaz\n",
	"Foo\nbar\n\\---\nbaz\n",

	// start: what comes before decides. A paragraph, a blank line, other flow, nothing.
	"===",
	"=",
	"--",
	"==\n==",
	"a\n\n===",
	"a\n \n===",
	"a\n\t\n---",
	"# a\n===",
	"# a\n--",
	"***\n===",
	"***\n--",
	"    code\n===",
	"    code\n--",
	"Foo\n===\n===",
	"Foo\n===\n---",
	"Foo\n---\n---",
	"Foo\n===\nbar\n---",
	"Foo\n===\nbar",

	// before and inside: each marker, a run of either length, a marker switch.
	"Foo\n=",
	"Foo\n-",
	"Foo\n--",
	"Foo\n" + strings.Repeat("=", 100),
	"Foo\n" + strings.Repeat("-", 100),
	"Foo\n=-",
	"Foo\n-=",
	"Foo\n==-",
	"Foo\n--=",
	"Foo\n===a",
	"Foo\n---a",
	"Foo\n===\\",

	// after: trailing whitespace, then end of input, a line ending, or something else.
	"Foo\n= ",
	"Foo\n=\t",
	"Foo\n=  \t ",
	"Foo\n===   \n",
	"Foo\n===\t\nbar",
	"Foo\n=== bar",
	"Foo\n=== =",
	"Foo\n--- a",
	"Foo\n---\t-",

	// End of input and line endings in every state.
	"Foo\n===",
	"Foo\n===\n",
	"Foo\n===\n\n",
	"Foo\r\n===\r\n",
	"Foo\r===\r",
	"Foo\r\n---\r\nbar",
	"Foo\r---\rbar",
	"Foo\r\n=== \r\n",
	"Foo\r\nbar\r\n===",

	// Indentation of the underline and the text.
	"Foo\n ===",
	"Foo\n   ===",
	"Foo\n \t===",
	"Foo\n\t===",
	"Foo\n  \t---",
	"   Foo\n   ---",
	"\tFoo\n---",
	" \tFoo\n---",

	// The text: trailing whitespace, escapes, hard breaks, non-ASCII and astral characters.
	"Foo \n===",
	"Foo\t\n===",
	"Foo\\\nbar\n===",
	"Foo  \nbar\n---",
	"é\n=",
	"𝔸\n---",
	"Foo 𝔸\n𝔸\n===",
	"a\u00a0\n===",
	"\\=\n===",
	"\\-\n---",
}

func TestSetextUnderlineEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, setextUnderlineFixtures)
}

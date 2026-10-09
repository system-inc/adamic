package micromark

import "testing"

// headingAtxFixtures hold no container, emphasis, code, html, link or setext syntax: those constructs
// are not ported yet, so a fixture leaning on them would differ for reasons that are not headingAtx's.
var headingAtxFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"# foo\n## foo\n### foo\n#### foo\n##### foo\n###### foo",
	"####### foo",
	"#5 bolt\n\n#hashtag",
	"\\## foo",
	"# foo \\*baz\\*",
	"#                  foo                     ",
	" ### foo\n  ## foo\n   # foo",
	"    # foo",
	"foo\n    # bar",
	"## foo ##\n  ###   bar    ###",
	"# foo ##################################\n##### foo ##",
	"### foo ###     ",
	"### foo ### b",
	"# foo#",
	"### foo \\###\n## foo #\\##\n# foo \\#",
	"****\n## foo\n****",
	"Foo bar\n# baz\nBar foo",
	"## \n#\n### ###",

	// The opening sequence: the limit, one past, and what may follow it.
	"#",
	"######",
	"#######",
	"###### a",
	"####### a",
	"#a",
	"#\ta",
	"#\t",
	"# ",
	"#\n",
	"#\n\n",
	"##\na",

	// The resolver: whitespace and sequences at either end.
	"# #",
	"# # #",
	"#  #",
	"#\t#",
	"# ## ##",
	"## #a#",
	"# a #b",
	"# #a",
	"# a#\n#b",
	"#  a  #  ",
	"# a\t#\t",
	"# a ###b ###",
	"# foo # bar #",
	"# a\\#",
	"# a\\",
	"# a  ",
	"# a  \n",
	"# a\u00a0#",
	"# é #",
	"# 𝔸 #",
	"# a b c",
	"#   ",
	"# \t a \t # \t ",

	// End of input and line endings in every state.
	"# a",
	"# a\n",
	"# a\n\n",
	"# a #",
	"# a #\n",
	"# a #\nb",

	// Around it: indentation, paragraphs, other flow.
	"   # a",
	"   \t# a",
	" \t# a",
	"\t# a",
	"a\n# b",
	"a\n#",
	"a\n#b",
	"a\n   # b",
	"# a\n# b",
	"# a\n===",
	"# a\n***",
	"***\n# a",
	"# a\nb\n# c",
}

func TestHeadingAtxEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, headingAtxFixtures)
}

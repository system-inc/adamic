package micromark

import "testing"

// gfmTaskListItemFixtures hold no definition, html text, liquid, math, wiki-link, GFM table,
// strikethrough, footnote or autolink literal syntax: those constructs are still pending stubs, so a
// fixture leaning on them would differ for reasons that are not the task list item's. So no `[x]:`, no
// `<`, `{`, `$`, `|`, `~`, `[^` or `[[`, and no `www.`, `http` or `@` in text. The failing checks fall
// through to labelStartLink and labelEnd, which are ported.
var gfmTaskListItemFixtures = []string{
	// GFM examples (5.3 Task list items).
	"- [ ] foo\n- [x] bar\n",
	"- [x] foo\n  - [ ] bar\n  - [x] baz\n- [ ] bim\n",

	// Values: space, `x`, `X`, other characters, none, two.
	"- [ ] foo",
	"- [x] foo",
	"- [X] foo",
	"- [y] foo",
	"- [*] foo",
	"- [-] foo",
	"- [] foo",
	"- [xx] foo",
	"- [ x] foo",
	"- [x ] foo",
	"- [  ] foo",
	"- [é] foo",
	"- [😀] foo",

	// Tab as the value: one column wide, then wide enough to leave virtual spaces before `]`.
	"- [\t] foo",
	"-  [\t] foo",
	"- [ \t] foo",

	// Line ending as the value (`[` at the end of the first line, `]` on a lazy line).
	"- [\n] foo",
	"- [\r\n] foo",
	"- [\r] foo",
	"- [\n  ] foo",

	// After the closing bracket: missing space, tab, several spaces, line endings, end of input.
	"- [x]foo",
	"- [ ]foo",
	"- [x]\tfoo",
	"- [ ]\t\tfoo",
	"- [x]  foo",
	"- [x] \tfoo",
	"- [x]",
	"- [ ]",
	"- [x] ",
	"- [x]   ",
	"- [x]\t",
	"- [x]\n",
	"- [x] \n",
	"- [x]\nfoo",
	"- [x]\r\nfoo",
	"- [x]\rfoo",
	"- [x]   \nfoo",
	"- [x]\n  foo",
	"- [x] \r\n  foo",
	"- [x]é",
	"- [x]😀",
	"- [x] é",
	"- [x] 😀",
	"- [x]*a*",
	"- [x] *a*",
	"- [x]\\ foo",
	"- [x](y) foo",
	"- [x] [y] foo",
	"- [x][y] foo",

	// End of input inside the check.
	"- [",
	"- [x",
	"- [ ",
	"- [\t",

	// Other bullet markers, ordered items, wider prefixes.
	"* [x] foo",
	"+ [ ] foo",
	"1. [x] foo",
	"1) [ ] foo",
	"123. [X] foo\n124. [ ] bar\n",
	"-   [x] foo",
	"-\t[x] foo",
	"-     [x] foo",
	"1. [x]\n   foo\n",

	// Nested items and lists in block quotes.
	"- - [x] foo",
	"1. - [ ] foo",
	"- foo\n  - [x] bar\n",
	"> - [x] foo",
	"- > [x] foo",

	// Not the first content of the list item.
	"- foo [x] bar",
	"- *a* [x] b",
	"- foo\n  [x] bar",
	"- foo\n\n  [x] bar",
	"- # [x] foo",
	"- # foo\n  [x] bar",
	"- \n  [x] foo",
	"-\n  [x] foo",
	"- ***\n- [x] foo",

	// Several items, loose and tight, and line endings between them.
	"- [x] a\n- [ ] b\n- c",
	"- [x] a\n\n- [ ] b\n",
	"- [x] a\r\n- [ ] b\r\n",
	"- [x] a\r- [ ] b\r",
	"- [x] foo\n  bar\n",
	"- [x] foo\n  ===\n",

	// Outside lists.
	"[x] foo",
	"[ ] foo",
	"> [x] foo",
	"# [x] foo",
	"foo\n[x] bar",
}

func TestGfmTaskListItemEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, gfmTaskListItemFixtures)
}

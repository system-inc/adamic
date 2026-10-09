package micromark

import "testing"

// codeFencedFixtures hold no block quote, list, setext, html, definition, emphasis, strikethrough or
// math syntax: those constructs are not ported yet, so a fixture leaning on them would differ for reasons
// that are not codeFenced's. Where a fence fails and its backticks fall to a paragraph, the fixture is
// shaped so codeText could not match there either. The lazy-line branch of nonLazyContinuation is only
// reachable inside a block quote or list item, and the unlimited closing prefix only when codeIndented
// is disabled, which the fork never does.
var codeFencedFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"```\n<\n >\n```\n",
	"~~~\n<\n >\n~~~\n",
	"```\naaa\n~~~\n```\n",
	"~~~\naaa\n```\n~~~\n",
	"````\naaa\n```\n``````\n",
	"~~~~\naaa\n~~~\n~~~~\n",
	"```\n",
	"`````\n\n```\naaa\n",
	"```\n\n  \n```\n",
	"```\n```\n",
	" ```\n aaa\naaa\n```\n",
	"  ```\naaa\n  aaa\naaa\n  ```\n",
	"   ```\n   aaa\n    aaa\n  aaa\n   ```\n",
	"    ```\n    aaa\n    ```\n",
	"```\naaa\n  ```\n",
	"   ```\naaa\n  ```\n",
	"```\naaa\n    ```\n",
	"~~~~~~\naaa\n~~~ ~~\n",
	"foo\n```\nbar\n```\nbaz\n",
	"```ruby\ndef foo(x)\n  return 3\nend\n```\n",
	"~~~~    ruby startline=3 $%@#$\ndef foo(x)\n  return 3\nend\n~~~~~~~\n",
	"````;\n````\n",
	"~~~ aa ``` ~~~\nfoo\n~~~\n",
	"```\n``` aaa\n```\n",

	// Opening sequence: too short, exactly three, longer, and a mixed run that stops the sequence.
	"``",
	"`",
	"```",
	"~~~",
	"~~~~~~~~~~",
	"``x",
	"~~~`",
	"```~~~",

	// End of input in every state of the opening fence.
	"``` ",
	"```js",
	"```js ",
	"```js eval",
	"```js eval ",
	"```\t",
	"```\tjs\tmeta\t",
	"~~~ js   eval  more  ",

	// Info and meta: backticks forbidden only after a backtick fence, escapes and references inside.
	"```a`b",
	"``` a b`c",
	"~~~a`b\nx\n~~~",
	"~~~ a b`c\nx\n~~~",
	"```a\\`b",
	"```&amp; &#96;\nx\n```",
	"```\\\nx\n```",
	"```jsé méta\nx\n```",
	"```\U0001F600 \U0001F600\n\U0001F600\n```",

	// Content: blank lines, empty lines at end, eof right after a line ending.
	"```\na",
	"```\na\n",
	"```\na\n\n",
	"```\n\n",
	"```\n\n\n",
	"```\n  \n",
	"```\n\t\n",

	// The initial prefix, removed from content up to its own width, tabs counted as columns.
	" ```\n  a\n a\na",
	"  ```\n    a\n\ta",
	"   ```\n\ta\n```",
	"\t```\nx",
	"  ```\n \t a\n  ```",
	" ```\n\n \n```",

	// The closing fence: indentation up to three, four is content, trailing whitespace, junk after it,
	// shorter and longer sequences, the other marker, and a tab prefix.
	"```\na\n   ```",
	"```\na\n```   ",
	"```\na\n```\t",
	"```\na\n``` x\n```",
	"```\na\n```x\n```",
	"````\na\n```\n````",
	"```\na\n`````",
	"~~~\na\n```\n~~~",
	"```\na\n\t```",
	"```\na\n \t```",
	"```\na\n``\n```",
	"```\na\n```\nafter",
	"```\na\n```\n\n```\nb\n```",

	// Line endings: carriage return and carriage return line feed throughout.
	"```js\r\na\r\n```\r\n",
	"```js\ra\r```\r",
	"~~~ x y\r\n\r\n  b\r\n~~~",
	"  ```\r  a\r```",

	// Interrupting a paragraph, and the following constructs.
	"a\n```",
	"a\n```js\nb\n```",
	"a\n~~~ x\nb",
	"# h\n```\nc\n```\n# h",
	"***\n```\nc\n```\n***",
	"    code\n```\nc\n```",
	"```\nc\n```\n    code",

	// Non-ASCII and astral next to the fence.
	"é```\nx",
	"```é\nx\n```",
	"```\n\U0001F600\n```\U0001F600",
}

func TestCodeFencedEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, codeFencedFixtures)
}

package micromark

import "testing"

// liquidFixtures hold no attention (`*`, `_`), label end (`]`, so no `[` or `![` either), html or
// autolink (`<`), definition, list, table (`|`), strikethrough (`~`), math (`$`), wiki-link, footnote,
// task list or GFM autolink literal (`www.`, `http`, `@`) syntax: those constructs are still pending,
// so a fixture leaning on them would differ for reasons that are not liquid's. Liquid is the fork's own
// syntax, so there are no CommonMark or GFM spec examples for it.
var liquidFixtures = []string{
	// Both forms, empty and filled.
	"{{ a }}",
	"{% a %}",
	"{{}}",
	"{%%}",
	"{{ }}",
	"{% %}",
	"{{a}}",
	"{% if a %}b{% endif %}",
	"{{ a | b: 'c' }}",

	// In a paragraph, against its neighbours.
	"a {{ b }} c",
	"a{{b}}c",
	"{{ a }}{{ b }}",
	"{{ a }}{% b %}",
	"{{ a }}\n",
	"{{ a }}\n\n",
	"{{ a }}  ",
	"  {{ a }}",
	"a\t{{ b }}\t",

	// start: what may follow the first brace, and end of input there.
	"{",
	"{a",
	"{ {a }}",
	"{a}",
	"{\n{ a }}",
	"}}",
	"%}",
	"{{",
	"{%",

	// inside and mayClose: unclosed, a closing code without its brace, the wrong closer.
	"{{ a",
	"{% a",
	"{{ a }",
	"{% a %",
	"{{ a } }",
	"{% a % }",
	"{{ a %}",
	"{% a }}",
	"{{ a }x}}",
	"{% a %x%}",
	"{% a %%}",
	"{{ a }}}",
	"{{{ a }}}",
	"{%{ a }%}",
	"{{ a }}} b }}",
	"{{ a  ",

	// Nesting attempts: the first closer wins.
	"{{ {{ a }} }}",
	"{% {% a %} %}",
	"{{ {% a %} }}",
	"{% {{ a }} %}",
	"{% if %}{{ a }}{% endif %}",

	// Line endings inside, in each form and between a closing code and its brace.
	"{{ a\nb }}",
	"{{ a\r\nb }}",
	"{{ a\rb }}",
	"{% a\nb %}",
	"{{\n}}",
	"{%\r\n%}",
	"{{ a }\n}}",
	"{{ a }\r\n}}",
	"{% a %\r}",
	"{{ a\n",
	"{{ a\n\nb }}",
	"{{ a\r\n\r\nb }}",
	"{{ a  \nb }}",
	"{{ a \nb }}",
	"{{ a\t\nb }}",
	"{{ a\\\nb }}",
	"{{ a\n  b }}",
	"{{ a\n    b }}",
	"{{ a\nb\nc }}",
	"{{ a }}  \nb",
	"{{ a }}\\\nb",

	// Flow around it: what interrupts the paragraph cuts the span short.
	"{{ a\n# b }}",
	"{{ a\n***\n}}",
	"{{ a\n> b }}",
	"{{ a\n```\n}}",
	"{{ a\n===",
	"{{ a }}\n===",
	"{{ a\nb }}\n---",
	"# {{ a }}",
	"# {{ a }} #",
	"## {% a %} ##",
	"# {{ a\n}}",
	"> {{ a }}",
	"> {{ a\n> b }}",
	"> {{ a\nb }}",
	"    {{ a }}",
	"\t{{ a }}",
	"```\n{{ a }}\n```",
	"```{{ a }}\n```",
	"***\n{{ a }}\n***",

	// Other ported text constructs next to it and inside it.
	"\\{{ a }}",
	"{\\{ a }}",
	"{{ a \\}}",
	"{{ a \\}} }}",
	"`{{ a }}`",
	"{{ `a` }}",
	"{{ a ` }} `",
	"`a {{ b` }}",
	"&amp;{{ a }}",
	"{{ &amp; }}",
	"&#123;{ a }}",
	"{{ a }}&copy;",
	"{&#123; a }}",

	// Non-ASCII and astral characters next to it and inside it.
	"é{{ é }}é",
	"𝔸{{ 𝔸 }}𝔸",
	"{{ 😀\n😀 }}",
	"{%\u00a0%}",
	"{{ a }\u00a0}}",
	"{{\ta\t}}",
}

func TestLiquidEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, liquidFixtures)
}

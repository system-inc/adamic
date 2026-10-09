package micromark

import "testing"

// mathFixtures cover math flow and math text. They hold no attention (`*`, `_`), label end (`]`), html
// (`<`), definition, list, liquid (`{{`, `{%`), wiki-link (`[[`), strikethrough (`~`), table (`|`),
// footnote or GFM autolink literal (`www.`, `http`, `@`) syntax: those constructs are still pending, so a
// fixture leaning on them would differ for reasons that are not math's. Neither CommonMark nor GFM has
// spec examples for math, so the first group is micromark-extension-math's readme and test examples
// that stay inside ported constructs.
var mathFixtures = []string{
	// The extension's own examples.
	"$$\n\\frac{1}{2}\n$$\n",
	"$\\frac{1}{2}$\n",
	"$$\\frac{1}{2}$$\n",
	"Lift($L$) can be determined by Lift Coefficient ($C_L$) like the following equation.\n",
	"$$\nL = \\frac{1}{2} \\rho v^2 S C_L\n$$\n",
	"$$asciimath\nx < y\n$$\n",
	"a \\$b$\n",
	"a \\$$b$$\n",
	"a $$b$\n",
	"$$ b $$\n",
	"$ $\n",
	"$  $\n",
	"$\na\n$\n",

	// mathText sequenceOpen: one, two, many, and the end of input inside the opening.
	"$",
	"$$",
	"a$",
	"a$$",
	"a$$$",
	"$a$",
	"$$a$$",
	"$$$$$$$$$$a$$$$$$$$$$",
	"$$$$$$$$$$a$$$$$$$$$",
	"x $a$$ y",
	"x $$a$ y",
	"x $$ $a$ $$ y",

	// mathText between and data: end of input after spaces and line endings, each line ending kind.
	"x $ ",
	"x $\n",
	"x $ a",
	"x $a b$",
	"x $a  b$",
	"x $   a   $",
	"x $a\nb$",
	"x $a\r\nb$",
	"x $a\rb$",
	"x $\na\n$",
	"x $\r\na\r\n$",
	"x $\ra\r$",
	"x $ \na\n $",
	"x $\n$",
	"x $a\tb$",
	"x $\ta\t$",
	"x $a\n\nb$",

	// Previous: a dollar after an escape, after a dollar, and non-ASCII and astral neighbors.
	"\\$$a$",
	"\\\\$a$",
	"é$a$é",
	"😀$a$😀",
	"$😀$",
	"x $é 😀 ü$ y",

	// mathFlow opening: too short, meta, meta with a dollar, whitespace, and end of input.
	"$$",
	"$$\n",
	"$$ ",
	"$$$",
	"$$$\na\n$$$",
	"$$$\na\n$$",
	"$$\na\n$$$$",
	"$$ meta\na\n$$",
	"$$meta  more \na\n$$",
	"$$ me$ta\na\n$$",
	"$$\t\ta\n$$",
	"$$ \\frac\n$$",

	// mathFlow content: empty lines, end of input in content, every line ending kind.
	"$$\n\n\na\n\n$$",
	"$$\na",
	"$$\na\n",
	"$$\r\na\r\n$$\r\n",
	"$$\ra\r$$\r",
	"$$\n$$",
	"$$\n$$\n$$\n$$",

	// mathFlow closing: indentation up to the limit and one past, trailing whitespace, trailing junk.
	"$$\na\n $$",
	"$$\na\n   $$",
	"$$\na\n    $$",
	"$$\na\n\t$$",
	"$$\na\n$$  ",
	"$$\na\n$$\t",
	"$$\na\n$$ b",
	"$$\na\n$$b\n$$",

	// initialSize: an indented opening strips up to that much indentation from content.
	" $$\n a\n  b\nc\n $$",
	"  $$\n  a\n   b\n\tc\n  $$",
	"   $$\n    a\n   $$",
	"   $$\n\ta\n   $$",
	"    $$\n    a\n    $$",

	// Interrupting a paragraph, and paragraphs around it.
	"a\n$$\nb\n$$",
	"a\n$$ meta\nb",
	"$$\na\n$$\nb",
	"a\n\n$$\nb\n$$\n\nc",

	// Containers: block quotes, lazy lines, and a closing fence outside the quote.
	"> $$\n> a\n> $$",
	"> $$\na\n$$",
	"> $$\n> a\n$$",
	"> $$",
	">  $$\n>  a\n>   b\n>  $$",
	"> x $a$ y",
}

func TestMathEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, mathFixtures)
}

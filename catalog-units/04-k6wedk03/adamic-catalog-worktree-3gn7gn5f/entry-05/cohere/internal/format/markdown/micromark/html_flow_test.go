package micromark

import "testing"

// htmlFlowFixtures hold no container, emphasis, code fence, code span, autolink, link end, table, math or
// liquid syntax outside the HTML itself: those constructs are not ported yet. Text that is not HTML (flow)
// also never forms HTML (text) or an autolink upstream, since html-text and autolink are still stubs, so
// a fixture that falls back to a paragraph only does so where upstream's inline `<` fails too. That rules
// out the lazy-line branches, which need a block quote or list, and a failed `<![` opening, whose `![`
// becomes a labelImage that only label-end's resolver turns back into data.
var htmlFlowFixtures = []string{
	// CommonMark examples, the ones that stay inside ported constructs.
	"<table>\n  <tr>\n    <td>\n           hi\n    </td>\n  </tr>\n</table>\n\nokay.\n",
	" <div>\n  *hello*\n         <foo><a>\n",
	"</div>\n*foo*\n",
	"<div id=\"foo\"\n  class=\"bar\">\n</div>\n",
	"<div id=\"foo\" class=\"bar\n  baz\">\n</div>\n",
	"<div id=\"foo\"\n*hi*\n",
	"<div class\nfoo\n",
	"<div *???-&&&-<---\n*foo*\n",
	"<div><a href=\"bar\">*foo*</a></div>\n",
	"<table><tr><td>\nfoo\n</td></tr></table>\n",
	"<a href=\"foo\">\n*bar*\n</a>\n",
	"<Warning>\n*bar*\n</Warning>\n",
	"<i class=\"foo\">\n*bar*\n</i>\n",
	"</ins>\n*bar*\n",
	"<del>\n*foo*\n</del>\n",
	"<pre language=\"haskell\"><code>\nimport Text.HTML.TagSoup\n\nmain :: IO ()\nmain = print $ parseTags tags\n</code></pre>\nokay\n",
	"<script type=\"text/javascript\">\n// JavaScript example\n\ndocument.getElementById(\"demo\").innerHTML = \"Hello JavaScript!\";\n</script>\nokay\n",
	"<textarea>\n\n*foo*\n\n_bar_\n\n</textarea>\n",
	"<style\n  type=\"text/css\">\nh1 {color:red;}\n\np {color:blue;}\n</style>\nokay\n",
	"<style\n  type=\"text/css\">\n\nfoo\n",
	"<script>\nfoo\n</script>1. *bar*\n",
	"<!-- Foo\n\nbar\n   baz -->\nokay\n",
	"<?php\n\n  echo '>';\n\n?>\nokay\n",
	"<!DOCTYPE html>\n",
	"<![CDATA[\nfunction matchwo(a,b)\n{\n  if (a < b && a < 0) then {\n    return 1;\n\n  } else {\n\n    return 0;\n  }\n}\n]]>\nokay\n",
	"  <!-- foo -->\n\n    <!-- foo -->\n",
	"  <div>\n\n    <div>\n",
	"Foo\n<div>\nbar\n</div>\n",
	"<div>\nbar\n</div>\n*foo*\n",
	"<div>\n*Emphasized* text.\n</div>\n",
	"<table>\n\n<tr>\n\n<td>\nHi\n</td>\n\n</tr>\n\n</table>\n",
	"<table>\n\n  <tr>\n\n    <td>\n      Hi\n    </td>\n\n  </tr>\n\n</table>\n",

	// Openings that fail, and end of input in each opening state.
	"<",
	"< a>",
	"<1>",
	"<!",
	"<!>",
	"<!-",
	"<!-x",
	"</",
	"</1>",
	"<a",
	"<dív>",

	// Kind 1, raw: the names, case, the end tag and its length limit.
	"<script>",
	"<pre",
	"<pre\nb",
	"<PRE>\na\n\nb</Pre>c\nd",
	"<style>a</style>",
	"<script>a</scriptx>\nb</script>",
	"<textarea>a</textarea>\nb",
	"<textarea>a</textareax>\nb\n</TEXTAREA>",
	"<script>a</ >b<</script>",
	"<pre/>",
	"</pre>",
	"<prex>",

	// Kind 2, comment.
	"<!---->",
	"<!-->",
	"<!--->",
	"<!-- a -- b -->c",
	"<!-- a --->",
	"<!-- -x-->",
	"<!-- a\nb",
	"<!-- a --",

	// Kind 3, instruction.
	"<?",
	"<?>",
	"<??>",
	"<?x\ny?>z\nw",

	// Kind 4, declaration.
	"<!A>",
	"<!a\nb>c\nd",
	"<!doctype html>trailing",

	// Kind 5, CDATA.
	"<![CDATA[]]>",
	"<![CDATA[",
	"<![CDATA[x]]]>",
	"<![CDATA[x]>]]>\ny",

	// Kind 6, basic: names, the self-closing slash, blank lines ending it.
	"<div",
	"<DIV>",
	"<h1>",
	"<search>",
	"<div\tclass>",
	"<div/>",
	"<div/",
	"<div/x",
	"<div>\n",
	"<div>\n\n",
	"<div>\n  \nx",
	"<div>\n\t\nx",
	"<div>\nfoo\n\n",
	"<div>\n# a",
	"<div>\n\n# a",

	// Kind 7, complete: attributes, quotes, closing tags, and what may follow the `>`.
	"<h7>",
	"<a>",
	"<a >",
	"<a/>",
	"<a />",
	"<a/",
	"<a b>",
	"<a b c>",
	"<a b=c>",
	"<a b='c'>",
	"<a b=\"c\">",
	"<a :b _c d.e-f:g>",
	"<a b = c>",
	"<a b=\"c\"/>",
	"<a\tb>",
	"<x-y>",
	"</a>",
	"</a >",
	"<a>  ",
	"<a>\t",
	"<a>\nb",
	"<a>\n\nb",
	"<a b",
	"<a b=",
	"<a b=\"",
	"<a b=\"c\"",
	"<a b=\">",
	"<a b=c",
	"<a =b>",
	"<a b=>",
	"<a b='c'd>",
	"<a b=`c>",
	"</a b>",
	"<a b=c\n",

	// Interrupting a paragraph: every kind but 7.
	"a\n<div>",
	"a\n<div>\nb",
	"a\n</div>",
	"a\n<div/>",
	"a\n<div/x",
	"a\n<!-- x -->",
	"a\n<?x?>",
	"a\n<!X>",
	"a\n<![CDATA[x]]>",
	"a\n<script>",
	"a\n<style\n",
	"a\n<pre",
	"a\n<x",
	"a\n<x\nb",
	"a\n  <div>",

	// Indentation, the line prefix the resolver folds in.
	"   <div>",
	"    <div>",
	"\t<div>",
	" \t<div>",
	"a\n\n   <!-- x -->\n   b",

	// Line endings.
	"<div>\r\n\r\nx",
	"<div>\rb\r\rc",
	"<!-- a\r\nb -->\r\nc",
	"<script>\ra\r</script>\rb",

	// Non-ASCII and astral characters next to it.
	"<div>é",
	"<div>𝔸",
	"<!-- 𝔸 -->",
	"é\n<div>",
	"<a b=\"𝔸\">",
	"<a b=é>",

	// Around it: other flow.
	"# a\n<div>",
	"***\n<div>",
	"<div>\n***",
	"<!-- a -->\n***",
}

func TestHtmlFlowEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, htmlFlowFixtures)
}

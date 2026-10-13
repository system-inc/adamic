package markdownblocks

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
)

// Not parallel: listLayoutShared setup cache and shared markdownMemory configuration; existing helper controls parallel execution. shared layout fixtures and memory-budget admission are handled by the setup helper. shared listLayoutShared and listLayoutBuckets; existing helper controls parallel execution.
// Not parallel: list-layout fixture and markdown memory configuration.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownListLayout(t *testing.T) {
	t.Parallel()
	buildListLayoutSetup(t)
}

// Not parallel: quoteLayoutShared is initialized for the parallel quote layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: quote lowered and Go build products initialized before parallel shards.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownQuoteLayout(t *testing.T) {
	quoteLayoutSetup(t)
}

// Not parallel: tableLayoutSharedProducts and tableLayoutInputs are initialized for the parallel table layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Legacy entry point reports shared preparation completed before m.Run.
func TestMarkdownTableLayout(t *testing.T) {
	t.Parallel()
	testMarkdownTableLayout(t)
}

type layoutFixture struct {
	selection                                                                  gatesample.Selection
	mutantNativeCases                                                          string
	mutantInputs                                                               []auditInput
	mutantWant                                                                 []byte
	root, directory, nativeCases, canonicalCases, main, fork, goLayout, script string
	inputs                                                                     []auditInput
	files                                                                      int
	want                                                                       []byte
	program                                                                    *ir.Program
}

// Not parallel: shared leaf layout shard setup and markdownMemory configuration; existing helper controls parallel execution. shared layout fixtures and memory-budget admission are handled by the setup helper. shared leafCompositionPrepared and leafCompositionBuilds; existing helper controls parallel execution.
// Not parallel: markdown memory configuration checked by parallelMarkdown before parallel admission.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownLeafComposition(t *testing.T) {
	t.Parallel()
	configureMarkdownMemory(t)
	testMarkdownLeafShards(t)
}

// Not parallel: structureLayoutComplete is initialized for the parallel structure layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: structure-layout fixture and markdown memory configuration.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownStructureLayout_Setup(t *testing.T) {
	configureMarkdownMemory(t)
	structureLayoutShared(t)
}

func blockCorpus(t *testing.T, root, slice string) ([]auditInput, int) {
	t.Helper()
	inputs, files := auditCorpus(t, root)
	for _, marker := range []string{"-", "*", "+", "0.", "1.", "1)", "9.", "10)", "99.", "999999999."} {
		for _, spacing := range []string{" ", "  ", "   ", "    ", "     ", "\t"} {
			for _, task := range []string{"", "[ ] ", "[x] ", "[X] "} {
				for _, body := range []string{"a *b _c_* `x`", "a\n\n    b\n\n    - c", "a\n\n    > q\n    > r", "a\n<div>\nx\n</div>", "\n\n\n    code"} {
					text := marker + spacing + task + body + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/list-layout/" + text, Text: text})
				}
			}
		}
	}
	for _, text := range []string{
		"0. a\n1. b\n", "0. a\n1. b\n1. c\n", "2. a\n1. b\n7. c\n", "999999999. a\n1. b\n",
		"- a\n* b\n+ c\n- d\n", "1. a\n2) b\n3. c\n", "-\n\n\n    x\n", "1.\n\n\n        x\n",
		"- [x] a\n\n    - [ ] b\n      c\n", "- a\n\n  ```js\n  let x = 1;\n  ```\n",
		"<!-- prettier-ignore -->\n*  a\n+   b\n", "> <!-- prettier-ignore -->\n> *  a\n>\n",
		"- a\n\n  |a|b|\n  |-|-|\n  |x|y|\n", "- a\n\n  ### h\n\n  ---\n",
		"- <div>\n  x\n  </div>\n", "- a\n\n  [x]: https://example.test\n  [y]: /y\n",
		"- a\n\n  [^x]: footnote\n\n- [^x]\n", "1.    中😀\n1.    _a*b*_\n",
	} {
		inputs = append(inputs, auditInput{Name: "generated/list-layout/edge/" + text, Text: text})
	}
	if slice != "lists" {
		for _, prefix := range []string{">", "> ", " > ", "  >  ", "   >\t", "> > ", ">> ", "> > > ", ">\t> "} {
			for _, body := range []string{"", "text", "a\nb", "a\n\nb", "# h\n\na", "a\n---", "- a\n- b", "1. a\n\n   - b", "- [x] a", "```js\nlet x=1;\n```", "    x\n    y", "<div>\nx\n</div>", "a\n<div>\nx", "[a]: /x\n[b]: /y", "| a | b |\n| - | - |\n| x | y |", "<!-- prettier-ignore -->\n+    a", "中😀 _a*b*_"} {
				text := prefix + strings.ReplaceAll(body, "\n", "\n"+prefix) + "\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/quote-layout/" + text, Text: text})
			}
		}
	}
	if slice != "lists" && slice != "quotes" {
		for _, align := range []string{"---", ":--", "--:", ":-:"} {
			for _, cell := range []string{"", "a", "abcde", "abcdef", "中😀", "*a _b_*", "`a|b`", "a\\|b", "[a](/b)", "<em>a</em>"} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := "| a | b |\n| " + align + " | " + align + " |\n| " + cell + " | x |\n| long text | y |\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/table-layout/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"a | b\n- | -\nc\nd | e | f\n", "a | b\n:- | -:\n | \n", "| a |\n| - |\n| x | y |\n", "| 😀 | 中 |\n| :-: | -: |\n| é | 👨‍👩‍👧‍👦 |\n"} {
			inputs = append(inputs, auditInput{Name: "generated/table-layout/edge/" + text, Text: text})
		}
	}
	if slice == "code" || (slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace") {
		for _, marker := range []string{"```", "````", "~~~~", "```````"} {
			for _, info := range []string{"", "js", "json", "yaml", "toml", "css", "html", "text title=foo", "x {#id .class}"} {
				for _, body := range []string{"", "a", "a\nb", "\n\nx\n", "a  \nb\t", "`a`", "```", "~~~~", "中😀"} {
					for _, prefix := range []string{"", "> ", "- ", "> - "} {
						text := marker + info + "\n" + body + "\n" + marker + "\n"
						lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
						text = prefix + lines[0] + "\n"
						continuation := prefix
						if strings.HasSuffix(prefix, "- ") {
							continuation = strings.TrimSuffix(prefix, "- ") + "  "
						}
						for _, line := range lines[1:] {
							text += continuation + line + "\n"
						}
						inputs = append(inputs, auditInput{Corpus: true, Name: "generated/code-layout/" + text, Text: text})
					}
				}
			}
		}
		for _, indent := range []string{"    ", "\t", "     ", "  \t", "\t "} {
			for _, body := range []string{"a", "a  \nb\t", "\n\nx\n", "中😀"} {
				text := indent + strings.ReplaceAll(body, "\n", "\n"+indent) + "\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/code-layout/indent/" + text, Text: text})
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, body := range []string{"<div>\nx  \n</div>", "<script>\nx  \n</script>", "<style>\nx\t\n</style>", "<pre>\nx  \n</pre>", "<!-- a  \nb\t -->", "<!-->", "<!--->", "<!--a-->", "<?xml\nx  \n?>", "<!DOCTYPE html>", "<![CDATA[\nx  \n]]>", "<table>\n<tr>\nx\n</tr>\n</table>", "<x-a a='b'>\nx\n</x-a>", "a <em>\nx  \n</em> b", "<div>\n\n# h\n\n</div>", "<!-- prettier-ignore -->\n<div>  \nx\n</div>", "<div>中😀</div>"} {
			for _, prefix := range []string{"", "> ", "> > ", "- ", "> - "} {
				for _, ending := range []string{"", "  ", "\t", "\u00a0", "\u2000", "\ufeff"} {
					text := body + ending + "\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/html-layout/" + text, Text: text})
				}
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, cell := range []string{"中", "Ａ", "α", "é", "©", "©️", "👩🏽‍⚕️", "👨🏻‍❤️‍💋‍👨🏿", "🧑🏿‍🦽‍➡️", "🏳️‍🌈", "🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f"} {
			for _, align := range []string{"---", ":--", "--:", ":-:"} {
				text := "| " + cell + " | a |\n| " + align + " | --- |\n| a | " + cell + " |\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/native-width/table/" + text, Text: text})
			}
		}
	}
	if slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for depth := 1; depth <= 6; depth++ {
			for _, body := range []string{"", "a", "*a _b_*", "中 👩🏽‍⚕️", "[a](/b)", "a #", "`a  b`"} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := prefix + strings.Repeat("#", depth) + " " + body + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/structure/atx/" + text, Text: text})
				}
			}
		}
		for _, marker := range []string{"=", "===", "---", "-----"} {
			for _, body := range []string{"a", "a\nb", "*a _b_*", "中\n文", "a\n\ntext", "[x](/y)"} {
				for _, spacing := range []string{"", "  ", "\t"} {
					text := body + "\n" + marker + spacing + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/structure/setext/" + text, Text: text})
				}
			}
		}
	}
	if slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, space := range []string{"", " ", "\t", "\u00a0", "\u2028", "\ufeff"} {
			for _, body := range []string{"+    a\n*   b", "a\n====", "|a|b|\n|-|-|\n|中|😀|", "```js\na  b\n```", "<div>  \nx\n</div>"} {
				for _, marker := range []string{"next", "range", "nested", "unmatched"} {
					comment := "<!--" + space + "prettier-ignore" + space + "-->"
					text := comment + "\n" + body + "\n\n# after\n"
					if marker != "next" {
						text = "before\n\n<!--" + space + "prettier-ignore-start" + space + "-->\n" + body + "\n"
						if marker == "nested" {
							text += "<!-- prettier-ignore-start -->\n+   extra\n"
						}
						if marker != "unmatched" {
							text += "<!--" + space + "prettier-ignore-end" + space + "-->\n\n# after\n"
						}
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/root/ignore/" + marker + "/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"", " \n", "a\n<div>\nx\n</div>\n", "[a]: /x\n[b]: /y\n", "<!-- prettier-ignore-start -->\n+    a\n<!-- prettier-ignore-end -->\n\n<!-- prettier-ignore-start -->\n*    b\n<!-- prettier-ignore-end -->\n", "<!-- prettier-ignore-end -->\n+    a\n"} {
			inputs = append(inputs, auditInput{Name: "generated/root/edge/" + text, Text: text})
		}
	}
	if slice == "leaves" || slice == "whitespace" {
		for _, label := range []string{"a", "中", "a b", "x-y"} {
			for _, url := range []string{"/x", "", "<x>", "/a%20b", "https://example.test"} {
				for _, title := range []string{"", " \"title\"", " 'a \"b\"'", " (a'b)"} {
					text := "[" + label + "](" + url + title + ") ![" + label + "](" + url + title + ")\n\n[" + label + "]: " + url + title + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/leaves/links/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"*a* **b** ~~c~~ _d_ 1*2*3 1***2***3\n", "_<https://example.test>_\n", "a  \nb\nc\\\nd\n", "[^x]\n\n[^x]: first\n    second\n\n    - a\n    - b\n", "- a\n\n  ***\n\n- b\n\n  ---\n", "$$ title\na+b\n$$\n\n$x + y$\n", "{{ a }} {% b %}\n", "[[wiki link]]\n", "![a][b] [a][b] ![b][] [b][] ![b] [b]\n\n[b]: /x 'title'\n"} {
			inputs = append(inputs, auditInput{Name: "generated/leaves/edge/" + text, Text: text})
		}
	}
	if slice == "whitespace" {
		for _, previous := range []string{"a", "中", "한", "。", "α", "…"} {
			for _, next := range []string{"a", "中", "한", "1.", "123)", "。", "-"} {
				for _, space := range []string{" ", "\n", "\t"} {
					text := previous + space + next + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/whitespace/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"[a  中][b]\n\n[b]: /x\n", "![a  中][b] [a\n中][b]\n\n[b]: /x\n", "|a b|中 c|\n|-|-|\n|x y|한 z|\n", "[^x]\n\n[^x]: a\n\n    {{ b }}\n"} {
			inputs = append(inputs, auditInput{Name: "generated/whitespace/edge/" + text, Text: text})
		}
	}
	return inputs, files
}

package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func TestMarkdownListLayout(t *testing.T) {
	t.Parallel()
	testBlockLayout(t, "lists")
}

func TestMarkdownQuoteLayout(t *testing.T) {
	t.Parallel()
	testBlockLayout(t, "quotes")
}

func TestMarkdownTableLayout(t *testing.T) {
	t.Parallel()
	testBlockLayout(t, "tables")
}

func TestMarkdownCodeBlockLayout(t *testing.T) {
	t.Parallel()
	testBlockLayout(t, "code")
}

func TestMarkdownHTMLBlockLayout(t *testing.T) {
	t.Parallel()
	testBlockLayout(t, "html")
}

func testBlockLayout(t *testing.T, slice string) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := auditCorpus(t, root)
	for _, marker := range []string{"-", "*", "+", "0.", "1.", "1)", "9.", "10)", "99.", "999999999."} {
		for _, spacing := range []string{" ", "  ", "   ", "    ", "     ", "\t"} {
			for _, task := range []string{"", "[ ] ", "[x] ", "[X] "} {
				for _, body := range []string{"a *b _c_* `x`", "a\n\n    b\n\n    - c", "a\n\n    > q\n    > r", "a\n<div>\nx\n</div>", "\n\n\n    code"} {
					text := marker + spacing + task + body + "\n"
					inputs = append(inputs, auditInput{Name: "generated/list-layout/" + text, Text: text})
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
				inputs = append(inputs, auditInput{Name: "generated/quote-layout/" + text, Text: text})
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
					inputs = append(inputs, auditInput{Name: "generated/table-layout/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"a | b\n- | -\nc\nd | e | f\n", "a | b\n:- | -:\n | \n", "| a |\n| - |\n| x | y |\n", "| 😀 | 中 |\n| :-: | -: |\n| é | 👨‍👩‍👧‍👦 |\n"} {
			inputs = append(inputs, auditInput{Name: "generated/table-layout/edge/" + text, Text: text})
		}
	}
	if slice == "code" || slice == "html" {
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
						inputs = append(inputs, auditInput{Name: "generated/code-layout/" + text, Text: text})
					}
				}
			}
		}
		for _, indent := range []string{"    ", "\t", "     ", "  \t", "\t "} {
			for _, body := range []string{"a", "a  \nb\t", "\n\nx\n", "中😀"} {
				text := indent + strings.ReplaceAll(body, "\n", "\n"+indent) + "\n"
				inputs = append(inputs, auditInput{Name: "generated/code-layout/indent/" + text, Text: text})
			}
		}
	}
	if slice == "html" {
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
					inputs = append(inputs, auditInput{Name: "generated/html-layout/" + text, Text: text})
				}
			}
		}
	}
	dir := t.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	bridge, err := filepath.Abs("testdata/list_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs("testdata/list_go.go")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_markdown_lists/main.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/adamic_lists.go"): bridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-list-fixtures")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go bridge: %v\n%s", err, output)
	}
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	want := execute(t, nil, goBinary, cases, nativeCases, canonicalCases)
	clean(t, "Go list fixtures", want)
	if keep := os.Getenv("ADAMIC_MARKDOWNLISTS_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
		for _, name := range []string{"native.txt", "canonical.txt"} {
			value, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(keep, name), value)
		}
	}
	layoutDriver, err := filepath.Abs("testdata/document_go.go")
	if err != nil {
		t.Fatal(err)
	}
	layoutPath := filepath.Join(cohere, "cmd/adamic_markdown_doclayout/main.go")
	layoutOverlay, err := json.Marshal(map[string]any{"Replace": map[string]string{layoutPath: layoutDriver}})
	if err != nil {
		t.Fatal(err)
	}
	layoutOverlayPath := filepath.Join(dir, "layout.overlay.json")
	write(t, layoutOverlayPath, layoutOverlay)
	goLayout := filepath.Join(dir, "go-layout")
	layoutCommand := bounded(t, "go", "build", "-overlay="+layoutOverlayPath, "-o", goLayout, layoutPath)
	layoutCommand.Dir = cohere
	if output, err := layoutCommand.CombinedOutput(); err != nil {
		t.Fatalf("Go layout: %v\n%s", err, output)
	}
	goResult := execute(t, nil, goLayout, canonicalCases)
	clean(t, "Go document layout", goResult)
	equal(t, "Go document layout", goResult.stdout, want.stdout)
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, nativeCases)
	for _, side := range []struct {
		name   string
		result run
	}{{"source Node lists", onNode(t, main, nativeCases)}, {"backend lists", onJavaScriptBackend(t, program, nativeCases)}, {"native lists", answer}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			t.Fatalf("%s output byte %d in %s\ngot %q\nwant %q", side.name, offset, inputs[index].Name, actual[index], expected[index])
		}
	}
	if report := leaks(t, program, binary, nativeCases); report != "" {
		t.Fatal(report)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	script, err := filepath.Abs("testdata/document_library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	original := execute(t, nil, "node", script, fork, canonicalCases)
	clean(t, "original document printer", original)
	equal(t, "original document printer", original.stdout, want.stdout)
	markdownScript, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	sourceAnswers := auditResults(t, "original Markdown parser/layout", execute(t, nil, "node", markdownScript, fork, cases, "fork", "off-only"))
	if len(sourceAnswers) != len(inputs) {
		t.Fatal("original Markdown oracle lost a document")
	}
	encodeOutput := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	wantLines := strings.Split(strings.TrimSuffix(string(want.stdout), "\n"), "\n")
	for index, answer := range sourceAnswers {
		if answer.Name != inputs[index].Name || encodeOutput.Replace(answer.Off) != wantLines[index] {
			t.Fatalf("original Markdown parser/layout disagreed with Go in %s: got %q, Go encoded %q", inputs[index].Name, answer.Off, wantLines[index])
		}
	}
	t.Logf("original fork full Markdown parsing/layout off agrees on all %d source documents", len(inputs))

	mutations := []struct{ name, from, to string }{
		{"unordered marker", "frame.sibling % 2 === 0 ? '- ' : '* '", "frame.sibling % 2 === 0 ? '+ ' : '* '"},
		{"task box", "'[x] '", "'[X] '"},
		{"ordered cap", "999999999", "999999998"},
	}
	mutantFile := "lists.ts"
	if slice == "quotes" {
		mutantFile = "quotes.ts"
		mutations = []struct{ name, from, to string }{
			{"quote marker", "arena.text('> ', 2)", "arena.text('>> ', 3)"},
			{"quote alignment", "arena.align('> ',", "arena.align('',"},
			{"quote blank line", "!overlapping && !definitions", "!overlapping && definitions"},
		}
	}
	if slice == "tables" {
		mutantFile = "tables.ts"
		mutations = []struct{ name, from, to string }{
			{"table minimum", "widths.push(3)", "widths.push(4)"},
			{"table center", "Math.floor(spaces / 2)", "Math.ceil(spaces / 2)"},
			{"table alignment", "align === 'right' ? spaces", "align === 'right' ? 0"},
		}
	}
	if slice == "code" {
		mutantFile = "codeblocks.ts"
		mutations = []struct{ name, from, to string }{
			{"fence minimum", "Math.max(3, longest + 1)", "Math.max(4, longest + 1)"},
			{"fence longest", "longest + 1", "longest"},
			{"code indentation", "' '.repeat(4)", "' '.repeat(3)"},
		}
	}
	if slice == "html" {
		mutantFile = "htmlblocks.ts"
		mutations = []struct{ name, from, to string }{
			{"HTML root trim", "if(frame.rootLast)", "if(false)"},
			{"HTML comment line", "value.startsWith('<!--')", "!value.startsWith('<!--')"},
			{"HTML literal root", "arena.add('r', '', 0, [literal])", "arena.add('d', '', 0, [literal])"},
		}
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(scratch, "markdowninline"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"inline.ts", "classes.ts"} {
				content, err := os.ReadFile(filepath.Join("../markdowninline", name))
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(scratch, "markdowninline", name), content)
			}
			for _, name := range []string{"document.ts", "commandStack.ts", "codec.ts", "lists.ts", "quotes.ts", "tables.ts", "codeblocks.ts", "htmlblocks.ts"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == mutantFile {
					if strings.Count(string(content), mutation.from) != 1 {
						t.Fatal("mutation site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(t, filepath.Join(scratch, name), content)
			}
			content, err := os.ReadFile(main)
			if err != nil {
				t.Fatal(err)
			}
			content = []byte(strings.Replace(string(content), "../../markdowninline/inline.ts", "../markdowninline/inline.ts", 1))
			mutantMain := filepath.Join(scratch, "testdata/list_probe.ts")
			write(t, mutantMain, content)
			result := nativelyRun(t, lowered(t, mutantMain), nativeCases)
			clean(t, "native list mutant", result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("list mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("output-only native mutant caught by %s at byte %d", inputs[index].Name, offset)
		})
	}
	fast := filepath.Join(dir, "native-fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(root, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{{"Go document layout", goLayout, []string{canonicalCases}}, {"native block/layout component", fast, []string{nativeCases}}, {"source Node block/layout component", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, nativeCases}}, {"original Node document layout", "node", []string{script, fork, canonicalCases}}} {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			result := execute(t, nil, side.command, side.args...)
			elapsed += time.Since(start)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	data, err := os.ReadFile(nativeCases)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d physical files, %d generated, %d whole-document contexts; %d native list frames, %d composed word nodes; Go/source/native/backend/original doc bytes identical", files, len(inputs)-files, len(inputs), bytes.Count(data, []byte("\nL\t")), bytes.Count(data, []byte("\nW\t")))
	t.Logf("%d native HTML frames", bytes.Count(data, []byte("\nH\t")))
	t.Logf("%d native code frames", bytes.Count(data, []byte("\nC\t")))
	t.Logf("%d native table frames", bytes.Count(data, []byte("\nT\t")))
	t.Logf("%d native quote frames", bytes.Count(data, []byte("\nQ\t")))
}

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
	for _, mutation := range []struct{ name, from, to string }{
		{"unordered marker", "frame.sibling % 2 === 0 ? '- ' : '* '", "frame.sibling % 2 === 0 ? '+ ' : '* '"},
		{"task box", "'[x] '", "'[X] '"},
		{"ordered cap", "999999999", "999999998"},
	} {
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
			for _, name := range []string{"document.ts", "commandStack.ts", "codec.ts", "lists.ts"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "lists.ts" {
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
	}{{"Go document layout", goLayout, []string{canonicalCases}}, {"native list/layout component", fast, []string{nativeCases}}, {"source Node list/layout component", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, nativeCases}}, {"original Node document layout", "node", []string{script, fork, canonicalCases}}} {
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
}

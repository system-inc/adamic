package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Parallel execution reserves memory: all parsed ASTs, full six-pass results and sanitizer clones share a large working set.
func TestMarkdownASTPreprocessing(t *testing.T) {
	parallelMarkdownMemory(t, 3)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus, files := blockCorpus(t, root, "whitespace")
	type input struct {
		Name, Text string
		Mode       string
		TabWidth   int
	}
	inputs := []input{}
	for _, item := range corpus {
		inputs = append(inputs, input{Name: item.Name, Text: item.Text, TabWidth: 4})
	}
	for _, text := range []string{"[[a\n[[wiki]]", "[[a\ntext ]]", "[[a **b** c]]", "[[a [x](/y) ]]", "> a\n> b", "> > a\n> > b", "> a >b\n> > c", " > a\n > b", "> **a\n> b**", "a\n    ===\nb", "\tcode\n", "\n    code", "![a [b] c](/x)", "![a\\]b](/x)", "![a\\[b](/x)", "![😀中](/x)", "![x][a]\n\n[a]: /x", "1. a\n2. b\n", "11. a\n1.  b\n", "11. a\n1. b\n", "1.   a\n2. b\n", "1.\n\n2.\n", "1. a\n\n        code\n", "1. a\n\n   - b\n", "a\u00a0 b\n", "a\u000b b\n", "a\u0085 b\n", "a\u3000中\n", "[[a\n**b**\n[[wiki]]", "a *b\n c* d", "\ufeffa\r\nb\rc\n"} {
		for _, tab := range []int{1, 2, 3, 4, 8} {
			inputs = append(inputs, input{Name: fmt.Sprintf("generated/AST/%d/%s", tab, text), Text: text, TabWidth: tab})
		}
	}
	inputs = append(inputs, input{Name: "synthetic/adjacent_text", Text: "a b", Mode: "adjacent_text", TabWidth: 4})
	dir := t.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, item := range inputs {
		if err := encoder.Encode(item); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_ast/main.go")
	driver, err := filepath.Abs("testdata/ast_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/ast_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	facts, err := filepath.Abs("testdata/ast_facts.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/adamic_ast.go"): bridge, filepath.Join(cohere, "internal/format/markdown/adamic_ast_facts.go"): facts}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-ast")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go preprocess %v %s", err, output)
	}
	nativeCases := filepath.Join(dir, "native.txt")
	want := execute(t, nil, goBinary, cases, nativeCases)
	clean(t, "Go preprocessing", want)
	if keep := os.Getenv("ADAMIC_MARKDOWNAST_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(nativeCases)
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "native.txt"), data)
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
		write(t, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned bundle", installed, pinned)
	main, err := filepath.Abs("testdata/ast_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, nativeCases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"actual Go preprocessing from AST transport", execute(t, nil, goBinary, "--facts", nativeCases)},
		{"actual original AST preprocessing", execute(t, nil, "node", "testdata/ast_library.mjs", fork, nativeCases)}, {"native AST preprocessing", answer}, {"source Node", onNode(t, main, nativeCases)}, {"backend", onJavaScriptBackend(t, program, nativeCases)},
	} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			a := strings.Split(string(side.result.stdout), "\n")
			b := strings.Split(string(want.stdout), "\n")
			if len(a) != len(b) {
				t.Fatalf("%s result count %d want %d", side.name, len(a), len(b))
			}
			for i, line := range b {
				if a[i] != line {
					offset := firstDifference(a[i], line)
					t.Fatalf("%s mismatch %s byte%d; lengths %d/%d, nearby got %q want %q", side.name, inputs[i].Name, offset, len(a[i]), len(line), a[i][max(0, offset-100):min(len(a[i]), offset+100)], line[max(0, offset-100):min(len(line), offset+100)])
				}
			}
		}
	}
	if report := leaks(t, program, binary, nativeCases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"raw source", "astPreprocess.ts", "node.raw = this.source.slice(node.start, node.end)", "node.raw = node.value"},
		{"wiki risk", "astPreprocess.ts", "this.riskyPositions.has(this.arena.node(paragraph).position)", "!this.riskyPositions.has(this.arena.node(paragraph).position)"},
		{"list alignment", "astLists.ts", "if(firstStart % tabWidth === 0)", "if(firstStart % tabWidth !== 0)"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"astArena.ts", "astWalk.ts", "astSource.ts", "astLists.ts", "astPreprocess.ts", "astProtocol.ts", "splitText.ts", "textTokens.ts", "textClasses.ts", "emojiMatcher.ts", "widthTables.ts", "widthRuneRanges.ts", "codec.ts", "testdata/ast_probe.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == m.file {
					if strings.Count(string(data), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/ast_probe.ts"), nativeCases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("survived")
			}
			observed := strings.Split(string(result.stdout), "\n")
			for i, line := range strings.Split(string(want.stdout), "\n") {
				if i >= len(observed) {
					t.Logf("caught by missing AST result %d", i)
					break
				}
				if observed[i] != line {
					t.Logf("caught by %s at projected AST byte%d", inputs[i].Name, firstDifference(observed[i], line))
					break
				}
			}
		})
	}
	fast := filepath.Join(dir, "fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{{"actual Go preprocess", goBinary, []string{"--facts", nativeCases}}, {"native preprocess", fast, []string{nativeCases}}, {"actual original Node preprocess", "node", []string{"testdata/ast_library.mjs", fork, nativeCases}}} {
		started := time.Now()
		for i := 0; i < 3; i++ {
			result := execute(t, nil, side.command, side.args...)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("%s %.1f documents/s; three runs, startup, tree transport and output included; parsing excluded on all three sides", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
	}
	t.Logf("%d AST contexts: %d physical Markdown files, 4943 layout corpus plus five-tab-width generated cases; complete six-pass projected fields and shared position identity agree", len(inputs), files)
}

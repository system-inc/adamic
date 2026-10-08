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

// Parallel execution reserves memory: the exhaustive scalar corpus and sanitizer builds have a large peak working set.
func TestMarkdownTextSplitting(t *testing.T) {
	parallelMarkdownMemory(t, 8)
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	inputs, files := auditCorpus(t, root)
	original := append([]auditInput(nil), inputs...)
	for _, input := range original {
		for i, line := range strings.Split(input.Text, "\n") {
			inputs = append(inputs, auditInput{Name: fmt.Sprintf("%s:%d", input.Name, i+1), Text: line})
		}
	}
	for code := rune(0); code <= 0x10ffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("U+%06X", code), Text: string(code)})
	}
	atoms := []string{"a", "!", "中", "。", "한", "\u3000", "\t", "\n", " ", "\ufe0f", "\U000e0100", "😀", "\U00020000"}
	for _, a := range atoms {
		for _, b := range atoms {
			for _, c := range atoms {
				inputs = append(inputs, auditInput{Name: "sequence/" + a + b + c, Text: a + b + c})
			}
		}
	}
	for _, point := range []rune{0x2c7, 0x1100, 0x3000, 0x4e00, 0xac00, 0xff10, 0x20000, 0x31350} {
		for _, selector := range []rune{0xfe00, 0xfe0f, 0xe0100, 0xe01ef} {
			for _, suffix := range []string{"", "a", "中", "!", " ", "😀"} {
				text := "a" + string(point) + string(selector) + suffix
				inputs = append(inputs, auditInput{Name: "selector/" + text, Text: text})
			}
		}
	}
	dir := t.TempDir()
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_text/main.go")
	driver, e := filepath.Abs("testdata/text_go.go")
	if e != nil {
		t.Fatal(e)
	}
	overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/adamic_text.go"): filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_bridge.go")}})
	if e != nil {
		t.Fatal(e)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-text")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, e := command.CombinedOutput(); e != nil {
		t.Fatalf("Go splitText: %v\n%s", e, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go splitText", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, e := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	originalBundle, e := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	equal(t, "pinned original markdown bundle", installed, originalBundle)
	library := execute(t, nil, "node", "testdata/text_library.mjs", fork, cases)
	clean(t, "actual original splitText", library)
	equal(t, "Go versus actual fork splitText", library.stdout, want.stdout)
	main, e := filepath.Abs("testdata/text_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onJavaScriptBackend(t, program, cases)}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			a := strings.Split(string(side.result.stdout), "\n")
			b := strings.Split(string(want.stdout), "\n")
			for i := range b {
				if i >= len(a) || a[i] != b[i] {
					t.Fatalf("%s mismatch in %q", side.name, inputs[i].Name)
				}
			}
		}
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"fake whitespace", "textTokens.ts", "!between &&", "between &&"},
		{"Hangul kind", "splitText.ts", "tokens.word(text, 'k-letter', false", "tokens.word(text, 'cj-letter', false"},
		{"variation selector", "splitText.ts", "if(inRanges(selector, variationSelectorRanges))", "if(!inRanges(selector, variationSelectorRanges))"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
				t.Fatal(e)
			}
			for _, f := range []string{"splitText.ts", "textTokens.ts", "textClasses.ts", "emojiMatcher.ts", "widthTables.ts", "widthRuneRanges.ts", "codec.ts", "testdata/text_probe.ts"} {
				b, e := os.ReadFile(f)
				if e != nil {
					t.Fatal(e)
				}
				if f == m.file {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, f), b)
			}
			r := onNode(t, filepath.Join(scratch, "testdata/text_probe.ts"), cases)
			clean(t, m.name, r)
			if bytes.Equal(r.stdout, want.stdout) {
				t.Fatal("survived")
			}
			t.Log("caught by observable token fields")
		})
	}
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := filepath.Join(dir, "fast")
		if e := native.Build(native.C(program), fast, native.Options{}); e != nil {
			t.Fatal(e)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node splitText", "node", []string{"testdata/text_library.mjs", fork, cases}}} {
			started := time.Now()
			for i := 0; i < 3; i++ {
				r := execute(t, nil, side.command, side.args...)
				clean(t, side.name, r)
				equal(t, side.name, r.stdout, want.stdout)
			}
			t.Logf("%s %.1f texts/s, startup, decoding and output included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
		}
	} else {
		result := releaseRun(t, program, cases)
		clean(t, "release splitText", result)
		equal(t, "release splitText", result.stdout, want.stdout)
	}
	command = bounded(t, "go", "run", "./stage1/cohere/markdownblocks/tools/generate_classes", "-check", "-formatter", cohereFormatter(t))
	command.Dir = root
	if output, e := command.CombinedOutput(); e != nil {
		t.Fatalf("regeneration %v %s", e, output)
	}
	t.Logf("%d splitText sources: %d repository files, source lines, every Unicode scalar and generated sequences", len(inputs), files)
}

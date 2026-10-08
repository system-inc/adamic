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

func TestFrontMatterStage(t *testing.T) {
	parallelMarkdown(t)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := auditCorpus(t, root)
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(t, root, inputs, files, 1)
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	original := len(inputs)
	for _, start := range []string{"---", "+++"} {
		for _, language := range []string{"", "yaml", "toml", "json", "YAML", " yaml ", "\t yaml\t", "\ufeffyaml\ufeff", "\u0085yaml\u0085", "😀", "\u2000\u3000", "\r"} {
			for _, end := range []string{"---", "+++", "...", "missing"} {
				for _, value := range []string{"", "x: 1", "😀😀\n漢字", "a\r\nb\rc", "a\t\x00b", "x\n---\ny", "\n\n", "\u2028\u2029"} {
					for _, suffix := range []string{"", "\n# h\n", "suffix\n", "😀\n\n"} {
						text := start + language + "\n" + value + "\n" + end + suffix
						inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/front-matter/%d", len(inputs)-original), Text: text})
					}
				}
			}
		}
	}
	for _, code := range []rune{9, 10, 11, 12, 13, 32, 0x85, 0xa0, 0x1680, 0x180e, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff, 0xffff, 0x10000, 0x10ffff} {
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/front-matter/trim-%x", code), Text: "---" + string(code) + "yaml" + string(code) + "\nx\n..."})
	}
	for code := rune(0x2000); code <= 0x200b; code++ {
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/front-matter/trim-%x", code), Text: "---" + string(code) + "yaml" + string(code) + "\nx\n..."})
	}
	for _, text := range []string{"", "---", "+++", "---\n---", "---\n...", "---yaml\n...", "---json\n...", "\ufeff---\nx\n---", "\n---\nx\n---", "---\nx\n---suffix", "---\nx\n...suffix", "+++\nx\n+++suffix"} {
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/front-matter/short-%d", len(inputs)), Text: text})
	}
	// A long literal checks UTF-16 blanking across all Unicode scalar ranges without
	// constructing 65,536 almost-identical one-character tests.
	var unicode strings.Builder
	for code := rune(0); code <= 0xffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		unicode.WriteRune(code)
	}
	inputs = append(inputs, auditInput{Name: "generated/front-matter/unicode", Text: "---\n" + unicode.String() + "😀\U0010ffff\n---\n# h"})
	dir := t.TempDir()
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + encode.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_frontmatter/main.go")
	driver, err := filepath.Abs("testdata/frontmatter_go.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-parser")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go build: %v\n%s", err, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go front matter", want)
	main, err := filepath.Abs("testdata/frontmatter_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	got, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native front matter", got}, {"source Node front matter", onNode(t, main, cases)},
		{"JavaScript backend front matter", onJavaScriptBackend(t, program, cases)},
	} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want.stdout)
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	for _, name := range []string{"standalone.js", "plugins/markdown.js"} {
		original, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles", name))
		if err != nil {
			t.Fatal(err)
		}
		installed, err := os.ReadFile(filepath.Join(fork, name))
		if err != nil {
			t.Fatal(err)
		}
		equal(t, "pinned parser "+name, installed, original)
	}
	script, err := filepath.Abs("testdata/frontmatter_library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	library := execute(t, nil, "node", script, fork, cases)
	clean(t, "original parser front matter", library)
	equal(t, "original parser front matter", library.stdout, want.stdout)
	for _, mutation := range []struct{ name, from, to string }{
		{"YAML fallback delimiter", "language === 'yaml'", "language === 'toml'"},
		{"UTF-16 blank prefix", "' '.repeat(index - start)", "' '.repeat(index - start + 1)"},
		{"JavaScript BOM trimming", "code === 0xfeff", "code === 0xfffe"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			scratch := t.TempDir()
			for _, name := range []string{"frontmatter.ts", "codec.ts"} {
				source, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "frontmatter.ts" {
					if strings.Count(string(source), mutation.from) != 1 {
						t.Fatal("mutant must change exactly one site")
					}
					source = []byte(strings.Replace(string(source), mutation.from, mutation.to, 1))
				}
				write(t, filepath.Join(scratch, name), source)
			}
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(main)
			if err != nil {
				t.Fatal(err)
			}
			mutantMain := filepath.Join(scratch, "testdata/frontmatter_probe.ts")
			write(t, mutantMain, source)
			answer := onNode(t, mutantMain, cases)
			clean(t, "source Node output-only mutant", answer)
			if bytes.Equal(answer.stdout, want.stdout) {
				t.Fatal("mutant survived")
			}
			offset := firstDifference(string(answer.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("source Node output-only mutant caught at output byte %d by %s", offset, inputs[index].Name)
		})
	}
	fast := filepath.Join(dir, "native-fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"Go front-matter stage", goBinary, []string{cases}},
		{"native front-matter stage", fast, []string{cases}},
		{"source Node front-matter stage", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, cases}},
	} {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			answer := execute(t, nil, side.command, side.args...)
			elapsed += time.Since(start)
			clean(t, side.name, answer)
			equal(t, side.name, answer.stdout, want.stdout)
		}
		t.Logf("parser-stage throughput %s %.1f texts/s, 3 runs %.6fs; startup/protocol/I/O included, no Markdown layout measured", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	t.Logf("front-matter stage %d physical Markdown files and %d generated cases, %d total: Go/native/source Node/backend/pinned library byte parity, ASan/UBSan/leaks", files, len(inputs)-files, len(inputs))
}

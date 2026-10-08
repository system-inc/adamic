package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMarkdownParserPrefixes(t *testing.T) {
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
	for _, text := range []string{"", "\ufeff", ">", "> ", ">\t", " > a", "   > a", "    > a", "\t> a", "\t > a", ">> a", "> > a", "\ufeff> a", "\r\na", "\ra", "a\r", "a\r\nb", "😀\t> a"} {
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/prefix/short/%d", len(inputs)), Text: text})
	}
	units := []string{"a", " ", "\t", "\r", "\n", "\x00", "\ufeff", "😀", ">"}
	for _, a := range units {
		for _, b := range units {
			for _, c := range units {
				text := a + b + c
				inputs = append(inputs, auditInput{Name: "generated/prefix/" + text, Text: text})
			}
		}
	}
	for column := 0; column <= 12; column++ {
		for _, prefix := range []string{" ", "\t", "  \t", "\t ", "\ufeff"} {
			for _, suffix := range []string{"", "a", " ", "\t", "\r\n", "😀"} {
				text := strings.Repeat(" ", column) + prefix + ">" + suffix
				inputs = append(inputs, auditInput{Name: "generated/prefix/indent/" + text, Text: text})
			}
		}
	}
	// All scalar ranges in one source check UTF-16 chunks and tab columns together.
	var unicode strings.Builder
	for code := rune(0); code <= 0xffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		unicode.WriteRune(code)
	}
	inputs = append(inputs, auditInput{Name: "generated/prefix/unicode", Text: unicode.String() + "😀\t\U0010ffff"})
	dir := t.TempDir()
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + encode.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_prefix/main.go")
	mainGo, err := filepath.Abs("testdata/prefix_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/prefix_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: mainGo, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_prefix.go"): bridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-prefix")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go prefix bridge: %v\n%s", err, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "actual Go parser primitives", want)
	main, err := filepath.Abs("testdata/prefix_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"native parser prefixes", answer}, {"source Node parser prefixes", onNode(t, main, cases)}, {"backend parser prefixes", onJavaScriptBackend(t, program, cases)}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Fatalf("%s differs at byte %d in %q", side.name, offset, inputs[index].Name)
		}
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, mutation := range []struct{ name, file, from, to string }{
		{"CRLF chunk", "preprocess.ts", "code: -3", "code: -4"},
		{"tab virtual spaces", "preprocess.ts", "column < next", "column <= next"},
		{"quote continuation limit", "quotePrefix.ts", "index >= 3", "index >= 4"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"preprocess.ts", "quotePrefix.ts", "codec.ts"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == mutation.file {
					if strings.Count(string(content), mutation.from) != 1 {
						t.Fatal("mutant site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(t, filepath.Join(scratch, name), content)
			}
			content, err := os.ReadFile(main)
			if err != nil {
				t.Fatal(err)
			}
			mutantMain := filepath.Join(scratch, "testdata/prefix_probe.ts")
			write(t, mutantMain, content)
			result := onNode(t, mutantMain, cases)
			clean(t, "source Node output-only prefix mutant", result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("output-only source Node mutant caught in %q at byte %d", inputs[index].Name, offset)
		})
	}
	fast := filepath.Join(dir, "native-fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, side := range []struct {
		name, command string
		args          []string
	}{{"actual Go preprocessing/constructs", goBinary, []string{cases}}, {"native preprocessing/prefix recognition", fast, []string{cases}}, {"source Node preprocessing/prefix recognition", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, cases}}} {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			result := execute(t, nil, side.command, side.args...)
			elapsed += time.Since(start)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("parser primitive throughput %s %.1f texts/s, three runs %.6fs; source preprocessing/chunks/prefix recognition and startup/protocol/output included; Go also drains tokenizer and emits construct events, native does not", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	t.Logf("%d physical files, %d generated, %d sources; exact chunk boundaries/codes and %d quote prefix contexts match actual Go/source/native/backend; ASan/UBSan/leaks", files, len(inputs)-files, len(inputs), len(inputs)*8)
}

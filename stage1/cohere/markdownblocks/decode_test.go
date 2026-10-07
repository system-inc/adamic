package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Not parallel: the exhaustive numeric corpus and sanitizer run need a bounded peak working set.
func TestMarkdownSourceDecoding(t *testing.T) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := auditCorpus(t, root)
	source := filepath.Join(root, "cohere/internal/format/markdown/micromark/entities_generated.go")
	tree, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	entities := 0
	ast.Inspect(tree, func(node ast.Node) bool {
		pair, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		name, err := strconv.Unquote(pair.Key.(*ast.BasicLit).Value)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range []string{"&" + name + ";", "\\&" + name + ";", "&" + name, "&" + name + "X;", "&&" + name + ";;", "a&" + name + ";中"} {
			inputs = append(inputs, auditInput{Name: "entity/" + text, Text: text})
		}
		entities++
		return false
	})
	for _, text := range []string{"", "&#;", "&#x;", "&#X;", "&;", "&unknown;", "&#12345678;", "&#x1234567;", "&" + strings.Repeat("a", 32) + ";", "&amp;copy;", "\\\\&amp;", "\\&#65;", "&#8;", "&#0000000;", "&#x000000;", "&#x0000000;", "&#00000000;", "&#9999999;", "&#xFFFFFF;"} {
		inputs = append(inputs, auditInput{Name: "edge/" + text, Text: text})
	}
	for code := 0; code < 256; code++ {
		text := "a\\" + string(rune(code)) + "&amp;中"
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("escape/%d", code), Text: text})
	}
	compact := len(inputs)
	for code := 0; code <= 0x110020; code++ {
		for _, text := range []string{fmt.Sprintf("&#%d;", code), fmt.Sprintf("&#x%X;", code)} {
			inputs = append(inputs, auditInput{Name: fmt.Sprintf("numeric/%d", code), Text: text})
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
	mainPath := filepath.Join(cohere, "cmd/adamic_decode/main.go")
	driver, err := filepath.Abs("testdata/decode_go.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-decode")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go decoder %v %s", err, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go decoder", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	original, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned original bundle", installed, original)
	main, err := filepath.Abs("testdata/decode_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"actual original decoder", execute(t, nil, "node", "testdata/decode_library.mjs", fork, cases)}, {"native decoder", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onJavaScriptBackend(t, program, cases)},
	} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			for i, line := range expected {
				if i >= len(actual) || actual[i] != line {
					t.Fatalf("%s mismatch %d %q: got %q want %q", side.name, i, inputs[i].Name, actual[i], line)
				}
			}
		}
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	var small strings.Builder
	for _, input := range inputs[:compact] {
		small.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	mutantCases := filepath.Join(dir, "mutants.txt")
	write(t, mutantCases, []byte(small.String()))
	mutantWant := execute(t, nil, goBinary, mutantCases)
	clean(t, "Go mutant corpus", mutantWant)
	for _, m := range []struct{ name, from, to string }{
		{"numeric replacement", "code < 9 ||", "code < 8 ||"},
		{"hex digit bound", "hexadecimal ? 6 : 7", "hexadecimal ? 7 : 7"},
		{"escape punctuation", "code >= 33 && code <= 47", "code >= 34 && code <= 47"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"decodeString.ts", "upperEntityNames.ts", "upperEntityValues.ts", "lowerEntityNames.ts", "lowerEntityValues.ts", "codec.ts", "testdata/decode_probe.ts"} {
				b, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "decodeString.ts" {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), b)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/decode_probe.ts"), mutantCases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, mutantWant.stdout) {
				t.Fatal("survived")
			}
			t.Log("caught by decoded bytes")
		})
	}
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := filepath.Join(dir, "fast")
		if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node", "node", []string{"testdata/decode_library.mjs", fork, cases}}} {
			started := time.Now()
			for i := 0; i < 3; i++ {
				result := execute(t, nil, side.command, side.args...)
				clean(t, side.name, result)
				equal(t, side.name, result.stdout, want.stdout)
			}
			t.Logf("%s %.1f texts/s, startup, decoding, serialization and stdout included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
		}
	} else {
		result := releaseRun(t, program, cases)
		clean(t, "release decoder", result)
		equal(t, "release decoder", result.stdout, want.stdout)
	}
	command = bounded(t, "go", "run", "./stage1/cohere/markdownblocks/tools/generate_entities", "-check", "-formatter", cohereFormatter(t))
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("regeneration %v %s", err, output)
	}
	t.Logf("%d source strings: %d repository files, %d named entities in six contexts, 256 escape units, all numeric code points in decimal/hex plus boundaries", len(inputs), files, entities)
}

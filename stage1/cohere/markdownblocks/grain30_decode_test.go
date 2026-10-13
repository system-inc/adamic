package markdownblocks

import (
	"bytes"
	"fmt"
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

func grainDecodeInputs(t *testing.T) (string, []auditInput, int, int, int) {
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
	return root, inputs, files, entities, compact
}

func runGrainDecode(t *testing.T, shard int) {
	products := grainReady(t, "decode")
	if shard == 11 {
		generator := grainGenerator(t, "decode")
		formatter := grainTool(t, "formatter")
		stop := grainOwn(t)
		defer stop()
		grainRegenerate(t, "decode", generator, formatter)
		grainPlanted(t)
		return
	}
	stop := grainOwn(t)
	defer stop()
	root, inputs, files, entities, compact := grainDecodeInputs(t)
	if shard >= grainCorpusShards {
		inputs = inputs[:compact]
	}
	if shard < grainCorpusShards {
		inputs = grainSelect(inputs, shard)
		files = (files + grainCorpusShards - 1 - shard) / grainCorpusShards
		if len(inputs) == 0 {
			grainPlanted(t)
			t.Log("empty corpus group")
			return
		}
	}
	grainPlanted(t)

	dir := t.TempDir()
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goBinary

	want := grainExecute(t, nil, goBinary, cases)
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
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, cases)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"actual original decoder", grainExecute(t, nil, "node", "testdata/decode_library.mjs", fork, cases)}, {"native decoder", answer}, {"source Node", grainNode(t, main, cases)}, {"backend", grainNode(t, products.backend, cases)},
		} {
			clean(t, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				actual := strings.Split(string(side.result.stdout), "\n")
				expected := strings.Split(string(want.stdout), "\n")
				for i, line := range expected {
					if i >= len(actual) || actual[i] != line {
						t.Fatalf("%s mismatch %d %q: got %q want %q", side.name, i, inputs[i].Name, actual[i], line)
					}
				}
			}
		}
		if report := grainLeaks(t, products, cases); report != "" {
			t.Fatal(report)
		}
	}
	mutantCases, mutantWant := cases, want

	for mutationIndex, m := range []struct{ name, from, to string }{
		{"numeric replacement", "code < 9 ||", "code < 8 ||"},
		{"hex digit bound", "hexadecimal ? 6 : 7", "hexadecimal ? 7 : 7"},
		{"escape punctuation", "code >= 33 && code <= 47", "code >= 34 && code <= 47"},
	} {
		if shard != grainCorpusShards+mutationIndex {
			continue
		}
		func(t *testing.T) {
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
			result := grainNode(t, filepath.Join(scratch, "testdata/decode_probe.ts"), mutantCases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, mutantWant.stdout) {
				t.Fatal("survived")
			}
			t.Log("caught by decoded bytes")
		}(t)
	}
	if shard < grainCorpusShards {

		if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
			fast := products.release

			for _, side := range []struct {
				name, command string
				args          []string
			}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node", "node", []string{"testdata/decode_library.mjs", fork, cases}}} {
				started := time.Now()
				for i := 0; i < 3; i++ {
					result := grainExecute(t, nil, side.command, side.args...)
					clean(t, side.name, result)
					equal(t, side.name, result.stdout, want.stdout)
				}
				t.Logf("%s %.1f texts/s, startup, decoding, serialization and stdout included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
			}
		} else {
			result := grainExecute(t, nil, products.release, cases)
			clean(t, "release decoder", result)
			equal(t, "release decoder", result.stdout, want.stdout)
		}
		t.Logf("%d source strings: %d repository files, %d named entities in six contexts, 256 escape units, all numeric code points in decimal/hex plus boundaries", len(inputs), files, entities)

	}

}

func TestMarkdownSourceDecoding_000(t *testing.T) { t.Parallel(); runGrainDecode(t, 0) }

func TestMarkdownSourceDecoding_001(t *testing.T) { t.Parallel(); runGrainDecode(t, 1) }

func TestMarkdownSourceDecoding_002(t *testing.T) { t.Parallel(); runGrainDecode(t, 2) }

func TestMarkdownSourceDecoding_003(t *testing.T) { t.Parallel(); runGrainDecode(t, 3) }

func TestMarkdownSourceDecoding_004(t *testing.T) { t.Parallel(); runGrainDecode(t, 4) }

func TestMarkdownSourceDecoding_005(t *testing.T) { t.Parallel(); runGrainDecode(t, 5) }

func TestMarkdownSourceDecoding_006(t *testing.T) { t.Parallel(); runGrainDecode(t, 6) }

func TestMarkdownSourceDecoding_007(t *testing.T) { t.Parallel(); runGrainDecode(t, 7) }

func TestMarkdownSourceDecoding_008(t *testing.T) { t.Parallel(); runGrainDecode(t, 8) }

func TestMarkdownSourceDecoding_009(t *testing.T) { t.Parallel(); runGrainDecode(t, 9) }

func TestMarkdownSourceDecoding_010(t *testing.T) { t.Parallel(); runGrainDecode(t, 10) }

func TestMarkdownSourceDecoding_011(t *testing.T) { t.Parallel(); runGrainDecode(t, 11) }

func TestMarkdownSourceDecodingUnion(t *testing.T) {
	t.Parallel()
	_, inputs, _, _, _ := grainDecodeInputs(t)
	grainUnion(t, inputs, 12, "TestMarkdownSourceDecoding", "runGrainDecode")
}

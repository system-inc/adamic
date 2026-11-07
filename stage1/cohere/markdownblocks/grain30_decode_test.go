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

func grainDecodeInputs(test *testing.T) (string, []auditInput, int, int, int) {
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	inputs, files := auditCorpus(test, root)
	source := filepath.Join(root, "cohere/internal/format/markdown/micromark/entities_generated.go")
	tree, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
	if err != nil {
		test.Fatal(err)
	}
	entities := 0
	ast.Inspect(tree, func(node ast.Node) bool {
		pair, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		name, err := strconv.Unquote(pair.Key.(*ast.BasicLit).Value)
		if err != nil {
			test.Fatal(err)
		}
		for _, text := range []string{
			"&" + name + ";",
			"\\&" + name + ";",
			"&" + name,
			"&" + name + "X;",
			"&&" + name + ";;",
			"a&" + name + ";中",
		} {
			inputs = append(inputs, auditInput{Name: "entity/" + text, Text: text})
		}
		entities++
		return false
	})
	for _, text := range []string{
		"",
		"&#;",
		"&#x;",
		"&#X;",
		"&;",
		"&unknown;",
		"&#12345678;",
		"&#x1234567;",
		"&" + strings.Repeat("a", 32) + ";",
		"&amp;copy;",
		"\\\\&amp;",
		"\\&#65;",
		"&#8;",
		"&#0000000;",
		"&#x000000;",
		"&#x0000000;",
		"&#00000000;",
		"&#9999999;",
		"&#xFFFFFF;",
	} {
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

func runGrainDecode(test *testing.T, shard int) {
	products := grainReady(test, "decode")
	if shard == 11 {
		generator := grainGenerator(test, "decode")
		formatter := grainTool(test, "formatter")
		stop := grainOwn(test)
		defer stop()
		grainRegenerate(test, "decode", generator, formatter)
		grainPlanted(test)
		return
	}
	stop := grainOwn(test)
	defer stop()
	root, inputs, files, entities, compact := grainDecodeInputs(test)
	if shard >= grainCorpusShards {
		inputs = inputs[:compact]
	}
	if shard < grainCorpusShards {
		inputs = grainSelect(inputs, shard)
		files = (files + grainCorpusShards - 1 - shard) / grainCorpusShards
		if len(inputs) == 0 {
			grainPlanted(test)
			test.Log("empty corpus group")
			return
		}
	}
	grainPlanted(test)

	dir := test.TempDir()
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(test, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goBinary

	want := grainExecute(test, nil, goBinary, cases)
	clean(test, "Go decoder", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	original, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if err != nil {
		test.Fatal(err)
	}
	installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if err != nil {
		test.Fatal(err)
	}
	equal(test, "pinned original bundle", installed, original)
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, cases)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"actual original decoder", grainExecute(test, nil, "node", "testdata/decode_library.mjs", fork, cases)},
			{"native decoder", answer},
			{"source Node", grainNode(test, main, cases)},
			{"backend", grainNode(test, products.backend, cases)},
		} {
			clean(test, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				actual := strings.Split(string(side.result.stdout), "\n")
				expected := strings.Split(string(want.stdout), "\n")
				for i, line := range expected {
					if i >= len(actual) || actual[i] != line {
						test.Fatalf("%s mismatch %d %q: got %q want %q", side.name, i, inputs[i].Name, actual[i], line)
					}
				}
			}
		}
		if report := grainLeaks(test, products, cases); report != "" {
			test.Fatal(report)
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
		func(test *testing.T) {
			scratch := test.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				test.Fatal(err)
			}
			for _, file := range []string{
				"decodeString.ts",
				"upperEntityNames.ts",
				"upperEntityValues.ts",
				"lowerEntityNames.ts",
				"lowerEntityValues.ts",
				"codec.ts",
				"testdata/decode_probe.ts",
			} {
				b, err := os.ReadFile(file)
				if err != nil {
					test.Fatal(err)
				}
				if file == "decodeString.ts" {
					if strings.Count(string(b), m.from) != 1 {
						test.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(test, filepath.Join(scratch, file), b)
			}
			result := grainNode(test, filepath.Join(scratch, "testdata/decode_probe.ts"), mutantCases)
			clean(test, m.name, result)
			if bytes.Equal(result.stdout, mutantWant.stdout) {
				test.Fatal("survived")
			}
			test.Log("caught by decoded bytes")
		}(test)
	}
	if shard < grainCorpusShards {

		if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
			fast := products.release

			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Go", goBinary, []string{cases}},
				{"native", fast, []string{cases}},
				{"actual original Node", "node", []string{"testdata/decode_library.mjs", fork, cases}},
			} {
				started := time.Now()
				for i := 0; i < 3; i++ {
					result := grainExecute(test, nil, side.command, side.args...)
					clean(test, side.name, result)
					equal(test, side.name, result.stdout, want.stdout)
				}
				test.Logf(
					"%s %.1f texts/s, startup, decoding, serialization and stdout included",
					side.name,
					float64(3*len(inputs))/time.Since(started).Seconds(),
				)
			}
		} else {
			result := grainExecute(test, nil, products.release, cases)
			clean(test, "release decoder", result)
			equal(test, "release decoder", result.stdout, want.stdout)
		}
		test.Logf(
			"%d source strings: %d repository files, %d named entities in six contexts, 256 escape units, all numeric code points in decimal/hex plus boundaries",
			len(inputs),
			files,
			entities,
		)

	}

}

func TestMarkdownSourceDecoding_000(test *testing.T) { test.Parallel(); runGrainDecode(test, 0) }

func TestMarkdownSourceDecoding_001(test *testing.T) { test.Parallel(); runGrainDecode(test, 1) }

func TestMarkdownSourceDecoding_002(test *testing.T) { test.Parallel(); runGrainDecode(test, 2) }

func TestMarkdownSourceDecoding_003(test *testing.T) { test.Parallel(); runGrainDecode(test, 3) }

func TestMarkdownSourceDecoding_004(test *testing.T) { test.Parallel(); runGrainDecode(test, 4) }

func TestMarkdownSourceDecoding_005(test *testing.T) { test.Parallel(); runGrainDecode(test, 5) }

func TestMarkdownSourceDecoding_006(test *testing.T) { test.Parallel(); runGrainDecode(test, 6) }

func TestMarkdownSourceDecoding_007(test *testing.T) { test.Parallel(); runGrainDecode(test, 7) }

func TestMarkdownSourceDecoding_008(test *testing.T) { test.Parallel(); runGrainDecode(test, 8) }

func TestMarkdownSourceDecoding_009(test *testing.T) { test.Parallel(); runGrainDecode(test, 9) }

func TestMarkdownSourceDecoding_010(test *testing.T) { test.Parallel(); runGrainDecode(test, 10) }

func TestMarkdownSourceDecoding_011(test *testing.T) { test.Parallel(); runGrainDecode(test, 11) }

func TestMarkdownSourceDecodingUnion(test *testing.T) {
	test.Parallel()
	_, inputs, _, _, _ := grainDecodeInputs(test)
	grainUnion(test, inputs, 12, "TestMarkdownSourceDecoding", "runGrainDecode")
}

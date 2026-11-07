package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// Parallel execution reserves memory: all parsed ASTs, full six-pass results and sanitizer clones share a large working
// set.
// Not parallel: ADAMIC_MARKDOWNAST_KEEP export directory and shared markdownMemory budget; existing helper controls parallel execution.
func TestMarkdownASTPreprocessing(test *testing.T) {
	parallelMarkdownMemory(test, 3)
	fixture := markdownASTPrepare(test)
	inputs, files, goBinary, nativeCases, fork, want := fixture.inputs, fixture.files, fixture.goBinary, fixture.nativeCases, fixture.fork, fixture.want
	if keep := os.Getenv("ADAMIC_MARKDOWNAST_KEEP"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		data, readError := os.ReadFile(nativeCases)
		if readError != nil {
			test.Fatal(readError)
		}
		write(test, filepath.Join(keep, "native.txt"), data)
		write(test, filepath.Join(keep, "want.txt"), want.stdout)
		write(test, filepath.Join(keep, "cases.jsonl"), fixture.batch)
	}
	// The lowered port and its sanitized and release builds are products (TestProduct_MarkdownAST*), not built
	// in this test each run.
	lowered := pilotLoweredProduct(test, "ast")
	products := grainProducts{
		backend:   filepath.Join(lowered, "program.mjs"),
		sanitized: pilotNativeProductFromLowered(test, "ast", lowered, true),
		release:   pilotNativeProductFromLowered(test, "ast", lowered, false),
	}
	main, absolutePathError := filepath.Abs("testdata/ast_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	answer := execute(test, environment, products.sanitized, nativeCases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"actual Go preprocessing from AST transport", execute(test, nil, goBinary, "--facts", nativeCases)},
		{
			"actual original AST preprocessing",
			execute(test, nil, "node", "testdata/ast_library.mjs", fork, nativeCases),
		},
		{"native AST preprocessing", answer},
		{"source Node", onNode(test, main, nativeCases)},
		{"backend", onNode(test, products.backend, nativeCases)},
	} {
		clean(test, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			a := strings.Split(string(side.result.stdout), "\n")
			b := strings.Split(string(want.stdout), "\n")
			if len(a) != len(b) {
				test.Fatalf("%s result count %d want %d", side.name, len(a), len(b))
			}
			for index, line := range b {
				if a[index] != line {
					offset := firstDifference(a[index], line)
					test.Fatalf(
						"%s mismatch %s byte%d; lengths %d/%d, nearby got %q want %q",
						side.name,
						inputs[index].Name,
						offset,
						len(a[index]),
						len(line),
						a[index][max(0, offset-100):min(len(a[index]), offset+100)],
						line[max(0, offset-100):min(len(line), offset+100)],
					)
				}
			}
		}
	}
	if report := grainLeaks(test, products, nativeCases); report != "" {
		test.Fatal(report)
	}
	for _, testCase := range []struct{ name, file, from, to string }{
		{
			"raw source",
			"astPreprocess.ts",
			"node.raw = this.source.slice(node.start, node.end)",
			"node.raw = node.value",
		},
		{
			"wiki risk",
			"astPreprocess.ts",
			"this.riskyPositions.has(this.arena.node(paragraph).position)",
			"!this.riskyPositions.has(this.arena.node(paragraph).position)",
		},
		{"list alignment", "astLists.ts", "if(firstStart % tabWidth === 0)", "if(firstStart % tabWidth !== 0)"},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, file := range []string{
				"astArena.ts",
				"astWalk.ts",
				"astSource.ts",
				"astLists.ts",
				"astPreprocess.ts",
				"astProtocol.ts",
				"splitText.ts",
				"textTokens.ts",
				"textClasses.ts",
				"emojiMatcher.ts",
				"widthTables.ts",
				"widthRuneRanges.ts",
				"codec.ts",
				"testdata/ast_probe.ts",
			} {
				data, readError := os.ReadFile(file)
				if readError != nil {
					test.Fatal(readError)
				}
				if file == testCase.file {
					if strings.Count(string(data), testCase.from) != 1 {
						test.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), testCase.from, testCase.to, 1))
				}
				write(test, filepath.Join(scratch, file), data)
			}
			result := onNode(test, filepath.Join(scratch, "testdata/ast_probe.ts"), nativeCases)
			clean(test, testCase.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("survived")
			}
			observed := strings.Split(string(result.stdout), "\n")
			for index, line := range strings.Split(string(want.stdout), "\n") {
				if index >= len(observed) {
					test.Logf("caught by missing AST result %d", index)
					break
				}
				if observed[index] != line {
					test.Logf(
						"caught by %s at projected AST byte%d",
						inputs[index].Name,
						firstDifference(observed[index], line),
					)
					break
				}
			}
		})
	}
	test.Logf(
		"%d AST contexts: %d physical Markdown files, 4943 layout corpus plus five-tab-width generated cases; complete six-pass projected fields and shared position identity agree",
		len(inputs),
		files,
	)
}

// The three-run timings of Go, the release port and the original preprocessing are their own test over the same
// contexts, every timed run still checked against Go's bytes, so neither test carries both the comparisons and the
// timings.
// Not parallel: parallelMarkdownMemory runs it in parallel under the shared markdownMemory budget, except while a KEEP export is requested; existing helper controls parallel execution.
func TestMarkdownASTPreprocessingThroughput(test *testing.T) {
	parallelMarkdownMemory(test, 3)
	fixture := markdownASTPrepare(test)
	inputs, goBinary, nativeCases, fork, want := fixture.inputs, fixture.goBinary, fixture.nativeCases, fixture.fork, fixture.want
	fast := pilotNativeProductFromLowered(test, "ast", pilotLoweredProduct(test, "ast"), false)
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go preprocess", goBinary, []string{"--facts", nativeCases}},
		{"native preprocess", fast, []string{nativeCases}},
		{"actual original Node preprocess", "node", []string{"testdata/ast_library.mjs", fork, nativeCases}},
	} {
		started := time.Now()
		for index := 0; index < 3; index++ {
			result := execute(test, nil, side.command, side.args...)
			clean(test, side.name, result)
			equal(test, side.name, result.stdout, want.stdout)
		}
		test.Logf(
			"%s %.1f documents/s; three runs, startup, tree transport and output included; parsing excluded on all three sides",
			side.name,
			float64(3*len(inputs))/time.Since(started).Seconds(),
		)
	}
}

// markdownASTInput is one AST preprocessing context as the Go driver and the probes read it.
type markdownASTInput struct {
	Name, Text string
	Mode       string
	TabWidth   int
}

// markdownASTFixture is what both AST preprocessing tests start from: the contexts, Go's projected bytes for them,
// the AST transport Go wrote for the other sides, and the pinned fork.
type markdownASTFixture struct {
	inputs                      []markdownASTInput
	files                       int
	batch                       []byte
	goBinary, nativeCases, fork string
	want                        run
}

func markdownASTPrepare(test *testing.T) markdownASTFixture {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus, files := blockCorpus(test, root, "whitespace")
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(test, root, corpus, files, 1)
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
	}
	corpus, files = selectedMarkdownInputs(corpus, files, selection)
	inputs := []markdownASTInput{}
	for _, item := range corpus {
		inputs = append(inputs, markdownASTInput{Name: item.Name, Text: item.Text, TabWidth: 4})
	}
	for _, text := range []string{
		"[[a\n[[wiki]]",
		"[[a\ntext ]]",
		"[[a **b** c]]",
		"[[a [x](/y) ]]",
		"> a\n> b",
		"> > a\n> > b",
		"> a >b\n> > c",
		" > a\n > b",
		"> **a\n> b**",
		"a\n    ===\nb",
		"\tcode\n",
		"\n    code",
		"![a [b] c](/x)",
		"![a\\]b](/x)",
		"![a\\[b](/x)",
		"![😀中](/x)",
		"![x][a]\n\n[a]: /x",
		"1. a\n2. b\n",
		"11. a\n1.  b\n",
		"11. a\n1. b\n",
		"1.   a\n2. b\n",
		"1.\n\n2.\n",
		"1. a\n\n        code\n",
		"1. a\n\n   - b\n",
		"a\u00a0 b\n",
		"a\u000b b\n",
		"a\u0085 b\n",
		"a\u3000中\n",
		"[[a\n**b**\n[[wiki]]",
		"a *b\n c* d",
		"\ufeffa\r\nb\rc\n",
	} {
		for _, tab := range []int{1, 2, 3, 4, 8} {
			inputs = append(
				inputs,
				markdownASTInput{Name: fmt.Sprintf("generated/AST/%d/%s", tab, text), Text: text, TabWidth: tab},
			)
		}
	}
	inputs = append(
		inputs,
		markdownASTInput{Name: "synthetic/adjacent_text", Text: "a b", Mode: "adjacent_text", TabWidth: 4},
	)
	dir := test.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, item := range inputs {
		if encodeError := encoder.Encode(item); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(test, cases, batch.Bytes())
	cohere := filepath.Join(root, "cohere")
	goBinary := astGoOracle(test)
	nativeCases := filepath.Join(dir, "native.txt")
	want := execute(test, nil, goBinary, cases, nativeCases)
	clean(test, "Go preprocessing", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, readError := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	pinned, readError := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	equal(test, "pinned bundle", installed, pinned)
	return markdownASTFixture{
		inputs:      inputs,
		files:       files,
		batch:       batch.Bytes(),
		goBinary:    goBinary,
		nativeCases: nativeCases,
		fork:        fork,
		want:        want,
	}
}

// astGoOracle is the Go side, cohere with the driver and bridge its overlay lays over it, a product built ahead
// (#5qykzj5).
func astGoOracle(test *testing.T) string { test.Helper(); return oracles.ASTGoOracle(test) }

func TestProduct_MarkdownASTLowered(test *testing.T) {
	test.Parallel()
	pilotLoweredProduct(test, "ast")
}

func TestProduct_MarkdownASTNativeSanitized(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "ast", true)
}

func TestProduct_MarkdownASTNativeRelease(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "ast", false)
}

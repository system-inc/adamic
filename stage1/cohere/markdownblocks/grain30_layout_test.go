package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var grainLayoutOnce sync.Once
var grainLayoutProducts leafCompositionProducts

func grainLayoutReady(test *testing.T) leafCompositionProducts {
	grainLayoutOnce.Do(func() {
		started := time.Now()
		root := markdownLayoutProductRoot(test)
		products := prepareLeafCompositionGo(test, root)
		prepareLeafCompositionLowered(test, &products)
		products.sanitized = leafCompositionNative(test, products.source, true)
		products.release = leafCompositionNative(test, products.source, false)
		grainLayoutProducts = products
		test.Logf("grain setup %.3fs", time.Since(started).Seconds())
	})
	return grainLayoutProducts
}
func runGrainLayout(test *testing.T, slice string, shard int) {
	products := grainLayoutReady(test)
	stop := grainOwn(test)
	defer stop()
	grainPlanted(test)
	all, files := blockCorpus(test, products.root, "whitespace")
	if shard < grainCorpusShards {
		inputs := grainSelect(all, shard)
		if len(inputs) == 0 {
			test.Log("empty corpus group")
			return
		}
		physical := 0
		for index := shard; index < files; index += grainCorpusShards {
			physical++
		}
		fixture := buildLeafCompositionFixture(test, inputs, physical, products)
		expected := bytes.Split(bytes.TrimSuffix(fixture.want, []byte("\n")), []byte("\n"))
		encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
		for index, input := range inputs {
			for side, output := range fixture.outputs {
				if grainDifferenceError := grainDifference(
					output[index],
					expected[index],
				); grainDifferenceError != nil {
					test.Fatalf("%s %s: %v", side, input.Name, grainDifferenceError)
				}
			}
			answer := fixture.answers[index]
			if answer.Name != input.Name || encode.Replace(answer.Off) != string(expected[index]) {
				test.Fatalf("original Markdown parser/layout disagreed in %s", input.Name)
			}
		}

		// Preserve the optional repeated throughput checks on each child's partition.
		if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Go document layout", products.goLayout, []string{fixture.canonicalCases}},
				{"native block/layout component", products.release, []string{fixture.nativeCases}},
				{"source Node block/layout component", "node", []string{
					"--disable-warning=ExperimentalWarning",
					filepath.Join(products.root, "oracle/node.mjs"),
					products.main,
					fixture.nativeCases,
				}},
				{
					"original Node document layout",
					"node",
					[]string{products.script, products.fork, fixture.canonicalCases},
				},
			} {
				started := time.Now()
				for round := 0; round < 3; round++ {
					result := grainExecute(test, nil, side.command, side.args...)
					clean(test, side.name, result)
					if grainDifferenceError := grainDifference(
						result.stdout,
						fixture.want,
					); grainDifferenceError != nil {
						test.Fatalf("%s throughput: %v", side.name, grainDifferenceError)
					}
				}
				test.Logf(
					"component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded",
					side.name,
					float64(3*len(inputs))/time.Since(started).Seconds(),
					time.Since(started).Seconds(),
				)
			}
		}
		test.Logf("cases=%d", len(inputs))
		return
	}
	encoded, want := grainLayoutMutantFixtures(test, products, all)
	nativeCases := filepath.Join(test.TempDir(), "native.txt")
	write(test, nativeCases, encoded)
	fixture := &layoutFixture{main: products.main, mutantNativeCases: nativeCases, mutantInputs: all, mutantWant: want}
	grainLayoutMutation(test, fixture, slice, shard-grainCorpusShards)
}

// grainLayoutMutantFixtures holds what every mutant shard of every slice compares against, made once per process:
// the whole whitespace corpus Go encoded for the port, its width slots poisoned, and Go's layout of it. Every slice's
// mutant shards read the same corpus, so each ran this same Go oracle over it before.
var grainLayoutMutantAnswers struct {
	once                sync.Once
	nativeCases, stdout []byte
}

func grainLayoutMutantFixtures(test *testing.T, products leafCompositionProducts, all []auditInput) ([]byte, []byte) {
	test.Helper()
	grainLayoutMutantAnswers.once.Do(func() {
		dir := test.TempDir()
		var batch bytes.Buffer
		for _, input := range all {
			if encodeError := json.NewEncoder(&batch).Encode(input); encodeError != nil {
				test.Fatal(encodeError)
			}
		}
		cases := filepath.Join(dir, "cases.jsonl")
		write(test, cases, batch.Bytes())
		nativeCases := filepath.Join(dir, "native.txt")
		want := grainExecute(test, nil, products.goList, cases, nativeCases, filepath.Join(dir, "canonical.txt"))
		clean(test, "Go mutant fixtures", want)
		poisonLeafCompositionCases(test, nativeCases)
		// Go writes the port's cases into the first caller's temporary directory, gone when that test ends, so their
		// bytes are kept.
		encoded, readError := os.ReadFile(nativeCases)
		if readError != nil {
			test.Fatal(readError)
		}
		grainLayoutMutantAnswers.stdout = want.stdout
		grainLayoutMutantAnswers.nativeCases = encoded
	})
	if grainLayoutMutantAnswers.nativeCases == nil {
		test.Fatal("the Go mutant fixtures failed earlier in this process")
	}
	return grainLayoutMutantAnswers.nativeCases, grainLayoutMutantAnswers.stdout
}
func grainLayoutMutation(test *testing.T, fixture *layoutFixture, slice string, selected int) {
	main := fixture.main
	mutations := []struct{ name, from, to string }{
		{"unordered marker", "frame.sibling % 2 === 0 ? '- ' : '* '", "frame.sibling % 2 === 0 ? '+ ' : '* '"},
		{"task box", "'[x] '", "'[X] '"},
		{"ordered cap", "999999999", "999999998"},
	}
	mutantFile := "lists.ts"
	if slice == "quotes" {
		mutantFile = "quotes.ts"
		mutations = []struct{ name, from, to string }{
			{"quote marker", "arena.text('> ')", "arena.text('>> ')"},
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
	if slice == "html" || slice == "structure" {
		mutantFile = "htmlblocks.ts"
		mutations = []struct{ name, from, to string }{
			{"HTML root trim", "if(frame.rootLast)", "if(false)"},
			{"HTML comment line", "value.startsWith('<!--')", "!value.startsWith('<!--')"},
			{"HTML literal root", "arena.add('r', '', 0, [literal])", "arena.add('d', '', 0, [literal])"},
		}
	}
	if slice == "structure" {
		mutantFile = "structure.ts"
		mutations = []struct{ name, from, to string }{
			{"heading depth", "'#'.repeat(depth)", "'#'.repeat(depth + 1)"},
			{"sentence separator", "parts.push(child.doc)", "parts.push(arena.text(''))"},
			{
				"paragraph fill tail",
				"for(let index = 1; index < node.children.length",
				"for(let index = 2; index < node.children.length",
			},
		}
	}
	if slice == "root" {
		mutantFile = "root.ts"
		mutations = []struct{ name, from, to string }{
			{"root double line", "!overlapping && !definitions", "overlapping && !definitions"},
			{"ignored source span", "source.slice(child.startOffset, child.endOffset)", "'mutant ignored text'"},
			{"ignore range closing comment", "arena.text(last.value)", "arena.text('')"},
		}
	}
	if slice == "leaves" {
		mutantFile = "leaves.ts"
		mutations = []struct{ name, from, to string }{
			{
				"strong delimiter",
				"frame.kind === 'strong'\n                ? '**'",
				"frame.kind === 'strong'\n                ? '*'",
			},
			{
				"reference collapse",
				"frame.referenceType === 'collapsed'\n                  ? '[]'",
				"frame.referenceType === 'collapsed'\n                  ? ''",
			},
			{"footnote indentation", "arena.align('    ',", "arena.align('   ',"},
		}
	}
	if slice == "whitespace" {
		mutantFile = "whitespace.ts"
		mutations = []struct{ name, from, to string }{
			{
				"preserved newline",
				"if(proseWrap === 'preserve' && frame.value === '\\n')",
				"if(proseWrap === 'never' && frame.value === '\\n')",
			},
			{"space removal", "return arena.text(canBeSpace ? ' ' : '')", "return arena.text('')"},
			{"ordered syntax", "value.endsWith('.')", "value.endsWith(':')"},
		}
	}
	for index, mutation := range mutations {
		if index != selected {
			continue
		}
		func(test *testing.T) {
			nativeCases, inputs := fixture.mutantNativeCases, fixture.mutantInputs
			want := run{stdout: fixture.mutantWant}
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if createDirectoryError := os.Mkdir(
				filepath.Join(scratch, "markdowninline"),
				0755,
			); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, name := range []string{"inline.ts", "classes.ts"} {
				content, readError := os.ReadFile(filepath.Join("../markdowninline", name))
				if readError != nil {
					test.Fatal(readError)
				}
				write(test, filepath.Join(scratch, "markdowninline", name), content)
			}
			for _, name := range []string{
				"document.ts",
				"commandStack.ts",
				"codec.ts",
				"lists.ts",
				"quotes.ts",
				"tables.ts",
				"codeblocks.ts",
				"htmlblocks.ts",
				"width.ts",
				"widthTables.ts",
				"widthRuneRanges.ts",
				"emojiMatcher.ts",
				"structure.ts",
				"root.ts",
				"leaves.ts",
				"children.ts",
				"splitText.ts",
				"textTokens.ts",
				"textClasses.ts",
				"preservedLabel.ts",
				"whitespace.ts",
				"whitespaceCodec.ts",
			} {
				content, readError := os.ReadFile(name)
				if readError != nil {
					test.Fatal(readError)
				}
				if name == "leaves.ts" {
					content = []byte(strings.Replace(
						string(content),
						"../markdowninline/inline.ts",
						"./markdowninline/inline.ts",
						1,
					))
				}
				if name == mutantFile {
					if strings.Count(string(content), mutation.from) != 1 {
						test.Fatal("mutation site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(test, filepath.Join(scratch, name), content)
			}
			content, readError := os.ReadFile(main)
			if readError != nil {
				test.Fatal(readError)
			}
			content = []byte(
				strings.Replace(string(content), "../../markdowninline/inline.ts", "../markdowninline/inline.ts", 1),
			)
			mutantMain := filepath.Join(scratch, "testdata/list_probe.ts")
			write(test, mutantMain, content)
			result := grainNode(test, mutantMain, nativeCases)
			clean(test, "source Node layout mutant", result)
			// One package canary: this copy imports every composed layout printer.
			if slice == "lists" && mutation.name == "unordered marker" {
				nativeResult := nativeMutant(test, lowered(test, mutantMain), nativeCases)
				clean(test, "native layout canary", nativeResult)
				equal(test, "edited native canary equals source Node", nativeResult.stdout, result.stdout)
			}
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("list mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			test.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
		}(test)
	}
}

func TestMarkdownCodeBlockLayout_000(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 0)
}

func TestMarkdownCodeBlockLayout_001(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 1)
}

func TestMarkdownCodeBlockLayout_002(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 2)
}

func TestMarkdownCodeBlockLayout_003(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 3)
}

func TestMarkdownCodeBlockLayout_004(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 4)
}

func TestMarkdownCodeBlockLayout_005(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 5)
}

func TestMarkdownCodeBlockLayout_006(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 6)
}

func TestMarkdownCodeBlockLayout_007(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 7)
}

func TestMarkdownCodeBlockLayout_008(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 8)
}

func TestMarkdownCodeBlockLayout_009(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 9)
}

func TestMarkdownCodeBlockLayout_010(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "code", 10)
}

func TestMarkdownCodeBlockLayoutUnion(test *testing.T) {
	test.Parallel()
	all, _ := blockCorpus(test, markdownLayoutProductRoot(test), "whitespace")
	grainUnion(test, all, 11, "TestMarkdownCodeBlockLayout", "runGrainLayout")
}

func TestMarkdownHTMLBlockLayout_000(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 0)
}

func TestMarkdownHTMLBlockLayout_001(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 1)
}

func TestMarkdownHTMLBlockLayout_002(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 2)
}

func TestMarkdownHTMLBlockLayout_003(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 3)
}

func TestMarkdownHTMLBlockLayout_004(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 4)
}

func TestMarkdownHTMLBlockLayout_005(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 5)
}

func TestMarkdownHTMLBlockLayout_006(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 6)
}

func TestMarkdownHTMLBlockLayout_007(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 7)
}

func TestMarkdownHTMLBlockLayout_008(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 8)
}

func TestMarkdownHTMLBlockLayout_009(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 9)
}

func TestMarkdownHTMLBlockLayout_010(test *testing.T) {
	test.Parallel()
	runGrainLayout(test, "html", 10)
}

func TestMarkdownHTMLBlockLayoutUnion(test *testing.T) {
	test.Parallel()
	all, _ := blockCorpus(test, markdownLayoutProductRoot(test), "whitespace")
	grainUnion(test, all, 11, "TestMarkdownHTMLBlockLayout", "runGrainLayout")
}

func TestMarkdownRootLayout_000(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 0) }

func TestMarkdownRootLayout_001(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 1) }

func TestMarkdownRootLayout_002(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 2) }

func TestMarkdownRootLayout_003(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 3) }

func TestMarkdownRootLayout_004(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 4) }

func TestMarkdownRootLayout_005(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 5) }

func TestMarkdownRootLayout_006(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 6) }

func TestMarkdownRootLayout_007(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 7) }

func TestMarkdownRootLayout_008(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 8) }

func TestMarkdownRootLayout_009(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 9) }

func TestMarkdownRootLayout_010(test *testing.T) { test.Parallel(); runGrainLayout(test, "root", 10) }

func TestMarkdownRootLayoutUnion(test *testing.T) {
	test.Parallel()
	all, _ := blockCorpus(test, markdownLayoutProductRoot(test), "whitespace")
	grainUnion(test, all, 11, "TestMarkdownRootLayout", "runGrainLayout")
}

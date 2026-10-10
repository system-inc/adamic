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

func grainLayoutReady(t *testing.T) leafCompositionProducts {
	grainLayoutOnce.Do(func() {
		started := time.Now()
		root := markdownLayoutProductRoot(t)
		p := prepareLeafCompositionGo(t, root)
		prepareLeafCompositionLowered(t, &p)
		p.sanitized = leafCompositionNative(t, p.source, true)
		p.release = leafCompositionNative(t, p.source, false)
		grainLayoutProducts = p
		t.Logf("grain setup %.3fs", time.Since(started).Seconds())
	})
	return grainLayoutProducts
}
func runGrainLayout(t *testing.T, slice string, shard int) {
	products := grainLayoutReady(t)
	stop := grainOwn(t)
	defer stop()
	grainPlanted(t)
	all, files := blockCorpus(t, products.root, "whitespace")
	if shard < grainCorpusShards {
		inputs := grainSelect(all, shard)
		if len(inputs) == 0 {
			t.Log("empty corpus group")
			return
		}
		physical := 0
		for index := shard; index < files; index += grainCorpusShards {
			physical++
		}
		fixture := buildLeafCompositionFixture(t, inputs, physical, products)
		expected := bytes.Split(bytes.TrimSuffix(fixture.want, []byte("\n")), []byte("\n"))
		encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
		for index, input := range inputs {
			for side, output := range fixture.outputs {
				if err := grainDifference(output[index], expected[index]); err != nil {
					t.Fatalf("%s %s: %v", side, input.Name, err)
				}
			}
			answer := fixture.answers[index]
			if answer.Name != input.Name || encode.Replace(answer.Off) != string(expected[index]) {
				t.Fatalf("original Markdown parser/layout disagreed in %s", input.Name)
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
				{"source Node block/layout component", "node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(products.root, "oracle/node.mjs"), products.main, fixture.nativeCases}},
				{"original Node document layout", "node", []string{products.script, products.fork, fixture.canonicalCases}},
			} {
				started := time.Now()
				for round := 0; round < 3; round++ {
					result := grainExecute(t, nil, side.command, side.args...)
					clean(t, side.name, result)
					if err := grainDifference(result.stdout, fixture.want); err != nil {
						t.Fatalf("%s throughput: %v", side.name, err)
					}
				}
				t.Logf("component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded", side.name, float64(3*len(inputs))/time.Since(started).Seconds(), time.Since(started).Seconds())
			}
		}
		t.Logf("cases=%d", len(inputs))
		return
	}
	dir := t.TempDir()
	var batch bytes.Buffer
	for _, input := range all {
		if err := json.NewEncoder(&batch).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	nativeCases := filepath.Join(dir, "native.txt")
	want := grainExecute(t, nil, products.goList, cases, nativeCases, filepath.Join(dir, "canonical.txt"))
	clean(t, "Go mutant fixtures", want)
	poisonLeafCompositionCases(t, nativeCases)
	fixture := &layoutFixture{main: products.main, mutantNativeCases: nativeCases, mutantInputs: all, mutantWant: want.stdout}
	grainLayoutMutation(t, fixture, slice, shard-grainCorpusShards)
}
func grainLayoutMutation(t *testing.T, fixture *layoutFixture, slice string, selected int) {
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
			{"paragraph fill tail", "for(let index = 1; index < node.children.length", "for(let index = 2; index < node.children.length"},
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
			{"strong delimiter", "frame.kind === 'strong'\n                ? '**'", "frame.kind === 'strong'\n                ? '*'"},
			{"reference collapse", "frame.referenceType === 'collapsed'\n                  ? '[]'", "frame.referenceType === 'collapsed'\n                  ? ''"},
			{"footnote indentation", "arena.align('    ',", "arena.align('   ',"},
		}
	}
	if slice == "whitespace" {
		mutantFile = "whitespace.ts"
		mutations = []struct{ name, from, to string }{
			{"preserved newline", "if(proseWrap === 'preserve' && frame.value === '\\n')", "if(proseWrap === 'never' && frame.value === '\\n')"},
			{"space removal", "return arena.text(canBeSpace ? ' ' : '')", "return arena.text('')"},
			{"ordered syntax", "value.endsWith('.')", "value.endsWith(':')"},
		}
	}
	for index, mutation := range mutations {
		if index != selected {
			continue
		}
		func(t *testing.T) {
			nativeCases, inputs := fixture.mutantNativeCases, fixture.mutantInputs
			want := run{stdout: fixture.mutantWant}
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
			for _, name := range []string{"document.ts", "commandStack.ts", "codec.ts", "lists.ts", "quotes.ts", "tables.ts", "codeblocks.ts", "htmlblocks.ts", "width.ts", "widthTables.ts", "widthRuneRanges.ts", "emojiMatcher.ts", "structure.ts", "root.ts", "leaves.ts", "children.ts", "splitText.ts", "textTokens.ts", "textClasses.ts", "preservedLabel.ts", "whitespace.ts", "whitespaceCodec.ts"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "leaves.ts" {
					content = []byte(strings.Replace(string(content), "../markdowninline/inline.ts", "./markdowninline/inline.ts", 1))
				}
				if name == mutantFile {
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
			result := grainNode(t, mutantMain, nativeCases)
			clean(t, "source Node layout mutant", result)
			// One package canary: this copy imports every composed layout printer.
			if slice == "lists" && mutation.name == "unordered marker" {
				nativeResult := nativeMutant(t, lowered(t, mutantMain), nativeCases)
				clean(t, "native layout canary", nativeResult)
				equal(t, "edited native canary equals source Node", nativeResult.stdout, result.stdout)
			}
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("list mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			t.Logf("output-only source Node mutant caught by %q at byte %d", inputs[index].Name, offset)
		}(t)
	}
}

func TestMarkdownCodeBlockLayout_000(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 0) }

func TestMarkdownCodeBlockLayout_001(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 1) }

func TestMarkdownCodeBlockLayout_002(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 2) }

func TestMarkdownCodeBlockLayout_003(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 3) }

func TestMarkdownCodeBlockLayout_004(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 4) }

func TestMarkdownCodeBlockLayout_005(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 5) }

func TestMarkdownCodeBlockLayout_006(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 6) }

func TestMarkdownCodeBlockLayout_007(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 7) }

func TestMarkdownCodeBlockLayout_008(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 8) }

func TestMarkdownCodeBlockLayout_009(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 9) }

func TestMarkdownCodeBlockLayout_010(t *testing.T) { t.Parallel(); runGrainLayout(t, "code", 10) }

func TestMarkdownCodeBlockLayoutUnion(t *testing.T) {
	t.Parallel()
	all, _ := blockCorpus(t, markdownLayoutProductRoot(t), "whitespace")
	grainUnion(t, all, 11, "TestMarkdownCodeBlockLayout", "runGrainLayout")
}

func TestMarkdownHTMLBlockLayout_000(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 0) }

func TestMarkdownHTMLBlockLayout_001(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 1) }

func TestMarkdownHTMLBlockLayout_002(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 2) }

func TestMarkdownHTMLBlockLayout_003(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 3) }

func TestMarkdownHTMLBlockLayout_004(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 4) }

func TestMarkdownHTMLBlockLayout_005(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 5) }

func TestMarkdownHTMLBlockLayout_006(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 6) }

func TestMarkdownHTMLBlockLayout_007(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 7) }

func TestMarkdownHTMLBlockLayout_008(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 8) }

func TestMarkdownHTMLBlockLayout_009(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 9) }

func TestMarkdownHTMLBlockLayout_010(t *testing.T) { t.Parallel(); runGrainLayout(t, "html", 10) }

func TestMarkdownHTMLBlockLayoutUnion(t *testing.T) {
	t.Parallel()
	all, _ := blockCorpus(t, markdownLayoutProductRoot(t), "whitespace")
	grainUnion(t, all, 11, "TestMarkdownHTMLBlockLayout", "runGrainLayout")
}

func TestMarkdownRootLayout_000(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 0) }

func TestMarkdownRootLayout_001(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 1) }

func TestMarkdownRootLayout_002(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 2) }

func TestMarkdownRootLayout_003(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 3) }

func TestMarkdownRootLayout_004(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 4) }

func TestMarkdownRootLayout_005(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 5) }

func TestMarkdownRootLayout_006(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 6) }

func TestMarkdownRootLayout_007(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 7) }

func TestMarkdownRootLayout_008(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 8) }

func TestMarkdownRootLayout_009(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 9) }

func TestMarkdownRootLayout_010(t *testing.T) { t.Parallel(); runGrainLayout(t, "root", 10) }

func TestMarkdownRootLayoutUnion(t *testing.T) {
	t.Parallel()
	all, _ := blockCorpus(t, markdownLayoutProductRoot(t), "whitespace")
	grainUnion(t, all, 11, "TestMarkdownRootLayout", "runGrainLayout")
}

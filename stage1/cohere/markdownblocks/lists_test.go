package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func TestMarkdownListLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "lists")
}

func TestMarkdownQuoteLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "quotes")
}

func TestMarkdownTableLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "tables")
}

func TestMarkdownCodeBlockLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "code")
}

func TestMarkdownHTMLBlockLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "html")
}

// Bound the large fixture streams below so layouts fit both memory and the gate deadline.
var layoutSlots = make(chan struct{}, 2)

type layoutFixture struct {
	selection                                                                  gatesample.Selection
	mutantNativeCases                                                          string
	mutantInputs                                                               []auditInput
	mutantWant                                                                 []byte
	root, directory, nativeCases, canonicalCases, main, fork, goLayout, script string
	inputs                                                                     []auditInput
	files                                                                      int
	want                                                                       []byte
	program                                                                    *ir.Program
}

var layoutOnce sync.Once
var completeLayout *layoutFixture

func fullLayoutFixture(t *testing.T) *layoutFixture {
	t.Helper()
	layoutOnce.Do(func() { completeLayout = buildLayoutFixture(t) })
	if completeLayout == nil {
		t.Fatal("the complete layout baseline failed")
	}
	return completeLayout
}

func TestMarkdownWhitespaceLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "whitespace")
}

func TestMarkdownLeafComposition(t *testing.T) {
	parallelMarkdown(t)
	testMarkdownLeafShards(t)
}

func TestMarkdownRootLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "root")
}

func TestMarkdownStructureLayout(t *testing.T) {
	parallelMarkdown(t)
	testBlockLayout(t, "structure")
}

func buildLayoutFixture(t *testing.T) *layoutFixture {
	var workers sync.WaitGroup
	defer workers.Wait()
	// This last cumulative corpus contains every earlier layout milestone's inputs.
	const slice = "whitespace"
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := blockCorpus(t, root, slice)
	fullInputs := inputs
	selection := selectMarkdownFiles(t, root, inputs, files, markdownCorpusStride)
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	dir := filepath.Join(artifactDirectory, "layout-fixtures")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	fullCases := cases
	if selection.Sample {
		fullCases = filepath.Join(dir, "mutant-cases.jsonl")
		var fullBatch bytes.Buffer
		fullEncoder := json.NewEncoder(&fullBatch)
		for _, input := range fullInputs {
			if err := fullEncoder.Encode(input); err != nil {
				t.Fatal(err)
			}
		}
		write(t, fullCases, fullBatch.Bytes())
	}
	bridge, err := filepath.Abs("testdata/list_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs("testdata/list_go.go")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_markdown_lists/main.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/adamic_lists.go"): bridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-list-fixtures")
	nativeCases, canonicalCases := filepath.Join(dir, "native.txt"), filepath.Join(dir, "canonical.txt")
	mutantNativeCases := nativeCases
	var mutantWant run
	listTask := startFixtureTask(&workers, func() (run, error) {
		command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			return run{}, fmt.Errorf("Go bridge: %v\n%s", err, output)
		}
		if selection.Sample {
			mutantNativeCases = filepath.Join(dir, "mutant-native.txt")
			var err error
			mutantWant, err = executeResult(t, nil, goBinary, fullCases, mutantNativeCases, filepath.Join(dir, "mutant-canonical.txt"))
			if err != nil {
				return run{}, err
			}
		}
		return executeResult(t, nil, goBinary, cases, nativeCases, canonicalCases)
	})
	layoutDriver, err := filepath.Abs("testdata/document_go.go")
	if err != nil {
		t.Fatal(err)
	}
	layoutPath := filepath.Join(cohere, "cmd/adamic_markdown_doclayout/main.go")
	layoutOverlay, err := json.Marshal(map[string]any{"Replace": map[string]string{layoutPath: layoutDriver}})
	if err != nil {
		t.Fatal(err)
	}
	layoutOverlayPath := filepath.Join(dir, "layout.overlay.json")
	write(t, layoutOverlayPath, layoutOverlay)
	goLayout := filepath.Join(dir, "go-layout")
	docBuild := startFixtureTask(&workers, func() (struct{}, error) {
		command := bounded(t, "go", "build", "-overlay="+layoutOverlayPath, "-o", goLayout, layoutPath)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			return struct{}{}, fmt.Errorf("Go layout: %v\n%s", err, output)
		}
		return struct{}{}, nil
	})
	main, err := filepath.Abs("testdata/list_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	lowerTask := startFixtureTask(&workers, func() (*ir.Program, error) { return loweredResult(main) })
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	script, err := filepath.Abs("testdata/document_library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	markdownScript, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	libraryTask := startFixtureTask(&workers, func() (run, error) {
		return executeResult(t, nil, "node", markdownScript, fork, cases, "fork", "off-only")
	})
	program := lowerTask.await(t)
	// Emit once; both native modes still compile separately with their original flags.
	source := native.C(program)
	sanitizedBuild := startFixtureTask(&workers, func() (string, error) {
		return nativeBinaryResult(source, native.Options{Sanitize: true})
	})
	releaseBuild := startFixtureTask(&workers, func() (string, error) { return nativeBinaryResult(source, native.Options{}) })
	want := listTask.await(t)
	clean(t, "Go list fixtures", want)
	if selection.Sample {
		clean(t, "full Go mutant fixtures", mutantWant)
	} else {
		mutantWant = want
	}
	// Width slots are retained only for the canonical Go/fork doc protocol.
	// Poison every text-width field in the native protocol to prove no oracle service remains.
	protocolPaths := []string{nativeCases}
	if selection.Sample {
		protocolPaths = append(protocolPaths, mutantNativeCases)
	}
	for _, protocolPath := range protocolPaths {
		raw, err := os.ReadFile(protocolPath)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(raw), "\n")
		for index, line := range lines {
			fields := strings.Split(line, "\t")
			switch fields[0] {
			case "D":
				if fields[1] == "t" {
					fields[3] = "-777"
				}
			case "W":
				fields[3] = "-777"
			case "H":
				fields[2] = ""
			case "C":
				fields[4], fields[5], fields[6] = "-777", "-777", ""
			case "T":
				fields[2] = ""
				rows := strings.Split(fields[3], ":")
				for r, row := range rows {
					cells := strings.Split(row, ";")
					for c, cell := range cells {
						pair := strings.Split(cell, ",")
						if len(pair) == 2 {
							pair[1] = "-777"
							cells[c] = strings.Join(pair, ",")
						}
					}
					rows[r] = strings.Join(cells, ";")
				}
				fields[3] = strings.Join(rows, ":")
			}
			lines[index] = strings.Join(fields, "\t")
		}
		write(t, protocolPath, []byte(strings.Join(lines, "\n")))
	}
	if keep := os.Getenv("ADAMIC_MARKDOWNLISTS_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
		for _, name := range []string{"native.txt", "canonical.txt"} {
			value, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(keep, name), value)
		}
	}
	docTask := startFixtureTask(&workers, func() (run, error) {
		if _, err := docBuild.result(); err != nil {
			return run{}, err
		}
		return executeResult(t, nil, goLayout, canonicalCases)
	})
	sourceTask := startFixtureTask(&workers, func() (run, error) { return onNodeResult(t, main, nativeCases) })
	backendPath := filepath.Join(dir, "program.mjs")
	write(t, backendPath, []byte(javascript.JavaScript(program)))
	backendTask := startFixtureTask(&workers, func() (run, error) { return onNodeResult(t, backendPath, nativeCases) })
	nativeTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := sanitizedBuild.result()
		if err != nil {
			return run{}, err
		}
		var environment []string
		if runtime.GOOS == "linux" {
			environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
		}
		return executeResult(t, environment, binary, nativeCases)
	})
	releaseTask := startFixtureTask(&workers, func() (run, error) {
		binary, err := releaseBuild.result()
		if err != nil {
			return run{}, err
		}
		return executeResult(t, nil, binary, nativeCases)
	})
	originalTask := startFixtureTask(&workers, func() (run, error) { return executeResult(t, nil, "node", script, fork, canonicalCases) })
	goResult := docTask.await(t)
	clean(t, "Go document layout", goResult)
	equal(t, "Go document layout", goResult.stdout, want.stdout)
	answer := nativeTask.await(t)
	binary := sanitizedBuild.await(t)
	for _, side := range []struct {
		name   string
		result run
	}{{"source Node lists", sourceTask.await(t)}, {"backend lists", backendTask.await(t)}, {"native lists", answer}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			actual := strings.Split(string(side.result.stdout), "\n")
			expected := strings.Split(string(want.stdout), "\n")
			t.Fatalf("%s output byte %d in %s\ngot %q\nwant %q", side.name, offset, inputs[index].Name, actual[index], expected[index])
		}
	}
	if report := leaks(t, program, binary, nativeCases); report != "" {
		t.Fatal(report)
	}
	original := originalTask.await(t)
	clean(t, "original document printer", original)
	equal(t, "original document printer", original.stdout, want.stdout)
	sourceAnswers := auditResults(t, "original Markdown parser/layout", libraryTask.await(t))
	if len(sourceAnswers) != len(inputs) {
		t.Fatal("original Markdown oracle lost a document")
	}
	encodeOutput := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	wantLines := strings.Split(strings.TrimSuffix(string(want.stdout), "\n"), "\n")
	for index, answer := range sourceAnswers {
		if answer.Name != inputs[index].Name || encodeOutput.Replace(answer.Off) != wantLines[index] {
			t.Fatalf("original Markdown parser/layout disagreed with Go in %s: got %q, Go encoded %q", inputs[index].Name, answer.Off, wantLines[index])
		}
	}
	t.Logf("original fork full Markdown parsing/layout off agrees on all %d source documents", len(inputs))
	release := releaseTask.await(t)
	clean(t, "release block layout", release)
	equal(t, "release block layout", release.stdout, want.stdout)
	return &layoutFixture{
		selection: selection, mutantNativeCases: mutantNativeCases, mutantInputs: fullInputs, mutantWant: mutantWant.stdout,
		root: root, directory: dir, nativeCases: nativeCases, canonicalCases: canonicalCases,
		main: main, fork: fork, goLayout: goLayout, script: script,
		inputs: inputs, files: files, want: want.stdout, program: program,
	}
}

func testBlockLayout(t *testing.T, slice string) {
	fixture := fullLayoutFixture(t)
	if fixture.selection.Sample {
		t.Log(fixture.selection.Log(t.Name()))
	}
	layoutSlots <- struct{}{}
	defer func() { <-layoutSlots }()
	root, nativeCases, canonicalCases := fixture.root, fixture.nativeCases, fixture.canonicalCases
	main, fork, goLayout, script := fixture.main, fixture.fork, fixture.goLayout, fixture.script
	inputs, files, program := fixture.inputs, fixture.files, fixture.program
	want := run{stdout: fixture.want}
	if slice == "whitespace" {
		testWhitespacePolicy(t, fixture.mutantNativeCases, fork)
	}

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
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
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
			result := onNode(t, mutantMain, nativeCases)
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
		})
	}
	// Repeat timings only on request; all corpus and mutant comparisons ran above.
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := nativeBinary(t, native.C(program), false)
		runner, err := filepath.Abs(filepath.Join(root, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go document layout", goLayout, []string{canonicalCases}}, {"native block/layout component", fast, []string{nativeCases}}, {"source Node block/layout component", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, nativeCases}}, {"original Node document layout", "node", []string{script, fork, canonicalCases}}} {
			var elapsed time.Duration
			for round := 0; round < 3; round++ {
				start := time.Now()
				result := execute(t, nil, side.command, side.args...)
				elapsed += time.Since(start)
				clean(t, side.name, result)
				equal(t, side.name, result.stdout, want.stdout)
			}
			t.Logf("component throughput %s %.1f documents/s, three runs %.6fs; fixture decoding/output/startup included, Markdown parsing and Go fixture generation excluded", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
		}
	}
	data, err := os.ReadFile(nativeCases)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d physical files, %d generated, %d whole-document contexts; %d native list frames, %d composed word nodes; Go/source/native/backend/original doc bytes identical", files, len(inputs)-files, len(inputs), bytes.Count(data, []byte("\nL\t")), bytes.Count(data, []byte("\nW\t")))
	t.Logf("%d native HTML frames", bytes.Count(data, []byte("\nH\t")))
	t.Logf("%d native code frames", bytes.Count(data, []byte("\nC\t")))
	t.Logf("%d native table frames", bytes.Count(data, []byte("\nT\t")))
	t.Logf("%d native quote frames", bytes.Count(data, []byte("\nQ\t")))
}

func blockCorpus(t *testing.T, root, slice string) ([]auditInput, int) {
	t.Helper()
	inputs, files := auditCorpus(t, root)
	for _, marker := range []string{"-", "*", "+", "0.", "1.", "1)", "9.", "10)", "99.", "999999999."} {
		for _, spacing := range []string{" ", "  ", "   ", "    ", "     ", "\t"} {
			for _, task := range []string{"", "[ ] ", "[x] ", "[X] "} {
				for _, body := range []string{"a *b _c_* `x`", "a\n\n    b\n\n    - c", "a\n\n    > q\n    > r", "a\n<div>\nx\n</div>", "\n\n\n    code"} {
					text := marker + spacing + task + body + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/list-layout/" + text, Text: text})
				}
			}
		}
	}
	for _, text := range []string{
		"0. a\n1. b\n", "0. a\n1. b\n1. c\n", "2. a\n1. b\n7. c\n", "999999999. a\n1. b\n",
		"- a\n* b\n+ c\n- d\n", "1. a\n2) b\n3. c\n", "-\n\n\n    x\n", "1.\n\n\n        x\n",
		"- [x] a\n\n    - [ ] b\n      c\n", "- a\n\n  ```js\n  let x = 1;\n  ```\n",
		"<!-- prettier-ignore -->\n*  a\n+   b\n", "> <!-- prettier-ignore -->\n> *  a\n>\n",
		"- a\n\n  |a|b|\n  |-|-|\n  |x|y|\n", "- a\n\n  ### h\n\n  ---\n",
		"- <div>\n  x\n  </div>\n", "- a\n\n  [x]: https://example.test\n  [y]: /y\n",
		"- a\n\n  [^x]: footnote\n\n- [^x]\n", "1.    中😀\n1.    _a*b*_\n",
	} {
		inputs = append(inputs, auditInput{Name: "generated/list-layout/edge/" + text, Text: text})
	}
	if slice != "lists" {
		for _, prefix := range []string{">", "> ", " > ", "  >  ", "   >\t", "> > ", ">> ", "> > > ", ">\t> "} {
			for _, body := range []string{"", "text", "a\nb", "a\n\nb", "# h\n\na", "a\n---", "- a\n- b", "1. a\n\n   - b", "- [x] a", "```js\nlet x=1;\n```", "    x\n    y", "<div>\nx\n</div>", "a\n<div>\nx", "[a]: /x\n[b]: /y", "| a | b |\n| - | - |\n| x | y |", "<!-- prettier-ignore -->\n+    a", "中😀 _a*b*_"} {
				text := prefix + strings.ReplaceAll(body, "\n", "\n"+prefix) + "\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/quote-layout/" + text, Text: text})
			}
		}
	}
	if slice != "lists" && slice != "quotes" {
		for _, align := range []string{"---", ":--", "--:", ":-:"} {
			for _, cell := range []string{"", "a", "abcde", "abcdef", "中😀", "*a _b_*", "`a|b`", "a\\|b", "[a](/b)", "<em>a</em>"} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := "| a | b |\n| " + align + " | " + align + " |\n| " + cell + " | x |\n| long text | y |\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/table-layout/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"a | b\n- | -\nc\nd | e | f\n", "a | b\n:- | -:\n | \n", "| a |\n| - |\n| x | y |\n", "| 😀 | 中 |\n| :-: | -: |\n| é | 👨‍👩‍👧‍👦 |\n"} {
			inputs = append(inputs, auditInput{Name: "generated/table-layout/edge/" + text, Text: text})
		}
	}
	if slice == "code" || (slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace") {
		for _, marker := range []string{"```", "````", "~~~~", "```````"} {
			for _, info := range []string{"", "js", "json", "yaml", "toml", "css", "html", "text title=foo", "x {#id .class}"} {
				for _, body := range []string{"", "a", "a\nb", "\n\nx\n", "a  \nb\t", "`a`", "```", "~~~~", "中😀"} {
					for _, prefix := range []string{"", "> ", "- ", "> - "} {
						text := marker + info + "\n" + body + "\n" + marker + "\n"
						lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
						text = prefix + lines[0] + "\n"
						continuation := prefix
						if strings.HasSuffix(prefix, "- ") {
							continuation = strings.TrimSuffix(prefix, "- ") + "  "
						}
						for _, line := range lines[1:] {
							text += continuation + line + "\n"
						}
						inputs = append(inputs, auditInput{Corpus: true, Name: "generated/code-layout/" + text, Text: text})
					}
				}
			}
		}
		for _, indent := range []string{"    ", "\t", "     ", "  \t", "\t "} {
			for _, body := range []string{"a", "a  \nb\t", "\n\nx\n", "中😀"} {
				text := indent + strings.ReplaceAll(body, "\n", "\n"+indent) + "\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/code-layout/indent/" + text, Text: text})
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, body := range []string{"<div>\nx  \n</div>", "<script>\nx  \n</script>", "<style>\nx\t\n</style>", "<pre>\nx  \n</pre>", "<!-- a  \nb\t -->", "<!-->", "<!--->", "<!--a-->", "<?xml\nx  \n?>", "<!DOCTYPE html>", "<![CDATA[\nx  \n]]>", "<table>\n<tr>\nx\n</tr>\n</table>", "<x-a a='b'>\nx\n</x-a>", "a <em>\nx  \n</em> b", "<div>\n\n# h\n\n</div>", "<!-- prettier-ignore -->\n<div>  \nx\n</div>", "<div>中😀</div>"} {
			for _, prefix := range []string{"", "> ", "> > ", "- ", "> - "} {
				for _, ending := range []string{"", "  ", "\t", "\u00a0", "\u2000", "\ufeff"} {
					text := body + ending + "\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/html-layout/" + text, Text: text})
				}
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, cell := range []string{"中", "Ａ", "α", "é", "©", "©️", "👩🏽‍⚕️", "👨🏻‍❤️‍💋‍👨🏿", "🧑🏿‍🦽‍➡️", "🏳️‍🌈", "🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f"} {
			for _, align := range []string{"---", ":--", "--:", ":-:"} {
				text := "| " + cell + " | a |\n| " + align + " | --- |\n| a | " + cell + " |\n"
				inputs = append(inputs, auditInput{Corpus: true, Name: "generated/native-width/table/" + text, Text: text})
			}
		}
	}
	if slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for depth := 1; depth <= 6; depth++ {
			for _, body := range []string{"", "a", "*a _b_*", "中 👩🏽‍⚕️", "[a](/b)", "a #", "`a  b`"} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := prefix + strings.Repeat("#", depth) + " " + body + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/structure/atx/" + text, Text: text})
				}
			}
		}
		for _, marker := range []string{"=", "===", "---", "-----"} {
			for _, body := range []string{"a", "a\nb", "*a _b_*", "中\n文", "a\n\ntext", "[x](/y)"} {
				for _, spacing := range []string{"", "  ", "\t"} {
					text := body + "\n" + marker + spacing + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/structure/setext/" + text, Text: text})
				}
			}
		}
	}
	if slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, space := range []string{"", " ", "\t", "\u00a0", "\u2028", "\ufeff"} {
			for _, body := range []string{"+    a\n*   b", "a\n====", "|a|b|\n|-|-|\n|中|😀|", "```js\na  b\n```", "<div>  \nx\n</div>"} {
				for _, marker := range []string{"next", "range", "nested", "unmatched"} {
					comment := "<!--" + space + "prettier-ignore" + space + "-->"
					text := comment + "\n" + body + "\n\n# after\n"
					if marker != "next" {
						text = "before\n\n<!--" + space + "prettier-ignore-start" + space + "-->\n" + body + "\n"
						if marker == "nested" {
							text += "<!-- prettier-ignore-start -->\n+   extra\n"
						}
						if marker != "unmatched" {
							text += "<!--" + space + "prettier-ignore-end" + space + "-->\n\n# after\n"
						}
					}
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/root/ignore/" + marker + "/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"", " \n", "a\n<div>\nx\n</div>\n", "[a]: /x\n[b]: /y\n", "<!-- prettier-ignore-start -->\n+    a\n<!-- prettier-ignore-end -->\n\n<!-- prettier-ignore-start -->\n*    b\n<!-- prettier-ignore-end -->\n", "<!-- prettier-ignore-end -->\n+    a\n"} {
			inputs = append(inputs, auditInput{Name: "generated/root/edge/" + text, Text: text})
		}
	}
	if slice == "leaves" || slice == "whitespace" {
		for _, label := range []string{"a", "中", "a b", "x-y"} {
			for _, url := range []string{"/x", "", "<x>", "/a%20b", "https://example.test"} {
				for _, title := range []string{"", " \"title\"", " 'a \"b\"'", " (a'b)"} {
					text := "[" + label + "](" + url + title + ") ![" + label + "](" + url + title + ")\n\n[" + label + "]: " + url + title + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/leaves/links/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"*a* **b** ~~c~~ _d_ 1*2*3 1***2***3\n", "_<https://example.test>_\n", "a  \nb\nc\\\nd\n", "[^x]\n\n[^x]: first\n    second\n\n    - a\n    - b\n", "- a\n\n  ***\n\n- b\n\n  ---\n", "$$ title\na+b\n$$\n\n$x + y$\n", "{{ a }} {% b %}\n", "[[wiki link]]\n", "![a][b] [a][b] ![b][] [b][] ![b] [b]\n\n[b]: /x 'title'\n"} {
			inputs = append(inputs, auditInput{Name: "generated/leaves/edge/" + text, Text: text})
		}
	}
	if slice == "whitespace" {
		for _, previous := range []string{"a", "中", "한", "。", "α", "…"} {
			for _, next := range []string{"a", "中", "한", "1.", "123)", "。", "-"} {
				for _, space := range []string{" ", "\n", "\t"} {
					text := previous + space + next + "\n"
					inputs = append(inputs, auditInput{Corpus: true, Name: "generated/whitespace/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{"[a  中][b]\n\n[b]: /x\n", "![a  中][b] [a\n中][b]\n\n[b]: /x\n", "|a b|中 c|\n|-|-|\n|x y|한 z|\n", "[^x]\n\n[^x]: a\n\n    {{ b }}\n"} {
			inputs = append(inputs, auditInput{Name: "generated/whitespace/edge/" + text, Text: text})
		}
	}
	return inputs, files
}

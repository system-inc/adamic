package markdownblocks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fixed hash shards keep existing cases in place as repository Markdown grows.
// ADAMIC_TEST_SHARD=i/n selects top-level shards by index modulo n; unset runs all.
const testMarkdownTextSplittingShards = 64

func textSplittingInputs(test *testing.T) (string, []auditInput) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, _ := auditCorpus(test, root)
	textVerifyCorpus(test, inputs)
	original := append([]auditInput(nil), inputs...)
	for _, input := range original {
		for index, line := range strings.Split(input.Text, "\n") {
			inputs = append(inputs, auditInput{Name: fmt.Sprintf("%s:%d", input.Name, index+1), Text: line})
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
	return root, inputs
}

func TestMarkdownTextSplitting_Setup(test *testing.T) {
	test.Parallel()
	textReadySetup(test)
}

func TestMarkdownTextSplittingUnion(test *testing.T) {
	test.Parallel()
	setup := textReadySetup(test)
	stop := textDeadline(test)
	defer stop()
	textVerifyLeaves(test)
	textShardInputs(test, setup.inputs, testMarkdownTextSplittingShards)
}

func textSplittingShard(test *testing.T, shard int) {
	if !textShardSelected(test, shard) {
		test.Skip("assigned to another worker")
	}
	setup := textReadySetup(test)
	configureMarkdownMemory(test)
	markdownMemory.acquire(2)
	test.Cleanup(func() { markdownMemory.release(2) })
	// Shared preparation and admission are complete before the case deadline starts.
	stopDeadline := textDeadline(test)
	defer stopDeadline()
	root, inputs, products := setup.root, setup.shards[shard], setup.products
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
	want := textExecute(test, nil, goBinary, cases)
	clean(test, "Go splitText", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, readError := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	originalBundle, readError := os.ReadFile(
		filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"),
	)
	if readError != nil {
		test.Fatal(readError)
	}
	equal(test, "pinned original markdown bundle", installed, originalBundle)
	library := textExecute(test, nil, "node", "testdata/text_library.mjs", fork, cases)
	clean(test, "actual original splitText", library)
	equal(test, "Go versus actual fork splitText", library.stdout, want.stdout)
	main := products.main
	binary := products.sanitized
	answer := textExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", answer},
		{"source Node", textOnNode(test, main, cases)},
		{"backend", textOnNode(test, products.backend, cases)},
	} {
		clean(test, side.name, side.result)
		if textOutputDifferenceError := textOutputDifference(
			inputs,
			side.result.stdout,
			want.stdout,
		); textOutputDifferenceError != nil {
			test.Fatalf("%s: %v", side.name, textOutputDifferenceError)
		}
	}
	if report := textLeakReport(test, products, cases); report != "" {
		test.Fatal(report)
	}
	for _, testCase := range []struct{ name, file, from, to string }{
		{"fake whitespace", "textTokens.ts", "!between &&", "between &&"},
		{"Hangul kind", "splitText.ts", "tokens.word(text, 'k-letter', false", "tokens.word(text, 'cj-letter', false"},
		{
			"variation selector",
			"splitText.ts",
			"if(inRanges(selector, variationSelectorRanges))",
			"if(!inRanges(selector, variationSelectorRanges))",
		},
	} {
		func() {
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, f := range []string{
				"splitText.ts",
				"textTokens.ts",
				"textClasses.ts",
				"emojiMatcher.ts",
				"widthTables.ts",
				"widthRuneRanges.ts",
				"codec.ts",
				"testdata/text_probe.ts",
			} {
				content, readError := os.ReadFile(f)
				if readError != nil {
					test.Fatal(readError)
				}
				if f == testCase.file {
					if strings.Count(string(content), testCase.from) != 1 {
						test.Fatal("mutation anchor")
					}
					content = []byte(strings.Replace(string(content), testCase.from, testCase.to, 1))
				}
				write(test, filepath.Join(scratch, f), content)
			}
			run := textOnNode(test, filepath.Join(scratch, "testdata/text_probe.ts"), cases)
			clean(test, testCase.name, run)
			if textShardFor(textMutantWitness(testCase.name), testMarkdownTextSplittingShards) == shard &&
				bytes.Equal(run.stdout, want.stdout) {
				test.Fatal("survived")
			}
			if !bytes.Equal(run.stdout, want.stdout) {
				test.Logf("%s caught in shard-%03d by observable token fields", testCase.name, shard)
			}
		}()
	}
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := products.release
		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"Go", goBinary, []string{cases}},
			{"native", fast, []string{cases}},
			{"actual original Node splitText", "node", []string{"testdata/text_library.mjs", fork, cases}},
		} {
			started := time.Now()
			for index := 0; index < 3; index++ {
				run := textExecute(test, nil, side.command, side.args...)
				clean(test, side.name, run)
				equal(test, side.name, run.stdout, want.stdout)
			}
			test.Logf(
				"%s %.1f texts/s, startup, decoding and output included",
				side.name,
				float64(3*len(inputs))/time.Since(started).Seconds(),
			)
		}
	} else {
		result := textExecute(test, nil, products.release, cases)
		clean(test, "release splitText", result)
		equal(test, "release splitText", result.stdout, want.stdout)
	}
	test.Logf("%d splitText sources in shard-%03d", len(inputs), shard)
}

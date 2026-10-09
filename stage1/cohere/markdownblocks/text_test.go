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

func textSplittingInputs(t *testing.T) (string, []auditInput) {
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	inputs, _ := auditCorpus(t, root)
	textVerifyCorpus(t, inputs)
	original := append([]auditInput(nil), inputs...)
	for _, input := range original {
		for i, line := range strings.Split(input.Text, "\n") {
			inputs = append(inputs, auditInput{Name: fmt.Sprintf("%s:%d", input.Name, i+1), Text: line})
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

func TestMarkdownTextSplittingUnion(t *testing.T) {
	parallelMarkdown(t)
	stop := textDeadline(t)
	defer stop()
	textVerifyLeaves(t)
	_, inputs := textSharedInputs(t)
	textShardInputs(t, inputs, testMarkdownTextSplittingShards)
}

func textSplittingShard(t *testing.T, shard int) {
	parallelMarkdownMemory(t, 2)
	stopDeadline := textDeadline(t)
	defer stopDeadline()
	setupStarted := time.Now()
	if !textShardSelected(t, shard) {
		t.Skip("assigned to another worker")
	}
	root, all := textSharedInputs(t)
	inputs := make([]auditInput, 0, len(all)/testMarkdownTextSplittingShards+1)
	for _, input := range all {
		if textShardFor(input.Name, testMarkdownTextSplittingShards) == shard {
			inputs = append(inputs, input)
		}
	}
	products := textSharedProducts(t, root)
	t.Logf("shard-%03d setup including product fetch/build %.3fs", shard, time.Since(setupStarted).Seconds())
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
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go splitText", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, e := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	originalBundle, e := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	equal(t, "pinned original markdown bundle", installed, originalBundle)
	library := execute(t, nil, "node", "testdata/text_library.mjs", fork, cases)
	clean(t, "actual original splitText", library)
	equal(t, "Go versus actual fork splitText", library.stdout, want.stdout)
	main := products.main
	binary := products.sanitized
	answer := execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onNode(t, products.backend, cases)}} {
		clean(t, side.name, side.result)
		if err := textOutputDifference(inputs, side.result.stdout, want.stdout); err != nil {
			t.Fatalf("%s: %v", side.name, err)
		}
	}
	if report := textLeakReport(t, products, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"fake whitespace", "textTokens.ts", "!between &&", "between &&"},
		{"Hangul kind", "splitText.ts", "tokens.word(text, 'k-letter', false", "tokens.word(text, 'cj-letter', false"},
		{"variation selector", "splitText.ts", "if(inRanges(selector, variationSelectorRanges))", "if(!inRanges(selector, variationSelectorRanges))"},
	} {
		func() {
			scratch := t.TempDir()
			if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
				t.Fatal(e)
			}
			for _, f := range []string{"splitText.ts", "textTokens.ts", "textClasses.ts", "emojiMatcher.ts", "widthTables.ts", "widthRuneRanges.ts", "codec.ts", "testdata/text_probe.ts"} {
				b, e := os.ReadFile(f)
				if e != nil {
					t.Fatal(e)
				}
				if f == m.file {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, f), b)
			}
			r := onNode(t, filepath.Join(scratch, "testdata/text_probe.ts"), cases)
			clean(t, m.name, r)
			if textShardFor(textMutantWitness(m.name), testMarkdownTextSplittingShards) == shard && bytes.Equal(r.stdout, want.stdout) {
				t.Fatal("survived")
			}
			if !bytes.Equal(r.stdout, want.stdout) {
				t.Logf("%s caught in shard-%03d by observable token fields", m.name, shard)
			}
		}()
	}
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := products.release
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node splitText", "node", []string{"testdata/text_library.mjs", fork, cases}}} {
			started := time.Now()
			for i := 0; i < 3; i++ {
				r := execute(t, nil, side.command, side.args...)
				clean(t, side.name, r)
				equal(t, side.name, r.stdout, want.stdout)
			}
			t.Logf("%s %.1f texts/s, startup, decoding and output included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
		}
	} else {
		result := execute(t, nil, products.release, cases)
		clean(t, "release splitText", result)
		equal(t, "release splitText", result.stdout, want.stdout)
	}
	t.Logf("%d splitText sources in shard-%03d", len(inputs), shard)
}

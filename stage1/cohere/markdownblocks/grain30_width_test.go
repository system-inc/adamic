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

func grainWidthInputs(t *testing.T) (string, []auditInput, int) {
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	inputs, files := auditCorpus(t, root)
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
	atoms := []string{"a", "\x7f", "中", "α", "\u0300", "\ufe0f", "\ufe0e", "\u200d", "©", "❤", "☀", "😀", "👩", "👨", "🧑", "🤝", "🏳", "🏴", "🇦", "🇿", "🏽", "🏻", "⚕", "♂", "♀", "1", "#", "\u20e3", "🦯", "➡"}
	for _, a := range atoms {
		for _, b := range atoms {
			for _, c := range atoms {
				inputs = append(inputs, auditInput{Name: "sequence/" + a + b + c, Text: a + b + c})
			}
		}
	}
	for _, text := range []string{"👩🏽‍⚕️", "👨🏻‍❤️‍💋‍👨🏿", "🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f", "👩‍👩‍👧‍👦", "🏳️‍🌈", "🧑🏿‍🦽‍➡️", "a\x7f", "中\x7f"} {
		inputs = append(inputs, auditInput{Name: "long/" + text, Text: text})
	}
	return root, inputs, files
}

func runGrainWidth(t *testing.T, shard int) {

	widthDependencies := os.Getenv("ADAMIC_MARKDOWNWIDTH_DEPS")
	if widthDependencies == "" {
		t.Skip("set ADAMIC_MARKDOWNWIDTH_DEPS to an npm install of emoji-regex@10.6.0, get-east-asian-width@1.6.0 and narrow-emojis@0.0.3; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	products := grainReady(t, "width")
	if shard == 11 {
		generator := grainGenerator(t, "width")
		formatter := grainTool(t, "formatter")
		stop := grainOwn(t)
		defer stop()
		grainRegenerate(t, "width", generator, formatter)
		grainPlanted(t)
		return
	}
	stop := grainOwn(t)
	defer stop()
	root, inputs, files := grainWidthInputs(t)
	if shard < grainCorpusShards {
		inputs = grainSelect(inputs, shard)
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
	clean(t, "Go StringWidth", want)
	library := grainExecute(t, nil, "node", "testdata/width_library.mjs", widthDependencies, cases)
	clean(t, "original dependencies", library)
	equal(t, "Go versus original dependencies", library.stdout, want.stdout)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	forkWidths := grainExecute(t, nil, "node", "testdata/width_fork.mjs", fork, cases)
	clean(t, "actual fork group-fit widths", forkWidths)
	equal(t, "actual fork group-fit widths", forkWidths.stdout, want.stdout)
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, cases)

		for _, side := range []struct {
			name   string
			result run
		}{{"native", answer}, {"source Node", grainNode(t, main, cases)}, {"backend", grainNode(t, products.backend, cases)}} {
			clean(t, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				a := strings.Split(string(side.result.stdout), "\n")
				b := strings.Split(string(want.stdout), "\n")
				for i := range b {
					if i >= len(a) || a[i] != b[i] {
						t.Fatalf("%s mismatch %q got %q want %q", side.name, inputs[i].Name, a[i], b[i])
					}
				}
			}
		}
		if report := grainLeaks(t, products, cases); report != "" {
			t.Fatal(report)
		}
	}

	for mutationIndex, m := range []struct{ name, from, to string }{{"East Asian width", "inRanges(code, wideRanges) ? 2 : 1", "inRanges(code, wideRanges) ? 1 : 1"}, {"narrow emoji", "=== end ? 1 : 2", "=== end ? 2 : 2"}, {"ASCII DEL shortcut", "code > 0x7f", "code > 0x7e"}} {
		if shard != grainCorpusShards+mutationIndex {
			continue
		}
		func(t *testing.T) {
			scratch := t.TempDir()
			if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
				t.Fatal(e)
			}
			for _, f := range []string{"width.ts", "widthTables.ts", "widthRuneRanges.ts", "emojiMatcher.ts", "codec.ts", "testdata/width_probe.ts"} {
				b, e := os.ReadFile(f)
				if e != nil {
					t.Fatal(e)
				}
				if f == "width.ts" {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, f), b)
			}
			r := grainNode(t, filepath.Join(scratch, "testdata/width_probe.ts"), cases)
			clean(t, m.name, r)
			if bytes.Equal(r.stdout, want.stdout) {
				t.Fatal("survived")
			}
			t.Log("caught by observable width output")
		}(t)
	}
	if shard < grainCorpusShards {

		if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
			fast := products.release

			for _, side := range []struct {
				name, command string
				args          []string
			}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"original Node dependencies", "node", []string{"testdata/width_library.mjs", widthDependencies, cases}}} {
				started := time.Now()
				for i := 0; i < 3; i++ {
					r := grainExecute(t, nil, side.command, side.args...)
					clean(t, side.name, r)
					equal(t, side.name, r.stdout, want.stdout)
				}
				t.Logf("%s %.1f texts/s, startup, decoding and output included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
			}
		} else {
			result := grainExecute(t, nil, products.release, cases)
			clean(t, "release width", result)
			equal(t, "release width", result.stdout, want.stdout)
		}
		t.Logf("%d widths: %d repository files, source lines, every Unicode scalar and generated sequences", len(inputs), files)

	}

}

func TestMarkdownUnicodeWidths_000(t *testing.T) { t.Parallel(); runGrainWidth(t, 0) }

func TestMarkdownUnicodeWidths_001(t *testing.T) { t.Parallel(); runGrainWidth(t, 1) }

func TestMarkdownUnicodeWidths_002(t *testing.T) { t.Parallel(); runGrainWidth(t, 2) }

func TestMarkdownUnicodeWidths_003(t *testing.T) { t.Parallel(); runGrainWidth(t, 3) }

func TestMarkdownUnicodeWidths_004(t *testing.T) { t.Parallel(); runGrainWidth(t, 4) }

func TestMarkdownUnicodeWidths_005(t *testing.T) { t.Parallel(); runGrainWidth(t, 5) }

func TestMarkdownUnicodeWidths_006(t *testing.T) { t.Parallel(); runGrainWidth(t, 6) }

func TestMarkdownUnicodeWidths_007(t *testing.T) { t.Parallel(); runGrainWidth(t, 7) }

func TestMarkdownUnicodeWidths_008(t *testing.T) { t.Parallel(); runGrainWidth(t, 8) }

func TestMarkdownUnicodeWidths_009(t *testing.T) { t.Parallel(); runGrainWidth(t, 9) }

func TestMarkdownUnicodeWidths_010(t *testing.T) { t.Parallel(); runGrainWidth(t, 10) }

func TestMarkdownUnicodeWidths_011(t *testing.T) { t.Parallel(); runGrainWidth(t, 11) }

func TestMarkdownUnicodeWidthsUnion(t *testing.T) {
	t.Parallel()
	_, inputs, _ := grainWidthInputs(t)
	grainUnion(t, inputs, 12, "TestMarkdownUnicodeWidths", "runGrainWidth")
}

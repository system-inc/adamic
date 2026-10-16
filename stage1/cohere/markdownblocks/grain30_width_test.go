package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

func grainWidthInputs(t *testing.T) (string, []auditInput, int) { return oracles.GrainWidthInputs(t) }

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
		files = (files + grainCorpusShards - 1 - shard) / grainCorpusShards
		if len(inputs) == 0 {
			grainPlanted(t)
			t.Log("empty corpus group")
			return
		}
	}
	grainPlanted(t)

	dir := t.TempDir()
	batch := grainWidthBatch(inputs)
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, batch)
	goBinary := products.goBinary
	fork := grainWidthFork(root)

	var want, forkWidths run
	if shard < grainCorpusShards {
		want = grainExecute(t, nil, goBinary, cases)
	} else {
		want, forkWidths = grainWidthFullAnswers(t, goBinary, fork, batch)
	}
	clean(t, "Go StringWidth", want)
	library := grainExecute(t, nil, "node", "testdata/width_library.mjs", widthDependencies, cases)
	clean(t, "original dependencies", library)
	equal(t, "Go versus original dependencies", library.stdout, want.stdout)
	if shard < grainCorpusShards {
		forkWidths = grainExecute(t, nil, "node", "testdata/width_fork.mjs", fork, cases)
	}
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

func grainWidthBatch(inputs []auditInput) []byte { return oracles.GrainWidthBatch(inputs) }

func grainWidthFork(root string) string { return oracles.GrainWidthFork(root) }

// The three mutant shards each need Go's and the fork's widths for the whole corpus, the same answers every time:
// about 14 s of one core per shard, the fork's group-fit search over every Unicode scalar most of it. They are a
// product keyed by the Go oracle's bytes, the fork bundle's, the fork script, Node's version and the whole batch.
// The corpus shards still run Go and the fork live on their own groups, and the original dependencies run live here.
func grainWidthFullAnswers(t *testing.T, binary, fork string, batch []byte) (run, run) {
	a, b := oracles.GrainWidthFullAnswers(t, binary, fork, batch)
	return run{stdout: a}, run{stdout: b}
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

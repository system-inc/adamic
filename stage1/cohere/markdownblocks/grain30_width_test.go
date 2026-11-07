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

func grainWidthInputs(test *testing.T) (string, []auditInput, int) {
	return oracles.GrainWidthInputs(test)
}

func runGrainWidth(test *testing.T, shard int) {

	widthDependencies := os.Getenv("ADAMIC_MARKDOWNWIDTH_DEPS")
	if widthDependencies == "" {
		test.Skip(
			"set ADAMIC_MARKDOWNWIDTH_DEPS to an npm install of emoji-regex@10.6.0, get-east-asian-width@1.6.0 and narrow-emojis@0.0.3; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	products := grainReady(test, "width")
	if shard == 11 {
		generator := grainGenerator(test, "width")
		formatter := grainTool(test, "formatter")
		stop := grainOwn(test)
		defer stop()
		grainRegenerate(test, "width", generator, formatter)
		grainPlanted(test)
		return
	}
	stop := grainOwn(test)
	defer stop()
	root, inputs, files := grainWidthInputs(test)
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
	batch := grainWidthBatch(inputs)
	cases := filepath.Join(dir, "cases.txt")
	write(test, cases, batch)
	goBinary := products.goBinary
	fork := grainWidthFork(root)

	var want, forkWidths run
	if shard < grainCorpusShards {
		want = grainExecute(test, nil, goBinary, cases)
	} else {
		want, forkWidths = grainWidthFullAnswers(test, goBinary, fork, batch)
	}
	clean(test, "Go StringWidth", want)
	library := grainExecute(test, nil, "node", "testdata/width_library.mjs", widthDependencies, cases)
	clean(test, "original dependencies", library)
	equal(test, "Go versus original dependencies", library.stdout, want.stdout)
	if shard < grainCorpusShards {
		forkWidths = grainExecute(test, nil, "node", "testdata/width_fork.mjs", fork, cases)
	}
	clean(test, "actual fork group-fit widths", forkWidths)
	equal(test, "actual fork group-fit widths", forkWidths.stdout, want.stdout)
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, cases)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"native", answer},
			{"source Node", grainNode(test, main, cases)},
			{"backend", grainNode(test, products.backend, cases)},
		} {
			clean(test, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				a := strings.Split(string(side.result.stdout), "\n")
				b := strings.Split(string(want.stdout), "\n")
				for i := range b {
					if i >= len(a) || a[i] != b[i] {
						test.Fatalf("%s mismatch %q got %q want %q", side.name, inputs[i].Name, a[i], b[i])
					}
				}
			}
		}
		if report := grainLeaks(test, products, cases); report != "" {
			test.Fatal(report)
		}
	}

	for mutationIndex, m := range []struct{ name, from, to string }{
		{"East Asian width", "inRanges(code, wideRanges) ? 2 : 1", "inRanges(code, wideRanges) ? 1 : 1"},
		{"narrow emoji", "=== end ? 1 : 2", "=== end ? 2 : 2"},
		{"ASCII DEL shortcut", "code > 0x7f", "code > 0x7e"},
	} {
		if shard != grainCorpusShards+mutationIndex {
			continue
		}
		func(test *testing.T) {
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, f := range []string{
				"width.ts",
				"widthTables.ts",
				"widthRuneRanges.ts",
				"emojiMatcher.ts",
				"codec.ts",
				"testdata/width_probe.ts",
			} {
				b, readError := os.ReadFile(f)
				if readError != nil {
					test.Fatal(readError)
				}
				if f == "width.ts" {
					if strings.Count(string(b), m.from) != 1 {
						test.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(test, filepath.Join(scratch, f), b)
			}
			r := grainNode(test, filepath.Join(scratch, "testdata/width_probe.ts"), cases)
			clean(test, m.name, r)
			if bytes.Equal(r.stdout, want.stdout) {
				test.Fatal("survived")
			}
			test.Log("caught by observable width output")
		}(test)
	}
	if shard < grainCorpusShards {

		if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
			fast := products.release

			for _, side := range []struct {
				name, command string
				args          []string
			}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {
				"original Node dependencies",
				"node",
				[]string{"testdata/width_library.mjs", widthDependencies, cases},
			}} {
				started := time.Now()
				for i := 0; i < 3; i++ {
					r := grainExecute(test, nil, side.command, side.args...)
					clean(test, side.name, r)
					equal(test, side.name, r.stdout, want.stdout)
				}
				test.Logf(
					"%s %.1f texts/s, startup, decoding and output included",
					side.name,
					float64(3*len(inputs))/time.Since(started).Seconds(),
				)
			}
		} else {
			result := grainExecute(test, nil, products.release, cases)
			clean(test, "release width", result)
			equal(test, "release width", result.stdout, want.stdout)
		}
		test.Logf(
			"%d widths: %d repository files, source lines, every Unicode scalar and generated sequences",
			len(inputs),
			files,
		)

	}

}

func grainWidthBatch(inputs []auditInput) []byte { return oracles.GrainWidthBatch(inputs) }

func grainWidthFork(root string) string { return oracles.GrainWidthFork(root) }

// The three mutant shards each need Go's and the fork's widths for the whole corpus, the same answers every time:
// about 14 s of one core per shard, the fork's group-fit search over every Unicode scalar most of it. They are a
// product keyed by the Go oracle's bytes, the fork bundle's, the fork script, Node's version and the whole batch.
// The corpus shards still run Go and the fork live on their own groups, and the original dependencies run live here.
func grainWidthFullAnswers(test *testing.T, binary, fork string, batch []byte) (run, run) {
	a, b := oracles.GrainWidthFullAnswers(test, binary, fork, batch)
	return run{stdout: a}, run{stdout: b}
}

func TestMarkdownUnicodeWidths_000(test *testing.T) { test.Parallel(); runGrainWidth(test, 0) }

func TestMarkdownUnicodeWidths_001(test *testing.T) { test.Parallel(); runGrainWidth(test, 1) }

func TestMarkdownUnicodeWidths_002(test *testing.T) { test.Parallel(); runGrainWidth(test, 2) }

func TestMarkdownUnicodeWidths_003(test *testing.T) { test.Parallel(); runGrainWidth(test, 3) }

func TestMarkdownUnicodeWidths_004(test *testing.T) { test.Parallel(); runGrainWidth(test, 4) }

func TestMarkdownUnicodeWidths_005(test *testing.T) { test.Parallel(); runGrainWidth(test, 5) }

func TestMarkdownUnicodeWidths_006(test *testing.T) { test.Parallel(); runGrainWidth(test, 6) }

func TestMarkdownUnicodeWidths_007(test *testing.T) { test.Parallel(); runGrainWidth(test, 7) }

func TestMarkdownUnicodeWidths_008(test *testing.T) { test.Parallel(); runGrainWidth(test, 8) }

func TestMarkdownUnicodeWidths_009(test *testing.T) { test.Parallel(); runGrainWidth(test, 9) }

func TestMarkdownUnicodeWidths_010(test *testing.T) { test.Parallel(); runGrainWidth(test, 10) }

func TestMarkdownUnicodeWidths_011(test *testing.T) { test.Parallel(); runGrainWidth(test, 11) }

func TestMarkdownUnicodeWidthsUnion(test *testing.T) {
	test.Parallel()
	_, inputs, _ := grainWidthInputs(test)
	grainUnion(test, inputs, 12, "TestMarkdownUnicodeWidths", "runGrainWidth")
}

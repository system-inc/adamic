package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func grainPathInputs(t *testing.T) (string, []auditInput, int) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := blockCorpus(t, root, "whitespace")
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(t, root, inputs, files, 1)
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	for depth := 1; depth <= 32; depth++ {
		text := strings.Repeat("> ", depth) + "- a *b*\n"
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/path/depth%d", depth), Text: text})
	}
	return root, inputs, files
}

func runGrainPath(t *testing.T, shard int) {
	products := grainReady(t, "path")
	stop := grainOwn(t)
	defer stop()
	root, inputs, files := grainPathInputs(t)
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
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, item := range inputs {
		if err := encoder.Encode(item); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goBinary

	nativeCases := filepath.Join(dir, "native.txt")
	want := grainExecute(t, nil, goBinary, cases, nativeCases)
	clean(t, "actual Go AstPath", want)
	if keep := os.Getenv("ADAMIC_PATH_KEEP"); keep != "" {
		keep = filepath.Join(keep, t.Name())
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(nativeCases)
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "native.txt"), data)
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, err := os.ReadFile(filepath.Join(fork, "standalone.js"))
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/standalone.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned standalone", installed, pinned)
	gapGo := grainExecute(t, nil, goBinary, "--key-gap")
	clean(t, "Go key witness", gapGo)
	equal(t, "Go key witness", gapGo.stdout, []byte("key=\"\" present=false\n"))
	gapFork := grainExecute(t, nil, "node", "testdata/path_library.mjs", fork, "--key-gap")
	clean(t, "fork key witness", gapFork)
	equal(t, "fork key witness", gapFork.stdout, []byte("key=0 type=number\n"))
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, nativeCases)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"actual Go AstPath from transport", grainExecute(t, nil, goBinary, "--facts", nativeCases)},
			{"actual original AstPath", grainExecute(t, nil, "node", "testdata/path_library.mjs", fork, nativeCases)}, {"native", answer}, {"source Node", grainNode(t, main, nativeCases)}, {"backend", grainNode(t, products.backend, nativeCases)},
		} {
			clean(t, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				a := strings.Split(string(side.result.stdout), "\n")
				b := strings.Split(string(want.stdout), "\n")
				if len(a) != len(b) {
					t.Fatalf("%s result count %d/%d", side.name, len(a), len(b))
				}
				for i, line := range b {
					if a[i] != line {
						offset := firstDifference(a[i], line)
						t.Fatalf("%s %s byte%d got %q want %q", side.name, inputs[i].Name, offset, a[i][max(0, offset-100):min(len(a[i]), offset+100)], line[max(0, offset-100):min(len(line), offset+100)])
					}
				}
			}
		}
		if report := grainLeaks(t, products, nativeCases); report != "" {
			t.Fatal(report)
		}
	}

	for mutationIndex, m := range []struct{ name, from, to string }{
		{"map callback result", "results.push(callback(path, index, array));", "results.push(callback(path, index, array) + 1);"},
		{"array frame", "this.at(-3)", "this.at(-5)"},
		{"ancestor order", "index -= 2", "index -= 4"},
		{"call restore", "this.truncate(length);\n        return result;", "this.truncate(length + 2);\n        return result;"},
	} {
		if shard != grainCorpusShards+mutationIndex {
			continue
		}
		func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"astArena.ts", "astWalk.ts", "astProtocol.ts", "codec.ts", "astPath.ts", "pathObserve.ts", "testdata/path_probe.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "astPath.ts" {
					if !strings.Contains(string(data), m.from) {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := grainNode(t, filepath.Join(scratch, "testdata/path_probe.ts"), nativeCases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("survived")
			}
			t.Logf("caught by path bytes, first difference%d", firstDifference(string(result.stdout), string(want.stdout)))
		}(t)
	}
	if shard < grainCorpusShards {
		fast := products.release

		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"actual Go AstPath", goBinary, []string{"--facts", nativeCases}},
			{"native AstPath", fast, []string{nativeCases}},
			{"actual original AstPath", "node", []string{"testdata/path_library.mjs", fork, nativeCases}},
		} {
			start := time.Now()
			for i := 0; i < 3; i++ {
				result := grainExecute(t, nil, side.command, side.args...)
				clean(t, side.name, result)
				equal(t, side.name, result.stdout, want.stdout)
			}
			t.Logf("%s %.1f documents/s, three runs, full path observation I/O, AST parsing excluded", side.name, float64(3*len(inputs))/time.Since(start).Seconds())
		}
		t.Logf("%d contexts, %d physical files, all per-node getter bytes and parent-node call/map/each/parent restore bytes agree; %d output bytes", len(inputs), files, len(want.stdout))

	}

}

func TestMarkdownAstPath_000(t *testing.T) { t.Parallel(); runGrainPath(t, 0) }

func TestMarkdownAstPath_001(t *testing.T) { t.Parallel(); runGrainPath(t, 1) }

func TestMarkdownAstPath_002(t *testing.T) { t.Parallel(); runGrainPath(t, 2) }

func TestMarkdownAstPath_003(t *testing.T) { t.Parallel(); runGrainPath(t, 3) }

func TestMarkdownAstPath_004(t *testing.T) { t.Parallel(); runGrainPath(t, 4) }

func TestMarkdownAstPath_005(t *testing.T) { t.Parallel(); runGrainPath(t, 5) }

func TestMarkdownAstPath_006(t *testing.T) { t.Parallel(); runGrainPath(t, 6) }

func TestMarkdownAstPath_007(t *testing.T) { t.Parallel(); runGrainPath(t, 7) }

func TestMarkdownAstPath_008(t *testing.T) { t.Parallel(); runGrainPath(t, 8) }

func TestMarkdownAstPath_009(t *testing.T) { t.Parallel(); runGrainPath(t, 9) }

func TestMarkdownAstPath_010(t *testing.T) { t.Parallel(); runGrainPath(t, 10) }

func TestMarkdownAstPath_011(t *testing.T) { t.Parallel(); runGrainPath(t, 11) }

func TestMarkdownAstPathUnion(t *testing.T) {
	t.Parallel()
	_, inputs, _ := grainPathInputs(t)
	grainUnion(t, inputs, 12, "TestMarkdownAstPath", "runGrainPath")
}

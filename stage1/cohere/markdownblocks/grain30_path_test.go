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

func grainPathInputs(test *testing.T) (string, []auditInput, int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, files := blockCorpus(test, root, "whitespace")
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(test, root, inputs, files, 1)
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	for depth := 1; depth <= 32; depth++ {
		text := strings.Repeat("> ", depth) + "- a *b*\n"
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/path/depth%d", depth), Text: text})
	}
	return root, inputs, files
}

func runGrainPath(test *testing.T, shard int) {
	products := grainReady(test, "path")
	stop := grainOwn(test)
	defer stop()
	root, inputs, files := grainPathInputs(test)
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

	goBinary := products.goBinary
	nativeCases, want := grainPathCases(test, goBinary, inputs)
	if keep := os.Getenv("ADAMIC_PATH_KEEP"); keep != "" {
		keep = filepath.Join(keep, test.Name())
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		data, readError := os.ReadFile(nativeCases)
		if readError != nil {
			test.Fatal(readError)
		}
		write(test, filepath.Join(keep, "native.txt"), data)
		write(test, filepath.Join(keep, "want.txt"), want.stdout)
	}
	fork := grainPathFork(test, root)
	gapGo := grainExecute(test, nil, goBinary, "--key-gap")
	clean(test, "Go key witness", gapGo)
	equal(test, "Go key witness", gapGo.stdout, []byte("key=\"\" present=false\n"))
	gapFork := grainExecute(test, nil, "node", "testdata/path_library.mjs", fork, "--key-gap")
	clean(test, "fork key witness", gapFork)
	equal(test, "fork key witness", gapFork.stdout, []byte("key=0 type=number\n"))
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, nativeCases)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"actual Go AstPath from transport", grainExecute(test, nil, goBinary, "--facts", nativeCases)},
			{
				"actual original AstPath",
				grainExecute(test, nil, "node", "testdata/path_library.mjs", fork, nativeCases),
			},
			{"native", answer},
			{"source Node", grainNode(test, main, nativeCases)},
			{"backend", grainNode(test, products.backend, nativeCases)},
		} {
			clean(test, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				a := strings.Split(string(side.result.stdout), "\n")
				b := strings.Split(string(want.stdout), "\n")
				if len(a) != len(b) {
					test.Fatalf("%s result count %d/%d", side.name, len(a), len(b))
				}
				for i, line := range b {
					if a[i] != line {
						offset := firstDifference(a[i], line)
						test.Fatalf(
							"%s %s byte%d got %q want %q",
							side.name,
							inputs[i].Name,
							offset,
							a[i][max(0, offset-100):min(len(a[i]), offset+100)],
							line[max(0, offset-100):min(len(line), offset+100)],
						)
					}
				}
			}
		}
		if report := grainLeaks(test, products, nativeCases); report != "" {
			test.Fatal(report)
		}
	}

	for mutationIndex, m := range []struct{ name, from, to string }{
		{
			"map callback result",
			"results.push(callback(path, index, array));",
			"results.push(callback(path, index, array) + 1);",
		},
		{"array frame", "this.at(-3)", "this.at(-5)"},
		{"ancestor order", "index -= 2", "index -= 4"},
		{
			"call restore",
			"this.truncate(length);\n        return result;",
			"this.truncate(length + 2);\n        return result;",
		},
	} {
		if shard != grainCorpusShards+mutationIndex {
			continue
		}
		func(test *testing.T) {
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, file := range []string{
				"astArena.ts",
				"astWalk.ts",
				"astProtocol.ts",
				"codec.ts",
				"astPath.ts",
				"pathObserve.ts",
				"testdata/path_probe.ts",
			} {
				data, readError := os.ReadFile(file)
				if readError != nil {
					test.Fatal(readError)
				}
				if file == "astPath.ts" {
					if !strings.Contains(string(data), m.from) {
						test.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(test, filepath.Join(scratch, file), data)
			}
			result := grainNode(test, filepath.Join(scratch, "testdata/path_probe.ts"), nativeCases)
			clean(test, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("survived")
			}
			test.Logf(
				"caught by path bytes, first difference%d",
				firstDifference(string(result.stdout), string(want.stdout)),
			)
		}(test)
	}
	if shard < grainCorpusShards {
		test.Logf(
			"%d contexts, %d physical files, all per-node getter bytes and parent-node call/map/each/parent restore bytes agree; %d output bytes",
			len(inputs),
			files,
			len(want.stdout),
		)
	}
}

// Each side's three timed runs are their own test beside the corpus shard of the same number,
// which owns the same contexts, so neither carries both the comparisons and the timings.
func runGrainPathThroughput(test *testing.T, shard int) {
	products := grainReady(test, "path")
	stop := grainOwn(test)
	defer stop()
	root, inputs, _ := grainPathInputs(test)
	inputs = grainSelect(inputs, shard)
	grainPlanted(test)
	if len(inputs) == 0 {
		test.Log("empty corpus group")
		return
	}
	goBinary := products.goBinary
	nativeCases, want := grainPathCases(test, goBinary, inputs)
	fork := grainPathFork(test, root)
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
			result := grainExecute(test, nil, side.command, side.args...)
			clean(test, side.name, result)
			equal(test, side.name, result.stdout, want.stdout)
		}
		test.Logf(
			"%s %.1f documents/s, three runs, full path observation I/O, AST parsing excluded",
			side.name,
			float64(3*len(inputs))/time.Since(start).Seconds(),
		)
	}
}

// grainPathCases writes a shard's contexts as the probes read them and returns Go's path bytes for them.
func grainPathCases(test *testing.T, goBinary string, inputs []auditInput) (string, run) {
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
	nativeCases := filepath.Join(dir, "native.txt")
	want := grainExecute(test, nil, goBinary, cases, nativeCases)
	clean(test, "actual Go AstPath", want)
	return nativeCases, want
}

// grainPathFork is the original AstPath's bundle, which must be the pinned one.
func grainPathFork(test *testing.T, root string) string {
	cohere := filepath.Join(root, "cohere")
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, readError := os.ReadFile(filepath.Join(fork, "standalone.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	pinned, readError := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/standalone.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	equal(test, "pinned standalone", installed, pinned)
	return fork
}

func TestMarkdownAstPath_000(test *testing.T) { test.Parallel(); runGrainPath(test, 0) }

func TestMarkdownAstPath_001(test *testing.T) { test.Parallel(); runGrainPath(test, 1) }

func TestMarkdownAstPath_002(test *testing.T) { test.Parallel(); runGrainPath(test, 2) }

func TestMarkdownAstPath_003(test *testing.T) { test.Parallel(); runGrainPath(test, 3) }

func TestMarkdownAstPath_004(test *testing.T) { test.Parallel(); runGrainPath(test, 4) }

func TestMarkdownAstPath_005(test *testing.T) { test.Parallel(); runGrainPath(test, 5) }

func TestMarkdownAstPath_006(test *testing.T) { test.Parallel(); runGrainPath(test, 6) }

func TestMarkdownAstPath_007(test *testing.T) { test.Parallel(); runGrainPath(test, 7) }

func TestMarkdownAstPath_008(test *testing.T) { test.Parallel(); runGrainPath(test, 8) }

func TestMarkdownAstPath_009(test *testing.T) { test.Parallel(); runGrainPath(test, 9) }

func TestMarkdownAstPath_010(test *testing.T) { test.Parallel(); runGrainPath(test, 10) }

func TestMarkdownAstPath_011(test *testing.T) { test.Parallel(); runGrainPath(test, 11) }

func TestMarkdownAstPathThroughput_000(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 0)
}

func TestMarkdownAstPathThroughput_001(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 1)
}

func TestMarkdownAstPathThroughput_002(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 2)
}

func TestMarkdownAstPathThroughput_003(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 3)
}

func TestMarkdownAstPathThroughput_004(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 4)
}

func TestMarkdownAstPathThroughput_005(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 5)
}

func TestMarkdownAstPathThroughput_006(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 6)
}

func TestMarkdownAstPathThroughput_007(test *testing.T) {
	test.Parallel()
	runGrainPathThroughput(test, 7)
}

func TestMarkdownAstPathUnion(test *testing.T) {
	test.Parallel()
	_, inputs, _ := grainPathInputs(test)
	grainUnion(test, inputs, 12, "TestMarkdownAstPath", "runGrainPath")
	// The timing shards select with the corpus shards' grainSelect, so the union above covers them too.
	grainVerifyChildren(test, "TestMarkdownAstPathThroughput", "runGrainPathThroughput", grainCorpusShards)
}

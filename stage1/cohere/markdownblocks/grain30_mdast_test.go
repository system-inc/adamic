package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func grainMdastInputs(test *testing.T) (string, []auditInput, int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, files := blockCorpus(test, root, "whitespace")
	smokeStart := len(inputs)
	for _, text := range []string{
		"",
		"~~a\nb~~\n",
		"$$ meta\na\nb\n$$\n",
		"$ a $ and $b$\n",
		"[[target]] {{x}} {% x\ny %}\n",
		"[^a]: x\n\n[^a]\n",
		"[x]: /u \"title\"\n\n![a &amp; b][x]\n",
		"| a | b |\n|:-:|--:|\n|`a\\|b`|~~x~~|\n",
		"- [x] \n- [ ] a\n\n  > b\n",
		"<a href=\"x\">t</a>\n",
		"a &NotEqualTilde; &#x1F600; &#0; \\*\n",
	} {
		inputs = append(inputs, auditInput{Name: "generated/mdast/" + text, Text: text})
	}
	for _, delimiter := range []string{"---", "+++"} {
		for _, language := range []string{"", " yaml", " toml", " JSON", "\uFEFFyaml\u00A0", "\u0085yaml"} {
			for _, closing := range []string{"---", "+++", "...", "---suffix"} {
				text := delimiter + language + "\n中😀: x\n" + closing + "\n\n# h\n"
				inputs = append(inputs, auditInput{Name: "generated/mdast-front/" + text, Text: text})
			}
		}
	}
	for _, text := range []string{
		"---\n---",
		"---\n...",
		"+++\n+++tail",
		"---yaml\n",
		"--- yaml\n---suffix\nx",
		"---\n😀\n...\n中",
	} {
		inputs = append(inputs, auditInput{Name: "generated/mdast-front-edge/" + text, Text: text})
	}
	if os.Getenv("ADAMIC_MDAST_SMOKE") != "" {
		inputs = append(inputs[smokeStart:smokeStart+11], inputs[len(inputs)-6:]...)
		files = 0
	}
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(test, root, inputs, files, 1)
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	return root, inputs, files
}

func runGrainMdast(test *testing.T, shard int) {
	products := grainReady(test, "mdast")
	stop := grainOwn(test)
	defer stop()
	root, inputs, files := grainMdastInputs(test)
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
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, i := range inputs {
		if encodeError := enc.Encode(i); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(test, cases, b.Bytes())
	cohere := filepath.Join(root, "cohere")
	goBinary := products.goBinary

	transport := filepath.Join(dir, "events.txt")
	want := grainExecute(test, nil, goBinary, cases, transport)
	clean(test, "actual Go mdast", want)
	if keep := os.Getenv("ADAMIC_MDAST_KEEP"); keep != "" {
		keep = filepath.Join(keep, test.Name())
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		data, readError := os.ReadFile(transport)
		if readError != nil {
			test.Fatal(readError)
		}
		write(test, filepath.Join(keep, "events.txt"), data)
		write(test, filepath.Join(keep, "want.txt"), want.stdout)
		write(test, filepath.Join(keep, "cases.jsonl"), b.Bytes())
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, readError := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	pinned, readError := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	equal(test, "pinned fork", installed, pinned)
	main := products.main
	if shard < grainCorpusShards {
		answer := grainExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, transport)

		for _, side := range []struct {
			name   string
			result run
		}{
			{"actual Go compiler from events", grainExecute(test, nil, goBinary, "--events", transport)},
			{
				"actual original compiler from events",
				grainExecute(test, nil, "node", "testdata/mdast_library.mjs", fork, "--events", transport),
			},
			{"actual original mdast", grainExecute(test, nil, "node", "testdata/mdast_library.mjs", fork, cases)},
			{"native", answer},
			{"source Node", grainNode(test, main, transport)},
			{"backend", grainNode(test, products.backend, transport)},
		} {
			clean(test, side.name, side.result)
			if grainDifference(side.result.stdout, want.stdout) != nil {
				a := strings.Split(string(side.result.stdout), "\n")
				z := strings.Split(string(want.stdout), "\n")
				if len(a) != len(z) {
					test.Fatalf("%s tree count%d/%d", side.name, len(a), len(z))
				}
				for i, row := range z {
					if row != a[i] {
						offset := firstDifference(a[i], row)
						test.Fatalf(
							"%s %s byte%d got%q want%q",
							side.name,
							inputs[i].Name,
							offset,
							a[i][max(0, offset-100):min(len(a[i]), offset+100)],
							row[max(0, offset-100):min(len(row), offset+100)],
						)
					}
				}
			}
		}
		if report := grainLeaks(test, products, transport); report != "" {
			test.Fatal(report)
		}
	}

	for mutationIndex, m := range []struct{ name, file, from, to string }{
		{
			"text construction",
			"mdastCompile.ts",
			"tail.value += this.serialize(token);",
			"tail.value += '!' + this.serialize(token);",
		},
		{"list spread", "mdastCompile.ts", "this.top().spread = token.spread;", "this.top().spread = !token.spread;"},
		{"front matter fallback", "parseFrontMatter.ts", "language === 'yaml'", "language === 'toml'"},
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
				"tokenizerEvents.ts",
				"tokenArena.ts",
				"tokenSource.ts",
				"mdastNode.ts",
				"inputChunks.ts",
				"codec.ts",
				"mdastArena.ts",
				"mdastCompile.ts",
				"parseFrontMatter.ts",
				"identifier.ts",
				"identifierCaseKeys.ts",
				"identifierCaseValues.ts",
				"decodeString.ts",
				"upperEntityNames.ts",
				"upperEntityValues.ts",
				"lowerEntityNames.ts",
				"lowerEntityValues.ts",
				"testdata/mdast_probe.ts",
			} {
				data, readError := os.ReadFile(file)
				if readError != nil {
					test.Fatal(readError)
				}
				if file == m.file {
					if !strings.Contains(string(data), m.from) {
						test.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(test, filepath.Join(scratch, file), data)
			}
			result := grainNode(test, filepath.Join(scratch, "testdata/mdast_probe.ts"), transport)
			clean(test, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("survived")
			}
			a := strings.Split(string(result.stdout), "\n")
			z := strings.Split(string(want.stdout), "\n")
			for i, row := range z {
				if i >= len(a) || a[i] != row {
					if i >= len(a) {
						test.Logf("caught by missing tree%d", i)
					} else {
						test.Logf("caught by %s at byte%d", inputs[i].Name, firstDifference(a[i], row))
					}
					break
				}
			}
		}(test)
	}
	if shard < grainCorpusShards {
		fast := products.release

		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"actual Go construction", goBinary, []string{"--events", transport}},
			{"native construction", fast, []string{transport}},
			{
				"actual original Node construction",
				"node",
				[]string{"testdata/mdast_library.mjs", fork, "--events", transport},
			},
		} {
			started := time.Now()
			for i := 0; i < 3; i++ {
				result := grainExecute(test, nil, side.command, side.args...)
				clean(test, side.name, result)
				equal(test, side.name, result.stdout, want.stdout)
			}
			test.Logf(
				"%s %.1f texts/s; three runs, identical event input and tree output, startup/I/O included, lexical parsing excluded",
				side.name,
				float64(3*len(inputs))/time.Since(started).Seconds(),
			)
		}
		test.Logf(
			"%d full constructed trees, %d physical files; actual Go and pinned fork, native/source/backend, sanitizer and leaks agree on every field and coordinate; %d bytes",
			len(inputs),
			files,
			len(want.stdout),
		)

	}

}

func TestNativeMdastConstruction_000(test *testing.T) { test.Parallel(); runGrainMdast(test, 0) }

func TestNativeMdastConstruction_001(test *testing.T) { test.Parallel(); runGrainMdast(test, 1) }

func TestNativeMdastConstruction_002(test *testing.T) { test.Parallel(); runGrainMdast(test, 2) }

func TestNativeMdastConstruction_003(test *testing.T) { test.Parallel(); runGrainMdast(test, 3) }

func TestNativeMdastConstruction_004(test *testing.T) { test.Parallel(); runGrainMdast(test, 4) }

func TestNativeMdastConstruction_005(test *testing.T) { test.Parallel(); runGrainMdast(test, 5) }

func TestNativeMdastConstruction_006(test *testing.T) { test.Parallel(); runGrainMdast(test, 6) }

func TestNativeMdastConstruction_007(test *testing.T) { test.Parallel(); runGrainMdast(test, 7) }

func TestNativeMdastConstruction_008(test *testing.T) { test.Parallel(); runGrainMdast(test, 8) }

func TestNativeMdastConstruction_009(test *testing.T) { test.Parallel(); runGrainMdast(test, 9) }

func TestNativeMdastConstruction_010(test *testing.T) { test.Parallel(); runGrainMdast(test, 10) }

func TestNativeMdastConstructionUnion(test *testing.T) {
	test.Parallel()
	_, inputs, _ := grainMdastInputs(test)
	grainUnion(test, inputs, 11, "TestNativeMdastConstruction", "runGrainMdast")
}

package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

func TestMicromarkInputChunks(t *testing.T) {
	parallelMarkdownMemory(t, 1)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus, files := blockCorpus(t, root, "whitespace")
	inputs := [][]uint16{{}}
	names := []string{"empty"}
	for _, item := range corpus {
		inputs = append(inputs, utf16.Encode([]rune(item.Text)))
		names = append(names, item.Name)
	}
	for unit := 0; unit <= 65535; unit++ {
		inputs = append(inputs, []uint16{uint16(unit)})
		names = append(names, fmt.Sprintf("unit/%d", unit))
	}
	alphabet := []uint16{0, 9, 10, 13, 65279, 65}
	for length := 1; length <= 5; length++ {
		count := 1
		for i := 0; i < length; i++ {
			count *= len(alphabet)
		}
		for n := 0; n < count; n++ {
			units := make([]uint16, length)
			number := n
			for i := range units {
				units[i] = alphabet[number%len(alphabet)]
				number /= len(alphabet)
			}
			inputs = append(inputs, units)
			names = append(names, fmt.Sprintf("sequence/%d/%d", length, n))
		}
	}
	for _, prefix := range []string{"", "a", "ab", "abc", "abcd", "abcde", "中", "😀", "a😀"} {
		for _, tail := range []string{"\t", "\t\t", "\r\n", "\rX\n", "\rX", "\x00\t", "\ufeff\t"} {
			inputs = append(inputs, utf16.Encode([]rune(prefix+tail)))
			names = append(names, fmt.Sprintf("columns/%q/%q", prefix, tail))
		}
	}
	selection := selectMarkdownFiles(t, root, corpus, files, markdownCorpusStride)
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
	}
	fullInputs, fullNames := inputs, names
	var rows []int
	inputs, names, rows = sampledNumericInputs(inputs, names, corpus, selection)
	batch := numericBatch(inputs)
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, batch)
	fullCases := cases
	if selection.Sample {
		fullCases = filepath.Join(dir, "mutant-cases.txt")
		write(t, fullCases, numericBatch(fullInputs))
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_chunks/main.go")
	driver, err := filepath.Abs("testdata/chunks_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/chunks_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_chunks.go"): bridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-chunks")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("Go chunks %v %s", err, output)
	}
	fullWant := execute(t, nil, goBinary, fullCases)
	clean(t, "full Go mutant oracle", fullWant)
	want := fullWant
	if selection.Sample {
		want = selectedNumericAnswers(t, fullWant, rows, len(fullInputs))
	}
	clean(t, "Go chunks", want)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned bundle", installed, pinned)
	main, err := filepath.Abs("testdata/chunks_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"actual original micromark preprocessing", execute(t, nil, "node", "testdata/chunks_library.mjs", fork, cases)},
		{"native", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onJavaScriptBackend(t, program, cases)},
	} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want.stdout)
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, from, to string }{
		{"initial BOM", "units[0] === 65279 ? 1 : 0", "units[0] === 65278 ? 1 : 0"},
		{"tab stop", "Math.ceil(column / 4) * 4", "Math.ceil(column / 3) * 3"},
		{"CRLF", "code === 10 && start === end && carriage", "code === 9 && start === end && carriage"},
	} {
		t.Run(m.name, func(t *testing.T) {
			cases, want, names := fullCases, fullWant, fullNames
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"inputChunks.ts", "testdata/chunks_probe.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "inputChunks.ts" {
					if strings.Count(string(data), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/chunks_probe.ts"), cases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("survived")
			}
			observed := strings.Split(string(result.stdout), "\n")
			for i, line := range strings.Split(string(want.stdout), "\n") {
				if i >= len(observed) {
					t.Logf("caught by missing chunk result%d", i)
					break
				}
				if observed[i] != line {
					t.Logf("caught by %s at chunk byte%d", names[i], firstDifference(observed[i], line))
					break
				}
			}
		})
	}
	fast := filepath.Join(dir, "fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node", "node", []string{"testdata/chunks_library.mjs", fork, cases}},
	} {
		started := time.Now()
		for i := 0; i < 3; i++ {
			result := execute(t, nil, side.command, side.args...)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("%s %.1f texts/s; three runs, startup and identical numeric UTF-16/chunk transport included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
	}
	t.Logf("%d cases: %d physical documents; 4943 layout documents; all 65536 UTF-16 units, including lone surrogates; special-code sequences through length five; column/BOM/NUL/CRLF/tab edges", len(inputs), len(selection.Paths))
}

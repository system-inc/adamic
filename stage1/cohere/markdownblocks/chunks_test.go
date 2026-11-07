package markdownblocks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// Not parallel: shared markdownMemory configuration; existing helper controls parallel execution.
func TestMicromarkInputChunks(test *testing.T) {
	parallelMarkdownMemory(test, 1)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus, files := blockCorpus(test, root, "whitespace")
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
	selection := selectMarkdownFiles(test, root, corpus, files, markdownCorpusStride)
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
	}
	fullInputs, fullNames := inputs, names
	var rows []int
	inputs, names, rows = sampledNumericInputs(inputs, names, corpus, selection)
	batch := numericBatch(inputs)
	dir := test.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	write(test, cases, batch)
	fullCases := cases
	if selection.Sample {
		fullCases = filepath.Join(dir, "mutant-cases.txt")
		write(test, fullCases, numericBatch(fullInputs))
	}
	cohere := filepath.Join(root, "cohere")
	goBinary := chunksGoOracle(test)
	fullWant := execute(test, nil, goBinary, fullCases)
	clean(test, "full Go mutant oracle", fullWant)
	want := fullWant
	if selection.Sample {
		want = selectedNumericAnswers(test, fullWant, rows, len(fullInputs))
	}
	clean(test, "Go chunks", want)
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
	equal(test, "pinned bundle", installed, pinned)
	main, absolutePathError := filepath.Abs("testdata/chunks_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	emitted := filepath.Join(probeEmitted(test, chunksProbe), "port.mjs")
	binary := filepath.Join(probeNative(test, chunksProbe, native.Options{Sanitize: true}), "port")
	fast := filepath.Join(probeNative(test, chunksProbe, native.Options{}), "port")
	answer := sanitizedRun(test, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{
			"actual original micromark preprocessing",
			execute(test, nil, "node", "testdata/chunks_library.mjs", fork, cases),
		},
		{"native", answer}, {"source Node", onNode(test, main, cases)}, {"backend", onNode(test, emitted, cases)},
	} {
		clean(test, side.name, side.result)
		equal(test, side.name, side.result.stdout, want.stdout)
	}
	// The leak check reads the built ports, macOS's leaks tool on the release one or LeakSanitizer on Linux running the
	// sanitized one again, as leaks does, without building a port inside this test.
	if runtime.GOOS == "darwin" {
		if report := execute(test, nil, "leaks", "--atExit", "--", fast, cases); report.exitCode != 0 {
			test.Fatal(string(report.stdout))
		}
	} else if report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases); report.exitCode != 0 {
		test.Fatalf("exit %d\n%s", report.exitCode, report.stderr)
	}
	for _, m := range []struct{ name, from, to string }{
		{"initial BOM", "units[0] === 65279 ? 1 : 0", "units[0] === 65278 ? 1 : 0"},
		{"tab stop", "Math.ceil(column / 4) * 4", "Math.ceil(column / 3) * 3"},
		{"CRLF", "code === 10 && start === end && carriage", "code === 9 && start === end && carriage"},
	} {
		test.Run(m.name, func(test *testing.T) {
			cases, want, names := fullCases, fullWant, fullNames
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, file := range []string{"inputChunks.ts", "testdata/chunks_probe.ts"} {
				data, readError := os.ReadFile(file)
				if readError != nil {
					test.Fatal(readError)
				}
				if file == "inputChunks.ts" {
					if strings.Count(string(data), m.from) != 1 {
						test.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(test, filepath.Join(scratch, file), data)
			}
			result := onNode(test, filepath.Join(scratch, "testdata/chunks_probe.ts"), cases)
			clean(test, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("survived")
			}
			observed := strings.Split(string(result.stdout), "\n")
			for i, line := range strings.Split(string(want.stdout), "\n") {
				if i >= len(observed) {
					test.Logf("caught by missing chunk result%d", i)
					break
				}
				if observed[i] != line {
					test.Logf("caught by %s at chunk byte%d", names[i], firstDifference(observed[i], line))
					break
				}
			}
		})
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go", goBinary, []string{cases}},
		{"native", fast, []string{cases}},
		{"actual original Node", "node", []string{"testdata/chunks_library.mjs", fork, cases}},
	} {
		started := time.Now()
		for i := 0; i < 3; i++ {
			result := execute(test, nil, side.command, side.args...)
			clean(test, side.name, result)
			equal(test, side.name, result.stdout, want.stdout)
		}
		test.Logf(
			"%s %.1f texts/s; three runs, startup and identical numeric UTF-16/chunk transport included",
			side.name,
			float64(3*len(inputs))/time.Since(started).Seconds(),
		)
	}
	test.Logf(
		"%d cases: %d physical documents; 4943 layout documents; all 65536 UTF-16 units, including lone surrogates; special-code sequences through length five; column/BOM/NUL/CRLF/tab edges",
		len(inputs),
		len(selection.Paths),
	)
}

// chunksGoOracle is the Go side, cohere with the driver and bridge its overlay lays over it, a product built ahead
// (#5qykzj5).
func chunksGoOracle(test *testing.T) string { test.Helper(); return oracles.ChunksGoOracle(test) }

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

// Parallel execution reserves memory: full event output and native mutant runs retain large corpus buffers.
func TestTokenizerEvents(t *testing.T) {
	parallelMarkdownMemory(t, 4)
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
	if os.Getenv("ADAMIC_EVENTS_SMOKE") != "" {
		inputs = [][]uint16{{}, {65}, {9}, {65, 9, 10, 13, 10, 0, 66}, {55296}, {55296, 56320}, {13, 9, 9, 10}}
		names = []string{"empty", "text", "tab", "mix", "surrogate", "astral", "lines"}
	}
	var batch strings.Builder
	for _, units := range inputs {
		if len(units) == 0 {
			batch.WriteString("-")
		} else {
			for i, unit := range units {
				if i > 0 {
					batch.WriteByte(',')
				}
				fmt.Fprint(&batch, unit)
			}
		}
		batch.WriteByte('\n')
	}
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_chunks/main.go")
	driver, err := filepath.Abs("testdata/events_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/events_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	transportBridge, err := filepath.Abs("testdata/events_transport.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_chunks.go"): bridge, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_event_transport.go"): transportBridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-chunks")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("Go events %v %s", err, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go events", want)
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
	main, err := filepath.Abs("testdata/events_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"actual original tokenizer", execute(t, nil, "node", "testdata/events_library.mjs", fork, cases)},
		{"native", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onJavaScriptBackend(t, program, cases)},
	} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want.stdout)
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, from, to string }{
		{"event rollback", "while(this.events.length > info.from) this.events.pop();", "while(this.events.length > info.from + 2) this.events.pop();"},
		{"virtual serialization", "if(!expandTabs && atTab) continue;", "if(expandTabs && atTab) continue;"},
		{"restored construct", "this.construct = info.construct;", "this.construct = 8;"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"inputChunks.ts", "tokenizerEvents.ts", "tokenArena.ts", "codec.ts", "testdata/events_probe.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "tokenizerEvents.ts" {
					if strings.Count(string(data), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/events_probe.ts"), cases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("survived")
			}
			observed := strings.Split(string(result.stdout), "\n")
			for i, line := range strings.Split(string(want.stdout), "\n") {
				if i >= len(observed) {
					t.Logf("caught by missing event result%d", i)
					break
				}
				if observed[i] != line {
					t.Logf("caught by %s at event byte%d", names[i], firstDifference(observed[i], line))
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
		{"actual Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"actual original Node", "node", []string{"testdata/events_library.mjs", fork, cases}},
	} {
		started := time.Now()
		for i := 0; i < 3; i++ {
			result := execute(t, nil, side.command, side.args...)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("%s %.1f texts/s; three runs, startup and identical numeric UTF-16/event transport included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
	}
	t.Logf("%d cases: %d physical documents; 4943 layout documents; all 65536 UTF-16 units, including lone surrogates; special-code sequences through length five; event/rollback/skip/fields/slice/CRLF/tab edges", len(inputs), files)
}

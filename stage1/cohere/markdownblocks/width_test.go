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
)

// Not parallel: the exhaustive scalar corpus and sanitizer builds have a large peak working set.
func TestMarkdownUnicodeWidths(t *testing.T) {
	widthDependencies := os.Getenv("ADAMIC_MARKDOWNWIDTH_DEPS")
	if widthDependencies == "" {
		t.Skip("set ADAMIC_MARKDOWNWIDTH_DEPS to an npm install of emoji-regex@10.6.0, get-east-asian-width@1.6.0 and narrow-emojis@0.0.3; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
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
	dir := t.TempDir()
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(batch.String()))
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_width/main.go")
	driver, e := filepath.Abs("testdata/width_go.go")
	if e != nil {
		t.Fatal(e)
	}
	overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver}})
	if e != nil {
		t.Fatal(e)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-width")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	command.Dir = cohere
	if output, e := command.CombinedOutput(); e != nil {
		t.Fatalf("Go width: %v\n%s", e, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go StringWidth", want)
	library := execute(t, nil, "node", "testdata/width_library.mjs", widthDependencies, cases)
	clean(t, "original dependencies", library)
	equal(t, "Go versus original dependencies", library.stdout, want.stdout)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	forkWidths := execute(t, nil, "node", "testdata/width_fork.mjs", fork, cases)
	clean(t, "actual fork group-fit widths", forkWidths)
	equal(t, "actual fork group-fit widths", forkWidths.stdout, want.stdout)
	main, e := filepath.Abs("testdata/width_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", answer}, {"source Node", onNode(t, main, cases)}, {"backend", onJavaScriptBackend(t, program, cases)}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			a := strings.Split(string(side.result.stdout), "\n")
			b := strings.Split(string(want.stdout), "\n")
			for i := range b {
				if i >= len(a) || a[i] != b[i] {
					t.Fatalf("%s mismatch %q got %q want %q", side.name, inputs[i].Name, a[i], b[i])
				}
			}
		}
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, from, to string }{{"East Asian width", "inRanges(code, wideRanges) ? 2 : 1", "inRanges(code, wideRanges) ? 1 : 1"}, {"narrow emoji", "=== end ? 1 : 2", "=== end ? 2 : 2"}, {"ASCII DEL shortcut", "code > 0x7f", "code > 0x7e"}} {
		t.Run(m.name, func(t *testing.T) {
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
			mutant := lowered(t, filepath.Join(scratch, "testdata/width_probe.ts"))
			r := nativelyRun(t, mutant, cases)
			clean(t, m.name, r)
			if bytes.Equal(r.stdout, want.stdout) {
				t.Fatal("survived")
			}
			t.Log("caught by observable width output")
		})
	}
	if os.Getenv("ADAMIC_MARKDOWN_BENCH") != "" {
		fast := filepath.Join(dir, "fast")
		if e := native.Build(native.C(program), fast, native.Options{}); e != nil {
			t.Fatal(e)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{cases}}, {"original Node dependencies", "node", []string{"testdata/width_library.mjs", widthDependencies, cases}}} {
			started := time.Now()
			for i := 0; i < 3; i++ {
				r := execute(t, nil, side.command, side.args...)
				clean(t, side.name, r)
				equal(t, side.name, r.stdout, want.stdout)
			}
			t.Logf("%s %.1f texts/s, startup, decoding and output included", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
		}
	} else {
		result := releaseRun(t, program, cases)
		clean(t, "release width", result)
		equal(t, "release width", result.stdout, want.stdout)
	}
	command = bounded(t, "go", "run", "./stage1/cohere/markdownblocks/tools/generate_width", "-check", "-formatter", cohereFormatter(t))
	command.Dir = root
	if output, e := command.CombinedOutput(); e != nil {
		t.Fatalf("regeneration %v %s", e, output)
	}
	t.Logf("%d widths: %d repository files, source lines, every Unicode scalar and generated sequences", len(inputs), files)
}

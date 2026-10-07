package markdownblocks

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Parallel execution reserves memory: the full event transport and multiple independently computed trees are large.
func TestNativeMdastConstruction(t *testing.T) {
	parallelMarkdownMemory(t, 3)
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	inputs, files := blockCorpus(t, root, "whitespace")
	smokeStart := len(inputs)
	for _, text := range []string{"", "~~a\nb~~\n", "$$ meta\na\nb\n$$\n", "$ a $ and $b$\n", "[[target]] {{x}} {% x\ny %}\n", "[^a]: x\n\n[^a]\n", "[x]: /u \"title\"\n\n![a &amp; b][x]\n", "| a | b |\n|:-:|--:|\n|`a\\|b`|~~x~~|\n", "- [x] \n- [ ] a\n\n  > b\n", "<a href=\"x\">t</a>\n", "a &NotEqualTilde; &#x1F600; &#0; \\*\n"} {
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
	for _, text := range []string{"---\n---", "---\n...", "+++\n+++tail", "---yaml\n", "--- yaml\n---suffix\nx", "---\n😀\n...\n中"} {
		inputs = append(inputs, auditInput{Name: "generated/mdast-front-edge/" + text, Text: text})
	}
	if os.Getenv("ADAMIC_MDAST_SMOKE") != "" {
		inputs = append(inputs[smokeStart:smokeStart+11], inputs[len(inputs)-6:]...)
		files = 0
	}
	dir := t.TempDir()
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, i := range inputs {
		if e := enc.Encode(i); e != nil {
			t.Fatal(e)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, b.Bytes())
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_mdast/main.go")
	replace := map[string]string{}
	for _, p := range []struct{ target, source string }{{mainPath, "testdata/mdast_go.go"}, {filepath.Join(cohere, "internal/format/markdown/mdast/adamic_mdast.go"), "testdata/mdast_bridge.go"}, {filepath.Join(cohere, "internal/format/markdown/micromark/adamic_events.go"), "testdata/events_transport.go"}} {
		source, e := filepath.Abs(p.source)
		if e != nil {
			t.Fatal(e)
		}
		replace[p.target] = source
	}
	overlay, e := json.Marshal(map[string]any{"Replace": replace})
	if e != nil {
		t.Fatal(e)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-mdast")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("Go mdast %v %s", e, output)
	}
	transport := filepath.Join(dir, "events.txt")
	want := execute(t, nil, goBinary, cases, transport)
	clean(t, "actual Go mdast", want)
	if keep := os.Getenv("ADAMIC_MDAST_KEEP"); keep != "" {
		if e := os.MkdirAll(keep, 0755); e != nil {
			t.Fatal(e)
		}
		data, e := os.ReadFile(transport)
		if e != nil {
			t.Fatal(e)
		}
		write(t, filepath.Join(keep, "events.txt"), data)
		write(t, filepath.Join(keep, "want.txt"), want.stdout)
		write(t, filepath.Join(keep, "cases.jsonl"), b.Bytes())
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, e := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	pinned, e := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if e != nil {
		t.Fatal(e)
	}
	equal(t, "pinned fork", installed, pinned)
	main, e := filepath.Abs("testdata/mdast_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, transport)
	for _, side := range []struct {
		name   string
		result run
	}{{"actual Go compiler from events", execute(t, nil, goBinary, "--events", transport)}, {"actual original compiler from events", execute(t, nil, "node", "testdata/mdast_library.mjs", fork, "--events", transport)}, {"actual original mdast", execute(t, nil, "node", "testdata/mdast_library.mjs", fork, cases)}, {"native", answer}, {"source Node", onNode(t, main, transport)}, {"backend", onJavaScriptBackend(t, program, transport)}} {
		clean(t, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			a := strings.Split(string(side.result.stdout), "\n")
			z := strings.Split(string(want.stdout), "\n")
			if len(a) != len(z) {
				t.Fatalf("%s tree count%d/%d", side.name, len(a), len(z))
			}
			for i, row := range z {
				if row != a[i] {
					offset := firstDifference(a[i], row)
					t.Fatalf("%s %s byte%d got%q want%q", side.name, inputs[i].Name, offset, a[i][max(0, offset-100):min(len(a[i]), offset+100)], row[max(0, offset-100):min(len(row), offset+100)])
				}
			}
		}
	}
	if report := leaks(t, program, binary, transport); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"text construction", "mdastCompile.ts", "tail.value += this.serialize(token);", "tail.value += '!' + this.serialize(token);"},
		{"list spread", "mdastCompile.ts", "this.top().spread = token.spread;", "this.top().spread = !token.spread;"},
		{"front matter fallback", "parseFrontMatter.ts", "language === 'yaml'", "language === 'toml'"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
				t.Fatal(e)
			}
			for _, file := range []string{"tokenizerEvents.ts", "tokenArena.ts", "tokenSource.ts", "mdastNode.ts", "inputChunks.ts", "codec.ts", "mdastArena.ts", "mdastCompile.ts", "parseFrontMatter.ts", "identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "decodeString.ts", "upperEntityNames.ts", "upperEntityValues.ts", "lowerEntityNames.ts", "lowerEntityValues.ts", "testdata/mdast_probe.ts"} {
				data, e := os.ReadFile(file)
				if e != nil {
					t.Fatal(e)
				}
				if file == m.file {
					if !strings.Contains(string(data), m.from) {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/mdast_probe.ts"), transport)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want.stdout) {
				t.Fatal("survived")
			}
			a := strings.Split(string(result.stdout), "\n")
			z := strings.Split(string(want.stdout), "\n")
			for i, row := range z {
				if i >= len(a) || a[i] != row {
					if i >= len(a) {
						t.Logf("caught by missing tree%d", i)
					} else {
						t.Logf("caught by %s at byte%d", inputs[i].Name, firstDifference(a[i], row))
					}
					break
				}
			}
		})
	}
	fast := filepath.Join(dir, "fast")
	if e := native.Build(native.C(program), fast, native.Options{}); e != nil {
		t.Fatal(e)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go construction", goBinary, []string{"--events", transport}}, {"native construction", fast, []string{transport}}, {"actual original Node construction", "node", []string{"testdata/mdast_library.mjs", fork, "--events", transport}},
	} {
		started := time.Now()
		for i := 0; i < 3; i++ {
			result := execute(t, nil, side.command, side.args...)
			clean(t, side.name, result)
			equal(t, side.name, result.stdout, want.stdout)
		}
		t.Logf("%s %.1f texts/s; three runs, identical event input and tree output, startup/I/O included, lexical parsing excluded", side.name, float64(3*len(inputs))/time.Since(started).Seconds())
	}
	t.Logf("%d full constructed trees, %d physical files; actual Go and pinned fork, native/source/backend, sanitizer and leaks agree on every field and coordinate; %d bytes", len(inputs), files, len(want.stdout))
}

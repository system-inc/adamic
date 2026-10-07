package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMdastMalformedEvents(t *testing.T) {
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_mdast_errors/main.go")
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
	goBinary := filepath.Join(dir, "go-errors")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("Go errors %v %s", e, output)
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	main, e := filepath.Abs("testdata/mdast_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	for _, name := range []string{"unclosed", "not-open", "mismatch"} {
		t.Run(name, func(t *testing.T) {
			truth := execute(t, nil, goBinary, "--error", name)
			clean(t, "Go error as value", truth)
			original := execute(t, nil, "node", "testdata/mdast_library.mjs", fork, "--error", name)
			clean(t, "fork error as value", original)
			forkMessages := map[string]string{
				"unclosed": "Cannot close document, a token (`paragraph`, 1:1-1:1) is still open\n",
				"not-open": "Cannot close `paragraph` (1:1-1:1): it’s not open\n",
				"mismatch": "Cannot close `strong` (1:1-1:1): a different token (`paragraph`, 1:1-1:1) is open\n",
			}
			equal(t, "retained fork diagnostic", original.stdout, []byte(forkMessages[name]))

			cases := "gaps/event_" + name + ".txt"

			answer, _ := natively(t, program, cases)
			for _, side := range []run{answer, onNode(t, main, cases), onJavaScriptBackend(t, program, cases)} {
				if side.exitCode != 70 {
					t.Fatalf("%s expected error exit70 got%d", name, side.exitCode)
				}
				equal(t, "empty pre-error stdout", side.stdout, nil)
				equal(t, "normalized actual error", side.stderr, []byte("adamic: panic: "+string(truth.stdout)))
			}
			scratch := t.TempDir()
			if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
				t.Fatal(e)
			}
			from, to := "Cannot close document, a token", "Cannot close document, token"
			if name == "not-open" {
				from, to = "it’s not open", "it’s already closed"
			}
			if name == "mismatch" {
				from, to = "a different token", "another token"
			}
			for _, file := range []string{"tokenizerEvents.ts", "tokenArena.ts", "tokenSource.ts", "mdastNode.ts", "inputChunks.ts", "codec.ts", "mdastArena.ts", "mdastCompile.ts", "parseFrontMatter.ts", "identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "decodeString.ts", "upperEntityNames.ts", "upperEntityValues.ts", "lowerEntityNames.ts", "lowerEntityValues.ts", "testdata/mdast_probe.ts"} {
				data, e := os.ReadFile(file)
				if e != nil {
					t.Fatal(e)
				}
				if file == "mdastCompile.ts" {
					if strings.Count(string(data), from) != 1 {
						t.Fatal("error mutation anchor")
					}
					data = []byte(strings.Replace(string(data), from, to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			mutant := onNode(t, filepath.Join(scratch, "testdata/mdast_probe.ts"), cases)
			if mutant.exitCode != 70 {
				t.Fatalf("error mutant did not reach expected error: %d", mutant.exitCode)
			}
			equal(t, "error mutant stdout", mutant.stdout, nil)
			if bytes.Equal(mutant.stderr, []byte("adamic: panic: "+string(truth.stdout))) {
				t.Fatal("error mutant survived")
			}
			t.Logf("stderr-only message mutant caught; Go=%q fork=%q", truth.stdout, original.stdout)
			t.Logf("%s", truth.stdout)
		})
	}
}

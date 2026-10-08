package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMdastIdentifierWitnesses(t *testing.T) {
	parallelMarkdownMemory(t, 1)
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_mdast_gap/main.go")
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
	binaryGo := filepath.Join(dir, "go-gap")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", binaryGo, mainPath)
	build.Dir = cohere
	if output, e := combinedOutput(build); e != nil {
		t.Fatalf("Go %v %s", e, output)
	}
	transport := filepath.Join(dir, "events.txt")
	goResult := execute(t, nil, binaryGo, "gaps/identifier_case.jsonl", transport)
	clean(t, "actual Go", goResult)
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
	equal(t, "pinned gap fork", installed, pinned)
	original := execute(t, nil, "node", "testdata/mdast_library.mjs", fork, "gaps/identifier_case.jsonl")
	clean(t, "actual fork", original)
	if bytes.Equal(goResult.stdout, original.stdout) {
		t.Fatal("the identifier witnesses closed; update GAPS.md")
	}
	if os.Getenv("ADAMIC_MDAST_WRITE_WITNESSES") != "" {
		write(t, "gaps/identifier_case_go.txt", goResult.stdout)
		write(t, "gaps/identifier_case_fork.txt", original.stdout)
	}
	for _, side := range []struct {
		name, path string
		result     run
	}{{"Go", "gaps/identifier_case_go.txt", goResult}, {"fork", "gaps/identifier_case_fork.txt", original}} {
		golden, e := os.ReadFile(side.path)
		if e != nil {
			t.Fatal(e)
		}
		equal(t, side.name+" witness", side.result.stdout, golden)
	}
	main, e := filepath.Abs("testdata/mdast_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, transport)
	for _, side := range []run{answer, onNode(t, main, transport), onJavaScriptBackend(t, program, transport)} {
		clean(t, "native Go-held identifier witness", side)
		equal(t, "native Go-held identifier witness", side.stdout, goResult.stdout)
	}
	if report := leaks(t, program, binary, transport); report != "" {
		t.Fatal(report)
	}
	t.Log("Three full-tree Go/fork casing disagreements retained; native/source/backend hold Go, sanitizers and leaks pass")
}

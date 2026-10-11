package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: fixed gaps/identifier_case_{go,fork}.txt witness files and shared markdownMemory budget; existing helper controls parallel execution.
func TestMdastIdentifierWitnesses(t *testing.T) {
	parallelMarkdownMemory(t, 1)
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	// The mdast grain's Go oracle is this witness's program: the same three overlay replacements (testdata/mdast_go.go as
	// a synthesized main, mdast_bridge.go and events_transport.go) built from cohere, so the witness reads that product.
	binaryGo := grainProduct(t, "mdast", "go").goBinary
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

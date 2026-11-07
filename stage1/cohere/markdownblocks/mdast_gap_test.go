package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: fixed gaps/identifier_case_{go,fork}.txt witness files and shared markdownMemory budget; existing helper controls parallel execution.
func TestMdastIdentifierWitnesses(test *testing.T) {
	parallelMarkdownMemory(test, 1)
	root, e := filepath.Abs(repository)
	if e != nil {
		test.Fatal(e)
	}
	dir := test.TempDir()
	cohere := filepath.Join(root, "cohere")
	// The mdast grain's Go oracle is this witness's program: the same three overlay replacements (testdata/mdast_go.go
	// as a synthesized main, mdast_bridge.go and events_transport.go) built from cohere, so the witness reads that
	// product.
	binaryGo := grainProduct(test, "mdast", "go").goBinary
	transport := filepath.Join(dir, "events.txt")
	goResult := execute(test, nil, binaryGo, "gaps/identifier_case.jsonl", transport)
	clean(test, "actual Go", goResult)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, e := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if e != nil {
		test.Fatal(e)
	}
	pinned, e := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if e != nil {
		test.Fatal(e)
	}
	equal(test, "pinned gap fork", installed, pinned)
	original := execute(test, nil, "node", "testdata/mdast_library.mjs", fork, "gaps/identifier_case.jsonl")
	clean(test, "actual fork", original)
	if bytes.Equal(goResult.stdout, original.stdout) {
		test.Fatal("the identifier witnesses closed; update GAPS.md")
	}
	if os.Getenv("ADAMIC_MDAST_WRITE_WITNESSES") != "" {
		write(test, "gaps/identifier_case_go.txt", goResult.stdout)
		write(test, "gaps/identifier_case_fork.txt", original.stdout)
	}
	for _, side := range []struct {
		name, path string
		result     run
	}{{"Go", "gaps/identifier_case_go.txt", goResult}, {"fork", "gaps/identifier_case_fork.txt", original}} {
		golden, e := os.ReadFile(side.path)
		if e != nil {
			test.Fatal(e)
		}
		equal(test, side.name+" witness", side.result.stdout, golden)
	}
	main, e := filepath.Abs("testdata/mdast_probe.ts")
	if e != nil {
		test.Fatal(e)
	}
	program := lowered(test, main)
	// The mdast grain's sanitized port is this program too, so the witness reads that product.
	binary := grainProduct(test, "mdast", "sanitized").sanitized
	answer := sanitizedRun(test, binary, transport)
	for _, side := range []run{answer, onNode(test, main, transport), onJavaScriptBackend(test, program, transport)} {
		clean(test, "native Go-held identifier witness", side)
		equal(test, "native Go-held identifier witness", side.stdout, goResult.stdout)
	}
	if report := leaks(test, program, binary, transport); report != "" {
		test.Fatal(report)
	}
	test.Log(
		"Three full-tree Go/fork casing disagreements retained; native/source/backend hold Go, sanitizers and leaks pass",
	)
}

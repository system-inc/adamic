package markdownblocks

import (
	"path/filepath"
	"testing"
)

// Optional declarations start as undefined on every backend, as Node decides.
func TestOptionalStringInitializationWitness(t *testing.T) {
	parallelMarkdownMemory(t, 1)
	path, err := filepath.Abs("gaps/6_uninitialized_optional_string.ts")
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	clean(t, "Node truth", source)
	equal(t, "Node truth", source.stdout, []byte("missing\n"))
	program := lowered(t, path)
	answer, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", answer}, {"backend", onJavaScriptBackend(t, program)}} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, source.stdout)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("Node, native and JavaScript backend agree on missing.")
}

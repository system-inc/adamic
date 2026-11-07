package markdownblocks

import (
	"path/filepath"
	"testing"
)

// Optional declarations start as undefined on every backend, as Node decides.
// Not parallel: shared markdownMemory configuration; existing helper controls parallel execution.
func TestOptionalStringInitializationWitness(test *testing.T) {
	parallelMarkdownMemory(test, 1)
	path, err := filepath.Abs("gaps/6_uninitialized_optional_string.ts")
	if err != nil {
		test.Fatal(err)
	}
	source := onNode(test, path)
	clean(test, "Node truth", source)
	equal(test, "Node truth", source.stdout, []byte("missing\n"))
	program := lowered(test, path)
	binary := filepath.Join(gapSanitized(test, "gaps/6_uninitialized_optional_string.ts"), "port")
	answer := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", answer}, {"backend", onJavaScriptBackend(test, program)}} {
		clean(test, side.name, side.result)
		equal(test, side.name, side.result.stdout, source.stdout)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("Node, native and JavaScript backend agree on missing.")
}

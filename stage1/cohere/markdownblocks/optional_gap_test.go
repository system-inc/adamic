package markdownblocks

import (
	"path/filepath"
	"testing"
)

// Explicitly recorded wrong behavior, never an oracle for the production decoder.
// Delete this observation and the workaround when the compiler gap closes.
func TestOptionalStringInitializationWitness(t *testing.T) {
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
	}{{"known native miscompile", answer}, {"known backend miscompile", onJavaScriptBackend(t, program)}} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, []byte("present\n"))
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("Node missing; native and JavaScript backend present. Decoder uses an explicit match flag.")
}

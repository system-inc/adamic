package markdownblocks

import (
	"path/filepath"
	"testing"
)

// Explicit gap observation, not the oracle for the production map, which uses push.
// Not parallel: shared markdownMemory configuration and artifacts build cache; existing helper controls parallel execution.
func TestArrayGrowthWitness(t *testing.T) {
	parallelMarkdownMemory(t, 1)
	path, e := filepath.Abs("gaps/13_array_growth.ts")
	if e != nil {
		t.Fatal(e)
	}
	source := onNode(t, path)
	clean(t, "Node truth", source)
	equal(t, "Node truth", source.stdout, []byte("1\n"))
	program := lowered(t, path)
	answer, _ := natively(t, program)
	if answer.exitCode != 70 {
		t.Fatalf("array growth gap changed: exit %d stderr %s", answer.exitCode, answer.stderr)
	}
	equal(t, "known native panic", answer.stderr, []byte("adamic: panic: index 0 is outside an array of length 0\n"))
	equal(t, "known native panic stdout", answer.stdout, nil)
	backend := onJavaScriptBackend(t, program)
	if backend.exitCode != 70 {
		t.Fatalf("backend growth gap changed: %d %s", backend.exitCode, backend.stderr)
	}
	equal(t, "known backend panic", backend.stderr, answer.stderr)
	equal(t, "known backend stdout", backend.stdout, nil)
	t.Log("Node extends the array; both compiled backends panic. The production map appends with push.")
}

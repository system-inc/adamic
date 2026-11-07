package markdownblocks

import (
	"path/filepath"
	"testing"
)

// Explicit gap observation, not the oracle for the production map, which uses push.
// Not parallel: shared markdownMemory configuration; existing helper controls parallel execution.
func TestArrayGrowthWitness(test *testing.T) {
	parallelMarkdownMemory(test, 1)
	path, absolutePathError := filepath.Abs("gaps/13_array_growth.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source := onNode(test, path)
	clean(test, "Node truth", source)
	equal(test, "Node truth", source.stdout, []byte("1\n"))
	program := lowered(test, path)
	answer := sanitizedRun(test, filepath.Join(gapSanitized(test, "gaps/13_array_growth.ts"), "port"))
	if answer.exitCode != 70 {
		test.Fatalf("array growth gap changed: exit %d stderr %s", answer.exitCode, answer.stderr)
	}
	equal(test, "known native panic", answer.stderr, []byte("adamic: panic: index 0 is outside an array of length 0\n"))
	equal(test, "known native panic stdout", answer.stdout, nil)
	backend := onJavaScriptBackend(test, program)
	if backend.exitCode != 70 {
		test.Fatalf("backend growth gap changed: %d %s", backend.exitCode, backend.stderr)
	}
	equal(test, "known backend panic", backend.stderr, answer.stderr)
	equal(test, "known backend stdout", backend.stdout, nil)
	test.Log("Node extends the array; both compiled backends panic. The production map appends with push.")
}

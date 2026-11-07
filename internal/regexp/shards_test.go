package regexp

import (
	"fmt"
	"os"
	"testing"
)

var runtimeReferenceShardNames = []string{
	"stop on embedded NUL", "drop Other_ID_Start", "drop Other_ID_Continue",
	"accept Pattern_Syntax letter", "wrap oversized Unicode escape",
	"silently accept clamped reversed bounds", "classify divergence as SyntaxError",
}

func checkRuntimeReferenceMutantCoverage(shards []runtimeReferenceMutant) error {
	seen := map[string]int{}
	for _, shard := range shards {
		seen[shard.name]++
	}
	for _, name := range runtimeReferenceShardNames {
		if seen[name] != 1 {
			return fmt.Errorf("mutant %q covered %d times, want 1", name, seen[name])
		}
	}
	if len(seen) != len(runtimeReferenceShardNames) {
		return fmt.Errorf("unexpected mutant shard")
	}
	return nil
}

func TestRuntimeReferenceShardCoverage(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("parser.go")
	if readError != nil {
		test.Fatal(readError)
	}
	shards := runtimeReferenceMutantShards(test, string(source))
	if checkRuntimeReferenceMutantCoverageError := checkRuntimeReferenceMutantCoverage(
		shards,
	); checkRuntimeReferenceMutantCoverageError != nil {
		test.Fatal(checkRuntimeReferenceMutantCoverageError)
	}
	if checkRuntimeReferenceMutantCoverageError := checkRuntimeReferenceMutantCoverage(
		shards[1:],
	); checkRuntimeReferenceMutantCoverageError == nil {
		test.Fatal("dropped mutant survived")
	}
	duplicated := append(append([]runtimeReferenceMutant{}, shards...), shards[0])
	if checkRuntimeReferenceMutantCoverageError := checkRuntimeReferenceMutantCoverage(
		duplicated,
	); checkRuntimeReferenceMutantCoverageError == nil {
		test.Fatal("duplicate mutant survived")
	}
	test.Logf("all %d unchanged mutants scheduled exactly once", len(shards))
}

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

func TestRuntimeReferenceShardCoverage(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("parser.go")
	if err != nil {
		t.Fatal(err)
	}
	shards := runtimeReferenceMutantShards(t, string(source))
	if err := checkRuntimeReferenceMutantCoverage(shards); err != nil {
		t.Fatal(err)
	}
	if err := checkRuntimeReferenceMutantCoverage(shards[1:]); err == nil {
		t.Fatal("dropped mutant survived")
	}
	duplicated := append(append([]runtimeReferenceMutant{}, shards...), shards[0])
	if err := checkRuntimeReferenceMutantCoverage(duplicated); err == nil {
		t.Fatal("duplicate mutant survived")
	}
	t.Logf("all %d unchanged mutants scheduled exactly once", len(shards))
}

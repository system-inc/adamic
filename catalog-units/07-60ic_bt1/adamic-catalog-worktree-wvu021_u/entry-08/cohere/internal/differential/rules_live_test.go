package differential

import (
	"context"
	"os"
	"sort"
	"testing"
)

// A live check against a real cohere binary, skipped unless one is named.
//
// The parse has to be proven against actual output rather than a fixture, because a fixture is a
// copy of what I believed the format was, and the failure this guards against is precisely the
// format not matching that belief. Run with:
//
//	COHERE_BINARY=/tmp/cohere-head COHERE_TREE=~/Projects/ahra go test -run TestCompiledRulesAgainstARealBinary ./internal/differential/
func TestCompiledRulesAgainstARealBinary(t *testing.T) {
	t.Parallel()
	binary := os.Getenv("COHERE_BINARY")
	tree := os.Getenv("COHERE_TREE")
	explainFile := os.Getenv("COHERE_EXPLAIN_FILE")
	if binary == "" || tree == "" || explainFile == "" {
		t.Skip("set COHERE_BINARY, COHERE_TREE and COHERE_EXPLAIN_FILE to run this against a real build")
	}

	names, err := CompiledRulesOf(context.Background(), GateCommand{
		Name: "cohere", Program: binary, Directory: tree,
	}, explainFile)
	if err != nil {
		t.Fatalf("could not read the rule list: %v", err)
	}

	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	t.Logf("%d rules: %v", len(sorted), sorted)

	// The count guard inside CompiledRulesOf already compared this against the binary's own
	// coverage line, so reaching here means they agreed. This asserts the population is plausible,
	// which catches the case where the binary reports a tiny number and the parse dutifully agrees.
	if len(names) < 10 {
		t.Fatalf("only %d rules parsed, which is too few to be a real build", len(names))
	}
}

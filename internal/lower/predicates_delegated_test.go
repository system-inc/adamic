package lower

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Probe the delegated declaration directly. A refusal of a lying helper earlier
// in a whole-file pass must not mask an unsound proof of its caller.
func delegatedProofFixture(t *testing.T, name, function, reason string) {
	t.Helper()
	program, err := load.Load([]string{"testdata/predicates_delegated/" + name + ".a"})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	var predicate *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypePredicate && node.Parent.Name() != nil && node.Parent.Name().Text() == function {
			predicate = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if predicate == nil {
		t.Fatal("missing declaration")
	}
	_, err = l.provePredicate(predicate)
	if reason == "" {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("wanted %q, got %v", reason, err)
	}
}
func TestDelegatedProofLyingHelper(t *testing.T) {
	t.Parallel()
	delegatedProofFixture(t, "lying", "delegated", "unconverged or opaque helper")
}
func TestDelegatedProofNegation(t *testing.T) {
	t.Parallel()
	delegatedProofFixture(t, "negation", "delegated", "false return does not exclude")
}
func TestDelegatedProofEffect(t *testing.T) {
	t.Parallel()
	delegatedProofFixture(t, "effect", "delegated", "unconverged or opaque helper")
}
func TestDelegatedProofIdentity(t *testing.T) {
	t.Parallel()
	delegatedProofFixture(t, "identity", "delegated", "different argument identity")
}
func TestDelegatedProofSeedlessCycle(t *testing.T) {
	t.Parallel()
	// Both summaries would establish this annotation if it were installed as a
	// seed. With empty summaries, neither helper has independent evidence.
	source := `function first(value: unknown): value is string {return second(value);} function second(value:unknown):value is string{return first(value);}`
	path := t.TempDir() + "/cycle.a"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := &lowering{program: program, checker: checked}
	for _, statement := range program.Files()[0].Statements.Nodes {
		if _, err = l.provePredicate(statement.Type()); err == nil || !strings.Contains(err.Error(), "unconverged") {
			t.Fatalf("annotation seeded recursive proof: %v", err)
		}
	}
}
func TestDelegatedProofPositiveComposition(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, function string }{{"primitive", "isOnlyString"}, {"literals", "isPositive"}, {"switch", "isSelected"}, {"switch", "isFirst"}, {"paths", "isConditionalString"}} {
		delegatedProofFixture(t, probe.name, probe.function, "")
	}
}

func TestDelegatedProofExtraCondition(t *testing.T) {
	t.Parallel()
	delegatedProofFixture(t, "mixed_false", "delegated", "false return does not exclude")
}

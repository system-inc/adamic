package lower

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// predicateCannotMatch skips, before the checker sees them, exactly the identifiers that can't be
// an identifier argument: those spelled otherwise. A declaration of the argument, a this, an
// access and an identifier spelled the same still reach isMatchingReference. On tsc it took
// mapToTypeNodes from 21.7 s to 0.5 s and the src/compiler latent census from 2,308 s to 1,629 s
// with all 28,637 findings unchanged (#p61gv45). Mutants: skipping every identifier turns
// TestPredicateUsesBelongToEachCall and the oracle's predicate refusals red; never skipping turns
// this test red.
func TestPredicatePrefilterSkipsOnlyOtherNames(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "const value: string | number = 1; const other = { value: 2 }; function f(): void { console.log(`${value}${other.value}`); }\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	var argument, declaration, sameName, otherName, propertyName *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch {
		case node.Kind == ast.KindVariableDeclaration && declaration == nil:
			declaration = node
		case node.Kind == ast.KindIdentifier && node.Text() == "value" && node.Parent.Kind == ast.KindTemplateSpan && argument == nil:
			argument = node
		case node.Kind == ast.KindIdentifier && node.Text() == "value" && node.Parent.Kind == ast.KindPropertyAccessExpression:
			sameName = node
		case node.Kind == ast.KindIdentifier && node.Text() == "other" && node.Parent.Kind == ast.KindPropertyAccessExpression:
			otherName = node
			propertyName = node.Parent
		}
		node.ForEachChild(visit)
		return false
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	if argument == nil || declaration == nil || sameName == nil || otherName == nil || propertyName == nil {
		t.Fatalf("probe nodes not found: %v %v %v %v %v", argument, declaration, sameName, otherName, propertyName)
	}
	for _, probe := range []struct {
		name  string
		node  *ast.Node
		skips bool
	}{
		{"another name", otherName, true},
		{"the same name", sameName, false},
		{"its declaration", declaration, false},
		{"an access", propertyName, false},
	} {
		if got := predicateCannotMatch(probe.node, argument); got != probe.skips {
			t.Errorf("%s: skipped %v, want %v", probe.name, got, probe.skips)
		}
	}
	if predicateCannotMatch(otherName, propertyName) {
		t.Error("an access argument skipped an identifier: only an identifier argument filters by name")
	}
}

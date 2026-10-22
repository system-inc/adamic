package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Reduced from builder.ts:2229. A qualified name is not the identifier const.
func TestCensusQualifiedCast(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "qualified.a")
	source := censusPanicSource(t, "qualified.a")
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "types.a"), []byte(censusPanicSource(t, "types.a")), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	var file *ast.SourceFile
	for _, candidate := range program.Files() {
		if program.FileName(candidate) == path {
			file = candidate
		}
	}
	if file == nil {
		t.Fatal("missing main file")
	}
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	var cast *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindAsExpression {
			cast = n
		}
		n.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if cast == nil {
		t.Fatal("missing cast")
	}
	if _, err := l.castProof(cast); err != nil {
		t.Fatal(err)
	}
	if err := l.refuseWidening(cast); err != nil {
		t.Fatal(err)
	}
}

// A failing generic validation inside an arrow must preserve its enclosing stack.
func TestCensusGenericRefusalInsideClosure(t *testing.T) {
	t.Parallel()
	source := censusPanicSource(t, "generic_closure.a")
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "instantiating a generic function makes a value of type string | undefined written where string is read") {
		t.Fatalf("want named generic mutation refusal, got %v", err)
	}
}

// Reduced from declarations/diagnostics.ts:606, whose table uses enum keys.
func TestCensusComputedRelation(t *testing.T) {
	t.Parallel()
	source := censusPanicSource(t, "computed_relation.a")
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "optional field arrow.suggestion has no proven compatible presence/type") {
		t.Fatalf("want named optional-field proof refusal, got %v", err)
	}
}

func censusPanicSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "census_panics", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

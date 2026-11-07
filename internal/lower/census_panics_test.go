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
func TestCensusQualifiedCast(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "qualified.a")
	source := censusPanicSource(test, "qualified.a")
	if writeError := os.WriteFile(
		filepath.Join(filepath.Dir(path), "types.a"),
		[]byte(censusPanicSource(test, "types.a")),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	var file *ast.SourceFile
	for _, candidate := range program.Files() {
		if program.FileName(candidate) == path {
			file = candidate
		}
	}
	if file == nil {
		test.Fatal("missing main file")
	}
	checked, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checked, result: &ir.Program{}}
	var cast *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindAsExpression {
			cast = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if cast == nil {
		test.Fatal("missing cast")
	}
	if _, castProofError := lowering.castProof(cast); castProofError != nil {
		test.Fatal(castProofError)
	}
	if refuseWideningError := lowering.refuseWidening(cast); refuseWideningError != nil {
		test.Fatal(refuseWideningError)
	}
}

// A generic call inside an arrow that binds string[] to T extends (string | undefined)[] is refused at
// its argument, the binding's relation site (#xywvn06), as append(result as (string | undefined)[], ...)
// would be. The instantiation's own validation, which named the write in the body, stays behind it as a
// witness the binding no longer lets this program reach.
func TestCensusGenericRefusalInsideClosure(test *testing.T) {
	test.Parallel()
	source := censusPanicSource(test, "generic_closure.a")
	_, lowerSourceError := lowerSource(test, source)
	if lowerSourceError == nil || !strings.Contains(
		lowerSourceError.Error(),
		"main.a:7:35: Adamic 0.1 refuses a value of type string[] seen as (string | undefined)[], which can write string | undefined where string is read",
	) {
		test.Fatalf("want the binding's invariant-mutable refusal, got %v", lowerSourceError)
	}
}

// Reduced from declarations/diagnostics.ts:606, whose table uses enum keys.
func TestCensusComputedRelation(test *testing.T) {
	test.Parallel()
	source := censusPanicSource(test, "computed_relation.a")
	_, lowerSourceError := lowerSource(test, source)
	if lowerSourceError == nil ||
		!strings.Contains(
			lowerSourceError.Error(),
			"optional field arrow.suggestion has no proven compatible presence/type",
		) {
		test.Fatalf("want named optional-field proof refusal, got %v", lowerSourceError)
	}
}

func censusPanicSource(test *testing.T, name string) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata", "census_panics", name))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

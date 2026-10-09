package type_checking

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// TestCommentRangeFactoryStaysUntouched holds what lets every GetCommentsInRange call share one
// factory: the scanner writes nothing to it. It scans trailing and leading comments of both kinds
// through the shared factory, checks they were all found, and then that the factory is still the
// zero value. If the scanner ever starts counting or pooling on its factory, this fails, and the
// sharing has to become one factory per caller.
func TestCommentRangeFactoryStaysUntouched(t *testing.T) {
	t.Parallel()
	text := "a /* one */ // two\n/* three */ b;"
	fileName := tspath.NormalizePath("/repository/source/Thing.ts")
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName),
		PathKey:  tspath.PathKey(fileName),
	}, text, core.ScriptKindTS)
	if sourceFile == nil {
		t.Fatal("could not parse")
	}
	var found []string
	for comment := range GetCommentsInRange(sourceFile, core.NewTextRange(1, len(text))) {
		found = append(found, text[comment.Pos():comment.End()])
	}
	if want := []string{"/* one */", "// two", "/* three */"}; !reflect.DeepEqual(found, want) {
		t.Fatalf("found %q, want %q", found, want)
	}
	if !reflect.DeepEqual(commentRangeFactory, ast.NodeFactory{}) {
		t.Fatal("the scanner wrote to the shared comment range factory, so sharing it across goroutines is a race")
	}
}

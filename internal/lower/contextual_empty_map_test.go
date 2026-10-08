package lower

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestContextualEmptyMap(t *testing.T) {
	source := `function rows<T>(xs: T[]): T[][] { return xs.map(() => []); }
 const numbers = rows([1,2]);
 const objects = rows([{ name: "x" }]);`
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindArrayLiteralExpression || n.Kind == ast.KindArrowFunction || n.Kind == ast.KindCallExpression {
			ctx := c.GetContextualType(n, checker.ContextFlagsNone)
			text := "nil"
			if ctx != nil {
				text = c.TypeToString(ctx)
			}
			t.Logf("%s own=%s context=%s", program.Where(n), c.TypeToString(c.GetTypeAtLocation(n)), text)
		}
		return n.ForEachChild(visit)
	}
	program.Files()[0].Node.ForEachChild(visit)
	_, err = Lower(context.Background(), program)
	if err == nil {
		t.Fatal("expected existing map-result invariance refusal")
	}
}

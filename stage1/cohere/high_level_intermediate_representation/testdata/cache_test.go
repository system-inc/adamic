//go:build lintoracle

package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

func TestStage1FileCacheOracle(t *testing.T) {
	t.Parallel()
	destination := os.Getenv("HIR_CACHE_ORACLE")
	if destination == "" {
		t.Skip("set HIR_CACHE_ORACLE")
	}
	if err := os.MkdirAll(destination, 0700); err != nil {
		t.Fatal(err)
	}
	sources := []string{
		"function Widget(flag) { let x = 1; if (flag) x = 2; else x = 3; const Child = () => <span>{x}</span>; return <Child />; }",
		"function Widget(flag) { let x = 7; if (flag) x = 8; else x = 9; const Child = () => <span>{x}</span>; return <Child />; }",
		"function Widget(n) { let x = 0; for (let i = 0; i < n; i++) { x += i; } return <div>{x}</div>; }",
	}
	var manifest, output strings.Builder
	for fixture, source := range sources {
		sourcePath := filepath.Join(destination, fmt.Sprintf("source-%d.tsx", fixture))
		if err := os.WriteFile(sourcePath, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		probe := rule.Rule{Name: "stage1-hir-cache", NeedsTypeChecker: true, Run: func(ctx rule.Context, _ any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(root *ast.Node) {
				symbols := constructionSymbols(root.AsSourceFile(), ctx.TypeChecker)
				factsPath := filepath.Join(destination, fmt.Sprintf("symbols-%d", fixture))
				if err := os.WriteFile(factsPath, []byte(symbols), 0600); err != nil {
					t.Fatal(err)
				}
				ordinal := 0
				forEachFunctionLike(root, func(node *ast.Node) {
					first := ForFunction(ctx, node)
					if first == nil {
						t.Fatal("cache probe declined")
					}
					before := oracleDump(first)
					second, third := ForFunction(ctx, node), ForFunction(ctx, node)
					if first != second || second != third || before != oracleDump(third) {
						t.Fatal("Go cache identity or construction changed")
					}
					unchecked := ctx
					unchecked.TypeChecker = nil
					if ForFunction(unchecked, node) != nil {
						t.Fatal("checker-less Go cache hit")
					}
					key := fmt.Sprintf("%d-%d", fixture, ordinal)
					ordinal++
					fmt.Fprintf(&manifest, "%s\t%s\t%s\t%d\t%d\n", key, sourcePath, factsPath, node.Pos(), node.End())
					fmt.Fprintf(&output, "case\t%s\nidentity 1 unchecked 1\n%s", key, before)
				})
			}}
		}}
		rule_testing.RunTyped(t, probe, fmt.Sprintf("cache-%d.tsx", fixture), source)
	}
	for name, data := range map[string]string{"manifest.tsv": manifest.String(), "go.dump": output.String()} {
		if err := os.WriteFile(filepath.Join(destination, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

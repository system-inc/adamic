package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func namespaceGraphForTest(t *testing.T, source string) (*namespaceCallGraph, []*ast.Node) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "graph.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checker, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	return &namespaceCallGraph{lowering: &lowering{program: program, checker: checker}, functions: map[*ast.Node]*namespaceCallNode{}}, program.Files()[0].Statements.Nodes
}

// Count body expansions rather than wall time: shared edges must not expand paths.
func TestNamespaceCallGraphLinearWork(t *testing.T) {
	t.Parallel()
	for _, depth := range []int{12, 24} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			source := "function f0():void {}\n"
			for level := 1; level <= depth; level++ {
				source += fmt.Sprintf("function f%d():void {if(false){f%d();f%d();}}\n", level, level-1, level-1)
			}
			source += "namespace Debug {export let ready=true;}"
			graph, statements := namespaceGraphForTest(t, source)
			graph.reach(statements[depth])
			if graph.walks != depth+1 {
				t.Fatalf("nonlinear function walks: got %d, want %d", graph.walks, depth+1)
			}
			for _, function := range statements[:depth+1] {
				graph.reach(function)
			}
			if graph.walks != depth+1 {
				t.Fatalf("memoized calls expanded bodies again: %d", graph.walks)
			}
		})
	}
}

func TestNamespaceCallGraphCycleUnion(t *testing.T) {
	t.Parallel()
	graph, statements := namespaceGraphForTest(t, `
 function one():boolean {two();return First.x;}
 function two():boolean {three();return Second.x;}
 function three():boolean {return one();}
 namespace First {export const x=true;}
 namespace Second {export const x=false;}`)
	for _, function := range statements[:3] {
		reaches := graph.reach(function)
		if len(reaches) != 2 {
			t.Fatalf("partial cycle reach for %s: %d namespaces", function.Name().Text(), len(reaches))
		}
		for _, declaration := range statements[3:] {
			if reaches[declaration] == nil {
				t.Fatalf("cycle lost %s", declaration.Name().Text())
			}
		}
	}
	if graph.walks != 3 {
		t.Fatalf("cycle walked %d bodies", graph.walks)
	}
}

func TestNamespaceEnumInitializationIndependentOfModuleAnalysis(t *testing.T) {
	t.Parallel()
	graph, _ := namespaceGraphForTest(t, "namespace N {function read():number{return E.A;} const x=read(); enum E {A}}")
	err := graph.lowering.namespaceInitialization(graph.lowering.program.Files())
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "enum before its runtime initialization") {
		t.Fatalf("namespace enum lost its independent refusal: %v", err)
	}
}

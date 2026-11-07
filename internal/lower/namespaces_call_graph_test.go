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

func namespaceGraphForTest(test *testing.T, source string) (*namespaceCallGraph, []*ast.Node) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "graph.a")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	checker, release := program.Checker(context.Background(), program.Files()[0])
	test.Cleanup(release)
	return &namespaceCallGraph{
		lowering:  &lowering{program: program, checker: checker},
		functions: map[*ast.Node]*namespaceCallNode{},
	}, program.Files()[0].Statements.Nodes
}

// Count body expansions rather than wall time: shared edges must not expand paths.
func TestNamespaceCallGraphLinearWork(test *testing.T) {
	test.Parallel()
	for _, depth := range []int{12, 24} {
		test.Run(fmt.Sprint(depth), func(test *testing.T) {
			source := "function f0():void {}\n"
			for level := 1; level <= depth; level++ {
				source += fmt.Sprintf("function f%d():void {if(false){f%d();f%d();}}\n", level, level-1, level-1)
			}
			source += "namespace Debug {export let ready=true;}"
			graph, statements := namespaceGraphForTest(test, source)
			graph.reach(statements[depth])
			if graph.walks != depth+1 {
				test.Fatalf("nonlinear function walks: got %d, want %d", graph.walks, depth+1)
			}
			for _, function := range statements[:depth+1] {
				graph.reach(function)
			}
			if graph.walks != depth+1 {
				test.Fatalf("memoized calls expanded bodies again: %d", graph.walks)
			}
		})
	}
}

func TestNamespaceCallGraphCycleUnion(test *testing.T) {
	test.Parallel()
	graph, statements := namespaceGraphForTest(test, `
 function one():boolean {two();return First.x;}
 function two():boolean {three();return Second.x;}
 function three():boolean {return one();}
 namespace First {export const x=true;}
 namespace Second {export const x=false;}`)
	for _, function := range statements[:3] {
		reaches := graph.reach(function)
		if len(reaches) != 2 {
			test.Fatalf("partial cycle reach for %s: %d namespaces", function.Name().Text(), len(reaches))
		}
		for _, declaration := range statements[3:] {
			if reaches[declaration] == nil {
				test.Fatalf("cycle lost %s", declaration.Name().Text())
			}
		}
	}
	if graph.walks != 3 {
		test.Fatalf("cycle walked %d bodies", graph.walks)
	}
}

func TestNamespaceEnumInitializationIndependentOfModuleAnalysis(test *testing.T) {
	test.Parallel()
	graph, _ := namespaceGraphForTest(
		test,
		"namespace N {function read():number{return E.A;} const x=read(); enum E {A}}",
	)
	namespaceInitializationError := graph.lowering.namespaceInitialization(graph.lowering.program.Files())
	var notYet *NotYet
	if !errors.As(namespaceInitializationError, &notYet) ||
		!strings.Contains(namespaceInitializationError.Error(), "enum before its runtime initialization") {
		test.Fatalf("namespace enum lost its independent refusal: %v", namespaceInitializationError)
	}
}

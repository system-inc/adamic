package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestProgramRegionTscNodeMemberSites(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/program_region/tsc_node_membership.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	typeChecker, release := loaded.Checker(context.Background(), entry)
	defer release()
	discovery, err := lowerChecked(loaded, typeChecker, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	var nodeType, arrayType *checker.Type
	allocations := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration && node.Name().Text() == "Node" {
			nodeType = typeChecker.GetDeclaredTypeOfSymbol(typeChecker.GetSymbolAtLocation(node.Name()))
		}
		if node.Kind == ast.KindVariableDeclaration && node.Name() != nil && node.Name().Text() == "nodes" {
			arrayType = typeChecker.GetTypeAtLocation(node.Name())
		}
		if node.Kind == ast.KindNewExpression || node.Kind == ast.KindArrayLiteralExpression {
			actual := typeChecker.GetTypeAtLocation(node)
			contextual := typeChecker.GetContextualType(node, checker.ContextFlagsNone)
			member := discovery.programPlan.types[int(actual.Id())]
			if contextual != nil {
				member = member || discovery.programPlan.types[int(contextual.Id())]
			}
			label := typeChecker.TypeToStringEx(actual, nil, checker.TypeFormatFlagsNoTruncation, nil)
			t.Logf("allocation %s at %s: selected=%t", label, loaded.Where(node), member)
			if !member {
				t.Fatalf("tsc graph allocation %s left counted at %s", label, loaded.Where(node))
			}
			allocations++
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	if nodeType == nil || arrayType == nil {
		t.Fatal("Node declaration or instantiated NodeArray missing")
	}
	for name, declared := range map[string]*checker.Type{"Node": nodeType, "NodeArray<Node>": arrayType} {
		label := typeChecker.TypeToStringEx(declared, nil, checker.TypeFormatFlagsNoTruncation, nil)
		if !discovery.programPlan.types[int(declared.Id())] {
			t.Fatalf("selector left %s (%s) counted", name, label)
		}
		t.Logf("selector member %s: %s", name, label)
	}
	if allocations != 4 {
		t.Fatalf("allocation sites %d, want BaseNode factory, empty declarations, populated declarations and NodeArray", allocations)
	}
	built, err := lowerChecked(loaded, typeChecker, entry, discovery.programPlan, false)
	if err != nil {
		t.Fatal(err)
	}
	objects, arrays := 0, 0
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if value.Class > 0 {
				if !built.result.Classes[value.Class-1].ProgramRegion {
					t.Fatal("factory Node allocation left counted")
				}
				objects++
			}
		case ir.ArrayLiteral:
			if !value.ProgramRegion {
				t.Fatal("NodeArray or declarations allocation left counted")
			}
			arrays++
		}
		return true
	}
	walk(built.result.Main, inspect)
	for _, function := range built.result.Functions {
		walk(function.Body, inspect)
	}
	if objects != 1 || arrays != 3 {
		t.Fatalf("lowered graph allocation sites objects=%d arrays=%d, want 1 and 3", objects, arrays)
	}
}

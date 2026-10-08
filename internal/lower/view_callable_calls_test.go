package lower

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvedCallableDeclarationMembership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolved.a")
	source := `interface Target { read():number; read(value:number):number; read(value:string):number; }
function probe(target:Target):number { return target.read("ok"); }`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	var call *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if call == nil {
		t.Fatal("missing witness call")
	}
	declared := checked.GetSignaturesOfType(checked.GetTypeAtLocation(call.AsCallExpression().Expression), checker.SignatureKindCall)
	if len(declared) != 3 || checked.GetResolvedSignature(call).Declaration() != declared[2].Declaration() {
		t.Fatal("witness did not resolve its third declaration")
	}
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	arguments := []ir.Expression{ir.StringConstant{}}
	if err := l.convertResolvedCallableArguments(call, declared, arguments); err != nil {
		t.Fatal(err)
	}
	if err := l.convertResolvedCallableArguments(call, declared[:2], arguments); err == nil || !strings.Contains(err.Error(), "outside its declared overload set") {
		t.Fatalf("undeclared resolution admitted: %v", err)
	}
}

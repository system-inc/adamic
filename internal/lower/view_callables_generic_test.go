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

func TestGenericCallableDeclarationIdentity(t *testing.T) {
	source := `interface Node { readonly value:number; }
interface Target { cloneNode<T extends Node | undefined>(node:T):T; }
function good<U extends Node | undefined>(node:U):U { return node; }
function badResult<U extends Node | undefined>(node:U):Node|undefined { return node; }
function badConstraint<U extends Node>(node:U):U { return node; }
function ordinary(node:Node):Node { return node; }`
	path := filepath.Join(t.TempDir(), "generic.a")
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
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	target := checked.GetTypeAtLocation(file.Statements.Nodes[1].Name())
	member := checked.GetPropertyOfType(target, "cloneNode")
	wanted := checked.GetSignaturesOfType(checked.GetTypeOfSymbol(member), checker.SignatureKindCall)[0]
	for _, declaration := range file.Statements.Nodes[2:] {
		if declaration.Kind != ast.KindFunctionDeclaration {
			t.Fatal("unexpected producer")
		}
		got := l.genericCallableMatches(checked.GetSignatureFromDeclaration(declaration), wanted)
		want := declaration.Name().Text() == "good"
		if got != want {
			t.Fatalf("generic declaration %s compatibility %t, want %t", declaration.Name().Text(), got, want)
		}
	}
}

func TestGenericCallableResultInstantiation(t *testing.T) {
	source := `interface Node { readonly value:number; }
interface Detail extends Node { readonly label:string; }
function clone<T extends Node | undefined>(node:T):T { return node; }`
	path := filepath.Join(t.TempDir(), "result.a")
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
	declaration := file.Statements.Nodes[2]
	parameter := checked.GetTypeAtLocation(declaration.TypeParameters()[0].Name())
	node := checked.GetTypeAtLocation(file.Statements.Nodes[0].Name())
	detail := checked.GetTypeAtLocation(file.Statements.Nodes[1].Name())
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	l.typeMapper = newTypeMapper([]*checker.Type{parameter}, []*checker.Type{detail})
	if err := l.checkGenericCallableReturns(declaration, declaration, detail); err != nil {
		t.Fatal(err)
	}
	l.typeMapper = newTypeMapper([]*checker.Type{parameter}, []*checker.Type{node})
	if err := l.checkGenericCallableReturns(declaration, declaration, detail); err == nil || !strings.Contains(err.Error(), "a generic callable result outside its instantiation") {
		t.Fatalf("constraint substituted for result instantiation: %v", err)
	}
}

package lower

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// intersectionViewSource declares one intersection of plain object types and one per member that
// keeps an intersection's view unsupported. Each is read from its alias or its parameter.
const intersectionViewSource = `interface Node { readonly kind: number; }
class Box { readonly kind: 201 = 201; }
type Plain = Node & { readonly kind: 201; readonly literal: 11 };
type Nominal = Box & { readonly literal: 11 };
type Callable = Node & { (): number };
type Indexed = Node & { [key: string]: number };
function generic<T extends Node>(value: T & { readonly literal: 11 }): number { return value.literal; }
console.log("loaded");
`

// intersectionViewTypes lowers nothing: it loads intersectionViewSource and gives a lowering over its
// checker, with each intersection by its alias's name, and the generic one as "Generic".
func intersectionViewTypes(t *testing.T) (*lowering, map[string]*checker.Type, map[string]*ast.Node) {
	t.Helper()
	path := writeAgreementSource(t, "main.ts", intersectionViewSource)
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
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	types := map[string]*checker.Type{}
	nodes := map[string]*ast.Node{}
	for _, statement := range file.Statements.Nodes {
		switch statement.Kind {
		case ast.KindTypeAliasDeclaration:
			types[statement.Name().Text()] = checked.GetTypeAtLocation(statement.Name())
			nodes[statement.Name().Text()] = statement
		case ast.KindFunctionDeclaration:
			parameter := statement.Parameters()[0]
			types["Generic"] = checked.GetTypeAtLocation(parameter.Name())
			nodes["Generic"] = parameter
		}
	}
	for name, proven := range types {
		if proven.Flags()&checker.TypeFlagsIntersection == 0 {
			t.Fatalf("%s is %s, not an intersection", name, checked.TypeToString(proven))
		}
	}
	return l, types, nodes
}

// An intersection of plain object types has a contract. A class instance member would lose its
// nominal check, a callable member its signature, an index signature its dictionary, and a type
// parameter has no layout, so each stays unsupported. On every program these lower, each of the four
// already stops at its representation, so only this test guards the rule itself.
func TestIntersectionViewFamily(t *testing.T) {
	t.Parallel()
	l, types, _ := intersectionViewTypes(t)
	for name, want := range map[string]string{"Plain": "", "Nominal": "intersection", "Callable": "intersection", "Indexed": "intersection", "Generic": "intersection"} {
		if got := l.unsupportedViewFamily(types[name]); got != want {
			t.Errorf("%s: family %q, want %q", name, got, want)
		}
	}
}

// The contract's fields are the intersection's resolved properties with their refined types: kind is
// 201 where Node's is a number, and literal is the added 11. Its mutant builds each field's contract
// from the first member that declares it, and kind's contract then admits any number.
func TestIntersectionViewContractRefinesFields(t *testing.T) {
	t.Parallel()
	l, types, nodes := intersectionViewTypes(t)
	id, err := l.viewContract(nodes["Plain"], types["Plain"])
	if err != nil {
		t.Fatal(err)
	}
	contract := l.result.ViewContracts[id-1]
	if contract.Unsupported != "" || contract.Kind != ir.ViewObject {
		t.Fatalf("Plain: %+v", contract)
	}
	allowed := map[string][]ir.ViewLiteral{}
	for _, field := range contract.Fields {
		allowed[field.Name] = l.result.ViewContracts[field.Contract-1].Allowed
	}
	for name, want := range map[string]float64{"kind": 201, "literal": 11} {
		if len(allowed[name]) != 1 || allowed[name][0].Of != ir.Number || allowed[name][0].Number != want {
			t.Errorf("%s: allowed %+v, want only %v", name, allowed[name], want)
		}
	}
}

// An intersection whose contract is built inside intersectionViewDepthBound others has none of its
// own: its descriptor names the bound, and a demanded read through it is NotYet.
func TestIntersectionViewDepthBound(t *testing.T) {
	t.Parallel()
	l, types, nodes := intersectionViewTypes(t)
	l.intersectionViewDepth = intersectionViewDepthBound
	id, err := l.viewContract(nodes["Plain"], types["Plain"])
	if err != nil {
		t.Fatal(err)
	}
	if got := l.result.ViewContracts[id-1].Unsupported; got != "intersection nested past the depth bound of 8" {
		t.Fatalf("family %q, want the depth bound named", got)
	}
	if l.intersectionViewDepth != intersectionViewDepthBound {
		t.Fatalf("depth %d after the descriptor, want %d", l.intersectionViewDepth, intersectionViewDepthBound)
	}
}

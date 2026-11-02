package lower

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// intersectionViewSource declares intersections of plain object types, one refining a node-typed
// field, a chain nested one past intersectionViewDepthBound, and one per member that keeps an
// intersection's view unsupported. Each is read from its declaration's name or its parameter.
const intersectionViewSource = `import type { Weak } from 'adamic';
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface ImportClause extends Node { readonly kind: 273; readonly isTypeOnly: boolean; readonly name?: Identifier | undefined; }
class Box { readonly kind: 201 = 201; }
type Plain = Node & { readonly kind: 201; readonly literal: 11 };
type TypeOnlyImportClause = ImportClause & { readonly isTypeOnly: true; readonly name: Identifier };
type Nominal = Box & { readonly literal: 11 };
type Callable = Node & { (): number };
type Indexed = Node & { [key: string]: number };
type Weakly = NonNullable<Weak<Identifier>>;
type D9 = Node & { readonly depth: 9 };
type D8 = Node & { readonly depth: 8; readonly next: D9 };
type D7 = Node & { readonly depth: 7; readonly next: D8 };
type D6 = Node & { readonly depth: 6; readonly next: D7 };
type D5 = Node & { readonly depth: 5; readonly next: D6 };
type D4 = Node & { readonly depth: 4; readonly next: D5 };
type D3 = Node & { readonly depth: 3; readonly next: D4 };
type D2 = Node & { readonly depth: 2; readonly next: D3 };
type D1 = Node & { readonly depth: 1; readonly next: D2 };
function generic<T extends Node>(value: T & { readonly literal: 11 }): number { return value.literal; }
console.log("loaded");
`

// intersectionViewTypes lowers nothing: it loads intersectionViewSource and gives a lowering over its
// checker, with each declared type by its name, and the generic parameter's as "Generic".
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
		case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration:
			types[statement.Name().Text()] = checked.GetTypeAtLocation(statement.Name())
			nodes[statement.Name().Text()] = statement
		case ast.KindFunctionDeclaration:
			parameter := statement.Parameters()[0]
			types["Generic"] = checked.GetTypeAtLocation(parameter.Name())
			nodes["Generic"] = parameter
		}
	}
	for _, name := range []string{"Plain", "TypeOnlyImportClause", "Nominal", "Callable", "Indexed", "Weakly", "D1", "D9", "Generic"} {
		if types[name] == nil || types[name].Flags()&checker.TypeFlagsIntersection == 0 {
			t.Fatalf("%s is not an intersection", name)
		}
	}
	return l, types, nodes
}

// An intersection of plain object types has a contract. A class instance member would lose its
// nominal check, a callable member its signature, an index signature its dictionary, a type
// parameter has no layout, and Weak<Identifier>'s present half is kept weakly, so each stays
// unsupported. On every program these lower, the first four already stop at their representation,
// so only this test guards the rule itself.
func TestIntersectionViewFamily(t *testing.T) {
	t.Parallel()
	l, types, _ := intersectionViewTypes(t)
	for name, want := range map[string]string{"Plain": "", "TypeOnlyImportClause": "", "Nominal": "intersection", "Callable": "intersection", "Indexed": "intersection", "Generic": "intersection", "Weakly": "intersection"} {
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

// A refinement that narrows a node field checks the narrower type: name is required and points to
// Identifier's own contract, not to the base's Identifier | undefined, and isTypeOnly admits only
// true. No program on main reaches it, since representation() knows an object intersection only by
// scalar fields, so the contract takes its object representation from its members. Under the
// mutant, name points to the base's union.
func TestIntersectionViewContractRefinesNodeField(t *testing.T) {
	t.Parallel()
	l, types, nodes := intersectionViewTypes(t)
	id, err := l.viewContract(nodes["TypeOnlyImportClause"], types["TypeOnlyImportClause"])
	if err != nil {
		t.Fatal(err)
	}
	contract := l.result.ViewContracts[id-1]
	if contract.Unsupported != "" || contract.Kind != ir.ViewObject || contract.Of != ir.Object {
		t.Fatalf("TypeOnlyImportClause: %+v", contract)
	}
	identifier := l.result.ViewContractTypes[int(types["Identifier"].Id())]
	if identifier == 0 {
		t.Fatal("no contract for Identifier")
	}
	fields := map[string]ir.ViewFieldContract{}
	for _, field := range contract.Fields {
		fields[field.Name] = field
	}
	if name := fields["name"]; name.Contract != identifier || name.Optional {
		t.Errorf("name: %+v points to %+v, want Identifier's required contract %d", name, l.result.ViewContracts[name.Contract-1], identifier)
	}
	if allowed := l.result.ViewContracts[fields["isTypeOnly"].Contract-1].Allowed; len(allowed) != 1 || allowed[0].Of != ir.Boolean || !allowed[0].Boolean {
		t.Errorf("isTypeOnly: allowed %+v, want only true", allowed)
	}
}

// Nested intersections resolve field by field, each inside the one before, up to
// intersectionViewDepthBound of them. The ninth, D9 inside D8's next, has no contract of its own:
// its descriptor names the bound, and a demanded read through it is NotYet. D8 has its contract.
func TestIntersectionViewDepthBound(t *testing.T) {
	t.Parallel()
	l, types, nodes := intersectionViewTypes(t)
	if _, err := l.viewContract(nodes["D1"], types["D1"]); err != nil {
		t.Fatal(err)
	}
	if got := l.result.ViewContracts[l.result.ViewContractTypes[int(types["D9"].Id())]-1].Unsupported; got != "intersection nested past the depth bound of 8" {
		t.Errorf("D9: family %q, want the depth bound named", got)
	}
	if got := l.result.ViewContracts[l.result.ViewContractTypes[int(types["D8"].Id())]-1].Unsupported; got != "" {
		t.Errorf("D8: family %q, want a contract", got)
	}
	if l.intersectionViewDepth != 0 {
		t.Errorf("depth %d after the contract, want 0", l.intersectionViewDepth)
	}
}

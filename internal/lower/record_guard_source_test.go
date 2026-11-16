package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// recordGuardImports is tsc's import nodes in miniature: an ImportSpecifier's parent is NamedImports,
// whose parent is the ImportClause holding phaseModifier, with a record in the same program.
const recordGuardImports = `enum SyntaxKind { Identifier, ImportClause, NamedImports, ImportSpecifier, TypeKeyword, DeferKeyword }
interface Node {
	readonly kind: SyntaxKind;
	readonly parent: Node | undefined;
}
interface ImportClause extends Node {
	readonly kind: SyntaxKind.ImportClause;
	readonly phaseModifier: SyntaxKind.TypeKeyword | SyntaxKind.DeferKeyword | undefined;
}
interface NamedImports extends Node {
	readonly kind: SyntaxKind.NamedImports;
	readonly parent: ImportClause;
}
interface ImportSpecifier extends Node {
	readonly kind: SyntaxKind.ImportSpecifier;
	readonly parent: NamedImports;
	readonly isTypeOnly: boolean;
}
const table: { [key: string]: number } = {};
function put(key: string): void {
	table[key] = 1;
}
put(String(1));
const clause: ImportClause = { kind: SyntaxKind.ImportClause, parent: undefined, phaseModifier: SyntaxKind.TypeKeyword };
const named: NamedImports = { kind: SyntaxKind.NamedImports, parent: clause };
const specifier: ImportSpecifier = { kind: SyntaxKind.ImportSpecifier, parent: named, isTypeOnly: false };
`

// tsc's two reads of an ImportClause's optional phaseModifier through an ImportSpecifier's grandparent
// (#5d9y3rs): utilitiesPublic.ts:1539's isTypeOnlyImportDeclaration, which downcasts to ImportClause in a
// case of its own too, and utilities.ts:11038's isTypeDeclaration, which reads only the grandparent. Each
// downcast is from Node, a type no record reaches, so the view's checks can't see a record, and the read
// lowers beside the program's record. Held to Node on JavaScript and natively under the sanitizers.
// Mutant: checkRecordGuardedReads asks ir.Program.CheckedField again, so any registered key stops it, and
// both rows stop as "a union or optional field read in a program with record storage".
func TestRecordGuardLowersTscImportShapes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"isTypeOnlyImportDeclaration", recordGuardImports + `function isTypeOnly(node: Node): boolean {
	switch (node.kind) {
		case SyntaxKind.ImportSpecifier:
			return (node as ImportSpecifier).isTypeOnly || (node as ImportSpecifier).parent.parent.phaseModifier === SyntaxKind.TypeKeyword;
		case SyntaxKind.ImportClause:
			return (node as ImportClause).phaseModifier === SyntaxKind.TypeKeyword;
	}
	return false;
}
console.log(String(isTypeOnly(specifier)));
console.log(String(isTypeOnly(clause)));
`, "true\ntrue\n"},
		{"isTypeDeclaration", recordGuardImports + `function isTypeDeclaration(node: Node): boolean {
	switch (node.kind) {
		case SyntaxKind.ImportSpecifier:
			return (node as ImportSpecifier).parent.parent.phaseModifier === SyntaxKind.TypeKeyword;
	}
	return false;
}
console.log(String(isTypeDeclaration(specifier)));
console.log(String(isTypeDeclaration(clause)));
`, "true\nfalse\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, out, errs := brandedNative(t, program); code != 0 || out != probe.want || errs != "" {
				t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", code, out, errs, probe.want)
			}
		})
	}
}

// What a registration's view says about records decides the record guard (#5d9y3rs), asked directly,
// since no source road reaches a record seen as a fixed type today (storage stops it first). A key
// registered by a view from Node, through ImportClause, lets the read through. Every other row keeps
// the guard's stop: a source a record reaches (Weak, every field optional, which `{}` relates to), a
// union with such a member (Node | Weak), no source at all, a key a record-free view and a reachable one
// both registered, in either order, and a key added past markCheckedField.
// Mutant: viewRecordFree answering true for every view, and the Weak, union and no-source rows lower.
func TestRecordGuardKeepsAReachableSource(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `enum SyntaxKind { ImportClause, TypeKeyword }
interface Node {
	readonly kind: SyntaxKind;
}
interface ImportClause extends Node {
	readonly kind: SyntaxKind.ImportClause;
	readonly phaseModifier: SyntaxKind.TypeKeyword | undefined;
}
interface Weak {
	readonly kind?: SyntaxKind;
}
type Either = Node | Weak;
const table: { [key: string]: number } = {};
function put(key: string): void {
	table[key] = 1;
}
put(String(1));
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	types := map[string]*checker.Type{}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind == ast.KindInterfaceDeclaration || statement.Kind == ast.KindTypeAliasDeclaration {
			types[statement.Name().Text()] = checked.GetDeclaredTypeOfSymbol(checked.GetSymbolAtLocation(statement.Name()))
		}
	}
	for _, name := range []string{"Node", "ImportClause", "Weak", "Either"} {
		if types[name] == nil {
			t.Fatalf("no type %s", name)
		}
	}
	clause := types["ImportClause"]
	key := ir.CheckedFieldKey(int(clause.Id()), "phaseModifier")
	for _, row := range []struct {
		name    string
		sources []*checker.Type
		direct  bool
		stops   bool
	}{
		{"from Node", []*checker.Type{types["Node"]}, false, false},
		{"from Weak", []*checker.Type{types["Weak"]}, false, true},
		{"from Node | Weak", []*checker.Type{types["Either"]}, false, true},
		{"with no source", []*checker.Type{nil}, false, true},
		{"from Node, then Weak", []*checker.Type{types["Node"], types["Weak"]}, false, true},
		{"from Weak, then Node", []*checker.Type{types["Weak"], types["Node"]}, false, true},
		{"past markCheckedField", nil, true, true},
	} {
		l := &lowering{program: program, checker: checked, result: &ir.Program{}}
		for _, source := range row.sources {
			l.markCheckedField(key, "phaseModifier", l.viewRecordFree(source, []*checker.Type{clause}))
		}
		if row.direct {
			l.result.CheckedFields = map[string]bool{key: true}
		}
		l.recordGuardedReads = append(l.recordGuardedReads, recordGuardedRead{node: file.Statements.Nodes[0], receiverTypeID: int(clause.Id()), name: "phaseModifier"})
		err := l.checkRecordGuardedReads()
		var notYet *NotYet
		if stopped := errors.As(err, &notYet); stopped != row.stops || err != nil && !stopped {
			t.Errorf("%s: want stopped %v, got %v", row.name, row.stops, err)
		}
	}
}

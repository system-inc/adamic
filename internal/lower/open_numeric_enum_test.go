package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Whether a numeric enum is open has one definition, wholeNumericEnums, which both readers call
// (#4cfvp4q, Compiler's should on #wcva20f): openNumericEnumType, the open-enum tag walk's and the
// relation rules', and holdsOpenNumericEnum, the arm filter's. They differ only where each is meant
// to: openNumericEnumType answers for a type that is exactly one enum's whole type, and
// holdsOpenNumericEnum for any union holding one, failing closed on an enum that isn't purely numeric.
// Each shape is pinned for both. Mutants: openNumericEnumType's size check dropped, under which the
// wider union reads open to it; holdsOpenNumericEnum ignoring impure, under which the heterogeneous
// literal reads closed to it; and wholeNumericEnums asking one member more than the enum has, under
// which the pure and single-member enums read closed to both callers. The same pins pass against main's
// two separate definitions (7e3c0682), so the helper moves no answer.
func TestOpenNumericEnumHasOneDefinition(test *testing.T) {
	test.Parallel()
	const source = `enum Kind { A, B, C }
enum One { A = 7 }
enum Mixed { A = 0, B = "b" }
let pure: Kind = Kind.A;
let partial: Kind.A | Kind.B = Kind.A;
let wider: Kind | "x" = "x";
let widerPartial: Kind.A | "x" = "x";
let single: One = One.A;
let mixed: Mixed.A = Mixed.A;
let member: Kind.B = Kind.B;
console.log(String(pure) + String(partial) + wider + widerPartial + String(single) + String(mixed) + String(member));
`
	path := filepath.Join(test.TempDir(), "main.ts")
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: typeChecker, program: program}
	declared := map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			declared[node.Name().Text()] = node.Name()
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	for _, probe := range []struct {
		name       string
		open, hold bool
	}{
		{"pure", true, true},
		{"partial", false, false},
		{"wider", false, true},
		{"widerPartial", false, false},
		{"single", true, true},
		{"mixed", false, true},
		{"member", false, false},
	} {
		name, found := declared[probe.name]
		if !found {
			test.Fatalf("no declaration %s", probe.name)
		}
		proven := typeChecker.GetTypeAtLocation(name)
		if got := l.openNumericEnumType(proven); got != probe.open {
			test.Errorf(
				"%s (%s): openNumericEnumType %t, want %t",
				probe.name,
				typeChecker.TypeToString(proven),
				got,
				probe.open,
			)
		}
		if got := l.holdsOpenNumericEnum(flattenedMembers(proven)); got != probe.hold {
			test.Errorf(
				"%s (%s): holdsOpenNumericEnum %t, want %t",
				probe.name,
				typeChecker.TypeToString(proven),
				got,
				probe.hold,
			)
		}
	}
}

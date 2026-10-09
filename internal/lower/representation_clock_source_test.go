package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Probe checked types before monomorphization. A structural base constrains
// properties, but cannot fix the runtime header of every possible subtype.
// Constraint-backed Union storage preserves those headers without choosing one.
func TestRepresentationClockSourceCheckedTypes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "probe.a")
	source := `interface CompilerType { readonly id: number; readonly payload: { readonly label: string }; }
interface CallableType extends CompilerType { (): void; }
declare const objectSource: CompilerType;
declare const callableSource: CallableType;
declare const arraySource: number[];
declare const stringSource: string;
function inspect<Source extends CompilerType>(source: Source): void {}
function lengthOnly<Length extends { readonly length: number }>(lengthSource: Length): void {}
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	types := map[string]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			switch name {
			case "objectSource", "callableSource", "arraySource", "stringSource", "source", "lengthSource":
				types[name] = checked.GetTypeAtLocation(node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	for _, probe := range []struct {
		name  string
		want  ir.Type
		known bool
	}{
		{"objectSource", ir.Object, true}, {"callableSource", ir.Closure, true},
		{"arraySource", ir.Array, true}, {"stringSource", ir.String, true},
		{"source", ir.Union, true}, {"lengthSource", 0, false},
	} {
		proven := types[probe.name]
		if proven == nil {
			t.Fatalf("missing checked type %s", probe.name)
		}
		got, known := l.representation(proven)
		if got != probe.want || known != probe.known {
			t.Errorf("%s: representation = (%v, %v), want (%v, %v)", probe.name, got, known, probe.want, probe.known)
		}
	}
	for _, relation := range [][2]string{{"callableSource", "source"}, {"arraySource", "lengthSource"}, {"stringSource", "lengthSource"}} {
		constraint := checked.GetBaseConstraintOfType(types[relation[1]])
		if constraint == nil || !checked.IsTypeAssignableTo(types[relation[0]], constraint) {
			t.Fatalf("%s must satisfy %s constraint", relation[0], relation[1])
		}
	}
	t.Log("CompilerType admits Object and Closure; length-only admits Array and String; Source uses tagged Union storage, not an Object header; length-only storage remains unsupported")
}

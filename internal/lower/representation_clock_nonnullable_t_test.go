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

// An isolated census binder has no ABI. Real type substitution, not an empty
// structural view, determines the storage of NonNullable<T>.
func TestClockNonNullableCheckedRepresentation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := "function keep<T>(value: NonNullable<T>): NonNullable<T> { return value; }\nconst text: string = 'text'; const node: { readonly name: string } = { name: 'node' };\n"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	c, release := program.Checker(context.Background(), file)
	defer release()
	var binder, value, text, object *checker.Type
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case "T":
				if proven := c.GetTypeAtLocation(node); proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
					binder = proven
				}
			case "value":
				value = c.GetTypeAtLocation(node)
			case "text":
				text = c.GetTypeAtLocation(node)
			case "node":
				object = c.GetTypeAtLocation(node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	if binder == nil || value == nil || text == nil || object == nil || value.Flags()&checker.TypeFlagsIntersection == 0 {
		t.Fatal("missing checked intersection probe")
	}
	l := &lowering{checker: c, program: program}
	if got, known := l.representation(value); known {
		t.Fatalf("isolated T & {} acquired storage %v", got)
	}
	for _, probe := range []struct {
		name   string
		target *checker.Type
		want   ir.Type
	}{{"string", text, ir.String}, {"object", object, ir.Object}} {
		t.Run(probe.name, func(t *testing.T) {
			l.typeMapper = newTypeMapper([]*checker.Type{binder}, []*checker.Type{probe.target})
			if got, known := l.representation(value); !known || got != probe.want {
				t.Fatalf("got %v, %v; want %v", got, known, probe.want)
			}
		})
	}
}

package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestPromisedShape(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	source := `async function allowed():Promise<void>{return;} async function required():Promise<string|undefined>{return;} function ordinary():unknown{return;} function inferred(){return;}`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	checked := 0
	sf.AsNode().ForEachChild(func(node *ast.Node) bool {
		if node.Kind != ast.KindFunctionDeclaration {
			return false
		}
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "FunctionDeclaration", "promised-shape")
		if node.Type() == nil {
			if err == nil {
				t.Error("promised-shape accepted unannotated function")
			}
			return false
		}
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) < 6 || fields[0] != "1" || fields[1] != "promised-shape" || fields[3] != "1" || fields[4] != "1" {
			t.Fatalf("bad frame %q", fields)
		}
		id, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) {
			t.Fatal("bad root identity")
		}
		want := c.GetTypeFromTypeNode(node.Type())
		if ast.GetFunctionFlags(node)&ast.FunctionFlagsAsync != 0 {
			want = c.GetPromisedTypeOfPromise(want)
		}
		if p.typesByID[id-1] != want {
			t.Fatalf("promised-shape returned %s, expected %s", c.TypeToString(p.typesByID[id-1]), c.TypeToString(want))
		}
		if node.Name().Text() == "allowed" && want.Flags() != checker.TypeFlagsVoid {
			t.Fatal("positive async void control failed")
		}
		if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "FunctionDeclaration", "promised-shape\nextra"); err == nil {
			t.Error("suffix accepted")
		}
		checked++
		return false
	})
	if checked != 3 {
		t.Fatalf("missing controls: checked %d", checked)
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "promised-shape"); err == nil || !strings.Contains(err.Error(), "annotated function") {
		t.Fatalf("value kind accepted: %v", err)
	}
}

package load

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestRequireBindingHasTheImportedModuleType(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", `import * as imported from 'node:fs';
const bare = require('fs');
const prefixed = require('node:fs');
const result = bare.existsSync('.');
`})
	program, err := Load(paths)
	if err != nil {
		t.Fatal(err)
	}
	checker, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	symbols := map[string]*ast.Symbol{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			symbols[node.Text()] = checker.GetSymbolAtLocation(node)
		}
		node.ForEachChild(visit)
		return false
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	imported := checker.GetTypeOfSymbol(symbols["imported"])
	for _, name := range []string{"bare", "prefixed"} {
		got := checker.GetTypeOfSymbol(symbols[name])
		if got != imported {
			t.Errorf("%s type %s differs from import %s", name, checker.TypeToString(got), checker.TypeToString(imported))
		}
		if checker.GetPropertyOfType(got, "existsSync") != checker.GetPropertyOfType(imported, "existsSync") {
			t.Errorf("%s lost member declaration identity", name)
		}
	}
}

func TestNodeBuiltinNamesAgreeWithNode(t *testing.T) {
	t.Parallel()
	output, err := exec.Command("node", "-e", `console.log(JSON.stringify(require('module').builtinModules))`).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, output)
	}
	var names []string
	if err := json.Unmarshal(output, &names); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if _, builtin := NodeBuiltin(name); !builtin {
			t.Errorf("Node builtin %q refused", name)
		}
	}
	for _, name := range []string{"./fs", "some-package", "node:not-a-module", "test", "sqlite", "sea"} {
		if _, builtin := NodeBuiltin(name); builtin {
			t.Errorf("non-builtin %q accepted", name)
		}
	}
}

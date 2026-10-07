package load

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestRequireBindingHasTheImportedModuleType(t *testing.T) {
	t.Parallel()
	for _, module := range []string{"fs", "path", "perf_hooks"} {
		t.Run(module, func(t *testing.T) {
			member := map[string]string{"fs": "existsSync", "path": "join", "perf_hooks": "performance"}[module]
			source := `import * as imported from 'node:fs';
const bare = require('fs');
const prefixed = require('node:fs');
`
			source = strings.ReplaceAll(source, "fs", module)
			paths := writeProgram(t, [2]string{"main.a", source})
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
				if checker.GetPropertyOfType(got, member) != checker.GetPropertyOfType(imported, member) {
					t.Errorf("%s lost member declaration identity", name)
				}
			}
		})
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

// The Node branch uses a destructured optional view inside a try and catch.
// Its host methods must keep the pinned declarations even in that view.
func TestRequirePerformanceCoreShapeChecks(t *testing.T) {
	t.Parallel()
	for _, specifier := range []string{"perf_hooks", "node:perf_hooks"} {
		t.Run(specifier, func(t *testing.T) {
			source := `function isNodeLikeSystem(): boolean { return true; }
function tryGetPerformance() {
 if (isNodeLikeSystem()) {
  try {
   const { performance } = require("perf_hooks") as Partial<typeof import("perf_hooks")>;
   if (performance) {
    const now: number = performance.now();
    const origin: number = performance.timeOrigin;
    performance.mark("start");
    performance.measure("span", "start");
    performance.clearMarks();
    performance.clearMeasures();
    return { shouldWriteNativeEvents: false, performance, now, origin };
   }
  } catch { }
 }
 return undefined;
}
const hooks = tryGetPerformance();`
			source = strings.ReplaceAll(source, "perf_hooks", specifier)
			if _, err := nodeSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLocalRequireDoesNotLoadNodeGlobals(t *testing.T) {
	t.Parallel()
	program, err := nodeSource(t, `function require(name: string): string { return name; } const module = { exports: 'local' }; console.log(require(module.exports));`)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range program.compiler.GetSourceFiles() {
		if IsNodeLibrary(file) {
			t.Fatal("local require or module activated Node globals")
		}
	}
}

package lower

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Test preflight separately from unsupported host member lowering. Neither host
// namespace has an executable body in the checked module graph.
func TestNamespaceAmbientHostInitialization(test *testing.T) {
	test.Parallel()
	// An executable namespace read before initialization must still be rejected.
	graph, _ := namespaceGraphForTest(test, "console.log(`${N.value}`); namespace N { export const value = true; }")
	namespaceInitializationError := graph.lowering.namespaceInitialization(graph.lowering.program.Files())
	var stop *NotYet
	if !errors.As(namespaceInitializationError, &stop) ||
		!strings.Contains(namespaceInitializationError.Error(), "before runtime initialization") {
		test.Fatalf("runtime namespace initialization guard lost: %v", namespaceInitializationError)
	}
	for _, name := range []string{"realpath", "cwd"} {
		test.Run(name, func(test *testing.T) {
			source, readError := os.ReadFile("../../stage3/namespace-init-sys/" + name + ".a")
			if readError != nil {
				test.Fatal(readError)
			}
			graph, _ := namespaceGraphForTest(test, string(source))
			if namespaceInitializationError := graph.lowering.namespaceInitialization(
				graph.lowering.program.Files(),
			); namespaceInitializationError != nil {
				test.Fatalf(
					"ambient host declaration acquired pending initialization: %v",
					namespaceInitializationError,
				)
			}
			// Graph-owned callbacks may use supported hosts; unsupported members
			// must still have their own named diagnostic.
			_, lowerSourceError := lowerSource(test, string(source))
			var missing *NotYet
			if name == "realpath" &&
				errors.As(lowerSourceError, &missing) &&
				strings.Contains(missing.What, "observing a callable namespace object") {
				test.Logf("independent callable namespace refusal: %v", lowerSourceError)
				return
			}
			if name == "realpath" && lowerSourceError == nil {
				return
			}
			if name == "cwd" {
				if lowerSourceError != nil {
					test.Fatalf("ambient cwd callback must be admitted: %v", lowerSourceError)
				}
				graphAccepted(test, string(source))
				return
			}
			if lowerSourceError == nil ||
				strings.Contains(lowerSourceError.Error(), "before runtime initialization") ||
				!strings.Contains(lowerSourceError.Error(), "node:") {
				test.Fatalf("expected independent host member NotYet, got %v", lowerSourceError)
			}
			test.Logf("independent lowering stop: %v", lowerSourceError)
		})
	}
}

func TestNamespaceAmbientContextsDoNotExecute(test *testing.T) {
	test.Parallel()
	_, statements := namespaceGraphForTest(test, `
        declare namespace Host { namespace Nested { const value: boolean; } }
        namespace Runtime { export const value = true; }
    `)
	ambient := statements[0]
	nested := namespaceStatements(ambient)[0]
	if namespaceRuntime(ambient) || namespaceRuntime(nested) {
		test.Fatal("ambient namespace or inherited ambient context executes")
	}
	if !namespaceRuntime(statements[1]) {
		test.Fatal("executable namespace erased")
	}
	// Declaration-file context is independently authoritative, even when the
	// explicit modifier and inherited parser flag are absent.
	nested.Flags &^= ast.NodeFlagsAmbient
	ast.GetSourceFileOfNode(nested).IsDeclarationFile = true
	if namespaceRuntime(nested) {
		test.Fatal("declaration-file namespace executes")
	}
}

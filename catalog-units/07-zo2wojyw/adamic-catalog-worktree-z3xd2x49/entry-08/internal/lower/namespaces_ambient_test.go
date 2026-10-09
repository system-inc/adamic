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
func TestNamespaceAmbientHostInitialization(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"realpath", "cwd"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../stage3/namespace-init-sys/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			graph, _ := namespaceGraphForTest(t, string(source))
			if err := graph.lowering.namespaceInitialization(graph.lowering.program.Files()); err != nil {
				t.Fatalf("ambient host declaration acquired pending initialization: %v", err)
			}
			// Unsupported hosts must still have their own named diagnostic.
			_, err = lowerSource(t, string(source))
			var refused *Refused
			if name == "cwd" && errors.As(err, &refused) && strings.Contains(err.Error(), "adamic/cycle-capable") {
				// The area's captured-callback ownership rule remains stricter than the host topic.
				t.Logf("independent ownership refusal: %v", err)
				return
			}
			if err == nil || strings.Contains(err.Error(), "before runtime initialization") || !strings.Contains(err.Error(), "node:") {
				t.Fatalf("expected independent host member NotYet, got %v", err)
			}
			t.Logf("independent lowering stop: %v", err)
		})
	}
}

func TestNamespaceAmbientContextsDoNotExecute(t *testing.T) {
	t.Parallel()
	_, statements := namespaceGraphForTest(t, `
        declare namespace Host { namespace Nested { const value: boolean; } }
        namespace Runtime { export const value = true; }
    `)
	ambient := statements[0]
	nested := namespaceStatements(ambient)[0]
	if namespaceRuntime(ambient) || namespaceRuntime(nested) {
		t.Fatal("ambient namespace or inherited ambient context executes")
	}
	if !namespaceRuntime(statements[1]) {
		t.Fatal("executable namespace erased")
	}
	// Declaration-file context is independently authoritative, even when the
	// explicit modifier and inherited parser flag are absent.
	nested.Flags &^= ast.NodeFlagsAmbient
	ast.GetSourceFileOfNode(nested).IsDeclarationFile = true
	if namespaceRuntime(nested) {
		t.Fatal("declaration-file namespace executes")
	}
}

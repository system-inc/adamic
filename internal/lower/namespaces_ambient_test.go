package lower

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Test preflight separately from unsupported host member lowering. The host
// namespace has no executable body in the checked module graph.
func TestNamespaceAmbientHostInitialization(t *testing.T) {
	t.Parallel()
	// An executable namespace read before initialization must still be rejected.
	graph, _ := namespaceGraphForTest(t, "console.log(`${N.value}`); namespace N { export const value = true; }")
	err := graph.lowering.namespaceInitialization(graph.lowering.program.Files())
	var stop *NotYet
	if !errors.As(err, &stop) || !strings.Contains(err.Error(), "before runtime initialization") {
		t.Fatalf("runtime namespace initialization guard lost: %v", err)
	}
	source := namespaceAmbientHostSource(t, "realpath")
	_, err = lowerSource(t, source)
	if err == nil || strings.Contains(err.Error(), "before runtime initialization") || !strings.Contains(err.Error(), "node:") {
		t.Fatalf("expected independent host member NotYet, got %v", err)
	}
	t.Logf("independent lowering stop: %v", err)
}

func TestNamespaceAmbientCwdInitialization(t *testing.T) {
	t.Parallel()
	// Closure targets prove the callback has no captured back-reference. The
	// host call is supported, so the old cycle refusal is now stale.
	lowersAndAgreesWithNode(t, namespaceAmbientHostSource(t, "cwd"))
}

func namespaceAmbientHostSource(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile("../../stage3/namespace-init-sys/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	graph, _ := namespaceGraphForTest(t, string(source))
	if err := graph.lowering.namespaceInitialization(graph.lowering.program.Files()); err != nil {
		t.Fatalf("ambient host declaration acquired pending initialization: %v", err)
	}
	return string(source)
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

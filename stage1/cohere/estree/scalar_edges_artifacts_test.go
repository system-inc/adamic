package estree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Build-phase units use the same recipes and keys as standalone shards. The Go oracle's row is its owner's, in
// stage1/cohere/estree/oracles.
func TestProduct_scalar_edges_lowered(test *testing.T) {
	test.Parallel()
	directory := scalarEdgeLowered(test, filepath.Join(root(test), "stage1/cohere/estree/main.ts"))
	buildcache.RequireArtifacts(
		test,
		directory,
		buildcache.Artifact{Name: "port.c"},
		buildcache.Artifact{Name: "port.mjs"},
	)
}

func TestProduct_scalar_edges_sanitized_native(test *testing.T) {
	test.Parallel()
	lowered := scalarEdgeLowered(test, filepath.Join(root(test), "stage1/cohere/estree/main.ts"))
	requireScalarEdgeBinary(test, scalarEdgeNative(test, lowered))
}

func requireScalarEdgeBinary(test *testing.T, binary string) {
	test.Helper()
	manifest := filepath.Join(test.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		test.Fatal(err)
	}
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest}},
	)
}

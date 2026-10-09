package estree

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// Build-phase units use the same recipes and keys as standalone shards.
func TestProduct_scalar_edges_go_oracle(t *testing.T) {
	t.Parallel()
	binary := scalarEdgeOracle(t)
	requireScalarEdgeBinary(t, binary)
}

func TestProduct_scalar_edges_lowered(t *testing.T) {
	t.Parallel()
	directory := scalarEdgeLowered(t, filepath.Join(root(t), "stage1/cohere/estree/main.ts"))
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port.c"}, buildcache.Artifact{Name: "port.mjs"})
}

func TestProduct_scalar_edges_sanitized_native(t *testing.T) {
	t.Parallel()
	lowered := scalarEdgeLowered(t, filepath.Join(root(t), "stage1/cohere/estree/main.ts"))
	requireScalarEdgeBinary(t, scalarEdgeNative(t, lowered))
}

func requireScalarEdgeBinary(t *testing.T, binary string) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest}})
}

package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_FactoryHooksLowered(t *testing.T) {
	t.Parallel()
	requireFactoryHooksLoweredArtifacts(t, factoryHooksLowered(t, context.Background(), 0))
}
func TestProduct_FactoryHooksMutantLowered(t *testing.T) {
	t.Parallel()
	requireFactoryHooksLoweredArtifacts(t, factoryHooksLowered(t, context.Background(), 1))
}
func TestProduct_FactoryHooksNative(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	binary := factoryHooksNative(t, ctx, factoryHooksLowered(t, ctx, 0))
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest}})
}

func requireFactoryHooksLoweredArtifacts(t *testing.T, script string) {
	t.Helper()
	buildcache.RequireArtifacts(t, filepath.Dir(script), buildcache.Artifact{Name: "lint.mjs"}, buildcache.Artifact{Name: "lint.c"})
}

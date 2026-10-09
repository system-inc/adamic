package lint

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// C emission and upstream capture exceed the build phase's 60s grain
// (#c5k975w, #3he8f8g). Shards still prepare those products themselves,
// before starting their own deadlines.
func TestProduct_RulesAgreeOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintOracleArtifacts(t, rulesAgreeOracle(t))
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}

// Probe the shared Go oracle with no cases, independently of shard preparation.
func requireLintOracleArtifacts(t *testing.T, directory string) {
	t.Helper()
	manifest := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		t.Fatal(err)
	}
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "oracle", Executable: true, Arguments: []string{"--manifest", manifest}})
}

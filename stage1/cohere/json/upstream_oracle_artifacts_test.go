package json

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_JSONUpstreamOracle(t *testing.T) {
	t.Parallel()
	// upstreamParityOracleProduct returns the oracle binary's path, as buildcache.GoTest does: its product directory
	// holds it.
	binary := upstreamParityOracleProduct(t)
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}})
}

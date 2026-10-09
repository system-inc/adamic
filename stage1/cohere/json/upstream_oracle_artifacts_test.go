package json

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_JSONUpstreamOracle(t *testing.T) {
	t.Parallel()
	buildcache.RequireArtifacts(t, upstreamParityOracleProduct(t), buildcache.Artifact{Name: "oracle", Executable: true, Arguments: []string{"-test.list=^$"}})
}

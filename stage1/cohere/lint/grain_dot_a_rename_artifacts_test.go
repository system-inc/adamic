package lint

import (
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// C emission exceeds the build phase's 60s grain. Keep lowered/native
// products shard-prepared until that production dependency is split.
func TestProduct_DotARenameSource(t *testing.T) {
	t.Parallel()
	started := time.Now()
	buildcache.RequireArtifacts(t, dotARenameSource(t), buildcache.Artifact{Name: "main.ts"}, buildcache.Artifact{Name: "rules/no-var/rule.ts"}, buildcache.Artifact{Name: "rules/no-var/rule.json"})
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}

func TestProduct_DotARenameOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintOracleArtifacts(t, dotARenameOracle(t))
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}
